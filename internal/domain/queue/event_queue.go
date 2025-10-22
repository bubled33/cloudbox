package queue

import (
	"context"

	"github.com/yourusername/cloud-file-storage/internal/domain/event"
)

type EventProducer interface {
	Produce(ctx context.Context, event *event.Event) error
}

type EventConsumer interface {
	Consume(ctx context.Context) (*event.Event, error)
}
