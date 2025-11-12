package event_service

import (
	"context"
	"fmt"

	"github.com/yourusername/cloud-file-storage/internal/domain/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
)

type EventService struct {
	queryRepo   event.QueryRepository
	commandRepo event.CommandRepository
	producer    queue.EventProducer
	instanceID  string
}

func NewEventService(
	queryRepo event.QueryRepository,
	commandRepo event.CommandRepository,
	producer queue.EventProducer,
	instanceID string,
) *EventService {
	return &EventService{
		queryRepo:   queryRepo,
		commandRepo: commandRepo,
		producer:    producer,
		instanceID:  instanceID,
	}
}

func (s *EventService) Create(ctx context.Context, name string, payload any) (*event.Event, error) {
	e, err := event.NewEvent(name, payload)
	if err != nil {
		return nil, err
	}

	e.Lock(s.instanceID)
	defer e.Unlock()

	if err := s.commandRepo.Save(ctx, e); err != nil {
		return nil, err
	}

	return e, nil
}

func (s *EventService) PublishPending(ctx context.Context, batchSize int, maxRetries int) error {
	events, err := s.queryRepo.GetPending(ctx, batchSize)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	var failedEvents []*event.Event

	for _, e := range events {
		if e.LockedAt == nil {
			e.Lock(s.instanceID)
			if err := s.commandRepo.Save(ctx, e); err != nil {
				failedEvents = append(failedEvents, e)
				continue
			}
		}

		if err := s.producer.Produce(ctx, e); err != nil {
			e.RetryCount++
			e.Unlock()

			if err := s.commandRepo.UpdateRetryCount(ctx, e.ID, e.RetryCount); err != nil {

			}

			if e.RetryCount >= maxRetries {

			}

			failedEvents = append(failedEvents, e)
			continue
		}

		e.MarkAsSent()
		e.Unlock()

		if err := s.commandRepo.MarkAsSent(ctx, e.ID); err != nil {

		}
	}

	if len(failedEvents) > 0 {
		return fmt.Errorf("%d/%d событий не удалось отправить", len(failedEvents), len(events))
	}

	return nil
}
