package magic_link

import (
	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type QueryRepository interface {
	GetByID(id uuid.UUID) (*MagicLink, error)
	GetByTokenHash(token value_objects.TokenHash) (*MagicLink, error)
	GetByUserID(userID uuid.UUID) ([]*MagicLink, error)
	GetAll() ([]*MagicLink, error)
}

type CommandRepository interface {
	Save(link *MagicLink) error
	Delete(id uuid.UUID) error
}
