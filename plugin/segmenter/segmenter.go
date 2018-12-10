package plugin

import (
	"github.com/yanyiwu/gojieba"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
	"github.com/tddhit/xsearch/plugin"
)

func init() {
	if err := plugin.Register(&segmenter{}); err != nil {
		log.Fatal(err)
	}
}

type segmenter struct {
	*gojieba.Jieba
	initial bool
}

func (s *segmenter) Init() error {
	s.initial = true
	s.Jieba = gojieba.NewJieba()
	return nil
}

func (s *segmenter) Type() int8 {
	return plugin.TYPE_ANALYSIS
}

func (s *segmenter) Name() string {
	return "segmenter"
}

func (s *segmenter) Priority() int8 {
	return 2
}

func (s *segmenter) Analyze(args *xsearchpb.QueryAnalysisArgs) error {
	if !s.initial {
		return nil
	}
	for _, term := range s.Cut(args.Queries[0].Raw, true) {
		args.Queries[0].Tokens = append(
			args.Queries[0].Tokens,
			&xsearchpb.Token{Term: term},
		)
	}
	return nil
}

func (s *segmenter) Rerank(args *xsearchpb.RerankArgs) error {
	return nil
}

func (s *segmenter) Cleanup() error {
	return nil
}
