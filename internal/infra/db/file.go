package db

import (
	"context"
	"database/sql"

	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

type FileCommandRepository struct {
}

func NewFileCommandRepository() *FileCommandRepository {
	return &FileCommandRepository{}
}

type FileQueryRepository struct {
	db *sql.DB
}

func NewFileQueryRepository(db *sql.DB) *FileQueryRepository {
	return &FileQueryRepository{db: db}
}

func scanFile(scanner scannable) (*file.File, error) {
	var f file.File
	var name, previewS3KeyStr, mime, status string
	var size uint64
	var versionNum int

	if err := scanner.Scan(
		&f.ID,
		name,
		previewS3KeyStr,
		mime,
		status,
		size,
		versionNum,
		&f.OwnerID,
		&f.UploadedBySessionId,
		&f.CreatedAt,
		&f.UpdatedAt,
	); err != nil {
		return nil, err
	}

	var err error

	f.Name, err = file.NewFileName(name)
	if err != nil {
		return nil, err
	}

	previewS3Key, err := file_version.NewS3Key(previewS3KeyStr)
	if err != nil {
		return nil, err
	}
	f.PreviewS3Key = &previewS3Key

	f.Mime, err = file_version.NewMimeType(mime)
	if err != nil {
		return nil, err
	}

	f.Status, err = file_version.NewFileStatus(status)
	if err != nil {
		return nil, err
	}

	f.Size, err = file_version.NewFileSize(size)
	if err != nil {
		return nil, err
	}

	f.VersionNum, err = file_version.NewFileVersionNum(versionNum)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

func (r *FileQueryRepository) GetByID(ctx context.Context, id uuid.UUID) (*file.File, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, preview_s3_key, mime, status, size, version_num,
		       owner_id, uploaded_by_session_id, created_at, updated_at
		FROM files
		WHERE id = $1
	`, id)

	f, err := scanFile(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return f, err
}

func (r *FileQueryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*file.File, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, preview_s3_key, mime, status, size, version_num,
		       owner_id, uploaded_by_session_id, created_at, updated_at
		FROM files
		WHERE user_id = $1
	`, userID)

	f, err := scanFile(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return f, err
}

func (r *FileQueryRepository) GetAll(ctx context.Context) ([]*file.File, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, preview_s3_key, mime, status, size, version_num,
		       owner_id, uploaded_by_session_id, created_at, updated_at
		FROM files
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*file.File
	for rows.Next() {
		f, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
