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
	// Проверяем, есть ли уже транзакция (используем строку "tx")
	if _, ok := ctx.Value("tx").(*sql.Tx); ok {
		return fn(ctx)
	}

	// Начинаем новую транзакцию
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Сохраняем транзакцию с ключом "tx" (как ожидает репозиторий)
	ctx = context.WithValue(ctx, "tx", tx)

	// Выполняем функцию
	err = fn(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
