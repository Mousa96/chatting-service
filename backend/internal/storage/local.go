package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	uploadDir string
	baseURL   string
}

func NewLocalStorage(uploadDir, baseURL string) *LocalStorage {
	return &LocalStorage{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

func (s *LocalStorage) Upload(filename string, content io.Reader, contentType string) (string, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Create file with safe path
	path := filepath.Join(s.uploadDir, filename)
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content to file
	if _, err := io.Copy(file, content); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return public URL
	return fmt.Sprintf("%s/%s", s.baseURL, filename), nil
}

func (s *LocalStorage) Delete(filename string) error {
	path := filepath.Join(s.uploadDir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
} 