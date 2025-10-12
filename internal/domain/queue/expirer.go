package events

import (
	"time"

	"github.com/google/uuid"
)

// Очередь, отвечающая за автоматическое истечение публичных ссылок
type Expirer interface {
	// Добавляет задачу истечения ссылки через указанный duration
	Enqueue(linkID uuid.UUID, duration time.Duration) error

	// (опционально) Удаляет задачу, если ссылка была вручную удалена до истечения
	Remove(linkID uuid.UUID) error
}
