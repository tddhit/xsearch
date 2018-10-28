package searchd

import (
	"errors"
	"path"
	"sync"

	"github.com/tddhit/diskqueue/pb"
	"github.com/tddhit/xsearch/searchd/pb"
)

var (
	errNotFound      = errors.New("not found")
	errAlreadyExists = errors.New("already exists")
)

type Resource struct {
	sync.RWMutex
	dir    string
	shards map[string]*shard
}

func NewResource(dir string) *Resource {
	return &Resource{
		dir:    dir,
		shards: make(map[string]*shard),
	}
}

func (r *Resource) getShard(id string) (*shard, bool) {
	r.RLock()
	defer r.RUnlock()

	s, ok := r.shards[id]
	return s, ok
}

func (r *Resource) createShard(
	id string,
	channel string,
	c diskqueuepb.DiskqueueGrpcClient) (*shard, error) {

	r.Lock()
	defer r.Unlock()

	if s, ok := r.shards[id]; ok {
		return s, errAlreadyExists
	}
	s := newShard(id, path.Join(r.dir, id), channel, c)
	r.shards[id] = s
	return s, nil
}

func (r *Resource) removeShard(id string) error {
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

func (r *Resource) marshalTo(rsp *searchdpb.InfoRsp) {
	r.RLock()
	defer r.RUnlock()

	for _, s := range r.shards {
		rsp.Shards = append(rsp.Shards, s.id)
	}
}
