package file_version_service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/app"
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
	previewProducer    queue.PreviewProducer
	eventService       *event_service.EventService
	uow                app.UnitOfWork
}

func NewFileVersionService(
	fileQueryRepo file.QueryRepository,
	fileCommandRepo file.CommandRepository,
	versionQueryRepo file_version.QueryRepository,
	versionCommandRepo file_version.CommandRepository,
	storage storage.Storage,
	previewConsumer queue.PreviewConsumer,
	previewProducer queue.PreviewProducer,
	eventService *event_service.EventService,
	uow app.UnitOfWork,
) *FileVersionService {
	return &FileVersionService{
		fileQueryRepo:      fileQueryRepo,
		fileCommandRepo:    fileCommandRepo,
		versionQueryRepo:   versionQueryRepo,
		versionCommandRepo: versionCommandRepo,
		storage:            storage,
		previewConsumer:    previewConsumer,
		eventService:       eventService,
		previewProducer:    previewProducer,
		uow:                uow,
	}
}

func generateS3Key(ownerID, fileID uuid.UUID, versionNum int, name string) string {
	return fmt.Sprintf("files/%s/%s/v%d/%s", ownerID, fileID, versionNum, name)
}

func (s *FileVersionService) UpdatePreview(ctx context.Context, versionID uuid.UUID, previewKey file_version.S3Key) error {
	return s.uow.Do(ctx, func(ctx context.Context) error {
		version, err := s.versionQueryRepo.GetByID(ctx, versionID)
		if err != nil {
			return err
		}
		version.MarkReady()
		version.PreviewS3Key = &previewKey
		if err := s.versionCommandRepo.Save(ctx, version); err != nil {
			return err
		}

		file, err := s.fileQueryRepo.GetByID(ctx, version.FileId)
		if err != nil {
			return err
		}

		if !file.VersionNum.Equal(version.VersionNum) {
			return nil
		}
		file.PreviewS3Key = &previewKey
		file.MarkReady()
		return s.fileCommandRepo.Save(ctx, file)
	})
}

func (s *FileVersionService) GetVersionsByStatus(ctx context.Context, status file_version.FileStatus) ([]*file_version.FileVersion, error) {
	return s.versionQueryRepo.GetAllByStatus(ctx, status)
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

// GetDownloadURL generates presigned download URL for a file version
func (s *FileVersionService) GetDownloadURL(ctx context.Context, versionID uuid.UUID, duration time.Duration) (*string, error) {
	version, err := s.versionQueryRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, fmt.Errorf("version not found")
	}
	if version.Status.Equal(file_version.FileStatusProcessing) {
		return nil, file_version.ErrVersionProcessing
	}

	if version.Status.Equal(file_version.FileStatusFailed) {
		return nil, file_version.ErrVersionFailed
	}

	url, err := s.storage.GenerateDownloadURL(ctx, version.S3Key.String(), duration)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (s *FileVersionService) UploadNewFile(ctx context.Context, ownerID, sessionID uuid.UUID, name string, size uint64, mime string) (*file.File, *file_version.FileVersion, string, error) {
	var f *file.File
	var version *file_version.FileVersion
	var uploadURL string

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		fileNameVO, err := file.NewFileName(name)
		if err != nil {
			return err
		}
		fileSizeVO, err := file_version.NewFileSize(size)
		if err != nil {
			return err
		}
		mimeVO, err := file_version.NewMimeType(mime)
		if err != nil {
			return err
		}
		versionNumVO, _ := file_version.NewFileVersionNum(1)

		f = file.NewFile(ownerID, fileNameVO, fileSizeVO, mimeVO, versionNumVO, sessionID)

		key := generateS3Key(ownerID, f.ID, versionNumVO.Int(), fileNameVO.String())
		s3Key, err := file_version.NewS3Key(key)
		if err != nil {
			return err
		}

		version = file_version.NewFileVersion(f.ID, sessionID, s3Key, mimeVO, fileSizeVO, versionNumVO)

		if err := s.fileCommandRepo.Save(ctx, f); err != nil {
			return err
		}
		if err := s.versionCommandRepo.Save(ctx, version); err != nil {
			return err
		}

		uploadURL, err = s.storage.GenerateUploadURLWithSize(ctx, s3Key.String(), uploadURLTTL, int64(version.Size.Uint64()))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, nil, "", err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionUploadedEvent(f.ID, version.ID, ownerID, f.Name.String(), version.VersionNum.Int())
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return f, version, uploadURL, nil
}

func (s *FileVersionService) CompleteUpload(ctx context.Context, versionID uuid.UUID) error {
	version, err := s.versionQueryRepo.GetByID(ctx, versionID)
	if err != nil {
		return err
	}
	file, err := s.fileQueryRepo.GetByID(ctx, version.FileId)
	if err != nil {
		return err
	}
	version.MarkUploaded()
	file.MarkUploaded()

	err = s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.versionCommandRepo.Save(ctx, version); err != nil {
			return err
		}

		if err := s.fileCommandRepo.Save(ctx, file); err != nil {
			return err
		}

		return nil

	})

	if err != nil {
		return err
	}

	err = s.previewProducer.Produce(ctx, versionID)

	return err
}
func (s *FileVersionService) UploadNewVersion(ctx context.Context, fileID, ownerID, sessionID uuid.UUID, name string, size uint64, mime string, versionNum int) (*file.File, *file_version.FileVersion, string, error) {
	var f *file.File
	var version *file_version.FileVersion
	var uploadURL string

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		f, err = s.fileQueryRepo.GetByID(ctx, fileID)
		if err != nil {
			return err
		}
		if f == nil {
			return file.ErrNotFound
		}

		fileNameVO, _ := file.NewFileName(name)
		fileSizeVO, _ := file_version.NewFileSize(size)
		mimeVO, _ := file_version.NewMimeType(mime)
		versionNumVO, _ := file_version.NewFileVersionNum(versionNum)

		s3Key := generateS3Key(ownerID, f.ID, versionNumVO.Int(), fileNameVO.String())
		s3, err := file_version.NewS3Key(s3Key)
		if err != nil {
			return err
		}

		version = file_version.NewFileVersion(f.ID, sessionID, s3, mimeVO, fileSizeVO, versionNumVO)

		f.UpdateFromVersion(version)

		if err := s.fileCommandRepo.Save(ctx, f); err != nil {
			return err
		}
		if err := s.versionCommandRepo.Save(ctx, version); err != nil {
			return err
		}

		uploadURL, err = s.storage.GenerateUploadURL(ctx, s3Key, uploadURLTTL)
		if err != nil {
			return err
		}

		return nil
	})

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
	var f *file.File
	var version *file_version.FileVersion

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		f, err = s.fileQueryRepo.GetByID(ctx, fileID)
		if err != nil {
			return err
		}
		if f == nil {
			return file.ErrNotFound
		}

		version, err = s.versionQueryRepo.GetByID(ctx, versionID)
		if err != nil {
			return err
		}
		if version == nil {
			return file_version.ErrVersionNotFound
		}

		f.UpdateFromVersion(version)
		return s.fileCommandRepo.Save(ctx, f)
	})

	if err != nil {
		return err
	}

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionRestoredEvent(f.ID, version.ID)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}

func (s *FileVersionService) DeleteVersion(ctx context.Context, fileID, versionID uuid.UUID) error {
	var version *file_version.FileVersion

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		f, err := s.fileQueryRepo.GetByID(ctx, fileID)
		if err != nil || f == nil {
			return file.ErrNotFound
		}

		version, err = s.versionQueryRepo.GetByID(ctx, versionID)
		if err != nil || version == nil {
			return file_version.ErrVersionNotFound
		}

		if version.VersionNum.Equal(f.VersionNum) {
			return file_version.ErrCannotDeleteCurr
		}
		if version.Status.Equal(file_version.FileStatusProcessing) {
			return file_version.ErrVersionProcessing
		}

		return s.versionCommandRepo.Delete(ctx, versionID)
	})

	if err != nil {
		return err
	}

	if version.PreviewS3Key != nil {
		_ = s.previewConsumer.Remove(ctx, version.ID)
		_ = s.storage.Delete(ctx, version.PreviewS3Key.String())
	}
	_ = s.storage.Delete(ctx, version.S3Key.String())

	if s.eventService != nil {
		eventName, payload := file.NewFileVersionDeletedEvent(fileID, version.ID)
		_, _ = s.eventService.Create(ctx, eventName, payload)
	}

	return nil
}
