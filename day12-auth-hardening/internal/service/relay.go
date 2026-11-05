package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/port"
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
		// The relay runs on its own goroutine, so ctx has no request in it.
		// Join the trace of the request that wrote this row, if its payload carries a valid traceparent.
		var tp struct {
			Traceparent string `json:"traceparent"`
		}
		_ = json.Unmarshal(ev.Payload, &tp)
		pc := ctx
		if sc, ok := obs.ParseTraceparent(tp.Traceparent); ok {
			pc = obs.ContextWithRemote(ctx, sc)
		}
		pctx, sp := obs.StartKind(pc, "outbox publish "+ev.Topic, obs.KindProducer, obs.Attr{Key: "messaging.destination.name", Val: ev.Topic})
		err := r.pub.Publish(pctx, ev.Topic, ev.Key, ev.Payload)
		sp.SetError(err)
		sp.End()
		if err != nil {
			obs.M().OutboxPublished.Inc("error")
		} else {
			obs.M().OutboxPublished.Inc("ok")
		}
		return err
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
					r.log.WarnContext(ctx, "outbox relay error", "err", err)
					break
				}
				if n < r.Batch {
					break
				}
			}
		}
	}
}
