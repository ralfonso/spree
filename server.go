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

	var file File
	var filename string
	var ptr int64
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if file == nil && in.Filename != "" {
			filename = path.Base(in.Filename)
			ll = ll.With(zap.String("filename", filename))
			file, err = s.storage.Create(filename)
			if err != nil {
				s.ll.Error("unable to open file", zap.Error(err))
				return errInternal
			}
			defer file.Close()
			ll.Info("created new file")
		}

		if file == nil {
			ll.Info("unknown file")
			return errUnknownFile
		}

		if in.Length > 0 {
			if int64(len(in.Data)) != in.Length {
				ll.With(
					zap.Int("in.data.len", len(in.Data)),
					zap.Int64("in.length", in.Length),
				).Error("data/length mismatch")
				return errInvalidArg
			}

			if in.Offset != ptr {
				file.Seek(in.Offset, io.SeekStart)
				ptr = in.Offset
			}

			n, err := file.Write(in.Data)
			if err != nil {
				ll.Error("unable to write to file file", zap.Error(err))
				return errInternal
			}

			ptr += in.Length

			resp := &CreateResponse{
				Offset:       ptr,
				BytesWritten: int64(n),
			}
			err = stream.Send(resp)
			if err != nil {
				ll.Error("unable to send response to client", zap.Error(err))
				return errInternal
			}
		}
	}

	shot := &Shot{
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Filename:  filename,
		Backend: &BackendDetails{
			Type: "file",
		},
	}
	shot.Id = s.md.GetId(shot)

	err := s.md.PutShot(shot)
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

func (s *Server) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	ll := s.ll.With(zap.String("method", "List"))
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
