package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/public_link"
)

type PublicLinkQueryRepository struct {
	db *sql.DB
}

func NewPublicLinkQueryRepository(db *sql.DB) *PublicLinkQueryRepository {
	return &PublicLinkQueryRepository{db: db}
}

func scanPublicLink(scanner scannable) (*public_link.PublicLink, error) {
	var p public_link.PublicLink
	var tokenHash string
	var isExpired bool
	var expiredAt sql.NullTime // может быть NULL в базе

	if err := scanner.Scan(
		&p.ID,
		&p.FileID,
		&p.CreatedByUserID,
		&tokenHash,
		&isExpired,
		&p.CreatedAt,
		&p.UpdatedAt,
		&expiredAt,
	); err != nil {
		return nil, err
	}

	p.TokenHash = tokenHash
	p.IsExpired = isExpired
	if expiredAt.Valid {
		p.ExpiredAt = expiredAt.Time
	}

	return &p, nil
}

func (r *PublicLinkQueryRepository) GetByID(ctx context.Context, id uuid.UUID) (*public_link.PublicLink, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, file_id, created_by_user_id, token_hash, is_expired,
		       created_at, updated_at, expired_at
		FROM public_links
		WHERE id = $1
	`, id)

	p, err := scanPublicLink(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *PublicLinkQueryRepository) GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*public_link.PublicLink, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, file_id, created_by_user_id, token_hash, is_expired,
		       created_at, updated_at, expired_at
		FROM public_links
		WHERE file_id = $1
	`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*public_link.PublicLink
	for rows.Next() {
		p, err := scanPublicLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *PublicLinkQueryRepository) GetAll(ctx context.Context) ([]*public_link.PublicLink, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, file_id, created_by_user_id, token_hash, is_expired,
		       created_at, updated_at, expired_at
		FROM public_links
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*public_link.PublicLink
	for rows.Next() {
		p, err := scanPublicLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
