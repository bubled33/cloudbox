package db

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
)

func TestUserQueryRepository_GetByEmail_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewUserQueryRepository(sqlDB)

	id := uuid.New()
	updatedAt := time.Now()

	mock.ExpectQuery(`SELECT id, email, display_name, is_email_verified, updated_at FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "display_name", "is_email_verified", "updated_at",
		}).AddRow(id, "test@example.com", "John Doe", true, updatedAt))

	u, err := repo.GetByEmail(context.Background(), "test@example.com")

	require.NoError(t, err)
	require.NotNil(t, u)
	require.Equal(t, "test@example.com", u.Email.String())
	require.Equal(t, "John Doe", u.DisplayName.String())
	require.True(t, u.IsEmailVerified)
	require.WithinDuration(t, updatedAt, u.UpdatedAt, time.Second)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserQueryRepository_GetByID_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewUserQueryRepository(sqlDB)

	id := uuid.New()
	updatedAt := time.Now()

	mock.ExpectQuery(`SELECT id, email, display_name, is_email_verified, updated_at FROM users WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "display_name", "is_email_verified", "updated_at",
		}).AddRow(id, "test@example.com", "John Doe", true, updatedAt))

	u, err := repo.GetByID(context.Background(), id)

	require.NoError(t, err)
	require.NotNil(t, u)
	require.Equal(t, "test@example.com", u.Email.String())
	require.Equal(t, "John Doe", u.DisplayName.String())
	require.True(t, u.IsEmailVerified)
	require.WithinDuration(t, updatedAt, u.UpdatedAt, time.Second)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserQueryRepository_GetAll_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewUserQueryRepository(sqlDB)

	id := uuid.New()
	updatedAt := time.Now()

	mock.ExpectQuery(`SELECT id, email, display_name, is_email_verified, updated_at FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "display_name", "is_email_verified", "updated_at",
		}).AddRow(id, "test@example.com", "John Doe", true, updatedAt))

	users, err := repo.GetAll(context.Background())

	require.NoError(t, err)
	require.NotNil(t, users)
	require.Len(t, users, 1)
	require.Equal(t, "test@example.com", users[0].Email.String())
	require.Equal(t, "John Doe", users[0].DisplayName.String())
	require.True(t, users[0].IsEmailVerified)
	require.WithinDuration(t, updatedAt, users[0].UpdatedAt, time.Second)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserQueryRepository_Save_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "tx", tx)
	repo := NewUserCommandRepository()

	email, err := user.NewEmail("test@example.com")
	require.NoError(t, err)

	displayName, err := user.NewDisplayName("John Doe")
	require.NoError(t, err)

	u := user.NewUser(
		email, displayName,
	)

	mock.ExpectExec(`INSERT INTO users`).WithArgs(u.ID, u.Email.String(), u.DisplayName.String(), u.IsEmailVerified, u.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(ctx, u)

	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserQueryRepository_Delete_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "tx", tx)
	repo := NewUserCommandRepository()

	email, err := user.NewEmail("test@example.com")
	require.NoError(t, err)

	displayName, err := user.NewDisplayName("John Doe")
	require.NoError(t, err)

	u := user.NewUser(
		email, displayName,
	)

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).WithArgs(u.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(ctx, u.ID)

	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
