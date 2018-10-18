package service

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type shardTable struct {
	sync.RWMutex
	namespace     string
	shardNum      int
	replicaFactor int
	initial       bool
	nodes         []*node
	groups        []*shardGroup
	todos         []*todo
	todoSequence  uint64
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
}

func newTable(namespace string, shardNum, replicaFactor int) *shardTable {
	t := &shardTable{
		namespace:     namespace,
		shardNum:      shardNum,
		replicaFactor: replicaFactor,
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

	for _, node := range t.nodes {
		if node.addr == n.addr {
			return errors.New("already exists")
		}
	}
	t.nodes = append(t.nodes, n)
	return nil
}

func (t *shardTable) removeNode(n *node) error {
	t.Lock()
	defer t.Unlock()

	i := -1
	for j, node := range t.nodes {
		if node.addr == n.addr {
			i = j
			break
		}
	}
	if i == -1 {
		return errors.New("not found")
	}
	t.nodes = append(t.nodes[:i], t.nodes[i+1:]...)
	return nil
}

func (t *shardTable) replaceNode(old, new *node) error {
	t.Lock()
	defer t.Unlock()

	i := -1
	for j, node := range t.nodes {
		if node.addr == old.addr {
			i = j
			break
		}
	}
	if i == -1 {
		return errors.New("not found")
	}
	t.nodes[i] = new
	return nil
}

func (t *shardTable) setShard(s *shard) {
	t.Lock()
	defer t.Unlock()

	t.groups[s.groupID].replicas[s.replicaID] = s
}

func (t *shardTable) getShard(shardID, replicaID int) (*shard, error) {
	t.RLock()
	defer t.RUnlock()

	if shardID >= t.shardNum {
		return nil, fmt.Errorf("shardID(%d)>shardNum(%d)", shardID, t.shardNum)
	}
	if replicaID >= t.replicaFactor {
		return nil, fmt.Errorf(
			"replicaID(%d)>replicaFactor(%d)",
			replicaID,
			t.replicaFactor,
		)
	}
	return t.groups[shardID].replicas[replicaID], nil
}

func (t *shardTable) autoBalance() error {
	if !t.initial {
		err := t.firstAllocate()
		t.initial = true
		return err
	}
	// TODO: diff
	return nil
}

func (t *shardTable) firstAllocate() error {
	if len(t.nodes) < t.replicaFactor {
		return errors.New("replicaFactor > nodesNum")
	}
	for i := 0; i < t.shardNum; i++ {
		k := i
		if k >= len(t.nodes) {
			k = 0
		}
		for j := 0; j < t.replicaFactor; j++ {
			if k >= len(t.nodes) {
				k = 0
			}
			s := &shard{
				id:        fmt.Sprintf("%s.%d.%d", t.namespace, i, j),
				groupID:   i,
				replicaID: j,
				table:     t,
				node:      t.nodes[k],
			}
			t.setShard(s)
			k++
		}
	}
	return nil
}

func (t *shardTable) migrate(s *shard, from, to *node, online *shardTable) error {
	if !t.initial {
		return errors.New("need to perfrom AutoBalance first")
	}
	id := atomic.AddUint64(&t.todoSequence, 1)
	d := newTodo(id, ACTION_MIGRATE_SHARD, from, to, s, online)
	t.todos = append(t.todos, d)
	return nil
}

func (t *shardTable) commit() error {
	for _, todo := range t.todos {
		if _, err := todo.do(); err != nil {
			return err
		}
	}
	return nil
}
