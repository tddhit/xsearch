package service

import (
	"context"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/lunny/log"
	diskqueuepb "github.com/tddhit/diskqueue/pb"

	"github.com/tddhit/xsearch/indexer"
	indexerOpt "github.com/tddhit/xsearch/indexer/option"
	searchdpb "github.com/tddhit/xsearch/searchd/pb"
)

type shard struct {
	id       string
	indexer  *indexer.Indexer
	dqClient diskqueuepb.DiskqueueGrpcClient
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func newShard(id, dir string, c diskqueuepb.DiskqueueGrpcClient) *shard {
	s := &shard{
		id:       id,
		indexer:  indexer.New(indexerOpt.WithIndexDir(dir)),
		dqClient: c,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.wg.Add(1)
	go func() {
		s.indexLoop()
		s.wg.Done()
	}()
	return s
}

func (s *shard) indexLoop() {
	for {
		rsp, err := s.dqClient.Pop(s.ctx, &diskqueuepb.PopRequest{
			Topic: s.id,
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
