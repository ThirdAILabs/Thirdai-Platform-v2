package storage

import "io"

type Storage interface {
	Read(path string) (io.Reader, error)

	Write(path string, data io.Reader) error

	Delete(path string) error
}
