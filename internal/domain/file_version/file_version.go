package file_version

import (
	"time"

	uuid "github.com/google/uuid"
)

type FileVersion struct {
	ID                  uuid.UUID
	FileId              uuid.UUID
	UploadedBySessionId uuid.UUID

	S3Key        S3Key
	Mime         MimeType
	PreviewS3Key *S3Key

	Status FileStatus

	Size       FileSize
	VersionNum FileVersionNum

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFileVersion(
	fileID uuid.UUID,
	uploadedBySessionID uuid.UUID,
	s3Key S3Key,
	mime MimeType,
	size FileSize,
	versionNum FileVersionNum,
) *FileVersion {
	now := time.Now()
	return &FileVersion{
		ID:                  uuid.New(),
		FileId:              fileID,
		UploadedBySessionId: uploadedBySessionID,
		S3Key:               s3Key,
		Mime:                mime,
		PreviewS3Key:        nil,
		Status:              FileStatusProcessing,
		Size:                size,
		VersionNum:          versionNum,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func (fv *FileVersion) SetStatus(status FileStatus) error {
	_, err := NewFileStatus(string(status))
	if err != nil {
		return err
	}
	fv.Status = status
	fv.UpdatedAt = time.Now()
	return nil
}

func (fv *FileVersion) SetPreviewS3Key(key S3Key) {
	fv.PreviewS3Key = &key
	fv.UpdatedAt = time.Now()
}

func (fv *FileVersion) SetMime(mime MimeType) {
	fv.Mime = mime
	fv.UpdatedAt = time.Now()
}

func (fv *FileVersion) SetSize(size FileSize) {
	fv.Size = size
	fv.UpdatedAt = time.Now()
}

func (fv *FileVersion) SetVersionNum(versionNum FileVersionNum) {
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
