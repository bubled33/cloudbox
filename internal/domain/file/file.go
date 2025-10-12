package file

import (
	"time"

	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

type File struct {
	ID                  uuid.UUID
	OwnerID             uuid.UUID
	UploadedBySessionId uuid.UUID

	Name         FileName
	Mime         file_version.MimeType
	PreviewS3Key *file_version.S3Key

	Status file_version.FileStatus

	Size       file_version.FileSize
	VersionNum file_version.FileVersionNum

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewFile создаёт новый файл с валидацией всех VO
func NewFile(
	ownerID uuid.UUID,
	name FileName,
	size file_version.FileSize,
	mime file_version.MimeType,
	versionNum file_version.FileVersionNum,
	uploadedBySessionId uuid.UUID,
) *File {
	now := time.Now()
	return &File{
		ID:                  uuid.New(),
		OwnerID:             ownerID,
		Name:                name,
		Size:                size,
		Mime:                mime,
		VersionNum:          versionNum,
		PreviewS3Key:        nil,
		Status:              file_version.FileStatusUploaded,
		UploadedBySessionId: uploadedBySessionId,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// Rename задаёт новое имя файла
func (f *File) Rename(newName FileName) {
	f.Name = newName
	f.UpdatedAt = time.Now()
}

// UpdateFromVersion обновляет свойства файла из FileVersion VO
func (f *File) UpdateFromVersion(fv *file_version.FileVersion) {
	f.Mime = fv.Mime
	f.PreviewS3Key = fv.PreviewS3Key
	f.Status = fv.Status
	f.Size = fv.Size
	f.VersionNum = fv.VersionNum
	f.UpdatedAt = time.Now()
}

// Методы для обновления статуса файла через VO
func (f *File) MarkUploaded() {
	f.Status = file_version.FileStatusUploaded
	f.UpdatedAt = time.Now()
}

func (f *File) MarkProcessing() {
	f.Status = file_version.FileStatusProcessing
	f.UpdatedAt = time.Now()
}

func (f *File) MarkReady() {
	f.Status = file_version.FileStatusReady
	f.UpdatedAt = time.Now()
}

func (f *File) MarkFailed() {
	f.Status = file_version.FileStatusFailed
	f.UpdatedAt = time.Now()
}
