package file_version_service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
	"github.com/yourusername/cloud-file-storage/internal/domain/storage"
)

const uploadURLTTL = 15 * time.Minute

type FileVersionService struct {
	fileQueryRepo      file.QueryRepository
	fileCommandRepo    file.CommandRepository
	versionQueryRepo   file_version.QueryRepository
	versionCommandRepo file_version.CommandRepository
	storage            storage.Storage
	previewConsumer    queue.PreviewConsumer
	eventService       *event_service.EventService
}

func NewFileVersionService(
	fileQueryRepo file.QueryRepository,
	fileCommandRepo file.CommandRepository,
	versionQueryRepo file_version.QueryRepository,
	versionCommandRepo file_version.CommandRepository,
	storage storage.Storage,
	previewConsumer queue.PreviewConsumer,
	eventService *event_service.EventService,
) *FileVersionService {
	return &FileVersionService{
		fileQueryRepo:      fileQueryRepo,
		fileCommandRepo:    fileCommandRepo,
		versionQueryRepo:   versionQueryRepo,
		versionCommandRepo: versionCommandRepo,
		storage:            storage,
		previewConsumer:    previewConsumer,
		eventService:       eventService,
	}
}

func generateS3Key(ownerID, fileID uuid.UUID, versionNum int, name string) string {
	return fmt.Sprintf("files/%s/%s/v%d/%s", ownerID, fileID, versionNum, name)
}

func (s *FileVersionService) UpdatePreview(ctx context.Context, versionID uuid.UUID, previewKey file_version.S3Key) error {
	version, err := s.versionQueryRepo.GetByID(ctx, versionID)
	if err != nil {
		return err
	}
	version.MarkReady()
	version.PreviewS3Key = &previewKey
	s.versionCommandRepo.Save(ctx, version)

	file, err := s.GetFileByID(ctx, version.FileId)
	if err != nil {
		return err
	}

	if !file.VersionNum.Equal(version.VersionNum) {
		return nil
	}
	file.PreviewS3Key = &previewKey
	file.MarkReady()
	s.fileCommandRepo.Save(ctx, file)

	return nil
}

func (s *FileVersionService) GetFileByID(ctx context.Context, fileID uuid.UUID) (*file.File, error) {
	return s.fileQueryRepo.GetByID(ctx, fileID)
}

func (s *FileVersionService) GetVersionsByFileID(ctx context.Context, fileID uuid.UUID) ([]*file_version.FileVersion, error) {
	return s.versionQueryRepo.GetByFileID(ctx, fileID)
}

func (s *FileVersionService) GetVersionByID(ctx context.Context, versionID uuid.UUID) (*file_version.FileVersion, error) {
	version, err := s.versionQueryRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, file_version.ErrVersionNotFound
	}
	return version, nil
}

func (s *FileVersionService) GetAllVersions(ctx context.Context) ([]*file_version.FileVersion, error) {
	return s.versionQueryRepo.GetAll(ctx)
}

func (s *FileVersionService) UploadNewFile(ctx context.Context, ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*file.File, *file_version.FileVersion, string, error) {
	fileNameVO, err := file.NewFileName(name)
	if err != nil {
		return nil, nil, "", err
	}
	fileSizeVO, err := file_version.NewFileSize(size)
	if err != nil {
		return nil, nil, "", err
	}
	mimeVO, err := file_version.NewMimeType(mime)
	if err != nil {
		return nil, nil, "", err
	}
	versionNumVO, _ := file_version.NewFileVersionNum(1)

	f := file.NewFile(ownerID, fileNameVO, fileSizeVO, mimeVO, versionNumVO, sessionID)

	key := generateS3Key(ownerID, f.ID, versionNumVO.Int(), fileNameVO.String())
	s3Key, err := file_version.NewS3Key(key)
	if err != nil {
		return nil, nil, "", err
	}

	version := file_version.NewFileVersion(f.ID, sessionID, s3Key, mimeVO, fileSizeVO, versionNumVO)

	if err := s.fileCommandRepo.Save(ctx, f); err != nil {
		return nil, nil, "", err
	}
	if err := s.versionCommandRepo.Save(ctx, version); err != nil {
		return nil, nil, "", err
	}

	uploadURL, err := s.storage.GenerateUploadURL(ctx, s3Key.String(), uploadURLTTL)
	if err != nil {
		return nil, nil, "", err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionUploadedEvent(f.ID, version.ID, ownerID, f.Name.String(), version.VersionNum.Int())
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return f, version, uploadURL, nil
}

func (s *FileVersionService) UploadNewVersion(ctx context.Context, fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*file.File, *file_version.FileVersion, string, error) {
	f, err := s.fileQueryRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, nil, "", err
	}
	if f == nil {
		return nil, nil, "", file.ErrNotFound
	}

	fileNameVO, _ := file.NewFileName(name)
	fileSizeVO, _ := file_version.NewFileSize(size)
	mimeVO, _ := file_version.NewMimeType(mime)
	versionNumVO, _ := file_version.NewFileVersionNum(versionNum)

	s3Key := generateS3Key(ownerID, f.ID, versionNumVO.Int(), fileNameVO.String())
	if err != nil {
		return nil, nil, "", err
	}
	s3, err := file_version.NewS3Key(s3Key)
	if err != nil {
		return nil, nil, "", err
	}

	version := file_version.NewFileVersion(f.ID, sessionID, s3, mimeVO, fileSizeVO, versionNumVO)

	f.UpdateFromVersion(version)

	if err := s.fileCommandRepo.Save(ctx, f); err != nil {
		return nil, nil, "", err
	}
	if err := s.versionCommandRepo.Save(ctx, version); err != nil {
		return nil, nil, "", err
	}

	uploadURL, err := s.storage.GenerateUploadURL(ctx, s3Key, uploadURLTTL)
	if err != nil {
		return nil, nil, "", err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionUploadedEvent(f.ID, version.ID, ownerID, f.Name.String(), f.VersionNum.Int())
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return f, version, uploadURL, nil
}

func (s *FileVersionService) RestoreVersion(ctx context.Context, fileID, versionID uuid.UUID) error {
	f, err := s.fileQueryRepo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}
	if f == nil {
		return file.ErrNotFound
	}

	version, err := s.versionQueryRepo.GetByID(ctx, versionID)
	if err != nil {
		return err
	}
	if version == nil {
		return file_version.ErrVersionNotFound
	}

	f.UpdateFromVersion(version)
	if err := s.fileCommandRepo.Save(ctx, f); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionRestoredEvent(f.ID, version.ID)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}

func (s *FileVersionService) DeleteVersion(ctx context.Context, fileID, versionID uuid.UUID) error {
	f, err := s.fileQueryRepo.GetByID(ctx, fileID)
	if err != nil || f == nil {
		return file.ErrNotFound
	}

	version, err := s.versionQueryRepo.GetByID(ctx, versionID)
	if err != nil || version == nil {
		return file_version.ErrVersionNotFound
	}

	if version.VersionNum.Equal(f.VersionNum) {
		return file_version.ErrCannotDeleteCurr
	}
	if version.Status.Equal(file_version.FileStatusProcessing) {
		return file_version.ErrVersionProcessing
	}

	if version.PreviewS3Key != nil {
		_ = s.previewConsumer.Remove(ctx, version.ID)
		_ = s.storage.Delete(ctx, version.PreviewS3Key.String())
	}
	_ = s.storage.Delete(ctx, version.S3Key.String())

	if err := s.versionCommandRepo.Delete(ctx, versionID); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionDeletedEvent(f.ID, version.ID)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}
