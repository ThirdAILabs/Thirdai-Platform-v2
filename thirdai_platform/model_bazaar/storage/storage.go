package storage

import "io"

type Storage interface {
	Read(path string) (io.ReadCloser, error)

	Write(path string, data io.Reader) error

	Append(path string, data io.Reader) error

	Delete(path string) error

	List(path string) ([]string, error)

	Exists(path string) (bool, error)

	Unzip(path string) error

	Zip(path string) error

	Size(path string) (int64, error)

	Usage() (UsageStats, error)

	Location() string
}

type UsageStats struct {
	TotalBytes uint64
	FreeBytes  uint64
}
