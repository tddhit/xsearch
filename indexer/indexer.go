package indexer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/tddhit/bindex"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/tools/mmap"
	"github.com/tddhit/xsearch/internal/util"
	"github.com/tddhit/xsearch/pb"
)

var (
	defaultOptions = indexerOptions{
		persistNum:    100000,
		mergeInterval: 60 * time.Second,
		dir:           "./",
		id:            "indexer",
		shardNum:      runtime.NumCPU(),
	}
)

type Indexer struct {
	sync.RWMutex
	opt        indexerOptions
	Shards     [][]*segment
	shardLocks []sync.RWMutex
	segmentID  int32
	removeDocs sync.Map // TODO: replace with bitmap
	NumDocs    uint64
	indexDocC  []chan *xsearchpb.Document
	indexRspC  []chan error
	exitC      chan struct{}
	exitMu     sync.RWMutex
	exitFlag   int32
	wg         sync.WaitGroup
}

func New(opts ...IndexerOption) (*Indexer, error) {
	opt := defaultOptions
	for _, o := range opts {
		o(&opt)
	}
	idx := &Indexer{
		opt:        opt,
		Shards:     make([][]*segment, opt.shardNum),
		shardLocks: make([]sync.RWMutex, opt.shardNum),
		indexDocC:  make([]chan *xsearchpb.Document, opt.shardNum),
		indexRspC:  make([]chan error, opt.shardNum),
		exitC:      make(chan struct{}),
	}
	if err := os.MkdirAll(idx.opt.dir, 0755); err != nil && !os.IsExist(err) {
		log.Error(err)
		return nil, err
	}
	if err := idx.loadSegments(); err != nil {
		log.Error(err)
		return nil, err
	}
	for i := 0; i < int(idx.opt.shardNum); i++ {
		idx.shardLocks[i] = sync.RWMutex{}
		idx.indexDocC[i] = make(chan *xsearchpb.Document)
		idx.indexRspC[i] = make(chan error)
		segID := atomic.AddInt32(&idx.segmentID, 1)
		seg, err := idx.openSegment(mmap.MODE_CREATE, segID)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		idx.Shards[i] = append(idx.Shards[i], seg)
		idx.wg.Add(1)
		go idx.indexLoop(i)
	}
	if idx.opt.persistInterval > 0 {
		idx.wg.Add(1)
		go idx.persistLoop()
	}
	if idx.opt.mergeInterval > 0 {
		idx.wg.Add(1)
		go idx.mergeLoop()
	}
	return idx, nil
}

func (idx *Indexer) openSegment(mode int, id int32) (*segment, error) {
	vocabPath := fmt.Sprintf("%s/%s_%d.vocab", idx.opt.dir, idx.opt.id, id)
	invertPath := fmt.Sprintf("%s/%s_%d.invert", idx.opt.dir, idx.opt.id, id)
	switch mode {
	case mmap.MODE_CREATE:
		return newSegment(vocabPath, invertPath, mode)
	case mmap.MODE_RDONLY:
		if _, err := os.Stat(vocabPath); err != nil {
			return nil, err
		}
		if _, err := os.Stat(invertPath); err != nil {
			return nil, err
		}
		return newSegment(vocabPath, invertPath, mode)
	default:
		return nil, errors.New("invalid mode")
	}
}

func (idx *Indexer) loadSegments() error {
	var (
		segids []int
		segs   []*segment
	)
	filepath.Walk(
		idx.opt.dir,
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".vocab") {
				name := info.Name()
				s := strings.Split(name[:len(name)-6], "_")
				if len(s) != 2 {
					return fmt.Errorf("invalid vocab/invert filename:%s", name)
				}
				indexID := s[0]
				if indexID != idx.opt.id {
					return fmt.Errorf("invalid vocab/invert filename:%s", name)
				}
				segmentID, err := strconv.Atoi(s[1])
				if err != nil {
					return err
				}
				segids = append(segids, segmentID)
			}
			return nil
		},
	)
	sort.Ints(segids)
	for _, segID := range segids {
		seg, err := idx.openSegment(mmap.MODE_RDONLY, int32(segID))
		if err != nil {
			return err
		}
		segs = append(segs, seg)
		idx.segmentID = int32(segID)
	}
	idx.shardBalance(segs)
	return nil
}

func (idx *Indexer) indexLoop(shardID int) {
	for {
		select {
		case doc := <-idx.indexDocC[shardID]:
			idx.indexRspC[shardID] <- idx.indexOne(shardID, doc)
		case <-idx.exitC:
			goto exit
		}
	}
exit:
	idx.wg.Done()
}

func (idx *Indexer) indexOne(shardID int, doc *xsearchpb.Document) error {
	idx.shardLocks[shardID].RLock()
	segs := idx.Shards[shardID]
	seg := segs[len(segs)-1]
	if err := seg.indexDocument(doc); err != nil {
		idx.shardLocks[shardID].RUnlock()
		return err
	}
	idx.shardLocks[shardID].RUnlock()

	NumDocs := atomic.LoadUint64(&seg.NumDocs)
	if idx.opt.persistInterval == 0 && NumDocs >= idx.opt.persistNum {
		idx.persist(shardID)
	}
	return nil
}

func (idx *Indexer) IndexDocument(doc *xsearchpb.Document) error {
	idx.exitMu.RLock()
	defer idx.exitMu.RUnlock()

	if atomic.LoadInt32(&idx.exitFlag) == 1 {
		return errors.New("Already Close.")
	}

	id, err := uuid.FromString(doc.ID)
	if err != nil {
		log.Error(err)
		return err
	}
	i := util.Hash(id[:]) % uint64(idx.opt.shardNum)
	log.Infof("doc(%s) is indexed by shard(%d).", doc.GetID(), i)
	idx.indexDocC[i] <- doc
	return <-idx.indexRspC[i]
}

func (idx *Indexer) RemoveDocument(docID string) {
	idx.removeDocs.Store(docID, struct{}{})
}

// TODO: Need internal DocID
func (idx *Indexer) UpdateDocument(doc *xsearchpb.Document) error {
	return nil
}

func (idx *Indexer) Search(query *xsearchpb.Query,
	start uint64, count int32) ([]*xsearchpb.Document, error) {

	shards, n := idx.copy()
	var (
		wg    sync.WaitGroup
		docsC = make(chan *xsearchpb.Document, n*count)
		docs  = make([]*xsearchpb.Document, 0, n*count)
	)
	for _, segs := range shards {
		for _, seg := range segs {
			wg.Add(1)
			go func(seg *segment) {
				docs, err := seg.search(query, start, count)
				if err == nil {
					for _, doc := range docs {
						docsC <- doc
					}
				}
				wg.Done()
			}(seg)
		}
	}
	wg.Wait()
	close(docsC)
	for doc := range docsC {
		if _, ok := idx.removeDocs.Load(doc.ID); ok {
			continue
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (idx *Indexer) copy() ([][]*segment, int32) {
	count := int32(0)
	shards := make([][]*segment, idx.opt.shardNum)
	for i := range shards {
		idx.shardLocks[i].RLock()
		shards[i] = make([]*segment, len(idx.Shards[i]))
		for j := range shards[i] {
			shards[i][j] = idx.Shards[i][j]
			count++
		}
		idx.shardLocks[i].RUnlock()
	}
	return shards, count
}

func (idx *Indexer) persist(shardID int) error {
	idx.RLock()
	defer idx.RUnlock()

	segID := atomic.AddInt32(&idx.segmentID, 1)
	newSeg, err := idx.openSegment(mmap.MODE_CREATE, segID)
	if err != nil {
		return err
	}
	idx.shardLocks[shardID].Lock()
	segs := idx.Shards[shardID]
	oldSeg := segs[len(segs)-1]
	// Use idx.shards to avoid append allocation of new variables
	idx.Shards[shardID] = append(idx.Shards[shardID], newSeg)
	idx.shardLocks[shardID].Unlock()

	idx.wg.Add(1)
	go func(seg *segment) {
		if err := seg.persistData(); err != nil {
			log.Error(err)
		}
		idx.wg.Done()
	}(oldSeg)
	return nil
}

func (idx *Indexer) persistAll() {
	idx.exitMu.RLock()
	defer idx.exitMu.RUnlock()

	if atomic.LoadInt32(&idx.exitFlag) == 1 {
		return
	}
	for i := 0; i < idx.opt.shardNum; i++ {
		idx.persist(i)
	}
}

func (idx *Indexer) mergeSegments() {
	var (
		NumDocs   uint64
		needMerge []*segment
		newSegs   []*segment
	)
	shards, _ := idx.copy()
	for i, segs := range shards {
		for j, seg := range segs {
			merge := false
			if atomic.LoadInt32(&seg.persist) == 1 {
				if seg.NumDocs > 0 && seg.NumDocs < idx.opt.persistNum {
					seg.recycle = true
					needMerge = append(needMerge, seg)
					NumDocs += seg.NumDocs
				} else if seg.NumDocs == 0 {
					seg.recycle = true
				}
			}
			if NumDocs >= idx.opt.persistNum {
				merge = true
			}
			if i == len(shards)-1 && j == len(segs)-1 {
				if len(needMerge) > 1 {
					merge = true
				} else if len(needMerge) == 1 {
					needMerge[0].recycle = false
				}
			}
			if merge {
				newSeg, err := idx.merge(needMerge)
				if err != nil {
					log.Error(err)
					return
				}
				newSegs = append(newSegs, newSeg)
				NumDocs = 0
				needMerge = needMerge[:0]
			}
		}
	}
	idx.shardBalance(newSegs)
}

func (idx *Indexer) shardBalance(newSegs []*segment) {
	idx.Lock()
	defer idx.Unlock()

	newShards := make([][]*segment, idx.opt.shardNum)
	i := 0
	for _, segs := range idx.Shards {
		for _, seg := range segs {
			if seg.recycle {
				seg.delete()
			} else {
				newShards[i] = append(newShards[i], seg)
				i = (i + 1) % idx.opt.shardNum
			}
		}
	}
	for _, seg := range newSegs {
		newShards[i] = append(newShards[i], seg)
		i = (i + 1) % idx.opt.shardNum
	}
	idx.Shards = newShards
}

func (idx *Indexer) merge(segs []*segment) (*segment, error) {
	k := len(segs)
	segID := atomic.AddInt32(&idx.segmentID, 1)
	newSeg, err := idx.openSegment(mmap.MODE_CREATE, segID)
	if err != nil {
		return nil, err
	}
	cursors := make([]*bindex.Cursor, k)
	var (
		max    []byte
		offset int64
	)
	for _, seg := range segs {
		newSeg.NumDocs += seg.NumDocs
		c := seg.vocab.NewCursor()
		key, _ := c.Last()
		if max == nil || bytes.Compare(key, max) > 0 {
			max = key
		}
	}
	max = append(max, '0')
	input := func(i int) interface{} {
		if i >= k {
			return max
		}
		if cursors[i] == nil {
			cursors[i] = segs[i].vocab.NewCursor()
			key, _ := cursors[i].First()
			return key
		}
		key, _ := cursors[i].Next()
		if key == nil {
			return max
		}
		return key
	}
	output := func(k interface{}) error {
		if newSeg.vocab.Get(k.([]byte)) != nil {
			return nil
		}
		var (
			total uint64
			start = offset
		)
		offset += 8
		for _, seg := range segs {
			b := seg.vocab.Get(k.([]byte))
			if b == nil {
				continue
			}
			off := int64(binary.LittleEndian.Uint64(b))
			count, err := seg.invert.Uint64At(off)
			if err != nil {
				return err
			}
			total += count
			data, err := seg.invert.ReadAt(off+8, int64(count*20))
			if err != nil {
				return err
			}
			newSeg.invert.WriteAt(data, offset)
			offset += int64(count * 20)
		}
		//b := pool.Get().([]byte)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(start))
		newSeg.vocab.Put(k.([]byte), b)
		newSeg.invert.PutUint64At(start, total)
		return nil
	}
	compare := func(a, b interface{}) int {
		return bytes.Compare(a.([]byte), b.([]byte))
	}
	util.KMerge(len(segs), []byte(""), max, input, output, compare)
	return newSeg, nil
}

func (idx *Indexer) mergeLoop() {
	ticker := time.NewTicker(idx.opt.mergeInterval)
	for {
		select {
		case <-ticker.C:
			idx.mergeSegments()
		case <-idx.exitC:
			goto exit
		}
	}
exit:
	ticker.Stop()
	idx.wg.Done()
}

func (idx *Indexer) persistLoop() {
	ticker := time.NewTicker(idx.opt.persistInterval)
	for {
		select {
		case <-ticker.C:
			idx.persistAll()
		case <-idx.exitC:
			goto exit
		}
	}
exit:
	ticker.Stop()
	idx.wg.Done()
}

func (idx *Indexer) Close() {
	idx.exitMu.RLock()
	defer idx.exitMu.RUnlock()

	if !atomic.CompareAndSwapInt32(&idx.exitFlag, 0, 1) {
		return
	}
	close(idx.exitC)
	idx.wg.Wait()
	var wg sync.WaitGroup
	wg.Add(idx.opt.shardNum)
	for i := 0; i < idx.opt.shardNum; i++ {
		go func(i int) {
			idx.shardLocks[i].Lock()
			segs := idx.Shards[i]
			seg := segs[len(segs)-1]
			if err := seg.persistData(); err != nil {
				log.Error(err)
			}
			for j := range segs {
				log.Error(i, j, "close")
				segs[j].close()
			}
			idx.shardLocks[i].Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()
}
