package searchd

import (
	"errors"
	"sync"

	diskqueuepb "github.com/tddhit/diskqueue/pb"
)

var (
	errNotFound      = errors.New("not found")
	errAlreadyExists = errors.New("already exists")
)

type resource struct {
	sync.RWMutex
	dir    string
	shards map[string]*shard
}

func newResource(dir string) *resource {
	return &resource{
		dir:    dir,
		shards: make(map[string]*shard),
	}
}

func (r *resource) getShard(id string) (*shard, bool) {
	r.RLock()
	defer r.RUnlock()

	s, ok := r.shards[id]
	return s, ok
}

func (r *resource) createShard(
	id string,
	c diskqueuepb.DiskqueueGrpcClient) (*shard, error) {

	r.Lock()
	defer r.Unlock()

	if s, ok := r.shards[id]; ok {
		return s, errAlreadyExists
	}
	s := newShard(id, r.dir, c)
	r.shards[id] = s
	return s, nil
}

func (r *resource) removeShard(id string) error {
	r.Lock()
	defer r.Unlock()

	if s, ok := r.shards[id]; !ok {
		return errNotFound
	} else {
		go func(s *shard) {
			s.close()
		}(s)
	}
	delete(r.shards, id)
	return nil
}
