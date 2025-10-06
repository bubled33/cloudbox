package domain

import "github.com/google/uuid"

type PreviewQueue interface {
	Enqueue(versuinID uuid.UUID) error
}
