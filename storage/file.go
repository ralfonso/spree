package storage

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	prefixChar = "0123456789abcdef"
)

type FileStorage struct {
	path string
}

func NewFileStorage(path string) (*FileStorage, error) {
	err := createPrefixPaths(path)
	if err != nil {
		return nil, err
	}
	return &FileStorage{
		path: path,
	}, nil
}

func createPrefixPaths(path string) error {
	if !os.IsExist(path) || !os.IsDir(path) {
		return errors.New("path does not exist for FileStorage:", path)
	}

	return createDirs(path, 1)
}

func createDirs(path string, depth int) error {
	for _, c := range prefixChar {
		cpath := filepath.Join(path, c)
		err := os.Mkdir(cpath)
		if err != nil {
			return err
		}
		if depth > 0 {
			err = createDirs(cpath, depth-1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (fs *FileStorage) Open(filename string) (*File, error) {
	filePath := filepath.Join(fs.path, filename)
	return os.Open(filePath)
}
