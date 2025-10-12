package file

import uuid "github.com/google/uuid"

type QueryRepository interface {
	GetByID(id uuid.UUID) (*File, error)
	GetByUserID(userID uuid.UUID) ([]*File, error)
	GetAll() ([]*File, error)
}

type CommandRepository interface {
	Save(file *File) error
	Delete(id uuid.UUID) error
}
