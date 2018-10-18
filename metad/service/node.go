package service

import (
	"errors"
	"sync"
	"time"

	"github.com/tddhit/box/transport"
	"github.com/tddhit/tools/log"

	metadpb "github.com/tddhit/xsearch/metad/pb"
	searchdpb "github.com/tddhit/xsearch/searchd/pb"
)

type node struct {
	addr      string
	adminAddr string
	client    searchdpb.SearchdAdminGrpcClient
	shards    []*shard
	readC     chan *metadpb.RegisterNodeReq
	writeC    chan *metadpb.RegisterNodeRsp
}

func newNode(addr, adminAddr string) (*node, error) {
	conn, err := transport.Dial("grpc://" + adminAddr)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	n := &node{
		addr:      addr,
		adminAddr: adminAddr,
		client:    searchdpb.NewSearchdAdminGrpcClient(conn),
		readC:     make(chan *metadpb.RegisterNodeReq, 2),
		writeC:    make(chan *metadpb.RegisterNodeRsp, 2),
	}
	return n, err
}

func (n *node) readLoop() {
	timer := time.NewTimer(3 * time.Second)
	for {
		select {
		case req := <-n.readC:
			if req == nil {
				goto exit
			}
			timer.Reset(3 * time.Second)
		case <-timer.C:
			goto exit
		}
	}
exit:
	timer.Stop()
	n.close()
	n.readC = nil
}

func (n *node) writeLoop(stream metadpb.Metad_RegisterNodeServer) {
	for {
		rsp := <-n.writeC
		if rsp == nil {
			goto exit
		}
		if err := stream.Send(rsp); err != nil {
			log.Error(err)
			goto exit
		}
	}
exit:
	n.writeC = nil
}

func (n *node) removeShard(s *shard) error {
	i := -1
	for j, shard := range n.shards {
		if shard.table.namespace == s.table.namespace &&
			shard.id == s.id &&
			shard.replicaID == s.replicaID {

			i = j
			break
		}
	}
	if i == -1 {
		return errors.New("not found")
	}
	n.shards = append(n.shards[:i], n.shards[i+1:]...)
	return nil
}

func (n *node) ioLoop(stream metadpb.Metad_RegisterNodeServer) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		n.readLoop()
		wg.Done()
	}()
	go func() {
		n.writeLoop(stream)
		wg.Done()
	}()
	wg.Wait()
}

func (n *node) close() {
	if n.readC != nil {
		close(n.readC)
	}
	if n.writeC != nil {
		close(n.writeC)
	}
}
