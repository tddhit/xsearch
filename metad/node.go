package metad

import (
	"context"
	"sync"
	"time"

	"github.com/tddhit/tools/log"

	"github.com/tddhit/xsearch/metad/pb"
)

type nodeStatus int

const (
	NODE_ISOLATED_OFFLINE nodeStatus = iota
	NODE_ISOLATED_ONLINE
	NODE_CLUSTER_OFFLINE
	NODE_CLUSTER_ONLINE
)

func (ns nodeStatus) String() string {
	switch ns {
	case NODE_ISOLATED_OFFLINE:
		return "isolated offline"
	case NODE_ISOLATED_ONLINE:
		return "isolated online"
	case NODE_CLUSTER_OFFLINE:
		return "cluster offline"
	case NODE_CLUSTER_ONLINE:
		return "cluster online"
	default:
		return "unknown"
	}
}

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

func (n *node) getAddr() string {
	if n == nil {
		return ""
	} else {
		return n.addr
	}
}

func (n *node) getClusterStatus() string {
	if n == nil {
		return ""
	} else {
		if n.status == NODE_CLUSTER_ONLINE {
			return "online"
		} else {
			return "offline"
		}
	}
}

func (n *node) getInfo() string {
	if n == nil {
		return ""
	} else {
		return n.addr + ":" + n.status.String()
	}
}

func (n *node) readLoop(ctx context.Context, res *resource, rec *reception) {
	timer := time.NewTimer(3 * time.Second)
	for {
		select {
		case req := <-n.readC:
			timer.Reset(3 * time.Second)
			if req == nil || req.Type == metadpb.RegisterNodeReq_Heartbeat {
				continue
			}
			shard, err := res.getShard(
				req.Namespace,
				int(req.GroupID),
				int(req.ReplicaID),
			)
			if err != nil {
				log.Error(err)
				continue
			}
			table, ok := res.getTable(req.Namespace)
			if !ok {
				log.Errorf("not found namespace %s", req.Namespace)
				continue
			}
			switch req.Type {
			case metadpb.RegisterNodeReq_PutShardOnline:
				if shard.node.addr == req.Addr &&
					shard.node.status == NODE_CLUSTER_OFFLINE {

					shard.node = n
					shard.node.status = NODE_CLUSTER_ONLINE
					table.removeNode(n.addr)
					table.addNode(n)
					rec.notifyByNamespace(req.Namespace, table.marshal())
				}
			case metadpb.RegisterNodeReq_RegisterShard:
				if shard.next.addr != req.Addr {
					log.Errorf("no match: %s != %s", shard.next.addr, req.Addr)
					continue
				}
				shard.execTodo()
			case metadpb.RegisterNodeReq_UnregisterShard:
				shard.execTodo()
			}
		case <-timer.C:
			log.Errorf("%s Heartbeat Timeout", n.addr)
			goto exit
		case <-ctx.Done():
			goto exit
		}
	}
exit:
	timer.Stop()
	n.close()
}

func (n *node) writeLoop(ctx context.Context,
	stream metadpb.Metad_RegisterNodeServer) {

	for {
		select {
		case rsp, ok := <-n.writeC:
			if !ok {
				goto exit
			}
			if rsp != nil {
				if err := stream.Send(rsp); err != nil {
					log.Error(err)
				}
			}
		case <-ctx.Done():
			goto exit
		}
	}
exit:
}

func (n *node) ioLoop(
	ctx context.Context,
	stream metadpb.Metad_RegisterNodeServer,
	res *resource,
	rec *reception) {

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		n.readLoop(ctx, res, rec)
		wg.Done()
	}()
	go func() {
		n.writeLoop(ctx, stream)
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
	if n.status == NODE_CLUSTER_ONLINE {
		n.status = NODE_CLUSTER_OFFLINE
	}
}
