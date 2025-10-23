package workers

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/app"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
	"github.com/yourusername/cloud-file-storage/internal/domain/storage"
)

type PreviewWorker struct {
	storage            storage.Storage
	consumer           queue.PreviewConsumer
	fileVersionService app.FileVersionService
}

func (w *PreviewWorker) Handle(ctx context.Context, versionID uuid.UUID) error {
	for {
		versionID, err := w.consumer.Consume(ctx)
		if err != nil {
			return err
		}

		version, err := w.fileVersionService.GetVersionByID(versionID)
		if err != nil {
			return err
		}

		if isImageFile(version.S3Key.String()) {
			previewKey, err := file_version.NewS3Key(addPreviewSuffix(version.S3Key.String()))
			if err != nil {
				return err
			}
			err = w.generateAndUploadImagePreview(ctx, version.S3Key.String(), previewKey.String())
			if err != nil {
				return err
			}

			err = w.fileVersionService.UpdatePreview(versionID, previewKey)
			if err != nil {
				return err
			}

			continue
		}

		previewKey, err := file_version.NewS3Key(addPreviewSuffix("default_preview.svg"))
		if err != nil {
			return err
		}
		err = w.fileVersionService.UpdatePreview(versionID, previewKey)
		if err != nil {
			return err
		}

	}
}

func (w *PreviewWorker) generateAndUploadImagePreview(ctx context.Context, fileKey, previewKey string) error {
	imgData, err := w.storage.DownloadFile(ctx, fileKey)
	if err != nil {
		return err
	}

	img, err := imaging.Decode(bytes.NewReader(imgData))
	if err != nil {
		return err
	}

	thumb := imaging.Resize(img, 200, 200, imaging.Lanczos)

	var buf bytes.Buffer
	err = imaging.Encode(&buf, thumb, imaging.PNG)
	if err != nil {
		return err
	}

	err = w.storage.UploadFile(ctx, previewKey, buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func isImageFile(fileKey string) bool {
	ext := filepath.Ext(fileKey)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return true
	default:
		return false
	}
}

func addPreviewSuffix(key string) string {
	dotPos := strings.LastIndex(key, ".")
	if dotPos == -1 {
		return key + "_preview"
	}
	return key[:dotPos] + "_preview" + key[dotPos:]
}
