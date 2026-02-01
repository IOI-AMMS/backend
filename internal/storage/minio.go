package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"time"

	"ioi-amms/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Service provides file storage operations
type Service interface {
	Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) (*UploadResult, error)
	Download(ctx context.Context, bucket, objectName string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, objectName string) error
	GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error)
	EnsureBucket(ctx context.Context, bucket string) error
	Health() map[string]string // [NEW] Check connectivity
}

// UploadResult contains info about an uploaded file
type UploadResult struct {
	Bucket     string `json:"bucket"`
	ObjectName string `json:"objectName"`
	Size       int64  `json:"size"`
	ETag       string `json:"etag"`
	URL        string `json:"url"`
}

// Health checks MinIO connectivity
func (s *minioService) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := "up"
	message := "Connected"

	// List buckets as a ping check
	_, err := s.client.ListBuckets(ctx)
	if err != nil {
		status = "down"
		message = err.Error()
	}

	return map[string]string{
		"status":  status,
		"message": message,
	}
}

type minioService struct {
	client   *minio.Client
	endpoint string
}

// NewService creates a new MinIO storage service
func NewService(cfg *config.Config) (Service, error) {
	endpoint := cfg.MinIO.Endpoint

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	slog.Info("Connected to MinIO", slog.String("endpoint", endpoint))

	return &minioService{
		client:   client,
		endpoint: endpoint,
	}, nil
}

// EnsureBucket creates a bucket if it doesn't exist
func (s *minioService) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		slog.Info("Created bucket", slog.String("bucket", bucket))
	}

	return nil
}

// Upload uploads a file to MinIO
func (s *minioService) Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) (*UploadResult, error) {
	// Ensure bucket exists
	if err := s.EnsureBucket(ctx, bucket); err != nil {
		return nil, err
	}

	info, err := s.client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadResult{
		Bucket:     bucket,
		ObjectName: objectName,
		Size:       info.Size,
		ETag:       info.ETag,
		URL:        fmt.Sprintf("http://%s/%s/%s", s.endpoint, bucket, objectName),
	}, nil
}

// Download retrieves a file from MinIO
func (s *minioService) Download(ctx context.Context, bucket, objectName string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return object, nil
}

// Delete removes a file from MinIO
func (s *minioService) Delete(ctx context.Context, bucket, objectName string) error {
	err := s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// GetPresignedURL generates a temporary URL for accessing a file
func (s *minioService) GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// GenerateObjectName creates a unique object name with timestamp
func GenerateObjectName(tenantID, assetID, filename string) string {
	ext := filepath.Ext(filename)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("tenants/%s/assets/%s/%d%s", tenantID, assetID, timestamp, ext)
}

// GetContentType returns a content type based on file extension
func GetContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
