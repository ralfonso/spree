package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/codegangsta/cli"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/ralfonso/spree"
	"github.com/uber-go/zap"
)

var (
	// global flags
	rpcAddrFlag = cli.StringFlag{
		Name:  "rpc.addr",
		Value: "localhost:4285",
		Usage: "Addr for the remote RPC server",
	}

	caCertFileFlag = cli.StringFlag{
		Name:   "ca.cert.file",
		Value:  "spree.dev.ca.crt",
		Usage:  "CA cert file for TLS client",
		EnvVar: "SPREE_CACERT_FILE",
	}
	certFileFlag = cli.StringFlag{
		Name:   "cert.file",
		Value:  "spreectl.crt",
		Usage:  "cert file for TLS client",
		EnvVar: "SPREE_CERT_FILE",
	}
	keyFileFlag = cli.StringFlag{
		Name:   "key.file",
		Value:  "spreectl.key",
		Usage:  "key file for TLS client",
		EnvVar: "SPREE_KEY_FILE",
	}

	// subcommand flags
	srcFlag = cli.StringFlag{
		Name:  "src",
		Value: "-",
		Usage: "The src file to upload. \"-\" for stdin",
	}
	filenameFlag = cli.StringFlag{
		Name:  "file",
		Value: "",
		Usage: "The file to upload",
	}
)

var GlobalFlags = []cli.Flag{
	rpcAddrFlag,
	caCertFileFlag,
	certFileFlag,
	keyFileFlag,
}

var (
	// commands
	uploadCmd = cli.Command{
		Name:   "upload",
		Usage:  "upload to the server",
		Action: UploadCommand,
		Flags: []cli.Flag{
			srcFlag,
			filenameFlag,
		},
	}
	listCmd = cli.Command{
		Name:   "list",
		Usage:  "list the files server",
		Action: ListCommand,
	}
)

var Commands = []cli.Command{
	uploadCmd,
	listCmd,
}

func UploadCommand(ctx *cli.Context) {
	filename := ctx.String(filenameFlag.Name)
	src := ctx.String(srcFlag.Name)

	var reader io.Reader
	var err error

	ll := zap.New(zap.NewJSONEncoder())

	if src == "-" {
		if filename == "" {
			ll.Fatal("You must specify \"file\" when using stdin")
		}
		reader = os.Stdin
	} else {
		if filename == "" {
			filename = path.Base(src)
		}
		reader, err = os.Open(src)
		if err != nil {
			ll.Fatal("could not open src file", zap.Error(err))
		}
	}

	c := mustSpreeClient(ctx, ll)
	cctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	srv, err := c.Create(cctx)
	if err != nil {
		ll.Fatal("could not call Create", zap.Error(err))
	}

	buf := make([]byte, 4096)
	msg := &spree.CreateRequest{
		Filename: path.Base(filename),
	}

	var offset int64
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			srv.CloseSend()

			resp, err := srv.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				ll.Fatal("error reading response from server", zap.Error(err))
			}

			if resp.Shot != nil {
				printProto(resp, ll)
			}

			srv.CloseSend()
			return
		}
		if err != nil {
			ll.Fatal("error reading file", zap.Error(err))
		}

		msg.Offset = offset
		msg.Length = int64(n)
		msg.Data = buf[:n]
		err = srv.Send(msg)
		if err != nil {
			ll.Fatal("error sending message", zap.Error(err))
		}

		resp, err := srv.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			ll.Fatal("error reading response from server", zap.Error(err))
		}

		if resp.BytesWritten != msg.Length {
			ll.Fatal("mismatch in bytes written to msg length",
				zap.Int64("written.bytes", resp.BytesWritten),
				zap.Int64("msg.bytes", msg.Length))
		}

		offset += msg.Length
	}
}

func ListCommand(ctx *cli.Context) {
	ll := zap.New(zap.NewJSONEncoder())
	c := mustSpreeClient(ctx, ll)
	req := &spree.ListRequest{}
	cctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := c.List(cctx, req)
	if err != nil {
		ll.Fatal("error in list response", zap.Error(err))
	}
	printProto(resp, ll)
}

func mustSpreeClient(ctx *cli.Context, ll zap.Logger) spree.SpreeClient {
	rpcAddr := ctx.GlobalString(rpcAddrFlag.Name)
	caCertFile := ctx.GlobalString(caCertFileFlag.Name)
	certFile := ctx.GlobalString(certFileFlag.Name)
	keyFile := ctx.GlobalString(keyFileFlag.Name)

	if caCertFile == "" || certFile == "" || keyFile == "" {
		ll.Fatal("must have CA cert, client cert, and client key")
	}

	certPool := x509.NewCertPool()
	caContents, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		ll.Fatal("unable to read CA certificate file",
			zap.String("file", caCertFile),
			zap.Error(err))
	}
	certPool.AppendCertsFromPEM(caContents)
	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}

	opts := make([]grpc.DialOption, 0)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		ll.Fatal("unable to load keypair",
			zap.String("cert.file", certFile),
			zap.String("key.file", keyFile))
	}
	tlsConfig.Certificates = []tls.Certificate{cert}
	tlsConfig.InsecureSkipVerify = true

	creds := credentials.NewTLS(tlsConfig)
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.Dial(rpcAddr, opts...)
	if err != nil {
		ll.Fatal("could not connect", zap.Error(err))
	}
	return spree.NewSpreeClient(conn)
}

func printProto(p proto.Message, ll zap.Logger) {
	jm := jsonpb.Marshaler{Indent: "    "}
	out, err := jm.MarshalToString(p)
	if err != nil {
		ll.Fatal("error marhsaling proto to json", zap.Error(err))
	}
	fmt.Println(out)
}
