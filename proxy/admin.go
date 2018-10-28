package proxy

import (
	"context"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/xsearch/proxy/pb"
)

type admin struct {
	resource *Resource
}

func NewAdmin(r *Resource) *admin {
	if !mw.IsWorker() {
		return nil
	}
	return &admin{r}
}

func (a *admin) Info(ctx context.Context,
	req *proxypb.InfoReq) (*proxypb.InfoRsp, error) {

	rsp := &proxypb.InfoRsp{}
	a.resource.marshalTo(rsp)
	return rsp, nil
}
