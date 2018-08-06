package engine

import (
	"errors"

	"github.com/tddhit/hunter/indexer"
)

var (
	errBufIsFull = errors.New("document buffer is full!")
)

type Engine struct {
	indexer   *indexer.Indexer
	documentC chan *pb.Document
}

func New() *Engine {
	e := &Engine{}
	return e
}

func (e *Engine) AddDoc(doc *pb.Document) error {
	select {
	case e.documentC <- doc:
		return nil
	default:
		return errBufIsFull
	}
}

func (e *Engine) RemoveDoc(doc *pb.Document) error {
}

func (e *Engine) Search(query *pb.Query) {

}
