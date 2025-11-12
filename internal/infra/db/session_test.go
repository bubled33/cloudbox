package db

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/cloud-file-storage/internal/domain/session"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

func TestSessionQueryRepository_GetByID_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewSessionQueryRepository(sqlDB)

	id := uuid.New()
	usedId := uuid.New()
	lastUsedAt := time.Now()
	createdAt := time.Now()
	updatedAt := time.Now()
	expiresAt := time.Now().AddDate(1, 0, 0)

	mock.ExpectQuery(`SELECT id, token_hash, refresh_token_hash, device_info, ip, is_revoked,
		       user_id, last_used_at, created_at, updated_at, expires_at FROM sessions WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "token_hash", "refresh_token_hash", "device_info", "ip", "is_revoked",
			"user_id", "last_used_at", "created_at", "updated_at", "expires_at",
		}).AddRow(id, "token_hash", "refresh_token_hash", "Windows 10", "192.168.1.1", true, usedId, lastUsedAt, createdAt, updatedAt, expiresAt))

	s, err := repo.GetByID(context.Background(), id)

	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, "Windows 10", s.DeviceInfo.String())
	require.Equal(t, "192.168.1.1", s.Ip.String())
	require.True(t, s.IsRevoked)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionQueryRepository_GetByUserID_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewSessionQueryRepository(sqlDB)
	userID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT id, token_hash, refresh_token_hash, device_info, ip, is_revoked,
		       user_id, last_used_at, created_at, updated_at, expires_at FROM sessions WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "token_hash", "refresh_token_hash", "device_info", "ip", "is_revoked",
			"user_id", "last_used_at", "created_at", "updated_at", "expires_at",
		}).AddRow(sessionID, "token_hash", "refresh_token_hash", "macOS", "10.0.0.2", false, userID, now, now, now, now.AddDate(1, 0, 0)))

	s, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, "macOS", s[0].DeviceInfo.String())
	require.Equal(t, "10.0.0.2", s[0].Ip.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionQueryRepository_GetAll_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewSessionQueryRepository(sqlDB)
	now := time.Now()
	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, token_hash, refresh_token_hash, device_info, ip, is_revoked,
		       user_id, last_used_at, created_at, updated_at, expires_at FROM sessions`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "token_hash", "refresh_token_hash", "device_info", "ip", "is_revoked",
			"user_id", "last_used_at", "created_at", "updated_at", "expires_at",
		}).AddRow(uuid.New(), "t1", "r1", "iPhone", "127.0.0.1", false, userID, now, now, now, now.AddDate(1, 0, 0)).
			AddRow(uuid.New(), "t2", "r2", "Windows", "192.168.1.5", true, userID, now, now, now, now.AddDate(1, 0, 0)))

	sessions, err := repo.GetAll(context.Background())
	require.NoError(t, err)
	require.Len(t, sessions, 2)
	require.Equal(t, "iPhone", sessions[0].DeviceInfo.String())
	require.Equal(t, "127.0.0.1", sessions[0].Ip.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionCommandRepository_Save_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	mock.ExpectBegin()
	tx, _ := sqlDB.Begin()
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := NewSessionCommandRepository()

	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	ip, _ := value_objects.NewIP(net.ParseIP("127.0.0.1"))
	expiresAt, _ := value_objects.NewExpiresAt(now.AddDate(1, 0, 0))

	session := session.Session{
		ID:               id,
		UserID:           userID,
		IsRevoked:        false,
		CreatedAt:        now,
		UpdatedAt:        now,
		LastUsedAt:       now,
		ExpiresAt:        expiresAt,
		DeviceInfo:       mustDeviceInfo("Windows"),
		TokenHash:        mustTokenHash("token1"),
		RefreshTokenHash: mustTokenHash("token2"),
		Ip:               ip,
	}

	mock.ExpectExec(`INSERT INTO sessions`).WithArgs(
		session.ID,
		session.TokenHash.String(),
		session.RefreshTokenHash.String(),
		session.DeviceInfo.String(),
		session.Ip.String(),
		session.IsRevoked,
		session.UserID,
		session.LastUsedAt,
		session.CreatedAt,
		session.UpdatedAt,
		session.ExpiresAt.Time(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(ctx, &session)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionCommandRepository_Save_NoTx(t *testing.T) {
	repo := NewSessionCommandRepository()
	err := repo.Save(context.Background(), &session.Session{})
	require.Error(t, err)
}

func TestSessionCommandRepository_Delete_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	mock.ExpectBegin()
	tx, _ := sqlDB.Begin()
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := NewSessionCommandRepository()
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM sessions WHERE id = \$1`).WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(ctx, id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionCommandRepository_Delete_NoTx(t *testing.T) {
	repo := NewSessionCommandRepository()
	err := repo.Delete(context.Background(), uuid.New())
	require.Error(t, err)
}

func mustTokenHash(v string) value_objects.TokenHash {
	h, err := value_objects.NewTokenHash(v)
	if err != nil {
		panic(err)
	}
	return h
}

func mustDeviceInfo(v string) value_objects.DeviceInfo {
	d, err := value_objects.NewDeviceInfo(v)
	if err != nil {
		panic(err)
	}
	return d
}
