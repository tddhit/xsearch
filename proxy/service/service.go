package service

import (
	"container/heap"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/tddhit/box/transport"
	"github.com/tddhit/tools/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	metadpb "github.com/tddhit/xsearch/metad/pb"
	xsearchpb "github.com/tddhit/xsearch/pb"
	proxypb "github.com/tddhit/xsearch/proxy/pb"
	searchdpb "github.com/tddhit/xsearch/searchd/pb"
)

type service struct {
	metad    metadpb.MetadGrpcClient
	resource *resource
	exitC    chan struct{}
}

func NewService(metadAddr, namespaces string) (*service, error) {
	conn, err := transport.Dial(metadAddr)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	s := &service{
		metad:    metadpb.NewMetadGrpcClient(conn),
		resource: newResource(),
		exitC:    make(chan struct{}),
	}
	nss := strings.Split(namespaces, ",")
	for _, ns := range nss {
		if err := s.registerClient(ns); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *service) registerClient(namespace string) error {
	stream, err := s.metad.RegisterClient(context.Background())
	if err != nil {
		return err
	}
	go s.watchShardTable(stream)
	go s.keepAliveWithMetad(namespace, stream)
	return nil
}

func (s *service) watchShardTable(stream metadpb.Metad_RegisterClientClient) {
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Error(err)
			return
		}
		table := s.resource.updateTable(
			rsp.Namespace,
			int(rsp.ShardNum),
			int(rsp.ReplicaFactor),
		)
		for _, ss := range rsp.Shards {
			n, err := s.resource.getOrCreateNode(ss.NodeAddr)
			if err != nil {
				log.Error(err)
				continue
			}
			table.setShard(
				&shard{
					id: fmt.Sprintf("%s.%d.%d",
						table.namespace, ss.GroupID, ss.ReplicaID),
					groupID:   int(ss.GroupID),
					replicaID: int(ss.ReplicaID),
					node:      n,
				},
			)
		}
	}
}

func (s *service) keepAliveWithMetad(
	namespace string,
	stream metadpb.Metad_RegisterClientClient) {

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			err := stream.Send(&metadpb.RegisterClientReq{
				Namespace: namespace,
			})
			if err != nil {
				log.Error(err)
			}
		case <-s.exitC:
			goto exit
		}
	}
exit:
	ticker.Stop()
}

func (s *service) IndexDoc(
	ctx context.Context,
	req *proxypb.IndexDocReq) (*proxypb.IndexDocRsp, error) {

	table, ok := s.resource.getTable(req.Namespace)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	key := hash(id[:])
	groupID := key % uint64(table.shardNum)
	replicaID := key % uint64(table.replicaFactor)
	shard, _ := table.getShard(int(groupID), int(replicaID))
	_, err = shard.node.client.IndexDoc(ctx, &searchdpb.IndexDocReq{
		ShardID: shard.id,
		Doc:     req.Doc,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proxypb.IndexDocRsp{DocID: string(id[:])}, nil
}

func (s *service) RemoveDoc(
	ctx context.Context,
	req *proxypb.RemoveDocReq) (*proxypb.RemoveDocRsp, error) {

	table, ok := s.resource.getTable(req.Namespace)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
	key := hash([]byte(req.DocID))
	groupID := key % uint64(table.shardNum)
	replicaID := key % uint64(table.replicaFactor)
	shard, _ := table.getShard(int(groupID), int(replicaID))
	_, err := shard.node.client.RemoveDoc(ctx, &searchdpb.RemoveDocReq{
		ShardID: shard.id,
		DocID:   req.DocID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proxypb.RemoveDocRsp{}, nil
}

func (s *service) Search(
	ctx context.Context,
	req *proxypb.SearchReq) (*proxypb.SearchRsp, error) {

	table, ok := s.resource.getTable(req.Namespace)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
	var (
		wg   sync.WaitGroup
		rsps = make([]*searchdpb.SearchRsp, table.shardNum)
	)
	for i := range table.groups {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			shard, _ := table.getShard(i, 0)
			rsp, err := shard.node.client.Search(ctx, &searchdpb.SearchReq{
				ShardID: shard.id,
				Query:   req.Query,
				Start:   req.Start,
				Count:   req.Count,
			})
			if err != nil {
				log.Error(err)
				return
			}
			rsps[i] = rsp
		}(i)
	}
	docHeap := &DocHeap{}
	heap.Init(docHeap)
	for _, rsp := range rsps {
		for _, doc := range rsp.Docs {
			heap.Push(docHeap, doc)
		}
	}
	var (
		i      = uint64(0)
		start  = req.Start
		count  = req.Count
		docNum = docHeap.Len()
		docs   []*xsearchpb.Document
	)
	for count > 0 && docNum > 0 {
		doc := heap.Pop(docHeap).(*xsearchpb.Document)
		if i >= start {
			docs = append(docs, doc)
			count--
			docNum--
		}
		i++
	}
	return &proxypb.SearchRsp{Docs: docs}, nil
}
