package indexer

import (
	"bytes"
	"container/heap"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/satori/go.uuid"

	"github.com/tddhit/bindex"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/tools/mmap"
	"github.com/tddhit/xsearch/internal/types"
	"github.com/tddhit/xsearch/pb"
)

const (
	maxMapSize         = 1 << 30 // 1G
	bm25_k1    float32 = 1.2
	bm25_b     float32 = 0.75
	version    uint32  = 1
	magic      uint32  = 0xC9D12F01
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

type segment struct {
	sync.Mutex
	ID          string
	CreateTime  int64
	NumDocs     uint64
	totalLength uint64
	avgLength   uint32
	docsLength  map[string]uint32
	invertList  map[string]*types.PostingList
	vocab       *bindex.BIndex
	invert      *mmap.MmapFile
	vocabPath   string
	invertPath  string
	persistFlag int32
	recycleFlag bool
}

func newSegment(vocabPath, invertPath string, mode int) (*segment, error) {
	vocab, err := bindex.New(vocabPath, mode, mmap.ADVISE_RANDOM)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	invert, err := mmap.New(invertPath, maxMapSize, mode, mmap.ADVISE_RANDOM)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var (
		numDocs    uint64
		createTime int64 = time.Now().UnixNano()
	)
	if mode == mmap.MODE_RDONLY {
		if num, ctime, err := validate(invert); err != nil {
			log.Error(err)
			return nil, err
		} else {
			numDocs = num
			createTime = int64(ctime)
		}
	}
	s := strings.Split(vocabPath[:len(vocabPath)-6], "/")
	return &segment{
		ID:         s[len(s)-1],
		NumDocs:    numDocs,
		CreateTime: createTime,
		docsLength: make(map[string]uint32),
		invertList: make(map[string]*types.PostingList),
		vocab:      vocab,
		invert:     invert,
		vocabPath:  vocabPath,
		invertPath: invertPath,
	}, nil
}

func validate(invert *mmap.MmapFile) (numDocs, ctime uint64, err error) {
	log.Trace(2, invert.Size())
	var tailer int64 = 40
	if invert.Size() < tailer {
		return 0, 0, fmt.Errorf("Invalid invert size: %d", invert.Size())
	}
	buf := (*[maxMapSize]byte)(invert.Buf(0))
	checksum := md5.Sum((*buf)[:invert.Size()-tailer])
	c, err := invert.ReadAt(invert.Size()-tailer, 16)
	if err != nil {
		log.Error(err)
		return
	}
	log.Errorf("check checksum fail:%x,%x", c, checksum[:])
	if bytes.Compare(c, checksum[:]) != 0 {
		err = fmt.Errorf("check checksum fail:%x,%x", c, checksum[:])
		log.Error(err)
		return
	}
	tailer -= 16
	m, err := invert.Uint32At(invert.Size() - tailer)
	if err != nil {
		log.Error(err)
		return
	}
	if m != magic {
		err = fmt.Errorf("check magic fail:%x", m)
		log.Error(err)
		return
	}
	tailer -= 4
	v, err := invert.Uint32At(invert.Size() - tailer)
	if err != nil {
		log.Error(err)
		return
	}
	if v != version {
		err = fmt.Errorf("check version fail:%d", v)
		log.Error(err)
		return
	}
	tailer -= 4
	numDocs, err = invert.Uint64At(invert.Size() - tailer)
	if err != nil {
		log.Error(err)
		return
	}
	tailer -= 8
	ctime, err = invert.Uint64At(invert.Size() - tailer)
	if err != nil {
		log.Error(err)
		return
	}
	tailer -= 8
	return
}

func (s *segment) indexDoc(doc *xsearchpb.Document) {
	s.Lock()
	defer s.Unlock()

	var (
		loc     uint32
		posting *types.Posting
	)
	for _, token := range doc.Tokens {
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
		loc++
	}
	docLength := len(doc.Tokens)
	s.totalLength += uint64(docLength)
	s.docsLength[doc.ID] = uint32(docLength)
	s.NumDocs++
}

func (s *segment) search(
	query *xsearchpb.Query,
	start uint64,
	count int32) (docs []*xsearchpb.Document, err error) {

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

func (s *segment) lookup(key []byte, doc2BM25 map[string]float32) error {
	value := s.vocab.Get(key)
	if value == nil {
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
		docID, err := uuid.FromBytes(bytes)
		if err != nil {
			log.Error(err)
			return err
		}
		off += 16
		bits, err := s.invert.Uint32At(off)
		if err != nil {
			return err
		}
		bm25 := math.Float32frombits(bits)
		off += 4
		doc2BM25[docID.String()] += bm25
	}
	return nil
}

func (s *segment) persist() error {
	s.Lock()
	defer s.Unlock()

	if !atomic.CompareAndSwapInt32(&s.persistFlag, 0, 1) {
		return fmt.Errorf("segment(%s) already persist.", s.ID)
	}
	if s.NumDocs == 0 {
		return nil
	}
	var off int64
	s.avgLength = uint32(s.totalLength / s.NumDocs)
	for term, postingList := range s.invertList {
		//b := pool.Get().([]byte)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(off))
		s.vocab.Put([]byte(term), b)
		//pool.Put(b)

		s.invert.PutUint64At(off, uint64(postingList.Len()))
		off += 8
		for p := postingList.Front(); p != nil; p = p.Next() {
			id, err := uuid.FromString(p.DocID)
			if err != nil {
				log.Error(err)
				return err
			}
			s.invert.WriteAt(id[:], off)
			off += 16
			idf := float32(math.Log2(float64(s.NumDocs)/float64(postingList.Len()) + 1))
			bm25 := idf * float32(p.Freq) * (bm25_k1 + 1) / (float32(p.Freq) + bm25_k1*(1-bm25_b+bm25_b*float32(s.docsLength[id.String()])/float32(s.avgLength)))
			bits := math.Float32bits(bm25)
			s.invert.PutUint32At(off, bits)
			off += 4
		}
	}
	buf := (*[maxMapSize]byte)(s.invert.Buf(0))
	checksum := md5.Sum((*buf)[:off])
	s.invert.WriteAt(checksum[:], off)
	log.Tracef(2, "checksum:%x", checksum[:])
	off += 16
	s.invert.PutUint32At(off, magic)
	off += 4
	s.invert.PutUint32At(off, version)
	off += 4
	s.invert.PutUint64At(off, s.NumDocs)
	off += 8
	s.invert.PutUint64At(off, uint64(s.CreateTime))
	off += 8
	s.invertList = nil
	s.docsLength = nil
	log.Trace(2, "compact start", off)
	if err := s.invert.Compact(); err != nil {
		log.Error(err)
		return err
	}
	log.Trace(2, "compact end", s.invert.Size())
	return nil
}

func (s *segment) delete() {
	s.close()
	os.Remove(s.vocabPath)
	os.Remove(s.invertPath)
}

func (s *segment) close() {
	s.Lock()
	defer s.Unlock()

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
