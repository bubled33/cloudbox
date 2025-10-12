package events

import "github.com/google/uuid"

// Очередь на генерацию превью файлов
type PreviewQueue interface {
	// Добавляет задачу генерации превью для конкретной версии файла
	Enqueue(versionID uuid.UUID) error

	// Удаляет задачу генерации превью для конкретной версии файла
	Remove(versionID uuid.UUID) error
}
