// Package redis implements port.Cache (works with Redis or Valkey).
package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/ratelimit"
)

type Cache struct{ c *goredis.Client }

func New(addr string) *Cache {
	return &Cache{c: goredis.NewClient(&goredis.Options{
		Addr: addr, DialTimeout: time.Second, ReadTimeout: 500 * time.Millisecond, WriteTimeout: 500 * time.Millisecond,
	})}
}

func (c *Cache) Ping(ctx context.Context) error { return c.c.Ping(ctx).Err() }
func (c *Cache) Close() error                   { return c.c.Close() }

func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	b, err := c.c.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func (c *Cache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return c.c.Set(ctx, key, val, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	return c.c.Del(ctx, keys...).Err()
}

// ---- Day 13: access-token blocklist and per-user rate limit, on the same Valkey ----

const (
	revTokenPrefix = "rev:jti:"
	revUserPrefix  = "rev:user:"
	rlPrefix       = "rl:"
)

// Revoker implements auth.Revoker. Entries expire on their own once the token could no longer be accepted anyway.
type Revoker struct{ c *goredis.Client }

func (c *Cache) Revoker() *Revoker { return &Revoker{c: c.c} }

func (r *Revoker) RevokeToken(ctx context.Context, jti string, until time.Time) error {
	ttl := time.Until(until)
	if ttl <= 0 {
		return nil // already unusable
	}
	return r.c.Set(ctx, revTokenPrefix+jti, "1", ttl).Err()
}

func (r *Revoker) RevokeUser(ctx context.Context, sub string, at time.Time, ttl time.Duration) error {
	return r.c.Set(ctx, revUserPrefix+sub, strconv.FormatInt(at.Unix(), 10), ttl).Err()
}

func (r *Revoker) Check(ctx context.Context, p auth.Principal) (bool, error) {
	pipe := r.c.Pipeline()
	ex := pipe.Exists(ctx, revTokenPrefix+p.JTI)
	cut := pipe.Get(ctx, revUserPrefix+p.Subject)
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, goredis.Nil) {
		return false, err
	}
	if n, _ := ex.Result(); n > 0 {
		return true, nil
	}
	v, err := cut.Result()
	if errors.Is(err, goredis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	at, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return true, nil // unreadable cut value: block rather than let a possibly stolen token through
	}
	return auth.Cut(p, time.Unix(at, 0)), nil
}

// Limiter implements ratelimit.Limiter with INCR + PEXPIRE (fixed window).
type Limiter struct{ c *goredis.Client }

func (c *Cache) Limiter() *Limiter { return &Limiter{c: c.c} }

func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (ratelimit.Result, error) {
	k := rlPrefix + key
	pipe := l.c.Pipeline()
	n := pipe.Incr(ctx, k)
	ttl := pipe.PTTL(ctx, k)
	if _, err := pipe.Exec(ctx); err != nil {
		return ratelimit.Result{}, err
	}
	left := ttl.Val()
	if left < 0 { // first hit, or a key that lost its expiry: (re)start the window
		if err := l.c.PExpire(ctx, k, window).Err(); err != nil {
			return ratelimit.Result{}, err
		}
		left = window
	}
	return ratelimit.Decide(int(n.Val()), limit, left), nil
}
