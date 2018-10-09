package service

import (
	"context"
	"io"
	"sync"

	"github.com/tddhit/tools/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	metadpb "github.com/tddhit/xsearch/metad/pb"
)

type service struct {
	reception *reception
	resource  *resource
}

func New() *service {
	return &service{
		reception: newReception(),
		resource:  newResource(),
	}
}

func (s *service) RegisterNode(stream metadpb.Metad_RegisterNodeServer) error {
	req, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	node, created, err := s.resource.getOrCreateNode(req.Addr, req.AdminAddr)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if !created {
		return status.Error(codes.AlreadyExists, req.Addr)
	}
	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				log.Error(err)
				return
			}
			node.readC <- req
		}
	}()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		node.readLoop()
		wg.Done()
	}()
	go func() {
		node.writeLoop(stream)
		wg.Done()
	}()
	wg.Wait()
	return nil
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
				log.Error(err)
				return
			}
			client.readC <- req
		}
	}()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		client.readLoop()
		wg.Done()
	}()
	go func() {
		client.writeLoop(stream)
		wg.Done()
	}()
	wg.Wait()
	return nil
}

func (s *service) CreateNamespace(
	ctx context.Context,
	req *metadpb.CreateNamespaceReq) (*metadpb.CreateNamespaceRsp, error) {

	_, created := s.resource.getOrCreateTable(
		req.Namespace,
		int(req.ShardNum),
		int(req.ReplicaFactor),
	)
	if !created {
		return nil, status.Error(codes.AlreadyExists, req.Namespace)
	}
	return &metadpb.CreateNamespaceRsp{}, nil
}

func (s *service) DropNamesapce(
	ctx context.Context,
	req *metadpb.DropNamespaceReq) (*metadpb.DropNamespaceRsp, error) {

	if err := s.resource.removeTable(req.Namespace); err != nil {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
	return &metadpb.DropNamespaceRsp{}, nil
}

func (s *service) CommitNamesapce(
	ctx context.Context,
	req *metadpb.CommitNamespaceReq) (*metadpb.CommitNamespaceRsp, error) {

	if err := s.resource.commit(req.Namespace); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &metadpb.CommitNamespaceRsp{}, nil
}

func (s *service) AddNodeToNamespace(
	ctx context.Context,
	req *metadpb.AddNodeToNamespaceReq) (*metadpb.AddNodeToNamespaceRsp, error) {

	node, ok := s.resource.getNode(req.Addr)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Addr)
	}
	table, ok := s.resource.getTable(req.Namespace)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
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
	table, ok := s.resource.getTable(req.Namespace)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
	if err := table.removeNode(node); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &metadpb.RemoveNodeFromNamespaceRsp{}, nil
}

func (s *service) ReplaceNodeInNamespace(
	ctx *context.Context,
	req *metadpb.ReplaceNodeInNamespaceReq,
) (*metadpb.ReplaceNodeInNamespaceRsp, error) {

	old, ok := s.resource.getNode(req.OldAddr)
	if !ok {
		return nil, status.Error(codes.NotFound, req.OldAddr)
	}
	new, ok := s.resource.getNode(req.NewAddr)
	if !ok {
		return nil, status.Error(codes.NotFound, req.NewAddr)
	}
	table, ok := s.resource.getTable(req.Namespace)
	if !ok {
		return nil, status.Error(codes.NotFound, req.Namespace)
	}
	if err := table.replaceNode(old, new); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &metadpb.ReplaceNodeInNamespaceRsp{}, nil
}
