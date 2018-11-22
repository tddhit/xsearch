package indexer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tddhit/bindex"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/tools/mmap"
	"github.com/tddhit/xsearch/internal/util"
	"github.com/tddhit/xsearch/pb"
)

var (
	defaultOptions = indexerOptions{
		persistInterval: 10 * time.Second,
		mergeInterval:   60 * time.Second,
		mergeNum:        100000,
		dir:             "./data/indexer/",
		id:              "indexer",
	}
)

type Indexer struct {
	opt        indexerOptions
	Segments   []*segment
	segmentID  int32
	segmentsMu sync.RWMutex
	persistMu  sync.RWMutex
	mergeMu    sync.RWMutex
	exitMu     sync.RWMutex
	exitC      chan struct{}
	exitFlag   int32
	wg         sync.WaitGroup
}

func New(opts ...IndexerOption) (*Indexer, error) {
	opt := defaultOptions
	for _, o := range opts {
		o(&opt)
	}
	idx := &Indexer{
		opt:   opt,
		exitC: make(chan struct{}),
	}
	if err := os.MkdirAll(idx.opt.dir, 0755); err != nil && !os.IsExist(err) {
		log.Error(err)
		return nil, err
	}
	if err := idx.loadSegments(); err != nil {
		return nil, err
	}
	if seg, err := idx.openSegment(mmap.MODE_CREATE, 0); err != nil {
		log.Error(err)
		return nil, err
	} else {
		idx.Segments = append(idx.Segments, seg)
	}
	if idx.opt.persistInterval > 0 {
		idx.wg.Add(1)
		go idx.persistLoop()
	}
	if idx.opt.mergeInterval > 0 && idx.opt.mergeNum > 0 {
		idx.wg.Add(1)
		go idx.mergeLoop()
	}
	return idx, nil
}

func (idx *Indexer) openSegment(mode int, id int32) (*segment, error) {
	if mode == mmap.MODE_CREATE {
		id = atomic.AddInt32(&idx.segmentID, 1)
	}
	vocabPath := fmt.Sprintf("%s/%d.vocab", idx.opt.dir, id)
	invertPath := fmt.Sprintf("%s/%d.invert", idx.opt.dir, id)
	if mode == mmap.MODE_RDONLY {
		if _, err := os.Stat(vocabPath); err != nil {
			log.Error(err)
			return nil, err
		}
		if _, err := os.Stat(invertPath); err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return newSegment(vocabPath, invertPath, mode)
}

func (idx *Indexer) loadSegments() error {
	var segids []int
	filepath.Walk(
		idx.opt.dir,
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".vocab") {
				name := info.Name()
				s := strings.Split(name[:len(name)-6], "/")
				if segmentID, err := strconv.Atoi(s[len(s)-1]); err != nil {
					log.Error(err)
					return err
				} else {
					segids = append(segids, segmentID)
				}
			}
			return nil
		},
	)
	sort.Ints(segids)
	for _, id := range segids {
		if seg, err := idx.openSegment(mmap.MODE_RDONLY, int32(id)); err != nil {
			return err
		} else {
			idx.Segments = append(idx.Segments, seg)
			idx.segmentID = int32(id)
		}
	}
	return nil
}

func (idx *Indexer) IndexDoc(doc *xsearchpb.Document) error {
	idx.exitMu.RLock()
	defer idx.exitMu.RUnlock()

	if atomic.LoadInt32(&idx.exitFlag) == 1 {
		log.Error("Already Close.")
		return errors.New("Already Close.")
	}

	idx.mergeMu.RLock()
	defer idx.mergeMu.RUnlock()

	idx.persistMu.RLock()
	defer idx.persistMu.RUnlock()

	idx.segmentsMu.RLock()
	seg := idx.Segments[len(idx.Segments)-1]
	idx.segmentsMu.RUnlock()

	seg.indexDoc(doc)
	return nil
}

func (idx *Indexer) Search(
	query *xsearchpb.Query,
	start uint64,
	count int32) ([]*xsearchpb.Document, error) {

	idx.exitMu.RLock()
	defer idx.exitMu.RUnlock()

	if atomic.LoadInt32(&idx.exitFlag) == 1 {
		return nil, errors.New("Already Close.")
	}

	idx.mergeMu.RLock()
	defer idx.mergeMu.RUnlock()

	idx.persistMu.RLock()
	defer idx.persistMu.RUnlock()

	var (
		wg    sync.WaitGroup
		docsC = make(chan *xsearchpb.Document, 100*count)
		docs  = make([]*xsearchpb.Document, 0, 100*count)
	)

	idx.segmentsMu.RLock()
	for _, seg := range idx.Segments {
		wg.Add(1)
		go func(seg *segment) {
			if docs, err := seg.search(query, start, count); err == nil {
				for _, doc := range docs {
					docsC <- doc
				}
			}
			wg.Done()
		}(seg)
	}
	idx.segmentsMu.RUnlock()

	wg.Wait()
	close(docsC)
	for doc := range docsC {
		docs = append(docs, doc)
	}
	return docs, nil
}

func (idx *Indexer) RemoveDoc(docID string) {
}

func (idx *Indexer) persist() {
	idx.exitMu.RLock()
	defer idx.exitMu.RUnlock()

	if atomic.LoadInt32(&idx.exitFlag) == 1 {
		return
	}

	idx.mergeMu.RLock()
	defer idx.mergeMu.RUnlock()

	idx.persistMu.Lock()
	defer idx.persistMu.Unlock()

	idx.segmentsMu.RLock()
	oldSeg := idx.Segments[len(idx.Segments)-1]
	idx.segmentsMu.RUnlock()

	idx.wg.Add(1)
	go func(seg *segment) {
		if err := seg.persist(); err != nil {
			log.Error(err)
		}
		log.Infof("Type=Persist\tIndexer=%s\tSegment=%s", idx.opt.id, seg.ID)
		idx.wg.Done()
	}(oldSeg)

	if newSeg, err := idx.openSegment(mmap.MODE_CREATE, 0); err != nil {
		return
	} else {
		idx.segmentsMu.Lock()
		idx.Segments = append(idx.Segments, newSeg)
		idx.segmentsMu.Unlock()
	}
}

func (idx *Indexer) mergeSegments() {
	idx.exitMu.RLock()
	defer idx.exitMu.RUnlock()

	if atomic.LoadInt32(&idx.exitFlag) == 1 {
		return
	}

	idx.mergeMu.Lock()
	defer idx.mergeMu.Unlock()

	idx.persistMu.RLock()
	defer idx.persistMu.RUnlock()

	idx.segmentsMu.RLock()
	needMerge := idx.pickSegments()
	idx.segmentsMu.RUnlock()

	var newSegs []*segment
	for _, segs := range needMerge {
		if seg, err := idx.merge(segs); err != nil {
			log.Error(err)
		} else {
			newSegs = append(newSegs, seg)
		}
	}

	idx.segmentsMu.RLock()
	for _, seg := range idx.Segments {
		if seg.recycleFlag {
			seg.delete()
		} else {
			newSegs = append(newSegs, seg)
		}
	}
	idx.segmentsMu.RUnlock()

	idx.Segments = newSegs
}

func (idx *Indexer) pickSegments() (res [][]*segment) {
	var (
		numDocs   uint64
		needMerge []*segment
	)
	for i, seg := range idx.Segments {
		merge := false
		if atomic.LoadInt32(&seg.persistFlag) == 0 {
			continue
		}
		if seg.NumDocs > 0 && seg.NumDocs < idx.opt.mergeNum {
			seg.recycleFlag = true
			needMerge = append(needMerge, seg)
			numDocs += seg.NumDocs
		} else if seg.NumDocs == 0 {
			seg.recycleFlag = true
		}
		if numDocs >= idx.opt.mergeNum {
			merge = true
		}
		if i == len(idx.Segments)-1 {
			if len(needMerge) > 1 {
				merge = true
			} else if len(needMerge) == 1 {
				needMerge[0].recycleFlag = false
			}
		}
		if merge {
			res = append(res, needMerge)
			needMerge = make([]*segment, 0)
		}
	}
	return
}

func (idx *Indexer) merge(segs []*segment) (*segment, error) {
	k := len(segs)
	newSeg, err := idx.openSegment(mmap.MODE_CREATE, 0)
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
			idx.persist()
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

	idx.mergeMu.RLock()
	defer idx.mergeMu.RUnlock()

	idx.persistMu.RLock()
	defer idx.persistMu.RUnlock()

	close(idx.exitC)
	idx.wg.Wait()

	idx.segmentsMu.RLock()
	for _, seg := range idx.Segments {
		seg.persist()
		seg.close()
	}
	idx.segmentsMu.RUnlock()
}
