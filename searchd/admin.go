package searchd

import (
	"context"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/xsearch/searchd/pb"
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
	req *searchdpb.InfoReq) (*searchdpb.InfoRsp, error) {

	rsp := &searchdpb.InfoRsp{}
	a.resource.marshalTo(rsp)
	return rsp, nil
}
