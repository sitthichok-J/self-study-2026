package http

import (
	"bytes"
	"encoding/json"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/memory"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/service"
)

const (
	alicePw = "alice-fake-password-1"
	loginOK = `{"username":"alice","password":"` + alicePw + `"}`
)

type authEnv struct {
	h     nethttp.Handler
	logs  *bytes.Buffer
	store *auth.MemStore
}

func authSetup(t *testing.T, cookie CookieConfig) *authEnv {
	t.Helper()
	fast := auth.Hasher{Iterations: auth.MinIterations}
	st := auth.NewMemStore()
	sess, err := auth.NewService(st, testVerifier, fast, 10*time.Minute, 12*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []struct{ n, p, r string }{
		{"alice", alicePw, auth.RoleUser}, {"bob", "bob-fake-password-22", auth.RoleUser},
	} {
		if err := sess.SetUser(t.Context(), fast, u.n, u.p, u.r); err != nil {
			t.Fatal(err)
		}
	}
	logs := &bytes.Buffer{}
	q := slog.New(slog.NewJSONHandler(logs, nil))
	svc := service.NewClaims(memory.NewRepo(), memory.NewCache(), q)
	return &authEnv{h: NewWithSessions(svc, nil, q, testVerifier, sess, cookie), logs: logs, store: st}
}

var secureCookie = CookieConfig{Secure: true}

type req struct {
	method, path, body string
	hdr                map[string]string
	cookie             string
}

func (e *authEnv) do(r req) *httptest.ResponseRecorder {
	rq := httptest.NewRequest(r.method, r.path, strings.NewReader(r.body))
	for k, v := range r.hdr {
		rq.Header.Set(k, v)
	}
	if r.cookie != "" {
		rq.AddCookie(&nethttp.Cookie{Name: "refresh_token", Value: r.cookie})
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, rq)
	return rec
}

var csrf = map[string]string{CSRFHeader: CSRFValue}

func (e *authEnv) login(t *testing.T) (access, refresh string) {
	t.Helper()
	rec := e.do(req{method: "POST", path: "/auth/login", body: loginOK})
	if rec.Code != 200 {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	var tr tokenResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &tr)
	return tr.AccessToken, cookieOf(rec, "refresh_token").Value
}

func cookieOf(rec *httptest.ResponseRecorder, name string) *nethttp.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestLoginReturnsAccessTokenAndRefreshCookie(t *testing.T) {
	e := authSetup(t, secureCookie)
	rec := e.do(req{method: "POST", path: "/auth/login", body: loginOK})
	if rec.Code != 200 {
		t.Fatalf("status %d body %s", rec.Code, rec.Body)
	}
	var tr tokenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &tr); err != nil {
		t.Fatal(err)
	}
	if tr.TokenType != "Bearer" || tr.ExpiresIn != 600 {
		t.Fatalf("response %+v", tr)
	}
	p, err := testVerifier.Verify(tr.AccessToken)
	if err != nil || p.Subject != "alice" {
		t.Fatalf("access token: %+v %v", p, err)
	}
	c := cookieOf(rec, "refresh_token")
	if c == nil || c.Value == "" {
		t.Fatal("no refresh cookie")
	}
	if !c.HttpOnly || !c.Secure || c.SameSite != nethttp.SameSiteStrictMode || c.Path != "/api/auth" {
		t.Fatalf("cookie attributes wrong: %+v", c)
	}
	if c.MaxAge != 12*3600 {
		t.Fatalf("cookie max-age %d", c.MaxAge)
	}
	if strings.Contains(rec.Body.String(), c.Value) {
		t.Fatal("refresh token must never be in the response body")
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("token responses must not be cacheable")
	}
}

func TestLoginTokenWorksOnClaims(t *testing.T) {
	e := authSetup(t, secureCookie)
	access, _ := e.login(t)
	rec := e.do(req{method: "POST", path: "/claims", body: `{"patient_ref":"FAKE-AB12","amount":100}`,
		hdr: map[string]string{"Authorization": "Bearer " + access, "Idempotency-Key": "k1"}})
	if rec.Code != 201 || !strings.Contains(rec.Body.String(), `"owner_id":"alice"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestCookieConfigIsApplied(t *testing.T) {
	e := authSetup(t, CookieConfig{Name: "rt", Path: "/auth", Secure: false})
	rec := e.do(req{method: "POST", path: "/auth/login", body: loginOK})
	c := cookieOf(rec, "rt")
	if c == nil || c.Path != "/auth" || c.Secure || !c.HttpOnly {
		t.Fatalf("%+v", c)
	}
}

func TestLoginFailuresLookIdentical(t *testing.T) {
	e := authSetup(t, secureCookie)
	wrongPw := e.do(req{method: "POST", path: "/auth/login", body: `{"username":"alice","password":"nope-nope-nope-1"}`})
	noUser := e.do(req{method: "POST", path: "/auth/login", body: `{"username":"mallory","password":"nope-nope-nope-1"}`})
	for name, rec := range map[string]*httptest.ResponseRecorder{"wrong password": wrongPw, "unknown user": noUser} {
		if rec.Code != 401 {
			t.Fatalf("%s: %d", name, rec.Code)
		}
		if len(rec.Result().Cookies()) != 0 {
			t.Fatalf("%s: a cookie was set", name)
		}
	}
	if want := `{"error":"invalid username or password"}` + "\n"; wrongPw.Body.String() != want {
		t.Fatalf("login failure body %q, want %q (the UI shows this message)", wrongPw.Body, want)
	}
	if wrongPw.Body.String() != noUser.Body.String() {
		t.Fatalf("bodies differ: %q vs %q", wrongPw.Body, noUser.Body)
	}
	logs := e.logs.String()
	if !strings.Contains(logs, `"security_event":"login_failed"`) {
		t.Fatal("failed login must be logged as a security event")
	}
	if strings.Contains(logs, "nope-nope") || strings.Contains(logs, "mallory") {
		t.Fatal("password or username leaked into the log")
	}
}

func TestLoginRejectsBadBodies(t *testing.T) {
	e := authSetup(t, secureCookie)
	big := `{"username":"alice","password":"` + strings.Repeat("x", 6000) + `"}`
	for name, body := range map[string]string{
		"empty":         ``,
		"not json":      `username=alice&password=x`,
		"unknown field": `{"username":"alice","password":"` + alicePw + `","role":"staff"}`,
		"too large":     big,
		"array":         `["alice"]`,
	} {
		rec := e.do(req{method: "POST", path: "/auth/login", body: body})
		if rec.Code != 400 {
			t.Errorf("%s: status %d", name, rec.Code)
		}
		if len(rec.Result().Cookies()) != 0 {
			t.Errorf("%s: cookie set", name)
		}
	}
}

func TestAuthRoutesAreOnlyPOST(t *testing.T) {
	e := authSetup(t, secureCookie)
	for _, p := range []string{"/auth/login", "/auth/refresh", "/auth/logout"} {
		if rec := e.do(req{method: "GET", path: p}); rec.Code != 405 {
			t.Errorf("GET %s: %d", p, rec.Code)
		}
	}
}

func TestRefreshNeedsCSRFHeader(t *testing.T) {
	e := authSetup(t, secureCookie)
	_, refresh := e.login(t)
	for name, hdr := range map[string]map[string]string{
		"missing": nil, "wrong value": {CSRFHeader: "XMLHttpRequest"}, "empty": {CSRFHeader: ""},
	} {
		rec := e.do(req{method: "POST", path: "/auth/refresh", cookie: refresh, hdr: hdr})
		if rec.Code != 403 {
			t.Errorf("%s: %d", name, rec.Code)
		}
	}
	// a rejected CSRF check must not burn the refresh token
	if rec := e.do(req{method: "POST", path: "/auth/refresh", cookie: refresh, hdr: csrf}); rec.Code != 200 {
		t.Fatalf("token was consumed by a request that failed the CSRF check: %d", rec.Code)
	}
}

func TestRefreshRotatesTheCookie(t *testing.T) {
	e := authSetup(t, secureCookie)
	_, r1 := e.login(t)
	rec := e.do(req{method: "POST", path: "/auth/refresh", cookie: r1, hdr: csrf})
	if rec.Code != 200 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	c := cookieOf(rec, "refresh_token")
	if c == nil || c.Value == "" || c.Value == r1 || !c.HttpOnly || !c.Secure || c.SameSite != nethttp.SameSiteStrictMode {
		t.Fatalf("rotated cookie wrong: %+v", c)
	}
	var tr tokenResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &tr)
	if p, err := testVerifier.Verify(tr.AccessToken); err != nil || p.Subject != "alice" {
		t.Fatalf("new access token: %+v %v", p, err)
	}
}

func TestRefreshWithoutOrWithBadCookieIs401AndClearsCookie(t *testing.T) {
	e := authSetup(t, secureCookie)
	for name, cookie := range map[string]string{"no cookie": "", "garbage": "garbage", "unknown": strings.Repeat("A", 43)} {
		rec := e.do(req{method: "POST", path: "/auth/refresh", cookie: cookie, hdr: csrf})
		if rec.Code != 401 {
			t.Errorf("%s: %d", name, rec.Code)
		}
		if c := cookieOf(rec, "refresh_token"); c == nil || c.MaxAge >= 0 {
			t.Errorf("%s: cookie not cleared: %+v", name, c)
		}
	}
}

func TestReusedRefreshTokenKillsTheSession(t *testing.T) {
	e := authSetup(t, secureCookie)
	_, r1 := e.login(t)
	rec := e.do(req{method: "POST", path: "/auth/refresh", cookie: r1, hdr: csrf})
	r2 := cookieOf(rec, "refresh_token").Value
	// replay of the old token
	rec = e.do(req{method: "POST", path: "/auth/refresh", cookie: r1, hdr: csrf})
	if rec.Code != 401 {
		t.Fatalf("replay: %d", rec.Code)
	}
	if !strings.Contains(e.logs.String(), `"security_event":"refresh_reuse"`) {
		t.Fatal("reuse must be logged as a security event")
	}
	// the newer token is dead as well
	if rec := e.do(req{method: "POST", path: "/auth/refresh", cookie: r2, hdr: csrf}); rec.Code != 401 {
		t.Fatalf("family survived reuse: %d", rec.Code)
	}
}

func TestLogoutClearsCookieAndEndsSession(t *testing.T) {
	e := authSetup(t, secureCookie)
	_, r1 := e.login(t)
	rec := e.do(req{method: "POST", path: "/auth/logout", cookie: r1, hdr: csrf})
	if rec.Code != 204 {
		t.Fatalf("%d", rec.Code)
	}
	if c := cookieOf(rec, "refresh_token"); c == nil || c.MaxAge >= 0 || c.Value != "" {
		t.Fatalf("cookie not cleared: %+v", c)
	}
	if rec := e.do(req{method: "POST", path: "/auth/refresh", cookie: r1, hdr: csrf}); rec.Code != 401 {
		t.Fatalf("refresh after logout: %d", rec.Code)
	}
	// twice, and with no cookie at all: still 204
	for _, c := range []string{r1, ""} {
		if rec := e.do(req{method: "POST", path: "/auth/logout", cookie: c, hdr: csrf}); rec.Code != 204 {
			t.Fatalf("repeat logout: %d", rec.Code)
		}
	}
}

func TestLogoutNeedsCSRFHeader(t *testing.T) {
	e := authSetup(t, secureCookie)
	_, r1 := e.login(t)
	if rec := e.do(req{method: "POST", path: "/auth/logout", cookie: r1}); rec.Code != 403 {
		t.Fatalf("%d", rec.Code)
	}
	if rec := e.do(req{method: "POST", path: "/auth/refresh", cookie: r1, hdr: csrf}); rec.Code != 200 {
		t.Fatal("a forged logout must not end the session")
	}
}

func TestStoreErrorsAreGeneric500s(t *testing.T) {
	e := authSetup(t, secureCookie)
	e.store.FailGet = errString("pq: password authentication failed for user secret_user")
	rec := e.do(req{method: "POST", path: "/auth/login", body: loginOK})
	if rec.Code != 500 || strings.Contains(rec.Body.String(), "secret_user") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func TestNoAuthRoutesWithoutSessions(t *testing.T) {
	h, _, _ := setup(nil) // plain New(): the Day 10 router
	for _, p := range []string{"/auth/login", "/auth/refresh", "/auth/logout"} {
		if rec := do(h, "POST", p, loginOK, map[string]string{"Authorization": ""}); rec.Code != 404 && rec.Code != 405 {
			t.Errorf("%s exists without a session service: %d", p, rec.Code)
		}
	}
}

func TestClaimsStillNeedAToken(t *testing.T) {
	e := authSetup(t, secureCookie)
	if rec := e.do(req{method: "GET", path: "/claims"}); rec.Code != 401 {
		t.Fatalf("%d", rec.Code)
	}
	// a refresh token is not an access token
	_, r1 := e.login(t)
	if rec := e.do(req{method: "GET", path: "/claims", hdr: map[string]string{"Authorization": "Bearer " + r1}}); rec.Code != 401 {
		t.Fatalf("refresh token accepted as access token: %d", rec.Code)
	}
}

func TestTokensNeverAppearInLogs(t *testing.T) {
	e := authSetup(t, secureCookie)
	access, r1 := e.login(t)
	rec := e.do(req{method: "POST", path: "/auth/refresh", cookie: r1, hdr: csrf})
	r2 := cookieOf(rec, "refresh_token").Value
	_ = e.do(req{method: "POST", path: "/auth/refresh", cookie: r1, hdr: csrf}) // reuse -> warning log
	_ = e.do(req{method: "GET", path: "/claims", hdr: map[string]string{"Authorization": "Bearer x" + access}})
	logs := e.logs.String()
	for name, secret := range map[string]string{"access": access, "refresh1": r1, "refresh2": r2, "password": alicePw} {
		if strings.Contains(logs, secret) {
			t.Errorf("%s token/secret found in logs", name)
		}
	}
}
