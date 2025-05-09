package storage

import "io"

type Storage interface {
    Upload(filename string, content io.Reader, contentType string) (string, error)
    Delete(filename string) error
} 