package main

import "github.com/codegangsta/cli"

var (
	caCertFileFlag = cli.StringFlag{
		Name:   "ca.cert.file",
		Value:  "/etc/spree/certs/spree.ca.crt",
		Usage:  "CA cert file for TLS server",
		EnvVar: "SPREE_CACERT_FILE",
	}
	certFileFlag = cli.StringFlag{
		Name:   "cert.file",
		Value:  "/etc/spree/certs/spree.server.crt",
		Usage:  "cert file for TLS server",
		EnvVar: "SPREE_CERT_FILE",
	}
	keyFileFlag = cli.StringFlag{
		Name:   "key.file",
		Value:  "/etc/spree/certs/spree.server.key",
		Usage:  "key file for TLS server",
		EnvVar: "SPREE_KEY_FILE",
	}
	rpcAddrFlag = cli.StringFlag{
		Name:   "rpc.addr",
		Value:  "localhost:4285",
		Usage:  "host:port for the grpc server",
		EnvVar: "SPREE_RPC_ADDR",
	}
	httpAddrFlag = cli.StringFlag{
		Name:   "http.addr",
		Value:  "localhost:8383",
		Usage:  "host:port for the http server",
		EnvVar: "SPREE_HTTP_ADDR",
	}
	dataDirFlag = cli.StringFlag{
		Name:   "data.dir",
		Value:  "/tmp",
		Usage:  "The directory in which to store uploaded files",
		EnvVar: "SPREE_DATA_DIR",
	}
	dbFileFlag = cli.StringFlag{
		Name:   "db.file",
		Value:  "/tmp/spree.boltdb",
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
	caCertFileFlag,
	certFileFlag,
	keyFileFlag,
	rpcAddrFlag,
	httpAddrFlag,
	dataDirFlag,
	dbFileFlag,
	dbBucketFlag,
}
