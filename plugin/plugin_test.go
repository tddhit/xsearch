package plugin

import (
	"testing"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
)

func TestPlugin(t *testing.T) {
	if err := Init("./testso"); err != nil {
		log.Fatal(err)
	}
	args := &xsearchpb.QueryAnalysisArgs{
		Queries: []*xsearchpb.Query{
			{Raw: "我是一个程序员"},
		},
	}
	if err := Analyze(args); err != nil {
		log.Fatal(err)
	}
	for _, query := range args.Queries {
		log.Debug(query.Raw)
		for _, token := range query.Tokens {
			log.Debug(token.Term)
		}
	}
}
