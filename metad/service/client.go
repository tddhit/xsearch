package service

import (
	"time"

	"github.com/tddhit/tools/log"

	metadpb "github.com/tddhit/xsearch/metad/pb"
)

type client struct {
	addr      string
	namespace string
	readC     chan *metadpb.RegisterClientReq
	writeC    chan *metadpb.RegisterClientRsp
}

func newClient(addr, namespace string) *client {
	return &client{
		addr:      addr,
		namespace: namespace,
		readC:     make(chan *metadpb.RegisterClientReq, 2),
		writeC:    make(chan *metadpb.RegisterClientRsp, 2),
	}
}

func (c *client) readLoop() {
	timer := time.NewTimer(3 * time.Second)
	for {
		select {
		case req := <-c.readC:
			if req == nil {
				return
			}
			timer.Reset(3 * time.Second)
		case <-timer.C:
			goto exit
		}
	}
exit:
	timer.Stop()
}

func (c *client) writeLoop(stream metadpb.Metad_RegisterClientServer) {
	for {
		select {
		case rsp := <-c.writeC:
			if rsp == nil {
				return
			}
			if err := stream.Send(rsp); err != nil {
				log.Error(err)
				return
			}
		}
	}
}

func (c *client) close() {
	if c.writeC != nil {
		close(c.writeC)
	}
	if c.readC != nil {
		close(c.readC)
	}
}
