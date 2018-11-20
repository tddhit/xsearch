package searchd

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/tddhit/diskqueue/pb"
	"github.com/tddhit/xsearch/searchd/pb"
	"github.com/wangbin/jiebago"
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
	addr string,
	segmenter *jiebago.Segmenter,
	stopwords map[string]struct{},
	c diskqueuepb.DiskqueueGrpcClient) (*shard, error) {

	r.Lock()
	defer r.Unlock()

	if s, ok := r.shards[id]; ok {
		return s, errAlreadyExists
	}
	s, err := newShard(id, path.Join(r.dir, id), addr, segmenter, stopwords, c)
	if err != nil {
		return nil, err
	}
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

	type segment struct {
		id   int
		info string
	}
	var segments []*segment
	for _, resourceShard := range r.shards {
		segments = segments[:0]
		pbs := &searchdpb.InfoRsp_Shard{
			ID: resourceShard.id,
		}
		for _, seg := range resourceShard.indexer.Segments {
			id, _ := strconv.Atoi(seg.ID)
			segments = append(segments, &segment{
				id: id,
				info: fmt.Sprintf(
					"ctime=%s    id=%-10dnum=%d",
					time.Unix(
						seg.CreateTime/int64(time.Second),
						seg.CreateTime%int64(time.Second),
					).Format("2006-01-02 15:04:05"),
					id,
					seg.NumDocs,
				),
			})
			pbs.NumDocs += seg.NumDocs
		}
		sort.Slice(segments, func(i, j int) bool {
			return segments[i].id < segments[j].id
		})
		for _, seg := range segments {
			pbs.Segments = append(pbs.Segments, seg.info)
		}
		rsp.Shards = append(rsp.Shards, pbs)
	}
}
