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
			Name:  "metad",
			Usage: "metad addr",
			Value: "127.0.0.1:10100",
		},
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
