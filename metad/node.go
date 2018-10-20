package metad

import (
	"sync"
	"time"

	"github.com/tddhit/tools/log"

	"github.com/tddhit/xsearch/metad/pb"
)

type nodeStatus int

const (
	NODE_INITIAL nodeStatus = iota
	NODE_ONLINE
	NODE_OFFLINE
)

type node struct {
	addr   string
	status nodeStatus
	readC  chan *metadpb.RegisterNodeReq
	writeC chan *metadpb.RegisterNodeRsp
}

func newNode(addr string, status nodeStatus) *node {
	return &node{
		addr:   addr,
		status: status,
		readC:  make(chan *metadpb.RegisterNodeReq, 100),
		writeC: make(chan *metadpb.RegisterNodeRsp, 100),
	}
}

func (n *node) readLoop(r *resource) {
	timer := time.NewTimer(3 * time.Second)
	for {
		select {
		case req := <-n.readC:
			if req == nil {
				goto exit
			}
			timer.Reset(3 * time.Second)
			switch req.Type {
			case metadpb.RegisterNodeReq_RegisterShard:
				shard, err := r.getShard(
					req.Namespace,
					int(req.GroupID),
					int(req.ReplicaID),
				)
				if err != nil {
					break
				}
				if shard.backup.addr != req.Addr {
					log.Errorf("no match: %s != %s", shard.backup.addr, req.Addr)
					break
				}
				shard.execTodo()
			case metadpb.RegisterNodeReq_UnregisterShard:
				shard, err := r.getShard(
					req.Namespace,
					int(req.GroupID),
					int(req.ReplicaID),
				)
				if err != nil {
					break
				}
				if shard.node.addr != req.Addr {
					log.Errorf("no match: %s != %s", shard.node.addr, req.Addr)
					break
				}
				shard.execTodo()
			case metadpb.RegisterNodeReq_Heartbeat:
			}
		case <-timer.C:
			goto exit
		}
	}
exit:
	timer.Stop()
	n.close()
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
}

func (n *node) ioLoop(stream metadpb.Metad_RegisterNodeServer, r *resource) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		n.readLoop(r)
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
		n.readC = nil
	}
	if n.writeC != nil {
		close(n.writeC)
		n.writeC = nil
	}
	if n.status == NODE_ONLINE {
		n.status = NODE_OFFLINE
	}
}
