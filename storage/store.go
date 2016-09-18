package storage

import "io"

type File interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

type Store interface {
	Open(filename string) (*File, error)
}
