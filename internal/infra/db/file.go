package db

import (
	"context"
	"database/sql"

	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

type FileCommandRepository struct {
}

func (r *FileCommandRepository) Save(ctx context.Context, v *file.File) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `
	INSERT INTO files (id, name, preview_s3_key, mime, status, size, version_num,
		       owner_id, uploaded_by_session_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT (id) DO UPDATE 
	SET s3_key = $2, preview_s3_key = $3, mime = $4, status = $5, size = $6, version_num = $7,
		       file_id = $8, uploaded_by_session_id = $9, created_at = $10, updated_at = $11
	`
	_, err := tx.ExecContext(ctx, query,
		v.ID,
		v.Name.String(),
		v.PreviewS3Key.String(),
		v.Mime.String(),
		v.Status.String(),
		v.Size.Uint64(),
		v.VersionNum.Int(),
		v.OwnerID,
		v.UploadedBySessionId,
		v.CreatedAt,
		v.UpdatedAt,
	)
	return err
}

func (r *FileCommandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `DELETE FROM files WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
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
		&name,
		&previewS3KeyStr,
		&mime,
		&status,
		&size,
		&versionNum,
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
