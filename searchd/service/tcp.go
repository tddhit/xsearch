package service

import (
	"bufio"
	"net"
	"runtime"
	"sync"
)

const defaultBufferSize = 16 * 1024

type peer struct {
	addr   string
	reader *bufio.Reader
	writer *bufio.Writer
	wg     sync.WaitGroup
}

func newPeer(addr string, conn *net.Conn) (*peer, error) {
	p := &peer{
		addr:   addr,
		reader: bufio.NewReaderSize(conn, defaultBufferSize),
		writer: bufio.NewWriterSize(conn, defaultBufferSize),
	}
	go p.writeLoop()
	go p.readLoop()
	return p, nil
}

func (p *peer) writeLoop() {
	for {
		p.writer.Write()
	}
}

func (p *peer) readLoop() {
	for {
		p.reader.Read()
	}
}

type tcpServer struct {
	addr     string
	listener *net.Listener
}

func newTCPServer(addr string) (*tcpServer, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &tcpServer{
		addr:     addr,
		listener: ln,
	}, nil
}

func (t *tcpServer) serve() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				log.Warn("temporary Accept() failure - %s", err)
				runtime.Gosched()
				continue
			}
			break
		}
		go t.handle(conn)
	}
}

func (t *tcpServer) handle(conn *net.Conn) {
	p := newPeer(conn)
	p.wg.Add(1)
	go func() {
		p.writeLoop()
		p.wg.Done()
	}()
	p.wg.Add(1)
	go func() {
		p.readLoop()
		p.wg.Done()
	}()
	p.wg.Wait()
}
