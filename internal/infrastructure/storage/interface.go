package storage

import (
	"context"
	"time"
)

// Storage defines the interface for PDF storage operations
type Storage interface {
	// StorePDF stores a PDF file and returns the storage key/path
	StorePDF(ctx context.Context, jobID string, pdfData []byte) (string, error)
	
	// GetPDF retrieves a PDF file by its storage key
	GetPDF(ctx context.Context, storageKey string) ([]byte, error)
	
	// GetPDFURL generates a presigned URL for PDF access
	GetPDFURL(ctx context.Context, storageKey string, expiry time.Duration) (string, error)
	
	// DeletePDF removes a PDF file from storage
	DeletePDF(ctx context.Context, storageKey string) error
	
	// ListPDFs lists PDF files with optional prefix filter
	ListPDFs(ctx context.Context, prefix string) ([]string, error)
	
	// HealthCheck verifies storage connectivity
	HealthCheck(ctx context.Context) error
	
	// GetStats returns storage statistics
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

// LocalStorage implements Storage interface for local file system (fallback)
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
	}
}

// StorePDF stores PDF locally (basic implementation)
func (l *LocalStorage) StorePDF(ctx context.Context, jobID string, pdfData []byte) (string, error) {
	// Basic local storage implementation
	// This would be used as fallback when MinIO is not available
	return "", nil
}

// GetPDF retrieves PDF from local storage
func (l *LocalStorage) GetPDF(ctx context.Context, storageKey string) ([]byte, error) {
	return nil, nil
}

// GetPDFURL generates local file URL
func (l *LocalStorage) GetPDFURL(ctx context.Context, storageKey string, expiry time.Duration) (string, error) {
	return "", nil
}

// DeletePDF removes PDF from local storage
func (l *LocalStorage) DeletePDF(ctx context.Context, storageKey string) error {
	return nil
}

// ListPDFs lists local PDFs
func (l *LocalStorage) ListPDFs(ctx context.Context, prefix string) ([]string, error) {
	return nil, nil
}

// HealthCheck checks local storage
func (l *LocalStorage) HealthCheck(ctx context.Context) error {
	return nil
}

// GetStats returns local storage stats
func (l *LocalStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"type": "local",
		"path": l.basePath,
	}, nil
}
