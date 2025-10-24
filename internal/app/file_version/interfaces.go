package file_version_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

type FileVersionServiceInterface interface {
	UpdatePreview(versionID uuid.UUID, previewKey file_version.S3Key) error
	GetFileByID(fileID uuid.UUID) (*file.File, error)
	GetVersionsByFileID(fileID uuid.UUID) ([]*file_version.FileVersion, error)
	GetVersionByID(versionID uuid.UUID) (*file_version.FileVersion, error)
	GetAllVersions() ([]*file_version.FileVersion, error)
	UploadNewFile(ctx context.Context, ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*file.File, *file_version.FileVersion, string, error)
	UploadNewVersion(ctx context.Context, fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*file.File, *file_version.FileVersion, string, error)
	RestoreVersion(fileID, versionID uuid.UUID) error
	DeleteVersion(ctx context.Context, fileID, versionID uuid.UUID) error
}
