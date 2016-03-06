package storage

import (
	"io"

	"github.com/ralfonso/spree/internal/metadata"
)

type Store interface {
	Save(src io.Reader, filename string) (*metadata.File, error)
}
