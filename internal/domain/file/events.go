package file

import (
	"github.com/google/uuid"
)

func NewFileRenamedEvent(f *File) (string, map[string]interface{}) {
	return "FileRenamed", map[string]interface{}{
		"file_id": f.ID,
		"name":    f.Name.String(),
	}
}

func NewFileDeletedEvent(f *File) (string, map[string]interface{}) {
	return "FileDeleted", map[string]interface{}{
		"file_id":  f.ID,
		"owner_id": f.OwnerID,
	}
}

func NewFileCreatedEvent(f *File) (string, map[string]interface{}) {
	return "FileCreated", map[string]interface{}{
		"file_id":  f.ID,
		"owner_id": f.OwnerID,
		"name":     f.Name.String(),
		"size":     f.Size,
		"mime":     f.Mime.String(),
	}
}

func NewFileVersionUploadedEvent(fileID, versionID uuid.UUID, ownerID uuid.UUID, name string, versionNum int) (string, map[string]interface{}) {
	return "FileVersionUploaded", map[string]interface{}{
		"file_id":    fileID,
		"version_id": versionID,
		"owner_id":   ownerID,
		"name":       name,
		"version":    versionNum,
	}
}

func NewFileVersionDeletedEvent(fileID, versionID uuid.UUID) (string, map[string]interface{}) {
	return "FileVersionDeleted", map[string]interface{}{
		"file_id":    fileID,
		"version_id": versionID,
	}
}

func NewFileVersionRestoredEvent(fileID, versionID uuid.UUID) (string, map[string]interface{}) {
	return "FileVersionRestored", map[string]interface{}{
		"file_id":    fileID,
		"version_id": versionID,
	}
}
