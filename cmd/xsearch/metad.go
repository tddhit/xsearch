package main

import (
	"github.com/urfave/cli"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/box/transport"
	tropt "github.com/tddhit/box/transport/option"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/metad"
	"github.com/tddhit/xsearch/metad/pb"
)

var metadCommand = cli.Command{
	Name:      "metad",
	Usage:     "start metad service",
	Action:    withLog(startMetad),
	UsageText: "xsearch metad [arguments...]",
	Flags: []cli.Flag{
		pidFlag,
		logPathFlag,
		logLevelFlag,
		cli.StringFlag{
			Name:  "addr",
			Usage: "listen address",
			Value: "127.0.0.1:10100",
		},
		cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "./data",
		},
	},
}

func startMetad(ctx *cli.Context) {
	addr := ctx.String("addr")
	pidPath := ctx.String("pidpath")
	svc := metad.NewService(ctx)
	server, err := transport.Listen(
		"grpc://"+addr,
		tropt.WithUnaryServerMiddleware(
			metad.CheckParams(svc),
		),
		tropt.WithBeforeClose(svc.Close),
	)
	if err != nil {
		log.Fatal(err)
	}
	server.Register(metadpb.MetadGrpcServiceDesc, svc)
	mw.Run(mw.WithServer(server), mw.WithPIDPath(pidPath))
}
