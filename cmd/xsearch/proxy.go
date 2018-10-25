package main

import (
	"github.com/urfave/cli"

	"github.com/tddhit/box/mw"
	"github.com/tddhit/box/transport"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/proxy"
	"github.com/tddhit/xsearch/proxy/pb"
)

var proxyCommand = cli.Command{
	Name:      "proxy",
	Usage:     "start proxy service",
	Action:    withLog(startProxy),
	UsageText: "xsearch proxy [arguments...]",
	Flags: []cli.Flag{
		pidFlag,
		logPathFlag,
		logLevelFlag,
		cli.StringFlag{
			Name:  "addr",
			Usage: "listen address",
			Value: "127.0.0.1:10300",
		},
		cli.StringFlag{
			Name:  "metad",
			Usage: "metad addr",
		},
		cli.StringFlag{
			Name:  "namespaces",
			Usage: "Business name separated by a comma. i.e. faq,news",
		},
	},
}

func startProxy(ctx *cli.Context) {
	addr := ctx.String("addr")
	pidPath := ctx.String("pidpath")
	svc := proxy.NewService(ctx)
	server, err := transport.Listen(
		"grpc://" + addr,
	)
	if err != nil {
		log.Fatal(err)
	}
	server.Register(proxypb.ProxyGrpcServiceDesc, svc)
	mw.Run(mw.WithServer(server), mw.WithPIDPath(pidPath))
}
