package searchd

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/wangbin/jiebago"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/box/transport"
	"github.com/tddhit/diskqueue/pb"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/metad/pb"
	"github.com/tddhit/xsearch/pb"
	"github.com/tddhit/xsearch/searchd/pb"
)

type service struct {
	addr      string
	resource  *Resource
	metad     metadpb.MetadGrpcClient
	diskq     diskqueuepb.DiskqueueGrpcClient
	segmenter *jiebago.Segmenter
	stopwords map[string]struct{}
	exitC     chan struct{}
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
	segmenter := &jiebago.Segmenter{}
	if err := segmenter.LoadDictionary(ctx.String("dict")); err != nil {
		log.Fatal(err)
	}
	if err := segmenter.LoadUserDictionary(ctx.String("userdict")); err != nil {
		log.Fatal(err)
	}
	stopwords := make(map[string]struct{})
	file, err := os.Open(ctx.String("stopdict"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		stopwords[strings.TrimSpace(line)] = struct{}{}
	}
	s := &service{
		addr:      ctx.String("addr"),
		resource:  r,
		metad:     metad,
		diskq:     diskq,
		segmenter: segmenter,
		stopwords: stopwords,
		exitC:     make(chan struct{}),
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
			log.Debug("create shard")
			_, err := s.resource.createShard(
				rsp.ShardID,
				s.addr,
				s.segmenter,
				s.stopwords,
				s.diskq,
			)
			if err != nil {
				log.Error(err)
				continue
			}
			log.Debug("register shard")
			err = stream.Send(&metadpb.RegisterNodeReq{
				Type:    metadpb.RegisterNodeReq_RegisterShard,
				Addr:    addr,
				ShardID: rsp.ShardID,
			})
			if err != nil {
				log.Error(err)
			}
		case metadpb.RegisterNodeRsp_RemoveShard:
			log.Debug("remove shard")
			if err := s.resource.removeShard(rsp.ShardID); err != nil {
				log.Error(err)
			}
			log.Debug("unregister shard")
			err := stream.Send(&metadpb.RegisterNodeReq{
				Type:    metadpb.RegisterNodeReq_UnregisterShard,
				Addr:    addr,
				ShardID: rsp.ShardID,
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

	log.Debug("search", req.Query.Raw)
	shard := ctx.Value(shardContextKey).(*shard)
	for term := range s.segmenter.Cut(req.Query.Raw, true) {
		if _, ok := s.stopwords[term]; !ok {
			log.Debug(term)
			req.Query.Tokens = append(
				req.Query.Tokens,
				&xsearchpb.Token{Term: term},
			)
		}
	}
	docs, err := shard.indexer.Search(req.Query, req.Start, int32(req.Count))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &searchdpb.SearchRsp{
		Docs: docs,
	}, nil
}