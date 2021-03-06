package metad

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/tddhit/tools/log"
	metadpb "github.com/tddhit/xsearch/metad/pb"
)

type resource struct {
	sync.Mutex
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

func (r *resource) rangeTables(f func(*shardTable) error) {
	for _, table := range r.tables {
		if err := f(table); err != nil {
			break
		}
	}
}

func (r *resource) getNode(addr string) (*node, bool) {
	n, ok := r.nodes[addr]
	return n, ok
}

func (r *resource) createNode(addr string) (*node, error) {
	if n, ok := r.nodes[addr]; ok {
		return n, fmt.Errorf("node(%s) already exists", addr)
	}
	n := newNode(addr, NODE_ISOLATED_ONLINE)
	r.nodes[addr] = n
	return n, nil
}

func (r *resource) removeNode(addr string) {
	delete(r.nodes, addr)
}

func (r *resource) persistTables() {
	for _, t := range r.tables {
		t.persist(r.dataDir)
	}
}

func (r *resource) createTable(namespace string, shardNum, replicaFactor int) error {
	if _, ok := r.tables[namespace]; ok {
		return fmt.Errorf("(%s)already exists in mirror table.", namespace)
	}
	r.tables[namespace] = newTable(namespace, shardNum, replicaFactor)
	return nil
}

func (r *resource) removeTable(namespace string) error {
	if _, ok := r.tables[namespace]; !ok {
		return fmt.Errorf("not found table(%s)", namespace)
	}
	delete(r.tables, namespace)
	return nil
}

func (r *resource) getTable(namespace string) (*shardTable, bool) {
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

/*
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
*/

func (r *resource) marshalTo(res *metadpb.Resource) {
	res.Tables = make(map[string]*metadpb.Resource_Table)
	for _, n := range r.nodes {
		res.Nodes = append(res.Nodes, n.getInfo())
	}
	for _, t := range r.tables {
		var (
			shards []*metadpb.Resource_Shard
			nodes  []string
		)
		for _, n := range t.prepare {
			nodes = append(nodes, n.getInfo())
		}
		for _, group := range t.groups {
			for _, replica := range group.replicas {
				shards = append(shards, &metadpb.Resource_Shard{
					ID:   replica.id,
					Node: replica.node.getInfo(),
					Next: replica.next.getInfo(),
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
