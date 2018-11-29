package main

import (
	"errors"

	"github.com/yanyiwu/gojieba"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
	"github.com/tddhit/xsearch/plugin"
)

var Object Segmenter

type Segmenter struct {
	*gojieba.Jieba
}

func (s *Segmenter) Init() error {
	s.Jieba = gojieba.NewJieba()
	return nil
}

func (s *Segmenter) Type() int8 {
	return plugin.TYPE_ANALYSIS
}

func (s *Segmenter) Name() string {
	return "Segmenter"
}

func (s *Segmenter) Priority() int8 {
	return 1
}

func (s *Segmenter) Analyze(args *xsearchpb.QueryAnalysisArgs) error {
	for _, term := range s.Cut(args.Queries[0].Raw, true) {
		args.Queries[0].Tokens = append(
			args.Queries[0].Tokens,
			&xsearchpb.Token{Term: term},
		)
	}
	return nil
}

func (s *Segmenter) Rerank(args *xsearchpb.RerankArgs) error {
	return nil
}

func (s *Segmenter) Cleanup() error {
	return nil
}
