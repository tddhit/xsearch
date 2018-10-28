package proxy

import (
	"sync"

	"github.com/gogo/protobuf/jsonpb"

	"github.com/tddhit/xsearch/proxy/pb"
)

type Resource struct {
	sync.RWMutex
	nodes  map[string]*node
	tables map[string]*shardTable
}

func NewResource() *Resource {
	return &Resource{
		nodes:  make(map[string]*node),
		tables: make(map[string]*shardTable),
	}
}

func (r *Resource) getOrCreateNode(addr, status string) (*node, error) {
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
	n, err := newNode(addr, status)
	if err != nil {
		return nil, err
	}
	r.nodes[addr] = n
	r.Unlock()

	return n, nil
}

func (r *Resource) getTable(namespace string) (*shardTable, bool) {
	r.RLock()
	defer r.RUnlock()

	table, ok := r.tables[namespace]
	return table, ok
}

func (r *Resource) updateTable(
	namespace string,
	shardNum int,
	replicaFactor int) *shardTable {

	r.Lock()
	defer r.Unlock()

	t := newTable(namespace, shardNum, replicaFactor)
	r.tables[namespace] = t
	return t
}

func (r *Resource) marshalTo(rsp *proxypb.InfoRsp) string {
	r.RLock()
	defer r.RUnlock()

	rsp.Tables = make(map[string]*proxypb.InfoRsp_Table)
	for _, t := range r.tables {
		var (
			shards []*proxypb.InfoRsp_Shard
		)
		for _, group := range t.groups {
			for _, replica := range group.replicas {
				shards = append(shards, &proxypb.InfoRsp_Shard{
					ID:   replica.id,
					Node: replica.node.addr + ":" + replica.node.status,
				})
			}
		}
		rsp.Tables[t.namespace] = &proxypb.InfoRsp_Table{
			Namespace:     t.namespace,
			ShardNum:      uint32(t.shardNum),
			ReplicaFactor: uint32(t.replicaFactor),
			Shards:        shards,
		}
	}
	marshaler := &jsonpb.Marshaler{
		EnumsAsInts: true,
		Indent:      "    ",
	}
	if s, err := marshaler.MarshalToString(rsp); err != nil {
		return err.Error()
	} else {
		return s
	}
}
