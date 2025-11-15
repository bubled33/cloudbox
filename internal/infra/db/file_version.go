package db

import (
	"context"
	"database/sql"
	"fmt"

	uuid "github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

type FileVersionCommandRepository struct {
}

func (r *FileVersionCommandRepository) Save(ctx context.Context, v *file_version.FileVersion) error {
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
    INSERT INTO file_versions (id, s3_key, preview_s3_key, mime, status, size, version_num,
               file_id, uploaded_by_session_id, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    ON CONFLICT (id) DO UPDATE 
    SET s3_key = EXCLUDED.s3_key, 
        preview_s3_key = EXCLUDED.preview_s3_key, 
        mime = EXCLUDED.mime, 
        status = EXCLUDED.status, 
        size = EXCLUDED.size, 
        version_num = EXCLUDED.version_num,
        file_id = EXCLUDED.file_id,
        uploaded_by_session_id = EXCLUDED.uploaded_by_session_id, 
        updated_at = EXCLUDED.updated_at
    `
	_, err := tx.ExecContext(ctx, query,
		v.ID,
		v.S3Key.String(),
		previewS3Key,
		v.Mime.String(),
		v.Status.String(),
		v.Size.Uint64(),
		v.VersionNum.Int(),
		v.FileId,
		v.UploadedBySessionId,
		v.CreatedAt,
		v.UpdatedAt,
	)
	return err
}

func (r *FileVersionCommandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `DELETE FROM file_versions WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
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

// scanFileVersion сканирует строку базы данных в объект FileVersion
func scanFileVersion(scanner scannable) (*file_version.FileVersion, error) {
	var v file_version.FileVersion
	var s3Key, mime, status string
	var previewS3KeyNullStr sql.NullString // Используем NullString для nullable поля
	var size uint64
	var versionNum int

	fmt.Println("scanFileVersion: Starting to scan row")

	if err := scanner.Scan(
		&v.ID,
		&s3Key,
		&previewS3KeyNullStr,
		&mime,
		&status,
		&size,
		&versionNum,
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

	if previewS3KeyNullStr.Valid {
		previewS3Key, err := file_version.NewS3Key(previewS3KeyNullStr.String)
		if err != nil {
			return nil, err
		}
		v.PreviewS3Key = &previewS3Key
	} else {
		v.PreviewS3Key = nil
	}

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

func (r *FileVersionQueryRepository) GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*file_version.FileVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, s3_key, preview_s3_key, mime, status, size, version_num,
               file_id, uploaded_by_session_id, created_at, updated_at
        FROM file_versions
        WHERE file_id = $1
        ORDER BY version_num DESC
    `, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query file versions: %w", err)
	}
	defer rows.Close()

	versions := make([]*file_version.FileVersion, 0)

	for rows.Next() {
		v, err := scanFileVersion(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file version: %w", err)
		}
		versions = append(versions, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return versions, nil
}

func (r *FileVersionQueryRepository) GetAll(ctx context.Context) ([]*file_version.FileVersion, error) {

	rows, err := r.db.QueryContext(ctx, `
        SELECT id, s3_key, preview_s3_key, mime, status, size, version_num,
               file_id, uploaded_by_session_id, created_at, updated_at
        FROM file_versions
    `)
	if err != nil {
		fmt.Printf("GetAll: Query error: %v\n", err)
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

func (r *FileVersionQueryRepository) GetAllByStatus(ctx context.Context, status file_version.FileStatus) ([]*file_version.FileVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, s3_key, preview_s3_key, mime, status, size, version_num,
               file_id, uploaded_by_session_id, created_at, updated_at
        FROM file_versions WHERE status = $1 ORDER BY created_at DESC
    `, status.String())
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
