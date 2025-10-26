package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/magic_link"
)

type MagicLinkRepository struct {
	db *sql.DB
}

func NewMagicLinkRepository(db *sql.DB) *MagicLinkRepository {
	return &MagicLinkRepository{
		db: db,
	}
}

// Create создаёт новую магическую ссылку
func (r *MagicLinkRepository) Create(ctx context.Context, ml *magic_link.MagicLink) error {
	query := `
		INSERT INTO magic_links (
			id, user_id, token_hash, device_info, purpose, 
			ip_address, expires_at, is_used, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		ml.ID, ml.UserID, ml.TokenHash, ml.DeviceInfo.String(), ml.Purpose,
		ml.IPAddress.String(), ml.ExpiresAt, ml.IsUsed, ml.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create magic link: %w", err)
	}

	return nil
}

// GetByTokenHash получает магическую ссылку по хешу токена
func (r *MagicLinkRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*magic_link.MagicLink, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, purpose,
			   ip_address, expires_at, is_used, created_at
		FROM magic_links
		WHERE token_hash = $1
	`

	ml := &magic_link.MagicLink{}
	var deviceInfoStr, ipAddressStr string

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&ml.ID, &ml.UserID, &ml.TokenHash, &deviceInfoStr, &ml.Purpose,
		&ipAddressStr, &ml.ExpiresAt, &ml.IsUsed, &ml.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, magic_link.ErrMagicLinkNotFound
		}
		return nil, fmt.Errorf("failed to get magic link by token hash: %w", err)
	}

	// Преобразовываем строки обратно в типы
	ml.DeviceInfo = magic_link.DeviceInfo(deviceInfoStr)
	ml.IPAddress = net.ParseIP(ipAddressStr)

	return ml, nil
}

// GetByID получает магическую ссылку по ID
func (r *MagicLinkRepository) GetByID(ctx context.Context, id uuid.UUID) (*magic_link.MagicLink, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, purpose,
			   ip_address, expires_at, is_used, created_at
		FROM magic_links
		WHERE id = $1
	`

	ml := &magic_link.MagicLink{}
	var deviceInfoStr, ipAddressStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ml.ID, &ml.UserID, &ml.TokenHash, &deviceInfoStr, &ml.Purpose,
		&ipAddressStr, &ml.ExpiresAt, &ml.IsUsed, &ml.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, magic_link.ErrMagicLinkNotFound
		}
		return nil, fmt.Errorf("failed to get magic link by ID: %w", err)
	}

	// Преобразовываем строки обратно в типы
	ml.DeviceInfo = magic_link.DeviceInfo(deviceInfoStr)
	ml.IPAddress = net.ParseIP(ipAddressStr)

	return ml, nil
}

// MarkAsUsed отмечает магическую ссылку как использованную
func (r *MagicLinkRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE magic_links SET is_used = true WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark magic link as used: %w", err)
	}

	return nil
}

// DeleteExpired удаляет просроченные магические ссылки
func (r *MagicLinkRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM magic_links WHERE expires_at < $1`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired magic links: %w", err)
	}

	return nil
}