package segment

import (
	"bytes"
	"container/heap"
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/tddhit/bindex"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/internal/types"
	"github.com/tddhit/xsearch/xsearchpb"
)

const (
	TimeFormat                 = "2006/01/02"
	maxMapSize                 = 1 << 37
	commitDocsNumber           = 100000           // generate a segment per 100k docs
	commitTimeInterval         = 10 * time.Second // generate a segment per 10s
	BM25_K1            float32 = 1.2
	BM25_B             float32 = 0.75
)

var (
	pool           sync.Pool
	errNotFoundKey = errors.New("bindex not found key")
)

func init() {
	pool.New = func() interface{} {
		buf := make([]byte, 8)
		return buf
	}
}

type Segment struct {
	NumDocs      uint64
	avgDocLength uint32
	docLength    map[uint64]uint32
	invertList   map[string]*list.List
	vocab        *bindex.BIndex
	invertFile   *os.File
	invertRef    []byte
}

func New(vocabPath, invertPath string, readOnly bool) *Segment {
	vocab, err := bindex.New(vocabPath, readOnly)
	if err != nil {
		log.Fatal(err)
	}
	var flag int
	if readOnly {
		flag = os.O_RDONLY
	} else {
		flag = os.O_CREATE | os.O_RDWR | os.O_APPEND
	}
	invertFile, err := os.OpenFile(invertPath, flag, 0644)
	if err != nil {
		log.Fatal(err)
	}
	invertRef, err := mmap(invertFile)
	if err != nil {
		log.Fatal(err)
	}
	return &Segment{
		docLength:  make(map[uint64]uint32),
		invertList: make(map[string]*list.List),
		vocab:      vocab,
		invertFile: invertFile,
		invertRef:  invertRef,
	}
}

func (s *Segment) Index(doc *xsearchpb.Document) {
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
		log.Error(errNotFoundKey, string(key))
		return errNotFoundKey
	}
	loc := binary.LittleEndian.Uint64(value)
	count := binary.LittleEndian.Uint64(s.invertRef[loc : loc+8 : loc+8])
	log.Debug("count:", count)
	loc += 8
	for i := uint64(0); i < count; i++ {
		docID := binary.LittleEndian.Uint64(s.invertRef[loc : loc+8 : loc+8])
		loc += 8
		bits := binary.LittleEndian.Uint32(s.invertRef[loc : loc+4 : loc+4])
		bm25 := math.Float32frombits(bits)
		loc += 4
		doc2BM25[docID] += bm25
	}
	return nil
}

func (s *Segment) munmap() error {
	if s.invertRef == nil {
		return nil
	}
	if err := syscall.Munmap(s.invertRef); err != nil {
		return err
	}
	return nil
}

func (s *Segment) Dump() {
	// init file
	vocabPath := fmt.Sprintf("%s/%d.vocab", s.opt.IndexDir, rand.Int())
	vocabTmpPath := fmt.Sprintf("%s/%d.vocab.tmp", s.opt.IndexDir, rand.Int())
	vocab, err := bindex.New(vocabPath, false)
	if err != nil {
		log.Fatal(err)
	}
	invertPath := fmt.Sprintf("%s/%d.invert.tmp", s.opt.IndexDir, rand.Int())
	invert, err := os.OpenFile(invertPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// dump vocab/invert
	var loc uint64
	for term, postingList := range s.invertList {
		var buf bytes.Buffer
		err := binary.Write(&buf, binary.LittleEndian, uint64(postingList.Len()))
		if err != nil {
			log.Fatal(err)
		}
		for e := postingList.Front(); e != nil; e = e.Next() {
			if posting, ok := e.Value.(*types.Posting); ok {
				err := binary.Write(&buf, binary.LittleEndian, posting.DocID)
				if err != nil {
					log.Fatal(err)
				}
				idf := float32(math.Log2(float64(s.NumDocs)/float64(postingList.Len()) + 1))
				bm25 := idf * float32(posting.Freq) * (BM25_K1 + 1) / (float32(posting.Freq) + BM25_K1*(1-BM25_B+BM25_B*float32(s.docLength[posting.DocID])/float32(s.avgDocLength)))
				err = binary.Write(&buf, binary.LittleEndian, bm25)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatalf("convert fail:%#v\n", e)
			}
		}
		n, err := invert.Write(buf.Bytes())
		if n != buf.Len() || err != nil {
			log.Fatalf("dump fail:n=%d,len=%d,err=%s\n", n, buf.Len(), err)
		}
		b := pool.Get().([]byte)
		binary.LittleEndian.PutUint64(b, loc)
		vocab.Put([]byte(term), b)
		loc += uint64(n)
	}
	vocab.Close()
	invert.Close()
	os.Rename(vocabPath, s.opt.VocabPath)
	os.Rename(invertPath, s.opt.InvertPath)
}

func (s *Segment) Reset() error {
	s.NumDocs = 0
	s.AvgDocLength = 0
	s.InvertList = make(map[string]*list.List)
	if s.vocab != nil {
		s.vocab.Close()
		s.vocab = nil
	}
	if s.invertRef != nil {
		if err := syscall.Munmap(s.invertRef); err != nil {
			return err
		}
		s.invertRef = nil
	}
	if s.invertFile != nil {
		s.invertFile.Sync()
		s.invertFile.Close()
		s.invertFile = nil
	}
	return nil
}

func (s *Segment) Close() {
	s.Reset()
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

func mmap(file *os.File) ([]byte, error) {
	buf, err := syscall.Mmap(int(file.Fd()), 0, maxMapSize,
		syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	if _, _, err := syscall.Syscall(syscall.SYS_MADVISE,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)),
		uintptr(syscall.MADV_RANDOM)); err != 0 {

		return nil, err
	}
	return buf, nil
}
