package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/port"
)

// Relay publishes outbox rows to Kafka (transactional outbox pattern).
// Delivery is at-least-once: a crash between publish and mark-sent re-sends.
type Relay struct {
	repo  port.ClaimRepository
	pub   port.EventPublisher
	log   *slog.Logger
	Batch int
	Every time.Duration
}

func NewRelay(repo port.ClaimRepository, pub port.EventPublisher, log *slog.Logger) *Relay {
	if log == nil {
		log = slog.Default()
	}
	return &Relay{repo: repo, pub: pub, log: log, Batch: 50, Every: 500 * time.Millisecond}
}

func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	return r.repo.ProcessOutbox(ctx, r.Batch, func(ev port.OutboxEvent) error {
		return r.pub.Publish(ctx, ev.Topic, ev.Key, ev.Payload)
	})
}

// Run polls until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	t := time.NewTicker(r.Every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			for { // drain
				n, err := r.RunOnce(ctx)
				if err != nil {
					r.log.Warn("outbox relay error", "err", err)
					break
				}
				if n < r.Batch {
					break
				}
			}
		}
	}
}
