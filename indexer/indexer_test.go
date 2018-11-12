package indexer

import (
	"bufio"
	"encoding/binary"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"
	"unsafe"

	uuid "github.com/satori/go.uuid"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/tools/mmap"
	"github.com/tddhit/xsearch/pb"
	"github.com/wangbin/jiebago"
)

func TestSearch(t *testing.T) {
	log.Init("test/indexer.log", log.ERROR)
	indexer, _ := New(WithDir("./test/"))
	tokens := []*xsearchpb.Token{
		&xsearchpb.Token{Term: "Linux"},
		&xsearchpb.Token{Term: "Go"},
	}
	docs, err := indexer.Search(&xsearchpb.Query{
		Tokens: tokens,
	}, 0, 20)
	if err != nil {
		log.Fatal(err)
	}
	for _, doc := range docs {
		log.Info(doc.GetID())
	}
}

const (
	docSize = 200
	numDocs = 2000000
)

func TestIndex(t *testing.T) {

	go func() {
		if err := http.ListenAndServe(":12345", nil); err != nil {
			log.Fatal(err)
		}
	}()

	log.Init("test/indexer.log", log.ERROR)
	sharding := runtime.NumCPU()
	docs := make([]*xsearchpb.Document, sharding)
	for i := range docs {
		docs[i] = &xsearchpb.Document{
			Tokens: make([]*xsearchpb.Token, docSize),
		}
	}
	file, err := os.Open("./test/test.txt")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan() && i < docSize; i++ {
		text := scanner.Text()
		for j := range docs {
			docs[j].Tokens[i] = &xsearchpb.Token{
				Term: text,
			}
		}
	}
	indexer, _ := New(WithDir("./test/"))
	var wg sync.WaitGroup
	for i := 0; i < sharding; i++ {
		wg.Add(1)
		go func(shardingID int) {
			for i := 0; i < numDocs/sharding; i++ {
				if err := indexer.IndexDocument(docs[shardingID]); err != nil {
					log.Fatal(err)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	indexer.Close()
}

func TestReadMmap(t *testing.T) {
	file, err := os.Open("test/0_0_0.invert")
	if err != nil {
		log.Fatal(err)
	}
	buf, err := syscall.Mmap(int(file.Fd()), 0, 1<<30,
		syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		log.Fatal(err)
	}
	if _, _, err := syscall.Syscall(syscall.SYS_MADVISE,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)),
		uintptr(syscall.MADV_RANDOM)); err != 0 {

		log.Fatal(err)
	}
	info, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Start")
	var (
		loc   = 0
		count uint64
	)
	for {
		if loc >= int(info.Size()) {
			break
		}
		count = binary.LittleEndian.Uint64(buf[loc : loc+8 : loc+8])
		loc += 8
		log.Info(count)
		for i := uint64(0); i < count; i++ {
			binary.LittleEndian.Uint64(buf[loc : loc+8 : loc+8])
			loc += 8
			bits := binary.LittleEndian.Uint32(buf[loc : loc+4 : loc+4])
			math.Float32frombits(bits)
			loc += 4
		}
	}
	log.Info("End")
}

func TestReadFile(t *testing.T) {
	file, err := os.Open("test/0_0_0.invert")
	if err != nil {
		log.Fatal(err)
	}
	info, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Start")
	var (
		loc   int64 = 0
		count uint64
	)
	buf := make([]byte, 8)
	buf2 := make([]byte, 4)
	for {
		if loc >= info.Size() {
			break
		}
		n, err := file.ReadAt(buf, loc)
		if err != nil || n != cap(buf) {
			log.Fatal(err)
		}
		count = binary.LittleEndian.Uint64(buf)
		loc += 8
		log.Info(count)
		for i := uint64(0); i < count; i++ {
			n, err := file.ReadAt(buf, loc)
			if err != nil || n != cap(buf) {
				log.Fatal(err)
			}
			binary.LittleEndian.Uint64(buf)
			loc += 8
			n, err = file.ReadAt(buf2, loc)
			if err != nil || n != cap(buf2) {
				log.Fatal(err)
			}
			bits := binary.LittleEndian.Uint32(buf2)
			math.Float32frombits(bits)
			loc += 4
		}
	}
	log.Info("End")
}

func TestWriteMmap(t *testing.T) {
	file, err := os.OpenFile("test/1_0_0.invert", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = syscall.Ftruncate(int(file.Fd()), 1<<30)
	if err != nil {
		log.Fatal(err)
	}
	buf, err := syscall.Mmap(int(file.Fd()), 0, 1<<30,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatal(err)
	}
	if _, _, err := syscall.Syscall(syscall.SYS_MADVISE,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)),
		uintptr(syscall.MADV_RANDOM)); err != 0 {

		log.Fatal(err)
	}
	log.Info("Start")
	var (
		loc  = 0
		size = 64
		buf2 = make([]byte, size)
	)
	buf2[0] = 1
	for {
		if loc >= 1<<30 {
			break
		}
		copy(buf[loc:loc+size:loc+size], buf2)
		//binary.LittleEndian.PutUint64(buf[loc:loc+8:loc+8], 1)
		loc += size
	}
	log.Info(loc)
	file.Sync()
	log.Info("End")
}

func TestWriteFile(t *testing.T) {
	file, err := os.OpenFile("test/1_0_0.invert", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Start")
	var (
		loc  int64 = 0
		size       = 64
		buf        = make([]byte, size)
	)
	buf[0] = 1
	for {
		if loc >= 1<<30 {
			break
		}
		file.WriteAt(buf, loc)
		loc += int64(size)
	}
	log.Info(loc)
	log.Info("End")
}

func TestWriteMmap2(t *testing.T) {
	file, err := mmap.New("test/1_0_0.invert", 1<<30, mmap.MODE_CREATE, mmap.ADVISE_RANDOM)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Start")
	var (
		loc  = 0
		size = 64
		buf2 = make([]byte, size)
	)
	buf2[0] = 1
	for {
		if loc >= 1<<30 {
			break
		}
		file.WriteAt(buf2, int64(loc))
		//binary.LittleEndian.PutUint64(buf[loc:loc+8:loc+8], 1)
		loc += size
	}
	log.Info(loc)
	file.Sync()
	log.Info("End")
}

func TestMergeSegments(t *testing.T) {
	log.Init("", log.INFO)
	idx, _ := New(
		WithDir("./test"),
		WithMergeInterval(10*time.Second),
		WithPersistInterval(2*time.Second),
		WithShardNum(1),
	)
	segmenter := &jiebago.Segmenter{}
	if err := segmenter.LoadDictionary("../cmd/xsearch/dict/segment.dict"); err != nil {
		log.Fatal(err)
	}
	docs := []string{
		"我是一个程序员",
		"我是一个linux程序员",
		"我是一个golang程序员",
		"我是一个算法工程师",
		"我是一个服务端工程师",
		"我是一个做深度学习算法的工程师",
	}
	for i, content := range docs {
		id, _ := uuid.NewV4()
		doc := &xsearchpb.Document{ID: id.String()}
		for term := range segmenter.Cut(content, true) {
			doc.Tokens = append(
				doc.Tokens,
				&xsearchpb.Token{Term: term},
			)
		}
		if err := idx.IndexDocument(doc); err != nil {
			log.Error(err)
		}
		log.Infof("index doc id=%s content=%s", doc.ID, docs[i])
		time.Sleep(time.Second)
	}
	time.Sleep(2 * time.Second)
	log.Info("----------------")
	res, err := idx.Search(&xsearchpb.Query{
		Tokens: []*xsearchpb.Token{
			{Term: "程序员"},
		},
	}, 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	for _, doc := range res {
		log.Info(doc.ID, doc.BM25Score)
	}
	time.Sleep(10 * time.Second)
	log.Info("----------------")
	res, err = idx.Search(&xsearchpb.Query{
		Tokens: []*xsearchpb.Token{
			{Term: "程序员"},
		},
	}, 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	for _, doc := range res {
		log.Info(doc.ID, doc.BM25Score)
	}
	log.Info("close")
	go func() {
		if err := http.ListenAndServe(":12345", nil); err != nil {
			log.Fatal(err)
		}
	}()
	idx.Close()
}
