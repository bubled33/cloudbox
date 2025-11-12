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
	"github.com/yourusername/cloud-file-storage/internal/domain/public_link"
)

func TestPublicLinkCommandRepository_Save_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := &PublicLinkCommandRepository{}

	id := uuid.New()
	fileID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	expiredAt := now.Add(time.Hour)

	p := public_link.PublicLink{
		ID:              id,
		FileID:          fileID,
		CreatedByUserID: userID,
		TokenHash:       "token123",
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiredAt:       expiredAt,
	}

	mock.ExpectExec(`INSERT INTO public_links`).
		WithArgs(
			p.ID,
			p.FileID,
			p.CreatedByUserID,
			p.TokenHash,
			p.CreatedAt,
			p.UpdatedAt,
			p.ExpiredAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(ctx, &p)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPublicLinkCommandRepository_Save_NoTransaction(t *testing.T) {
	repo := &PublicLinkCommandRepository{}
	ctx := context.Background()
	p := public_link.PublicLink{}

	err := repo.Save(ctx, &p)
	require.ErrorIs(t, err, domainerrors.ErrTransactionNotFound)
}

func TestPublicLinkCommandRepository_Delete_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := &PublicLinkCommandRepository{}
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM public_links WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(ctx, id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPublicLinkCommandRepository_Delete_NoTransaction(t *testing.T) {
	repo := &PublicLinkCommandRepository{}
	ctx := context.Background()
	id := uuid.New()

	err := repo.Delete(ctx, id)
	require.ErrorIs(t, err, domainerrors.ErrTransactionNotFound)
}

func TestPublicLinkQueryRepository_GetByID_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewPublicLinkQueryRepository(sqlDB)

	id := uuid.New()
	fileID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	expiredAt := now.Add(time.Hour)

	mock.ExpectQuery(`SELECT id, file_id, created_by_user_id, token_hash,`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "file_id", "created_by_user_id", "token_hash",
			"created_at", "updated_at", "expired_at",
		}).AddRow(id, fileID, userID, "token123", now, now, expiredAt))

	p, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, "token123", p.TokenHash)
	require.Equal(t, fileID, p.FileID)
}

func TestPublicLinkQueryRepository_GetByID_NoRows(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewPublicLinkQueryRepository(sqlDB)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, file_id, created_by_user_id, token_hash,`).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	p, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.Nil(t, p)
}

func TestPublicLinkQueryRepository_GetByFileID_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewPublicLinkQueryRepository(sqlDB)
	fileID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT id, file_id, created_by_user_id, token_hash,`).
		WithArgs(fileID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "file_id", "created_by_user_id", "token_hash",
			"created_at", "updated_at", "expired_at",
		}).AddRow(uuid.New(), fileID, uuid.New(), "token1", now, now, now.Add(time.Hour)).
			AddRow(uuid.New(), fileID, uuid.New(), "token2", now, now, now.Add(2*time.Hour)))

	links, err := repo.GetByFileID(context.Background(), fileID)
	require.NoError(t, err)
	require.Len(t, links, 2)
	require.Equal(t, "token1", links[0].TokenHash)
	require.Equal(t, "token2", links[1].TokenHash)
}

func TestPublicLinkQueryRepository_GetAll_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewPublicLinkQueryRepository(sqlDB)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, file_id, created_by_user_id, token_hash,`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "file_id", "created_by_user_id", "token_hash",
			"created_at", "updated_at", "expired_at",
		}).AddRow(uuid.New(), uuid.New(), uuid.New(), "token1", now, now, now.Add(time.Hour)).
			AddRow(uuid.New(), uuid.New(), uuid.New(), "token2", now, now, now.Add(2*time.Hour)))

	links, err := repo.GetAll(context.Background())
	require.NoError(t, err)
	require.Len(t, links, 2)
	require.Equal(t, "token1", links[0].TokenHash)
	require.Equal(t, "token2", links[1].TokenHash)
}
