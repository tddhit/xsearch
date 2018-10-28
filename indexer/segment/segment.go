package segment

import (
	"container/heap"
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
	xsearchpb "github.com/tddhit/xsearch/pb"
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
	invertList   map[string]*types.PostingList
	vocab        *bindex.BIndex
	invert       *mmap.MmapFile
	persist      int32
}

func New(vocabPath, invertPath string, mode, advise int) *Segment {
	vocab, err := bindex.New(vocabPath, mode, advise)
	if err != nil {
		log.Fatal(err)
	}
	invert, err := mmap.New(invertPath, maxMapSize, mode, advise)
	if err != nil {
		log.Fatal(err)
	}
	s := strings.Split(vocabPath, "/")
	return &Segment{
		name:       s[len(s)-1],
		docLength:  make(map[uint64]uint32),
		invertList: make(map[string]*types.PostingList),
		vocab:      vocab,
		invert:     invert,
	}
}

func (s *Segment) IndexDocument(doc *xsearchpb.Document) error {
	var (
		loc     uint32
		posting *types.Posting
	)
	for _, token := range doc.Tokens {
		//log.Error(len(doc.Tokens))
		term := token.GetTerm()
		if term == "" {
			continue
		}
		if _, ok := s.invertList[term]; !ok {
			s.invertList[term] = types.NewPostingList()
		}
		postingList := s.invertList[term]
		posting = postingList.Back()
		if posting == nil || posting.DocID != doc.ID {
			posting = &types.Posting{}
			posting.DocID = doc.ID
			posting.Freq = 1
			postingList.PushBack(posting)
		} else {
			posting.Freq++
		}
		//posting.Loc = append(posting.Loc, loc)
		loc++
	}
	s.NumDocs++
	return nil
}

func (s *Segment) Search(query *xsearchpb.Query,
	start uint64, count int32) (docs []*xsearchpb.Document, err error) {

	doc2BM25 := make(map[string]float32)
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

func (s *Segment) lookup(key []byte, doc2BM25 map[[16]byte]float32) error {
	value := s.vocab.Get(key)
	if value == nil {
		log.Error(s.name, errNotFoundKey, string(key))
		return errNotFoundKey
	}
	off := int64(binary.LittleEndian.Uint64(value))
	count, err := s.invert.Uint64At(off)
	if err != nil {
		return err
	}
	off += 8
	for i := uint64(0); i < count; i++ {
		bytes, err := s.invert.ReadAt(off, 16)
		if err != nil {
			return err
		}
		var docID [16]byte
		for i := 0; i < len(docID); i++ {
			docID[i] = bytes[i]
		}
		off += 16
		bits, err := s.invert.Uint32At(off)
		if err != nil {
			return err
		}
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

	log.Error(s.name, "persist start", len(s.invertList))
	var off int64
	for term, postingList := range s.invertList {
		b := pool.Get().([]byte)
		binary.LittleEndian.PutUint64(b, uint64(off))
		s.vocab.Put([]byte(term), b)
		pool.Put(b)

		s.invert.PutUint64At(off, uint64(postingList.Len()))
		off += 8
		for p := postingList.Front(); p != nil; p = p.Next() {
			s.invert.PutUint64At(off, p.DocID)
			off += 8
			idf := float32(math.Log2(float64(s.NumDocs)/float64(postingList.Len()) + 1))
			bm25 := idf * float32(p.Freq) * (BM25_K1 + 1) / (float32(p.Freq) + BM25_K1*(1-BM25_B+BM25_B*float32(s.docLength[p.DocID])/float32(s.avgDocLength)))
			bits := math.Float32bits(bm25)
			s.invert.PutUint32At(off, bits)
			off += 4
		}
	}
	log.Error(s.name, "persist end", len(s.invertList))
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
