package file_version_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

// MockFileVersionService удовлетворяет FileVersionServiceInterface
type MockFileVersionService struct {
	// Функции
	UpdatePreviewFunc       func(ctx context.Context, versionID uuid.UUID, previewKey file_version.S3Key) error
	GetFileByIDFunc         func(ctx context.Context, fileID uuid.UUID) (*file.File, error)
	GetVersionsByFileIDFunc func(ctx context.Context, fileID uuid.UUID) ([]*file_version.FileVersion, error)
	GetVersionByIDFunc      func(ctx context.Context, versionID uuid.UUID) (*file_version.FileVersion, error)
	GetAllVersionsFunc      func(ctx context.Context) ([]*file_version.FileVersion, error)
	UploadNewFileFunc       func(ctx context.Context, ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*file.File, *file_version.FileVersion, string, error)
	UploadNewVersionFunc    func(ctx context.Context, fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*file.File, *file_version.FileVersion, string, error)
	RestoreVersionFunc      func(ctx context.Context, fileID, versionID uuid.UUID) error
	DeleteVersionFunc       func(ctx context.Context, fileID, versionID uuid.UUID) error

	// Трассировка вызовов
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
	Ctx        context.Context
	VersionID  uuid.UUID
	PreviewKey file_version.S3Key
}

type GetFileByIDCall struct {
	Ctx    context.Context
	FileID uuid.UUID
}

type GetVersionsByFileIDCall struct {
	Ctx    context.Context
	FileID uuid.UUID
}

type GetVersionByIDCall struct {
	Ctx       context.Context
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
	Ctx       context.Context
	FileID    uuid.UUID
	VersionID uuid.UUID
}

type DeleteVersionCall struct {
	Ctx       context.Context
	FileID    uuid.UUID
	VersionID uuid.UUID
}

func (m *MockFileVersionService) UpdatePreview(ctx context.Context, versionID uuid.UUID, previewKey file_version.S3Key) error {
	m.UpdatePreviewCalls = append(m.UpdatePreviewCalls, UpdatePreviewCall{
		Ctx:        ctx,
		VersionID:  versionID,
		PreviewKey: previewKey,
	})
	if m.UpdatePreviewFunc != nil {
		return m.UpdatePreviewFunc(ctx, versionID, previewKey)
	}
	return nil
}

func (m *MockFileVersionService) GetFileByID(ctx context.Context, fileID uuid.UUID) (*file.File, error) {
	m.GetFileByIDCalls = append(m.GetFileByIDCalls, GetFileByIDCall{
		Ctx:    ctx,
		FileID: fileID,
	})
	if m.GetFileByIDFunc != nil {
		return m.GetFileByIDFunc(ctx, fileID)
	}
	return nil, nil
}

func (m *MockFileVersionService) GetVersionsByFileID(ctx context.Context, fileID uuid.UUID) ([]*file_version.FileVersion, error) {
	m.GetVersionsByFileIDCalls = append(m.GetVersionsByFileIDCalls, GetVersionsByFileIDCall{
		Ctx:    ctx,
		FileID: fileID,
	})
	if m.GetVersionsByFileIDFunc != nil {
		return m.GetVersionsByFileIDFunc(ctx, fileID)
	}
	return []*file_version.FileVersion{}, nil
}

func (m *MockFileVersionService) GetVersionByID(ctx context.Context, versionID uuid.UUID) (*file_version.FileVersion, error) {
	m.GetVersionByIDCalls = append(m.GetVersionByIDCalls, GetVersionByIDCall{
		Ctx:       ctx,
		VersionID: versionID,
	})
	if m.GetVersionByIDFunc != nil {
		return m.GetVersionByIDFunc(ctx, versionID)
	}
	return nil, nil
}

func (m *MockFileVersionService) GetAllVersions(ctx context.Context) ([]*file_version.FileVersion, error) {
	m.GetAllVersionsCalls++
	if m.GetAllVersionsFunc != nil {
		return m.GetAllVersionsFunc(ctx)
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

func (m *MockFileVersionService) RestoreVersion(ctx context.Context, fileID, versionID uuid.UUID) error {
	m.RestoreVersionCalls = append(m.RestoreVersionCalls, RestoreVersionCall{
		Ctx:       ctx,
		FileID:    fileID,
		VersionID: versionID,
	})
	if m.RestoreVersionFunc != nil {
		return m.RestoreVersionFunc(ctx, fileID, versionID)
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
