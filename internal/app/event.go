package app

import (
	"fmt"

	"github.com/yourusername/cloud-file-storage/internal/domain/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
)

// EventService управляет доменными событиями и их отправкой в очередь
type EventService struct {
	repo       event.EventRepository
	queue      queue.EventQueue
	instanceID string // уникальный идентификатор инстанса для блокировки
}

// NewEventService создаёт сервис для работы с событиями
func NewEventService(repo event.EventRepository, queue queue.EventQueue, instanceID string) *EventService {
	return &EventService{
		repo:       repo,
		queue:      queue,
		instanceID: instanceID,
	}
}

// Create создаёт новое событие и сохраняет его в репозитории
func (s *EventService) Create(name string, payload any) (*event.Event, error) {
	e, err := event.NewEvent(name, payload)
	if err != nil {
		return nil, err
	}

	// Блокируем событие на момент сохранения
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
		// Пробуем заблокировать событие
		if e.LockedAt == nil {
			e.Lock(s.instanceID)
			if err := s.repo.Save(e); err != nil {
				failedEvents = append(failedEvents, e)
				continue
			}
		}

		// Пробуем отправить событие в очередь
		if err := s.queue.Enqueue(e); err != nil {
			// Увеличиваем RetryCount и снимаем блокировку
			e.RetryCount++
			e.Unlock()
			_ = s.repo.Save(e)

			// Если достигнут maxRetries, можно логировать или сигнализировать
			if e.RetryCount >= maxRetries {
				// Здесь можно пометить событие как "проваленное" или отправить в dead-letter
			}

			failedEvents = append(failedEvents, e)
			continue
		}

		// Событие успешно отправлено
		e.MarkAsSent()
		e.Unlock()
		_ = s.repo.Save(e)
	}

	if len(failedEvents) > 0 {
		// Возвращаем ошибку, чтобы можно было логировать частичную неудачу
		return fmt.Errorf("%d/%d событий не удалось отправить", len(failedEvents), len(events))
	}

	return nil
}
