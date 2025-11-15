package event_service

import (
	"context"

	"github.com/yourusername/cloud-file-storage/internal/domain/event"
)

type EventPublisher interface {
	Create(ctx context.Context, name string, payload any) (*event.Event, error)
	PublishPending(ctx context.Context, batchSize int, maxRetries int) error
}

type EventManager interface {
	EventPublisher

	GetPendingCount(ctx context.Context) (int, error)
	GetFailedEvents(ctx context.Context, limit int) ([]*event.Event, error)
	RetryFailed(ctx context.Context, eventID string) error
}
