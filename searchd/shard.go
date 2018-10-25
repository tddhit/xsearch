package searchd

import (
	"context"
	"sync"

	"github.com/gogo/protobuf/proto"

	"github.com/tddhit/diskqueue/pb"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/indexer"
	indexerOpt "github.com/tddhit/xsearch/indexer/option"
	"github.com/tddhit/xsearch/searchd/pb"
)

type shard struct {
	id      string
	indexer *indexer.Indexer
	diskq   diskqueuepb.DiskqueueGrpcClient
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func newShard(id, dir, channel string, c diskqueuepb.DiskqueueGrpcClient) *shard {
	s := &shard{
		id:      id,
		indexer: indexer.New(indexerOpt.WithIndexDir(dir)),
		diskq:   c,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.wg.Add(1)
	go func() {
		s.indexLoop(channel)
		s.wg.Done()
	}()
	return s
}

func (s *shard) indexLoop(channel string) {
	for {
		rsp, err := s.diskq.Pop(s.ctx, &diskqueuepb.PopReq{
			Topic:   s.id,
			Channel: channel,
		})
		if err != nil {
			log.Error(err)
			break
		}
		msg := rsp.GetMessage()
		if msg == nil {
			continue
		}
		data := msg.GetData()
		cmd := &searchdpb.Command{}
		if err := proto.Unmarshal(data, cmd); err != nil {
			log.Error(err)
			continue
		}
		switch cmd.Type {
		case searchdpb.Command_INDEX:
			if doc, ok := cmd.DocOneof.(*searchdpb.Command_Doc); ok {
				if err := s.indexer.IndexDocument(doc.Doc); err != nil {
					log.Error(err)
				}
			}
		case searchdpb.Command_REMOVE:
			if docID, ok := cmd.DocOneof.(*searchdpb.Command_DocID); ok {
				s.indexer.RemoveDocument(docID.DocID)
			}
		}
	}
}

func (s *shard) close() {
	s.cancel()
	s.wg.Wait()
	s.indexer.Close()
}
