package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/auth"
)

// Repo also implements auth.Store: users and rotating refresh tokens.
var _ auth.Store = (*Repo)(nil)

func (r *Repo) GetUser(ctx context.Context, username string) (auth.User, error) {
	var u auth.User
	err := r.pool.QueryRow(ctx,
		`SELECT username, password_hash, role FROM users WHERE username = $1`, username,
	).Scan(&u.Username, &u.PasswordHash, &u.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.User{}, auth.ErrNoUser
	}
	return u, err
}

func (r *Repo) UpsertUser(ctx context.Context, u auth.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)
		ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role`,
		u.Username, u.PasswordHash, u.Role)
	return err
}

func (r *Repo) SaveRefresh(ctx context.Context, rec auth.RefreshRecord) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (token_hash, family_id, username, expires_at) VALUES ($1, $2, $3, $4)`,
		rec.Hash, rec.Family, rec.Username, rec.Expires)
	return err
}

// UseRefresh marks a valid token as used in ONE statement, so two requests
// carrying the same token cannot both win: the second UPDATE re-checks
// used_at after the first commits and matches no row.
func (r *Repo) UseRefresh(ctx context.Context, hash string, now time.Time) (auth.RefreshRecord, auth.UseStatus, error) {
	var rec auth.RefreshRecord
	err := r.pool.QueryRow(ctx, `
		UPDATE refresh_tokens SET used_at = $2
		WHERE token_hash = $1 AND used_at IS NULL AND revoked_at IS NULL AND expires_at > $2
		RETURNING token_hash, family_id, username, expires_at`, hash, now,
	).Scan(&rec.Hash, &rec.Family, &rec.Username, &rec.Expires)
	if err == nil {
		return rec, auth.UseOK, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return auth.RefreshRecord{}, auth.UseUnknown, err
	}
	// Not usable: find out why (same order as the in-memory store).
	var used, revoked *time.Time
	err = r.pool.QueryRow(ctx, `
		SELECT token_hash, family_id, username, expires_at, used_at, revoked_at
		FROM refresh_tokens WHERE token_hash = $1`, hash,
	).Scan(&rec.Hash, &rec.Family, &rec.Username, &rec.Expires, &used, &revoked)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.RefreshRecord{}, auth.UseUnknown, nil
	}
	if err != nil {
		return auth.RefreshRecord{}, auth.UseUnknown, err
	}
	switch {
	case used != nil:
		return rec, auth.UseReused, nil
	case revoked != nil:
		return rec, auth.UseRevoked, nil
	default:
		return rec, auth.UseExpired, nil
	}
}

func (r *Repo) FamilyOf(ctx context.Context, hash string) (string, bool, error) {
	var family string
	err := r.pool.QueryRow(ctx, `SELECT family_id FROM refresh_tokens WHERE token_hash = $1`, hash).Scan(&family)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return family, err == nil, err
}

func (r *Repo) RevokeFamily(ctx context.Context, family string, now time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = $2 WHERE family_id = $1 AND revoked_at IS NULL`, family, now)
	return err
}
