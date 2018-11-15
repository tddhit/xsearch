package proxy

import (
	"context"
	"unsafe"

	"github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tddhit/box/interceptor"
	"github.com/tddhit/box/transport/common"
	"github.com/tddhit/xsearch/internal/types"
)

func CheckParams(s *service) interceptor.UnaryServerMiddleware {
	return func(next interceptor.UnaryHandler) interceptor.UnaryHandler {
		return func(ctx context.Context, req interface{},
			info *common.UnaryServerInfo) (interface{}, error) {

			iface := (*types.Iface)(unsafe.Pointer(&req))
			header := (*types.ReqHeader)(unsafe.Pointer(iface.Data))
			if header.TraceID == "" {
				if id, err := uuid.NewV4(); err != nil {
					return nil, status.Error(codes.Internal, err.Error())
				} else {
					header.TraceID = id.String()
				}
			}
			return next(ctx, req, info)
		}
	}
}
