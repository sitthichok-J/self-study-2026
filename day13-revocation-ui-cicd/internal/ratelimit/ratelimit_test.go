package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemWindow(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	l := NewMem()
	l.Now = func() time.Time { return now }
	ctx := context.Background()
	for i := 1; i <= 3; i++ {
		r, err := l.Allow(ctx, "alice", 3, time.Minute)
		if err != nil || !r.Allowed || r.Remaining != 3-i {
			t.Fatalf("call %d: %+v %v", i, r, err)
		}
	}
	now = now.Add(20 * time.Second)
	r, _ := l.Allow(ctx, "alice", 3, time.Minute)
	if r.Allowed || r.RetryAfter != 40*time.Second {
		t.Fatalf("4th call must be blocked with 40s left, got %+v", r)
	}
	// another key is unaffected
	if r, _ := l.Allow(ctx, "bob", 3, time.Minute); !r.Allowed {
		t.Fatal("bob must not share alice's counter")
	}
	// the window resets exactly at its end
	now = now.Add(40 * time.Second)
	if r, _ := l.Allow(ctx, "alice", 3, time.Minute); !r.Allowed || r.Remaining != 2 {
		t.Fatalf("new window must start fresh, got %+v", r)
	}
}

func TestMemError(t *testing.T) {
	l := NewMem()
	l.Fail = errors.New("down")
	if _, err := l.Allow(context.Background(), "x", 1, time.Minute); err == nil {
		t.Fatal("error must be returned")
	}
}

func TestDecideBoundary(t *testing.T) {
	cases := []struct {
		count, limit int
		allowed      bool
		remaining    int
	}{{1, 2, true, 1}, {2, 2, true, 0}, {3, 2, false, 0}}
	for _, c := range cases {
		r := Decide(c.count, c.limit, time.Second)
		if r.Allowed != c.allowed || r.Remaining != c.remaining {
			t.Errorf("Decide(%d,%d)=%+v", c.count, c.limit, r)
		}
	}
}
