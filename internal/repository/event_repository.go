package repository

import (
	"context"
	"errors"
	"github.com/isaacjstriker/reliable-webhooks/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateEvent = errors.New("duplicate event")

type EventRepository interface {
	InsertIfNew(ctx context.Context, e *domain.Event) error
	ClaimNext(ctx context.Context) (*domain.Event, error)
	MarkProcessed(ctx context.Context, id int64) error
	MarkProcessing(ctx context.Context, id int64) error
	Ping() error
}

type eventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Ping() error {
	return r.db.Ping(context.Background())
}

func (r *eventRepository) InsertIfNew(ctx context.Context, e *domain.Event) error {
	q := `INSERT INTO events (provider, event_id, payload) VALUES ($1,$2,$3)
		  ON CONFLICT (provider, event_id) DO NOTHING
		  RETURNING id, status, created_at, updated_at`
	row := r.db.QueryRow(ctx, q, e.Provider, e.EventID, e.Payload)
	if err := row.Scan(&e.ID, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
		// If no rows returned, it's a duplicate
		if errors.Is(err, context.Canceled) {
			return err
		}
		if err.Error() == "no rows in result set" {
			return ErrDuplicateEvent
		}
		return err
	}
	return nil
}

func (r *eventRepository) ClaimNext(ctx context.Context) (*domain.Event, error) {
	// Simple row-level lock pattern; improvement: priority ordering, batching
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := `SELECT id, provider, event_id, payload, status, attempts, created_at, updated_at, processed_at
		  FROM events
		  WHERE status = 'received'
		  ORDER BY created_at
		  FOR UPDATE SKIP LOCKED
		  LIMIT 1`
	row := tx.QueryRow(ctx, q)
	var e domain.Event
	if err := row.Scan(&e.ID, &e.Provider, &e.EventID, &e.Payload, &e.Status, &e.Attempts, &e.CreatedAt, &e.UpdatedAt, &e.ProcessedAt); err != nil {
		return nil, err
	}

	// Mark processing
	_, err = tx.Exec(ctx, `UPDATE events SET status='processing', attempts=attempts+1 WHERE id=$1`, e.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	e.Status = domain.StatusProcessing
	e.Attempts += 1
	return &e, nil
}

func (r *eventRepository) MarkProcessing(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `UPDATE events SET status='processing', attempts=attempts+1 WHERE id=$1 AND status='received'`, id)
	return err
}

func (r *eventRepository) MarkProcessed(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `UPDATE events SET status='processed', processed_at=now() WHERE id=$1`, id)
	return err
}