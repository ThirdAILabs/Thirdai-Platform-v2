package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type SharedDisk struct {
	basepath string
}

func NewSharedDisk(basepath string) Storage {
	return &SharedDisk{basepath: basepath}
}

func (s *SharedDisk) Read(path string) (io.Reader, error) {
	file, err := os.Open(filepath.Join(s.basepath, path))
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("error reading file %v: %v", path, err)
	}

	return file, nil
}

func (s *SharedDisk) Write(path string, data io.Reader) error {
	file, err := os.Create(filepath.Join(s.basepath, path))
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

func (s *SharedDisk) Delete(path string) error {
	err := os.RemoveAll(filepath.Join(s.basepath, path))
	if err != nil {
		return fmt.Errorf("error deleting file %v: %v", path, err)
	}
	return nil
}
