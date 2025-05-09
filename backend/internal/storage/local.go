package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	uploadDir    string
	publicPath   string
}

func NewLocalStorage(uploadDir, publicPath string) Storage {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create upload directory: %v", err))
	}
	return &LocalStorage{
		uploadDir:    uploadDir,
		publicPath:   publicPath,
	}
}

func (s *LocalStorage) Upload(filename string, src io.Reader, contentType string) (string, error) {
	// Remove leading slash from filename if present
	filename = filepath.Clean(filename)
	if filename[0] == '/' {
		filename = filename[1:]
	}

	// Create the full path
	fullPath := filepath.Join(s.uploadDir, filename)

	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy the file content
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return the public URL path
	return filepath.Join(s.publicPath, filename), nil
}

func (s *LocalStorage) Delete(filename string) error {
	path := filepath.Join(s.uploadDir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
} 