package storage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type SharedDiskStorage struct {
	basepath string
}

func NewSharedDisk(basepath string) Storage {
	return &SharedDiskStorage{basepath: basepath}
}

func (s *SharedDiskStorage) Read(path string) (io.ReadCloser, error) {
	file, err := os.Open(filepath.Join(s.basepath, path))
	if err != nil {
		return nil, fmt.Errorf("error reading file %v: %v", path, err)
	}

	return file, nil
}

func (s *SharedDiskStorage) Write(path string, data io.Reader) error {
	return s.writeData(path, data, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
}

func (s *SharedDiskStorage) Append(path string, data io.Reader) error {
	return s.writeData(path, data, os.O_RDWR|os.O_CREATE|os.O_APPEND)
}

func (s *SharedDiskStorage) writeData(path string, data io.Reader, flags int) error {
	fullpath := filepath.Join(s.basepath, path)

	err := os.MkdirAll(filepath.Dir(fullpath), 0777)
	if err != nil {
		return fmt.Errorf("error creating parent directory %v: %v", path, err)
	}

	file, err := os.OpenFile(fullpath, flags, 0666)
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

func (s *SharedDiskStorage) List(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error listing entries at %v: %w", path, err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, entry.Name())
	}

	return paths, nil
}

func (s *SharedDiskStorage) Unzip(path string) error {
	zip, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zip.Close()

	newPath := strings.TrimSuffix(path, ".zip")

	for _, file := range zip.File {
		if strings.HasSuffix(file.Name, "/") {
			continue // directory
		}

		fileData, err := file.Open()
		if err != nil {
			return fmt.Errorf("error opening file in zipfile %v: %w", file.Name, err)
		}
		defer fileData.Close()

		err = s.Write(filepath.Join(newPath, file.Name), fileData)
		if err != nil {
			return fmt.Errorf("error writing contents from zipfile %v: %w", file.Name, err)
		}
	}

	return nil
}

func (s *SharedDiskStorage) Zip(path string) error {
	zipfile, err := os.Create(path + ".zip")
	if err != nil {
		return fmt.Errorf("error creading file to store zip archive: %w", err)
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = archive.AddFS(os.DirFS(path))
	if err != nil {
		return fmt.Errorf("error writing directory to zipfile: %w", err)
	}

	return nil
}

func (s *SharedDiskStorage) Location() string {
	return s.basepath
}
