package http

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/memory"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/ratelimit"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/service"
)

type guardEnv struct {
	h    nethttp.Handler
	rev  *auth.MemRevoker
	lim  *ratelimit.Mem
	repo *memory.Repo
	logs *bytes.Buffer
}

func guardSetup(t *testing.T, limit int) *guardEnv {
	t.Helper()
	fast := auth.Hasher{Iterations: auth.MinIterations}
	st := auth.NewMemStore()
	sess, err := auth.NewService(st, testVerifier, fast, 10*time.Minute, 12*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.SetUser(t.Context(), fast, "alice", alicePw, auth.RoleUser); err != nil {
		t.Fatal(err)
	}
	logs := &bytes.Buffer{}
	q := slog.New(slog.NewJSONHandler(logs, nil))
	repo := memory.NewRepo()
	svc := service.NewClaims(repo, memory.NewCache(), q)
	e := &guardEnv{rev: auth.NewMemRevoker(), lim: ratelimit.NewMem(), repo: repo, logs: logs}
	e.h = NewFull(svc, nil, q, testVerifier, sess, secureCookie,
		Options{Revoker: e.rev, Limiter: e.lim, UserLimit: limit, UserWindow: time.Minute})
	return e
}

func (e *guardEnv) call(method, path, body, tok string, extra map[string]string) *httptest.ResponseRecorder {
	rq := httptest.NewRequest(method, path, strings.NewReader(body))
	if tok != "" {
		rq.Header.Set("Authorization", "Bearer "+tok)
	}
	for k, v := range extra {
		rq.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, rq)
	return rec
}

func principalOf(t *testing.T, tok string) auth.Principal {
	t.Helper()
	p, err := testVerifier.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRevokedTokenIsRejected(t *testing.T) {
	e := guardSetup(t, 0)
	tok := tokenFor("alice", "user")
	if rec := e.call("GET", "/claims", "", tok, nil); rec.Code != 200 {
		t.Fatalf("before revoke: %d", rec.Code)
	}
	p := principalOf(t, tok)
	_ = e.rev.RevokeToken(t.Context(), p.JTI, testVerifier.RevokeUntil(p))
	rec := e.call("GET", "/claims", "", tok, nil)
	if rec.Code != 401 || rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatalf("revoked: %d %v", rec.Code, rec.Header())
	}
	if !strings.Contains(e.logs.String(), `"reason":"revoked"`) {
		t.Fatalf("log must say why: %s", e.logs)
	}
	if rec := e.call("GET", "/claims", "", tokenFor("alice", "user"), nil); rec.Code != 200 {
		t.Fatalf("a new token of the same user must work: %d", rec.Code)
	}
}

func TestUserWideCutRejectsOldTokens(t *testing.T) {
	e := guardSetup(t, 0)
	old := tokenFor("alice", "user")
	_ = e.rev.RevokeUser(t.Context(), "alice", time.Now().Add(time.Second), time.Hour)
	if rec := e.call("GET", "/claims", "", old, nil); rec.Code != 401 {
		t.Fatalf("old token after cut: %d", rec.Code)
	}
	if rec := e.call("GET", "/claims", "", tokenFor("bob", "user"), nil); rec.Code != 200 {
		t.Fatalf("bob: %d", rec.Code)
	}
}

func TestRevocationStoreDownFailsClosed(t *testing.T) {
	e := guardSetup(t, 0)
	e.rev.Fail = errors.New("valkey down")
	rec := e.call("GET", "/claims", "", tokenFor("alice", "user"), nil)
	if rec.Code != 503 || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("want 503 + Retry-After, got %d %v", rec.Code, rec.Header())
	}
	if strings.Contains(rec.Body.String(), "valkey") {
		t.Fatal("error detail leaked")
	}
	// an invalid token is still just 401: the store is not asked
	if rec := e.call("GET", "/claims", "", "garbage", nil); rec.Code != 401 {
		t.Fatalf("garbage token: %d", rec.Code)
	}
	if rec := e.call("GET", "/healthz", "", "", nil); rec.Code != 200 {
		t.Fatalf("healthz must not depend on the blocklist: %d", rec.Code)
	}
}

func TestLogoutRevokesPresentedAccessToken(t *testing.T) {
	e := guardSetup(t, 0)
	rec := e.call("POST", "/auth/login", loginOK, "", nil)
	if rec.Code != 200 {
		t.Fatalf("login %d", rec.Code)
	}
	var refresh string
	for _, c := range rec.Result().Cookies() {
		if c.Name == "refresh_token" {
			refresh = c.Value
		}
	}
	access := tokenFor("alice", "user")
	if r := e.call("GET", "/claims", "", access, nil); r.Code != 200 {
		t.Fatal("token should work before logout")
	}
	rq := httptest.NewRequest("POST", "/auth/logout", nil)
	rq.Header.Set(CSRFHeader, CSRFValue)
	rq.Header.Set("Authorization", "Bearer "+access)
	rq.AddCookie(&nethttp.Cookie{Name: "refresh_token", Value: refresh})
	out := httptest.NewRecorder()
	e.h.ServeHTTP(out, rq)
	if out.Code != 204 {
		t.Fatalf("logout %d", out.Code)
	}
	if r := e.call("GET", "/claims", "", access, nil); r.Code != 401 {
		t.Fatalf("access token must be dead right after logout, got %d", r.Code)
	}
	if !strings.Contains(e.logs.String(), `"security_event":"token_revoked"`) {
		t.Fatal("logout must leave a log line")
	}
}

func TestLogoutWithoutOrWithBadBearerStillSucceeds(t *testing.T) {
	e := guardSetup(t, 0)
	for _, hdr := range []map[string]string{{CSRFHeader: CSRFValue}, {CSRFHeader: CSRFValue, "Authorization": "Bearer junk"}} {
		if r := e.call("POST", "/auth/logout", "", "", hdr); r.Code != 204 {
			t.Fatalf("logout %v -> %d", hdr, r.Code)
		}
	}
}

func TestLogoutReportsFailureWhenBlocklistDown(t *testing.T) {
	e := guardSetup(t, 0)
	e.rev.Fail = errors.New("down")
	hdr := map[string]string{CSRFHeader: CSRFValue}
	rq := httptest.NewRequest("POST", "/auth/logout", nil)
	for k, v := range hdr {
		rq.Header.Set(k, v)
	}
	rq.Header.Set("Authorization", "Bearer "+tokenFor("alice", "user"))
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, rq)
	if rec.Code != 503 {
		t.Fatalf("want 503 (logout incomplete), got %d", rec.Code)
	}
	if !strings.Contains(strings.Join(rec.Header().Values("Set-Cookie"), ";"), "Max-Age=0") {
		t.Fatalf("cookie should still be cleared: %v", rec.Header())
	}
}

func TestPerUserRateLimit(t *testing.T) {
	e := guardSetup(t, 3)
	now := time.Unix(2_000_000, 0)
	e.lim.Now = func() time.Time { return now }
	alice := tokenFor("alice", "user")
	for i := 1; i <= 3; i++ {
		rec := e.call("GET", "/claims", "", alice, nil)
		if rec.Code != 200 || rec.Header().Get("X-RateLimit-Remaining-User") == "" {
			t.Fatalf("call %d: %d %v", i, rec.Code, rec.Header())
		}
	}
	now = now.Add(20 * time.Second)
	rec := e.call("GET", "/claims", "", alice, nil)
	if rec.Code != 429 || rec.Header().Get("Retry-After") != "40" {
		t.Fatalf("4th call: %d Retry-After=%q", rec.Code, rec.Header().Get("Retry-After"))
	}
	if !strings.Contains(e.logs.String(), `"security_event":"user_rate_limited"`) {
		t.Fatal("must log the event")
	}
	if rec := e.call("GET", "/claims", "", tokenFor("bob", "user"), nil); rec.Code != 200 {
		t.Fatalf("bob has his own counter: %d", rec.Code)
	}
	if rec := e.call("GET", "/healthz", "", "", nil); rec.Code != 200 {
		t.Fatal("healthz is not limited")
	}
	now = now.Add(40 * time.Second) // window over
	if rec := e.call("GET", "/claims", "", alice, nil); rec.Code != 200 {
		t.Fatalf("after the window: %d", rec.Code)
	}
}

func TestRateLimitCountsOnlyAuthenticatedRequests(t *testing.T) {
	e := guardSetup(t, 1)
	for i := 0; i < 5; i++ {
		if rec := e.call("GET", "/claims", "", "garbage", nil); rec.Code != 401 {
			t.Fatalf("garbage: %d", rec.Code)
		}
	}
	if rec := e.call("GET", "/claims", "", tokenFor("alice", "user"), nil); rec.Code != 200 {
		t.Fatalf("failed attempts must not burn alice's quota: %d", rec.Code)
	}
}

func TestRateLimitStoreDownFailsOpen(t *testing.T) {
	e := guardSetup(t, 1)
	e.lim.Fail = errors.New("down")
	for i := 0; i < 3; i++ {
		rec := e.call("GET", "/claims", "", tokenFor("alice", "user"), nil)
		if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"items"`) { // the request really reached the handler
			t.Fatalf("call %d: %d %q", i, rec.Code, rec.Body)
		}
	}
	if !strings.Contains(e.logs.String(), "rate_limit_unavailable") {
		t.Fatal("must be logged")
	}
}

// stubLimiter answers with a fixed Result (a real window never reports RetryAfter <= 0, but a limiter might).
type stubLimiter struct{ res ratelimit.Result }

func (s stubLimiter) Allow(context.Context, string, int, time.Duration) (ratelimit.Result, error) {
	return s.res, nil
}

func TestRetryAfterIsAtLeastOneSecond(t *testing.T) {
	for name, ra := range map[string]time.Duration{"zero": 0, "negative": -time.Second, "fraction": 300 * time.Millisecond, "just over": 1100 * time.Millisecond} {
		e := guardSetup(t, 1)
		fast := auth.Hasher{Iterations: auth.MinIterations}
		sess, _ := auth.NewService(auth.NewMemStore(), testVerifier, fast, 10*time.Minute, 12*time.Hour)
		q := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
		h := NewFull(service.NewClaims(memory.NewRepo(), memory.NewCache(), q), nil, q, testVerifier, sess, secureCookie,
			Options{Limiter: stubLimiter{ratelimit.Result{Allowed: false, RetryAfter: ra}}, UserLimit: 1, UserWindow: time.Minute})
		_ = e
		rq := httptest.NewRequest("GET", "/claims", nil)
		rq.Header.Set("Authorization", "Bearer "+tokenFor("alice", "user"))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, rq)
		want := map[string]string{"zero": "1", "negative": "1", "fraction": "1", "just over": "2"}[name]
		if rec.Code != 429 || rec.Header().Get("Retry-After") != want {
			t.Errorf("%s: %d Retry-After=%q want %q", name, rec.Code, rec.Header().Get("Retry-After"), want)
		}
	}
}

func TestRateLimitOffWhenZero(t *testing.T) {
	e := guardSetup(t, 0)
	for i := 0; i < 10; i++ {
		if rec := e.call("GET", "/claims", "", tokenFor("alice", "user"), nil); rec.Code != 200 {
			t.Fatalf("call %d: %d", i, rec.Code)
		}
	}
}

// ---- manual review: approve / reject ----

func (e *guardEnv) pendingClaim(t *testing.T, key string) string {
	t.Helper()
	rec := e.call("POST", "/claims", `{"patient_ref":"FAKE-REV1","amount":3000000}`, tokenFor("alice", "user"), map[string]string{"Idempotency-Key": key})
	if rec.Code != 201 {
		t.Fatalf("create: %d %s", rec.Code, rec.Body)
	}
	c, err := e.repo.List(t.Context(), "alice", nil, 10, 0)
	if err != nil || len(c) == 0 {
		t.Fatal("no claim stored")
	}
	for _, x := range c {
		if x.IdempotencyKey == key {
			return x.ID
		}
	}
	t.Fatal("claim not found")
	return ""
}

func TestStaffApprovesAndRejects(t *testing.T) {
	e := guardSetup(t, 0)
	staff := tokenFor("ops", "staff")
	for _, tc := range []struct{ action, status, reason string }{
		{"approve", "APPROVED", "manual_approved"}, {"reject", "REJECTED", "manual_rejected"},
	} {
		id := e.pendingClaim(t, "k-"+tc.action)
		rec := e.call("POST", "/claims/"+id+"/"+tc.action, "", staff, nil)
		if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"status":"`+tc.status+`"`) || !strings.Contains(rec.Body.String(), tc.reason) {
			t.Fatalf("%s: %d %s", tc.action, rec.Code, rec.Body)
		}
		// a second decision is a conflict, and the first one stands
		again := e.call("POST", "/claims/"+id+"/"+map[string]string{"approve": "reject", "reject": "approve"}[tc.action], "", staff, nil)
		if again.Code != 409 {
			t.Fatalf("second decision: %d", again.Code)
		}
		got, _ := e.repo.Get(t.Context(), id)
		if string(got.Status) != tc.status {
			t.Fatalf("status flipped to %s", got.Status)
		}
	}
}

func TestOnlyStaffMayDecide(t *testing.T) {
	e := guardSetup(t, 0)
	id := e.pendingClaim(t, "k1")
	for _, action := range []string{"approve", "reject"} {
		if rec := e.call("POST", "/claims/"+id+"/"+action, "", tokenFor("alice", "user"), nil); rec.Code != 403 {
			t.Fatalf("owner (not staff) %s: %d", action, rec.Code)
		}
		if rec := e.call("POST", "/claims/"+id+"/"+action, "", "", nil); rec.Code != 401 {
			t.Fatalf("no token %s: %d", action, rec.Code)
		}
	}
	got, _ := e.repo.Get(t.Context(), id)
	if got.Status != domain.StatusPending {
		t.Fatalf("claim changed to %s", got.Status)
	}
}

func TestDecideErrors(t *testing.T) {
	e := guardSetup(t, 0)
	staff := tokenFor("ops", "staff")
	if rec := e.call("POST", "/claims/not-a-uuid/approve", "", staff, nil); rec.Code != 400 {
		t.Fatalf("bad id: %d", rec.Code)
	}
	if rec := e.call("POST", "/claims/00000000-0000-4000-8000-0000000000ff/approve", "", staff, nil); rec.Code != 404 {
		t.Fatalf("unknown id: %d", rec.Code)
	}
	if rec := e.call("GET", "/claims/00000000-0000-4000-8000-0000000000ff/approve", "", staff, nil); rec.Code != 405 {
		t.Fatalf("GET on approve: %d", rec.Code)
	}
}
