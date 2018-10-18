package service

import (
	"sync"
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
		readC:     make(chan *metadpb.RegisterClientReq, 10),
		writeC:    make(chan *metadpb.RegisterClientRsp, 10),
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
	c.close()
	c.readC = nil
}

func (c *client) writeLoop(stream metadpb.Metad_RegisterClientServer) {
	for {
		rsp := <-c.writeC
		if rsp == nil {
			return
		}
		if err := stream.Send(rsp); err != nil {
			log.Error(err)
			goto exit
		}
	}
exit:
	c.writeC = nil
}

func (c *client) ioLoop(stream metadpb.Metad_RegisterClientServer) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		c.readLoop()
		wg.Done()
	}()
	go func() {
		c.writeLoop(stream)
		wg.Done()
	}()
	wg.Wait()
}

func (c *client) close() {
	if c.writeC != nil {
		close(c.writeC)
	}
	if c.readC != nil {
		close(c.readC)
	}
}
