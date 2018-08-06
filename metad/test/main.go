package main

import (
	"context"
	"time"

	pb "github.com/tddhit/xsearch/metad/pb"
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

		c := pb.NewMetadGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		reply, err := c.Echo(ctx, &pb.EchoRequest{Msg: "hello"})
		if err != nil {
			log.Fatalf("could not echo: %v", err)
		}
		log.Debug("Grpc Echo: ", reply.Msg)
	}
}
