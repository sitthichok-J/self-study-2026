// Package baseline snapshots current behaviour into testdata/baseline.json.
// Run `go test ./internal/baseline -update` to regenerate after an intended change.
package baseline

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/adapter/http"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/auth"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/auth/authtest"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/memory"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/port"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/service"
)

var update = flag.Bool("update", false, "rewrite testdata/baseline.json")

const file = "../../testdata/baseline.json"

func snapshot() map[string]any {
	out := map[string]any{}

	out["topics"] = []string{port.TopicSubmitted, port.TopicDLQ}
	out["limits"] = map[string]any{
		"over_limit_satang": domain.OverLimit, "max_amount_satang": domain.MaxAmount,
		"default_page": service.DefaultLimit, "max_page": service.MaxLimit,
		"claim_cache_ttl": service.ClaimTTL.String(), "list_cache_ttl": service.ListTTL.String(),
	}

	// decision table
	dec := map[string]string{}
	for _, c := range []struct {
		ref string
		amt int64
	}{{"FAKE-AB12", 1}, {"FAKE-AB12", domain.OverLimit}, {"FAKE-AB12", domain.OverLimit + 1}, {domain.PoisonRef, 1}} {
		d, err := domain.Evaluate(domain.Claim{PatientRef: c.ref, Amount: c.amt})
		res := string(d.Status) + "/" + d.Reason
		if err != nil {
			res = "error:transient"
		}
		dec[fmt.Sprintf("%s@%d", c.ref, c.amt)] = res
	}
	out["decisions"] = dec

	// transition matrix
	tr := map[string]bool{}
	all := []domain.Status{domain.StatusPending, domain.StatusApproved, domain.StatusRejected}
	for _, a := range all {
		for _, b := range all {
			tr[string(a)+"->"+string(b)] = domain.CanTransition(a, b) == nil
		}
	}
	out["transitions"] = tr

	// HTTP status codes
	q := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewClaims(memory.NewRepo(), memory.NewCache(), q)
	ver, _ := auth.NewVerifier(authtest.KeySet("k1"), "claim-tracker", "claims-api")
	h := http.New(svc, nil, q, ver)
	tok := func(sub, role string) string { t, _ := ver.Mint(sub, role, 10*time.Minute); return t }
	// callAs sends the request with a token for sub ("" = no Authorization header).
	callAs := func(sub, method, path, body, key string) (int, *httptest.ResponseRecorder) {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		if sub != "" {
			req.Header.Set("Authorization", "Bearer "+tok(sub, "user"))
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code, rec
	}
	// The original Day 8 table, now for an authenticated user: unchanged.
	call := func(method, path, body, key string) int {
		c, _ := callAs("alice", method, path, body, key)
		return c
	}
	good := `{"patient_ref":"FAKE-AB12","amount":100}`
	out["http"] = map[string]int{
		"POST create":              call("POST", "/claims", good, "a"),
		"POST replay":              call("POST", "/claims", good, "a"),
		"POST conflict":            call("POST", "/claims", `{"patient_ref":"FAKE-AB12","amount":101}`, "a"),
		"POST no key":              call("POST", "/claims", good, ""),
		"POST real-looking ref":    call("POST", "/claims", `{"patient_ref":"1100100123456","amount":1}`, "b"),
		"GET unknown id":           call("GET", "/claims/00000000-0000-4000-8000-0000000000ff", "", ""),
		"GET bad id":               call("GET", "/claims/x", "", ""),
		"GET list":                 call("GET", "/claims", "", ""),
		"GET list bad status":      call("GET", "/claims?status=x", "", ""),
		"GET list limit too large": call("GET", "/claims?limit=1000", "", ""),
		"GET healthz":              call("GET", "/healthz", "", ""),
		"GET readyz":               call("GET", "/readyz", "", ""),
	}
	_ = context.Background

	// ---- Day 10 additions ----
	var created struct {
		ID string `json:"id"`
	}
	_, rec := callAs("alice", "POST", "/claims", good, "day10-key")
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	sec := map[string]int{}
	sec["GET list without token"], _ = callAs("", "GET", "/claims", "", "")
	sec["POST without token"], _ = callAs("", "POST", "/claims", good, "z")
	sec["GET own claim"], _ = callAs("alice", "GET", "/claims/"+created.ID, "", "")
	sec["GET other user's claim"], _ = callAs("bob", "GET", "/claims/"+created.ID, "", "")
	sec["GET healthz without token"], _ = callAs("", "GET", "/healthz", "", "")
	sec["GET /metrics on api port"], _ = callAs("alice", "GET", "/metrics", "", "")
	sec["same idempotency key, other user"], _ = callAs("bob", "POST", "/claims", good, "day10-key")
	out["security_http"] = sec
	_, hdr := callAs("", "GET", "/healthz", "", "")
	hs := map[string]string{}
	for _, k := range []string{"X-Content-Type-Options", "Cache-Control", "Content-Security-Policy", "X-Frame-Options", "Referrer-Policy"} {
		hs[k] = hdr.Header().Get(k)
	}
	out["security_headers"] = hs
	out["jwt"] = map[string]any{
		"alg": "RS256 only", "min_rsa_bits": auth.MinRSABits,
		"roles": []string{auth.RoleUser, auth.RoleStaff},
	}

	// ---- Day 11 additions: login / refresh / logout ----
	out["auth_http"] = authSnapshot(q)

	// metric families are a contract with Prometheus rules and the dashboard
	var mb strings.Builder
	obs.NewMetrics().Reg.Write(&mb)
	fams := map[string]string{}
	for _, l := range strings.Split(mb.String(), "\n") {
		if f := strings.Fields(l); len(f) == 4 && f[0] == "#" && f[1] == "TYPE" {
			fams[f[2]] = f[3]
		}
	}
	out["metric_families"] = fams
	return out
}

// authSnapshot records the observable behaviour of the session routes:
// status codes, cookie attributes and the response shape. Tokens themselves are random and never recorded.
func authSnapshot(q *slog.Logger) map[string]any {
	ver, _ := auth.NewVerifier(authtest.KeySet("k1"), "claim-tracker", "claims-api")
	fast := auth.Hasher{Iterations: auth.MinIterations}
	st := auth.NewMemStore()
	sess, _ := auth.NewService(st, ver, fast, 10*time.Minute, 12*time.Hour)
	_ = sess.SetUser(context.Background(), fast, "alice", "alice-fake-password-1", auth.RoleUser)
	svc := service.NewClaims(memory.NewRepo(), memory.NewCache(), q)
	h := http.NewWithSessions(svc, nil, q, ver, sess, http.CookieConfig{Secure: true})

	type resp struct {
		code   int
		rec    *httptest.ResponseRecorder
		cookie *nethttp.Cookie
	}
	call := func(method, path, body, refresh string, csrf bool) resp {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if csrf {
			req.Header.Set(http.CSRFHeader, http.CSRFValue)
		}
		if refresh != "" {
			req.AddCookie(&nethttp.Cookie{Name: "refresh_token", Value: refresh})
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		var c *nethttp.Cookie
		for _, k := range rec.Result().Cookies() {
			if k.Name == "refresh_token" {
				c = k
			}
		}
		return resp{rec.Code, rec, c}
	}
	good := `{"username":"alice","password":"alice-fake-password-1"}`
	login := call("POST", "/auth/login", good, "", false)
	var body map[string]any
	_ = json.Unmarshal(login.rec.Body.Bytes(), &body)
	keys := []string{}
	for k := range body {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	r1 := login.cookie.Value
	refresh := call("POST", "/auth/refresh", "", r1, true)
	r2 := refresh.cookie.Value
	reuse := call("POST", "/auth/refresh", "", r1, true)
	afterReuse := call("POST", "/auth/refresh", "", r2, true)
	l2 := call("POST", "/auth/login", good, "", false)
	logout := call("POST", "/auth/logout", "", l2.cookie.Value, true)
	afterLogout := call("POST", "/auth/refresh", "", l2.cookie.Value, true)

	return map[string]any{
		"status": map[string]int{
			"login ok":                      login.code,
			"login wrong password":          call("POST", "/auth/login", `{"username":"alice","password":"wrong-password-123"}`, "", false).code,
			"login unknown user":            call("POST", "/auth/login", `{"username":"zed","password":"wrong-password-123"}`, "", false).code,
			"login bad body":                call("POST", "/auth/login", `{"username":"alice","password":"x","role":"staff"}`, "", false).code,
			"login GET":                     call("GET", "/auth/login", "", "", false).code,
			"refresh ok":                    refresh.code,
			"refresh reused token":          reuse.code,
			"refresh after reuse (new one)": afterReuse.code,
			"refresh without csrf header":   call("POST", "/auth/refresh", "", r2, false).code,
			"refresh without cookie":        call("POST", "/auth/refresh", "", "", true).code,
			"logout ok":                     logout.code,
			"refresh after logout":          afterLogout.code,
			"logout without csrf header":    call("POST", "/auth/logout", "", l2.cookie.Value, false).code,
		},
		"login_response_keys": keys,
		"token_type":          body["token_type"],
		"expires_in":          body["expires_in"],
		"refresh_cookie": map[string]any{
			"name": login.cookie.Name, "http_only": login.cookie.HttpOnly, "secure": login.cookie.Secure,
			"same_site": fmt.Sprint(login.cookie.SameSite), "path": login.cookie.Path, "max_age": login.cookie.MaxAge,
		},
		"csrf_header":        http.CSRFHeader + "=" + http.CSRFValue,
		"password_hash_form": "pbkdf2-sha256, min iterations " + fmt.Sprint(auth.DefaultIterations),
		"password_min_len":   auth.MinPasswordLen,
	}
}

func TestBaseline(t *testing.T) {
	got := snapshot()
	if *update {
		b, _ := json.MarshalIndent(got, "", "  ")
		if err := os.WriteFile(file, append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("missing baseline (run with -update): %v", err)
	}
	var want map[string]any
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(got)
	var gotNorm map[string]any
	_ = json.Unmarshal(b, &gotNorm)
	if !reflect.DeepEqual(gotNorm, want) {
		g, _ := json.MarshalIndent(gotNorm, "", "  ")
		w, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("behaviour changed vs baseline.\n--- want\n%s\n--- got\n%s\n(if intended: go test ./internal/baseline -update)", w, g)
	}
}
