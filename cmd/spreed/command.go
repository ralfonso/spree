package main

import "github.com/codegangsta/cli"

var (
	addrFlag = cli.StringFlag{
		Name:  "addr",
		Value: "localhost:4285",
		Usage: "host:port for the http server",
	}
	dataDirFlag = cli.StringFlag{
		Name:  "data.dir",
		Value: "",
		Usage: "The directory in which to store uploaded files",
	}
	authTokenFlag = cli.StringFlag{
		Name:  "auth.token",
		Value: "",
		Usage: "The token to use in the auth header",
	}
	urlPrefixFlag = cli.StringFlag{
		Name:  "url.prefix",
		Value: "http:/localhost:4285",
		Usage: "The partial URL to use for hosting files",
	}
	dbFileFlag = cli.StringFlag{
		Name:  "db.file",
		Value: "",
		Usage: "The file to use for the database",
	}
	dbBucketFlag = cli.StringFlag{
		Name:  "db.bucket",
		Value: "spree",
		Usage: "The bucket to use in the database",
	}
)

var GlobalFlags = []cli.Flag{
	addrFlag,
	dataDirFlag,
	urlPrefixFlag,
	dbFileFlag,
	dbBucketFlag,
	authTokenFlag,
}
