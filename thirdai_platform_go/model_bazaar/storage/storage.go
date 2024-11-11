package storage

import "io"

type Storage interface {
	Read(path string) (io.ReadCloser, error)

	Write(path string, data io.Reader) error

	Append(path string, data io.Reader) error

	Delete(path string) error

	List(path string) ([]string, error)

	Unzip(path string) error

	Zip(path string) error

	Location() string
}
