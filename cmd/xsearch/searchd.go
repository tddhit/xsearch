package main

import (
	"github.com/urfave/cli"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/box/transport"
	tropt "github.com/tddhit/box/transport/option"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/searchd"
	"github.com/tddhit/xsearch/searchd/pb"
)

var searchdCommand = cli.Command{
	Name:      "searchd",
	Usage:     "start searchd service",
	Action:    withLog(startSearchd),
	UsageText: "xsearch searchd [arguments...]",
	Flags: []cli.Flag{
		pidFlag,
		logPathFlag,
		logLevelFlag,
		cli.StringFlag{
			Name:  "addr",
			Usage: "listen address",
			Value: "127.0.0.1:10200",
		},
		cli.StringFlag{
			Name:  "admin",
			Usage: "admin address",
			Value: "127.0.0.1:10201",
		},
		cli.StringFlag{
			Name:  "metad",
			Usage: "metad addr",
			Value: "127.0.0.1:10100",
		},
		cli.StringFlag{
			Name:  "diskqueue",
			Usage: "diskqueue addr",
		},
		cli.StringFlag{
			Name:  "dict",
			Usage: "segment dict",
			Value: "./dict/segment.dict",
		},
		cli.StringFlag{
			Name:  "userdict",
			Usage: "user defined segment dict",
			Value: "./dict/user_segment.dict",
		},
		cli.StringFlag{
			Name:  "stopdict",
			Usage: "stopwords dict",
			Value: "./dict/stopwords.dict",
		},
		cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "./data",
		},
	},
}

func startSearchd(ctx *cli.Context) {
	var (
		resource *searchd.Resource
		err      error
	)
	if mw.IsWorker() {
		resource, err = searchd.NewResource(ctx.String("datadir"))
		if err != nil {
			log.Fatal(err)
		}
	}

	svc := searchd.NewService(ctx, resource)
	server, err := transport.Listen(
		"grpc://"+ctx.String("addr"),
		tropt.WithUnaryServerMiddleware(
			searchd.CheckParams(svc),
		),
		tropt.WithAfterClose(svc.Close),
	)
	if err != nil {
		log.Fatal(err)
	}
	server.Register(searchdpb.SearchdGrpcServiceDesc, svc)

	admin := searchd.NewAdmin(resource)
	adminServer, err := transport.Listen("grpc://" + ctx.String("admin"))
	if err != nil {
		log.Fatal(err)
	}
	adminServer.Register(searchdpb.AdminGrpcServiceDesc, admin)

	mw.Run(
		mw.WithServer(server),
		mw.WithServer(adminServer),
		mw.WithPIDPath(ctx.String("pidpath")),
	)
}
