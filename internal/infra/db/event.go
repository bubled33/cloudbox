package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/domainerrors"
	"github.com/yourusername/cloud-file-storage/internal/domain/event"
)

// EventCommandRepository - реализация команд
type EventCommandRepository struct{}

func NewEventCommandRepository() *EventCommandRepository {
	return &EventCommandRepository{}
}

func (r *EventCommandRepository) Save(ctx context.Context, e *event.Event) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `
    INSERT INTO events (id, name, data, created_at, sent, locked_at, locked_by, retry_count)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    ON CONFLICT (id) DO UPDATE 
    SET name = $2, data = $3, sent = $5,
        locked_at = $6, locked_by = $7, retry_count = $8
    `
	_, err := tx.ExecContext(ctx, query,
		e.ID,
		e.Name,
		e.Data,
		e.CreatedAt,
		e.Sent,
		e.LockedAt,
		e.LockedBy,
		e.RetryCount,
	)
	return err
}

func (r *EventCommandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `DELETE FROM events WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}

func (r *EventCommandRepository) MarkAsSent(ctx context.Context, id uuid.UUID) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `UPDATE events SET sent = true WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}

func (r *EventCommandRepository) UpdateRetryCount(ctx context.Context, id uuid.UUID, retryCount int) error {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return domainerrors.ErrTransactionNotFound
	}

	query := `UPDATE events SET retry_count = $2 WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id, retryCount)
	return err
}

// EventQueryRepository - реализация запросов
type EventQueryRepository struct {
	db *sql.DB
}

func NewEventQueryRepository(db *sql.DB) *EventQueryRepository {
	return &EventQueryRepository{db: db}
}

func scanEvent(scanner scannable) (*event.Event, error) {
	var e event.Event
	var lockedAt sql.NullTime
	var lockedBy sql.NullString

	if err := scanner.Scan(
		&e.ID,
		&e.Name,
		&e.Data,
		&e.CreatedAt,
		&e.Sent,
		&lockedAt,
		&lockedBy,
		&e.RetryCount,
	); err != nil {
		return nil, err
	}

	if lockedAt.Valid {
		e.LockedAt = &lockedAt.Time
	}

	if lockedBy.Valid {
		e.LockedBy = &lockedBy.String
	}

	return &e, nil
}

func (r *EventQueryRepository) GetByID(ctx context.Context, id uuid.UUID) (*event.Event, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, name, data, created_at, sent, locked_at, locked_by, retry_count
        FROM events
        WHERE id = $1
    `, id)

	e, err := scanEvent(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return e, err
}

func (r *EventQueryRepository) GetPending(ctx context.Context, limit int) ([]*event.Event, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, name, data, created_at, sent, locked_at, locked_by, retry_count
        FROM events
        WHERE sent = false
        ORDER BY created_at ASC
        LIMIT $1
    `, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*event.Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventQueryRepository) GetAll(ctx context.Context) ([]*event.Event, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, name, data, created_at, sent, locked_at, locked_by, retry_count
        FROM events
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*event.Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
