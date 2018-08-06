package service

import (
	"context"

	"github.com/tddhit/hunter/builder"
	pb "github.com/tddhit/xsearch/searchd/pb"
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
	in *pb.AddDocRequest) (*pb.AddDocReply, error) {

	s.engine.AddDoc(in.GetDoc())
}

func (s *service) DeleteDoc(ctx context.Context,
	in *pb.DeleteDocRequest) (*pb.DeleteDocReply, error) {

}

func (s *service) Search(ctx context.Context,
	in *pb.SearchRequest) (*pb.SearchReply, error) {

}
