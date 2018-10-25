package proxy

import (
	"sync"
)

type resource struct {
	sync.RWMutex
	nodes  map[string]*node
	tables map[string]*shardTable
}

func newResource() *resource {
	return &resource{
		nodes:  make(map[string]*node),
		tables: make(map[string]*shardTable),
	}
}

func (r *resource) getOrCreateNode(addr string) (*node, error) {
	r.RLock()
	if n, ok := r.nodes[addr]; ok {
		r.RUnlock()
		return n, nil
	}
	r.RUnlock()

	r.Lock()
	if n, ok := r.nodes[addr]; ok {
		r.Unlock()
		return n, nil
	}
	n, err := newNode(addr)
	if err != nil {
		return nil, err
	}
	r.nodes[addr] = n
	r.Unlock()

	return n, nil
}

func (r *resource) getTable(namespace string) (*shardTable, bool) {
	r.RLock()
	defer r.RUnlock()

	table, ok := r.tables[namespace]
	return table, ok
}

func (r *resource) updateTable(
	namespace string,
	shardNum int,
	replicaFactor int) *shardTable {

	r.Lock()
	defer r.Unlock()

	t := newTable(namespace, shardNum, replicaFactor)
	r.tables[namespace] = t
	return t
}
