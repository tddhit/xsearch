package metad

import (
	"strconv"
	"strings"
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

func (n *node) readLoop(r *resource) {
	timer := time.NewTimer(3 * time.Second)
	for {
		select {
		case req := <-n.readC:
			if req == nil {
				goto exit
			}
			timer.Reset(3 * time.Second)
			v := strings.Split(req.ShardID, ".")
			if len(v) != 3 {
				continue
			}
			namespace := v[0]
			groupID, err := strconv.Atoi(v[1])
			if err != nil {
				log.Error(err)
				continue
			}
			replicaID, err := strconv.Atoi(v[2])
			if err != nil {
				log.Error(err)
				continue
			}
			switch req.Type {
			case metadpb.RegisterNodeReq_RegisterShard:
				log.Debug("register shard")
				shard, err := r.getShard(namespace, groupID, replicaID)
				log.Debug("!")
				if err != nil {
					log.Debug("!")
					log.Debug(err)
					break
				}
				log.Debug("!")
				if shard.next.addr != req.Addr {
					log.Debug("!")
					log.Errorf("no match: %s != %s", shard.next.addr, req.Addr)
					break
				}
				log.Debug("!")
				log.Debug("exec todo")
				shard.execTodo()
				log.Debug("!")
			case metadpb.RegisterNodeReq_UnregisterShard:
				log.Debug("unregister shard")
				shard, err := r.getShard(namespace, groupID, replicaID)
				if err != nil {
					break
				}
				if err != nil {
					break
				}
				if shard.node.addr != req.Addr {
					log.Errorf("no match: %s != %s", shard.node.addr, req.Addr)
					break
				}
				shard.execTodo()
			case metadpb.RegisterNodeReq_Heartbeat:
				log.Debug("heartbeat")
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
		log.Debug("rsp")
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
	if n.status == NODE_CLUSTER_ONLINE {
		n.status = NODE_CLUSTER_OFFLINE
	}
}
