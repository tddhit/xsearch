package metad

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/tddhit/tools/log"
	metadpb "github.com/tddhit/xsearch/metad/pb"
)

type resource struct {
	sync.RWMutex
	dataDir string
	nodes   map[string]*node
	tables  map[string]*shardTable
}

func newResource(dataDir string) *resource {
	r := &resource{
		dataDir: dataDir,
		nodes:   make(map[string]*node),
		tables:  make(map[string]*shardTable),
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".meta") {
			metaPath := filepath.Join(dataDir, file.Name())
			r.loadTable(metaPath)
		}
	}
	return r
}

func (r *resource) getNode(addr string) (*node, bool) {
	r.RLock()
	defer r.RUnlock()

	n, ok := r.nodes[addr]
	return n, ok
}

func (r *resource) createNode(addr string) (*node, error) {
	r.Lock()
	defer r.Unlock()

	if n, ok := r.nodes[addr]; ok {
		return n, fmt.Errorf("node(%s) already exists", addr)
	}
	n := newNode(addr, NODE_ISOLATED_ONLINE)
	r.nodes[addr] = n
	return n, nil
}

func (r *resource) removeNode(addr string) {
	r.Lock()
	defer r.Unlock()

	delete(r.nodes, addr)
}

func (r *resource) persistTables() {
	r.RLock()
	defer r.RUnlock()

	for _, t := range r.tables {
		t.persist(r.dataDir)
	}
}

func (r *resource) createTable(namespace string, shardNum, replicaFactor int) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.tables[namespace]; ok {
		return fmt.Errorf("(%s)already exists in mirror table.", namespace)
	}
	r.tables[namespace] = newTable(namespace, shardNum, replicaFactor)
	return nil
}

func (r *resource) removeTable(namespace string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.tables[namespace]; !ok {
		return fmt.Errorf("not found table(%s)", namespace)
	}
	delete(r.tables, namespace)
	return nil
}

func (r *resource) getTable(namespace string) (*shardTable, bool) {
	r.RLock()
	defer r.RUnlock()

	table, ok := r.tables[namespace]
	return table, ok
}

func (r *resource) loadTable(metaPath string) {
	data, err := ioutil.ReadFile(metaPath)
	if err != nil {
		log.Fatal(err)
	}
	meta := &metadpb.Metadata{}
	if err := proto.Unmarshal(data, meta); err != nil {
		log.Fatal(err)
	}
	err = r.createTable(meta.Namespace, int(meta.ShardNum), int(meta.ReplicaFactor))
	if err != nil {
		log.Fatal(err)
	}
	table, ok := r.getTable(meta.Namespace)
	if !ok {
		log.Fatalf("online table(%s) does not exist", meta.Namespace)
	}
	for _, s := range meta.Shards {
		shard := newShard(meta.Namespace, int(s.GroupID), int(s.ReplicaID))
		shard.node = newNode(s.NodeAddr, NODE_CLUSTER_OFFLINE)
		table.setShard(shard)
	}
}

func (r *resource) getShard(
	namespace string,
	groupID int,
	replicaID int) (*shard, error) {

	table, ok := r.getTable(namespace)
	if !ok {
		log.Errorf("not found table(%s)", namespace)
		return nil, fmt.Errorf("not found table(%s)", namespace)
	}
	shard, err := table.getShard(groupID, replicaID)
	if err != nil {
		return nil, err
	}
	return shard, nil
}

func (r *resource) activeShards(n *node) {
	for _, table := range r.tables {
		for _, group := range table.groups {
			for _, replica := range group.replicas {
				if replica.node.getAddr() == n.addr {
					replica.node = n
				}
			}
		}
	}
}

func (r *resource) marshalTo(res *metadpb.Resource) {
	r.RLock()
	defer r.RUnlock()

	res.Tables = make(map[string]*metadpb.Resource_Table)
	for _, n := range r.nodes {
		res.Nodes = append(res.Nodes, n.getStatus())
	}
	for _, t := range r.tables {
		var (
			shards []*metadpb.Resource_Shard
			nodes  []string
		)
		for _, n := range t.prepare {
			nodes = append(nodes, n.getStatus())
		}
		for _, group := range t.groups {
			for _, replica := range group.replicas {
				shards = append(shards, &metadpb.Resource_Shard{
					ID:   replica.id,
					Node: replica.node.getStatus(),
					Next: replica.next.getStatus(),
					Todo: replica.todo.getInfo(),
				})
			}
		}
		res.Tables[t.namespace] = &metadpb.Resource_Table{
			Namespace:     t.namespace,
			ShardNum:      uint32(t.shardNum),
			ReplicaFactor: uint32(t.replicaFactor),
			Nodes:         nodes,
			Shards:        shards,
		}
	}
}
