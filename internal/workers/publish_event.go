package workers

import (
	"context"
	"log"
	"time"

	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
)

// PublishEventsWorker запускает периодическую публикацию pending событий
type PublishEventsWorker struct {
	eventService *event_service.EventService
	interval     time.Duration
	batchSize    int
	maxRetries   int
	stopCh       chan struct{}
}

func NewPublishEventsWorker(eventService *event_service.EventService, interval time.Duration, batchSize, maxRetries int) *PublishEventsWorker {
	return &PublishEventsWorker{
		eventService: eventService,
		interval:     interval,
		batchSize:    batchSize,
		maxRetries:   maxRetries,
		stopCh:       make(chan struct{}),
	}
}

// Start запускает воркер в фоне с регулярным интервалом
func (w *PublishEventsWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Println("PublishEventsWorker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("PublishEventsWorker stopped by context")
			return
		case <-w.stopCh:
			log.Println("PublishEventsWorker stopped")
			return
		case <-ticker.C:
			if err := w.publishPending(ctx); err != nil {
				log.Printf("PublishEventsWorker error publishing events: %v", err)
			}
		}
	}
}

// Stop останавливает воркер
func (w *PublishEventsWorker) Stop() {
	close(w.stopCh)
}

// publishPending вызывает EventService для публикации pending событий
func (w *PublishEventsWorker) publishPending(ctx context.Context) error {
	return w.eventService.PublishPending(ctx, w.batchSize, w.maxRetries)
}
