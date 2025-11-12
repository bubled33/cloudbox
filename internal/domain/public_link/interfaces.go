package public_link

import (
	"context"

	"github.com/google/uuid"
)

type PublicLinkQueryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*PublicLink, error)
	GetByToken(ctx context.Context, token string) (*PublicLink, error)
	GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*PublicLink, error)
	GetAll(ctx context.Context) ([]*PublicLink, error)
}

type PublicLinkCommandRepository interface {
	Save(ctx context.Context, link *PublicLink) error
	Delete(ctx context.Context, id uuid.UUID) error
}
