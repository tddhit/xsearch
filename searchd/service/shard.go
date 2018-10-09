package service

import (
	"sync"

	"github.com/tddhit/hunter/indexer"
)

type shard struct {
	dataDir   string
	namespace string
	id        int
	indexer   *indexer.Indexer
	blog      *binLog
	wg        sync.WaitGroup
}

func newShard(dataDir, namespace string, id int) *shard {
	return &shard{
		dataDir:   dataDir,
		namespace: namespace,
		id:        id,
		indexer:   indexer.New(),
		blog:      newBinLog(dataDir, fmt.Sprintf("%s.%d", namespace, id)),
	}
}

func (s *shard) addDoc(doc *xsearchpb.Document) {
	s.indexer.IndexDocument(doc)
	s.blog.WriteLog(&pb.LogEntry{
		Action: pb.ADD,
		Doc:    doc,
	})
}

func (s *shard) removeDoc(docID int) {
	s.indexer.RemoveDocument(docID)
	s.blog.WriteLog(&pb.LogEntry{
		Action: pb.REMOVE,
		DocID:  docID,
	})
}

func (s *shard) search(
	query *xsearchpb.Query,
	start uint64,
	count int32) ([]*xsearchpb.Document, error) {

	return s.indexer.Search(query, start, count)
}
