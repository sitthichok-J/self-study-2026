package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth"
)

// These tests need a real Valkey/Redis. They are skipped unless REDIS_TEST_ADDR is set
// (scripts/run-all.ps1 sets it and runs them on the compose network against the real valkey container).
func testCache(t *testing.T) *Cache {
	t.Helper()
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR not set")
	}
	c := New(addr)
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("valkey at %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func uniq(prefix string) string { return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()) }

func TestRevokerTokenAndExpiry(t *testing.T) {
	r := testCache(t).Revoker()
	ctx := context.Background()
	p := auth.Principal{Subject: uniq("u"), JTI: uniq("jti"), IssuedAt: time.Now()}
	if bad, err := r.Check(ctx, p); bad || err != nil {
		t.Fatalf("fresh: %v %v", bad, err)
	}
	if err := r.RevokeToken(ctx, p.JTI, time.Now().Add(1500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	if bad, _ := r.Check(ctx, p); !bad {
		t.Fatal("must be blocked")
	}
	if err := r.RevokeToken(ctx, "already-past", time.Now().Add(-time.Second)); err != nil {
		t.Fatalf("a past expiry is a no-op, not an error: %v", err)
	}
	time.Sleep(2 * time.Second)
	if bad, _ := r.Check(ctx, p); bad {
		t.Fatal("entry must expire by itself")
	}
}

func TestRevokerUserCut(t *testing.T) {
	r := testCache(t).Revoker()
	ctx := context.Background()
	sub := uniq("u")
	now := time.Now()
	old := auth.Principal{Subject: sub, JTI: uniq("a"), IssuedAt: now.Add(-time.Minute)}
	newer := auth.Principal{Subject: sub, JTI: uniq("b"), IssuedAt: now.Add(2 * time.Second)}
	other := auth.Principal{Subject: uniq("other"), JTI: uniq("c"), IssuedAt: now.Add(-time.Minute)}
	if err := r.RevokeUser(ctx, sub, now, time.Minute); err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		p    auth.Principal
		want bool
	}{"old": {old, true}, "newer": {newer, false}, "other user": {other, false}} {
		if got, err := r.Check(ctx, tc.p); err != nil || got != tc.want {
			t.Errorf("%s: blocked=%v err=%v want %v", name, got, err, tc.want)
		}
	}
}

func TestLimiterWindow(t *testing.T) {
	l := testCache(t).Limiter()
	ctx := context.Background()
	key := uniq("user")
	for i := 1; i <= 3; i++ {
		r, err := l.Allow(ctx, key, 3, 2*time.Second)
		if err != nil || !r.Allowed || r.Remaining != 3-i {
			t.Fatalf("call %d: %+v %v", i, r, err)
		}
	}
	r, err := l.Allow(ctx, key, 3, 2*time.Second)
	if err != nil || r.Allowed || r.RetryAfter <= 0 || r.RetryAfter > 2*time.Second {
		t.Fatalf("4th call must be blocked with a reset time: %+v %v", r, err)
	}
	if r, _ := l.Allow(ctx, uniq("other"), 3, 2*time.Second); !r.Allowed {
		t.Fatal("another key has its own counter")
	}
	time.Sleep(2200 * time.Millisecond)
	if r, _ := l.Allow(ctx, key, 3, 2*time.Second); !r.Allowed || r.Remaining != 2 {
		t.Fatalf("new window: %+v", r)
	}
}
