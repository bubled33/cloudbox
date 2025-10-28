package db

import (
	"context"
	"database/sql"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/magic_link"
	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type MagicLinkCommandRepository struct {
}

type MagicLinkQueryRepository struct {
	db *sql.DB
}

func NewMagicLinkQueryRepository(db *sql.DB) *MagicLinkQueryRepository {
	return &MagicLinkQueryRepository{db: db}
}

func NewMagicLinkCommandRepository() *MagicLinkCommandRepository {
	return &MagicLinkCommandRepository{}
}
func (r *MagicLinkCommandRepository) Save(ctx context.Context, m *magic_link.MagicLink) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `
    INSERT INTO magic_links (id, user_id, token_hash, device_info, purpose, ip,
               is_used, used_at, created_at, updated_at, expired_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    ON CONFLICT (id) DO UPDATE 
    SET user_id=$2, token_hash=$3, device_info=$4, purpose=$5, ip=$6,
        is_used=$7, used_at=$8, created_at=$9, updated_at=$10, expired_at=$11
    `
	_, err := tx.ExecContext(ctx, query,
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
	)
	return err
}

func (r *MagicLinkCommandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `DELETE FROM magic_links WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}

func scanMagicLink(scanner scannable) (*magic_link.MagicLink, error) {
	var m magic_link.MagicLink
	var tokenHashStr, deviceInfoStr, ipStr, purposeStr string
	var usedAt sql.NullTime
	var expiredAt time.Time

	if err := scanner.Scan(
		&m.ID,
		&m.UserID,
		&tokenHashStr,
		&deviceInfoStr,
		&purposeStr,
		&ipStr,
		&m.IsUsed,
		&usedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
		&expiredAt,
	); err != nil {
		return nil, err
	}

	var err error

	m.TokenHash, err = value_objects.NewTokenHash(tokenHashStr)
	if err != nil {
		return nil, err
	}

	m.DeviceInfo, err = value_objects.NewDeviceInfo(deviceInfoStr)
	if err != nil {
		return nil, err
	}

	m.Ip, err = value_objects.NewIP(net.ParseIP(ipStr))
	if err != nil {
		return nil, err
	}

	m.Purpose, err = magic_link.NewPurpose(purposeStr)
	if err != nil {
		return nil, err
	}

	if usedAt.Valid {
		m.UsedAt = &usedAt.Time
	} else {
		m.UsedAt = nil
	}

	m.ExpiredAt, err = value_objects.NewExpiresAt(expiredAt)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *MagicLinkQueryRepository) GetByID(ctx context.Context, id uuid.UUID) (*magic_link.MagicLink, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, user_id, token_hash, device_info, purpose, ip,
               is_used, used_at, created_at, updated_at, expired_at
        FROM magic_links
        WHERE id = $1
    `, id)

	m, err := scanMagicLink(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func (r *MagicLinkQueryRepository) GetByTokenHash(ctx context.Context, token value_objects.TokenHash) (*magic_link.MagicLink, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, user_id, token_hash, device_info, purpose, ip,
               is_used, used_at, created_at, updated_at, expired_at
        FROM magic_links
        WHERE token_hash = $1
    `, token.String())

	m, err := scanMagicLink(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func (r *MagicLinkQueryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*magic_link.MagicLink, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, user_id, token_hash, device_info, purpose, ip,
               is_used, used_at, created_at, updated_at, expired_at
        FROM magic_links
        WHERE user_id = $1
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*magic_link.MagicLink
	for rows.Next() {
		m, err := scanMagicLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *MagicLinkQueryRepository) GetAll(ctx context.Context) ([]*magic_link.MagicLink, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, user_id, token_hash, device_info, purpose, ip,
               is_used, used_at, created_at, updated_at, expired_at
        FROM magic_links
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*magic_link.MagicLink
	for rows.Next() {
		m, err := scanMagicLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
