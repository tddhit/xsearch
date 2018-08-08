package segment

import (
	"container/heap"
	"container/list"
	"encoding/binary"
	"errors"
	"math"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/tddhit/bindex"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/tools/mmap"
	"github.com/tddhit/xsearch/internal/types"
	"github.com/tddhit/xsearch/xsearchpb"
)

const (
	TimeFormat         = "2006/01/02"
	maxMapSize         = 1 << 37
	BM25_K1    float32 = 1.2
	BM25_B     float32 = 0.75
)

var (
	pool           sync.Pool
	errNotFoundKey = errors.New("bindex not found key")
)

func init() {
	pool.New = func() interface{} {
		b := make([]byte, 8)
		return b
	}
}

type Segment struct {
	mu           sync.RWMutex
	NumDocs      uint64
	avgDocLength uint32
	name         string
	docLength    map[uint64]uint32
	invertList   map[string]*list.List
	vocab        *bindex.BIndex
	invert       *mmap.MmapFile
	persist      int32
}

func New(vocabPath, invertPath string, mode int) *Segment {
	vocab, err := bindex.New(vocabPath, mode)
	if err != nil {
		log.Fatal(err)
	}
	invert, err := mmap.New(invertPath, maxMapSize, mode)
	if err != nil {
		log.Fatal(err)
	}
	s := strings.Split(vocabPath, "/")
	return &Segment{
		name:       s[len(s)-1],
		docLength:  make(map[uint64]uint32),
		invertList: make(map[string]*list.List),
		vocab:      vocab,
		invert:     invert,
	}
}

func (s *Segment) IndexDocument(doc *xsearchpb.Document) error {
	var loc uint32
	for _, token := range doc.Tokens {
		term := token.GetTerm()
		var (
			lastestPosting *types.Posting
			posting        *types.Posting
		)
		if _, ok := s.invertList[term]; !ok {
			s.invertList[term] = list.New()
		}
		postingList := s.invertList[term]
		elem := postingList.Back()
		if elem != nil {
			if p, ok := elem.Value.(*types.Posting); ok {
				lastestPosting = p
			}
		}
		if lastestPosting != nil && lastestPosting.DocID == doc.ID {
			posting = lastestPosting
		} else {
			posting = &types.Posting{}
			postingList.PushBack(posting)
		}
		posting.DocID = doc.ID
		posting.Freq++
		posting.Loc = append(posting.Loc, loc)
		loc++
	}
	s.NumDocs++
	return nil
}

func (s *Segment) Search(query *xsearchpb.Query,
	start uint64, count int32) (docs []*xsearchpb.Document, err error) {

	doc2BM25 := make(map[uint64]float32)
	docHeap := &DocHeap{}
	heap.Init(docHeap)
	for _, token := range query.GetTokens() {
		s.lookup([]byte(token.GetTerm()), doc2BM25)
	}
	for docID, bm25 := range doc2BM25 {
		d := &xsearchpb.Document{
			ID:        docID,
			BM25Score: bm25,
		}
		heap.Push(docHeap, d)
	}
	docNum := docHeap.Len()
	var i uint64
	for count > 0 && docNum > 0 {
		doc := heap.Pop(docHeap).(*xsearchpb.Document)
		if i >= start {
			docs = append(docs, doc)
			count--
			docNum--
		}
		i++
	}
	return
}

func (s *Segment) lookup(key []byte, doc2BM25 map[uint64]float32) error {
	value := s.vocab.Get(key)
	if value == nil {
		log.Error(s.name, errNotFoundKey, string(key))
		return errNotFoundKey
	}
	off := int64(binary.LittleEndian.Uint64(value))
	count := s.invert.Uint64At(off)
	off += 8
	for i := uint64(0); i < count; i++ {
		docID := s.invert.Uint64At(off)
		off += 8
		bits := s.invert.Uint32At(off)
		bm25 := math.Float32frombits(bits)
		off += 4
		doc2BM25[docID] += bm25
	}
	return nil
}

func (s *Segment) Persist() error {
	if !atomic.CompareAndSwapInt32(&s.persist, 0, 1) {
		return errors.New("Already persist.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	var off int64
	for term, postingList := range s.invertList {
		b := pool.Get().([]byte)
		binary.LittleEndian.PutUint64(b, uint64(off))
		s.vocab.Put([]byte(term), b)
		pool.Put(b)

		s.invert.PutUint64At(off, uint64(postingList.Len()))
		off += 8
		for e := postingList.Front(); e != nil; e = e.Next() {
			if posting, ok := e.Value.(*types.Posting); ok {
				s.invert.PutUint64At(off, posting.DocID)
				off += 8
				idf := float32(math.Log2(float64(s.NumDocs)/float64(postingList.Len()) + 1))
				bm25 := idf * float32(posting.Freq) * (BM25_K1 + 1) / (float32(posting.Freq) + BM25_K1*(1-BM25_B+BM25_B*float32(s.docLength[posting.DocID])/float32(s.avgDocLength)))
				bits := math.Float32bits(bm25)
				s.invert.PutUint32At(off, bits)
				off += 4
			} else {
				log.Fatalf("convert fail:%#v\n", e)
			}
		}
	}
	s.invertList = nil
	s.docLength = nil
	return nil
}

func (s *Segment) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vocab != nil {
		s.vocab.Close()
		s.vocab = nil
	}
	if s.invert != nil {
		s.invert.Close()
		s.invert = nil
	}
}

type DocHeap []*xsearchpb.Document

func (h DocHeap) Len() int           { return len(h) }
func (h DocHeap) Less(i, j int) bool { return h[i].BM25Score > h[j].BM25Score }
func (h DocHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *DocHeap) Push(x interface{}) {
	*h = append(*h, x.(*xsearchpb.Document))
}

func (h *DocHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
