package metadata

type Store interface {
	PutFile(*File) error
	ListFiles() ([]File, error)
}
