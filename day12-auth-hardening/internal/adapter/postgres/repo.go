// Package postgres implements port.ClaimRepository with pgx v5.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/port"
)

type Repo struct{ pool *pgxpool.Pool }

func New(ctx context.Context, dsn string) (*Repo, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Repo{pool: pool}, nil
}

func (r *Repo) Close()                         { r.pool.Close() }
func (r *Repo) Ping(ctx context.Context) error { return r.pool.Ping(ctx) }

// PoolStats feeds the db_pool_* gauges (USE: utilization/saturation of the pool).
func (r *Repo) PoolStats() (acquired, idle, total, max int32) {
	s := r.pool.Stat()
	return s.AcquiredConns(), s.IdleConns(), s.TotalConns(), s.MaxConns()
}

const cols = `id, owner_id, patient_ref, amount, status, decision_reason, idempotency_key, request_hash, created_at, updated_at`

type scanner interface{ Scan(dest ...any) error }

func scan(s scanner) (domain.Claim, error) {
	var c domain.Claim
	var st string
	err := s.Scan(&c.ID, &c.OwnerID, &c.PatientRef, &c.Amount, &st, &c.DecisionReason, &c.IdempotencyKey, &c.RequestHash, &c.CreatedAt, &c.UpdatedAt)
	c.Status = domain.Status(st)
	return c, err
}

func (r *Repo) Create(ctx context.Context, c domain.Claim, ev port.OutboxEvent) (domain.Claim, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Claim{}, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	row := tx.QueryRow(ctx, `
		INSERT INTO claims (id, owner_id, patient_ref, amount, status, idempotency_key, request_hash, created_at, updated_at)
		VALUES ($1,$8,$2,$3,$4,$5,$6,$7,$7)
		ON CONFLICT (owner_id, idempotency_key) DO NOTHING
		RETURNING `+cols,
		c.ID, c.PatientRef, c.Amount, string(c.Status), c.IdempotencyKey, c.RequestHash, c.CreatedAt, c.OwnerID)
	saved, err := scan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		// key already used: replay or conflict
		existing, err := scan(tx.QueryRow(ctx, `SELECT `+cols+` FROM claims WHERE owner_id = $1 AND idempotency_key = $2`, c.OwnerID, c.IdempotencyKey))
		if err != nil {
			return domain.Claim{}, false, fmt.Errorf("load existing by key: %w", err)
		}
		if existing.RequestHash != c.RequestHash {
			return domain.Claim{}, false, domain.ErrIdempotencyConflict
		}
		return existing, false, nil
	}
	if err != nil {
		return domain.Claim{}, false, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO outbox (topic, key, payload) VALUES ($1,$2,$3)`,
		ev.Topic, ev.Key, ev.Payload); err != nil {
		return domain.Claim{}, false, fmt.Errorf("insert outbox: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Claim{}, false, err
	}
	return saved, true, nil
}

func (r *Repo) Get(ctx context.Context, id string) (domain.Claim, error) {
	c, err := scan(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM claims WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Claim{}, domain.ErrNotFound
	}
	return c, err
}

func (r *Repo) List(ctx context.Context, owner string, status *domain.Status, limit, offset int) ([]domain.Claim, error) {
	var st any // nil => no filter
	if status != nil {
		st = string(*status)
	}
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+` FROM claims
		WHERE ($1::text IS NULL OR status = $1::text)
		  AND ($4::text = '' OR owner_id = $4::text)
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`, st, limit, offset, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Claim
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repo) UpdateStatus(ctx context.Context, id string, from, to domain.Status, reason string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE claims SET status = $3, decision_reason = $4, updated_at = now()
		WHERE id = $1 AND status = $2`, id, string(from), string(to), reason)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// ProcessOutbox: lock unsent rows (SKIP LOCKED so several relays can run),
// publish in id order, mark the successful ones, commit.
func (r *Repo) ProcessOutbox(ctx context.Context, limit int, publish func(port.OutboxEvent) error) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	rows, err := tx.Query(ctx, `
		SELECT id, topic, key, payload FROM outbox
		WHERE sent_at IS NULL ORDER BY id LIMIT $1
		FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return 0, err
	}
	var evs []port.OutboxEvent
	for rows.Next() {
		var e port.OutboxEvent
		if err := rows.Scan(&e.ID, &e.Topic, &e.Key, &e.Payload); err != nil {
			rows.Close()
			return 0, err
		}
		evs = append(evs, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	var sentIDs []int64
	var pubErr error
	for _, e := range evs {
		if pubErr = publish(e); pubErr != nil {
			break
		}
		sentIDs = append(sentIDs, e.ID)
	}
	if len(sentIDs) > 0 {
		if _, err := tx.Exec(ctx, `UPDATE outbox SET sent_at = now() WHERE id = ANY($1)`, sentIDs); err != nil {
			return 0, err
		}
		if err := tx.Commit(ctx); err != nil {
			return 0, err // events were published but not marked: re-sent later (at-least-once)
		}
	}
	return len(sentIDs), pubErr
}
