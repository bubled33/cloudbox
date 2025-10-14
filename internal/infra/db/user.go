package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
)

type UserCommandRepository struct{}

func NewUserCommandRepository() *UserCommandRepository {
	return &UserCommandRepository{}
}

func (r *UserCommandRepository) Save(ctx context.Context, u *user.User) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `
	INSERT INTO users (id, email, display_name, is_email_verified, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id) DO UPDATE 
	SET email=$2, display_name=$3, is_email_verified=$4, updated_at=$5
	`
	_, err := tx.ExecContext(ctx, query,
		u.ID,
		u.Email.String(),
		u.DisplayName.String(),
		u.IsEmailVerified,
		u.UpdatedAt,
	)
	return err
}

func (r *UserCommandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `DELETE FROM users WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}

type UserQueryRepository struct {
	db *sql.DB
}

func NewUserQueryRepository(db *sql.DB) *UserQueryRepository {
	return &UserQueryRepository{db: db}
}

func (r *UserQueryRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, display_name, is_email_verified, updated_at
		FROM users
		WHERE id = $1
	`, id)

	var u user.User
	var email, displayName string
	if err := row.Scan(&u.ID, &email, &displayName, &u.IsEmailVerified, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var err error

	u.Email, err = user.NewEmail(email)
	if err != nil {
		return nil, err
	}

	u.DisplayName, err = user.NewDisplayName(displayName)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserQueryRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, display_name, is_email_verified, updated_at
		FROM users
		WHERE email = $1
	`, email)

	var u user.User
	var dbEmail, displayName string
	if err := row.Scan(&u.ID, &dbEmail, &displayName, &u.IsEmailVerified, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var err error

	u.Email, err = user.NewEmail(dbEmail)
	if err != nil {
		return nil, err
	}

	u.DisplayName, err = user.NewDisplayName(displayName)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserQueryRepository) GetAll(ctx context.Context) ([]*user.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, email, display_name, is_email_verified, updated_at
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var u user.User
		var email, displayName string
		if err := rows.Scan(&u.ID, &email, &displayName, &u.IsEmailVerified, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Email, err = user.NewEmail(email)
		u.DisplayName, err = user.NewDisplayName(displayName)
		users = append(users, &u)
	}
	return users, nil
}
