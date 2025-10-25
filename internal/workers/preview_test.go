package workers

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/google/uuid"
	file_version_service "github.com/yourusername/cloud-file-storage/internal/app/file_version"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

func createTestPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestPreviewWorker_Handle(t *testing.T) {
	mockStorage := &MockStorage{
		DownloadFileFunc: func(ctx context.Context, key string) ([]byte, error) {
			return createTestPNG(), nil
		},
		UploadFileFunc: func(ctx context.Context, key string, fileData []byte) error {
			return nil
		},
	}

	callCount := 0
	mockConsumer := &MockPreviewConsumer{
		ConsumeFunc: func(ctx context.Context) (uuid.UUID, error) {
			callCount++
			if callCount > 2 {
				return uuid.Nil, context.Canceled
			}
			return uuid.New(), nil
		},
	}

	s3Key, _ := file_version.NewS3Key("files/test.jpg")
	mockService := &file_version_service.MockFileVersionService{
		GetVersionByIDFunc: func(id uuid.UUID) (*file_version.FileVersion, error) {
			return &file_version.FileVersion{ID: id, S3Key: s3Key}, nil
		},
		UpdatePreviewFunc: func(versionID uuid.UUID, previewKey file_version.S3Key) error {
			return nil
		},
	}

	worker := NewPreviewWorker(mockStorage, mockConsumer, mockService)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	worker.Handle(ctx)

	if callCount < 2 {
		t.Errorf("expected at least 2 calls, got %d", callCount)
	}
}
