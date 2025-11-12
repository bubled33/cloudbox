package session

import (
	"context"

	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type QueryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	GetAll(ctx context.Context) ([]*Session, error)
	GetByAccessToken(ctx context.Context, tokenHash *value_objects.TokenHash) (*Session, error)
	GetByRefreshToken(ctx context.Context, tokenHash *value_objects.TokenHash) (*Session, error) // Добавлено

}

type CommandRepository interface {
	Save(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id uuid.UUID) error
}
