package queue

import "github.com/google/uuid"

type PreviewQueue interface {
	Enqueue(versionID uuid.UUID) error
	Remove(versionID uuid.UUID) error
}
