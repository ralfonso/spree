package spree

import "io"

type File interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

type Storage interface {
	Open(filename string) (File, error)
	Create(filename string) (File, error)
}
