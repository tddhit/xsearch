package metad

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/metad/pb"
)

type shardTable struct {
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
	next      *node
}

func newTable(namespace string, shardNum, replicaFactor int) *shardTable {
	t := &shardTable{
		namespace:     namespace,
		shardNum:      shardNum,
		replicaFactor: replicaFactor,
		prepare:       make(map[string]*node),
		groups:        make([]*shardGroup, shardNum),
	}
	for i, _ := range t.groups {
		t.groups[i] = &shardGroup{
			replicas: make([]*shard, replicaFactor),
		}
		for j, _ := range t.groups[i].replicas {
			s := newShard(namespace, i, j)
			t.setShard(s)
		}
	}
	return t
}

func (t *shardTable) addNode(n *node) error {
	if _, ok := t.prepare[n.addr]; ok {
		return fmt.Errorf("node(%s) already exist", n.addr)
	}
	t.prepare[n.addr] = n
	return nil
}

func (t *shardTable) removeNode(addr string) {
	delete(t.prepare, addr)
}

func (t *shardTable) getNode(addr string) (*node, bool) {
	n, ok := t.prepare[addr]
	return n, ok
}

func (t *shardTable) setShard(s *shard) error {
	if s.groupID >= t.shardNum || s.replicaID >= t.replicaFactor {
		return errors.New("invalid groupID/replicaID ")
	}
	t.groups[s.groupID].replicas[s.replicaID] = s
	s.table = t
	return nil
}

func (t *shardTable) getShard(groupID, replicaID int) (*shard, error) {
	if groupID >= t.shardNum || replicaID >= t.replicaFactor {
		return nil, errors.New("invalid groupID/replicaID ")
	}
	return t.groups[groupID].replicas[replicaID], nil
}

func (t *shardTable) autoBalance() error {
	if !t.initial {
		if err := t.firstAllocate(); err != nil {
			return err
		}
		t.initial = true
	} else {
		// TODO
	}
	t.diff()
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
			t.groups[i].replicas[j].next = nodes[k]
			k++
		}
	}
	return nil
}

func (t *shardTable) diff() {
	for _, group := range t.groups {
		for _, replica := range group.replicas {
			if replica.next == nil {
				continue
			}
			if replica.node == nil {
				newTodo(ACTION_CREATE_SHARD, replica.next, nil, replica)
			} else {
				newTodo(ACTION_MIGRATE_SHARD, replica.node, replica.next, replica)
			}
		}
	}
}

func (t *shardTable) migrate(s *shard, from, to *node) error {
	if !t.initial {
		return errors.New("Need to perfrom AutoBalance first")
	}
	s.next = to
	newTodo(ACTION_MIGRATE_SHARD, from, to, s)
	return nil
}

func (t *shardTable) commit() error {
	var wg sync.WaitGroup
	for _, group := range t.groups {
		for _, replica := range group.replicas {
			if replica.todo == nil {
				continue
			}
			wg.Add(1)
			replica.todo.do(&wg)
		}
	}
	wg.Wait()
	return nil
}

func (t *shardTable) persist(path string) {
	meta := &metadpb.Metadata{
		Namespace:     t.namespace,
		ShardNum:      uint32(t.shardNum),
		ReplicaFactor: uint32(t.replicaFactor),
	}
	for _, group := range t.groups {
		for _, replica := range group.replicas {
			meta.Shards = append(meta.Shards, &metadpb.Metadata_Shard{
				GroupID:    uint32(replica.groupID),
				ReplicaID:  uint32(replica.replicaID),
				NodeAddr:   replica.node.getAddr(),
				NodeStatus: replica.node.getClusterStatus(),
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

func (t *shardTable) marshal() *metadpb.Metadata {
	meta := &metadpb.Metadata{
		Namespace:     t.namespace,
		ShardNum:      uint32(t.shardNum),
		ReplicaFactor: uint32(t.replicaFactor),
	}
	for _, group := range t.groups {
		for _, replica := range group.replicas {
			meta.Shards = append(meta.Shards, &metadpb.Metadata_Shard{
				GroupID:    uint32(replica.groupID),
				ReplicaID:  uint32(replica.replicaID),
				NodeAddr:   replica.node.getAddr(),
				NodeStatus: replica.node.getClusterStatus(),
			})
		}
	}
	return meta
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
