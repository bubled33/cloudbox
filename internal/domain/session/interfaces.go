package session

import uuid "github.com/google/uuid"

type QueryRepository interface {
	GetByID(id uuid.UUID) (*Session, error)
	GetByUserID(userID uuid.UUID) ([]*Session, error)
	GetAll() ([]*Session, error)
}

type CommandRepository interface {
	Save(session *Session) error
	Delete(id uuid.UUID) error
}
