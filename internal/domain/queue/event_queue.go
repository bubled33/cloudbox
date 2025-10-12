package queue

import "github.com/yourusername/cloud-file-storage/internal/domain/event"

// EventQueue отвечает за публикацию доменных событий в асинхронную очередь
type EventQueue interface {
	// Добавляет событие в очередь
	Enqueue(event *event.Event) error
}
