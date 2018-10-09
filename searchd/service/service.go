package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tddhit/xsearch/searchd/engine"
	"github.com/tddhit/xsearch/searchd/searchdpb"
)

type service struct {
	engine *engine.Engine
}

func NewService() *service {
	s := &service{
		engine: engine.New(),
	}
	return s
}

func (s *service) AddDoc(ctx context.Context,
	in *searchdpb.AddDocRequest) (*searchdpb.AddDocReply, error) {

	err := s.engine.AddDoc(in.GetDoc())
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	return nil, nil
}

func (s *service) RemoveDoc(ctx context.Context,
	in *searchdpb.RemoveDocRequest) (*searchdpb.RemoveDocReply, error) {

	s.engine.RemoveDoc(in.GetDoc())
	return nil, nil
}

func (s *service) Search(ctx context.Context,
	in *searchdpb.SearchRequest) (*searchdpb.SearchReply, error) {

	docs, err := s.engine.Search(in.GetQuery(), in.GetStart(), in.GetCount())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &searchdpb.SearchReply{
		Docs: docs,
	}, nil
}
