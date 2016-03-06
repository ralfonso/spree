package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	// global flags
	uriFlag = cli.StringFlag{
		Name:  "uri",
		Value: "http://localhost:4285/api/image",
		Usage: "URI for upload",
	}
	authTokenFlag = cli.StringFlag{
		Name:  "auth.token",
		Value: "",
		Usage: "The token to use in the auth header",
	}

	// subcommand flags
	srcFlag = cli.StringFlag{
		Name:  "src",
		Value: "-",
		Usage: "The src file to upload. \"-\" for stdin",
	}
	filenameFlag = cli.StringFlag{
		Name:  "file.name",
		Value: "",
		Usage: "The filename to use for the upload",
	}
)

var (
	// commands
	uploadCmd = cli.Command{
		Name:   "upload",
		Usage:  "upload to the server",
		Action: UploadCommand,
		Flags: []cli.Flag{
			uriFlag,
			srcFlag,
			filenameFlag,
			authTokenFlag,
		},
	}
)

var Commands = []cli.Command{
	uploadCmd,
}

func UploadCommand(ctx *cli.Context) {
	uri := ctx.String(uriFlag.Name)
	filename := ctx.String(filenameFlag.Name)
	authToken := ctx.String(authTokenFlag.Name)
	src := ctx.String(srcFlag.Name)

	var reader io.Reader
	var err error

	if src == "-" {
		if filename == "" {
			log.Fatal("You must specify \"file.name\" when using stdin")
		}
		reader = os.Stdin
	} else {
		if filename == "" {
			filename = path.Base(src)
		}
		reader, err = os.Open(src)
		if err != nil {
			log.WithError(err).Fatal("could not open src file")
		}
	}

	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		log.WithError(err).Fatal("could not read src file")
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		log.WithError(err).Fatalf("Could not create multipart writer from file %v\n", src)
	}

	part.Write(contents)
	_ = writer.WriteField("filename", filename)
	err = writer.Close()
	if err != nil {
		log.WithError(err).Fatalf("Could not close multipart writer for file %v\n", src)
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		log.WithError(err).Fatalf("Error creating upload request for file %v\n", src)
	}

	req.Header.Set("X-Auth-Token", authToken)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Fatalf("Error making HTTP request to %v for file %v\n", uri, src)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Fatalf("Error reading response body for file %v\n", src)
	}

	fmt.Println(string(respBody))

	if resp.StatusCode != http.StatusCreated {
		os.Exit(1)
	}
}
