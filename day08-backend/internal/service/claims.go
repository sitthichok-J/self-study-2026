// Package service holds the use cases. It only talks to port interfaces.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/port"
)

const (
	ClaimTTL = 60 * time.Second
	ListTTL  = 10 * time.Second

	DefaultLimit = 20
	MaxLimit     = 100
)

type Claims struct {
	repo  port.ClaimRepository
	cache port.Cache
	log   *slog.Logger
	now   func() time.Time
	newID func() string
}

func NewClaims(repo port.ClaimRepository, cache port.Cache, log *slog.Logger) *Claims {
	if log == nil {
		log = slog.Default()
	}
	return &Claims{repo: repo, cache: cache, log: log, now: time.Now, newID: domain.NewUUID}
}

// WithClock / WithIDs are for tests.
func (s *Claims) WithClock(f func() time.Time) *Claims { s.now = f; return s }
func (s *Claims) WithIDs(f func() string) *Claims      { s.newID = f; return s }

type SubmitInput struct {
	IdempotencyKey string
	PatientRef     string
	Amount         int64
}

// SubmitEvent is the payload of claims.submitted. It carries only the id:
// the consumer reads the current row, so it never acts on stale data.
type SubmitEvent struct {
	Event      string    `json:"event"`
	Version    int       `json:"version"`
	ClaimID    string    `json:"claim_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (s *Claims) Submit(ctx context.Context, in SubmitInput) (domain.Claim, bool, error) {
	now := s.now()
	c, err := domain.NewClaim(s.newID(), in.IdempotencyKey, in.PatientRef, in.Amount, now)
	if err != nil {
		return domain.Claim{}, false, err
	}
	payload, _ := json.Marshal(SubmitEvent{Event: port.TopicSubmitted, Version: 1, ClaimID: c.ID, OccurredAt: now.UTC()})
	ev := port.OutboxEvent{Topic: port.TopicSubmitted, Key: c.ID, Payload: payload}

	saved, created, err := s.repo.Create(ctx, c, ev)
	if err != nil {
		return domain.Claim{}, false, err
	}
	return saved, created, nil
}

func claimKey(id string) string { return "claim:" + id }

func (s *Claims) Get(ctx context.Context, id string) (domain.Claim, error) {
	if !domain.IsUUID(id) {
		return domain.Claim{}, fmt.Errorf("%w: id is not a uuid", domain.ErrInvalid)
	}
	var c domain.Claim
	if s.cacheGet(ctx, claimKey(id), &c) {
		return c, nil
	}
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Claim{}, err
	}
	s.cacheSet(ctx, claimKey(id), c, ClaimTTL)
	return c, nil
}

func (s *Claims) List(ctx context.Context, status *domain.Status, limit, offset int) ([]domain.Claim, error) {
	if limit == 0 {
		limit = DefaultLimit
	}
	if limit < 0 || limit > MaxLimit || offset < 0 {
		return nil, fmt.Errorf("%w: limit must be 1..%d and offset >= 0", domain.ErrInvalid, MaxLimit)
	}
	st := "all"
	if status != nil {
		st = string(*status)
	}
	key := fmt.Sprintf("claims:list:%s:%d:%d", st, limit, offset)
	var out []domain.Claim
	if s.cacheGet(ctx, key, &out) {
		return out, nil
	}
	out, err := s.repo.List(ctx, status, limit, offset)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []domain.Claim{}
	}
	s.cacheSet(ctx, key, out, ListTTL) // lists may be stale up to ListTTL (documented)
	return out, nil
}

// Process is the worker use case. It is idempotent: a claim that already left
// PENDING is skipped, so redelivery from Kafka (at-least-once) is harmless.
func (s *Claims) Process(ctx context.Context, id string) error {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if c.Status != domain.StatusPending {
		s.log.Info("skip already processed", "claim_id", id, "status", c.Status)
		return nil
	}
	d, err := domain.Evaluate(c)
	if err != nil {
		return err
	}
	if err := domain.CanTransition(c.Status, d.Status); err != nil {
		return err
	}
	updated, err := s.repo.UpdateStatus(ctx, id, domain.StatusPending, d.Status, d.Reason)
	if err != nil {
		return err
	}
	if !updated {
		s.log.Info("lost race, already processed", "claim_id", id)
	}
	if s.cache != nil {
		if err := s.cache.Delete(ctx, claimKey(id)); err != nil {
			s.log.Warn("cache delete failed", "err", err)
		}
	}
	return nil
}

func (s *Claims) cacheGet(ctx context.Context, key string, dst any) bool {
	if s.cache == nil {
		return false
	}
	b, ok, err := s.cache.Get(ctx, key)
	if err != nil {
		s.log.Warn("cache get failed", "key", key, "err", err)
		return false
	}
	if !ok {
		return false
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return false
	}
	return true
}

func (s *Claims) cacheSet(ctx context.Context, key string, v any, ttl time.Duration) {
	if s.cache == nil {
		return
	}
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	if err := s.cache.Set(ctx, key, b, ttl); err != nil {
		s.log.Warn("cache set failed", "key", key, "err", err)
	}
}

// IsPermanent: errors that retrying cannot fix.
func IsPermanent(err error) bool {
	return errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalid) ||
		errors.Is(err, domain.ErrInvalidTransition)
}
