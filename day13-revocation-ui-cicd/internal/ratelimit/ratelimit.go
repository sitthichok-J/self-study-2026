// Package ratelimit is a per-key fixed-window counter. Kong already limits per IP; this one limits per
// signed-in user, so one account cannot hide behind many IPs (and many users behind one office IP are not blocked together).
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Result of one Allow call.
type Result struct {
	Allowed    bool
	Remaining  int           // requests left in this window (0 when blocked)
	RetryAfter time.Duration // until the window resets; only meaningful when blocked
}

type Limiter interface {
	// Allow counts one request for key. An error means the counter could not be read.
	Allow(ctx context.Context, key string, limit int, window time.Duration) (Result, error)
}

// Mem is the in-memory Limiter for tests.
type Mem struct {
	mu   sync.Mutex
	Now  func() time.Time
	Fail error
	m    map[string]*bucket
}

type bucket struct {
	n     int
	reset time.Time
}

func NewMem() *Mem { return &Mem{Now: time.Now, m: map[string]*bucket{}} }

func (l *Mem) Allow(_ context.Context, key string, limit int, window time.Duration) (Result, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.Fail != nil {
		return Result{}, l.Fail
	}
	now := l.Now()
	b := l.m[key]
	if b == nil || !now.Before(b.reset) {
		b = &bucket{reset: now.Add(window)}
		l.m[key] = b
	}
	b.n++
	return Decide(b.n, limit, b.reset.Sub(now)), nil
}

// Decide turns the counter value into a Result. Used by every implementation so they agree.
func Decide(count, limit int, untilReset time.Duration) Result {
	if count > limit {
		return Result{Allowed: false, RetryAfter: untilReset}
	}
	return Result{Allowed: true, Remaining: limit - count}
}
