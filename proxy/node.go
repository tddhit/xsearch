package proxy

import (
	"sync/atomic"
	"unsafe"

	"github.com/tddhit/box/transport"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/searchd/pb"
)

type nodeStatus int32

const (
	nodeOffline nodeStatus = iota
	nodeOnline
	nodeConnectFail
)

func (n nodeStatus) String() string {
	switch n {
	case nodeOffline:
		return "offline"
	case nodeOnline:
		return "online"
	default:
		return "invalid node status"
	}
}

type node struct {
	addr   string
	status nodeStatus
	conn   transport.ClientConn
	client searchdpb.SearchdGrpcClient
}

func newNode(addr string) *node {
	return &node{
		addr:   addr,
		status: nodeOnline,
	}
}

func (n *node) connect() error {
	conn, err := transport.Dial("grpc://" + n.addr)
	if err != nil {
		log.Error(err)
		return err
	}
	n.conn = conn
	n.client = searchdpb.NewSearchdGrpcClient(conn)
	return nil
}

func (n *node) setStatus(status nodeStatus) {
	atomic.StoreInt32((*int32)(unsafe.Pointer(&n.status)), int32(status))
}

func (n *node) isOnline() bool {
	i := atomic.LoadInt32((*int32)(unsafe.Pointer(&n.status)))
	return nodeStatus(i) == nodeOnline
}
