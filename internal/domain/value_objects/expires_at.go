package value_objects

import (
	"fmt"
	"time"

	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
)

type ExpiresAt struct {
	value time.Time
}

func NewExpiresAt(t time.Time) (ExpiresAt, error) {
	if t.IsZero() {
		return ExpiresAt{}, domainerrors.ErrInvalidExpiry
	}

	return ExpiresAt{value: t}, nil
}

func (e ExpiresAt) Time() time.Time {
	fmt.Println("Test time", e.value)
	return e.value
}

func (e ExpiresAt) IsExpired() bool {
	return time.Now().After(e.value)
}

func (e ExpiresAt) String() string {
	return e.value.Format(time.RFC3339)
}
