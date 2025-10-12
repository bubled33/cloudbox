package value_objects

import (
	"time"

	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
)

type ExpiresAt struct {
	value time.Time
}

func NewExpiresAt(t time.Time) (ExpiresAt, error) {
	if time.Until(t) <= 0 {
		return ExpiresAt{}, domainerrors.ErrInvalidExpiry
	}
	return ExpiresAt{value: t}, nil
}

func (e ExpiresAt) Time() time.Time {
	return e.value
}

func (e ExpiresAt) IsExpired() bool {
	return time.Now().After(e.value)
}
