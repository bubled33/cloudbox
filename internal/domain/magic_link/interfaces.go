package magic_link

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type QueryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*MagicLink, error)
	GetByTokenHash(ctx context.Context, token value_objects.TokenHash) (*MagicLink, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*MagicLink, error)
	GetAll(ctx context.Context) ([]*MagicLink, error)
}

type CommandRepository interface {
	Save(ctx context.Context, link *MagicLink) error
	Delete(ctx context.Context, id uuid.UUID) error
}
