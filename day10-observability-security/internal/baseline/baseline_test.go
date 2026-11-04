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
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/http"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/auth"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/memory"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/port"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/service"
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
	ver, _ := auth.NewVerifier([]byte("baseline-secret-baseline-secret-0000"), "claim-tracker", "claims-api")
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
		"alg": "HS256 only", "min_secret_bytes": auth.MinSecretLen,
		"roles": []string{auth.RoleUser, auth.RoleStaff},
	}

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
