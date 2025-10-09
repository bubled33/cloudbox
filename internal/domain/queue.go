package domain

import (
	"time"

	"github.com/google/uuid"
)

// Очередь на генерацию превью файлов
type PreviewQueue interface {
	// Добавляет задачу генерации превью для конкретной версии файла
	Enqueue(versionID uuid.UUID) error

	// Удаляет задачу генерации превью для конкретной версии файла
	Remove(versionID uuid.UUID) error
}

// Очередь, отвечающая за автоматическое истечение публичных ссылок
type Expirer interface {
	// Добавляет задачу истечения ссылки через указанный duration
	Enqueue(linkID uuid.UUID, duration time.Duration) error

	// (опционально) Удаляет задачу, если ссылка была вручную удалена до истечения
	Remove(linkID uuid.UUID) error
}
