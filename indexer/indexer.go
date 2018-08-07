package indexer

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tddhit/xsearch/indexer/option"
	"github.com/tddhit/xsearch/indexer/segment"
	"github.com/tddhit/xsearch/xsearchpb"
)

var (
	defaultOptions = option.IndexerOptions{
		CommitNumDocs:      100000,
		CommitTimeInterval: 10 * time.Second,
		IndexDir:           "./",
		Sharding:           runtime.NumCPU(),
	}
)

type Indexer struct {
	opt        option.IndexerOptions
	mu         sync.RWMutex
	segs       [][]*segment.Segment
	removeDocs sync.Map
}

func New() *Indexer {
	idx := &Indexer{
		opt:  defaultOptions,
		segs: make([]*segment, defaultOptions.Sharding),
	}
	for i := range idx.segs {
		idx.segs[i] = make([]*segment.Segment, 1, 10)
		idx.segs[i][0] = idx.newSegment()
	}
	return idx
}

func (idx *Indexer) IndexDocument(doc *xsearchpb.Document) error {
	// TODO: Select an appropriate hash function
	i := doc.GetID() % idx.opt.Sharding

	idx.mu.RLock()
	segs := idx.segs[i]
	seg := segs[len(segs)-1]
	idx.mu.Unlock()

	if err := seg.IndexDocument(doc); err != nil {
		return err
	}
	numDocs := atomic.LoadUint64(&seg.NumDocs)
	if numDocs >= idx.opt.CommitNumDocs {
		if err := seg.Persist(); err != nil {
			log.Fatal(err)
			return err
		}
		newSeg := idx.newSegment()
		idx.mu.Lock()
		idx.segs[i] = append(idx.segs[i], newSeg)
		idx.mu.Unlock()
	}
}

func (idx *Indexer) RemoveDocument(docID uint64) {
	idx.removeDocs.Store(docID, struct{}{})
}

// TODO: Need internal DocID
func (idx *Indexer) UpdateDocument(doc *xsearchpb.Document) error {
	return nil
}

func (idx *Indexer) Search(query *xsearchpb.Query,
	start uint64, count int32) (docs []*xsearchpb.Document, err error) {

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
	for doc := range docsC {
		docs = append(docs, doc)
	}
	return
}

func (idx *Indexer) copy() ([][]*segment.Segment, int) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	count := 0
	segs := make([][]*segment.Segment, len(idx.segs))
	for i := range segs {
		segs[i] = make([]*segment.Segment, len(idx.segs[i]))
		for j := range segs[i] {
			segs[i][j] = idx.segs[i][j]
			count++
		}
	}
	return count
}
