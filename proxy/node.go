package service

import (
	"github.com/tddhit/box/transport"
	"github.com/tddhit/tools/log"

	searchdpb "github.com/tddhit/xsearch/searchd/pb"
)

type node struct {
	addr   string
	conn   transport.ClientConn
	client searchdpb.SearchdGrpcClient
}

func newNode(addr string) (*node, error) {
	conn, err := transport.Dial("grpc://" + addr)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	n := &node{
		addr:   addr,
		client: searchdpb.NewSearchdGrpcClient(conn),
	}
	return n, nil
}
