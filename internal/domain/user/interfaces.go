package user

import (
	"context"

	uuid "github.com/google/uuid"
)

type QueryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
}

type CommandRepository interface {
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
