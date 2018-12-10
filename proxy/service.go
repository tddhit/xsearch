package proxy

import (
	"container/heap"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/urfave/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/box/transport"
	"github.com/tddhit/diskqueue/pb"
	_ "github.com/tddhit/diskqueue/resolver"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/internal/util"
	"github.com/tddhit/xsearch/metad/pb"
	"github.com/tddhit/xsearch/pb"
	"github.com/tddhit/xsearch/plugin"
	"github.com/tddhit/xsearch/proxy/pb"
	"github.com/tddhit/xsearch/searchd/pb"
)

type service struct {
	resource *Resource
	metad    metadpb.MetadGrpcClient
	diskq    diskqueuepb.DiskqueueGrpcClient
	exitC    chan struct{}
}

func NewService(ctx *cli.Context, r *Resource) *service {
	if !mw.IsWorker() {
		return nil
	}
	conn, err := transport.Dial(ctx.String("metad"))
	if err != nil {
		log.Fatal(err)
	}
	metad := metadpb.NewMetadGrpcClient(conn)
	conn, err = transport.Dial(ctx.String("diskqueue"))
	if err != nil {
		log.Fatal(err)
	}
	diskq := diskqueuepb.NewDiskqueueGrpcClient(conn)
	s := &service{
		resource: r,
		metad:    metad,
		diskq:    diskq,
		exitC:    make(chan struct{}),
	}
	nss := strings.Split(ctx.String("namespaces"), ",")
	for _, ns := range nss {
		if err := s.registerClient(ns); err != nil {
			log.Fatal(err)
		}
	}
	plugins := strings.Split(ctx.String("plugins"), ",")
	for _, name := range plugins {
		p, ok := plugin.Get(name)
		if !ok {
			continue
		}
		if err := p.Init(); err != nil {
			log.Fatal(err)
		}
	}
	return s
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
			rsp.Table.Namespace,
			int(rsp.Table.ShardNum),
			int(rsp.Table.ReplicaFactor),
		)
		for _, ss := range rsp.Table.Shards {
			n, _ := s.resource.getOrCreateNode(ss.NodeAddr)
			switch ss.NodeStatus {
			case "online":
				n.setStatus(nodeOnline)
				if err := n.connect(); err != nil {
					n.setStatus(nodeConnectFail)
				}
			case "offline":
				n.setStatus(nodeOffline)
			}
			table.setShard(
				&shard{
					id: fmt.Sprintf("%s.%d.%d",
						rsp.Table.Namespace, ss.GroupID, ss.ReplicaID),
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
	key := util.Hash(id[:])
	groupID := key % uint64(table.shardNum)
	segmenter, _ := plugin.Get("segmenter")
	query := &xsearchpb.QueryAnalysisArgs{
		Queries: []*xsearchpb.Query{
			{Raw: req.Doc.Content},
		},
	}
	segmenter.Analyze(query)
	data, err := proto.Marshal(&xsearchpb.Command{
		Type: xsearchpb.Command_INDEX,
		DocOneof: &xsearchpb.Command_Doc{
			Doc: &xsearchpb.Document{
				ID:     id.String(),
				Tokens: query.Queries[0].Tokens,
			},
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_, err = s.diskq.Push(context.Background(), &diskqueuepb.PushReq{
		Topic: fmt.Sprintf("%s.%d", table.namespace, groupID),
		Data:  data,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Infof("Type=IndexDoc\tTraceID=%s\tDocID=%s\tShard=%s.%d",
		req.TraceID, id.String(), table.namespace, groupID)
	return &proxypb.IndexDocRsp{DocID: id.String()}, nil
}

func (s *service) RemoveDoc(
	ctx context.Context,
	req *proxypb.RemoveDocReq) (*proxypb.RemoveDocRsp, error) {

	table, ok := s.resource.getTable(req.Namespace)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
	id, err := uuid.FromString(req.DocID)
	if err != nil {
		log.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	key := util.Hash(id[:])
	groupID := key % uint64(table.shardNum)
	data, err := proto.Marshal(&xsearchpb.Command{
		Type: xsearchpb.Command_REMOVE,
		DocOneof: &xsearchpb.Command_DocID{
			DocID: id.String(),
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_, err = s.diskq.Push(context.Background(), &diskqueuepb.PushReq{
		Topic: fmt.Sprintf("%s.%d", table.namespace, groupID),
		Data:  data,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Infof("Type=RemoveDoc\tTraceID=%s\tDocID=%s\tShard=%s.%d",
		req.TraceID, id.String(), table.namespace, groupID)
	return &proxypb.RemoveDocRsp{}, nil
}

func (s *service) Search(
	ctx context.Context,
	req *proxypb.SearchReq) (*proxypb.SearchRsp, error) {

	table, ok := s.resource.getTable(req.Namespace)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
	args := &xsearchpb.QueryAnalysisArgs{
		Queries: []*xsearchpb.Query{
			req.Query,
		},
	}
	if err := plugin.Analyze(args); err != nil {
		return nil, err
	}
	var (
		wg   sync.WaitGroup
		rsps = make([]*searchdpb.SearchRsp, table.shardNum)
		docs []*xsearchpb.Document
	)
	for i := range table.groups {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			var (
				key       = util.Hash([]byte(req.Query.Raw))
				replicaID = key % uint64(table.replicaFactor)
				shard     *shard
				tryCount  int
			)
			for {
				tryCount++
				shard, _ = table.getShard(i, int(replicaID)%table.replicaFactor)
				if shard.node.isOnline() {
					log.Debugf("Type=Search\tTraceID=%s\tQuery=%s\tShard=%s",
						req.TraceID, req.Query.Raw, shard.id)
					break
				}
				replicaID++
				if tryCount == table.replicaFactor {
					return
				}
			}
			rsp, err := shard.node.client.Search(ctx, &searchdpb.SearchReq{
				TraceID: req.TraceID,
				ShardID: shard.id,
				Query:   args.Queries[0],
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
	wg.Wait()
	for _, rsp := range rsps {
		if rsp != nil {
			for _, doc := range rsp.Docs {
				docs = append(docs, doc)
			}
		}
	}
	rerankArgs := &xsearchpb.RerankArgs{
		Query: args.Queries[0],
		Docs:  docs,
	}
	if err := plugin.Rerank(rerankArgs); err != nil {
		return nil, err
	}
	docHeap := &DocHeap{}
	heap.Init(docHeap)
	for _, doc := range rerankArgs.Docs {
		heap.Push(docHeap, doc)
	}
	var (
		i         = uint64(0)
		start     = req.Start
		count     = req.Count
		docNum    = docHeap.Len()
		finalDocs []*xsearchpb.Document
	)
	for count > 0 && docNum > 0 {
		doc := heap.Pop(docHeap).(*xsearchpb.Document)
		if i >= start {
			finalDocs = append(finalDocs, doc)
			count--
			docNum--
		}
		i++
	}
	log.Infof("Type=Search\tTraceID=%s\tQuery=%s",
		req.TraceID, req.Query.Raw)
	return &proxypb.SearchRsp{
		Query: args.Queries[0],
		Docs:  finalDocs,
	}, nil
}
