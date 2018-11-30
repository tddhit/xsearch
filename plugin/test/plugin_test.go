package plugin

import (
	"testing"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
	"github.com/tddhit/xsearch/plugin"
	_ "github.com/tddhit/xsearch/plugin/classifier"
	_ "github.com/tddhit/xsearch/plugin/segmenter"
)

func TestPlugin(t *testing.T) {
	args := &xsearchpb.QueryAnalysisArgs{
		Queries: []*xsearchpb.Query{
			{Raw: "我是一个程序员"},
		},
	}
	if err := plugin.Analyze(args); err != nil {
		log.Fatal(err)
	}
	for _, query := range args.Queries {
		log.Debugf("%s,%+v", query.Raw, query.Tags)
		for _, token := range query.Tokens {
			log.Debug(token.Term)
		}
	}
}
