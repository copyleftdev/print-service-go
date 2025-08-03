package services

import (
	"fmt"
	"os"
	"path/filepath"
)

// StorageService provides file storage functionality
type StorageService struct {
	basePath string
}

// NewStorageService creates a new storage service
func NewStorageService(basePath string) *StorageService {
	return &StorageService{
		basePath: basePath,
	}
}

// GetPath returns the full path for a filename
func (ss *StorageService) GetPath(filename string) string {
	return filepath.Join(ss.basePath, filename)
}

// WriteFile writes data to a file
func (ss *StorageService) WriteFile(path string, data []byte) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadFile reads data from a file
func (ss *StorageService) ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// DeleteFile deletes a file
func (ss *StorageService) DeleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// FileExists checks if a file exists
func (ss *StorageService) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
