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
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

func mustS3Key(val string) file_version.S3Key {
	k, err := file_version.NewS3Key(val)
	if err != nil {
		panic(err)
	}
	return k
}

func mustMime(val string) file_version.MimeType {
	m, err := file_version.NewMimeType(val)
	if err != nil {
		panic(err)
	}
	return m
}

func mustStatus(val string) file_version.FileStatus {
	s, err := file_version.NewFileStatus(val)
	if err != nil {
		panic(err)
	}
	return s
}

func mustFileVersionNum(v int) file_version.FileVersionNum {
	num, err := file_version.NewFileVersionNum(v)
	if err != nil {
		panic(err)
	}
	return num
}

func TestFileVersionCommandRepository_Save_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := NewFileVersionCommandRepository()

	id := uuid.New()
	fileID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	fileSize, err := file_version.NewFileSize(2048)
	require.NoError(t, err)

	previewS3Key, err := file_version.NewS3Key("preview-key")
	require.NoError(t, err)

	v := &file_version.FileVersion{
		ID:                  id,
		S3Key:               mustS3Key("main-file-key"),
		PreviewS3Key:        &previewS3Key,
		Mime:                mustMime("image/png"),
		Status:              mustStatus("uploaded"),
		Size:                fileSize,
		VersionNum:          mustFileVersionNum(1),
		FileId:              fileID,
		UploadedBySessionId: sessionID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(
			v.ID,
			v.S3Key.String(),
			v.PreviewS3Key.String(),
			v.Mime.String(),
			v.Status.String(),
			v.Size.Uint64(),
			v.VersionNum.Int(),
			v.FileId,
			v.UploadedBySessionId,
			v.CreatedAt,
			v.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(ctx, v)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFileVersionCommandRepository_Save_NoTransaction(t *testing.T) {
	repo := NewFileVersionCommandRepository()
	ctx := context.Background() // без tx
	v := &file_version.FileVersion{}

	err := repo.Save(ctx, v)
	require.ErrorIs(t, err, domainerrors.ErrTransactionNotFound)
}

func TestFileVersionCommandRepository_Delete_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := NewFileVersionCommandRepository()

	id := uuid.New()

	mock.ExpectExec(`DELETE FROM file_versions WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(ctx, id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFileVersionCommandRepository_Delete_NoTransaction(t *testing.T) {
	repo := NewFileVersionCommandRepository()
	ctx := context.Background()
	id := uuid.New()

	err := repo.Delete(ctx, id)
	require.ErrorIs(t, err, domainerrors.ErrTransactionNotFound)
}

func TestFileVersionQueryRepository_GetByID_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewFileVersionQueryRepository(sqlDB)

	id := uuid.New()
	fileID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT id, s3_key, preview_s3_key, mime, status, size, version_num,
		       file_id, uploaded_by_session_id, created_at, updated_at FROM file_versions WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "s3_key", "preview_s3_key", "mime", "status", "size", "version_num",
			"file_id", "uploaded_by_session_id", "created_at", "updated_at",
		}).AddRow(id, "main-key", "preview-key", "image/png", "uploaded", 2048, 1, fileID, sessionID, now, now))

	v, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.Equal(t, "main-key", v.S3Key.String())
	require.Equal(t, "preview-key", v.PreviewS3Key.String())
	require.Equal(t, "image/png", v.Mime.String())
	require.Equal(t, "uploaded", v.Status.String())
	require.Equal(t, uint64(2048), v.Size.Uint64())
	require.Equal(t, 1, v.VersionNum.Int())
}

func TestFileVersionQueryRepository_GetByID_NoRows(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewFileVersionQueryRepository(sqlDB)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, s3_key, preview_s3_key, mime, status, size, version_num,
		       file_id, uploaded_by_session_id, created_at, updated_at FROM file_versions WHERE id = \$1`).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	v, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.Nil(t, v)
}
