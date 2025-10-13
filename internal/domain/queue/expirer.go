package queue

import (
	"time"

	"github.com/google/uuid"
)

type Expirer interface {
	Enqueue(linkID uuid.UUID, duration time.Duration) error

	Remove(linkID uuid.UUID) error
}
