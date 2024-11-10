package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type SharedDiskStorage struct {
	basepath string
}

func NewSharedDisk(basepath string) Storage {
	return &SharedDiskStorage{basepath: basepath}
}

func (s *SharedDiskStorage) Read(path string) (io.Reader, error) {
	file, err := os.Open(filepath.Join(s.basepath, path))
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("error reading file %v: %v", path, err)
	}

	return file, nil
}

func (s *SharedDiskStorage) Write(path string, data io.Reader) error {
	fullpath := filepath.Join(s.basepath, path)

	err := os.MkdirAll(filepath.Dir(fullpath), 0777)
	if err != nil {
		return fmt.Errorf("error creating parent directory %v: %v", path, err)
	}

	file, err := os.Create(fullpath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("error opening file %v: %v", path, err)
	}

	_, err = io.Copy(file, data)
	if err != nil {
		return fmt.Errorf("error writing to file %v: %v", path, err)
	}

	return nil
}

func (s *SharedDiskStorage) Delete(path string) error {
	err := os.RemoveAll(filepath.Join(s.basepath, path))
	if err != nil {
		return fmt.Errorf("error deleting file %v: %v", path, err)
	}
	return nil
}

func (s *SharedDiskStorage) Location() string {
	return s.basepath
}
