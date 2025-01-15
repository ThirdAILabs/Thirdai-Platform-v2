package storage

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

type SharedDiskStorage struct {
	basepath string
}

func NewSharedDisk(basepath string) Storage {
	slog.Info("creating new shared disk storage", "basepath", basepath)
	return &SharedDiskStorage{basepath: basepath}
}

func (s *SharedDiskStorage) fullpath(path string) string {
	return filepath.Join(s.basepath, path)
}

func (s *SharedDiskStorage) Read(path string) (io.ReadCloser, error) {
	fullpath := s.fullpath(path)
	file, err := os.Open(fullpath)
	if err != nil {
		slog.Error("error opening file for read", "path", fullpath, "error", err)
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
	fullpath := s.fullpath(path)

	err := os.MkdirAll(filepath.Dir(fullpath), 0777)
	if err != nil {
		slog.Error("error creating parent directory", "path", fullpath, "error", err)
		return fmt.Errorf("error creating parent directory %v: %v", path, err)
	}

	file, err := os.OpenFile(fullpath, flags, 0666)
	if err != nil {
		slog.Error("error opening file for writing", "path", fullpath, "error", err)
		return fmt.Errorf("error opening file %v: %v", path, err)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		slog.Error("error writing to file", "path", fullpath, "error", err)
		return fmt.Errorf("error writing to file %v: %v", path, err)
	}

	return nil
}

func (s *SharedDiskStorage) Delete(path string) error {
	fullpath := s.fullpath(path)
	err := os.RemoveAll(fullpath)
	if err != nil {
		slog.Error("error deleting file", "path", fullpath, "error", err)
		return fmt.Errorf("error deleting file %v: %v", path, err)
	}
	return nil
}

func (s *SharedDiskStorage) List(path string) ([]string, error) {
	fullpath := s.fullpath(path)
	entries, err := os.ReadDir(fullpath)
	if err != nil {
		slog.Error("error listing entries", "path", fullpath, "error", err)
		return nil, fmt.Errorf("error listing entries at %v: %w", path, err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, entry.Name())
	}

	return paths, nil
}

func (s *SharedDiskStorage) Exists(path string) (bool, error) {
	fullpath := s.fullpath(path)
	_, err := os.Stat(fullpath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	slog.Error("error checking if file exists", "path", fullpath, "error", err)
	return false, fmt.Errorf("error checking if file %v exists: %w", fullpath, err)
}

func (s *SharedDiskStorage) Unzip(path string) error {
	fullpath := s.fullpath(path)
	zip, err := zip.OpenReader(fullpath)
	if err != nil {
		slog.Error("error opening zip reader", "path", fullpath, "error", err)
		return fmt.Errorf("error opening zip reader: %w", err)
	}
	defer zip.Close()

	newPath := strings.TrimSuffix(path, ".zip")

	for _, file := range zip.File {
		if strings.HasSuffix(file.Name, "/") {
			continue // directory
		}

		fileData, err := file.Open()
		if err != nil {
			slog.Error("error opening file in zipfile", "path", fullpath, "name", file.Name, "error", err)
			return fmt.Errorf("error opening file in zipfile %v: %w", file.Name, err)
		}
		defer fileData.Close()

		err = s.Write(filepath.Join(newPath, file.Name), fileData)
		if err != nil {
			slog.Error("error writing contents of file in zipfile", "path", fullpath, "name", file.Name, "error", err)
			return fmt.Errorf("error writing contents from zipfile %v: %w", file.Name, err)
		}
	}

	return nil
}

func (s *SharedDiskStorage) Zip(path string) error {
	fullpath := s.fullpath(path)
	zipfile, err := os.Create(fullpath + ".zip")
	if err != nil {
		slog.Error("error creating file to store zip archive", "path", fullpath, "error", err)
		return fmt.Errorf("error creating file to store zip archive: %w", err)
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = archive.AddFS(os.DirFS(fullpath))
	if err != nil {
		slog.Error("error writing directory to zip archive", "path", fullpath, "error", err)
		return fmt.Errorf("error writing directory '%v' to zipfile: %w", fullpath, err)
	}

	return nil
}

func (s *SharedDiskStorage) Size(path string) (int64, error) {
	fullpath := s.fullpath(path)

	info, err := os.Stat(fullpath)
	if err != nil {
		slog.Error("error getting stats for file", "path", fullpath, "error", err)
		return 0, fmt.Errorf("error gettings stats for file %v: %w", fullpath, err)
	}

	return info.Size(), nil
}

func (s *SharedDiskStorage) Usage() (UsageStats, error) {
	var stat unix.Statfs_t

	err := unix.Statfs(s.basepath, &stat)
	if err != nil {
		slog.Error("error getting disk usage for shared storage", "path", s.basepath, "error", err)
		return UsageStats{}, fmt.Errorf("error getting disk usage stats: %w", err)
	}

	return UsageStats{
		TotalBytes: stat.Blocks * uint64(stat.Bsize),
		FreeBytes:  stat.Bfree * uint64(stat.Bsize),
	}, nil
}

func (s *SharedDiskStorage) Location() string {
	return s.basepath
}
