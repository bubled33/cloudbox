package event

import (
	"context"

	"github.com/google/uuid"
)

// CommandRepository отвечает за изменение данных (Write)
type CommandRepository interface {
	Save(ctx context.Context, event *Event) error
	Delete(ctx context.Context, id uuid.UUID) error
	MarkAsSent(ctx context.Context, id uuid.UUID) error
	UpdateRetryCount(ctx context.Context, id uuid.UUID, retryCount int) error
}

// QueryRepository отвечает за чтение данных (Read)
type QueryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
	GetPending(ctx context.Context, limit int) ([]*Event, error)
	GetAll(ctx context.Context) ([]*Event, error)
}
