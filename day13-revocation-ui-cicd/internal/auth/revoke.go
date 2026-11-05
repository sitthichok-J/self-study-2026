package auth

import (
	"context"
	"sync"
	"time"
)

// Revoker is the blocklist for access tokens. A signed JWT stays valid until exp by itself, so
// logout and "this refresh token was stolen" need a server-side list to cut it short.
//
//   - RevokeToken blocks one token (by jti) until its own exp: after that the list entry is useless and may expire.
//   - RevokeUser blocks every token of one user issued at or before `at` (logout everywhere,
//     refresh-token reuse). `ttl` only needs to cover the longest access token lifetime.
//   - Check says whether a verified Principal is blocked. An error means "could not tell":
//     the caller must fail closed.
type Revoker interface {
	RevokeToken(ctx context.Context, jti string, until time.Time) error
	RevokeUser(ctx context.Context, sub string, at time.Time, ttl time.Duration) error
	Check(ctx context.Context, p Principal) (bool, error)
}

// MemRevoker is the in-memory Revoker for tests.
type MemRevoker struct {
	mu    sync.Mutex
	Now   func() time.Time
	jtis  map[string]time.Time // jti -> until
	users map[string]userCut   // sub -> cut
	Fail  error                // set to make every call fail (tests)
}

type userCut struct {
	at    time.Time
	until time.Time
}

func NewMemRevoker() *MemRevoker {
	return &MemRevoker{Now: time.Now, jtis: map[string]time.Time{}, users: map[string]userCut{}}
}

func (m *MemRevoker) RevokeToken(_ context.Context, jti string, until time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Fail != nil {
		return m.Fail
	}
	m.jtis[jti] = until
	return nil
}

func (m *MemRevoker) RevokeUser(_ context.Context, sub string, at time.Time, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Fail != nil {
		return m.Fail
	}
	m.users[sub] = userCut{at: at, until: m.Now().Add(ttl)}
	return nil
}

func (m *MemRevoker) Check(_ context.Context, p Principal) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Fail != nil {
		return false, m.Fail
	}
	now := m.Now()
	if until, ok := m.jtis[p.JTI]; ok && now.Before(until) {
		return true, nil
	}
	if c, ok := m.users[p.Subject]; ok && now.Before(c.until) && Cut(p, c.at) {
		return true, nil
	}
	return false, nil
}

// Cut reports whether a token falls before a user-wide cut. Whole seconds, because iat has no finer
// resolution: a token issued in the same second as the cut is blocked too (safe side). A token with no iat is blocked.
func Cut(p Principal, at time.Time) bool {
	return p.IssuedAt.IsZero() || p.IssuedAt.Unix() <= at.Unix()
}

// RevokeUntil is the time a blocklist entry for p may expire: its exp plus the verifier leeway.
func (v *Verifier) RevokeUntil(p Principal) time.Time { return p.Expires.Add(v.Leeway) }

// MaxTokenLife is how long any access token of this verifier can still be accepted (cap + leeway).
func (v *Verifier) MaxTokenLife() time.Duration { return v.MaxTTL + v.Leeway }
