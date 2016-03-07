package metadata

type Store interface {
	PutFile(*File) error
	ListFiles() ([]File, error)
	GetFileById(id string) (*File, error)
}
