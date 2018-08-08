package indexer

import (
	"bufio"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/indexer/option"
	"github.com/tddhit/xsearch/xsearchpb"
)

func TestSearch(t *testing.T) {
	log.Init("test/indexer.log", log.ERROR)
	indexer := New(option.WithIndexDir("./test/"))
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

func TestIndex(t *testing.T) {
	log.Init("test/indexer.log", log.ERROR)
	indexer := New(option.WithIndexDir("./test/"))
	docC := make(chan *xsearchpb.Document, 1000)
	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			for doc := range docC {
				if err := indexer.IndexDocument(doc); err != nil {
					log.Fatal(err)
				}
			}
			wg.Done()
		}()
	}

	file, err := os.Open("./test/seg.txt")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	i := uint64(0)
	for scanner.Scan() {
		text := scanner.Text()
		terms := strings.Split(text, " ")
		tokens := make([]*xsearchpb.Token, 0, 20)
		for _, term := range terms {
			tokens = append(tokens, &xsearchpb.Token{Term: term})
		}
		doc := &xsearchpb.Document{
			ID:     i,
			Tokens: tokens,
		}
		docC <- doc
		i++
	}
	close(docC)
	wg.Wait()
	indexer.Close()
}
