package db

import (
	"context"
	"database/sql"
	"fmt"

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

	var previewS3Key sql.NullString
	if v.PreviewS3Key != nil {
		previewS3Key = sql.NullString{
			String: v.PreviewS3Key.String(),
			Valid:  true,
		}
	}

	query := `
    INSERT INTO files (id, name, preview_s3_key, mime, status, size, version_num,
               owner_id, uploaded_by_session_id, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    ON CONFLICT (id) DO UPDATE 
    SET name = EXCLUDED.name, 
        preview_s3_key = EXCLUDED.preview_s3_key, 
        mime = EXCLUDED.mime, 
        status = EXCLUDED.status, 
        size = EXCLUDED.size, 
        version_num = EXCLUDED.version_num,
        uploaded_by_session_id = EXCLUDED.uploaded_by_session_id, 
        updated_at = EXCLUDED.updated_at
    `
	_, err := tx.ExecContext(ctx, query,
		v.ID,
		v.Name.String(),
		previewS3Key,
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

// scanFile сканирует строку базы данных в объект File
func scanFile(scanner scannable) (*file.File, error) {
	var f file.File
	var name, mime, status string
	var previewS3KeyNullStr sql.NullString
	var size uint64
	var versionNum int

	if err := scanner.Scan(
		&f.ID,
		&name,
		&previewS3KeyNullStr,
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

	if previewS3KeyNullStr.Valid {
		previewS3Key, err := file_version.NewS3Key(previewS3KeyNullStr.String)
		if err != nil {
			return nil, err
		}
		f.PreviewS3Key = &previewS3Key
	} else {
		f.PreviewS3Key = nil
	}

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

func (r *FileQueryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*file.File, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, name, preview_s3_key, mime, status, size, version_num,
               owner_id, uploaded_by_session_id, created_at, updated_at
        FROM files
        WHERE owner_id = $1
        ORDER BY created_at DESC
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	files := make([]*file.File, 0)

	for rows.Next() {
		f, err := scanFile(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, f)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return files, nil
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

// SearchByName поиск файлов пользователя по названию с пагинацией
func (r *FileQueryRepository) SearchByName(ctx context.Context, userID uuid.UUID, query string, limit int, skip int) ([]*file.File, int64, error) {
	// Получить общее количество совпадений
	var total int64
	countErr := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM files
        WHERE owner_id = $1 AND LOWER(name) LIKE LOWER($2)
    `, userID, "%"+query+"%").Scan(&total)
	if countErr != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", countErr)
	}

	rows, err := r.db.QueryContext(ctx, `
        SELECT id, name, preview_s3_key, mime, status, size, version_num,
               owner_id, uploaded_by_session_id, created_at, updated_at
        FROM files
        WHERE owner_id = $1 AND LOWER(name) LIKE LOWER($2)
        ORDER BY created_at DESC
        LIMIT $3 OFFSET $4
    `, userID, "%"+query+"%", limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	files := make([]*file.File, 0)

	for rows.Next() {
		f, err := scanFile(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, f)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return files, total, nil
}
