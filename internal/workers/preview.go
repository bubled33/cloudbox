package workers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	file_version_service "github.com/yourusername/cloud-file-storage/internal/app/file_version"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
	"github.com/yourusername/cloud-file-storage/internal/domain/storage"
)

type PreviewWorker struct {
	storage            storage.Storage
	consumer           queue.PreviewConsumer
	fileVersionService file_version_service.FileVersionServiceInterface

	thumbWidth  int
	thumbHeight int
	format      imaging.Format

	retryDelay time.Duration
}

func NewPreviewWorker(
	st storage.Storage,
	cons queue.PreviewConsumer,
	fvs file_version_service.FileVersionServiceInterface,
	opts ...func(*PreviewWorker),
) *PreviewWorker {
	w := &PreviewWorker{
		storage:            st,
		consumer:           cons,
		fileVersionService: fvs,
		thumbWidth:         200,
		thumbHeight:        200,
		format:             imaging.PNG,
		retryDelay:         50 * time.Second,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func WithThumbSize(width int, height int) func(*PreviewWorker) {
	return func(p *PreviewWorker) {
		p.thumbWidth = width
		p.thumbHeight = height
	}
}

func WithFormat(format imaging.Format) func(*PreviewWorker) {
	return func(p *PreviewWorker) {
		p.format = format
	}
}

func WithRetryDelay(retryDelay time.Duration) func(*PreviewWorker) {
	return func(p *PreviewWorker) {
		p.retryDelay = retryDelay
	}
}

func (w *PreviewWorker) Handle(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		versionID, err := w.consumer.Consume(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			time.Sleep(w.retryDelay)
			continue
		}

		if err := w.ProcessVersion(ctx, versionID); err != nil {
			time.Sleep(w.retryDelay)
			continue
		}
	}
}

func (w *PreviewWorker) ProcessVersion(ctx context.Context, versionID uuid.UUID) error {
	version, err := w.fileVersionService.GetVersionByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to get version by id: %w", err)
	}

	if version == nil {
		return errors.New("version is nil")
	}

	origKey := version.S3Key.String()

	var previewKey *file_version.S3Key
	if isImageFile(origKey) {
		genKey, err := file_version.NewS3Key(addPreviewSuffix(origKey))
		if err != nil {
			return fmt.Errorf("failed to create preview s3 key: %w", err)
		}

		if err := w.generateAndUploadImagePreview(ctx, origKey, genKey.String()); err != nil {
			return fmt.Errorf("failed to generate and upload image preview: %w", err)
		}

		previewKey = &genKey
	} else {

		genKey, err := file_version.NewS3Key(addPreviewSuffix("default_preview.svg"))
		if err != nil {
			return fmt.Errorf("failed to create default preview key: %w", err)
		}
		previewKey = &genKey
	}

	if err := w.fileVersionService.UpdatePreview(ctx, versionID, *previewKey); err != nil {
		return fmt.Errorf("failed to update preview: %w", err)
	}

	return nil
}

func (w *PreviewWorker) generateAndUploadImagePreview(ctx context.Context, fileKey, previewKey string) error {
	imgData, err := w.storage.DownloadFile(ctx, fileKey)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	img, err := imaging.Decode(bytes.NewReader(imgData))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	thumb := imaging.Resize(img, w.thumbWidth, w.thumbHeight, imaging.Lanczos)

	var buf bytes.Buffer
	err = imaging.Encode(&buf, thumb, w.format)
	if err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	err = w.storage.UploadFile(ctx, previewKey, buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to upload preview to storage: %w", err)
	}

	return nil
}

func isImageFile(fileKey string) bool {
	ext := strings.ToLower(filepath.Ext(fileKey))
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
