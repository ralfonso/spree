package spree

type Server struct {
	md Metadata
	backend Store
}

var _ &Server{} Spree

func NewServer(md Metadata, backend Store) *Server {
	return &Server{
		md: md,
		backend: backend,
	}
}

func (s *Server) Create(stream pb.Spree_CreateServer) error {

	ll := log.WithFields(log.Fields{
		"req": req,
	})

	var shot *Shot
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if create.filename != "" {
			shot = s.backend.Open(create.filename)
		}
		for _, note := range s.routeNotes[key] {
			if err := stream.Send(note); err != nil {
				return err
			}
		}
	}
}
