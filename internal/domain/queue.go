package domain

import "github.com/google/uuid"

type PreviewQueue interface {
	Enqueue(versuinID uuid.UUID) error
}

type PublicExpirer interface {
	Enqueue(linkID uuid.UUID) error
}
