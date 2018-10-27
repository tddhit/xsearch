package proxy

import (
	"fmt"
	"sync"
)

type shardTable struct {
	sync.RWMutex
	namespace     string
	shardNum      int
	replicaFactor int
	groups        []*shardGroup
}

type shardGroup struct {
	replicas []*shard
}

type shard struct {
	id        string
	groupID   int
	replicaID int
	node      *node
}

func newTable(namespace string, shardNum, replicaFactor int) *shardTable {
	t := &shardTable{
		namespace:     namespace,
		shardNum:      shardNum,
		replicaFactor: replicaFactor,
		groups:        make([]*shardGroup, shardNum),
	}
	for i, _ := range t.groups {
		t.groups[i] = &shardGroup{
			replicas: make([]*shard, replicaFactor),
		}
	}
	return t
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
