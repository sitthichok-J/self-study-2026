// Package port declares what the service needs from the outside world.
package port

import (
	"context"
	"time"

	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/domain"
)

const (
	TopicSubmitted = "claims.submitted"
	TopicDLQ       = "claims.submitted.dlq"
)

// OutboxEvent is written in the same DB transaction as the claim.
type OutboxEvent struct {
	ID      int64
	Topic   string
	Key     string
	Payload []byte
}

type ClaimRepository interface {
	// Create stores claim + outbox event atomically. Idempotency keys are
	// scoped to the owner: two users may use the same key string. If the
	// owner's key already exists with the same request hash it returns the existing claim
	// and created=false; with a different hash it returns ErrIdempotencyConflict.
	Create(ctx context.Context, c domain.Claim, ev OutboxEvent) (claim domain.Claim, created bool, err error)
	Get(ctx context.Context, id string) (domain.Claim, error)
	// List returns newest first. owner == "" means every owner (staff only).
	List(ctx context.Context, owner string, status *domain.Status, limit, offset int) ([]domain.Claim, error)
	// UpdateStatus moves from -> to only if the row is still in `from`.
	// updated=false means someone else already moved it (not an error).
	UpdateStatus(ctx context.Context, id string, from, to domain.Status, reason string) (updated bool, err error)
	// ProcessOutbox locks up to limit unsent events, calls publish in order and
	// marks as sent the ones that succeeded. It stops at the first failure.
	ProcessOutbox(ctx context.Context, limit int, publish func(OutboxEvent) error) (sent int, err error)
}

type Message struct {
	Topic string
	Key   string
	Value []byte
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, value []byte) error
}

// Cache is best effort: callers must treat errors as a miss.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}
