package service

import (
	"errors"
	"fmt"
	"sync"
)

type resource struct {
	sync.RWMutex
	nodes   map[string]*node
	online  map[string]*shardTable
	offline map[string]*shardTable
}

func newResource() *resource {
	return &resource{
		nodes:   make(map[string]*node),
		online:  make(map[string]*shardTable),
		offline: make(map[string]*shardTable),
	}
}

func (r *resource) getNode(addr string) (*node, bool) {
	r.RLock()
	defer r.RUnlock()

	n, ok := r.nodes[addr]
	return n, ok
}

func (r *resource) getOrCreateNode(addr, adminAddr string) (*node, bool, error) {
	r.RLock()
	if n, ok := r.nodes[addr]; ok {
		r.RUnlock()
		return n, false, nil
	}
	r.RUnlock()

	r.Lock()
	if n, ok := r.nodes[addr]; ok {
		r.Unlock()
		return n, false, nil
	}
	n, err := newNode(addr, adminAddr)
	if err != nil {
		return nil, false, err
	}
	r.nodes[addr] = n
	r.Unlock()

	return n, true, nil
}

func (r *resource) removeNode(addr string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.nodes[addr]; !ok {
		return errors.New("not found")
	}
	delete(r.nodes, addr)
	return nil
}

func (r *resource) getTable(namespace string) (*shardTable, bool) {
	r.RLock()
	defer r.RUnlock()

	table, ok := r.offline[namespace]
	return table, ok
}

func (r *resource) getOrCreateTable(
	namespace string,
	shardNum int,
	replicaFactor int) (_ *shardTable, created bool) {

	r.RLock()
	if t, ok := r.offline[namespace]; ok {
		r.RUnlock()
		return t, false
	}
	r.RUnlock()

	r.Lock()
	if t, ok := r.offline[namespace]; ok {
		r.RUnlock()
		return t, false
	}
	t := newTable(namespace, shardNum, replicaFactor)
	r.offline[namespace] = t
	r.RUnlock()

	return t, true
}

func (r *resource) removeTable(namespace string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.offline[namespace]; !ok {
		return fmt.Errorf("not found table(%s)", namespace)
	}
	delete(r.offline, namespace)
	return nil
}

func (r *resource) commit(namespace string) error {
	return nil
}

func (r *resource) diff() {
}

/*
func (r *resource) allocateShards(
	namespace string,
	shardNum int,
	replicaFactor int,
	startNodeID int) *shardTable {

	table := newTable(namespace, sahrdNum, replicaFactor)
	nodeID := startNodeID
	for i := 0; i < shardNum; i++ {
		nodeID++
		if nodeID >= len(r.nodes) {
			nodeID = 0
		}
		k := nodeID
		for j := 0; j < replicaFactor; j++ {
			s := &shard{
				id:        i,
				namespace: namespace,
			}
			table.groups[i].replicas[j] = s
			if j == 0 {
				s.status = PrimaryShard
				s.node = r.nodeSlice[k]
			}
			k++
		}
	}
	return table
}

func (r *resource) grow() {
	r.Lock()
	defer r.Unlock()

	for {
		minNode, min := r.hasMinShards()
		maxNode, max := r.hasMaxShards()
		if max-min <= 1 {
			break
		}
		shard := maxNode.shards[max-1]
		minNode.shards = append(minNode.shards, shard)
		maxNode.shards = append(maxNode.shards[:max-1], maxNode.shards[max:]...)
		shard.node = minNode
	}
}

func (r *resource) shrink() {
	r.Lock()
	defer r.Unlock()

	for {
		node, shard := waitingForMigrate()
		if shard == nil {
			break
		}
		minNode, _ := hasMinShards()
		node.shards = append(node.shards[:i-1], node.shards[i+1:]...)
		minNode.shards = append(minNode.shards, shard)
		shard.node = minNode
	}
}

func (r *resource) hasMinShards() (*node, int) {
	var (
		min     = 0
		minNode *node
	)
	for _, n := range r.nodes {
		if min >= len(n.shards) {
			min = len(n.shards)
			minNode = n
		}
	}
	return minNode, min
}

func (r *resource) hasMaxShards() (*node, int) {
	var (
		max     = 0
		maxNode *node
	)
	for _, n := range r.nodes {
		if max <= len(n.shards) {
			max = len(n.shards)
			maxNode = n
		}
	}
	return maxNode, max
}

func (r *resource) waitingForMigrate() (*node, *shard) {
	for _, n := range r.nodes {
		if n.status == NodeRemove {
			if len(n.shards) != 0 {
				return n, n.shards[len(n.shards)-1]
			}
		}
	}
	return nil, nil
}
*/
