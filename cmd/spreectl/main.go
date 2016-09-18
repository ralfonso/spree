package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "spree-client"
	app.Usage = "upload stuff"
	app.Flags = GlobalFlags
	app.Commands = Commands
	app.Run(os.Args)
}
