package main

import (
	"context"

	"github.com/urfave/cli"

	"github.com/tddhit/box/transport"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/metad/pb"
)

var clientCommand = cli.Command{
	Name:      "client",
	Usage:     "client tools for manipulating clusters",
	Action:    startClient,
	UsageText: "xsearch client [create|drop|add|remove|balance|migrate|info|commit] [arguments...]",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "metad",
			Usage: "meatd address",
			Value: "grpc://127.0.0.1:10100",
		},
		cli.StringFlag{
			Name:  "n",
			Usage: "namespace",
		},
		cli.StringFlag{
			Name:  "g",
			Usage: "groupID",
		},
		cli.StringFlag{
			Name:  "r",
			Usage: "replicaID",
		},
		cli.StringFlag{
			Name:  "f",
			Usage: "fromNode address",
		},
		cli.StringFlag{
			Name:  "t",
			Usage: "toNode address",
		},
		cli.StringFlag{
			Name:  "a",
			Usage: "node address",
		},
	},
}

func startClient(params *cli.Context) {
	conn, err := transport.Dial(params.String("metad"))
	if err != nil {
		log.Fatal(err)
	}
	client := metadpb.NewMetadGrpcClient(conn)
	switch params.Args().First() {
	case "create":
		execCreate(params, client)
	case "drop":
		execDrop(params, client)
	case "add":
		execAdd(params, client)
	case "remove":
		execRemove(params, client)
	case "balance":
		execBalance(params, client)
	case "migrate":
		execMigrate(params, client)
	case "info":
		execInfo(params, client)
	case "commmit":
		execCommit(params, client)
	}
}

func execCreate(params *cli.Context, client metadpb.MetadGrpcClient) {
	log.Debug(params.String("n"))
	_, err := client.CreateNamespace(
		context.Background(),
		&metadpb.CreateNamespaceReq{
			Namespace:     params.String("n"),
			ShardNum:      uint32(params.Int("s")),
			ReplicaFactor: uint32(params.Int("r")),
		},
	)
	if err != nil {
		log.Error(err)
	}
}

func execDrop(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.DropNamespace(
		context.Background(),
		&metadpb.DropNamespaceReq{
			Namespace: params.String("n"),
		},
	)
	if err != nil {
		log.Error(err)
	}
}

func execAdd(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.AddNodeToNamespace(
		context.Background(),
		&metadpb.AddNodeToNamespaceReq{
			Namespace: params.String("n"),
			Addr:      params.String("a"),
		},
	)
	if err != nil {
		log.Error(err)
	}
}

func execRemove(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.RemoveNodeFromNamespace(
		context.Background(),
		&metadpb.RemoveNodeFromNamespaceReq{
			Namespace: params.String("n"),
			Addr:      params.String("a"),
		},
	)
	if err != nil {
		log.Error(err)
	}
}

func execBalance(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.AutoBalance(
		context.Background(),
		&metadpb.AutoBalanceReq{
			Namespace: params.String("n"),
		},
	)
	if err != nil {
		log.Error(err)
	}
}

func execMigrate(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.MigrateShard(
		context.Background(),
		&metadpb.MigrateShardReq{
			Namespace: params.String("n"),
			GroupID:   uint32(params.Int("g")),
			ReplicaID: uint32(params.Int("r")),
			FromNode:  params.String("f"),
			ToNode:    params.String("t"),
		},
	)
	if err != nil {
		log.Error(err)
	}
}

func execInfo(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.Info(
		context.Background(),
		&metadpb.InfoReq{
			Namespace: params.String("n"),
		},
	)
	if err != nil {
		log.Error(err)
	}
}

func execCommit(params *cli.Context, client metadpb.MetadGrpcClient) {
	_, err := client.Commit(
		context.Background(),
		&metadpb.CommitReq{
			Namespace: params.String("n"),
		},
	)
	if err != nil {
		log.Error(err)
	}
}
