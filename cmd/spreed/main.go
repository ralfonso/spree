package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	app := cli.NewApp()
	app.Flags = GlobalFlags
	app.Name = "spree"
	app.Usage = "upload stuff"
	app.Action = func(ctx *cli.Context) {
		boltKV, err := backends.NewBoltKV(ctx.String(dbFileFlag.Name), ctx.String(dbBucketFlag.Name), ctx.String(urlPrefixFlag.Name))
		if err != nil {
			log.WithError(err).Fatalf("unable to create BoltKV")
		}

		dataDir := ctx.String(dataDirFlag.Name)
		store, err := storage.NewFileStore(dataDir)
		if err != nil {
			log.WithError(err).Fatalf("unable to create FileStore")
		}

		server := NewServer(boltKV, store)
		lis, err := net.Listen("tcp", ctx.String(addrFlag.Name))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		pb.RegisterRouteGuideServer(grpcServer, server)
		go grpcServer.Serve(lis)

		httpServer := NewHTTPServer(ctx, boltKV, store)
		go httpServer.Run()

		shutdownFn := func() {
			err := s.KV.Close()
			if err != nil {
				log.WithError(err).Error("Error closing KV DB")
			}
		}

		go handleSignals(shutdownFn)
	}

	app.Run(os.Args)
}

type HTTPServer struct {
	Addr      string
	DataDir   string
	AuthToken string
	Store     storage.Store
	KV        metadata.Store
}

func NewHTTPServer(ctx *cli.Context, md Metdata, store Store) *HTTPServer {
	return &Server{
		Addr:      ctx.String(addrFlag.Name),
		DataDir:   dataDir,
		Store:     store,
		KV:        boltKV,
		AuthToken: ctx.String(authTokenFlag.Name),
	}
}

func (s *HTTPServer) Run() {
	r := mux.NewRouter()
	r.HandleFunc("/", s.IndexHandler)
	r.HandleFunc("/p/{id}", s.DisplayPageHandler)
	r.PathPrefix("/r").Handler(http.StripPrefix("/r/", http.FileServer(http.Dir(s.DataDir))))

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/image", auth.AuthHandler(s.AuthToken, s.ImageHandler))

	log.WithFields(log.Fields{"addr": s.Addr}).Info("Starting HTTP server")

	loggingRouter := handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(s.Addr, loggingRouter))
}

func (s *HTTPServer) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<img src=\"http://thumbs1.ebaystatic.com/d/l225/m/mwtBoKyCDL2DSVbHojb7KNQ.jpg\" style=\"width:100%; height:100%\">"))
}

func (s *HTTPServer) ImageHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.ListHandler(w, r)
	case "POST":
		s.UploadHandler(w, r)
	}
}

func (s *HTTPServer) DisplayPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	file, err := s.KV.GetFileById(id)
	ll := log.WithFields(log.Fields{"id": id})
	if err != nil {
		ll.WithError(err).Error("error getting file")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if file == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// increment the views asynchronously
	file.Views++
	go s.KV.PutFile(file)

	http.Redirect(w, r, file.DirectUrl, http.StatusFound)
}

func (s *HTTPServer) ListHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *HTTPServer) UploadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("filename")
	src, _, err := r.FormFile("file")
	ll := log.WithFields(log.Fields{"filename": filename})

	if err != nil {
		ll.WithError(err).Error("error reading uploaded file")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer src.Close()

	file, err := s.Store.Save(src, filename)
	if err != nil {
		ll.WithError(err).Error("error storing file to backend")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = s.KV.PutFile(file)
	if err != nil {
		ll.WithError(err).Error("error storing file to db")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonFile, err := json.Marshal(file)
	if err != nil {
		ll.WithError(err).Error("error encoding file object to json")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, string(jsonFile))
}

func handleSignals(shutdownFn func()) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGINT)
	defer signal.Stop(sig)
	for s := range sig {
		switch s {
		case syscall.SIGUSR1:
			debug.PrintStack()
		case syscall.SIGHUP:
			log.Warn("shutting down due to signal")
			shutdownFn()
			os.Exit(1)
			return
		case syscall.SIGINT:
			log.Warn("shutting down due to signal")
			shutdownFn()
			os.Exit(0)
			return
		default:
			log.WithFields(log.Fields{"signal": s}).Warn("received signal")
			shutdownFn()
			os.Exit(2)
			return
		}
	}
}
