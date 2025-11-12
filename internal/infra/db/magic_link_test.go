package db

import (
	"context"
	"database/sql"
	"net"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/magic_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

func mustIP(val string) value_objects.IP {
	ip, err := value_objects.NewIP(net.ParseIP(val))
	if err != nil {
		panic(err)
	}
	return ip
}

func mustPurpose(val string) magic_link.Purpose {
	p, err := magic_link.NewPurpose(val)
	if err != nil {
		panic(err)
	}
	return p
}

func mustExpiresAt(t time.Time) value_objects.ExpiresAt {
	e, err := value_objects.NewExpiresAt(t)
	if err != nil {
		panic(err)
	}
	return e
}

func TestMagicLinkCommandRepository_Save_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := &MagicLinkCommandRepository{}

	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	usedAt := now.Add(-time.Hour)
	expiredAt := now.Add(time.Hour)

	m := magic_link.MagicLink{
		ID:         id,
		UserID:     userID,
		TokenHash:  mustTokenHash("token123"),
		DeviceInfo: mustDeviceInfo("iPhone"),
		Ip:         mustIP("127.0.0.1"),
		Purpose:    mustPurpose("login"),
		IsUsed:     true,
		UsedAt:     &usedAt,
		CreatedAt:  now,
		UpdatedAt:  now,
		ExpiredAt:  mustExpiresAt(expiredAt),
	}

	mock.ExpectExec(`INSERT INTO magic_links`).
		WithArgs(
			m.ID,
			m.UserID,
			m.TokenHash.String(),
			m.DeviceInfo.String(),
			m.Purpose.String(),
			m.Ip.String(),
			m.IsUsed,
			m.UsedAt,
			m.CreatedAt,
			m.UpdatedAt,
			m.ExpiredAt.Time(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(ctx, &m)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMagicLinkCommandRepository_Save_NoTransaction(t *testing.T) {
	repo := &MagicLinkCommandRepository{}
	ctx := context.Background()
	m := magic_link.MagicLink{}

	err := repo.Save(ctx, &m)
	require.ErrorIs(t, err, domainerrors.ErrTransactionNotFound)
}

func TestMagicLinkCommandRepository_Delete_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectBegin()
	tx, err := sqlDB.Begin()
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), "tx", tx)

	repo := &MagicLinkCommandRepository{}
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM magic_links WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(ctx, id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMagicLinkCommandRepository_Delete_NoTransaction(t *testing.T) {
	repo := &MagicLinkCommandRepository{}
	ctx := context.Background()
	id := uuid.New()

	err := repo.Delete(ctx, id)
	require.ErrorIs(t, err, domainerrors.ErrTransactionNotFound)
}

func TestMagicLinkQueryRepository_GetByID_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewMagicLinkQueryRepository(sqlDB)

	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT id, user_id, token_hash, device_info, purpose, ip,`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "token_hash", "device_info", "purpose", "ip",
			"is_used", "used_at", "created_at", "updated_at", "expired_at",
		}).AddRow(id, userID, "token123", "iPhone", "login", "127.0.0.1", true, now, now, now, now.Add(time.Hour)))

	m, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, "token123", m.TokenHash.String())
	require.Equal(t, "iPhone", m.DeviceInfo.String())
	require.Equal(t, "127.0.0.1", m.Ip.String())
	require.Equal(t, "login", m.Purpose.String())
}

func TestMagicLinkQueryRepository_GetByID_NoRows(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewMagicLinkQueryRepository(sqlDB)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, user_id, token_hash, device_info, purpose, ip,`).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	m, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.Nil(t, m)
}

func TestMagicLinkQueryRepository_GetAll_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewMagicLinkQueryRepository(sqlDB)
	now := time.Now()
	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, user_id, token_hash, device_info, purpose, ip,`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "token_hash", "device_info", "purpose", "ip",
			"is_used", "used_at", "created_at", "updated_at", "expired_at",
		}).AddRow(uuid.New(), userID, "token1", "iPhone", "login", "127.0.0.1", true, now, now, now, now.Add(time.Hour)).
			AddRow(uuid.New(), userID, "token2", "Android", "login", "127.0.0.2", false, nil, now, now, now.Add(time.Hour)))

	links, err := repo.GetAll(context.Background())
	require.NoError(t, err)
	require.Len(t, links, 2)
	require.Equal(t, "token1", links[0].TokenHash.String())
	require.Equal(t, "token2", links[1].TokenHash.String())
}
