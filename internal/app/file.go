package app

import (
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

type FileService struct {
	fileRepo        domain.FileRepository
	fileVersionRepo domain.FileVersionRepository
}

func NewFileService(fileRepo domain.FileRepository) *FileService {
	return &FileService{
		fileRepo: fileRepo,
	}
}

func (s *FileService) GetByID(fileID uuid.UUID) (*domain.File, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.New("file not found")
	}
	return file, nil
}

func (s *FileService) GetAllByUser(userID uuid.UUID) ([]*domain.File, error) {
	files, err := s.fileRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (s *FileService) GetAll() ([]*domain.File, error) {
	files, err := s.fileRepo.GetAll()
	if err != nil {
		return nil, err
	}
	return files, nil
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

func (s *FileService) Delete(fileID uuid.UUID) error {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil || file == nil {
		return errors.New("file not found")
	}

	versions, err := s.fileVersionRepo.GetByFileID(fileID)
	if err != nil {
		return err
	}

	// Проверка, есть ли PENDING/PROCESSING версии
	for _, v := range versions {
		if v.Status == domain.FileStatusProcessing {
			return errors.New("cannot delete file, some versions are processing")
		}
	}

	// Удаляем версии
	for _, v := range versions {
		if err := s.fileVersionRepo.Delete(v.ID); err != nil {
			return err
		}
	}

	// Удаляем сам файл
	return s.fileRepo.Delete(fileID)
}
