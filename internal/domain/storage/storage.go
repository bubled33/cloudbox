package storage

import "time"

type Storage interface {
	GenerateUploadURL(key string, expiresIn time.Duration) (string, error)
	GenerateDownloadURL(key string, expiresIn time.Duration) (string, error)
	Delete(key string) error
}
