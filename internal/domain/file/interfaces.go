package file

import (
	"context"

	uuid "github.com/google/uuid"
)

type QueryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*File, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*File, error)
	GetAll(ctx context.Context) ([]*File, error)
}

type CommandRepository interface {
	Save(ctx context.Context, file *File) error
	Delete(ctx context.Context, id uuid.UUID) error
}
