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

	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/adapter/http"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/memory"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/port"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/service"
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
	h := http.New(svc, nil, q)
	call := func(method, path, body, key string) int {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
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
