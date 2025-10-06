package domain

import (
	"time"

	uuid "github.com/google/uuid"
)

type FileRepository interface {
	GetByID(id uuid.UUID) (*File, error)
	Save(file *File) error
	Delete(id uuid.UUID) error
}

type FileVersionRepository interface {
	GetByID(id uuid.UUID) (*FileVersion, error)
	GetByFileID(fileID uuid.UUID) ([]*FileVersion, error)
	Save(version *FileVersion) error
	Delete(id uuid.UUID) error
}

type UserRepository interface {
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	Save(user *User) error
}

type SessionRepository interface {
	GetByID(id uuid.UUID) (*Session, error)
	GetByUserID(userID uuid.UUID) ([]*Session, error)
	Save(session *Session) error
	Revoke(sessionID uuid.UUID) error
	CleanupExpired(before time.Time) error
}

type PublicLinkRepository interface {
	GetByID(id uuid.UUID) (*PublicLink, error)
	GetByFileID(fileID uuid.UUID) ([]*PublicLink, error)
	Save(link *PublicLink) error
	Delete(id uuid.UUID) error
	CleanupExpired(before time.Time) error
}

type MagicLinkRepository interface {
	GetByID(id uuid.UUID) (*MagicLink, error)
	GetByUserID(userID uuid.UUID) ([]*MagicLink, error)
	Save(link *MagicLink) error
	MarkAsUsed(id uuid.UUID) error
	CleanupExpired(before time.Time) error
}
