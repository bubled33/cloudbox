package event

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	Sent      bool      `json:"sent"`

	LockedAt   *time.Time `json:"locked_at,omitempty"`
	LockedBy   *string    `json:"locked_by,omitempty"`
	RetryCount int        `json:"retry_count"`
}

func NewEvent(name string, payload any) (*Event, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Event{
		ID:        uuid.New(),
		Name:      name,
		Data:      string(data),
		CreatedAt: time.Now(),
		Sent:      false,
	}, nil
}

func (e *Event) DecodePayload(target any) error {
	return json.Unmarshal([]byte(e.Data), target)
}

func (e *Event) Lock(instanceID string) {
	now := time.Now()
	e.LockedAt = &now
	e.LockedBy = &instanceID
}

func (e *Event) Unlock() {
	e.LockedAt = nil
	e.LockedBy = nil
}

func (e *Event) MarkAsSent() {
	e.Sent = true
}
