package app

import (
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

type FileService struct {
	fileQueryRepo      file.QueryRepository
	fileCommandRepo    file.CommandRepository
	versionQueryRepo   file_version.QueryRepository
	versionCommandRepo file_version.CommandRepository
	eventService       *EventService
}

func NewFileService(
	fileQueryRepo file.QueryRepository,
	fileCommandRepo file.CommandRepository,
	versionQueryRepo file_version.QueryRepository,
	versionCommandRepo file_version.CommandRepository,
	eventService *EventService,
) *FileService {
	return &FileService{
		fileQueryRepo:      fileQueryRepo,
		fileCommandRepo:    fileCommandRepo,
		versionQueryRepo:   versionQueryRepo,
		versionCommandRepo: versionCommandRepo,
		eventService:       eventService,
	}
}

// --- Queries ---

func (s *FileService) GetByID(fileID uuid.UUID) (*file.File, error) {
	f, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, file.ErrNotFound
	}
	return f, nil
}

func (s *FileService) GetAllByUser(userID uuid.UUID) ([]*file.File, error) {
	return s.fileQueryRepo.GetByUserID(userID)
}

func (s *FileService) GetAll() ([]*file.File, error) {
	return s.fileQueryRepo.GetAll()
}

// --- Commands ---

func (s *FileService) RenameFile(fileID uuid.UUID, newName string) error {
	f, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if f == nil {
		return file.ErrNotFound
	}

	// создаём VO для имени файла
	nameVO, err := file.NewFileName(newName)
	if err != nil {
		return err
	}

	f.Rename(nameVO)

	if err := s.fileCommandRepo.Save(f); err != nil {
		return err
	}

	// событие переименования
	if s.eventService != nil {
		eventName, payload := file.NewFileRenamedEvent(f)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}

func (s *FileService) Delete(fileID uuid.UUID) error {
	f, err := s.fileQueryRepo.GetByID(fileID)
	if err != nil || f == nil {
		return file.ErrNotFound
	}

	versions, err := s.versionQueryRepo.GetByFileID(fileID)
	if err != nil {
		return err
	}

	for _, v := range versions {
		if v.Status.Equal(file_version.FileStatusProcessing) {
			return file_version.ErrVersionProcessing
		}
	}

	for _, v := range versions {
		if err := s.versionCommandRepo.Delete(v.ID); err != nil {
			return err
		}
	}

	if err := s.fileCommandRepo.Delete(fileID); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileDeletedEvent(f)
		_, _ = s.eventService.Create(eventName, payload)
	}

	return nil
}
