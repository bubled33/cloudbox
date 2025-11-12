package app

import (
	"context"
	"database/sql"
)

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {

	if _, ok := ctx.Value("tx").(*sql.Tx); ok {
		return fn(ctx)
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "tx", tx)

	err = fn(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
