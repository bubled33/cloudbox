package queue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ExpirerProducer interface {
	Produce(ctx context.Context, linkID uuid.UUID, duration time.Duration) error
}

type ExpirerConsumer interface {
	Consume(ctx context.Context) (uuid.UUID, time.Duration, error)
	Remove(ctx context.Context, linkID uuid.UUID) error
}
