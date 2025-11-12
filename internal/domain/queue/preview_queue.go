package queue

import (
	"context"

	"github.com/google/uuid"
)

type PreviewProducer interface {
	Produce(ctx context.Context, versionID uuid.UUID) error
}

type PreviewConsumer interface {
	Consume(ctx context.Context) (uuid.UUID, error)
	Remove(ctx context.Context, versionID uuid.UUID) error
}
