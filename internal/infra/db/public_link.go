package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/public_link"
)

type PublicLinkCommandRepository struct {
}

func (r *PublicLinkCommandRepository) Save(ctx context.Context, p *public_link.PublicLink) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `
    INSERT INTO public_links (id, file_id, created_by_user_id, token_hash,
               created_at, updated_at, expired_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    ON CONFLICT (id) DO UPDATE 
    SET file_id = $2, created_by_user_id = $3, token_hash = $4, 
        created_at = $5, updated_at = $6, expired_at = $7
    `
	_, err := tx.ExecContext(ctx, query,
		p.ID,
		p.FileID,
		p.CreatedByUserID,
		p.TokenHash,
		p.CreatedAt,
		p.UpdatedAt,
		p.ExpiredAt,
	)
	return err
}

func (r *PublicLinkCommandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `DELETE FROM public_links WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}

type PublicLinkQueryRepository struct {
	db *sql.DB
}

func NewPublicLinkCommandRepository() *PublicLinkCommandRepository {
	return &PublicLinkCommandRepository{}
}
func NewPublicLinkQueryRepository(db *sql.DB) *PublicLinkQueryRepository {
	return &PublicLinkQueryRepository{db: db}
}

func scanPublicLink(scanner scannable) (*public_link.PublicLink, error) {
	var p public_link.PublicLink
	var tokenHash string
	var expiredAt time.Time

	if err := scanner.Scan(
		&p.ID,
		&p.FileID,
		&p.CreatedByUserID,
		&tokenHash,
		&p.CreatedAt,
		&p.UpdatedAt,
		&expiredAt,
	); err != nil {
		return nil, err
	}

	p.TokenHash = tokenHash
	p.ExpiredAt = expiredAt

	return &p, nil
}

func (r *PublicLinkQueryRepository) GetByID(ctx context.Context, id uuid.UUID) (*public_link.PublicLink, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, file_id, created_by_user_id, token_hash,
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

func (r *PublicLinkQueryRepository) GetByToken(ctx context.Context, token string) (*public_link.PublicLink, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, file_id, created_by_user_id, token_hash,
               created_at, updated_at, expired_at
        FROM public_links
        WHERE token_hash = $1
    `, token)

	p, err := scanPublicLink(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *PublicLinkQueryRepository) GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*public_link.PublicLink, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, file_id, created_by_user_id, token_hash,
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
        SELECT id, file_id, created_by_user_id, token_hash,
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
