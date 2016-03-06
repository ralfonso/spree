package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ralfonso/spree/internal/auth"
	"github.com/ralfonso/spree/internal/metadata"
	"github.com/ralfonso/spree/internal/metadata/backends"
	"github.com/ralfonso/spree/internal/storage"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	app := cli.NewApp()
	app.Flags = GlobalFlags
	app.Name = "spree"
	app.Usage = "upload stuff"
	app.Action = func(c *cli.Context) {
		server := NewServer(c)
		server.Run()
	}

	app.Run(os.Args)
}

type Server struct {
	Addr      string
	DataDir   string
	AuthToken string
	Store     storage.Store
	KV        metadata.Store
}

func NewServer(ctx *cli.Context) *Server {
	dataDir := ctx.String(dataDirFlag.Name)
	store := storage.NewFileStore(dataDir, ctx.String(urlPrefixFlag.Name))
	boltKV, err := backends.NewBoltKV(ctx.String(dbFileFlag.Name), ctx.String(dbBucketFlag.Name))
	if err != nil {
		log.WithError(err).Fatal("unable to connect to KV")
	}

	return &Server{
		Addr:      ctx.String(addrFlag.Name),
		DataDir:   dataDir,
		Store:     store,
		KV:        boltKV,
		AuthToken: ctx.String(authTokenFlag.Name),
	}
}

func (s *Server) Run() {
	r := mux.NewRouter()
	r.HandleFunc("/", s.IndexHandler)
	r.PathPrefix("/r").Handler(http.StripPrefix("/r/", http.FileServer(http.Dir(s.DataDir))))

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/image", auth.AuthHandler(s.AuthToken, s.ImageHandler))

	log.WithFields(log.Fields{"addr": s.Addr}).Info("Starting HTTP server")

	loggingRouter := handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(s.Addr, loggingRouter))
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<img src=\"http://thumbs1.ebaystatic.com/d/l225/m/mwtBoKyCDL2DSVbHojb7KNQ.jpg\" style=\"width:100%; height:100%\">"))
}

func (s *Server) ImageHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.ListHandler(w, r)
	case "POST":
		s.UploadHandler(w, r)
	}
}

func (s *Server) ListHandler(w http.ResponseWriter, r *http.Request) {
	files, err := s.KV.ListFiles()
	if err != nil {
		log.WithError(err).Error("error listing files")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonBody, err := json.Marshal(files)
	if err != nil {
		log.WithError(err).Error("error encoding file list to json")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(jsonBody))
}

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("filename")
	src, _, err := r.FormFile("file")
	if err != nil {
		log.WithError(err).Error("error reading uploaded file")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer src.Close()

	file, err := s.Store.Save(src, filename)
	if err != nil {
		log.WithError(err).Error("error storing file to backend")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = s.KV.PutFile(file)
	if err != nil {
		log.WithError(err).Error("error storing file to backend")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonFile, err := json.Marshal(file)
	if err != nil {
		log.WithError(err).Error("error encoding file object to json")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, string(jsonFile))
}
