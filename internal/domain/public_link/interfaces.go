package public_link

import (
	"time"

	"github.com/google/uuid"
)

type PublicLinkQueryRepository interface {
	GetByID(id uuid.UUID) (*PublicLink, error)
	GetByFileID(fileID uuid.UUID) ([]*PublicLink, error)
	GetAll() ([]*PublicLink, error)
}

type PublicLinkCommandRepository interface {
	Save(link *PublicLink) error
	Delete(id uuid.UUID) error
	CleanupExpired(before time.Time) error
}
