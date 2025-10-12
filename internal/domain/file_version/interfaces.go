package file_version

import uuid "github.com/google/uuid"

type QueryRepository interface {
	GetByID(id uuid.UUID) (*FileVersion, error)
	GetByFileID(fileID uuid.UUID) ([]*FileVersion, error)
	GetAll() ([]*FileVersion, error)
}

type CommandRepository interface {
	Save(version *FileVersion) error
	Delete(id uuid.UUID) error
}
