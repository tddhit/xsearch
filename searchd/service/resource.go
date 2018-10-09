package service

import "sync"

type resource struct {
	sync.RWMutex
	shards map[string]*shard
}

func (r *resource) getOrCreateShard(namespace string, shardID int) (*shard, bool) {
	key := fmt.Sprintf("%s.%d", namespace, shardID)
	r.RLock()
	if s, ok := r.shards[key]; ok {
		r.RUnlock()
		return s, false
	}
	r.RUnlock()

	r.Lock()
	if s, ok := r.shards[key]; ok {
		r.Unlock()
		return s, false
	}
	s := newShard(namespace, shardID)
	r.shards[key] = s
	r.Unlock()

	return s, true
}
