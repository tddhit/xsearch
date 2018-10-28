package searchd

import (
	"context"
	"io"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/urfave/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/box/transport"
	"github.com/tddhit/diskqueue/pb"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/metad/pb"
	"github.com/tddhit/xsearch/searchd/pb"
)

type service struct {
	addr     string
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
		addr:     ctx.String("addr"),
		resource: r,
		metad:    metad,
		diskq:    diskq,
		exitC:    make(chan struct{}),
	}
	if err := s.registerNode(s.addr); err != nil {
		log.Fatal(err)
	}
	return s
}

func (s *service) registerNode(addr string) error {
	stream, err := s.metad.RegisterNode(context.Background())
	if err != nil {
		return err
	}
	go s.waitCommand(addr, stream)
	go s.keepAliveWithMetad(addr, stream)
	return nil
}

func (s *service) waitCommand(addr string, stream metadpb.Metad_RegisterNodeClient) {
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Error(err)
			return
		}
		switch rsp.Type {
		case metadpb.RegisterNodeRsp_CreateShard:
			log.Debug("create shard")
			_, err := s.resource.createShard(rsp.ShardID, s.addr, s.diskq)
			if err != nil {
				log.Error(err)
				continue
			}
			log.Debug("register shard")
			err = stream.Send(&metadpb.RegisterNodeReq{
				Type:    metadpb.RegisterNodeReq_RegisterShard,
				Addr:    addr,
				ShardID: rsp.ShardID,
			})
			if err != nil {
				log.Error(err)
			}
		case metadpb.RegisterNodeRsp_RemoveShard:
			log.Debug("remove shard")
			if err := s.resource.removeShard(rsp.ShardID); err != nil {
				log.Error(err)
			}
			log.Debug("unregister shard")
			err := stream.Send(&metadpb.RegisterNodeReq{
				Type:    metadpb.RegisterNodeReq_UnregisterShard,
				Addr:    addr,
				ShardID: rsp.ShardID,
			})
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func (s *service) keepAliveWithMetad(
	addr string,
	stream metadpb.Metad_RegisterNodeClient) {

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			err := stream.Send(&metadpb.RegisterNodeReq{
				Type: metadpb.RegisterNodeReq_Heartbeat,
				Addr: addr,
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

func (s *service) IndexDoc(ctx context.Context,
	req *searchdpb.IndexDocReq) (*searchdpb.IndexDocRsp, error) {

	log.Debug(req.Doc.Raw)
	data, err := proto.Marshal(&searchdpb.Command{
		Type: searchdpb.Command_INDEX,
		DocOneof: &searchdpb.Command_Doc{
			Doc: req.Doc,
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Debug(len(data))
	_, err = s.diskq.Push(context.Background(), &diskqueuepb.PushReq{
		Topic: req.ShardID,
		Data:  data,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &searchdpb.IndexDocRsp{}, nil
}

func (s *service) RemoveDoc(ctx context.Context,
	req *searchdpb.RemoveDocReq) (*searchdpb.RemoveDocRsp, error) {

	data, err := proto.Marshal(&searchdpb.Command{
		Type: searchdpb.Command_REMOVE,
		DocOneof: &searchdpb.Command_DocID{
			DocID: req.DocID,
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_, err = s.diskq.Push(context.Background(), &diskqueuepb.PushReq{
		Topic: req.ShardID,
		Data:  data,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &searchdpb.RemoveDocRsp{}, nil
}

func (s *service) Search(ctx context.Context,
	req *searchdpb.SearchReq) (*searchdpb.SearchRsp, error) {

	shard := ctx.Value(shardContextKey).(*shard)
	docs, err := shard.indexer.Search(req.Query, req.Start, int32(req.Count))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &searchdpb.SearchRsp{
		Docs: docs,
	}, nil
}
