package main

import "github.com/codegangsta/cli"

var (
	addrFlag = cli.StringFlag{
		Name:   "addr",
		Value:  "localhost:4285",
		Usage:  "host:port for the http server",
		EnvVar: "SPREE_ADDR",
	}
	dataDirFlag = cli.StringFlag{
		Name:   "data.dir",
		Value:  "",
		Usage:  "The directory in which to store uploaded files",
		EnvVar: "SPREE_DATA_DIR",
	}
	authTokenFlag = cli.StringFlag{
		Name:   "auth.token",
		Value:  "",
		Usage:  "The token to use in the auth header",
		EnvVar: "SPREE_AUTH_TOKEN",
	}
	urlPrefixFlag = cli.StringFlag{
		Name:   "url.prefix",
		Value:  "http://localhost:4285",
		Usage:  "The partial URL to use for hosting files",
		EnvVar: "SPREE_URL_PREFIX",
	}
	dbFileFlag = cli.StringFlag{
		Name:   "db.file",
		Value:  "",
		Usage:  "The file to use for the database",
		EnvVar: "SPREE_DB_FILE",
	}
	dbBucketFlag = cli.StringFlag{
		Name:   "db.bucket",
		Value:  "spree",
		Usage:  "The bucket to use in the database",
		EnvVar: "SPREE_DB_BUCKET",
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
