package searchd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/urfave/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/box/transport"
	"github.com/tddhit/diskqueue/pb"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/metad/pb"
	"github.com/tddhit/xsearch/searchd/pb"
)

type service struct {
	addr     string
	resource *Resource
	metad    metadpb.MetadGrpcClient
	diskq    diskqueuepb.DiskqueueGrpcClient
	exitC    chan struct{}
}

func NewService(ctx *cli.Context, r *Resource) *service {
	if !mw.IsWorker() {
		return nil
	}
	conn, err := transport.Dial(ctx.String("metad"))
	if err != nil {
		log.Fatal(err)
	}
	metad := metadpb.NewMetadGrpcClient(conn)
	conn, err = transport.Dial(ctx.String("diskqueue"))
	if err != nil {
		log.Fatal(err)
	}
	diskq := diskqueuepb.NewDiskqueueGrpcClient(conn)
	s := &service{
		addr:     ctx.String("addr"),
		resource: r,
		metad:    metad,
		diskq:    diskq,
		exitC:    make(chan struct{}),
	}
	if err := r.loadShards(s.addr, diskq); err != nil {
		log.Fatal(err)
	}
	if err := s.registerNode(s.addr); err != nil {
		log.Fatal(err)
	}
	return s
}

func (s *service) registerNode(addr string) error {
	stream, err := s.metad.RegisterNode(context.Background())
	if err != nil {
		return err
	}
	s.resource.rangeShards(func(s *shard) error {
		err := stream.Send(&metadpb.RegisterNodeReq{
			Type:      metadpb.RegisterNodeReq_PutShardOnline,
			Addr:      addr,
			Namespace: s.namespace,
			GroupID:   uint32(s.groupID),
			ReplicaID: uint32(s.replicaID),
		})
		if err != nil {
			log.Error(err)
		}
		return err
	})
	go s.waitCommand(addr, stream)
	go s.keepAliveWithMetad(addr, stream)
	return nil
}

func (s *service) waitCommand(addr string, stream metadpb.Metad_RegisterNodeClient) {
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Error(err)
			return
		}
		switch rsp.Type {
		case metadpb.RegisterNodeRsp_CreateShard:
			_, err := s.resource.createShard(
				fmt.Sprintf("%s.%d.%d", rsp.Namespace, rsp.GroupID, rsp.ReplicaID),
				s.addr,
				s.diskq,
			)
			if err != nil {
				log.Error(err)
				continue
			}
			err = stream.Send(&metadpb.RegisterNodeReq{
				Type:      metadpb.RegisterNodeReq_RegisterShard,
				Addr:      addr,
				Namespace: rsp.Namespace,
				GroupID:   rsp.GroupID,
				ReplicaID: rsp.ReplicaID,
			})
			if err != nil {
				log.Error(err)
			}
		case metadpb.RegisterNodeRsp_RemoveShard:
			id := fmt.Sprintf("%s.%s.%s", rsp.Namespace, rsp.GroupID, rsp.ReplicaID)
			if err := s.resource.removeShard(id); err != nil {
				log.Error(err)
			}
			err := stream.Send(&metadpb.RegisterNodeReq{
				Type:      metadpb.RegisterNodeReq_UnregisterShard,
				Addr:      addr,
				Namespace: rsp.Namespace,
				GroupID:   rsp.GroupID,
				ReplicaID: rsp.ReplicaID,
			})
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func (s *service) keepAliveWithMetad(
	addr string,
	stream metadpb.Metad_RegisterNodeClient) {

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			err := stream.Send(&metadpb.RegisterNodeReq{
				Type: metadpb.RegisterNodeReq_Heartbeat,
				Addr: addr,
			})
			if err != nil {
				log.Error(err)
			}
		case <-s.exitC:
			goto exit
		}
	}
exit:
	ticker.Stop()
}

func (s *service) Search(ctx context.Context,
	req *searchdpb.SearchReq) (*searchdpb.SearchRsp, error) {

	shard := ctx.Value(shardContextKey).(*shard)
	docs, err := shard.indexer.Search(req.Query, req.Start, int32(req.Count))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Infof("Type=Search\tTraceID=%s\tQuery=%s\tShard=%s\tCount=%d",
		req.TraceID, req.Query.Raw, shard.id, len(docs))
	return &searchdpb.SearchRsp{
		Docs: docs,
	}, nil
}

func (s *service) Close() {
	s.resource.close()
}
