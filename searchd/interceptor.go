package searchd

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tddhit/box/interceptor"
	"github.com/tddhit/box/transport/common"
	"github.com/tddhit/xsearch/searchd/pb"
)

type contextKey struct {
	name string
}

var (
	shardContextKey = &contextKey{"shard"}
)

func CheckParams(s *service) interceptor.UnaryServerMiddleware {
	return func(next interceptor.UnaryHandler) interceptor.UnaryHandler {
		return func(ctx context.Context, req interface{},
			info *common.UnaryServerInfo) (interface{}, error) {

			var shardID string
			switch info.FullMethod {
			case "/searchdpb.Searchd/IndexDoc":
				shardID = req.(*searchdpb.IndexDocReq).ShardID
			case "/searchdpb.Searchd/RemoveDoc":
				shardID = req.(*searchdpb.RemoveDocReq).ShardID
			case "/searchdpb.Searchd/Search":
				shardID = req.(*searchdpb.SearchReq).ShardID
			}
			if shardID == "" {
				return nil, status.Error(codes.NotFound, "shardID")
			}
			shard, ok := s.resource.getShard(shardID)
			if !ok {
				return nil, status.Error(codes.NotFound, shardID)
			}
			ctx = context.WithValue(ctx, shardContextKey, shard)
			return next(ctx, req, info)
		}
	}
}
