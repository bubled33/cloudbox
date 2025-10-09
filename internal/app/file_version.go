package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

const uploadURLTTL = 15 * time.Minute

type FileVersionService struct {
	fileRepo     domain.FileRepository
	versionRepo  domain.FileVersionRepository
	storage      domain.Storage
	previewQueue domain.PreviewQueue
}

func NewFileVersionService(
	fileRepo domain.FileRepository,
	versionRepo domain.FileVersionRepository,
	storage domain.Storage,
	previewQueue domain.PreviewQueue,
) *FileVersionService {
	return &FileVersionService{
		fileRepo:     fileRepo,
		versionRepo:  versionRepo,
		storage:      storage,
		previewQueue: previewQueue,
	}
}

func generateS3Key(ownerID, fileID uuid.UUID, versionNum int, name string) string {
	return fmt.Sprintf("files/%s/%s/v%d/%s", ownerID, fileID, versionNum, name)
}

// Создание нового файла и первой версии
func (s *FileVersionService) UploadNewFile(ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*domain.File, *domain.FileVersion, string, error) {
	file := domain.NewFile(ownerID, name, size, mime, 1, sessionID)
	s3Key := generateS3Key(ownerID, file.ID, 1, name)
	version := domain.NewFileVersion(file.ID, sessionID, s3Key, mime, size, 1)

	if err := s.fileRepo.Save(file); err != nil {
		return nil, nil, "", err
	}
	if err := s.versionRepo.Save(version); err != nil {
		return nil, nil, "", err
	}

	uploadURL, err := s.storage.GenerateUploadURL(s3Key, uploadURLTTL)
	if err != nil {
		return nil, nil, "", err
	}

	return file, version, uploadURL, nil
}

// Загрузка новой версии файла
func (s *FileVersionService) UploadNewVersion(fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*domain.File, *domain.FileVersion, string, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, nil, "", err
	}
	if file == nil {
		return nil, nil, "", errors.New("file not found")
	}

	s3Key := generateS3Key(ownerID, file.ID, versionNum, name)
	version := domain.NewFileVersion(file.ID, sessionID, s3Key, mime, size, versionNum)
	file.UpdateFromVersion(version)

	if err := s.fileRepo.Save(file); err != nil {
		return nil, nil, "", err
	}
	if err := s.versionRepo.Save(version); err != nil {
		return nil, nil, "", err
	}

	uploadURL, err := s.storage.GenerateUploadURL(s3Key, uploadURLTTL)
	if err != nil {
		return nil, nil, "", err
	}

	return file, version, uploadURL, nil
}

// Получение всех версий файла
func (s *FileVersionService) GetByFileID(fileID uuid.UUID) ([]*domain.FileVersion, error) {
	versions, err := s.versionRepo.GetByFileID(fileID)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// Получение версии по ID
func (s *FileVersionService) GetByID(versionID uuid.UUID) (*domain.FileVersion, error) {
	version, err := s.versionRepo.GetByID(versionID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, errors.New("version not found")
	}
	return version, nil
}

// Получение всех версий в системе
func (s *FileVersionService) GetAll() ([]*domain.FileVersion, error) {
	versions, err := s.versionRepo.GetAll()
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// Восстановление версии
func (s *FileVersionService) Restore(fileID, versionID uuid.UUID) error {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("file not found")
	}

	version, err := s.versionRepo.GetByID(versionID)
	if err != nil {
		return err
	}
	if version == nil {
		return errors.New("version not found")
	}

	file.UpdateFromVersion(version)
	return s.fileRepo.Save(file)
}

func (s *FileVersionService) Delete(fileID, versionID uuid.UUID) error {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil || file == nil {
		return errors.New("file not found")
	}

	version, err := s.versionRepo.GetByID(versionID)
	if err != nil || version == nil {
		return errors.New("version not found")
	}

	if version.VersionNum == file.VersionNum {
		return errors.New("cannot delete current version")
	}

	switch version.Status {
	case domain.FileStatusProcessing:
		return errors.New("cannot delete version while processing (pending)")

	case domain.FileStatusFailed, domain.FileStatusUploaded:
		if version.PreviewS3Key != nil {
			s.previewQueue.Remove(version.ID)
		}
		if err := s.storage.Delete(version.S3Key); err != nil {
			return err
		}

	case domain.FileStatusReady:
		if version.PreviewS3Key != nil {
			if err := s.storage.Delete(*version.PreviewS3Key); err != nil {
				return err
			}
		}
		if err := s.storage.Delete(version.S3Key); err != nil {
			return err
		}
	}

	return s.versionRepo.Delete(versionID)
}
