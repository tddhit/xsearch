package main

import (
	"os"

	"github.com/tddhit/tools/log"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "xsearch"
	app.Usage = "xsearch command-line tool"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		metadCommand,
		searchdCommand,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
