package metad

import (
	"fmt"
	"strings"
	"sync"

	"github.com/tddhit/xsearch/metad/pb"
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

func (r *reception) notifyByNamespace(namespace string, meta *metadpb.Metadata) {
	r.RLock()
	defer r.RUnlock()

	for key, c := range r.clients {
		if strings.HasPrefix(key, namespace) {
			select {
			case c.writeC <- &metadpb.RegisterClientRsp{
				Table: meta,
			}:
			default:
			}
		}
	}
}

func (r *reception) broadcast(meta *metadpb.Metadata) {
	r.RLock()
	defer r.RUnlock()

	for _, c := range r.clients {
		c.writeC <- &metadpb.RegisterClientRsp{
			Table: meta,
		}
	}
}
