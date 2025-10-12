package app

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2/storage"
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
)

const uploadURLTTL = 15 * time.Minute

type FileVersionService struct {
	fileQueryRepo      file.QueryRepository
	fileCommandRepo    file.CommandRepository
	versionQueryRepo   file_version.QueryRepository
	versionCommandRepo file_version.CommandRepository
	storage            storage.Storage
	previewQueue       queue.PreviewQueue
	eventService       *EventService
}

func NewFileVersionService(
	fileQueryRepo file.QueryRepository,
	fileCommandRepo file.CommandRepository,
	versionQueryRepo file_version.QueryRepository,
	versionCommandRepo file_version.CommandRepository,
	storage storage.Storage,
	previewQueue queue.PreviewQueue,
	eventService *EventService,
) *FileVersionService {
	return &FileVersionService{
		fileQueryRepo:      fileQueryRepo,
		fileCommandRepo:    fileCommandRepo,
		versionQueryRepo:   versionQueryRepo,
		versionCommandRepo: versionCommandRepo,
		storage:            storage,
		previewQueue:       previewQueue,
		eventService:       eventService,
	}
}

func generateS3Key(ownerID, fileID uuid.UUID, versionNum int, name string) string {
	return fmt.Sprintf("files/%s/%s/v%d/%s", ownerID, fileID, versionNum, name)
}

// --- Queries ---

func (s *FileVersionService) GetFileByID(fileID uuid.UUID) (*file.File, error) {
	return s.fileQueryRepo.GetByID(fileID)
}

func (s *FileVersionService) GetVersionsByFileID(fileID uuid.UUID) ([]*file_version.FileVersion, error) {
	return s.versionQueryRepo.GetByFileID(fileID)
}

func (s *FileVersionService) GetVersionByID(versionID uuid.UUID) (*file_version.FileVersion, error) {
	version, err := s.versionQueryRepo.GetByID(versionID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, file_version.ErrVersionNotFound
	}
	return version, nil
}

func (s *FileVersionService) GetAllVersions() ([]*file_version.FileVersion, error) {
	return s.versionQueryRepo.GetAll()
}

// --- Commands ---

func (s *FileVersionService) UploadNewFile(ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*file.File, *file_version.FileVersion, string, error) {
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
	s3, err := file_version.NewS3Key(key)
	if err != nil {
		return nil, nil, "", err
	}

	version := file_version.NewFileVersion(f.ID, sessionID, s3, mimeVO, fileSizeVO, versionNumVO)

	if err := s.fileCommandRepo.Save(f); err != nil {
		return nil, nil, "", err
	}
	if err := s.versionCommandRepo.Save(version); err != nil {
		return nil, nil, "", err
	}

	uploadURL, err := s.storage.GenerateUploadURL(s3Key, uploadURLTTL)
	if err != nil {
		return nil, nil, "", err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileUploadedEvent(f, version)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return f, version, uploadURL, nil
}

func (s *FileVersionService) UploadNewVersion(fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*file.File, *file_version.FileVersion, string, error) {
	f, err := s.fileQueryRepo.GetByID(fileID)
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
	version := file_version.NewFileVersion(f.ID, sessionID, file_version.NewS3Key{s3Key}, mimeVO, fileSizeVO, versionNumVO)

	f.UpdateFromVersion(version)

	if err := s.fileCommandRepo.Save(f); err != nil {
		return nil, nil, "", err
	}
	if err := s.versionCommandRepo.Save(version); err != nil {
		return nil, nil, "", err
	}

	uploadURL, err := s.storage.GenerateUploadURL(s3Key, uploadURLTTL)
	if err != nil {
		return nil, nil, "", err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionUploadedEvent(f, version)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return f, version, uploadURL, nil
}

func (s *FileVersionService) RestoreVersion(fileID, versionID uuid.UUID) error {
	f, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if f == nil {
		return file.ErrNotFound
	}

	version, err := s.versionQueryRepo.GetByID(versionID)
	if err != nil {
		return err
	}
	if version == nil {
		return file_version.ErrVersionNotFound
	}

	f.UpdateFromVersion(version)
	if err := s.fileCommandRepo.Save(f); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionRestoredEvent(f, version)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}

func (s *FileVersionService) DeleteVersion(fileID, versionID uuid.UUID) error {
	f, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil || f == nil {
		return file.ErrNotFound
	}

	version, err := s.versionQueryRepo.GetByID(versionID)
	if err != nil || version == nil {
		return file_version.ErrVersionNotFound
	}

	if version.VersionNum.Equal(f.VersionNum) {
		return file_version.ErrCannotDeleteCurrentVersion
	}
	if version.Status.Equal(file_version.FileStatusProcessing) {
		return file_version.ErrVersionProcessing
	}

	if version.PreviewS3Key != nil {
		_ = s.previewQueue.Remove(version.ID)
		_ = s.storage.Delete(version.PreviewS3Key.String())
	}
	_ = s.storage.Delete(version.S3Key.String())

	if err := s.versionCommandRepo.Delete(versionID); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionDeletedEvent(f, version)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}
