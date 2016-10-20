package spree

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/uber-go/zap"

	_ "github.com/ralfonso/spree/statik"
)

type HTTPServer struct {
	ll      zap.Logger
	addr    string
	storage Storage
	md      Metadata
	jm      jsonpb.Marshaler
}

const (
	displayPath = "/p"
	directPath  = "/r"
)

func NewHTTPServer(addr string, md Metadata, storage Storage, ll zap.Logger) *HTTPServer {
	return &HTTPServer{
		ll:      ll,
		addr:    addr,
		storage: storage,
		md:      md,
		jm:      jsonpb.Marshaler{Indent: "  "},
	}
}

func (s *HTTPServer) Run() {
	r := mux.NewRouter()

	statikFS, err := fs.New()
	if err != nil {
		s.ll.Fatal("error opening static filesystem", zap.Error(err))
	}
	r.HandleFunc("/", s.IndexHandler)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(statikFS)))
	r.HandleFunc("/p/{id}", s.DisplayPageHandler)
	r.HandleFunc("/r/{filename}", s.DirectHandler)

	s.ll.Info("Starting HTTP server",
		zap.String("addr", s.addr))

	loggingRouter := handlers.LoggingHandler(os.Stdout, r)
	s.ll.Fatal("http server stopped unexpectedly", zap.Error(http.ListenAndServe(s.addr, loggingRouter)))
}

func (s *HTTPServer) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<img src=\"/static/spree.jpg\" style=\"width:100%; height:100%\">"))
}

func (s *HTTPServer) DisplayPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	shot, err := s.md.GetShotById(id)
	ll := s.ll.With(zap.String("id", id))
	if err != nil {
		ll.Error("error getting shot", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if shot == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// increment the views asynchronously
	go s.md.IncrementViews(id)

	http.Redirect(w, r, directUrl(shot), http.StatusFound)
}

func (s *HTTPServer) DirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if filename == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	ll := s.ll.With(zap.String("filename", filename))
	ll.Info("fetching file")

	file, err := s.storage.Open(filename)
	if err != nil {
		ll.Error("error reading file from storage", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ll.Info("sending file")
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	w.Header().Set("Content-Type", mimeType)
	io.Copy(w, file)
}

func directUrl(shot *Shot) string {
	return fmt.Sprintf("%s/%s", directPath, shot.Filename)
}
