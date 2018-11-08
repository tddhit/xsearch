package indexer

import (
	"bytes"
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
		CommitNumDocs: 100000,
		IndexDir:      "./",
		Sharding:      runtime.NumCPU(),
	}
)

type Indexer struct {
	opt        indexerOptions
	mu         []sync.RWMutex
	segmentID  int32
	segs       [][]*Segment
	removeDocs sync.Map // TODO: replace with bitmap
	indexDocC  []chan *xsearchpb.Document
	indexRspC  []chan error
	exitC      chan struct{}
	exitMu     sync.RWMutex
	exitFlag   int32
	persistWG  sync.WaitGroup
}

func New(opts ...IndexerOption) *Indexer {
	ops := defaultOptions
	for _, o := range opts {
		o(&ops)
	}
	idx := &Indexer{
		opt:       ops,
		mu:        make([]sync.RWMutex, ops.Sharding),
		segs:      make([][]*Segment, ops.Sharding),
		indexDocC: make([]chan *xsearchpb.Document, ops.Sharding),
		indexRspC: make([]chan error, ops.Sharding),
		exitC:     make(chan struct{}),
	}
	if err := os.MkdirAll(idx.opt.IndexDir, 0755); err != nil && !os.IsExist(err) {
		log.Error(err)
		return nil
	}
	idx.loadSegments()
	for i := 0; i < int(idx.opt.Sharding); i++ {
		idx.mu[i] = sync.RWMutex{}
		idx.indexDocC[i] = make(chan *xsearchpb.Document)
		idx.indexRspC[i] = make(chan error)
		idx.segs[i] = append(idx.segs[i], idx.createSegment())
		go idx.indexLoop(i)
	}
	if idx.opt.CommitTimeInterval > 0 {
		go idx.persistByTimer()
	}
	return idx
}

func (idx *Indexer) persistByTimer() {
	ticker := time.NewTicker(idx.opt.CommitTimeInterval)
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
}

func (idx *Indexer) createSegment() *Segment {
	vocabPath := fmt.Sprintf("%s/%d_%d.vocab", idx.opt.IndexDir,
		idx.opt.IndexID, idx.segmentID)
	invertPath := fmt.Sprintf("%s/%d_%d.invert", idx.opt.IndexDir,
		idx.opt.IndexID, idx.segmentID)
	atomic.AddInt32(&idx.segmentID, 1)

	return NewSegment(vocabPath, invertPath, mmap.CREATE, mmap.RANDOM)
}

func (idx *Indexer) loadSegments() {
	id := make(map[int][]int, idx.opt.Sharding)
	shardingID := 0
	filepath.Walk(idx.opt.IndexDir,
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".vocab") {
				name := info.Name()
				s := strings.Split(name[:len(name)-6], "_")
				if len(s) != 2 {
					log.Fatalf("Invalid vocab/invert filename:%s", name)
				}
				indexID := s[0]
				if indexID != idx.opt.IndexID {
					log.Fatalf("Find a different segment, indexID is %s not %s",
						indexID, idx.opt.IndexID)
				}
				segmentID, err := strconv.Atoi(s[1])
				if err != nil {
					log.Fatal(err)
				}
				id[shardingID%idx.opt.Sharding] = append(id[shardingID], segmentID)
				shardingID++
			}
			return nil
		})
	for k, v := range id {
		sort.Ints(v)
		for i := range v {
			vocabPath := fmt.Sprintf("%s/%s_%d.vocab", idx.opt.IndexDir,
				idx.opt.IndexID, v[i])
			invertPath := fmt.Sprintf("%s/%s_%d.invert", idx.opt.IndexDir,
				idx.opt.IndexID, v[i])
			_, err := os.Stat(invertPath)
			if err != nil && os.IsNotExist(err) {
				log.Fatalf("%s is not exist.", invertPath)
			}
			idx.segs[k] = append(idx.segs[k],
				NewSegment(vocabPath, invertPath, mmap.RDONLY, mmap.RANDOM),
			)
			if idx.segmentID < int32(v[i]) {
				idx.segmentID = int32(v[i])
			}
		}
	}
}

func (idx *Indexer) indexLoop(shardingID int) {
	for {
		select {
		case doc := <-idx.indexDocC[shardingID]:
			idx.indexRspC[shardingID] <- idx.indexOne(shardingID, doc)
		}
	}
}

func (idx *Indexer) indexOne(shardingID int, doc *xsearchpb.Document) error {
	idx.mu[shardingID].RLock()
	segs := idx.segs[shardingID]
	seg := segs[len(segs)-1]
	if err := seg.IndexDocument(doc); err != nil {
		idx.mu[shardingID].RUnlock()
		return err
	}
	idx.mu[shardingID].RUnlock()

	numDocs := atomic.LoadUint64(&seg.NumDocs)
	if idx.opt.CommitTimeInterval == 0 && numDocs >= idx.opt.CommitNumDocs {
		//log.Error(shardingID, doc.GetID())
		idx.persist(shardingID)
	}
	return nil
}

func (idx *Indexer) IndexDocument(doc *xsearchpb.Document) error {
	// TODO: Select an appropriate hash function
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
	i := util.Hash(id[:]) % uint64(idx.opt.Sharding)
	log.Infof("doc(%s) is indexed by segment(%d).", doc.GetID(), i)
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

	segs, n := idx.copy()
	var (
		wg    sync.WaitGroup
		docsC = make(chan *xsearchpb.Document, n*count)
		docs  = make([]*xsearchpb.Document, 0, n*count)
	)
	for i := range segs {
		wg.Add(1)
		go func(i int) {
			for j := range segs[i] {
				docs, err := segs[i][j].Search(query, start, count)
				if err == nil {
					for _, doc := range docs {
						docsC <- doc
					}
				}
			}
			wg.Done()
		}(i)
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

func (idx *Indexer) copy() ([][]*Segment, int32) {
	count := int32(0)
	segs := make([][]*Segment, idx.opt.Sharding)
	for i := range segs {
		idx.mu[i].RLock()
		segs[i] = make([]*Segment, len(idx.segs[i]))
		for j := range segs[i] {
			segs[i][j] = idx.segs[i][j]
			count++
		}
		idx.mu[i].RUnlock()
	}
	return segs, count
}

func (idx *Indexer) persist(shardingID int) {
	newSeg := idx.createSegment()

	idx.mu[shardingID].Lock()
	segs := idx.segs[shardingID]
	oldSeg := segs[len(segs)-1]
	// Use idx.segs to avoid append allocation of new variables
	idx.segs[shardingID] = append(idx.segs[shardingID], newSeg)
	idx.mu[shardingID].Unlock()

	idx.persistWG.Add(1)
	go func(seg *Segment) {
		if err := seg.Persist(); err != nil {
			log.Error(err)
		}
		idx.persistWG.Done()
	}(oldSeg)
}

func (idx *Indexer) persistAll() {
	idx.exitMu.RLock()
	defer idx.exitMu.RUnlock()

	if atomic.LoadInt32(&idx.exitFlag) == 1 {
		return
	}

	for i := 0; i < idx.opt.Sharding; i++ {
		idx.persist(i)
	}
}

func (idx *Indexer) mergeSegments() {
	var (
		numDocs   uint64
		needMerge []*Segment
	)
	segs, _ := idx.copy()
	for i := range segs {
		for j := range segs[i] {
			if atomic.LoadInt32(&segs[i][j].persist) == 1 {
				if segs[i][j].NumDocs < idx.opt.CommitNumDocs {
					if numDocs+segs[i][j].NumDocs < idx.opt.CommitNumDocs {
						numDocs += segs[i][j].NumDocs
						needMerge = append(needMerge, segs[i][j])
					} else {
						idx.merge(needMerge)
						numDocs = 0
						needMerge = needMerge[:0]
					}
				}
			}
		}
	}
}

type kv struct {
	key   []byte
	value []byte
}

func (idx *Indexer) merge(segs []*Segment) *Segment {
	newSeg := idx.createSegment()
	cursors := make([]*bindex.Cursor, len(segs))
	min, max := &kv{}, &kv{}
	for _, seg := range segs {
		c := seg.vocab.NewCursor()
		k, v := c.First()
		if bytes.Compare(k, min.key) < 0 {
			min = &kv{k, v}
		}
		if bytes.Compare(k, max.key) > 0 {
			max = &kv{k, v}
		}
	}
	input := func(i int) interface{} {
		if cursors[i] == nil {
			cursors[i] = segs[i].vocab.NewCursor()
			k, v := cursors[i].First()
			return &kv{
				key:   k,
				value: v,
			}
		}
		k, v := cursors[i].Next()
		return &kv{
			key:   k,
			value: v,
		}
	}
	output := func(i interface{}) {
		newSeg.vocab.Put(i.(*kv).key, i.(*kv).value)
	}
	compare := func(a, b interface{}) int {
		return bytes.Compare(a.(*kv).key, b.(*kv).value)
	}
	util.KMerge(len(segs), min, max, input, output, compare)
	return newSeg
}

func (idx *Indexer) Close() {
	idx.exitMu.RLock()
	defer idx.exitMu.RUnlock()

	if !atomic.CompareAndSwapInt32(&idx.exitFlag, 0, 1) {
		return
	}

	idx.persistWG.Wait()
	var wg sync.WaitGroup
	wg.Add(idx.opt.Sharding)
	for i := 0; i < idx.opt.Sharding; i++ {
		go func(i int) {
			idx.mu[i].Lock()
			segs := idx.segs[i]
			seg := segs[len(segs)-1]
			if err := seg.Persist(); err != nil {
				log.Error(err)
			}
			for j := range segs {
				log.Error(i, j, "close")
				segs[j].Close()
			}
			idx.mu[i].Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()
}
