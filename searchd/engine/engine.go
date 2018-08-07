package engine

import (
	"errors"

	"github.com/tddhit/hunter/indexer"
	xsearchpb "github.com/tddhit/xsearch/xsearchpb"
)

var (
	errBufIsFull = errors.New("document buffer is full!")
)

type Engine struct {
	indexer   *indexer.Indexer
	documentC chan *xsearchpb.Document
}

func New() *Engine {
	e := &Engine{
		indexer:   indexer.New(),
		documentC: make(chan *xsearchpb.Document),
	}
	return e
}

func (e *Engine) AddDoc(doc *xsearchpb.Document) error {
	select {
	case e.documentC <- doc:
		return nil
	default:
		return errBufIsFull
	}
}

func (e *Engine) RemoveDoc(doc *xsearchpb.Document) error {
	select {
	case e.documentC <- doc:
		return nil
	default:
		return errBufIsFull
	}
}

func (e *Engine) Search(query *xsearchpb.Query, start uint64, count int32) ([]*xsearchpb.Document, error) {
	return e.indexer.Search(query, start, count)
}
