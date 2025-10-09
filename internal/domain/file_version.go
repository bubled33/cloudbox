package domain

import (
	"fmt"
	"time"

	uuid "github.com/google/uuid"
)

type FileVersion struct {
	ID                  uuid.UUID
	FileId              uuid.UUID
	UploadedBySessionId uuid.UUID

	S3Key        string
	Mime         string
	PreviewS3Key *string

	Status FileStatus

	Size       uint64
	VersionNum int

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFileVersion(
	fileID uuid.UUID,
	uploadedBySessionID uuid.UUID,
	s3Key string,
	mime string,
	size uint64,
	versionNum int,
) *FileVersion {
	now := time.Now()
	return &FileVersion{
		ID:                  uuid.New(),
		FileId:              fileID,
		UploadedBySessionId: uploadedBySessionID,
		S3Key:               s3Key,
		Mime:                mime,
		PreviewS3Key:        nil,
		Status:              FileStatusUploaded,
		Size:                size,
		VersionNum:          versionNum,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func (fv *FileVersion) SetStatus(status FileStatus) error {
	switch status {
	case FileStatusUploaded, FileStatusProcessing, FileStatusReady, FileStatusFailed:
		fv.Status = status
		fv.UpdatedAt = time.Now()
		return nil
	default:
		return fmt.Errorf("invalid file status: %s", status)
	}
}

func (fv *FileVersion) SetPreviewS3Key(key string) {
	fv.PreviewS3Key = &key
	fv.UpdatedAt = time.Now()
}

func (fv *FileVersion) Rename(newName string) {
	fv.Mime = newName
	fv.UpdatedAt = time.Now()
}
func (fv *FileVersion) SetMime(mime string) {
	fv.Mime = mime
	fv.UpdatedAt = time.Now()
}

func (fv *FileVersion) SetSize(size uint64) error {
	const MaxFileSize = 10 * 1024 * 1024 * 1024
	if size > MaxFileSize {
		return fmt.Errorf("file size exceeds 10GB limit")
	}
	fv.Size = size
	fv.UpdatedAt = time.Now()
	return nil
}

func (fv *FileVersion) SetVersionNum(versionNum int) {
	fv.VersionNum = versionNum
	fv.UpdatedAt = time.Now()
}

func (f *FileVersion) MarkUploaded() {
	f.Status = FileStatusUploaded
	f.UpdatedAt = time.Now()
}

func (f *FileVersion) MarkProcessing() {
	f.Status = FileStatusProcessing
	f.UpdatedAt = time.Now()
}

func (f *FileVersion) MarkReady() {
	f.Status = FileStatusReady
	f.UpdatedAt = time.Now()
}

func (f *FileVersion) MarkFailed() {
	f.Status = FileStatusFailed
	f.UpdatedAt = time.Now()
}
