package workers

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/yourusername/cloud-file-storage/internal/domain/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/queue"
)

var (
	eventsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_processed_total",
			Help: "Total number of processed events",
		},
		[]string{"event_name", "status"},
	)

	eventRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "event_retries_total",
			Help: "Total number of event retries",
		},
		[]string{"event_name"},
	)

	eventProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "event_processing_duration_seconds",
			Help:    "Time taken to process events",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"event_name"},
	)

	eventQueueAge = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "event_queue_age_seconds",
			Help:    "Time since event creation until processing",
			Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
		},
		[]string{"event_name"},
	)

	lockedEventsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "events_locked_current",
			Help: "Current number of locked events",
		},
	)
)

type MetricsWorker struct {
	consumer     queue.EventConsumer
	pollInterval time.Duration
	stopCh       chan struct{}
}

func NewMetricsWorker(consumer queue.EventConsumer, pollInterval time.Duration) *MetricsWorker {
	return &MetricsWorker{
		consumer:     consumer,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
	}
}

func (w *MetricsWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	log.Println("MetricsWorker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("MetricsWorker stopped by context")
			return ctx.Err()
		case <-w.stopCh:
			log.Println("MetricsWorker stopped")
			return nil
		case <-ticker.C:
			if err := w.consumeAndRecordMetrics(ctx); err != nil {
				log.Printf("Error consuming events for metrics: %v", err)
				continue
			}
		}
	}
}

func (w *MetricsWorker) Stop() {
	close(w.stopCh)
}

func (w *MetricsWorker) consumeAndRecordMetrics(ctx context.Context) error {
	event, err := w.consumer.Consume(ctx)
	if err != nil {
		return err
	}

	w.recordEventMetrics(event)

	return nil
}

func (w *MetricsWorker) recordEventMetrics(e *event.Event) {
	start := time.Now()

	status := w.determineEventStatus(e)

	eventsProcessedTotal.WithLabelValues(e.Name, status).Inc()

	if e.RetryCount > 0 {
		eventRetries.WithLabelValues(e.Name).Add(float64(e.RetryCount))
	}

	queueAge := time.Since(e.CreatedAt).Seconds()
	eventQueueAge.WithLabelValues(e.Name).Observe(queueAge)

	if e.LockedAt != nil {
		lockedEventsGauge.Inc()
	}

	duration := time.Since(start).Seconds()
	eventProcessingDuration.WithLabelValues(e.Name).Observe(duration)
}

func (w *MetricsWorker) determineEventStatus(e *event.Event) string {
	if e.Sent {
		return "sent"
	}
	if e.RetryCount > 0 {
		return "retry"
	}
	if e.LockedAt != nil {
		return "locked"
	}
	return "pending"
}
