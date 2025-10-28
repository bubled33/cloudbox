package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/session"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type SessionCommandRepository struct{}

func (r *SessionCommandRepository) Save(ctx context.Context, s *session.Session) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	fmt.Printf("=== SAVE DEBUG ===\n")
	fmt.Printf("Session ID: %v\n", s.ID)
	fmt.Printf("Session.ExpiresAt type: %T\n", s.ExpiresAt)
	fmt.Printf("Session.ExpiresAt.Time(): %v\n", s.ExpiresAt.Time())
	fmt.Printf("Is zero: %v\n", s.ExpiresAt.Time().IsZero())

	expiresAtTime := s.ExpiresAt.Time()

	fmt.Printf("Params for DB:\n")
	fmt.Printf("  $1 (id): %v\n", s.ID)
	fmt.Printf("  $2 (token_hash): %v\n", s.TokenHash.String())
	fmt.Printf("  $11 (expires_at): %v (type: %T)\n", expiresAtTime, expiresAtTime)

	query := `
    INSERT INTO sessions (id, token_hash, refresh_token_hash, device_info, ip, is_revoked,
               user_id, last_used_at, created_at, updated_at, expired_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    ON CONFLICT (id) DO UPDATE 
    SET token_hash = $2, 
        refresh_token_hash = $3, 
        device_info = $4, 
        ip = $5, 
        is_revoked = $6,
        user_id = $7, 
        last_used_at = $8,
        created_at = $9, 
        updated_at = $10,
        expired_at = $11
    `

	_, err := tx.ExecContext(ctx, query,
		s.ID,                        // $1
		s.TokenHash.String(),        // $2
		s.RefreshTokenHash.String(), // $3
		s.DeviceInfo.String(),       // $4
		s.Ip.String(),               // $5
		s.IsRevoked,                 // $6
		s.UserID,                    // $7
		s.LastUsedAt,                // $8
		s.CreatedAt,                 // $9
		s.UpdatedAt,                 // $10
		expiresAtTime,               // $11
	)

	if err != nil {
		fmt.Printf("ERROR executing query: %v\n", err)
	} else {
		fmt.Printf("Query executed successfully\n")
	}
	fmt.Printf("=== END SAVE DEBUG ===\n")

	return err
}

func (r *SessionCommandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `DELETE FROM sessions WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}

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
		s.LastUsedAt = time.Time{}
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

func (r *SessionQueryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*session.Session, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, token_hash, refresh_token_hash, device_info, ip, is_revoked,
               user_id, last_used_at, created_at, updated_at, expires_at
        FROM sessions
        WHERE user_id = $1
    `, userID)
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
func (r *SessionQueryRepository) GetByAccessToken(ctx context.Context, tokenHash *value_objects.TokenHash) (*session.Session, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, token_hash, refresh_token_hash, device_info, ip, is_revoked,
               user_id, last_used_at, created_at, updated_at, expired_at
        FROM sessions
        WHERE token_hash = $1
    `, tokenHash.String())

	s, err := scanSession(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}
