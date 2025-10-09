package domain

import (
	"time"

	uuid "github.com/google/uuid"
)

type FileRepository interface {
	GetByID(id uuid.UUID) (*File, error)
	GetByUserID(userID uuid.UUID) ([]*File, error)

	GetAll() ([]*File, error)

	Save(file *File) error

	Delete(id uuid.UUID) error
}

type FileVersionRepository interface {
	GetByID(id uuid.UUID) (*FileVersion, error)
	GetByFileID(fileID uuid.UUID) ([]*FileVersion, error)

	GetAll() ([]*FileVersion, error)

	Save(version *FileVersion) error

	Delete(id uuid.UUID) error
}

type UserRepository interface {
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)

	GetAll() ([]*User, error)
	Delete(id uuid.UUID) error

	Save(user *User) error
}

type SessionRepository interface {
	GetByID(id uuid.UUID) (*Session, error)
	GetByUserID(userID uuid.UUID) ([]*Session, error)

	GetAll() ([]*Session, error)

	Save(session *Session) error

	Delete(id uuid.UUID) error
}

type PublicLinkRepository interface {
	GetByID(id uuid.UUID) (*PublicLink, error)
	GetByFileID(fileID uuid.UUID) ([]*PublicLink, error)

	GetAll() ([]*PublicLink, error)

	Save(link *PublicLink) error

	Delete(id uuid.UUID) error

	CleanupExpired(before time.Time) error
}

type MagicLinkRepository interface {
	GetByID(id uuid.UUID) (*MagicLink, error)
	GetByTokenHash(token string) (*MagicLink, error)
	GetByUserID(userID uuid.UUID) ([]*MagicLink, error)

	GetAll() ([]*MagicLink, error)

	Save(link *MagicLink) error

	Delete(id uuid.UUID) error
}
