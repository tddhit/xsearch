package metad

import (
	"context"

	"github.com/tddhit/box/interceptor"
	"github.com/tddhit/box/transport/common"
	"github.com/tddhit/tools/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	metadpb "github.com/tddhit/xsearch/metad/pb"
)

type contextKey struct {
	name string
}

var (
	tableContextKey = &contextKey{"table"}
)

func CheckParams(s *service) interceptor.UnaryServerMiddleware {
	return func(next interceptor.UnaryHandler) interceptor.UnaryHandler {
		return func(ctx context.Context, req interface{},
			info *common.UnaryServerInfo) (interface{}, error) {

			log.Info(info.FullMethod)
			var (
				namespace  string
				checkTable bool = true
			)
			switch info.FullMethod {
			case "/metadpb.Metad/CreateNamespace":
				namespace = req.(*metadpb.CreateNamespaceReq).Namespace
				checkTable = false
			case "/metadpb.Metad/DropNamespace":
				namespace = req.(*metadpb.DropNamespaceReq).Namespace
				checkTable = false
			case "/metadpb.Metad/AddNodeToNamespace":
				namespace = req.(*metadpb.AddNodeToNamespaceReq).Namespace
			case "/metadpb.Metad/RemoveNodeFromNamespace":
				namespace = req.(*metadpb.RemoveNodeFromNamespaceReq).Namespace
			case "/metadpb.Metad/ReplaceNodeInNamespace":
				namespace = req.(*metadpb.ReplaceNodeInNamespaceReq).Namespace
			case "/metadpb.Metad/AutoBalance":
				namespace = req.(*metadpb.AutoBalanceReq).Namespace
			case "/metadpb.Metad/MigrateShard":
				namespace = req.(*metadpb.MigrateShardReq).Namespace
			case "/metadpb.Metad/Commit":
				namespace = req.(*metadpb.CommitReq).Namespace
			default:
				return next(ctx, req, info)
			}
			if namespace == "" {
				return nil, status.Error(codes.NotFound, "namespace")
			}
			if checkTable {
				table, ok := s.resource.getTable(namespace)
				if !ok {
					return nil, status.Error(codes.NotFound, namespace)
				}
				ctx = context.WithValue(ctx, tableContextKey, table)
			}
			return next(ctx, req, info)
		}
	}
}
