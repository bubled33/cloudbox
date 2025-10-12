package event

type EventRepository interface {
	Save(event *Event) error

	GetPending(limit int) ([]*Event, error)
}
