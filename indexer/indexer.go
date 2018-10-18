package indexer

import (
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

	"github.com/tddhit/tools/log"
	"github.com/tddhit/tools/mmap"
	"github.com/tddhit/xsearch/indexer/option"
	"github.com/tddhit/xsearch/indexer/segment"
	xsearchpb "github.com/tddhit/xsearch/pb"
)

var (
	defaultOptions = option.IndexerOptions{
		CommitNumDocs: 100000,
		IndexDir:      "./",
		Sharding:      runtime.NumCPU(),
	}
)

type Indexer struct {
	opt            option.IndexerOptions
	mu             []sync.RWMutex
	segmentID      []int
	segs           [][]*segment.Segment
	removeDocs     sync.Map // TODO: replace with bitmap
	indexDocumentC []chan *xsearchpb.Document
	indexRspC      []chan error
	exitC          chan struct{}
	exitMu         sync.RWMutex
	exitFlag       int32
	persistWG      sync.WaitGroup
}

func New(opts ...option.IndexerOption) *Indexer {
	ops := defaultOptions
	for _, o := range opts {
		o(&ops)
	}
	idx := &Indexer{
		opt:            ops,
		mu:             make([]sync.RWMutex, ops.Sharding),
		segmentID:      make([]int, ops.Sharding),
		segs:           make([][]*segment.Segment, ops.Sharding),
		indexDocumentC: make([]chan *xsearchpb.Document, ops.Sharding),
		indexRspC:      make([]chan error, ops.Sharding),
		exitC:          make(chan struct{}),
	}
	idx.loadSegments()
	for i := 0; i < int(idx.opt.Sharding); i++ {
		idx.mu[i] = sync.RWMutex{}
		idx.indexDocumentC[i] = make(chan *xsearchpb.Document)
		idx.indexRspC[i] = make(chan error)
		idx.segs[i] = append(idx.segs[i], idx.createSegment(i))
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

func (idx *Indexer) createSegment(shardingID int) *segment.Segment {
	idx.mu[shardingID].Lock()
	vocabPath := fmt.Sprintf("%s/%d_%d_%d.vocab", idx.opt.IndexDir,
		idx.opt.IndexID, shardingID, idx.segmentID[shardingID])
	invertPath := fmt.Sprintf("%s/%d_%d_%d.invert", idx.opt.IndexDir,
		idx.opt.IndexID, shardingID, idx.segmentID[shardingID])
	idx.segmentID[shardingID]++
	idx.mu[shardingID].Unlock()

	return segment.New(vocabPath, invertPath, mmap.CREATE, mmap.RANDOM)
}

func (idx *Indexer) loadSegments() {
	id := make(map[int][]int, idx.opt.Sharding)
	filepath.Walk(idx.opt.IndexDir,
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".vocab") {
				name := info.Name()
				s := strings.Split(name[:len(name)-6], "_")
				if len(s) != 3 {
					log.Fatalf("Invalid vocab/invert filename:%s", name)
				}
				indexID, err := strconv.Atoi(s[0])
				if indexID != idx.opt.IndexID {
					log.Fatalf("Find a different segment, indexID is %d not %d",
						indexID, idx.opt.IndexID)
				}
				shardingID, err := strconv.Atoi(s[1])
				if shardingID > int(idx.opt.Sharding)-1 {
					log.Fatalf("Invalid shardingID (%d), sharding number is %d",
						shardingID, idx.opt.Sharding)
				}
				segmentID, err := strconv.Atoi(s[2])
				if err != nil {
					log.Fatal(err)
				}
				id[shardingID] = append(id[shardingID], segmentID)
			}
			return nil
		})
	for k, v := range id {
		sort.Ints(v)
		for i := range v {
			vocabPath := fmt.Sprintf("%s/%d_%d_%d.vocab", idx.opt.IndexDir,
				idx.opt.IndexID, k, v[i])
			invertPath := fmt.Sprintf("%s/%d_%d_%d.invert", idx.opt.IndexDir,
				idx.opt.IndexID, k, v[i])
			_, err := os.Stat(invertPath)
			if err != nil && os.IsNotExist(err) {
				log.Fatalf("%s is not exist.", invertPath)
			}
			idx.segs[k] = append(idx.segs[k],
				segment.New(vocabPath, invertPath, mmap.RDONLY, mmap.RANDOM),
			)
		}
		idx.segmentID[k] = len(v)
	}
}

func (idx *Indexer) indexLoop(shardingID int) {
	for {
		select {
		case doc := <-idx.indexDocumentC[shardingID]:
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

	i := doc.GetID() % uint64(idx.opt.Sharding)
	log.Infof("doc(%d) is indexed by segment(%d).", doc.GetID(), i)
	idx.indexDocumentC[i] <- doc
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
		if _, ok := idx.removeDocs.Load(doc.GetID()); ok {
			continue
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (idx *Indexer) copy() ([][]*segment.Segment, int32) {
	count := int32(0)
	segs := make([][]*segment.Segment, idx.opt.Sharding)
	for i := range segs {
		idx.mu[i].RLock()
		segs[i] = make([]*segment.Segment, len(idx.segs[i]))
		for j := range segs[i] {
			segs[i][j] = idx.segs[i][j]
			count++
		}
		idx.mu[i].RUnlock()
	}
	return segs, count
}

func (idx *Indexer) persist(shardingID int) {
	newSeg := idx.createSegment(shardingID)

	idx.mu[shardingID].Lock()
	segs := idx.segs[shardingID]
	oldSeg := segs[len(segs)-1]
	// Use idx.segs to avoid append allocation of new variables
	idx.segs[shardingID] = append(idx.segs[shardingID], newSeg)
	idx.mu[shardingID].Unlock()

	idx.persistWG.Add(1)
	go func(seg *segment.Segment) {
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
