package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/session"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// Create создаёт новую сессию
func (r *SessionRepository) Create(ctx context.Context, s *session.Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, token_hash, refresh_token_hash, device_info,
			ip_address, expires_at, is_revoked, last_accessed_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.UserID, s.TokenHash, s.RefreshTokenHash, s.DeviceInfo,
		s.IPAddress.String(), s.ExpiresAt, s.IsRevoked, s.LastAccessedAt, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID получает сессию по ID
func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash, device_info,
			   ip_address, expires_at, is_revoked, last_accessed_at, created_at
		FROM sessions
		WHERE id = $1
	`

	s := &session.Session{}
	var ipAddressStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.RefreshTokenHash, &s.DeviceInfo,
		&ipAddressStr, &s.ExpiresAt, &s.IsRevoked, &s.LastAccessedAt, &s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, session.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	// Преобразовываем IP-адрес из строки
	s.IPAddress = net.ParseIP(ipAddressStr)

	return s, nil
}

// GetByTokenHash получает сессию по хешу токена
func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*session.Session, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash, device_info,
			   ip_address, expires_at, is_revoked, last_accessed_at, created_at
		FROM sessions
		WHERE token_hash = $1
	`

	s := &session.Session{}
	var ipAddressStr string

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.RefreshTokenHash, &s.DeviceInfo,
		&ipAddressStr, &s.ExpiresAt, &s.IsRevoked, &s.LastAccessedAt, &s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, session.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by token hash: %w", err)
	}

	// Преобразовываем IP-адрес из строки
	s.IPAddress = net.ParseIP(ipAddressStr)

	return s, nil
}

// Update обновляет сессию
func (r *SessionRepository) Update(ctx context.Context, s *session.Session) error {
	query := `
		UPDATE sessions 
		SET last_accessed_at = $2, is_revoked = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, s.ID, s.LastAccessedAt, s.IsRevoked)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// Touch обновляет время последнего доступа к сессии
func (r *SessionRepository) Touch(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET last_accessed_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to touch session: %w", err)
	}

	return nil
}

// Revoke отменяет сессию
func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET is_revoked = true WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

// DeleteExpired удаляет просроченные сессии
func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < $1 OR is_revoked = true`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

// GetAllByUserID получает все активные сессии пользователя
func (r *SessionRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*session.Session, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash, device_info,
			   ip_address, expires_at, is_revoked, last_accessed_at, created_at
		FROM sessions
		WHERE user_id = $1 AND is_revoked = false AND expires_at > $2
		ORDER BY last_accessed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user ID: %w", err)
	}
	defer rows.Close()

	var sessions []*session.Session
	for rows.Next() {
		s := &session.Session{}
		var ipAddressStr string

		err := rows.Scan(
			&s.ID, &s.UserID, &s.TokenHash, &s.RefreshTokenHash, &s.DeviceInfo,
			&ipAddressStr, &s.ExpiresAt, &s.IsRevoked, &s.LastAccessedAt, &s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Преобразовываем IP-адрес из строки
		s.IPAddress = net.ParseIP(ipAddressStr)
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}