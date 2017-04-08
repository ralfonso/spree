package spree

import (
	"io"
	"path"
	"time"

	"github.com/uber-go/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	errInternal    = grpc.Errorf(codes.Internal, "operation failed")
	errUnknownFile = grpc.Errorf(codes.FailedPrecondition, "no file specified")
	errInvalidArg  = grpc.Errorf(codes.InvalidArgument, "invalid argument")
)

type Server struct {
	ll      zap.Logger
	md      Metadata
	storage Storage
}

var _ SpreeServer = &Server{}

func NewServer(md Metadata, storage Storage, ll zap.Logger) *Server {
	return &Server{
		ll:      ll,
		md:      md,
		storage: storage,
	}
}

func (s *Server) Create(stream Spree_CreateServer) error {
	ll := s.ll.With(
		zap.String("method", "Create"),
	)
	ll.Info("starting rpc")

	shot, err := s.handleFileUpload(stream, ll)

	if shot == nil {
		return errInternal
	}

	err = s.md.PutShot(shot)
	if err != nil {
		ll.With(zap.Object("shot", shot)).Error("unable to put shot", zap.Error(err))
		return err
	}
	resp := &CreateResponse{
		Shot: shot,
	}
	err = stream.Send(resp)
	if err != nil {
		ll.With(zap.Object("shot", shot)).Error("unable to send shot response", zap.Error(err))
		return err
	}
	return nil
}

func (s *Server) cleanupFile(filename string, success bool, ll zap.Logger) {
	if !success {
		ll.Info("cleaning up unsuccessful upload")
		err := s.storage.Remove(filename)
		if err != nil {
			ll.Error("unable to remove file", zap.Error(err))
			return
		}
	}
}

func (s *Server) newFile(filename string, ll zap.Logger) (File, error) {
	file, err := s.storage.Create(filename)
	if err != nil {
		s.ll.Error("unable to open file", zap.Error(err))
		return nil, err
	}
	ll.Info("created new file")
	return file, nil
}

func (s *Server) handleFileUpload(stream Spree_CreateServer, ll zap.Logger) (*Shot, error) {
	var file File
	var filename string

	in, err := stream.Recv()
	if err == io.EOF {
		return nil, errUnknownFile
	}
	if err != nil {
		return nil, err
	}

	var success bool

	if in.Filename != "" {
		filename = path.Base(in.Filename)
		ll = ll.With(zap.String("filename", filename))
		var err error
		file, err = s.newFile(filename, ll)
		if err != nil {
			ll.Error("could not create new file", zap.Error(err))
			return nil, errInternal
		}
		defer func() {
			file.Close()
			s.cleanupFile(filename, success, ll)
		}()

		if err != nil {
			return nil, errUnknownFile
		}
	}

	ll.Info("handling file content", zap.String("filename", filename))

	var ptr int64
	for {
		if in.Length > 0 {
			ll.With(
				zap.Int("in.data.len", len(in.Data)),
				zap.Int64("in.length", in.Length),
				zap.Int64("in.offset", in.Offset),
			).Info("reading file")
			if int64(len(in.Data)) != in.Length {
				ll.With(
					zap.Int("in.data.len", len(in.Data)),
					zap.Int64("in.length", in.Length),
				).Error("data/length mismatch")
				success = false
				return nil, errInvalidArg
			}

			if in.Offset != ptr {
				file.Seek(in.Offset, io.SeekStart)
				ptr = in.Offset
			}

			n, err := file.Write(in.Data)
			if err != nil {
				ll.Error("unable to write to file file", zap.Error(err))
				success = false
				return nil, errInternal
			}

			ptr += in.Length

			resp := &CreateResponse{
				Offset:       ptr,
				BytesWritten: int64(n),
			}
			err = stream.Send(resp)
			if err != nil {
				ll.Error("unable to send response to client", zap.Error(err))
				success = false
				return nil, errInternal
			}
		}

		in, err = stream.Recv()
		if err == io.EOF {
			success = ptr > 0
			break
		}
		if err != nil {
			success = false
			return nil, err
		}
	}

	shot := &Shot{
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Filename:  filename,
		SizeBytes: uint64(ptr),
		Backend: &BackendDetails{
			Type: "file",
		},
	}
	shot.Id = s.md.GetId(shot)
	return shot, nil
}

func (s *Server) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	ll := s.ll.With(zap.String("method", "List"))
	ll.Info("starting rpc")
	shots, err := s.md.ListShots()
	if err != nil {
		ll.Error("error listing shots", zap.Error(err))
		return nil, errInternal
	}

	resp := &ListResponse{
		Shots: shots,
	}

	return resp, nil
}
