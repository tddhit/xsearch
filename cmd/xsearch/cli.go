package main

import (
	"context"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/urfave/cli"

	"github.com/tddhit/box/transport"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/metad/pb"
	"github.com/tddhit/xsearch/pb"
	"github.com/tddhit/xsearch/proxy/pb"
	"github.com/tddhit/xsearch/searchd/pb"
)

var clientCommand = cli.Command{
	Name:   "client",
	Usage:  "client tools for manipulating clusters",
	Action: startClient,
	UsageText: `xsearch client service operation [arguments...]

service list:
metad -> operate cluster
searchd -> specify sharding to query the document
proxy -> index or query documents

metad operation list:
create -> create namespace, require namespace/shardnum/factor
drop -> drop namespace, require namespace
add -> add node to namespace, require namespace/node
remove -> remove node from namespace, require namespace/node
balance -> auto balance, require namespace
migrate -> migrate shard, require namespace/groupid/replicaid/from/to
info ->  node/namespace information, no args
commit -> commit operation, require namespace

searchd operation list:
search -> search query, require shardid/query/start/count
info -> shards information, no args

proxy operation list:
index -> build indexes, require namespace/content
remove -> remove document from indexes, require namespace/docid
search -> search query, require namespace/query/start/count
info -> shard tables information, no args

`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Usage: "meatd/searchd/proxy address",
			Value: "grpc://127.0.0.1:10100",
		},
		cli.StringFlag{
			Name: "namespace",
		},
		cli.IntFlag{
			Name: "groupid",
		},
		cli.IntFlag{
			Name:  "shardnum",
			Value: 3,
		},
		cli.IntFlag{
			Name: "replicaid",
		},
		cli.IntFlag{
			Name:  "factor",
			Usage: "replica factor",
			Value: 2,
		},
		cli.StringFlag{
			Name:  "from",
			Usage: "node address",
		},
		cli.StringFlag{
			Name:  "to",
			Usage: "node address",
		},
		cli.StringFlag{
			Name:  "node",
			Usage: "node address",
		},
		cli.StringFlag{
			Name: "shardid",
		},
		cli.StringFlag{
			Name: "docid",
		},
		cli.StringFlag{
			Name: "content",
		},
		cli.StringFlag{
			Name: "query",
		},
		cli.IntFlag{
			Name: "start",
		},
		cli.IntFlag{
			Name: "count",
		},
	},
}

func startClient(params *cli.Context) {
	conn, err := transport.Dial(params.String("addr"))
	if err != nil {
		log.Fatal(err)
	}
	switch params.Args().First() {
	case "metad":
		connectMetad(params, conn)
	case "searchd":
		connectSearchd(params, conn)
	case "proxy":
		connectProxy(params, conn)
	default:
		log.Fatal("invalid service")
	}
	println("ok")
}

func connectMetad(params *cli.Context, conn transport.ClientConn) {
	client := metadpb.NewMetadGrpcClient(conn)
	switch params.Args().Get(1) {
	case "create":
		execMetadCreate(params, client)
	case "drop":
		execMetadDrop(params, client)
	case "add":
		execMetadAdd(params, client)
	case "remove":
		execMetadRemove(params, client)
	case "balance":
		execMetadBalance(params, client)
	case "migrate":
		execMetadMigrate(params, client)
	case "info":
		execMetadInfo(params, client)
	case "commit":
		execMetadCommit(params, client)
	default:
		log.Fatal("invalid metad operation")
	}
}

func connectSearchd(params *cli.Context, conn transport.ClientConn) {
	client := searchdpb.NewSearchdGrpcClient(conn)
	adminClient := searchdpb.NewAdminGrpcClient(conn)
	switch params.Args().Get(1) {
	case "search":
		execSearchdSearch(params, client)
	case "info":
		execSearchdInfo(params, adminClient)
	default:
		log.Fatal("invalid searchd operation")
	}
}

func connectProxy(params *cli.Context, conn transport.ClientConn) {
	client := proxypb.NewProxyGrpcClient(conn)
	adminClient := proxypb.NewAdminGrpcClient(conn)
	switch params.Args().Get(1) {
	case "index":
		execProxyIndex(params, client)
	case "remove":
		execProxyRemove(params, client)
	case "search":
		execProxySearch(params, client)
	case "info":
		execProxyInfo(params, adminClient)
	default:
		log.Fatal("invalid proxy operation")
	}
}

func execMetadCreate(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.CreateNamespace(
		context.Background(),
		&metadpb.CreateNamespaceReq{
			Namespace:     params.String("namespace"),
			ShardNum:      uint32(params.Int("shardnum")),
			ReplicaFactor: uint32(params.Int("factor")),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func execMetadDrop(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.DropNamespace(
		context.Background(),
		&metadpb.DropNamespaceReq{
			Namespace: params.String("namespace"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func execMetadAdd(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.AddNodeToNamespace(
		context.Background(),
		&metadpb.AddNodeToNamespaceReq{
			Namespace: params.String("namespace"),
			Addr:      params.String("node"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func execMetadRemove(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.RemoveNodeFromNamespace(
		context.Background(),
		&metadpb.RemoveNodeFromNamespaceReq{
			Namespace: params.String("namespace"),
			Addr:      params.String("node"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func execMetadBalance(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.AutoBalance(
		context.Background(),
		&metadpb.AutoBalanceReq{
			Namespace: params.String("namespace"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func execMetadMigrate(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.MigrateShard(
		context.Background(),
		&metadpb.MigrateShardReq{
			Namespace: params.String("namespace"),
			GroupID:   uint32(params.Int("groupid")),
			ReplicaID: uint32(params.Int("replicaid")),
			FromNode:  params.String("from"),
			ToNode:    params.String("to"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func execMetadInfo(params *cli.Context, client metadpb.MetadGrpcClient) {
	rsp, err := client.Info(
		context.Background(),
		&metadpb.InfoReq{},
	)
	if err != nil {
		log.Fatal(err)
	}
	print(rsp)
}

func execMetadCommit(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.Commit(
		context.Background(),
		&metadpb.CommitReq{
			Namespace: params.String("namespace"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func execSearchdSearch(params *cli.Context, client searchdpb.SearchdGrpcClient) {
	rsp, err := client.Search(
		context.Background(),
		&searchdpb.SearchReq{
			ShardID: params.String("shardid"),
			Query: &xsearchpb.Query{
				Raw: params.String("query"),
			},
			Start: uint64(params.Int("start")),
			Count: uint32(params.Int("count")),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	print(rsp)
}

func execSearchdInfo(params *cli.Context, client searchdpb.AdminGrpcClient) {
	rsp, err := client.Info(
		context.Background(),
		&searchdpb.InfoReq{},
	)
	if err != nil {
		log.Fatal(err)
	}
	print(rsp)
}

func execProxyIndex(params *cli.Context, client proxypb.ProxyGrpcClient) {
	rsp, err := client.IndexDoc(
		context.Background(),
		&proxypb.IndexDocReq{
			Namespace: params.String("namespace"),
			Doc: &xsearchpb.Document{
				Content: params.String("content"),
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	println("docID:", rsp.DocID)
}

func execProxyRemove(params *cli.Context, client proxypb.ProxyGrpcClient) {
	_, err := client.RemoveDoc(
		context.Background(),
		&proxypb.RemoveDocReq{
			Namespace: params.String("namespace"),
			DocID:     params.String("docid"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func execProxySearch(params *cli.Context, client proxypb.ProxyGrpcClient) {
	rsp, err := client.Search(
		context.Background(),
		&proxypb.SearchReq{
			Namespace: params.String("namespace"),
			Query: &xsearchpb.Query{
				Raw: params.String("query"),
			},
			Start: uint64(params.Int("start")),
			Count: uint32(params.Int("count")),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	print(rsp)
}

func execProxyInfo(params *cli.Context, client proxypb.AdminGrpcClient) {
	rsp, err := client.Info(
		context.Background(),
		&proxypb.InfoReq{},
	)
	if err != nil {
		log.Fatal(err)
	}
	print(rsp)
}

func print(rsp proto.Message) {
	marshaler := &jsonpb.Marshaler{
		EnumsAsInts: true,
		Indent:      "    ",
	}
	if s, err := marshaler.MarshalToString(rsp); err != nil {
		log.Fatal(err)
	} else {
		println(s)
	}
}
