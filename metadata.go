package spree

type Metadata interface {
	GetId(*Shot) string
	PutShot(*Shot) error
	ListShots() ([]*Shot, error)
	GetShotById(id string) (*Shot, error)
	IncrementViews(id string) (*Shot, error)
	Close() error
}
