package service

import (
	"errors"
	"sync"
)

type shardTable struct {
	sync.RWMutex
	namespace     string
	shardNum      int
	replicaFactor int
	nodes         map[string]*node
	groups        []*shardGroup
}

type shardGroup struct {
	replicas []*shard
}

type shard struct {
	id        int
	replicaID int // 0 is primary shard, 1...N is replica shard
	table     *shardTable
	node      *node
}

func newTable(namespace string, shardNum, replicaFactor int) *shardTable {
	t := &shardTable{
		namespace:     namespace,
		shardNum:      shardNum,
		replicaFactor: replicaFactor,
		nodes:         make(map[string]*node),
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

	if _, ok := t.nodes[n.addr]; ok {
		return errors.New("already exists")
	}
	t.nodes[n.addr] = n
	return nil
}

func (t *shardTable) removeNode(n *node) error {
	t.Lock()
	defer t.Unlock()

	if _, ok := t.nodes[n.addr]; !ok {
		return errors.New("not found")
	}
	delete(t.nodes, n.addr)
	return nil
}

func (t *shardTable) replaceNode(old, new *node) error {
	t.Lock()
	defer t.Unlock()

	if _, ok := t.nodes[old.addr]; !ok {
		return errors.New("not found")
	}
	delete(t.nodes, old.addr)
	t.nodes[new.addr] = new
	return nil
}

func (t *shardTable) setShard(s *shard) {
	t.Lock()
	defer t.Unlock()

	t.groups[s.id].replicas[s.replicaID] = s
}

/*
func (t *table) diff(another *table) {
}

func (s *shard) createInSearchd() {
	_, err := s.node.client.CreateShard(
		context.Background(),
		&searchdpb.CreateShardReq{
			ShardID:   uint32(s.id),
			Namespace: s.namespace,
		},
	)
	if err != nil {
		log.Error(err)
	}
}

func (s *shard) removeInSearchd() {
}
*/
