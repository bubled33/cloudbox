package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

const uploadURLTTL = 15 * time.Minute

type FileService struct {
	fileRepo     domain.FileRepository
	versionRepo  domain.FileVersionRepository
	storage      domain.Storage
	previewQueue domain.PreviewQueue
}

func NewFileService(fileRepo domain.FileRepository, versionRepo domain.FileVersionRepository, storage domain.Storage, previewQueue domain.PreviewQueue) *FileService {
	return &FileService{
		fileRepo:     fileRepo,
		versionRepo:  versionRepo,
		storage:      storage,
		previewQueue: previewQueue,
	}
}

func generateS3Key(ownerID, fileID uuid.UUID, versionNum int, name string) string {
	return fmt.Sprintf("files/%s/%s/v%d/%s", ownerID, fileID, versionNum, name)
}

func (s *FileService) UploadNewFile(ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*domain.File, *domain.FileVersion, string, error) {
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

func (s *FileService) UploadNewVersion(fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*domain.File, *domain.FileVersion, string, error) {
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

func (s *FileService) RenameFile(fileID uuid.UUID, name string) error {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("file not found")
	}

	file.Rename(name)
	return s.fileRepo.Save(file)
}

func (s *FileService) DeleteVersion(versionID uuid.UUID) error {
	version, err := s.versionRepo.GetByID(versionID)
	if err != nil {
		return err
	}
	if version == nil {
		return errors.New("version not found")
	}

	file, err := s.fileRepo.GetByID(version.FileId)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("file not found")
	}

	if version.VersionNum == file.VersionNum {
		return errors.New("cannot delete current version")
	}
	if version.Status != domain.FileStatusReady {
		return errors.New("cannot delete version that is not READY")
	}

	if err := s.storage.DeleteFile(version.S3Key); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}
	if version.PreviewS3Key != nil {
		if err := s.storage.DeleteFile(*version.PreviewS3Key); err != nil {
			return fmt.Errorf("failed to delete preview from storage: %w", err)
		}
	}

	return s.versionRepo.Delete(versionID)
}

func (s *FileService) DeleteFile(fileID uuid.UUID) error {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("file not found")
	}

	versions, err := s.versionRepo.GetByFileID(fileID)
	if err != nil {
		return err
	}

	for _, version := range versions {
		if version.VersionNum != file.VersionNum {
			if err := s.DeleteVersion(version.ID); err != nil {
				return err
			}
		}
	}

	if err := s.storage.DeleteFile(versions[file.VersionNum-1].S3Key); err != nil {
		return fmt.Errorf("failed to delete current file from storage: %w", err)
	}
	if versions[file.VersionNum-1].PreviewS3Key != nil {
		if err := s.storage.DeleteFile(*versions[file.VersionNum-1].PreviewS3Key); err != nil {
			return fmt.Errorf("failed to delete current preview from storage: %w", err)
		}
	}
	if err := s.versionRepo.Delete(versions[file.VersionNum-1].ID); err != nil {
		return fmt.Errorf("failed to delete current version from repo: %w", err)
	}

	return s.fileRepo.Delete(fileID)
}

func (s *FileService) RestoreVersion(fileID, versionID uuid.UUID) error {
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
