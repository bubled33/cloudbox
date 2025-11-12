package file_version

import (
	"context"

	uuid "github.com/google/uuid"
)

type QueryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*FileVersion, error)
	GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*FileVersion, error)
	GetAll(ctx context.Context) ([]*FileVersion, error)
	GetAllByStatus(ctx context.Context, status FileStatus) ([]*FileVersion, error)
}

type CommandRepository interface {
	Save(ctx context.Context, version *FileVersion) error
	Delete(ctx context.Context, id uuid.UUID) error
}
