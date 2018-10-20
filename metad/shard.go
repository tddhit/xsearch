package metad

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/tddhit/tools/log"

	"github.com/tddhit/xsearch/metad/pb"
)

type shardTable struct {
	sync.RWMutex
	namespace     string
	shardNum      int
	replicaFactor int
	initial       bool
	prepare       map[string]*node
	groups        []*shardGroup
}

type shardGroup struct {
	replicas []*shard
}

type shard struct {
	id        string
	groupID   int
	replicaID int // 0 is primary shard, 1...N is replica shard
	table     *shardTable
	node      *node
	todo      *todo
	backup    *node
}

func newTable(namespace string, shardNum, replicaFactor int) *shardTable {
	t := &shardTable{
		namespace:     namespace,
		shardNum:      shardNum,
		replicaFactor: replicaFactor,
		prepare:       make(map[string]*node),
		groups:        make([]*shardGroup, shardNum),
	}
	for _, group := range t.groups {
		group.replicas = make([]*shard, replicaFactor)
	}
	return t
}

func (t *shardTable) addNode(n *node) error {
	t.Lock()
	defer t.Unlock()

	if _, ok := t.prepare[n.addr]; ok {
		return fmt.Errorf("node(%s) already exist", n.addr)
	}
	t.prepare[n.addr] = n
	return nil
}

func (t *shardTable) removeNode(addr string) {
	t.Lock()
	defer t.Unlock()

	delete(t.prepare, addr)
}

func (t *shardTable) setShard(s *shard) error {
	t.Lock()
	defer t.Unlock()

	if s.groupID >= t.shardNum || s.replicaID >= t.replicaFactor {
		return errors.New("invalid groupID/replicaID ")
	}
	t.groups[s.groupID].replicas[s.replicaID] = s
	s.table = t
	return nil
}

func (t *shardTable) getShard(groupID, replicaID int) (*shard, error) {
	t.RLock()
	defer t.RUnlock()

	if groupID >= t.shardNum || replicaID >= t.replicaFactor {
		return nil, errors.New("invalid groupID/replicaID ")
	}
	return t.groups[groupID].replicas[replicaID], nil
}

func (t *shardTable) autoBalance() error {
	t.Lock()
	defer t.Unlock()

	if !t.initial {
		err := t.firstAllocate()
		t.initial = true
		return err
	}
	// TODO: diff
	return nil
}

func (t *shardTable) firstAllocate() error {
	if len(t.prepare) < t.replicaFactor {
		return errors.New("The replicaFactor is greater than the number of nodes")
	}
	var nodes []*node
	for _, n := range t.prepare {
		nodes = append(nodes, n)
	}
	for i := 0; i < t.shardNum; i++ {
		k := i
		if k >= len(nodes) {
			k = 0
		}
		for j := 0; j < t.replicaFactor; j++ {
			if k >= len(nodes) {
				k = 0
			}
			s := newShard(t.namespace, i, j)
			s.node = nodes[k]
			t.setShard(s)
			k++
		}
	}
	return nil
}

func (t *shardTable) migrate(s *shard, from, to *node) error {
	t.Lock()
	defer t.Unlock()

	if !t.initial {
		return errors.New("Need to perfrom AutoBalance first")
	}
	todo := newTodo(ACTION_MIGRATE_SHARD, from, to, s)
	s.todo = todo
	s.backup = to
	return nil
}

func (t *shardTable) commit() error {
	t.Lock()
	defer t.Unlock()

	for _, group := range t.groups {
		for _, replica := range group.replicas {
			replica.todo.do()
		}
	}
	return nil
}

func (t *shardTable) persist(path string) {
	t.RLock()
	defer t.RUnlock()

	meta := &metadpb.Metadata{
		Namespace:     t.namespace,
		ShardNum:      uint32(t.shardNum),
		ReplicaFactor: uint32(t.replicaFactor),
	}
	for _, group := range t.groups {
		for _, replica := range group.replicas {
			meta.Shards = append(meta.Shards, &metadpb.Metadata_Shard{
				GroupID:   uint32(replica.groupID),
				ReplicaID: uint32(replica.replicaID),
				NodeAddr:  replica.node.addr,
			})
		}
	}
	data, err := proto.Marshal(meta)
	if err != nil {
		log.Error(err)
		return
	}
	tmp := fmt.Sprintf("%s.%d.tmp", path, rand.Int())
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Error(err)
	}
	n, err := f.Write(data)
	if n != len(data) {
		log.Error("persist table fail.")
		return
	}
	if err != nil {
		log.Error(err)
		return
	}
	f.Sync()
	f.Close()
	if err := os.Rename(tmp, path); err != nil {
		log.Error(err)
	}
}

func newShard(namespace string, groupID, replicaID int) *shard {
	return &shard{
		id:        fmt.Sprintf("%s.%d.%d", namespace, groupID, replicaID),
		groupID:   groupID,
		replicaID: replicaID,
	}
}

func (s *shard) execTodo() error {
	if s.todo == nil {
		return nil
	}
	count, err := s.todo.do()
	if count == 0 {
		s.todo = nil
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}
