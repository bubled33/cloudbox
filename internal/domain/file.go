package domain

import (
	"time"

	uuid "github.com/google/uuid"
)

type FileStatus string

const (
	FileStatusUploaded   FileStatus = "UPLOADED"
	FileStatusProcessing FileStatus = "PROCESSING"
	FileStatusReady      FileStatus = "READY"
	FileStatusFailed     FileStatus = "FAILED"
)

type File struct {
	ID                  uuid.UUID
	OwnerID             uuid.UUID
	UploadedBySessionId uuid.UUID

	Name         string
	Mime         string
	PreviewS3Key *string

	Status FileStatus

	Size       uint64
	VersionNum int

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFile(ownerID uuid.UUID, name string, size uint64, mime string, versionNum int, uploadedBySessionId uuid.UUID) *File {
	now := time.Now()
	return &File{
		ID:                  uuid.New(),
		OwnerID:             ownerID,
		Name:                name,
		Size:                size,
		Mime:                mime,
		VersionNum:          versionNum,
		PreviewS3Key:        nil,
		Status:              FileStatusUploaded,
		UploadedBySessionId: uploadedBySessionId,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func (f *File) Rename(newName string) {
	f.Name = newName
	f.UpdatedAt = time.Now()
}

func (f *File) UpdateFromVersion(fv *FileVersion) {
	f.Mime = fv.Mime
	f.PreviewS3Key = fv.PreviewS3Key
	f.Status = fv.Status
	f.Size = fv.Size
	f.VersionNum = fv.VersionNum
	f.UpdatedAt = time.Now()
}

func (f *File) MarkUploaded() {
	f.Status = FileStatusUploaded
	f.UpdatedAt = time.Now()
}

func (f *File) MarkProcessing() {
	f.Status = FileStatusProcessing
	f.UpdatedAt = time.Now()
}

func (f *File) MarkReady() {
	f.Status = FileStatusReady
	f.UpdatedAt = time.Now()
}

func (f *File) MarkFailed() {
	f.Status = FileStatusFailed
	f.UpdatedAt = time.Now()
}
