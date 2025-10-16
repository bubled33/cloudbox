package db

import (
	"context"
	"database/sql"

	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

type FileVersionCommandRepository struct {
}

func NewFileVersionCommandRepository() *FileVersionCommandRepository {
	return &FileVersionCommandRepository{}
}

type FileVersionQueryRepository struct {
	db *sql.DB
}

func NewFileVersionQueryRepository(db *sql.DB) *FileVersionQueryRepository {
	return &FileVersionQueryRepository{db: db}
}

func scanFileVersion(scanner scannable) (*file_version.FileVersion, error) {
	var v file_version.FileVersion
	var s3Key, previewS3KeyStr, mime, status string
	var size uint64
	var versionNum int

	if err := scanner.Scan(
		&v.ID,
		s3Key,
		previewS3KeyStr,
		mime,
		status,
		size,
		versionNum,
		&v.FileId,
		&v.UploadedBySessionId,
		&v.CreatedAt,
		&v.UpdatedAt,
	); err != nil {
		return nil, err
	}

	var err error

	v.S3Key, err = file_version.NewS3Key(s3Key)
	if err != nil {
		return nil, err
	}

	previewS3Key, err := file_version.NewS3Key(previewS3KeyStr)
	if err != nil {
		return nil, err
	}
	v.PreviewS3Key = &previewS3Key

	v.Mime, err = file_version.NewMimeType(mime)
	if err != nil {
		return nil, err
	}

	v.Status, err = file_version.NewFileStatus(status)
	if err != nil {
		return nil, err
	}

	v.Size, err = file_version.NewFileSize(size)
	if err != nil {
		return nil, err
	}

	v.VersionNum, err = file_version.NewFileVersionNum(versionNum)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (r *FileVersionQueryRepository) GetByID(ctx context.Context, id uuid.UUID) (*file_version.FileVersion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, s3_key, preview_s3_key, mime, status, size, version_num,
		       file_id, uploaded_by_session_id, created_at, updated_at
		FROM file_versions
		WHERE id = $1
	`, id)

	v, err := scanFileVersion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return v, err
}

func (r *FileVersionQueryRepository) GetByFileID(ctx context.Context, fileID uuid.UUID) (*file_version.FileVersion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, s3_key, preview_s3_key, mime, status, size, version_num,
		       file_id, uploaded_by_session_id, created_at, updated_at
		FROM file_versions
		WHERE file_id = $1
	`, fileID)

	v, err := scanFileVersion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return v, err
}

func (r *FileVersionQueryRepository) GetAll(ctx context.Context) ([]*file_version.FileVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, s3_key, preview_s3_key, mime, status, size, version_num,
		       file_id, uploaded_by_session_id, created_at, updated_at
		FROM file_versions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fileVersions []*file_version.FileVersion
	for rows.Next() {
		s, err := scanFileVersion(rows)
		if err != nil {
			return nil, err
		}
		fileVersions = append(fileVersions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fileVersions, nil
}
