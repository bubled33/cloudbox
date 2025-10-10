package domain

import (
	"time"

	uuid "github.com/google/uuid"
)

// --- FILE ---

type FileQueryRepository interface {
	GetByID(id uuid.UUID) (*File, error)
	GetByUserID(userID uuid.UUID) ([]*File, error)
	GetAll() ([]*File, error)
}

type FileCommandRepository interface {
	Save(file *File) error
	Delete(id uuid.UUID) error
}

type FileVersionQueryRepository interface {
	GetByID(id uuid.UUID) (*FileVersion, error)
	GetByFileID(fileID uuid.UUID) ([]*FileVersion, error)
	GetAll() ([]*FileVersion, error)
}

type FileVersionCommandRepository interface {
	Save(version *FileVersion) error
	Delete(id uuid.UUID) error
}

// --- USER ---

type UserQueryRepository interface {
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]*User, error)
}

type UserCommandRepository interface {
	Save(user *User) error
	Delete(id uuid.UUID) error
}

// --- SESSION ---

type SessionQueryRepository interface {
	GetByID(id uuid.UUID) (*Session, error)
	GetByUserID(userID uuid.UUID) ([]*Session, error)
	GetAll() ([]*Session, error)
}

type SessionCommandRepository interface {
	Save(session *Session) error
	Delete(id uuid.UUID) error
}

// --- PUBLIC LINK ---

type PublicLinkQueryRepository interface {
	GetByID(id uuid.UUID) (*PublicLink, error)
	GetByFileID(fileID uuid.UUID) ([]*PublicLink, error)
	GetAll() ([]*PublicLink, error)
}

type PublicLinkCommandRepository interface {
	Save(link *PublicLink) error
	Delete(id uuid.UUID) error
	CleanupExpired(before time.Time) error
}

// --- MAGIC LINK ---

type MagicLinkQueryRepository interface {
	GetByID(id uuid.UUID) (*MagicLink, error)
	GetByTokenHash(token string) (*MagicLink, error)
	GetByUserID(userID uuid.UUID) ([]*MagicLink, error)
	GetAll() ([]*MagicLink, error)
}

type MagicLinkCommandRepository interface {
	Save(link *MagicLink) error
	Delete(id uuid.UUID) error
}
