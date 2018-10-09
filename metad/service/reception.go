package service

import (
	"fmt"
	"sync"
)

type reception struct {
	sync.RWMutex
	clients map[string]*client
}

func newReception() *reception {
	return &reception{
		clients: make(map[string]*client),
	}
}

func (r *reception) getClient(namespace, addr string) (*client, bool) {
	r.RLock()
	defer r.RUnlock()

	key := fmt.Sprintf("%s.%s", namespace, addr)
	c, ok := r.clients[key]
	return c, ok
}

func (r *reception) getOrCreateClient(namespace, addr string) (*client, bool) {
	key := fmt.Sprintf("%s.%d", namespace, addr)
	r.RLock()
	if c, ok := r.clients[key]; ok {
		r.RUnlock()
		return c, false
	}
	r.RUnlock()

	r.Lock()
	if c, ok := r.clients[key]; ok {
		r.Unlock()
		return c, false
	}
	c := newClient(addr, namespace)
	r.clients[key] = c
	r.Unlock()

	return c, true
}
