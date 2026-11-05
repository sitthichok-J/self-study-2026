// Package service holds the use cases. It only talks to port interfaces.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/port"
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
	Owner          string // who is submitting; comes from the verified token, never from the body
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
	// Traceparent lets the worker continue the trace that created the claim.
	Traceparent string `json:"traceparent,omitempty"`
}

func (s *Claims) Submit(ctx context.Context, in SubmitInput) (domain.Claim, bool, error) {
	now := s.now()
	c, err := domain.NewClaim(s.newID(), in.Owner, in.IdempotencyKey, in.PatientRef, in.Amount, now)
	if err != nil {
		return domain.Claim{}, false, err
	}
	payload, _ := json.Marshal(SubmitEvent{
		Event: port.TopicSubmitted, Version: 1, ClaimID: c.ID, OccurredAt: now.UTC(),
		Traceparent: obs.TraceparentFromContext(ctx),
	})
	ev := port.OutboxEvent{Topic: port.TopicSubmitted, Key: c.ID, Payload: payload}

	ctx, sp := obs.Start(ctx, "repo.Create")
	saved, created, err := s.repo.Create(ctx, c, ev)
	sp.SetError(err)
	sp.End()
	if err != nil {
		return domain.Claim{}, false, err
	}
	if created {
		obs.M().ClaimsSubmitted.Inc("created")
	} else {
		obs.M().ClaimsSubmitted.Inc("replayed")
	}
	return saved, created, nil
}

func claimKey(id string) string { return "claim:" + id }

// Get returns the claim if the viewer owns it (or is staff). The check runs
// after the cache lookup too, so a cached copy never bypasses it. A claim that
// exists but belongs to someone else yields ErrForbidden; the HTTP layer
// answers 404 for both cases so ids cannot be probed.
func (s *Claims) Get(ctx context.Context, v domain.Viewer, id string) (domain.Claim, error) {
	if !domain.IsUUID(id) {
		return domain.Claim{}, fmt.Errorf("%w: id is not a uuid", domain.ErrInvalid)
	}
	c, err := s.load(ctx, id)
	if err != nil {
		return domain.Claim{}, err
	}
	if !v.Staff && (v.ID == "" || c.OwnerID != v.ID) {
		return domain.Claim{}, domain.ErrForbidden
	}
	return c, nil
}

func (s *Claims) load(ctx context.Context, id string) (domain.Claim, error) {
	var c domain.Claim
	if s.cacheGet(ctx, claimKey(id), &c) {
		return c, nil
	}
	ctx, sp := obs.Start(ctx, "repo.Get")
	c, err := s.repo.Get(ctx, id)
	sp.SetError(err)
	sp.End()
	if err != nil {
		return domain.Claim{}, err
	}
	s.cacheSet(ctx, claimKey(id), c, ClaimTTL)
	return c, nil
}

// List returns the viewer's own claims; staff see all. A viewer without an id
// that is not staff gets nothing (fail closed).
func (s *Claims) List(ctx context.Context, v domain.Viewer, status *domain.Status, limit, offset int) ([]domain.Claim, error) {
	owner, scope := v.ID, "u:"+v.ID
	if v.Staff {
		owner, scope = "", "all"
	} else if v.ID == "" {
		return nil, domain.ErrForbidden
	}
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
	key := fmt.Sprintf("claims:list:%s:%s:%d:%d", scope, st, limit, offset)
	var out []domain.Claim
	if s.cacheGet(ctx, key, &out) {
		return out, nil
	}
	ctx, sp := obs.Start(ctx, "repo.List")
	out, err := s.repo.List(ctx, owner, status, limit, offset)
	sp.SetError(err)
	sp.End()
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []domain.Claim{}
	}
	s.cacheSet(ctx, key, out, ListTTL) // lists may be stale up to ListTTL (documented)
	return out, nil
}

// Decide is the manual review: staff approve or reject a claim that is still PENDING.
// A claim someone else already decided gives ErrInvalidTransition (the HTTP layer answers 409).
func (s *Claims) Decide(ctx context.Context, v domain.Viewer, id string, to domain.Status, reason string) (domain.Claim, error) {
	if !v.Staff {
		return domain.Claim{}, domain.ErrForbidden
	}
	if !domain.IsUUID(id) {
		return domain.Claim{}, fmt.Errorf("%w: id is not a uuid", domain.ErrInvalid)
	}
	if err := domain.CanTransition(domain.StatusPending, to); err != nil {
		return domain.Claim{}, err
	}
	updated, err := s.repo.UpdateStatus(ctx, id, domain.StatusPending, to, reason)
	if err != nil {
		return domain.Claim{}, err
	}
	if s.cache != nil {
		if err := s.cache.Delete(ctx, claimKey(id)); err != nil {
			s.log.WarnContext(ctx, "cache delete failed", "err", err)
		}
	}
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Claim{}, err
	}
	if !updated {
		return domain.Claim{}, fmt.Errorf("%w: %s -> %s", domain.ErrInvalidTransition, c.Status, to)
	}
	obs.M().ClaimsDecided.Inc(string(to), reason)
	s.log.InfoContext(ctx, "claim decided manually", "claim_id", id, "status", to, "by", v.ID)
	return c, nil
}

// Process is the worker use case. It is idempotent: a claim that already left
// PENDING is skipped, so redelivery from Kafka (at-least-once) is harmless.
func (s *Claims) Process(ctx context.Context, id string) error {
	gctx, gsp := obs.Start(ctx, "repo.Get")
	c, err := s.repo.Get(gctx, id)
	gsp.SetError(err)
	gsp.End()
	if err != nil {
		return err
	}
	if c.Status != domain.StatusPending {
		s.log.InfoContext(ctx, "skip already processed", "claim_id", id, "status", c.Status)
		return nil
	}
	d, err := domain.Evaluate(c)
	if err != nil {
		return err
	}
	if d.Status != domain.StatusPending { // PENDING + reason = waiting for a person; no transition to check
		if err := domain.CanTransition(c.Status, d.Status); err != nil {
			return err
		}
	}
	uctx, usp := obs.Start(ctx, "repo.UpdateStatus")
	updated, err := s.repo.UpdateStatus(uctx, id, domain.StatusPending, d.Status, d.Reason)
	usp.SetError(err)
	usp.End()
	if err != nil {
		return err
	}
	if !updated {
		s.log.InfoContext(ctx, "lost race, already processed", "claim_id", id)
	} else {
		obs.M().ClaimsDecided.Inc(string(d.Status), d.Reason)
	}
	if s.cache != nil {
		if err := s.cache.Delete(ctx, claimKey(id)); err != nil {
			s.log.WarnContext(ctx, "cache delete failed", "err", err)
		}
	}
	return nil
}

func (s *Claims) cacheGet(ctx context.Context, key string, dst any) bool {
	if s.cache == nil {
		return false
	}
	cctx, sp := obs.Start(ctx, "cache.Get")
	b, ok, err := s.cache.Get(cctx, key)
	sp.SetAttr("cache.hit", ok)
	sp.SetError(err)
	sp.End()
	if err != nil {
		s.log.WarnContext(ctx, "cache get failed", "key", key, "err", err)
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
		s.log.WarnContext(ctx, "cache set failed", "key", key, "err", err)
	}
}

// IsPermanent: errors that retrying cannot fix.
func IsPermanent(err error) bool {
	return errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalid) ||
		errors.Is(err, domain.ErrInvalidTransition)
}
