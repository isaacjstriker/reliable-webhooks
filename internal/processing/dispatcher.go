package processing

import (
	"context"
	"errors"
	"log/slog"
	"time"

	// "github.com/isaacjstriker/reliable-webhooks/internal/domain"
	"github.com/isaacjstriker/reliable-webhooks/internal/repository"
)

type Dispatcher struct {
	repo    repository.EventRepository
	logger  *slog.Logger
	queue   chan int64
	workers int
}

func NewDispatcher(repo repository.EventRepository, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		repo:    repo,
		logger:  logger,
		queue:   make(chan int64, 1000),
		workers: 4,
	}
}

func (d *Dispatcher) Enqueue(id int64) {
	select {
	case d.queue <- id:
	default:
		d.logger.Error("queue_full_dropping_event", "id", id)
	}
}

func (d *Dispatcher) Run(ctx context.Context) {
	for i := 0; i < d.workers; i++ {
		go d.worker(ctx, i)
	}
	<-ctx.Done()
	d.logger.Info("dispatcher_stopping")
}

func (d *Dispatcher) worker(ctx context.Context, workerID int) {
	d.logger.Info("worker_started", "worker", workerID)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("worker_exit", "worker", workerID)
			return
		case <-ticker.C:
			select {
			case <-ctx.Done():
				return
			case id := <-d.queue:
				d.process(ctx, id, workerID)
			default:
				evt, err := d.repo.ClaimNext(ctx)
				if err != nil {
					continue
				}
				d.process(ctx, evt.ID, workerID)
			}
		case id := <-d.queue:
			d.process(ctx, id, workerID)
		}
	}
}

func (d *Dispatcher) process(ctx context.Context, id int64, workerID int) {
	// Simplified: We re-claim the event ensuring it's processing.
	// In improvement iterations, fetch full payload / domain logic from ClaimNext.
	evt, err := d.repo.ClaimNext(ctx)
	if err != nil {
		// Could be no rows, or lock race, safe to ignore if mismatch
		return
	}
	if evt.ID != id {
		// Another event claimed; we still process claimed one.
	}

	time.Sleep(10 * time.Millisecond) // placeholder

	if err := d.repo.MarkProcessed(ctx, evt.ID); err != nil {
		d.logger.Error("mark_processed_error", "event_id", evt.ID, "error", err)
		return
	}
	d.logger.Info("event_processed", "id", evt.ID, "provider", evt.Provider, "worker", workerID)
}

// (Optional) differentiate error kinds later
var ErrNoEvent = errors.New("no event")