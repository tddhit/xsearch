package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tddhit/xsearch/searchd/searchdpb"
)

type admin struct {
	resource *resource
}

func NewAdmin(r *resource) *admin {
	a := &admin{
		resource: r,
	}
	return a
}

func (a *admin) CreateShard(
	ctx context.Context,
	in *searchdpb.CreateShardReq) (*searchdpb.CreateShardRsp, error) {

	_, created := a.resource.getOrCreateShard(in.Namespace, in.ShardID)
	if !created {
		return nil, status.Error(
			codes.AlreadyExists,
			fmt.Sprintf("namespace(%d).shardID(%d)", in.Namespace, in.ShardID),
		)
	}
	return &searchdpb.CreateShardRsp{}, nil
}

func (a *admin) MigrateShard(
	ctx context.Context,
	in *searchdpb.MigrateShardReq) (*searchdpb.MigrateShardRsp, error) {

}

func (a *admin) ReplicateShard(
	ctx context.Context,
	in *searchdpb.ReplicateShardReq) (*searchdpb.ReplicateShardRsp, error) {

}
