package storage

import (
	"context"
	"io"
	"time"
)

type Storage interface {
	Upload(ctx context.Context, input *UploadInput) (*UploadOutput, error)
	Delete(ctx context.Context, url string) error
	GetURL(key string) string
	GetPresignedUploadURL(ctx context.Context, input *PresignedUploadInput) (*PresignedUploadOutput, error)
	GetPresignedDownloadURL(ctx context.Context, key string, expiresIn time.Duration) (string, error)
}

type UploadInput struct {
	File        io.Reader
	Filename    string
	ContentType string
	Size        int64
	Folder      string
}

type UploadOutput struct {
	URL      string
	Key      string
	Size     int64
	Filename string
}

type PresignedUploadInput struct {
	Filename    string
	ContentType string
	Folder      string
	ExpiresIn   time.Duration
}

type PresignedUploadOutput struct {
	UploadURL string
	Key       string
	FinalURL  string
	ExpiresIn time.Duration
	ExpiresAt time.Time
}
