package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/session"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type SessionCommandRepository struct{}

func NewSessionCommandRepository() *SessionCommandRepository {
	return &SessionCommandRepository{}
}

type SessionQueryRepository struct {
	db *sql.DB
}

func NewSessionQueryRepository(db *sql.DB) *SessionQueryRepository {
	return &SessionQueryRepository{db: db}
}

func scanSession(scanner scannable) (*session.Session, error) {
	var s session.Session
	var tokenHash, refreshTokenHash, deviceInfo string
	var expiresAt time.Time
	var lastUsedAt sql.NullTime

	if err := scanner.Scan(
		&s.ID,
		&tokenHash,
		&refreshTokenHash,
		&deviceInfo,
		&s.Ip,
		&s.IsRevoked,
		&s.UserID,
		&lastUsedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
		&expiresAt,
	); err != nil {
		return nil, err
	}

	var err error
	s.TokenHash, err = value_objects.NewTokenHash(tokenHash)
	if err != nil {
		return nil, err
	}

	s.RefreshTokenHash, err = value_objects.NewTokenHash(refreshTokenHash)
	if err != nil {
		return nil, err
	}

	s.DeviceInfo, err = value_objects.NewDeviceInfo(deviceInfo)
	if err != nil {
		return nil, err
	}

	s.ExpiresAt, err = value_objects.NewExpiresAt(expiresAt)
	if err != nil {
		return nil, err
	}

	if lastUsedAt.Valid {
		s.LastUsedAt = lastUsedAt.Time
	} else {
		s.LastUsedAt = time.Time{} // или оставить zero value
	}

	return &s, nil
}

func (r *SessionQueryRepository) GetByID(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, token_hash, refresh_token_hash, device_info, ip, is_revoked,
		       user_id, last_used_at, created_at, updated_at, expires_at
		FROM sessions
		WHERE id = $1
	`, id)

	s, err := scanSession(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *SessionQueryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*session.Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, token_hash, refresh_token_hash, device_info, ip, is_revoked,
		       user_id, last_used_at, created_at, updated_at, expires_at
		FROM sessions
		WHERE user_id = $1
	`, userID)

	s, err := scanSession(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *SessionQueryRepository) GetAll(ctx context.Context) ([]*session.Session, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, token_hash, refresh_token_hash, device_info, ip, is_revoked,
		       user_id, last_used_at, created_at, updated_at, expires_at
		FROM sessions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*session.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}
