package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage handles PDF storage operations using MinIO
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
}

// StorageConfig holds MinIO configuration
type StorageConfig struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	BucketName      string
	UseSSL          bool
	CreateBucket    bool
}

// NewMinIOStorage creates a new MinIO storage client
func NewMinIOStorage(config StorageConfig) (*MinIOStorage, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	storage := &MinIOStorage{
		client:     minioClient,
		bucketName: config.BucketName,
	}

	// Create bucket if it doesn't exist and CreateBucket is true
	if config.CreateBucket {
		if err := storage.ensureBucket(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
		}
	}

	return storage, nil
}

// NewMinIOStorageFromEnv creates MinIO storage from environment variables
func NewMinIOStorageFromEnv() (*MinIOStorage, error) {
	config := StorageConfig{
		Endpoint:     getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:    getEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:    getEnvOrDefault("MINIO_SECRET_KEY", "minioadmin123"),
		BucketName:   getEnvOrDefault("MINIO_BUCKET", "print-service-pdfs"),
		UseSSL:       getEnvOrDefault("MINIO_USE_SSL", "false") == "true",
		CreateBucket: true,
	}

	return NewMinIOStorage(config)
}

// ensureBucket creates the bucket if it doesn't exist
func (s *MinIOStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created MinIO bucket: %s", s.bucketName)
	}

	return nil
}

// StorePDF stores a PDF file in MinIO and returns the object key
func (s *MinIOStorage) StorePDF(ctx context.Context, jobID string, pdfData []byte) (string, error) {
	objectKey := fmt.Sprintf("pdfs/%s/%s.pdf", time.Now().Format("2006/01/02"), jobID)
	
	reader := bytes.NewReader(pdfData)
	
	_, err := s.client.PutObject(ctx, s.bucketName, objectKey, reader, int64(len(pdfData)), minio.PutObjectOptions{
		ContentType: "application/pdf",
		UserMetadata: map[string]string{
			"job-id":     jobID,
			"created-at": time.Now().UTC().Format(time.RFC3339),
		},
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to store PDF in MinIO: %w", err)
	}

	log.Printf("Stored PDF for job %s at %s", jobID, objectKey)
	return objectKey, nil
}

// GetPDF retrieves a PDF file from MinIO
func (s *MinIOStorage) GetPDF(ctx context.Context, objectKey string) ([]byte, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get PDF from MinIO: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF data: %w", err)
	}

	return data, nil
}

// GetPDFURL generates a presigned URL for PDF access
func (s *MinIOStorage) GetPDFURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = 24 * time.Hour // Default 24 hours
	}

	url, err := s.client.PresignedGetObject(ctx, s.bucketName, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// DeletePDF removes a PDF file from MinIO
func (s *MinIOStorage) DeletePDF(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete PDF from MinIO: %w", err)
	}

	log.Printf("Deleted PDF at %s", objectKey)
	return nil
}

// ListPDFs lists all PDF files in the bucket with optional prefix
func (s *MinIOStorage) ListPDFs(ctx context.Context, prefix string) ([]string, error) {
	var objects []string

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		objects = append(objects, object.Key)
	}

	return objects, nil
}

// HealthCheck verifies MinIO connectivity
func (s *MinIOStorage) HealthCheck(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("MinIO health check failed: %w", err)
	}
	return nil
}

// GetStats returns storage statistics
func (s *MinIOStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	objects, err := s.ListPDFs(ctx, "pdfs/")
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"bucket_name":  s.bucketName,
		"total_pdfs":   len(objects),
		"endpoint":     s.client.EndpointURL().String(),
		"health_check": "ok",
	}

	return stats, nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
