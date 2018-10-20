package main

import (
	"github.com/tddhit/tools/log"
	"github.com/urfave/cli"
)

var (
	pidFlag = cli.StringFlag{
		Name:  "pidpath",
		Usage: "(default: ./xxx.pid)",
	}
	logPathFlag = cli.StringFlag{
		Name:  "logpath",
		Usage: "(default: stderr)",
	}
	logLevelFlag = cli.IntFlag{
		Name:  "loglevel",
		Usage: "trace:1, debug:2, info:3, warn:4, error:5",
		Value: 2,
	}
)

func withLog(f func(ctx *cli.Context)) func(ctx *cli.Context) {
	return func(ctx *cli.Context) {
		logPath := ctx.String("logpath")
		logLevel := ctx.Int("loglevel")
		log.Init(logPath, logLevel)
		f(ctx)
	}
}
