package app

import (
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

type FileService struct {
	fileQueryRepo      domain.FileQueryRepository
	fileCommandRepo    domain.FileCommandRepository
	versionQueryRepo   domain.FileVersionQueryRepository
	versionCommandRepo domain.FileVersionCommandRepository
}

func NewFileService(
	fileQueryRepo domain.FileQueryRepository,
	fileCommandRepo domain.FileCommandRepository,
	versionQueryRepo domain.FileVersionQueryRepository,
	versionCommandRepo domain.FileVersionCommandRepository,
) *FileService {
	return &FileService{
		fileQueryRepo:      fileQueryRepo,
		fileCommandRepo:    fileCommandRepo,
		versionQueryRepo:   versionQueryRepo,
		versionCommandRepo: versionCommandRepo,
	}
}

// --- Queries ---

func (s *FileService) GetByID(fileID uuid.UUID) (*domain.File, error) {
	file, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, ErrFileNotFound
	}
	return file, nil
}

func (s *FileService) GetAllByUser(userID uuid.UUID) ([]*domain.File, error) {
	return s.fileQueryRepo.GetByUserID(userID)
}

func (s *FileService) GetAll() ([]*domain.File, error) {
	return s.fileQueryRepo.GetAll()
}

// --- Commands ---

func (s *FileService) RenameFile(fileID uuid.UUID, name string) error {
	file, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return ErrFileNotFound
	}

	file.Rename(name)
	return s.fileCommandRepo.Save(file)
}

func (s *FileService) Delete(fileID uuid.UUID) error {
	file, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil || file == nil {
		return ErrFileNotFound
	}

	versions, err := s.versionQueryRepo.GetByFileID(fileID)
	if err != nil {
		return err
	}

	// Проверка статусов версий
	for _, v := range versions {
		if v.Status == domain.FileStatusProcessing {
			return ErrVersionProcessing
		}
	}

	// Удаляем версии
	for _, v := range versions {
		if err := s.versionCommandRepo.Delete(v.ID); err != nil {
			return err
		}
	}

	// Удаляем сам файл
	return s.fileCommandRepo.Delete(fileID)
}
