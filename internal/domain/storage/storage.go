package storage

import (
	"context"
	"time"
)

type Storage interface {
	GenerateUploadURL(ctx context.Context, key string, expiresIn time.Duration) (string, error)
	GenerateDownloadURL(ctx context.Context, key string, expiresIn time.Duration) (string, error)
	Delete(ctx context.Context, key string) error

	UploadFile(ctx context.Context, key string, fileData []byte) error
	DownloadFile(ctx context.Context, key string) ([]byte, error)
}
