package file_version_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

// MockFileVersionService - мок для FileVersionServiceInterface
type MockFileVersionService struct {
	UpdatePreviewFunc       func(versionID uuid.UUID, previewKey file_version.S3Key) error
	GetFileByIDFunc         func(fileID uuid.UUID) (*file.File, error)
	GetVersionsByFileIDFunc func(fileID uuid.UUID) ([]*file_version.FileVersion, error)
	GetVersionByIDFunc      func(versionID uuid.UUID) (*file_version.FileVersion, error)
	GetAllVersionsFunc      func() ([]*file_version.FileVersion, error)
	UploadNewFileFunc       func(ctx context.Context, ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*file.File, *file_version.FileVersion, string, error)
	UploadNewVersionFunc    func(ctx context.Context, fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*file.File, *file_version.FileVersion, string, error)
	RestoreVersionFunc      func(fileID, versionID uuid.UUID) error
	DeleteVersionFunc       func(ctx context.Context, fileID, versionID uuid.UUID) error

	// Для отслеживания вызовов
	UpdatePreviewCalls       []UpdatePreviewCall
	GetFileByIDCalls         []GetFileByIDCall
	GetVersionsByFileIDCalls []GetVersionsByFileIDCall
	GetVersionByIDCalls      []GetVersionByIDCall
	GetAllVersionsCalls      int
	UploadNewFileCalls       []UploadNewFileCall
	UploadNewVersionCalls    []UploadNewVersionCall
	RestoreVersionCalls      []RestoreVersionCall
	DeleteVersionCalls       []DeleteVersionCall
}

// Структуры для отслеживания вызовов
type UpdatePreviewCall struct {
	VersionID  uuid.UUID
	PreviewKey file_version.S3Key
}

type GetFileByIDCall struct {
	FileID uuid.UUID
}

type GetVersionsByFileIDCall struct {
	FileID uuid.UUID
}

type GetVersionByIDCall struct {
	VersionID uuid.UUID
}

type UploadNewFileCall struct {
	Ctx       context.Context
	OwnerID   uuid.UUID
	SessionID uuid.UUID
	Name      string
	Size      uint64
	Mime      string
}

type UploadNewVersionCall struct {
	Ctx        context.Context
	FileID     uuid.UUID
	OwnerID    uuid.UUID
	SessionID  uuid.UUID
	Name       string
	Size       uint64
	Mime       string
	VersionNum int
}

type RestoreVersionCall struct {
	FileID    uuid.UUID
	VersionID uuid.UUID
}

type DeleteVersionCall struct {
	Ctx       context.Context
	FileID    uuid.UUID
	VersionID uuid.UUID
}

// Реализация методов интерфейса
func (m *MockFileVersionService) UpdatePreview(versionID uuid.UUID, previewKey file_version.S3Key) error {
	m.UpdatePreviewCalls = append(m.UpdatePreviewCalls, UpdatePreviewCall{
		VersionID:  versionID,
		PreviewKey: previewKey,
	})

	if m.UpdatePreviewFunc != nil {
		return m.UpdatePreviewFunc(versionID, previewKey)
	}
	return nil
}

func (m *MockFileVersionService) GetFileByID(fileID uuid.UUID) (*file.File, error) {
	m.GetFileByIDCalls = append(m.GetFileByIDCalls, GetFileByIDCall{FileID: fileID})

	if m.GetFileByIDFunc != nil {
		return m.GetFileByIDFunc(fileID)
	}
	return nil, nil
}

func (m *MockFileVersionService) GetVersionsByFileID(fileID uuid.UUID) ([]*file_version.FileVersion, error) {
	m.GetVersionsByFileIDCalls = append(m.GetVersionsByFileIDCalls, GetVersionsByFileIDCall{FileID: fileID})

	if m.GetVersionsByFileIDFunc != nil {
		return m.GetVersionsByFileIDFunc(fileID)
	}
	return []*file_version.FileVersion{}, nil
}

func (m *MockFileVersionService) GetVersionByID(versionID uuid.UUID) (*file_version.FileVersion, error) {
	m.GetVersionByIDCalls = append(m.GetVersionByIDCalls, GetVersionByIDCall{VersionID: versionID})

	if m.GetVersionByIDFunc != nil {
		return m.GetVersionByIDFunc(versionID)
	}
	return nil, nil
}

func (m *MockFileVersionService) GetAllVersions() ([]*file_version.FileVersion, error) {
	m.GetAllVersionsCalls++

	if m.GetAllVersionsFunc != nil {
		return m.GetAllVersionsFunc()
	}
	return []*file_version.FileVersion{}, nil
}

func (m *MockFileVersionService) UploadNewFile(ctx context.Context, ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*file.File, *file_version.FileVersion, string, error) {
	m.UploadNewFileCalls = append(m.UploadNewFileCalls, UploadNewFileCall{
		Ctx:       ctx,
		OwnerID:   ownerID,
		SessionID: sessionID,
		Name:      name,
		Size:      size,
		Mime:      mime,
	})

	if m.UploadNewFileFunc != nil {
		return m.UploadNewFileFunc(ctx, ownerID, sessionID, name, size, mime)
	}
	return nil, nil, "", nil
}

func (m *MockFileVersionService) UploadNewVersion(ctx context.Context, fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*file.File, *file_version.FileVersion, string, error) {
	m.UploadNewVersionCalls = append(m.UploadNewVersionCalls, UploadNewVersionCall{
		Ctx:        ctx,
		FileID:     fileID,
		OwnerID:    ownerID,
		SessionID:  sessionID,
		Name:       name,
		Size:       size,
		Mime:       mime,
		VersionNum: versionNum,
	})

	if m.UploadNewVersionFunc != nil {
		return m.UploadNewVersionFunc(ctx, fileID, ownerID, sessionID, name, size, mime, versionNum)
	}
	return nil, nil, "", nil
}

func (m *MockFileVersionService) RestoreVersion(fileID, versionID uuid.UUID) error {
	m.RestoreVersionCalls = append(m.RestoreVersionCalls, RestoreVersionCall{
		FileID:    fileID,
		VersionID: versionID,
	})

	if m.RestoreVersionFunc != nil {
		return m.RestoreVersionFunc(fileID, versionID)
	}
	return nil
}

func (m *MockFileVersionService) DeleteVersion(ctx context.Context, fileID, versionID uuid.UUID) error {
	m.DeleteVersionCalls = append(m.DeleteVersionCalls, DeleteVersionCall{
		Ctx:       ctx,
		FileID:    fileID,
		VersionID: versionID,
	})

	if m.DeleteVersionFunc != nil {
		return m.DeleteVersionFunc(ctx, fileID, versionID)
	}
	return nil
}
