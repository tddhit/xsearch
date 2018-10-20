package searchd

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/tddhit/box/transport"
	diskqueuepb "github.com/tddhit/diskqueue/pb"
	"github.com/tddhit/tools/log"
	"github.com/urfave/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tddhit/xsearch/metad/pb"
	"github.com/tddhit/xsearch/searchd/pb"
)

type service struct {
	resource *resource
	metad    metadpb.MetadGrpcClient
	diskq    diskqueuepb.DiskqueueGrpcClient
}

func NewService(ctx *cli.Context) *service {
	dataDir := ctx.String("datadir")
	metadAddr := ctx.String("metad")
	dqAddr := ctx.String("diskqueue")
	if dqAddr == "" || metadAddr == "" {
		log.Fatal("invalid diskqueue/metad address")
	}
	conn, err := transport.Dial(metadAddr)
	if err != nil {
		log.Fatal(err)
	}
	diskq := diskqueuepb.NewDiskqueueGrpcClient(conn)
	conn, err = transport.Dial(dqAddr)
	if err != nil {
		log.Fatal(err)
	}
	metad := metadpb.NewMetadGrpcClient(conn)
	return &service{
		resource: newResource(dataDir),
		metad:    metad,
		diskq:    diskq,
	}
}

func (s *service) IndexDoc(ctx context.Context,
	req *searchdpb.IndexDocReq) (*searchdpb.IndexDocRsp, error) {

	data, err := proto.Marshal(&searchdpb.Command{
		Type: searchdpb.Command_INDEX,
		DocOneof: &searchdpb.Command_Doc{
			Doc: req.Doc,
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_, err = s.diskq.Push(context.Background(), &diskqueuepb.PushRequest{
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
	_, err = s.diskq.Push(context.Background(), &diskqueuepb.PushRequest{
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
