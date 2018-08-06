package main

import (
	"context"
	"time"

	pb "github.com/tddhit/xsearch/proxy/pb"
	"github.com/tddhit/box/transport"
	"github.com/tddhit/tools/log"
)

func main() {
	{
		conn, err := transport.Dial("grpc://127.0.0.1:9000")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		c := pb.NewProxyGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		reply, err := c.Echo(ctx, &pb.EchoRequest{Msg: "hello"})
		if err != nil {
			log.Fatalf("could not echo: %v", err)
		}
		log.Debug("Grpc Echo: ", reply.Msg)
	}
	{
			conn, err := transport.Dial("http://127.0.0.1:9010")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		c := pb.NewProxyHttpClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 
				time.Second)
		defer cancel()
		reply, err := c.Echo(ctx, &pb.EchoRequest{Msg: "hello"})
		if err != nil {
			log.Fatalf("could not echo: %v", err)
		}
		log.Debug("Http Echo: ", reply.Msg)
	}
}
