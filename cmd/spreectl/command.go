package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/codegangsta/cli"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/ralfonso/spree"
	"github.com/ralfonso/spree/auth"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultAccessTokenFileName = "config.json"
	chunkSizeBytes             = 1.049e+6
)

var (
	// global flags
	rpcAddrFlag = cli.StringFlag{
		Name:  "rpc.addr",
		Value: "",
		Usage: "Addr for the remote RPC server",
	}
	oauthConfigFileFlag = cli.StringFlag{
		Name:  "oauth.config.file",
		Value: "",
		Usage: "Addr for the remote RPC server",
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
	caCertFileFlag = cli.StringFlag{
		Name:  "ca.cert.file",
		Value: "",
		Usage: "CA cert file",
	}

	oauthScopes = []string{
		"https://www.googleapis.com/auth/userinfo.email",
	}
)

var GlobalFlags = []cli.Flag{
	rpcAddrFlag,
}

var (
	// commands
	authCmd = cli.Command{
		Name:   "auth",
		Usage:  "authenticates with Google and stores the token",
		Action: AuthCommand,
	}
	uploadCmd = cli.Command{
		Name:   "upload",
		Usage:  "upload to the server",
		Action: UploadCommand,
		Flags: []cli.Flag{
			srcFlag,
			filenameFlag,
			caCertFileFlag,
		},
	}
	listCmd = cli.Command{
		Name:   "list",
		Usage:  "list the files server",
		Action: ListCommand,
		Flags: []cli.Flag{
			caCertFileFlag,
		},
	}
)

var Commands = []cli.Command{
	authCmd,
	uploadCmd,
	listCmd,
}

func AuthCommand(ctx *cli.Context) {
	ll, _ := zap.NewDevelopment()
	oauthConf := mustOauthConfFromFile(ctx, ll)
	fmt.Println("Opening web browser to log in with Google.")

	authURL := oauthConf.AuthCodeURL("test")
	err := open.Run(authURL)
	if err != nil {
		ll.Fatal("could not open url in browser", zap.Error(err))
	}

	fmt.Print("Paste Google Developer OAuth Code: ")
	reader := bufio.NewReader(os.Stdin)
	code, _ := reader.ReadString('\n')
	code = strings.TrimRight(code, "\n")
	token, err := oauthConf.Exchange(context.Background(), code)
	if err != nil {
		ll.Fatal("could not exchange code for token", zap.Error(err))
	}

	jwt, err := auth.NewClientJWTFromOauth2(token, ll)
	if err != nil {
		ll.Fatal("could not convert oauth2 token to JWT", zap.Error(err))
	}

	clientConf := &auth.ClientConfig{
		OauthToken: token,
		JWT:        jwt,
	}

	err = storeConfig(clientConf, ll)
	if err != nil {
		ll.Fatal("could not store token in config")
	}
}

func UploadCommand(ctx *cli.Context) {
	filename := ctx.String(filenameFlag.Name)
	src := ctx.String(srcFlag.Name)

	var rdr io.Reader
	var err error

	ll, _ := zap.NewDevelopment()

	if src == "-" {
		if filename == "" {
			ll.Fatal("You must specify \"file\" when using stdin")
		}
		rdr = os.Stdin
	} else {
		if filename == "" {
			filename = path.Base(src)
		}
		rdr, err = os.Open(src)
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

	start := time.Now()
	uploadedBytes, err := doUpload(srv, filename, rdr, ll)

	finish := time.Since(start)
	ll.Info("completed upload",
		zap.Float64("MiB/s", float64(uploadedBytes/1024)/finish.Seconds()))
}

func doUpload(srv spree.Spree_CreateClient, filename string, rdr io.Reader, ll *zap.Logger) (int64, error) {
	buf := make([]byte, chunkSizeBytes)
	msg := &spree.CreateRequest{
		Filename: path.Base(filename),
	}

	var offset int64
	for {
		n, err := rdr.Read(buf)
		if err == io.EOF {
			srv.CloseSend()

			resp, err := srv.Recv()
			if err == io.EOF {
				return offset, err
			}
			if err != nil {
				ll.Fatal("error reading response from server", zap.Error(err))
			}

			if resp.Shot != nil {
				printProto(resp, ll)
			}

			return offset, nil
		}
		if err != nil {
			ll.Fatal("error reading file", zap.Error(err))
		}

		msg.Offset = offset
		msg.Length = int64(n)
		msg.Data = buf[:n]
		ll.With(
			zap.Int64("msg.offset", msg.Offset),
			zap.Int64("msg.length", msg.Length),
		).Info("sending message part")

		err = srv.Send(msg)
		if err != nil {
			ll.Fatal("error sending message", zap.Error(err))
		}

		resp, err := srv.Recv()
		if err == io.EOF {
			return offset, err
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
	ll, _ := zap.NewDevelopment()
	c := mustSpreeClient(ctx, ll)
	req := &spree.ListRequest{}
	cctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	ll.Info("making list request")
	resp, err := c.List(cctx, req)
	if err != nil {
		ll.Fatal("error in list response", zap.Error(err))
	}
	printProto(resp, ll)
}

func mustSpreeClient(ctx *cli.Context, ll *zap.Logger) spree.SpreeClient {
	rpcAddr := ctx.GlobalString(rpcAddrFlag.Name)
	caCertFile := ctx.String(caCertFileFlag.Name)
	oauthConfig := mustOauthConfFromFile(ctx, ll)
	tlsConfig := &tls.Config{}

	if caCertFile != "" {
		caCert, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			ll.Fatal("could not read ca cert file", zap.Error(err),
				zap.String("ca.cert.file", caCertFile))
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = certPool
	}

	clientConf, err := getConfig(ll)
	if err != nil {
		ll.Fatal("unable to retrieve client configuration")
	}
	if clientConf == nil || clientConf.JWT == nil {
		ll.Fatal("missing token. please use the auth command first")
	}

	// allow the config file to override only if the flag was not set
	if rpcAddr == "" {
		rpcAddr = clientConf.RPCAddr
	}

	// local dev hax
	rpcParts := strings.Split(rpcAddr, ":")
	if rpcParts[0] == "localhost" {
		tlsConfig.InsecureSkipVerify = true
	}
	ll = ll.With(zap.String("rpc.addr", rpcAddr))

	creds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	a := auth.NewAuthenticator(ll)
	oauthToken, jwt, err := a.RefreshJWT(cctx, clientConf, oauthConfig)
	if err != nil {
		ll.Fatal("could not refresh jwt", zap.Error(err))
	}
	cancel()
	if oauthToken.AccessToken != clientConf.OauthToken.AccessToken || jwt.Token != clientConf.JWT.Token {
		clientConf.OauthToken = oauthToken
		clientConf.JWT = jwt
		err = storeConfig(clientConf, ll)
		if err != nil {
			ll.Fatal("unable to write new access token to config", zap.Error(err))
		}
	}

	opts = append(opts, grpc.WithPerRPCCredentials(jwt))

	ll.Debug("dialing rpc endpoint")
	conn, err := grpc.Dial(rpcAddr, opts...)
	if err != nil {
		ll.Fatal("could not connect", zap.Error(err))
	}
	return spree.NewSpreeClient(conn)
}

func mustOauthConfFromFile(ctx *cli.Context, ll *zap.Logger) *oauth2.Config {
	oauthConfFilename := ctx.GlobalString(oauthConfigFileFlag.Name)
	if oauthConfFilename == "" {
		// try the default location
		oauthConfFilename = filepath.Join(configHome(), "oauth.json")
		if _, err := os.Stat(oauthConfFilename); os.IsNotExist(err) {
			ll.Fatal("oauth config file not specified or found at default location",
				zap.String("default", oauthConfFilename))
		}
	}
	jsonConf, err := ioutil.ReadFile(oauthConfFilename)
	if err != nil {
		ll.Fatal("could not read oauth config file",
			zap.Error(err), zap.String("oauth.config.file", oauthConfFilename))
	}
	oauthConf, err := google.ConfigFromJSON(jsonConf, oauthScopes...)
	if err != nil {
		ll.Fatal("could not parse JSON oauth config",
			zap.Error(err))
	}

	return oauthConf
}

func printProto(p proto.Message, ll *zap.Logger) {
	jm := jsonpb.Marshaler{Indent: "    "}
	out, err := jm.MarshalToString(p)
	if err != nil {
		ll.Fatal("error marhsaling proto to json", zap.Error(err))
	}
	fmt.Println(out)
}
