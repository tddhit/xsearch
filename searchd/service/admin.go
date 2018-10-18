package service

import (
	"context"

	diskqueuepb "github.com/tddhit/diskqueue/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	searchdpb "github.com/tddhit/xsearch/searchd/pb"
)

type admin struct {
	resource *resource
	dqClient diskqueuepb.DiskqueueGrpcClient
}

func NewAdmin(r *resource, c diskqueuepb.DiskqueueGrpcClient) *admin {
	return &admin{
		resource: r,
		dqClient: c,
	}
}

func (a *admin) CreateShard(
	ctx context.Context,
	req *searchdpb.CreateShardReq) (*searchdpb.CreateShardRsp, error) {

	if _, err := a.resource.createShard(req.ShardID, a.dqClient); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &searchdpb.CreateShardRsp{}, nil
}

func (a *admin) RemoveShard(
	ctx context.Context,
	req *searchdpb.RemoveShardReq) (*searchdpb.RemoveShardRsp, error) {

	if err := a.resource.removeShard(req.ShardID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &searchdpb.RemoveShardRsp{}, nil
}
