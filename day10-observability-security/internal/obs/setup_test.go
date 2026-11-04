package obs

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsMuxServesMetricsAndHealth(t *testing.T) {
	m := NewMetrics()
	m.ClaimsSubmitted.Inc("created")
	h := MetricsMux(m.Reg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `claims_submitted_total{result="created"} 1`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/claims", nil))
	if rec.Code != 404 {
		t.Fatalf("the metrics listener must serve nothing else, got %d", rec.Code)
	}
}

func TestSetupTracingWithoutEndpointStillGivesIDs(t *testing.T) {
	old := SetTracer(NewTracer("x", nil))
	defer SetTracer(old)
	stop := SetupTracing("svc", "")
	ctx, sp := Start(context.Background(), "a")
	if TraceIDFromContext(ctx) == "" || len(TraceparentFromContext(ctx)) != 55 {
		t.Fatal("ids are needed for log correlation even when nothing is exported")
	}
	sp.End()
	if err := stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSetupTracingWithEndpointShutsDown(t *testing.T) {
	old := SetTracer(NewTracer("x", nil))
	defer SetTracer(old)
	stop := SetupTracing("svc", "http://127.0.0.1:1") // nothing listens; export errors are counted, not fatal
	_, sp := Start(context.Background(), "a")
	sp.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := stop(ctx); err != nil {
		t.Fatal(err)
	}
}
