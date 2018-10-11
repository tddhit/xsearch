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
	todos   []*todo
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

func (r *resource) createTable(namespace string, shardNum, replicaFactor int) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.offline[namespace]; ok {
		return fmt.Errorf("(%s)already exists in offline table.", namespace)
	}
	if _, ok := r.online[namespace]; ok {
		return fmt.Errorf("(%s)already exists in online table.", namespace)
	}
	r.online[namespace] = newTable(namespace, shardNum, replicaFactor)
	r.offline[namespace] = newTable(namespace, shardNum, replicaFactor)
	return nil
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

func (r *resource) getOnlineTable(namespace string) (*shardTable, bool) {
	r.RLock()
	defer r.RUnlock()

	table, ok := r.online[namespace]
	return table, ok
}
