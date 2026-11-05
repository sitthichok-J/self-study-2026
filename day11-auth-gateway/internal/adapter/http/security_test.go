package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/memory"
	"github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/service"
)

func bearer(sub, role string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + tokenFor(sub, role)}
}

func withKey(h map[string]string, key string) map[string]string {
	h["Idempotency-Key"] = key
	return h
}

const goodBody = `{"patient_ref":"FAKE-AB12","amount":1500}`

func newMetrics(t *testing.T) *obs.Metrics {
	t.Helper()
	m := obs.NewMetrics()
	old := obs.SetMetrics(m)
	t.Cleanup(func() { obs.SetMetrics(old) })
	return m
}

// ---- authentication ----

func TestClaimsRoutesRequireAuth(t *testing.T) {
	newMetrics(t)
	h, _, _ := setup(nil)
	none := map[string]string{"Authorization": ""}
	for _, rt := range [][2]string{
		{"POST", "/claims"}, {"GET", "/claims"}, {"GET", "/claims/00000000-0000-4000-8000-0000000000ff"},
	} {
		r := do(h, rt[0], rt[1], goodBody, withKey(map[string]string{"Authorization": ""}, "k"))
		if r.Code != 401 || r.Header().Get("WWW-Authenticate") == "" {
			t.Errorf("%v: got %d", rt, r.Code)
		}
		_ = none
	}
}

func TestBadTokensAre401WithOneGenericBody(t *testing.T) {
	m := newMetrics(t)
	h, _, _ := setup(nil)
	good := tokenFor("alice", "user")
	parts := strings.Split(good, ".")
	cases := map[string]string{
		"garbage":     "Bearer not.a.jwt",
		"bad sig":     "Bearer " + parts[0] + "." + parts[1] + ".AAAA",
		"basic":       "Basic YWxpY2U6cHc=",
		"alg none":    "Bearer eyJhbGciOiJub25lIn0." + parts[1] + ".",
		"empty":       "Bearer ",
		"only scheme": "Bearer",
	}
	var bodies = map[string]bool{}
	for n, a := range cases {
		r := do(h, "GET", "/claims", "", map[string]string{"Authorization": a})
		if r.Code != 401 {
			t.Errorf("%s: got %d", n, r.Code)
		}
		bodies[r.Body.String()] = true
	}
	if len(bodies) != 1 {
		t.Fatalf("error body must not reveal why: %v", bodies)
	}
	if m.AuthFailures.Value("signature") == 0 || m.AuthFailures.Value("alg") == 0 || m.AuthFailures.Value("missing") == 0 {
		t.Fatal("failures must be counted per reason")
	}
}

func TestAuthFailureIsLoggedWithoutTheToken(t *testing.T) {
	newMetrics(t)
	var buf bytes.Buffer
	log := obs.NewLogger(&buf, slog.LevelDebug)
	repo := memory.NewRepo()
	h := New(service.NewClaims(repo, nil, log), nil, log, testVerifier)
	req := httptest.NewRequest("GET", "/claims", nil)
	req.Header.Set("Authorization", "Bearer SUPER.SECRET.TOKEN")
	h.ServeHTTP(httptest.NewRecorder(), req)
	out := buf.String()
	if !strings.Contains(out, `"security_event":"auth_failed"`) {
		t.Fatalf("security event missing: %s", out)
	}
	if strings.Contains(out, "SUPER.SECRET.TOKEN") || strings.Contains(out, "SECRET") {
		t.Fatalf("token leaked into log: %s", out)
	}
}

func TestHealthEndpointsStayOpen(t *testing.T) {
	h, _, _ := setup(nil)
	for _, p := range []string{"/healthz", "/readyz"} {
		if r := do(h, "GET", p, "", map[string]string{"Authorization": ""}); r.Code != 200 {
			t.Errorf("%s: %d", p, r.Code)
		}
	}
}

func TestMetricsNotServedOnPublicPort(t *testing.T) {
	h, _, _ := setup(nil)
	for _, tok := range []map[string]string{{"Authorization": ""}, bearer("alice", "staff")} {
		if r := do(h, "GET", "/metrics", "", tok); r.Code != 404 {
			t.Fatalf("/metrics on the API port must not exist, got %d", r.Code)
		}
	}
}

func TestNewRefusesNilVerifier(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("New must fail closed without a verifier")
		}
	}()
	New(service.NewClaims(memory.NewRepo(), nil, nil), nil, nil, nil)
}

// ---- authorization / IDOR (exercise 1: attack yourself) ----

func submitAs(t *testing.T, h nethttp.Handler, sub, key string) domain.Claim {
	t.Helper()
	r := do(h, "POST", "/claims", goodBody, withKey(bearer(sub, "user"), key))
	if r.Code != 201 {
		t.Fatalf("submit as %s: %d %s", sub, r.Code, r.Body)
	}
	var c domain.Claim
	_ = json.Unmarshal(r.Body.Bytes(), &c)
	return c
}

func TestIDORReadingAnotherUsersClaimLooksLikeNotFound(t *testing.T) {
	m := newMetrics(t)
	h, _, _ := setup(nil)
	alices := submitAs(t, h, "alice", "k1")
	missing := do(h, "GET", "/claims/00000000-0000-4000-8000-0000000000ff", "", bearer("mallory", "user"))
	stolen := do(h, "GET", "/claims/"+alices.ID, "", bearer("mallory", "user"))
	if stolen.Code != 404 || stolen.Body.String() != missing.Body.String() {
		t.Fatalf("other user's claim must be indistinguishable from a missing one: %d %s vs %d %s",
			stolen.Code, stolen.Body, missing.Code, missing.Body)
	}
	if strings.Contains(stolen.Body.String(), "FAKE-AB12") {
		t.Fatal("no data may leak")
	}
	if m.AuthzDenied.Value("claim") != 1 {
		t.Fatal("denied access must be counted")
	}
	if own := do(h, "GET", "/claims/"+alices.ID, "", bearer("alice", "user")); own.Code != 200 {
		t.Fatalf("owner must still read it: %d", own.Code)
	}
}

func TestIDORStillBlockedWhenTheClaimIsCached(t *testing.T) {
	newMetrics(t)
	h, _, _ := setup(nil)
	c := submitAs(t, h, "alice", "k1")
	do(h, "GET", "/claims/"+c.ID, "", bearer("alice", "user")) // fills the cache
	do(h, "GET", "/claims/"+c.ID, "", bearer("alice", "user")) // served from cache
	if r := do(h, "GET", "/claims/"+c.ID, "", bearer("mallory", "user")); r.Code != 404 {
		t.Fatalf("cache must not bypass the owner check: %d", r.Code)
	}
}

func TestIDORAccessIsLoggedAsSecurityEvent(t *testing.T) {
	newMetrics(t)
	var buf bytes.Buffer
	log := obs.NewLogger(&buf, slog.LevelDebug)
	svc := service.NewClaims(memory.NewRepo(), nil, log)
	h := New(svc, nil, log, testVerifier)
	c, _, _ := svc.Submit(context.Background(), service.SubmitInput{Owner: "alice", IdempotencyKey: "k", PatientRef: "FAKE-AB12", Amount: 5})
	req := httptest.NewRequest("GET", "/claims/"+c.ID, nil)
	req.Header.Set("Authorization", "Bearer "+tokenFor("mallory", "user"))
	h.ServeHTTP(httptest.NewRecorder(), req)
	out := buf.String()
	if !strings.Contains(out, `"security_event":"access_denied"`) || !strings.Contains(out, `"sub":"mallory"`) {
		t.Fatalf("missing security log: %s", out)
	}
	if strings.Contains(out, "FAKE-AB12") {
		t.Fatal("patient_ref must not be logged")
	}
}

func TestListShowsOnlyOwnClaimsAndStaffSeesAll(t *testing.T) {
	newMetrics(t)
	h, _, _ := setup(nil)
	a := submitAs(t, h, "alice", "k1")
	b := submitAs(t, h, "bob", "k1")
	ids := func(who map[string]string) map[string]bool {
		r := do(h, "GET", "/claims", "", who)
		var out struct{ Items []domain.Claim }
		_ = json.Unmarshal(r.Body.Bytes(), &out)
		m := map[string]bool{}
		for _, c := range out.Items {
			m[c.ID] = true
		}
		return m
	}
	if got := ids(bearer("alice", "user")); len(got) != 1 || !got[a.ID] {
		t.Fatalf("alice sees %v", got)
	}
	if got := ids(bearer("bob", "user")); len(got) != 1 || !got[b.ID] {
		t.Fatalf("bob sees %v", got)
	}
	if got := ids(bearer("carol", "user")); len(got) != 0 {
		t.Fatalf("carol sees %v", got)
	}
	if got := ids(bearer("ops", "staff")); len(got) != 2 {
		t.Fatalf("staff sees %v", got)
	}
}

func TestListCacheIsNotSharedBetweenUsers(t *testing.T) {
	newMetrics(t)
	h, _, _ := setup(nil)
	submitAs(t, h, "alice", "k1")
	do(h, "GET", "/claims", "", bearer("alice", "user")) // caches alice's page
	r := do(h, "GET", "/claims", "", bearer("bob", "user"))
	if strings.Contains(r.Body.String(), "FAKE-AB12") {
		t.Fatalf("bob got alice's cached list: %s", r.Body)
	}
	// and staff does not poison it either
	do(h, "GET", "/claims", "", bearer("ops", "staff"))
	if r := do(h, "GET", "/claims", "", bearer("bob", "user")); strings.Contains(r.Body.String(), "FAKE-AB12") {
		t.Fatal("staff list leaked into a user's page")
	}
}

func TestStaffCanReadAnyClaim(t *testing.T) {
	newMetrics(t)
	h, _, _ := setup(nil)
	c := submitAs(t, h, "alice", "k1")
	if r := do(h, "GET", "/claims/"+c.ID, "", bearer("ops", "staff")); r.Code != 200 {
		t.Fatalf("%d", r.Code)
	}
}

func TestOwnerComesFromTheTokenNotTheBody(t *testing.T) {
	newMetrics(t)
	h, _, _ := setup(nil)
	r := do(h, "POST", "/claims", `{"patient_ref":"FAKE-AB12","amount":5,"owner_id":"bob"}`, withKey(bearer("mallory", "user"), "k"))
	if r.Code != 400 {
		t.Fatalf("unknown field owner_id must be refused, got %d", r.Code)
	}
	c := submitAs(t, h, "mallory", "k2")
	if c.OwnerID != "mallory" {
		t.Fatalf("owner = %q", c.OwnerID)
	}
}

func TestIdempotencyKeysAreScopedPerUser(t *testing.T) {
	newMetrics(t)
	h, _, repo := setup(nil)
	a := submitAs(t, h, "alice", "same-key")
	// bob re-uses alice's key with a different payload: must be his own new claim, not 409 and not alice's claim
	r := do(h, "POST", "/claims", `{"patient_ref":"FAKE-ZZ99","amount":77}`, withKey(bearer("bob", "user"), "same-key"))
	if r.Code != 201 {
		t.Fatalf("bob: %d %s", r.Code, r.Body)
	}
	var b domain.Claim
	_ = json.Unmarshal(r.Body.Bytes(), &b)
	if b.ID == a.ID || b.OwnerID != "bob" || repo.Count() != 2 {
		t.Fatalf("keys leaked across users: %+v", b)
	}
	// and bob re-using the same key with the SAME payload as alice's must not replay alice's claim
	r = do(h, "POST", "/claims", goodBody, withKey(bearer("carol", "user"), "same-key"))
	var c domain.Claim
	_ = json.Unmarshal(r.Body.Bytes(), &c)
	if r.Code != 201 || c.ID == a.ID {
		t.Fatalf("carol must not get alice's claim back: %d %+v", r.Code, c)
	}
}

// ---- headers ----

func TestSecurityHeadersOnEveryResponse(t *testing.T) {
	h, _, _ := setup(nil)
	for _, rq := range []struct{ method, path string }{
		{"GET", "/healthz"}, {"GET", "/claims"}, {"GET", "/nope"}, {"DELETE", "/claims/x"},
	} {
		r := do(h, rq.method, rq.path, "", map[string]string{"Authorization": ""})
		for k, want := range map[string]string{
			"X-Content-Type-Options":  "nosniff",
			"Cache-Control":           "no-store",
			"X-Frame-Options":         "DENY",
			"Referrer-Policy":         "no-referrer",
			"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
		} {
			if r.Header().Get(k) != want {
				t.Errorf("%s %s: %s = %q", rq.method, rq.path, k, r.Header().Get(k))
			}
		}
	}
}

func TestRequestAndTraceHeadersAreHandled(t *testing.T) {
	newMetrics(t)
	h, _, _ := setup(nil)
	r := do(h, "GET", "/healthz", "", map[string]string{"Authorization": "", "X-Request-Id": "abc-1"})
	if r.Header().Get("X-Request-Id") != "abc-1" {
		t.Fatal("request id not echoed")
	}
}

func TestHTTPMetricsUseRouteTemplates(t *testing.T) {
	m := newMetrics(t)
	h, _, _ := setup(nil)
	c := submitAs(t, h, "alice", "k1")
	do(h, "GET", "/claims/"+c.ID, "", bearer("alice", "user"))
	do(h, "GET", "/claims/00000000-0000-4000-8000-0000000000ff", "", bearer("alice", "user"))
	if m.HTTPRequests.Value("GET", "/claims/{id}", "200") != 1 || m.HTTPRequests.Value("GET", "/claims/{id}", "404") != 1 {
		t.Fatal("per-id URLs must collapse into /claims/{id}")
	}
	if m.HTTPRequests.Value("POST", "/claims", "201") != 1 {
		t.Fatal("POST not counted")
	}
	if m.ClaimsSubmitted.Value("created") != 1 {
		t.Fatal("business counter not updated")
	}
}
