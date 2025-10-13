package app

import (
	"fmt"

	"github.com/yourusername/cloud-file-storage/internal/domain/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
)

type EventService struct {
	repo       event.EventRepository
	queue      queue.EventQueue
	instanceID string
}

func NewEventService(repo event.EventRepository, queue queue.EventQueue, instanceID string) *EventService {
	return &EventService{
		repo:       repo,
		queue:      queue,
		instanceID: instanceID,
	}
}
func (s *EventService) Create(name string, payload any) (*event.Event, error) {
	e, err := event.NewEvent(name, payload)
	if err != nil {
		return nil, err
	}

	e.Lock(s.instanceID)
	defer e.Unlock()

	if err := s.repo.Save(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (s *EventService) PublishPending(batchSize int, maxRetries int) error {
	events, err := s.repo.GetPending(batchSize)
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
			if err := s.repo.Save(e); err != nil {
				failedEvents = append(failedEvents, e)
				continue
			}
		}

		if err := s.queue.Enqueue(e); err != nil {

			e.RetryCount++
			e.Unlock()
			_ = s.repo.Save(e)

			if e.RetryCount >= maxRetries {
				// TODO: Здесь можно пометить событие как "проваленное" или отправить в dead-letter
			}

			failedEvents = append(failedEvents, e)
			continue
		}

		e.MarkAsSent()
		e.Unlock()
		_ = s.repo.Save(e)
	}

	if len(failedEvents) > 0 {
		return fmt.Errorf("%d/%d событий не удалось отправить", len(failedEvents), len(events))
	}

	return nil
}
