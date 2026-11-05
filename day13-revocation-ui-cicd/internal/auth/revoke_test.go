package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

var bg = context.Background()

func TestMintedTokenCarriesUniqueJTI(t *testing.T) {
	v, _ := newV(t)
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		tok, _ := v.Mint("alice", "", 10*time.Minute)
		p, err := v.Verify(tok)
		if err != nil || p.JTI == "" {
			t.Fatalf("verify: %v %+v", err, p)
		}
		if seen[p.JTI] {
			t.Fatal("jti repeated")
		}
		seen[p.JTI] = true
		if !p.Expires.Equal(v.Now().Add(10*time.Minute).Truncate(time.Second)) || p.IssuedAt.IsZero() {
			t.Fatalf("principal times wrong: %+v", p)
		}
	}
}

func TestJTIIsRequiredAndChecked(t *testing.T) {
	v, _ := newV(t)
	for name, j := range map[string]any{"missing": nil, "short": "abc", "bad chars": "0123456789abcdef!!!!", "too long": "0123456789012345678901234567890123456789012345678901234567890123456789"} {
		tok := craft(v, hs, payload(v, map[string]any{"jti": j}))
		if _, err := v.Verify(tok); !errors.Is(err, ErrClaims) {
			t.Errorf("%s jti: want ErrClaims, got %v", name, err)
		}
	}
}

func TestReasonRevoked(t *testing.T) {
	if Reason(ErrRevoked) != "revoked" {
		t.Fatal("reason label")
	}
	found := false
	for _, r := range Reasons {
		found = found || r == "revoked"
	}
	if !found {
		t.Fatal("revoked must be in Reasons so its metric series exists")
	}
}

func TestRevokeUntilAndMaxTokenLife(t *testing.T) {
	v, _ := newV(t)
	tok, _ := v.Mint("alice", "", 10*time.Minute)
	p, _ := v.Verify(tok)
	if got := v.RevokeUntil(p); !got.Equal(p.Expires.Add(30 * time.Second)) {
		t.Fatalf("until %v", got)
	}
	if v.MaxTokenLife() != time.Hour+30*time.Second {
		t.Fatalf("max life %v", v.MaxTokenLife())
	}
}

func TestMemRevokerToken(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	r := NewMemRevoker()
	r.Now = func() time.Time { return now }
	p := Principal{Subject: "alice", JTI: "jti-aaaaaaaaaaaaaaaa", IssuedAt: now}
	if bad, err := r.Check(bg, p); bad || err != nil {
		t.Fatalf("fresh token must pass: %v %v", bad, err)
	}
	_ = r.RevokeToken(bg, p.JTI, now.Add(time.Minute))
	if bad, _ := r.Check(bg, p); !bad {
		t.Fatal("revoked token must be blocked")
	}
	if bad, _ := r.Check(bg, Principal{Subject: "alice", JTI: "jti-bbbbbbbbbbbbbbbb", IssuedAt: now}); bad {
		t.Fatal("another token of the same user must not be blocked")
	}
	now = now.Add(time.Minute) // entry reaches its end: the token is expired by then anyway
	if bad, _ := r.Check(bg, p); bad {
		t.Fatal("entry must lapse at its end")
	}
}

func TestMemRevokerUserCut(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	r := NewMemRevoker()
	r.Now = func() time.Time { return now }
	old := Principal{Subject: "alice", JTI: "jti-aaaaaaaaaaaaaaaa", IssuedAt: now.Add(-time.Minute)}
	sameSec := Principal{Subject: "alice", JTI: "jti-bbbbbbbbbbbbbbbb", IssuedAt: now}
	later := Principal{Subject: "alice", JTI: "jti-cccccccccccccccc", IssuedAt: now.Add(time.Second)}
	noIat := Principal{Subject: "alice", JTI: "jti-dddddddddddddddd"}
	other := Principal{Subject: "bob", JTI: "jti-eeeeeeeeeeeeeeee", IssuedAt: now.Add(-time.Minute)}
	_ = r.RevokeUser(bg, "alice", now, time.Hour)
	for name, tc := range map[string]struct {
		p    Principal
		want bool
	}{"issued before": {old, true}, "same second": {sameSec, true}, "issued after": {later, false}, "no iat": {noIat, true}, "other user": {other, false}} {
		if got, _ := r.Check(bg, tc.p); got != tc.want {
			t.Errorf("%s: blocked=%v want %v", name, got, tc.want)
		}
	}
	now = now.Add(2 * time.Hour) // cut itself expires after ttl
	if got, _ := r.Check(bg, old); got {
		t.Fatal("cut must lapse after its ttl")
	}
}

func TestMemRevokerFailure(t *testing.T) {
	r := NewMemRevoker()
	r.Fail = errors.New("down")
	if _, err := r.Check(bg, Principal{}); err == nil {
		t.Fatal("Check must return the error")
	}
	if r.RevokeToken(bg, "x", time.Now().Add(time.Hour)) == nil || r.RevokeUser(bg, "x", time.Now(), time.Hour) == nil {
		t.Fatal("revoke calls must return the error")
	}
}

func TestRefreshReuseCutsAccessTokens(t *testing.T) {
	e := setup(t)
	rv := NewMemRevoker()
	rv.Now = e.clk.now
	e.svc.SetRevoker(rv)
	t1, _ := e.svc.Login(bg, "alice", alicePw)
	other, _ := e.svc.Login(bg, "bob", "bob-fake-password-22")
	e.clk.t = e.clk.t.Add(time.Second)
	if _, err := e.svc.Refresh(bg, t1.Refresh); err != nil { // legit rotation
		t.Fatal(err)
	}
	pa, _ := e.ver.Verify(t1.Access)
	if bad, _ := rv.Check(bg, pa); bad {
		t.Fatal("normal refresh must not revoke anything")
	}
	e.clk.t = e.clk.t.Add(time.Second)
	if _, err := e.svc.Refresh(bg, t1.Refresh); !errors.Is(err, ErrRefreshReused) { // replay of the copied token
		t.Fatalf("want reuse, got %v", err)
	}
	if bad, _ := rv.Check(bg, pa); !bad {
		t.Fatal("after reuse, alice's earlier access token must be blocked")
	}
	pb, _ := e.ver.Verify(other.Access)
	if bad, _ := rv.Check(bg, pb); bad {
		t.Fatal("bob must not be affected")
	}
}

func TestRefreshReuseFailsWhenRevokerFails(t *testing.T) {
	e := setup(t)
	rv := NewMemRevoker()
	e.svc.SetRevoker(rv)
	t1, _ := e.svc.Login(bg, "alice", alicePw)
	_, _ = e.svc.Refresh(bg, t1.Refresh)
	rv.Fail = errors.New("down")
	if _, err := e.svc.Refresh(bg, t1.Refresh); err == nil || errors.Is(err, ErrRefreshReused) {
		t.Fatalf("a failed revoke must surface as an error, got %v", err)
	}
}
