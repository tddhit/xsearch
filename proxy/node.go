package proxy

import (
	"github.com/tddhit/box/transport"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/searchd/pb"
)

type node struct {
	addr   string
	status string
	conn   transport.ClientConn
	client searchdpb.SearchdGrpcClient
}

func newNode(addr, status string) (*node, error) {
	n := &node{
		addr:   addr,
		status: status,
	}
	if status == "online" {
		conn, err := transport.Dial("grpc://" + addr)
		if err != nil {
			log.Error(err)
			n.status = "fail"
		} else {
			n.conn = conn
			n.client = searchdpb.NewSearchdGrpcClient(conn)
		}
	}
	return n, nil
}
