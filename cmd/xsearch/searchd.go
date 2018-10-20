package main

import (
	"github.com/tddhit/box/mw"
	"github.com/tddhit/box/transport"
	tropt "github.com/tddhit/box/transport/option"
	"github.com/tddhit/tools/log"
	"github.com/urfave/cli"

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
			Name:  "diskqueue",
			Usage: "diskqueue addr",
		},
		cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "./data",
		},
	},
}

func startSearchd(ctx *cli.Context) {
	addr := ctx.String("addr")
	if addr == "" {
		addr = "127.0.0.1:10200"
	}
	pidPath := ctx.String("pidpath")
	svc := searchd.NewService(ctx)
	server, err := transport.Listen(
		"grpc://"+addr,
		tropt.WithUnaryServerMiddleware(
			searchd.CheckParams(svc),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	server.Register(searchdpb.SearchdGrpcServiceDesc, svc)
	mw.Run(mw.WithServer(server), mw.WithPIDPath(pidPath))
}
