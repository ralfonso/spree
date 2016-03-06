package storage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ralfonso/spree/internal/metadata"
)

type FileStore struct {
	DataDir   string
	UrlPrefix string
}

func NewFileStore(dataDir, urlPrefix string) *FileStore {
	return &FileStore{
		DataDir:   dataDir,
		UrlPrefix: urlPrefix,
	}
}

func (s *FileStore) Save(src io.Reader, filename string) (*metadata.File, error) {
	nano := time.Now().UnixNano()
	filename = fmt.Sprintf("%d_%s", nano, filename)
	outputFilename := filepath.Join(s.DataDir, filename)
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return nil, err
	}

	bufWrite := bufio.NewWriter(outputFile)
	bufRead := bufio.NewReader(src)

	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := bufRead.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := bufWrite.Write(buf[:n]); err != nil {
			return nil, err
		}
	}

	file := &metadata.File{
		FullPath: outputFilename,
		Url:      s.UrlPrefix + "/r/" + filename,
	}

	return file, nil
}
