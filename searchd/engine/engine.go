package engine

import (
	"errors"

	"github.com/tddhit/xsearch/indexer"
	xsearchpb "github.com/tddhit/xsearch/xsearchpb"
)

var (
	errBufIsFull = errors.New("document buffer is full!")
)

type Engine struct {
	indexer *indexer.Indexer
}

func New() *Engine {
	e := &Engine{
		indexer: indexer.New(),
	}
	return e
}

func (e *Engine) AddDoc(doc *xsearchpb.Document) error {
	return e.indexer.IndexDocument(doc)
}

func (e *Engine) RemoveDoc(doc *xsearchpb.Document) {
	e.indexer.RemoveDocument(doc.GetID())
}

func (e *Engine) Search(query *xsearchpb.Query, start uint64, count int32) ([]*xsearchpb.Document, error) {
	return e.indexer.Search(query, start, count)
}
