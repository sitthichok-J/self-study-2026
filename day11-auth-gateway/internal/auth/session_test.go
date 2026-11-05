package auth

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

const (
	testSecret = "session-test-secret-session-test-0000"
	alicePw    = "alice-fake-password-1"
)

type env struct {
	svc   *Service
	store *MemStore
	clk   *clock
	ver   *Verifier
}

func setup(t *testing.T) *env {
	t.Helper()
	ver, err := NewVerifier([]byte(testSecret), "claim-tracker", "claims-api")
	if err != nil {
		t.Fatal(err)
	}
	clk := &clock{t: time.Date(2026, 11, 5, 9, 0, 0, 0, time.UTC)}
	ver.Now = clk.now
	st := NewMemStore()
	svc, err := NewService(st, ver, fast, 10*time.Minute, 12*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	svc.Now = clk.now
	ctx := context.Background()
	if err := svc.SetUser(ctx, fast, "alice", alicePw, RoleUser); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetUser(ctx, fast, "bob", "bob-fake-password-22", RoleUser); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetUser(ctx, fast, "ops", "ops-fake-password-333", RoleStaff); err != nil {
		t.Fatal(err)
	}
	return &env{svc: svc, store: st, clk: clk, ver: ver}
}

func TestLoginIssuesVerifiableAccessToken(t *testing.T) {
	e := setup(t)
	tok, err := e.svc.Login(context.Background(), "alice", alicePw)
	if err != nil {
		t.Fatal(err)
	}
	p, err := e.ver.Verify(tok.Access)
	if err != nil || p.Subject != "alice" || p.Staff {
		t.Fatalf("principal %+v err %v", p, err)
	}
	if tok.ExpiresIn != 600 || tok.Subject != "alice" {
		t.Fatalf("expires_in %d subject %q", tok.ExpiresIn, tok.Subject)
	}
	if !tok.RefreshExpires.Equal(e.clk.t.Add(12*time.Hour)) || tok.RefreshMaxAge != 12*3600 {
		t.Fatalf("refresh expiry %v max-age %d", tok.RefreshExpires, tok.RefreshMaxAge)
	}
	if len(tok.Refresh) < 40 || tok.Refresh == tok.Access {
		t.Fatalf("refresh token looks wrong: %q", tok.Refresh)
	}
}

func TestStaffRoleReachesTheToken(t *testing.T) {
	e := setup(t)
	tok, err := e.svc.Login(context.Background(), "ops", "ops-fake-password-333")
	if err != nil {
		t.Fatal(err)
	}
	if p, err := e.ver.Verify(tok.Access); err != nil || !p.Staff {
		t.Fatalf("staff role missing: %+v %v", p, err)
	}
}

func TestLoginFailuresAreIdentical(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	long := make([]byte, 300)
	for i := range long {
		long[i] = 'x'
	}
	for name, c := range map[string][2]string{
		"wrong password":     {"alice", "not-the-password-1"},
		"unknown user":       {"mallory", alicePw},
		"empty both":         {"", ""},
		"bad username chars": {"al ice", alicePw},
		"sql-ish username":   {"alice' OR '1'='1", alicePw},
		"overlong password":  {"alice", string(long)},
	} {
		tok, err := e.svc.Login(ctx, c[0], c[1])
		if !errors.Is(err, ErrInvalidCredentials) || tok != (Tokens{}) {
			t.Errorf("%s: got %v / %+v", name, err, tok)
		}
	}
}

type spy struct {
	calls  int
	inner  PasswordChecker
	stored []string
}

func (s *spy) Verify(pw, stored string) bool {
	s.calls++
	s.stored = append(s.stored, stored)
	return s.inner.Verify(pw, stored)
}

func TestUnknownUserStillCostsOneHashCheck(t *testing.T) {
	e := setup(t)
	sp := &spy{inner: fast}
	svc, err := newService(e.store, e.ver, fast, sp, 10*time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	svc.Now = e.clk.now
	_, _ = svc.Login(context.Background(), "mallory", "whatever-password-1")
	_, _ = svc.Login(context.Background(), "alice", "wrong-password-1234")
	_, _ = svc.Login(context.Background(), "bad name", "whatever-password-1")
	if sp.calls != 3 {
		t.Fatalf("every failed login must do exactly one hash check, got %d for 3 attempts", sp.calls)
	}
	if sp.stored[0] != svc.dummy || sp.stored[2] != svc.dummy {
		t.Fatal("unknown or invalid usernames must be checked against the dummy hash")
	}
}

func TestLoginStoreErrorIsNotACredentialError(t *testing.T) {
	e := setup(t)
	e.store.FailGet = errors.New("db down")
	_, err := e.svc.Login(context.Background(), "alice", alicePw)
	if err == nil || errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("a store failure must surface as an error, got %v", err)
	}
}

func TestRefreshRotates(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	t1, _ := e.svc.Login(ctx, "alice", alicePw)
	e.clk.t = e.clk.t.Add(5 * time.Minute)
	t2, err := e.svc.Refresh(ctx, t1.Refresh)
	if err != nil {
		t.Fatal(err)
	}
	if t2.Refresh == t1.Refresh || t2.Access == t1.Access {
		t.Fatal("refresh must hand out a new refresh token and a new access token")
	}
	if !t2.RefreshExpires.Equal(e.clk.t.Add(12 * time.Hour)) {
		t.Fatalf("new refresh expiry %v", t2.RefreshExpires)
	}
	if p, err := e.ver.Verify(t2.Access); err != nil || p.Subject != "alice" {
		t.Fatalf("new access token invalid: %+v %v", p, err)
	}
	if _, err := e.svc.Refresh(ctx, t2.Refresh); err != nil {
		t.Fatalf("the new refresh token must work: %v", err)
	}
}

func TestRefreshReuseRevokesTheWholeFamily(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	t1, _ := e.svc.Login(ctx, "alice", alicePw)
	t2, err := e.svc.Refresh(ctx, t1.Refresh)
	if err != nil {
		t.Fatal(err)
	}
	// an attacker replays the first token
	_, err = e.svc.Refresh(ctx, t1.Refresh)
	if !errors.Is(err, ErrRefreshReused) || !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("reuse must be reported as ErrRefreshReused (an ErrInvalidRefresh), got %v", err)
	}
	// the legitimate holder's newer token is dead too
	if _, err := e.svc.Refresh(ctx, t2.Refresh); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("family must be revoked after reuse, got %v", err)
	}
	if got := e.store.Live(familyOf(t, e.store, t2.Refresh)); got != 0 {
		t.Fatalf("%d refresh tokens still live in the revoked family", got)
	}
}

func familyOf(t *testing.T, st *MemStore, raw string) string {
	t.Helper()
	f, ok, err := st.FamilyOf(context.Background(), HashRefresh(raw))
	if err != nil || !ok {
		t.Fatalf("family lookup: %v %v", ok, err)
	}
	return f
}

func TestReuseOnlyAffectsThatFamily(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	a1, _ := e.svc.Login(ctx, "alice", alicePw)
	a2, _ := e.svc.Login(ctx, "alice", alicePw) // second device
	b1, _ := e.svc.Login(ctx, "bob", "bob-fake-password-22")
	_, _ = e.svc.Refresh(ctx, a1.Refresh)
	_, _ = e.svc.Refresh(ctx, a1.Refresh) // reuse: first device's family dies
	if _, err := e.svc.Refresh(ctx, a2.Refresh); err != nil {
		t.Fatalf("alice's other session must survive: %v", err)
	}
	if _, err := e.svc.Refresh(ctx, b1.Refresh); err != nil {
		t.Fatalf("bob's session must survive: %v", err)
	}
}

func TestRefreshRejectsBadTokens(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	for name, raw := range map[string]string{
		"empty":    "",
		"short":    "abc",
		"unknown":  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"too long": string(make([]byte, 500)),
	} {
		if _, err := e.svc.Refresh(ctx, raw); !errors.Is(err, ErrInvalidRefresh) || errors.Is(err, ErrRefreshReused) {
			t.Errorf("%s: %v", name, err)
		}
	}
	// an access token is not a refresh token
	tok, _ := e.svc.Login(ctx, "alice", alicePw)
	if _, err := e.svc.Refresh(ctx, tok.Access); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("access token accepted as refresh token: %v", err)
	}
}

func TestRefreshExpires(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	tok, _ := e.svc.Login(ctx, "alice", alicePw)
	e.clk.t = e.clk.t.Add(12*time.Hour - time.Second)
	t2, err := e.svc.Refresh(ctx, tok.Refresh)
	if err != nil {
		t.Fatalf("just before expiry must work: %v", err)
	}
	e.clk.t = e.clk.t.Add(12 * time.Hour)
	if _, err := e.svc.Refresh(ctx, t2.Refresh); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("expired token accepted: %v", err)
	}
}

func TestLogoutRevokesFamily(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	t1, _ := e.svc.Login(ctx, "alice", alicePw)
	t2, _ := e.svc.Refresh(ctx, t1.Refresh)
	if err := e.svc.Logout(ctx, t2.Refresh); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.Refresh(ctx, t2.Refresh); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("token still works after logout: %v", err)
	}
	// logging out twice, or with junk, is harmless
	for _, raw := range []string{t2.Refresh, "", "junk", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"} {
		if err := e.svc.Logout(ctx, raw); err != nil {
			t.Errorf("logout(%q): %v", raw, err)
		}
	}
}

func TestLogoutOnlyEndsOneSession(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	a1, _ := e.svc.Login(ctx, "alice", alicePw)
	a2, _ := e.svc.Login(ctx, "alice", alicePw)
	_ = e.svc.Logout(ctx, a1.Refresh)
	if _, err := e.svc.Refresh(ctx, a2.Refresh); err != nil {
		t.Fatalf("logout on one device ended the other: %v", err)
	}
}

func TestRefreshAfterAccountRemoved(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	tok, _ := e.svc.Login(ctx, "bob", "bob-fake-password-22")
	e.store.DeleteUser("bob")
	if _, err := e.svc.Refresh(ctx, tok.Refresh); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("removed user still refreshed: %v", err)
	}
}

func TestRefreshPicksUpRoleChange(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	tok, _ := e.svc.Login(ctx, "bob", "bob-fake-password-22")
	u, _ := e.store.GetUser(ctx, "bob")
	u.Role = RoleStaff
	_ = e.store.UpsertUser(ctx, u)
	t2, err := e.svc.Refresh(ctx, tok.Refresh)
	if err != nil {
		t.Fatal(err)
	}
	if p, _ := e.ver.Verify(t2.Access); !p.Staff {
		t.Fatal("role is read again at refresh time")
	}
}

func TestOnlyHashesAreStored(t *testing.T) {
	e := setup(t)
	tok, _ := e.svc.Login(context.Background(), "alice", alicePw)
	if _, ok := e.store.tokens[tok.Refresh]; ok {
		t.Fatal("the raw refresh token is used as the storage key")
	}
	if _, ok := e.store.tokens[HashRefresh(tok.Refresh)]; !ok {
		t.Fatal("hash of the refresh token not stored")
	}
	for k, v := range e.store.tokens {
		if k == tok.Refresh || v.rec.Hash == tok.Refresh {
			t.Fatal("raw token found in the store")
		}
	}
}

func TestConcurrentRefreshAllowsExactlyOne(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	tok, _ := e.svc.Login(ctx, "alice", alicePw)
	var ok, reused, other int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := e.svc.Refresh(ctx, tok.Refresh)
			switch {
			case err == nil:
				atomic.AddInt32(&ok, 1)
			case errors.Is(err, ErrRefreshReused):
				atomic.AddInt32(&reused, 1)
			default:
				atomic.AddInt32(&other, 1)
			}
		}()
	}
	wg.Wait()
	if ok != 1 || reused != 19 || other != 0 {
		t.Fatalf("ok=%d reused=%d other=%d, want 1/19/0", ok, reused, other)
	}
}

func TestNewServiceValidatesTTLs(t *testing.T) {
	ver, _ := NewVerifier([]byte(testSecret), "i", "a")
	st := NewMemStore()
	for name, c := range map[string][2]time.Duration{
		"zero access":       {0, time.Hour},
		"access over max":   {2 * time.Hour, 5 * time.Hour},
		"refresh <= access": {10 * time.Minute, 10 * time.Minute},
		"refresh < access":  {10 * time.Minute, time.Minute},
		"negative access":   {-time.Minute, time.Hour},
	} {
		if _, err := NewService(st, ver, fast, c[0], c[1]); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := NewService(nil, ver, fast, time.Minute, time.Hour); err == nil {
		t.Error("nil store accepted")
	}
	if _, err := NewService(st, nil, fast, time.Minute, time.Hour); err == nil {
		t.Error("nil verifier accepted")
	}
}

func TestSetUserValidation(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	if err := e.svc.SetUser(ctx, fast, "bad name", "long-enough-password", RoleUser); err == nil {
		t.Error("bad username accepted")
	}
	if err := e.svc.SetUser(ctx, fast, "carol", "long-enough-password", "admin"); err == nil {
		t.Error("unknown role accepted")
	}
	if err := e.svc.SetUser(ctx, fast, "carol", "short", RoleUser); err == nil {
		t.Error("short password accepted")
	}
	u, _ := e.store.GetUser(ctx, "alice")
	if u.PasswordHash == alicePw || u.PasswordHash == "" {
		t.Error("password stored unhashed")
	}
}
