package queue

import "github.com/yourusername/cloud-file-storage/internal/domain/event"

type EventQueue interface {
	Enqueue(event *event.Event) error
}
