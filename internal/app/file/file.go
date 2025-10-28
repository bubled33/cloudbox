package app

import (
	"context"

	"github.com/google/uuid"
	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

type FileService struct {
	fileQueryRepo      file.QueryRepository
	fileCommandRepo    file.CommandRepository
	versionQueryRepo   file_version.QueryRepository
	versionCommandRepo file_version.CommandRepository
	eventService       *event_service.EventService
}

func NewFileService(
	fileQueryRepo file.QueryRepository,
	fileCommandRepo file.CommandRepository,
	versionQueryRepo file_version.QueryRepository,
	versionCommandRepo file_version.CommandRepository,
	eventService *event_service.EventService,
) *FileService {
	return &FileService{
		fileQueryRepo:      fileQueryRepo,
		fileCommandRepo:    fileCommandRepo,
		versionQueryRepo:   versionQueryRepo,
		versionCommandRepo: versionCommandRepo,
		eventService:       eventService,
	}
}

func (s *FileService) GetByID(ctx context.Context, fileID uuid.UUID) (*file.File, error) {
	f, err := s.fileQueryRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, file.ErrNotFound
	}
	return f, nil
}

func (s *FileService) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]*file.File, error) {
	return s.fileQueryRepo.GetByUserID(ctx, userID)
}

func (s *FileService) GetAll(ctx context.Context) ([]*file.File, error) {
	return s.fileQueryRepo.GetAll(ctx)
}

func (s *FileService) RenameFile(ctx context.Context, fileID uuid.UUID, newName string) error {
	f, err := s.fileQueryRepo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}
	if f == nil {
		return file.ErrNotFound
	}

	nameVO, err := file.NewFileName(newName)
	if err != nil {
		return err
	}

	f.Rename(nameVO)

	if err := s.fileCommandRepo.Save(ctx, f); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileRenamedEvent(f)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}

func (s *FileService) Delete(ctx context.Context, fileID uuid.UUID) error {
	f, err := s.fileQueryRepo.GetByID(ctx, fileID)
	if err != nil || f == nil {
		return file.ErrNotFound
	}

	versions, err := s.versionQueryRepo.GetByFileID(ctx, fileID)
	if err != nil {
		return err
	}

	for _, v := range versions {
		if v.Status.Equal(file_version.FileStatusProcessing) {
			return file_version.ErrVersionProcessing
		}
	}

	for _, v := range versions {
		if err := s.versionCommandRepo.Delete(ctx, v.ID); err != nil {
			return err
		}
	}

	if err := s.fileCommandRepo.Delete(ctx, fileID); err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileDeletedEvent(f)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}
