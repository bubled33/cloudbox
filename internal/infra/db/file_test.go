package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/file"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

// --- Вспомогательные функции ---
func mustFileName(val string) file.FileName {
	n, err := file.NewFileName(val)
	if err != nil {
		panic(err)
	}
	return n
}

func mustFileSize(size uint64) file_version.FileSize {
	s, err := file_version.NewFileSize(size)
	if err != nil {
		panic(err)
	}
	return s
}

// --- TESTS ---
func mustS3KeyPtr(val string) *file_version.S3Key {
	k := mustS3Key(val)
	return &k
}

func TestFileCommandRepository_Save_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := NewFileCommandRepository()

	id := uuid.New()
	ownerID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	f := &file.File{
		ID:                  id,
		Name:                mustFileName("file1.txt"),
		PreviewS3Key:        mustS3KeyPtr("preview-key"),
		Mime:                mustMime("text/plain"),
		Status:              mustStatus("uploaded"),
		Size:                mustFileSize(1024),
		VersionNum:          mustFileVersionNum(1),
		OwnerID:             ownerID,
		UploadedBySessionId: sessionID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	mock.ExpectExec(`INSERT INTO files`).
		WithArgs(
			f.ID,
			f.Name.String(),
			f.PreviewS3Key.String(),
			f.Mime.String(),
			f.Status.String(),
			f.Size.Uint64(),
			f.VersionNum.Int(),
			f.OwnerID,
			f.UploadedBySessionId,
			f.CreatedAt,
			f.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(ctx, f)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFileCommandRepository_Save_NoTransaction(t *testing.T) {
	repo := NewFileCommandRepository()
	ctx := context.Background()
	f := &file.File{}

	err := repo.Save(ctx, f)
	require.ErrorIs(t, err, domainerrors.ErrTransactionNotFound)
}

func TestFileCommandRepository_Delete_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := NewFileCommandRepository()

	id := uuid.New()

	mock.ExpectExec(`DELETE FROM files WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(ctx, id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFileCommandRepository_Delete_NoTransaction(t *testing.T) {
	repo := NewFileCommandRepository()
	ctx := context.Background()
	id := uuid.New()

	err := repo.Delete(ctx, id)
	require.ErrorIs(t, err, domainerrors.ErrTransactionNotFound)
}

func TestFileQueryRepository_GetByID_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewFileQueryRepository(sqlDB)

	id := uuid.New()
	ownerID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT id, name, preview_s3_key, mime, status, size, version_num,`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "preview_s3_key", "mime", "status", "size", "version_num",
			"owner_id", "uploaded_by_session_id", "created_at", "updated_at",
		}).AddRow(id, "file1.txt", "preview-key", "text/plain", "uploaded", 1024, 1, ownerID, sessionID, now, now))

	f, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, f)
	require.Equal(t, "file1.txt", f.Name.String())
	require.Equal(t, "preview-key", f.PreviewS3Key.String())
	require.Equal(t, "text/plain", f.Mime.String())
	require.Equal(t, "uploaded", f.Status.String())
	require.Equal(t, uint64(1024), f.Size.Uint64())
	require.Equal(t, 1, f.VersionNum.Int())
}

func TestFileQueryRepository_GetByID_NoRows(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewFileQueryRepository(sqlDB)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, name, preview_s3_key, mime, status, size, version_num,`).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	f, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.Nil(t, f)
}

func TestFileQueryRepository_GetAll_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewFileQueryRepository(sqlDB)

	now := time.Now()
	ownerID := uuid.New()
	sessionID := uuid.New()

	mock.ExpectQuery(`SELECT id, name, preview_s3_key, mime, status, size, version_num,`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "preview_s3_key", "mime", "status", "size", "version_num",
			"owner_id", "uploaded_by_session_id", "created_at", "updated_at",
		}).AddRow(uuid.New(), "file1.txt", "preview-key", "text/plain", "uploaded", 1024, 1, ownerID, sessionID, now, now).
			AddRow(uuid.New(), "file2.txt", "preview-key2", "image/png", "uploaded", 2048, 2, ownerID, sessionID, now, now))

	files, err := repo.GetAll(context.Background())
	require.NoError(t, err)
	require.Len(t, files, 2)
	require.Equal(t, "file1.txt", files[0].Name.String())
	require.Equal(t, "file2.txt", files[1].Name.String())
}
