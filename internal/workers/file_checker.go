package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/cloud-file-storage/internal/app"
	file_version_service "github.com/yourusername/cloud-file-storage/internal/app/file_version"
	"github.com/yourusername/cloud-file-storage/internal/domain/storage"
)

type FileChecker struct {
	fileVersionService *file_version_service.FileVersionService
	uow                app.UnitOfWork
	storage            storage.Storage
	interval           time.Duration
	maxRetries         int
}

func NewFileChecker(
	fileVersionService *file_version_service.FileVersionService,
	uow app.UnitOfWork,
	storage storage.Storage,
	interval time.Duration,
) *FileChecker {
	return &FileChecker{
		fileVersionService: fileVersionService,
		uow:                uow,
		storage:            storage,
		interval:           interval,
		maxRetries:         3,
	}
}

// Start запускает фоновый воркер
func (fc *FileChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(fc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := fc.checkAndUpdateFiles(ctx)
			if err != nil {

			}
		}
	}
}

// checkAndUpdateFiles проверяет файлы в статусе "processing"
func (fc *FileChecker) checkAndUpdateFiles(ctx context.Context) error {

	versions, err := fc.fileVersionService.GetVersionsByStatus(ctx, "processing")
	if err != nil {
		return fmt.Errorf("failed to get versions by status: %w", err)
	}

	if len(versions) == 0 {
		return nil
	}

	successCount := 0
	errorCount := 0

	for _, version := range versions {

		exists, err := fc.storage.FileExists(ctx, version.S3Key.String())
		if err != nil {
			errorCount++
			continue
		}

		if exists {
			err := fc.fileVersionService.CompleteUpload(ctx, version.ID)
			if err != nil {
				errorCount++
				continue
			}

			successCount++
		}
	}

	return nil
}
