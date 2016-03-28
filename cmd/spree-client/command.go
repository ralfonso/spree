package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/codegangsta/cli/altsrc"
)

var (
	// global flags
	cfgFlag = cli.StringFlag{
		Name: "cfg",
	}

	uriFlag = altsrc.NewStringFlag(cli.StringFlag{
		Name:   "uri",
		Value:  "http://localhost:4285/api/image",
		Usage:  "URI for upload",
		EnvVar: "SPREE_URI",
	})

	authTokenFlag = altsrc.NewStringFlag(cli.StringFlag{
		Name:   "auth.token",
		Value:  "",
		Usage:  "The token to use in the auth header",
		EnvVar: "SPREE_AUTH_TOKEN",
	})

	filenameFlag = cli.StringFlag{
		Name:  "file.name",
		Value: "",
		Usage: "The filename to use for the upload",
	}
)

var Commands = []cli.Command{}

func init() {
	usr, _ := user.Current()
	cfgFlag.Value = filepath.Join(usr.HomeDir, ".config", "spree-client.yaml")

	// commands
	uploadCmd := cli.Command{
		Name:   "upload",
		Usage:  "upload to the server",
		Action: UploadCommand,
		Flags: []cli.Flag{
			cfgFlag,
			uriFlag,
			filenameFlag,
			authTokenFlag,
		},
	}

	uploadCmd.Before = altsrc.InitInputSourceWithContext(uploadCmd.Flags, altsrc.NewYamlSourceFromFlagFunc("cfg"))
	Commands = append(Commands, uploadCmd)
}

func UploadCommand(ctx *cli.Context) {
	uri := ctx.String(uriFlag.Name)
	filename := ctx.String(filenameFlag.Name)
	authToken := ctx.String(authTokenFlag.Name)
	src := ctx.Args().First()

	if src == "" {
		fmt.Println("You must specify a source file or \"-\" for stdin")
		os.Exit(1)
	}

	var reader io.Reader
	var err error

	if src == "-" {
		if filename == "" {
			fmt.Println("You must specify \"file.name\" when using stdin")
			os.Exit(1)
		}
		reader = os.Stdin
	} else {
		if filename == "" {
			filename = path.Base(src)
		}
		reader, err = os.Open(src)
		if err != nil {
			fmt.Println(err)
			fmt.Println("could not open src file")
			os.Exit(1)
		}
	}

	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(err)
		fmt.Println("could not read src file")
		os.Exit(1)
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Could not create multipart writer from file %v\n", src)
		os.Exit(1)
	}

	part.Write(contents)
	_ = writer.WriteField("filename", filename)
	err = writer.Close()
	if err != nil {
		fmt.Printf("Could not close multipart writer for file %v\n", src)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		fmt.Printf("Error creating upload request for file %v\n", src)
		os.Exit(1)
	}

	req.Header.Set("X-Auth-Token", authToken)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making HTTP request to %v for file %v\n", uri, src)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body for file %v\n", src)
		os.Exit(1)
	}

	fmt.Println(string(respBody))

	if resp.StatusCode != http.StatusCreated {
		os.Exit(1)
	}
}
