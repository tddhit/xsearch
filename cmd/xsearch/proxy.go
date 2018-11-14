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
			Name:  "admin",
			Usage: "admin address",
			Value: "127.0.0.1:10301",
		},
		cli.StringFlag{
			Name:  "metad",
			Usage: "metad addr",
		},
		cli.StringFlag{
			Name:  "diskqueue",
			Usage: "diskqueue addr",
		},
		cli.StringFlag{
			Name:  "namespaces",
			Usage: "Business name separated by a comma. i.e. faq,news",
		},
		cli.StringFlag{
			Name:  "sodir",
			Usage: "query analysis and rerank algorithm plugin directory",
			Value: "./plugin",
		},
	},
}

func startProxy(ctx *cli.Context) {
	var resource *proxy.Resource
	if mw.IsWorker() {
		resource = proxy.NewResource()
	}

	svc := proxy.NewService(ctx, resource)
	server, err := transport.Listen("grpc://" + ctx.String("addr"))
	if err != nil {
		log.Fatal(err)
	}
	server.Register(proxypb.ProxyGrpcServiceDesc, svc)

	admin := proxy.NewAdmin(resource)
	adminServer, err := transport.Listen("grpc://" + ctx.String("admin"))
	if err != nil {
		log.Fatal(err)
	}
	adminServer.Register(proxypb.AdminGrpcServiceDesc, admin)

	mw.Run(
		mw.WithServer(server),
		mw.WithServer(adminServer),
		mw.WithPIDPath(ctx.String("pidpath")),
	)
}
