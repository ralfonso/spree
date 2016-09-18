package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/codegangsta/cli"
	"github.com/ralfonso/spree"
	"github.com/uber-go/zap"
)

var Version = "0.2.0"

func main() {
	rand.Seed(time.Now().UnixNano())
	app := cli.NewApp()
	app.Flags = GlobalFlags
	app.Name = "spree"
	app.Usage = "upload stuff"
	app.Action = serve
	app.Run(os.Args)
}

func serve(ctx *cli.Context) {
	ll := zap.New(zap.NewJSONEncoder())
	boltKV, err := spree.NewBoltKV(ctx.String(dbFileFlag.Name), ctx.String(dbBucketFlag.Name), ll)
	if err != nil {
		ll.Fatal("unable to create BoltKV", zap.Error(err))
	}

	dataDir := ctx.String(dataDirFlag.Name)
	store, err := spree.NewFileStorage(dataDir)
	if err != nil {
		ll.Fatal("unable to create FileStore", zap.Error(err))
	}

	server := spree.NewServer(boltKV, store, ll)
	rpcAddr := ctx.GlobalString(rpcAddrFlag.Name)
	caCertFile := ctx.GlobalString(caCertFileFlag.Name)
	certFile := ctx.GlobalString(certFileFlag.Name)
	keyFile := ctx.GlobalString(keyFileFlag.Name)

	if caCertFile == "" || certFile == "" || keyFile == "" {
		ll.Fatal("must have CA cert, server cert, and server key")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		ll.Fatal("failed to load keypair", zap.Error(err))
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	certPool := x509.NewCertPool()
	caContents, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		ll.Fatal("unable to read CA certificate file",
			zap.String("file", caCertFile),
			zap.Error(err))
	}
	certPool.AppendCertsFromPEM(caContents)
	tlsConfig.ClientCAs = certPool
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

	lis, err := tls.Listen("tcp", rpcAddr, tlsConfig)
	if err != nil {
		ll.Fatal("failed to listen", zap.Error(err))
	}
	grpcServer := grpc.NewServer()
	spree.RegisterSpreeServer(grpcServer, server)
	ll.Info("Starting RPC server", zap.String("rpc.addr", rpcAddr))
	go grpcServer.Serve(lis)

	httpAddr := ctx.String(httpAddrFlag.Name)
	httpServer := spree.NewHTTPServer(httpAddr, boltKV, store, ll)
	go httpServer.Run()
	select {}
}
