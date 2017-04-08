package spree

import (
	"fmt"
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
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist for FileStorage:", path)
	}

	return createDirs(path, 1)
}

func createDirs(path string, depth int) error {
	for _, c := range prefixChar {
		cpath := filepath.Join(path, string(c))
		err := os.Mkdir(cpath, 0755)
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

func (fs *FileStorage) Open(filename string) (File, error) {
	filePath := filepath.Join(fs.path, filename)
	return os.Open(filePath)
}

func (fs *FileStorage) Create(filename string) (File, error) {
	filePath := filepath.Join(fs.path, filename)
	return os.Create(filePath)
}

func (fs *FileStorage) Remove(filename string) error {
	filePath := filepath.Join(fs.path, filename)
	return os.Remove(filePath)
}
