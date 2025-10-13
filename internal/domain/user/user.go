package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID

	Email           Email
	DisplayName     DisplayName
	IsEmailVerified bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(email Email, displayName DisplayName) *User {
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

func (u *User) Rename(displayName DisplayName) {
	u.DisplayName = displayName
	u.UpdatedAt = time.Now()
}
