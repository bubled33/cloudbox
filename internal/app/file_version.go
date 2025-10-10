package app

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

const uploadURLTTL = 15 * time.Minute

type FileVersionService struct {
	fileQueryRepo      domain.FileQueryRepository
	fileCommandRepo    domain.FileCommandRepository
	versionQueryRepo   domain.FileVersionQueryRepository
	versionCommandRepo domain.FileVersionCommandRepository
	storage            domain.Storage
	previewQueue       domain.PreviewQueue
}

func NewFileVersionService(
	fileQueryRepo domain.FileQueryRepository,
	fileCommandRepo domain.FileCommandRepository,
	versionQueryRepo domain.FileVersionQueryRepository,
	versionCommandRepo domain.FileVersionCommandRepository,
	storage domain.Storage,
	previewQueue domain.PreviewQueue,
) *FileVersionService {
	return &FileVersionService{
		fileQueryRepo:      fileQueryRepo,
		fileCommandRepo:    fileCommandRepo,
		versionQueryRepo:   versionQueryRepo,
		versionCommandRepo: versionCommandRepo,
		storage:            storage,
		previewQueue:       previewQueue,
	}
}

func generateS3Key(ownerID, fileID uuid.UUID, versionNum int, name string) string {
	return fmt.Sprintf("files/%s/%s/v%d/%s", ownerID, fileID, versionNum, name)
}

// --- Queries ---

func (s *FileVersionService) GetFileByID(fileID uuid.UUID) (*domain.File, error) {
	return s.fileQueryRepo.GetByID(fileID)
}

func (s *FileVersionService) GetVersionsByFileID(fileID uuid.UUID) ([]*domain.FileVersion, error) {
	return s.versionQueryRepo.GetByFileID(fileID)
}

func (s *FileVersionService) GetVersionByID(versionID uuid.UUID) (*domain.FileVersion, error) {
	version, err := s.versionQueryRepo.GetByID(versionID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, ErrVersionNotFound
	}
	return version, nil
}

func (s *FileVersionService) GetAllVersions() ([]*domain.FileVersion, error) {
	return s.versionQueryRepo.GetAll()
}

// --- Commands ---

func (s *FileVersionService) UploadNewFile(ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*domain.File, *domain.FileVersion, string, error) {
	file := domain.NewFile(ownerID, name, size, mime, 1, sessionID)
	s3Key := generateS3Key(ownerID, file.ID, 1, name)
	version := domain.NewFileVersion(file.ID, sessionID, s3Key, mime, size, 1)

	if err := s.fileCommandRepo.Save(file); err != nil {
		return nil, nil, "", err
	}
	if err := s.versionCommandRepo.Save(version); err != nil {
		return nil, nil, "", err
	}

	uploadURL, err := s.storage.GenerateUploadURL(s3Key, uploadURLTTL)
	if err != nil {
		return nil, nil, "", err
	}

	return file, version, uploadURL, nil
}

func (s *FileVersionService) UploadNewVersion(fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*domain.File, *domain.FileVersion, string, error) {
	file, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil {
		return nil, nil, "", err
	}
	if file == nil {
		return nil, nil, "", ErrFileNotFound
	}

	s3Key := generateS3Key(ownerID, file.ID, versionNum, name)
	version := domain.NewFileVersion(file.ID, sessionID, s3Key, mime, size, versionNum)
	file.UpdateFromVersion(version)

	if err := s.fileCommandRepo.Save(file); err != nil {
		return nil, nil, "", err
	}
	if err := s.versionCommandRepo.Save(version); err != nil {
		return nil, nil, "", err
	}

	uploadURL, err := s.storage.GenerateUploadURL(s3Key, uploadURLTTL)
	if err != nil {
		return nil, nil, "", err
	}

	return file, version, uploadURL, nil
}

func (s *FileVersionService) RestoreVersion(fileID, versionID uuid.UUID) error {
	file, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return ErrFileNotFound
	}

	version, err := s.versionQueryRepo.GetByID(versionID)
	if err != nil {
		return err
	}
	if version == nil {
		return ErrVersionNotFound
	}

	file.UpdateFromVersion(version)
	return s.fileCommandRepo.Save(file)
}

func (s *FileVersionService) DeleteVersion(fileID, versionID uuid.UUID) error {
	file, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil || file == nil {
		return ErrFileNotFound
	}

	version, err := s.versionQueryRepo.GetByID(versionID)
	if err != nil || version == nil {
		return ErrVersionNotFound
	}

	if version.VersionNum == file.VersionNum {
		return ErrCannotDeleteCurr
	}

	if version.Status == domain.FileStatusProcessing {
		return ErrVersionProcessing
	}

	// Удаляем из хранилища
	if version.PreviewS3Key != nil {
		_ = s.previewQueue.Remove(version.ID)
		_ = s.storage.Delete(*version.PreviewS3Key)
	}
	_ = s.storage.Delete(version.S3Key)

	return s.versionCommandRepo.Delete(versionID)
}
