package app

import (
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

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

func (s *FileService) UploadNewFile(ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*domain.File, *domain.FileVersion, error) {
	file := domain.NewFile(ownerID, name, size, mime, 1, sessionID)
	version := domain.NewFileVersion(file.ID, sessionID, "", mime, size, 1)

	if err := s.fileRepo.Save(file); err != nil {
		return nil, nil, err
	}
	if err := s.versionRepo.Save(version); err != nil {
		return nil, nil, err
	}

	if err := s.previewQueue.Enqueue(version.ID); err != nil {
		return nil, nil, err
	}

	return file, version, nil
}
