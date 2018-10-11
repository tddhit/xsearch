package service

import (
	"context"

	"github.com/tddhit/box/interceptor"
	"github.com/tddhit/box/transport/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	metadpb "github.com/tddhit/xsearch/metad/pb"
)

type contextKey struct {
	name string
}

var (
	tableContextKey  = &contextKey{"offline-table"}
	onlineContextKey = &contextKey{"online-table"}
)

func CheckNamespaceWithUnary(s *service) interceptor.UnaryServerMiddleware {
	return func(next interceptor.UnaryHandler) interceptor.UnaryHandler {
		return func(ctx context.Context, req interface{},
			info *common.UnaryServerInfo) (interface{}, error) {

			var namespace string
			switch info.FullMethod {
			case "/metad.Metad/AddNodeToNamespace":
				namespace = req.(*metadpb.AddNodeToNamespaceReq).Namespace
			case "/metad.Metad/RemoveNodeFromNamespace":
				namespace = req.(*metadpb.RemoveNodeFromNamespaceReq).Namespace
			case "/metad.Metad/ReplaceNodeInNamespace":
				namespace = req.(*metadpb.ReplaceNodeInNamespaceReq).Namespace
			case "/metad.Metad/AutoBalance":
				namespace = req.(*metadpb.AutoBalanceReq).Namespace
			case "/metad.Metad/MigrateShard":
				namespace = req.(*metadpb.MigrateShardReq).Namespace
			case "/metad.Metad/Commit":
				namespace = req.(*metadpb.CommitReq).Namespace
			}
			if namespace == "" {
				return nil, status.Error(codes.NotFound, "namespace")
			}
			table, ok := s.resource.getTable(namespace)
			if !ok {
				return nil, status.Error(codes.NotFound, namespace)
			}
			online, ok := s.resource.getOnlineTable(namespace)
			if !ok {
				return nil, status.Error(codes.NotFound, namespace)
			}
			ctx = context.WithValue(ctx, tableContextKey, table)
			ctx = context.WithValue(ctx, onlineContextKey, online)
			return next(ctx, req, info)
		}
	}
}
