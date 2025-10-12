package user

import uuid "github.com/google/uuid"

// --- USER ---

type QueryRepository interface {
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]*User, error)
}

type CommandRepository interface {
	Save(user *User) error
	Delete(id uuid.UUID) error
}
