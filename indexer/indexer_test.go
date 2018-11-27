package indexer

import (
	"bufio"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/wangbin/jiebago"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
)

var (
	data      []string
	smallData []string
	segmenter *jiebago.Segmenter
)

func loadData(path string) (data []string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 102400)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(buf, 102400)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return
}

func init() {
	os.Remove("./test/test.log")
	//log.Init("./test/test.log", log.TRACE)
	log.Init("", log.TRACE)
	data = loadData("./test/test.txt")
	smallData = loadData("./test/small.txt")
	segmenter = &jiebago.Segmenter{}
	if err := segmenter.LoadDictionary("./test/segment.dict"); err != nil {
		log.Fatal(err)
	}
}

func index(data []string, idx *Indexer) {
	for i, content := range data {
		id, _ := uuid.NewV4()
		doc := &xsearchpb.Document{
			ID:      id.String(),
			Content: content,
		}
		for term := range segmenter.Cut(content, true) {
			doc.Tokens = append(
				doc.Tokens,
				&xsearchpb.Token{Term: term},
			)
		}
		if err := idx.IndexDoc(doc); err != nil {
			log.Error(err)
		}
		log.Infof("index %d doc id=%s", i, doc.ID)
	}
}

func TestIndex(t *testing.T) {
	os.RemoveAll("./test/index1")
	idx, err := New(
		WithID("index1"),
		WithDir("./test/index1"),
		WithMergeInterval(0),
		WithMergeNum(0),
		WithPersistInterval(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	index(data, idx)
	idx.Close()
}

func TestMerge(t *testing.T) {
	os.RemoveAll("./test/merge1")
	idx, err := New(
		WithID("merge1"),
		WithDir("./test/merge1"),
		WithMergeInterval(30*time.Second),
		WithMergeNum(100000),
		WithPersistInterval(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	index(data, idx)
	idx.Close()
}

func TestParallelIndex(t *testing.T) {
	os.RemoveAll("./test/index2")
	idx, err := New(
		WithID("index2"),
		WithDir("./test/index2"),
		WithMergeInterval(0),
		WithMergeNum(0),
		WithPersistInterval(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			index(data, idx)
			wg.Done()
		}()
	}
	wg.Wait()
	idx.Close()
}

func TestLoad(t *testing.T) {
	idx, err := New(
		WithID("index0"),
		WithDir("./test/index0"),
		WithMergeInterval(0),
		WithMergeNum(0),
		WithPersistInterval(0),
	)
	if err != nil {
		log.Fatal(err)
	}
	idx.Close()
}

func TestSmallIndex(t *testing.T) {
	os.RemoveAll("./test/index3")
	idx, err := New(
		WithID("index3"),
		WithDir("./test/index3"),
		WithMergeInterval(0),
		WithMergeNum(0),
		WithPersistInterval(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	index(smallData, idx)
	idx.Close()
}

func TestSearch(t *testing.T) {
	os.RemoveAll("./test/index5")
	idx, err := New(
		WithID("index5"),
		WithDir("./test/index5"),
		WithMergeInterval(0),
		WithMergeNum(0),
		WithPersistInterval(1*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	id, _ := uuid.NewV4()
	err = idx.IndexDoc(&xsearchpb.Document{
		ID: id.String(),
		Tokens: []*xsearchpb.Token{
			{Term: "我"},
			{Term: "是"},
			{Term: "一个"},
			{Term: "程序员"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	docs, err := idx.Search(&xsearchpb.Query{
		Tokens: []*xsearchpb.Token{
			{Term: "程序员"},
		},
	}, 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(len(docs))
	for _, doc := range docs {
		log.Info(doc)
	}
	idx.Close()
}

func TestMergeSearch(t *testing.T) {
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
	os.RemoveAll("./test/index6")
	idx, err := New(
		WithID("index6"),
		WithDir("./test/index6"),
		WithMergeInterval(5*time.Second),
		WithMergeNum(10),
		WithPersistInterval(1*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	id, _ := uuid.NewV4()
	err = idx.IndexDoc(&xsearchpb.Document{
		ID: id.String(),
		Tokens: []*xsearchpb.Token{
			{Term: "我"},
			{Term: "是"},
			{Term: "一个"},
			{Term: "程序员"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(10 * time.Second)
	docs, err := idx.Search(&xsearchpb.Query{
		Tokens: []*xsearchpb.Token{
			{Term: "程序员"},
		},
	}, 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(len(docs))
	for _, doc := range docs {
		log.Info(doc)
	}
	idx.Close()
}
