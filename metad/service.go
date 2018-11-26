package metad

import (
	"context"
	"io"

	"github.com/urfave/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/metad/pb"
)

type service struct {
	reception *reception
	resource  *resource
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewService(params *cli.Context) *service {
	if !mw.IsWorker() {
		return nil
	}
	dataDir := params.String("datadir")
	ctx, cancel := context.WithCancel(context.Background())
	return &service{
		reception: newReception(),
		resource:  newResource(dataDir),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (s *service) RegisterClient(stream metadpb.Metad_RegisterClientServer) error {
	p, ok := peer.FromContext(stream.Context())
	if !ok || p == nil || p.Addr == nil || p.Addr.String() == "" {
		return status.Error(codes.FailedPrecondition,
			"The client address could not be obtained")
	}
	req, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	client, created := s.reception.getOrCreateClient(
		req.Namespace,
		p.Addr.String(),
	)
	if !created {
		return status.Error(codes.AlreadyExists, req.Namespace+p.Addr.String())
	}
	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				return
			}
			client.readC <- req
		}
	}()
	if table, ok := s.resource.getTable(req.Namespace); ok {
		client.writeC <- &metadpb.RegisterClientRsp{
			Table: table.marshal(),
		}
	}
	client.ioLoop(s.ctx, stream)
	return nil
}

func (s *service) RegisterNode(stream metadpb.Metad_RegisterNodeServer) error {
	req, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	n, err := s.resource.createNode(req.Addr)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	n.readC <- req
	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				log.Error(err)
				return
			}
			n.readC <- req
		}
	}()
	n.ioLoop(s.ctx, stream, s.resource, s.reception)
	s.resource.removeNode(n.addr)
	s.resource.rangeTables(func(table *shardTable) error {
		if _, ok := table.getNode(n.addr); !ok {
			log.Trace(2, "not get node")
			return nil
		}
		log.Trace(2, table.namespace)
		s.reception.notifyByNamespace(table.namespace, table.marshal())
		return nil
	})
	return nil
}

func (s *service) CreateNamespace(
	ctx context.Context,
	req *metadpb.CreateNamespaceReq) (*metadpb.CreateNamespaceRsp, error) {

	err := s.resource.createTable(
		req.Namespace,
		int(req.ShardNum),
		int(req.ReplicaFactor),
	)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return &metadpb.CreateNamespaceRsp{}, nil
}

func (s *service) DropNamespace(
	ctx context.Context,
	req *metadpb.DropNamespaceReq) (*metadpb.DropNamespaceRsp, error) {

	if err := s.resource.removeTable(req.Namespace); err != nil {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
	return &metadpb.DropNamespaceRsp{}, nil
}

func (s *service) AddNodeToNamespace(
	ctx context.Context,
	req *metadpb.AddNodeToNamespaceReq) (*metadpb.AddNodeToNamespaceRsp, error) {

	node, ok := s.resource.getNode(req.Addr)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Addr)
	}
	table := ctx.Value(tableContextKey).(*shardTable)
	if err := table.addNode(node); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &metadpb.AddNodeToNamespaceRsp{}, nil
}

func (s *service) RemoveNodeFromNamespace(
	ctx context.Context,
	req *metadpb.RemoveNodeFromNamespaceReq,
) (*metadpb.RemoveNodeFromNamespaceRsp, error) {

	node, ok := s.resource.getNode(req.Addr)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Addr)
	}
	table := ctx.Value(tableContextKey).(*shardTable)
	table.removeNode(node.addr)
	return &metadpb.RemoveNodeFromNamespaceRsp{}, nil
}

func (s *service) AutoBalance(
	ctx context.Context,
	req *metadpb.AutoBalanceReq) (*metadpb.AutoBalanceRsp, error) {

	table := ctx.Value(tableContextKey).(*shardTable)
	if err := table.autoBalance(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &metadpb.AutoBalanceRsp{}, nil
}

func (s *service) MigrateShard(
	ctx context.Context,
	req *metadpb.MigrateShardReq) (*metadpb.MigrateShardRsp, error) {

	table := ctx.Value(tableContextKey).(*shardTable)
	shard, err := table.getShard(int(req.GroupID), int(req.ReplicaID))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	from, ok := s.resource.getNode(req.FromNode)
	if !ok {
		return nil, status.Error(codes.NotFound, req.FromNode)
	}
	to, ok := s.resource.getNode(req.ToNode)
	if !ok {
		return nil, status.Error(codes.NotFound, req.ToNode)
	}
	if err := table.migrate(shard, from, to); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &metadpb.MigrateShardRsp{}, nil
}

func (s *service) Info(
	ctx context.Context,
	req *metadpb.InfoReq) (*metadpb.InfoRsp, error) {

	res := &metadpb.Resource{}
	s.resource.marshalTo(res)
	return &metadpb.InfoRsp{Resource: res}, nil
}

func (s *service) Commit(
	ctx context.Context,
	req *metadpb.CommitReq) (*metadpb.CommitRsp, error) {

	table := ctx.Value(tableContextKey).(*shardTable)
	if err := table.commit(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	meta := table.marshal()
	s.reception.broadcast(meta)
	return &metadpb.CommitRsp{}, nil
}

func (s *service) Close() {
	s.cancel()
}
