package service

import (
	"context"

	"github.com/gogo/protobuf/proto"
	diskqueuepb "github.com/tddhit/diskqueue/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	searchdpb "github.com/tddhit/xsearch/searchd/pb"
)

type service struct {
	resource *resource
	dqClient diskqueuepb.DiskqueueGrpcClient
}

func NewService(r *resource, c diskqueuepb.DiskqueueGrpcClient) *service {
	return &service{
		resource: r,
		dqClient: c,
	}
}

func (s *service) IndexDoc(ctx context.Context,
	req *searchdpb.IndexDocReq) (*searchdpb.IndexDocRsp, error) {

	_, ok := s.resource.getShard(req.ShardID)
	if !ok {
		return nil, status.Error(codes.NotFound, req.ShardID)
	}
	data, err := proto.Marshal(&searchdpb.Command{
		Type: searchdpb.Command_INDEX,
		DocOneof: &searchdpb.Command_Doc{
			Doc: req.Doc,
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_, err = s.dqClient.Push(context.Background(), &diskqueuepb.PushRequest{
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

	_, ok := s.resource.getShard(req.ShardID)
	if !ok {
		return nil, status.Error(codes.NotFound, req.ShardID)
	}
	data, err := proto.Marshal(&searchdpb.Command{
		Type: searchdpb.Command_REMOVE,
		DocOneof: &searchdpb.Command_DocID{
			DocID: req.DocID,
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_, err = s.dqClient.Push(context.Background(), &diskqueuepb.PushRequest{
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

	shard, ok := s.resource.getShard(req.ShardID)
	if !ok {
		return nil, status.Error(codes.NotFound, req.ShardID)
	}
	docs, err := shard.indexer.Search(req.Query, req.Start, int32(req.Count))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &searchdpb.SearchRsp{
		Docs: docs,
	}, nil
}
