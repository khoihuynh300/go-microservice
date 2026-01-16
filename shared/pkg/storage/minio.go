package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

type MinIOConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}

func NewMinIOStorage(cfg MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	storage := &MinIOStorage{
		client:     client,
		bucketName: cfg.BucketName,
		endpoint:   cfg.Endpoint,
		useSSL:     cfg.UseSSL,
	}

	if err := storage.ensureBucket(context.Background()); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *MinIOStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		return fmt.Errorf("bucket %s does not exist", s.bucketName)
	}

	return nil
}

func (s *MinIOStorage) Upload(ctx context.Context, input *UploadInput) (*UploadOutput, error) {
	if input.File == nil {
		return nil, fmt.Errorf("file is required")
	}
	if input.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	key := s.generateKey(input.Filename, input.Folder)

	_, err := s.client.PutObject(
		ctx,
		s.bucketName,
		key,
		input.File,
		input.Size,
		minio.PutObjectOptions{
			ContentType: input.ContentType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadOutput{
		URL:      s.GetURL(key),
		Key:      key,
		Size:     input.Size,
		Filename: input.Filename,
	}, nil
}

func (s *MinIOStorage) Delete(ctx context.Context, url string) error {
	key := s.extractKeyFromURL(url)
	if key == "" {
		return fmt.Errorf("invalid URL")
	}

	err := s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *MinIOStorage) GetPresignedUploadURL(ctx context.Context, input *PresignedUploadInput) (*PresignedUploadOutput, error) {
	key := s.generateKey(input.Filename, input.Folder)

	presignedURL, err := s.client.PresignedPutObject(
		ctx,
		s.bucketName,
		key,
		input.ExpiresIn,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return &PresignedUploadOutput{
		UploadURL: presignedURL.String(),
		Key:       key,
		FinalURL:  s.GetURL(key),
		ExpiresIn: input.ExpiresIn,
		ExpiresAt: time.Now().Add(input.ExpiresIn),
	}, nil
}

func (s *MinIOStorage) GetPresignedDownloadURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(
		ctx,
		s.bucketName,
		key,
		expiresIn,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create presigned download URL: %w", err)
	}

	return presignedURL.String(), nil
}

func (s *MinIOStorage) GetURL(key string) string {
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucketName, key)
}

func (s *MinIOStorage) generateKey(filename, folder string) string {
	ext := filepath.Ext(filename)
	baseFilename := strings.TrimSuffix(filepath.Base(filename), ext)
	uniqueID := uuid.New().String()

	baseFilename = strings.ReplaceAll(baseFilename, " ", "-")
	baseFilename = strings.ToLower(baseFilename)

	return fmt.Sprintf("%s/%s-%s%s",
		folder,
		baseFilename,
		uniqueID[:8],
		ext,
	)
}

func (s *MinIOStorage) extractKeyFromURL(url string) string {
	parts := strings.SplitN(url, s.bucketName+"/", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
