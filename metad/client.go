package metad

import (
	"context"
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
		readC:     make(chan *metadpb.RegisterClientReq, 100),
		writeC:    make(chan *metadpb.RegisterClientRsp, 100),
	}
}

func (c *client) readLoop(ctx context.Context) {
	timer := time.NewTimer(3 * time.Second)
	for {
		select {
		case <-c.readC:
			timer.Reset(3 * time.Second)
		case <-timer.C:
			goto exit
		case <-ctx.Done():
			goto exit
		}
	}
exit:
	timer.Stop()
	c.close()
}

func (c *client) writeLoop(ctx context.Context,
	stream metadpb.Metad_RegisterClientServer) {

	for {
		select {
		case rsp, ok := <-c.writeC:
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

func (c *client) ioLoop(
	ctx context.Context,
	stream metadpb.Metad_RegisterClientServer) {

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		c.readLoop(ctx)
		wg.Done()
	}()
	go func() {
		c.writeLoop(ctx, stream)
		wg.Done()
	}()
	wg.Wait()
}

func (c *client) close() {
	if c.readC != nil {
		close(c.readC)
		c.readC = nil
	}
	if c.writeC != nil {
		close(c.writeC)
		c.writeC = nil
	}
}
