package domain

import (
	"time"

	uuid "github.com/google/uuid"
)

type User struct {
	ID uuid.UUID

	Email string

	IsEmailVerified bool

	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewUser(email string, displayName string) *User {
	now := time.Now()
	return &User{
		ID:              uuid.New(),
		Email:           email,
		DisplayName:     displayName,
		IsEmailVerified: false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (u *User) VerifyEmail() {
	u.IsEmailVerified = true
	u.UpdatedAt = time.Now()
}
