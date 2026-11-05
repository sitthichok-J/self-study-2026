package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)

func TestNewClaimValidation(t *testing.T) {
	cases := []struct {
		name, key, ref string
		amount         int64
		want           error
	}{
		{"ok", "k1", "FAKE-AB12", 1000, nil},
		{"no key", "", "FAKE-AB12", 1000, ErrIdempotencyRequired},
		{"key too long", strings.Repeat("k", 101), "FAKE-AB12", 1000, ErrInvalid},
		{"real-looking ref", "k", "1234567890123", 1000, ErrInvalid},
		{"lowercase ref", "k", "fake-ab12", 1000, ErrInvalid},
		{"short ref", "k", "FAKE-A", 1000, ErrInvalid},
		{"zero amount", "k", "FAKE-AB12", 0, ErrInvalid},
		{"negative amount", "k", "FAKE-AB12", -5, ErrInvalid},
		{"too big", "k", "FAKE-AB12", MaxAmount + 1, ErrInvalid},
		{"max ok", "k", "FAKE-AB12", MaxAmount, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NewClaim("id", "alice", c.key, c.ref, c.amount, t0)
			if !errors.Is(err, c.want) {
				t.Fatalf("err=%v want %v", err, c.want)
			}
			if err == nil && got.Status != StatusPending {
				t.Fatalf("new claim must be PENDING, got %s", got.Status)
			}
		})
	}
}

func TestTransitions(t *testing.T) {
	ok := [][2]Status{{StatusPending, StatusApproved}, {StatusPending, StatusRejected}}
	for _, p := range ok {
		if err := CanTransition(p[0], p[1]); err != nil {
			t.Errorf("%v should be allowed: %v", p, err)
		}
	}
	bad := [][2]Status{
		{StatusApproved, StatusRejected}, {StatusRejected, StatusApproved},
		{StatusApproved, StatusPending}, {StatusPending, StatusPending}, {StatusApproved, StatusApproved},
	}
	for _, p := range bad {
		if err := CanTransition(p[0], p[1]); !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("%v should be rejected, got %v", p, err)
		}
	}
}

func TestEvaluate(t *testing.T) {
	mk := func(ref string, amt int64) Claim { return Claim{PatientRef: ref, Amount: amt} }
	d, err := Evaluate(mk("FAKE-A1B", ReviewLimit))
	if err != nil || d.Status != StatusApproved || d.Reason != "ok" {
		t.Fatalf("at review limit should approve: %v %v", d, err)
	}
	for _, amt := range []int64{ReviewLimit + 1, OverLimit} { // Day 13: the middle band waits for a person
		d, err = Evaluate(mk("FAKE-A1B", amt))
		if err != nil || d.Status != StatusPending || d.Reason != "needs_review" {
			t.Fatalf("amount %d should wait for review: %v %v", amt, d, err)
		}
	}
	d, _ = Evaluate(mk("FAKE-A1B", OverLimit+1))
	if d.Status != StatusRejected || d.Reason != "over_limit" {
		t.Fatalf("over limit should reject: %v", d)
	}
	if _, err := Evaluate(mk(PoisonRef, 10)); !errors.Is(err, ErrTransient) {
		t.Fatalf("poison must be transient error, got %v", err)
	}
}

func TestRequestHashStable(t *testing.T) {
	if RequestHash("FAKE-A1B", 5) != RequestHash("FAKE-A1B", 5) {
		t.Fatal("hash not stable")
	}
	if RequestHash("FAKE-A1B", 5) == RequestHash("FAKE-A1B", 6) {
		t.Fatal("different amount must change hash")
	}
	if RequestHash("FAKE-A1", 56) == RequestHash("FAKE-A15", 6) {
		t.Fatal("separator must prevent ambiguity")
	}
}

func TestUUID(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		u := NewUUID()
		if !IsUUID(u) || seen[u] {
			t.Fatalf("bad or duplicate uuid %q", u)
		}
		seen[u] = true
	}
	for _, s := range []string{"", "abc", "not-a-uuid", "00000000-0000-0000-0000-000000000000"} {
		if IsUUID(s) {
			t.Errorf("%q should not be a uuid", s)
		}
	}
}

func TestParseStatus(t *testing.T) {
	if s, err := ParseStatus("APPROVED"); err != nil || s != StatusApproved {
		t.Fatal("APPROVED should parse")
	}
	if _, err := ParseStatus("approved"); !errors.Is(err, ErrInvalid) {
		t.Fatal("lowercase must be invalid")
	}
}
