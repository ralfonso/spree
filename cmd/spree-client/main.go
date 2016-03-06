package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "spree-client"
	app.Usage = "upload stuff"
	app.Commands = Commands
	app.Run(os.Args)
}
