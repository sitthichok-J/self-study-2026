package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/auth"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/auth/authtest"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/memory"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/service"
)

var testVerifier = func() *auth.Verifier {
	v, err := auth.NewVerifier(authtest.KeySet("k1"), "claim-tracker", "claims-api")
	if err != nil {
		panic(err)
	}
	return v
}()

func tokenFor(sub, role string) string {
	tok, _ := testVerifier.Mint(sub, role, 10*time.Minute)
	return tok
}

func setup(ready func(context.Context) error) (nethttp.Handler, *service.Claims, *memory.Repo) {
	q := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := memory.NewRepo()
	svc := service.NewClaims(repo, memory.NewCache(), q)
	return New(svc, ready, q, testVerifier), svc, repo
}

// do sends the request as "alice" unless hdr sets Authorization itself
// (an empty value means: send no Authorization header at all).
func do(h nethttp.Handler, method, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if _, ok := hdr["Authorization"]; !ok {
		req.Header.Set("Authorization", "Bearer "+tokenFor("alice", "user"))
	}
	for k, v := range hdr {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func idem(k string) map[string]string { return map[string]string{"Idempotency-Key": k} }

func TestPostCreatesThenReplays(t *testing.T) {
	h, _, repo := setup(nil)
	body := `{"patient_ref":"FAKE-AB12","amount":1500}`
	r1 := do(h, "POST", "/claims", body, idem("k1"))
	if r1.Code != 201 || r1.Header().Get("Location") == "" {
		t.Fatalf("first: %d %s", r1.Code, r1.Body)
	}
	r2 := do(h, "POST", "/claims", body, idem("k1"))
	if r2.Code != 200 || r2.Header().Get("Idempotent-Replayed") != "true" {
		t.Fatalf("replay: %d hdr=%v", r2.Code, r2.Header())
	}
	var a, b domain.Claim
	_ = json.Unmarshal(r1.Body.Bytes(), &a)
	_ = json.Unmarshal(r2.Body.Bytes(), &b)
	if a.ID != b.ID || repo.Count() != 1 {
		t.Fatal("replay must return the same claim")
	}
}

func TestPostErrors(t *testing.T) {
	h, _, _ := setup(nil)
	good := `{"patient_ref":"FAKE-AB12","amount":1500}`
	cases := []struct {
		name, body string
		hdr        map[string]string
		code       int
	}{
		{"missing key", good, nil, 400},
		{"bad json", `{`, idem("k"), 400},
		{"unknown field", `{"patient_ref":"FAKE-AB12","amount":5,"x":1}`, idem("k"), 400},
		{"real id", `{"patient_ref":"1234567890123","amount":5}`, idem("k"), 400},
		{"zero amount", `{"patient_ref":"FAKE-AB12","amount":0}`, idem("k"), 400},
		{"float amount", `{"patient_ref":"FAKE-AB12","amount":1.5}`, idem("k"), 400},
		{"empty body", ``, idem("k"), 400},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if r := do(h, "POST", "/claims", c.body, c.hdr); r.Code != c.code {
				t.Fatalf("code %d body %s", r.Code, r.Body)
			}
		})
	}
}

func TestPostConflict(t *testing.T) {
	h, _, _ := setup(nil)
	do(h, "POST", "/claims", `{"patient_ref":"FAKE-AB12","amount":1}`, idem("k1"))
	r := do(h, "POST", "/claims", `{"patient_ref":"FAKE-AB12","amount":2}`, idem("k1"))
	if r.Code != 409 {
		t.Fatalf("code %d", r.Code)
	}
}

func TestPostBodyTooLarge(t *testing.T) {
	h, _, _ := setup(nil)
	big := `{"patient_ref":"` + strings.Repeat("A", 2<<20) + `","amount":1}`
	if r := do(h, "POST", "/claims", big, idem("k")); r.Code != 400 {
		t.Fatalf("code %d", r.Code)
	}
}

func TestGet(t *testing.T) {
	h, svc, _ := setup(nil)
	c, _, _ := svc.Submit(context.Background(), service.SubmitInput{Owner: "alice", IdempotencyKey: "k", PatientRef: "FAKE-AB12", Amount: 9})
	if r := do(h, "GET", "/claims/"+c.ID, "", nil); r.Code != 200 || !strings.Contains(r.Body.String(), `"PENDING"`) {
		t.Fatalf("%d %s", r.Code, r.Body)
	}
	if r := do(h, "GET", "/claims/00000000-0000-4000-8000-0000000000ff", "", nil); r.Code != 404 {
		t.Fatalf("missing: %d", r.Code)
	}
	if r := do(h, "GET", "/claims/abc", "", nil); r.Code != 400 {
		t.Fatalf("bad id: %d", r.Code)
	}
}

func TestListQueryValidation(t *testing.T) {
	h, _, _ := setup(nil)
	cases := map[string]int{
		"/claims":                         200,
		"/claims?status=PENDING":          200,
		"/claims?status=pending":          400,
		"/claims?status=NOPE":             400,
		"/claims?limit=abc":               400,
		"/claims?limit=101":               400,
		"/claims?offset=-1":               400,
		"/claims?limit=5&offset=10":       200,
		"/claims?status=APPROVED&limit=1": 200,
	}
	for path, want := range cases {
		if r := do(h, "GET", path, "", nil); r.Code != want {
			t.Errorf("%s: got %d want %d (%s)", path, r.Code, want, r.Body)
		}
	}
}

func TestListShape(t *testing.T) {
	h, _, _ := setup(nil)
	r := do(h, "GET", "/claims", "", nil)
	var out struct {
		Items  []domain.Claim `json:"items"`
		Limit  int            `json:"limit"`
		Offset int            `json:"offset"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &out); err != nil || out.Items == nil || out.Limit != 20 {
		t.Fatalf("%s %v", r.Body, err)
	}
}

func TestHealthAndReady(t *testing.T) {
	h, _, _ := setup(nil)
	if r := do(h, "GET", "/healthz", "", nil); r.Code != 200 {
		t.Fatal(r.Code)
	}
	if r := do(h, "GET", "/readyz", "", nil); r.Code != 200 {
		t.Fatal(r.Code)
	}
	h2, _, _ := setup(func(context.Context) error { return errors.New("db down") })
	if r := do(h2, "GET", "/readyz", "", nil); r.Code != 503 {
		t.Fatalf("not-ready should be 503, got %d", r.Code)
	}
	if r := do(h2, "GET", "/healthz", "", nil); r.Code != 200 {
		t.Fatal("liveness must not depend on the DB")
	}
}

func TestInternalErrorDoesNotLeakDetails(t *testing.T) {
	h, _, repo := setup(nil)
	repo.FailGet = errors.New("pq: password authentication failed for user secret_user")
	r := do(h, "GET", "/claims/00000000-0000-4000-8000-0000000000ff", "", nil)
	if r.Code != 500 || strings.Contains(r.Body.String(), "secret_user") {
		t.Fatalf("%d %s", r.Code, r.Body)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h, _, _ := setup(nil)
	if r := do(h, "DELETE", "/claims/x", "", nil); r.Code != 405 {
		t.Fatalf("got %d", r.Code)
	}
}
