package obs

import (
	"sync/atomic"
	"time"
)

// Metrics is the fixed set of application metrics. Names are the contract with
// deploy/prometheus and the Grafana dashboard (tests/test_infra.py checks that).
type Metrics struct {
	Reg *Registry

	HTTPRequests *CounterVec   // method, route, code
	HTTPDuration *HistogramVec // method, route
	InFlight     atomic.Int64

	ClaimsSubmitted *CounterVec // result=created|replayed
	ClaimsDecided   *CounterVec // status, reason
	OutboxPublished *CounterVec // result=ok|error
	ConsumerOutcome *CounterVec // outcome=processed|retried|dlq
	AuthFailures    *CounterVec // reason
	AuthzDenied     *CounterVec // resource
	Panics          *CounterVec
	SpansDropped    *CounterVec
	OTLPExport      *CounterVec // result=ok|error
}

func NewMetrics() *Metrics {
	r := NewRegistry()
	m := &Metrics{Reg: r}
	m.HTTPRequests = r.NewCounterVec("http_requests_total", "HTTP requests by method, route and status code.", "method", "route", "code")
	m.HTTPDuration = r.NewHistogramVec("http_request_duration_seconds", "HTTP request latency.", nil, "method", "route")
	r.NewGaugeFunc("http_requests_in_flight", "Requests being served right now.", func() float64 { return float64(m.InFlight.Load()) })
	m.ClaimsSubmitted = r.NewCounterVec("claims_submitted_total", "Claim submissions.", "result")
	m.ClaimsDecided = r.NewCounterVec("claims_decided_total", "Claims moved out of PENDING by the worker.", "status", "reason")
	m.OutboxPublished = r.NewCounterVec("outbox_publish_total", "Outbox events handed to the broker.", "result")
	m.ConsumerOutcome = r.NewCounterVec("consumer_messages_total", "Messages handled by the worker.", "outcome")
	m.AuthFailures = r.NewCounterVec("auth_failures_total", "Rejected authentication attempts.", "reason")
	m.AuthzDenied = r.NewCounterVec("authz_denied_total", "Requests refused because the caller does not own the resource.", "resource")
	m.Panics = r.NewCounterVec("http_panics_total", "Recovered handler panics.")
	m.SpansDropped = r.NewCounterVec("trace_spans_dropped_total", "Spans dropped because the export queue was full.")
	m.OTLPExport = r.NewCounterVec("trace_export_total", "OTLP export attempts.", "result")
	r.RegisterRuntime()
	m.initSeries()
	return m
}

// initSeries creates the known series at 0. Prometheus' increase()/rate()
// cannot see the first increment of a series that appears already at 1, so
// without this an alert like "dlq > 0" would miss the very first dead letter.
func (m *Metrics) initSeries() {
	for _, v := range []string{"created", "replayed"} {
		m.ClaimsSubmitted.Add(0, v)
	}
	m.ClaimsDecided.Add(0, "APPROVED", "ok")
	m.ClaimsDecided.Add(0, "REJECTED", "over_limit")
	for _, v := range []string{"ok", "error"} {
		m.OutboxPublished.Add(0, v)
		m.OTLPExport.Add(0, v)
	}
	for _, v := range []string{"processed", "retried", "dlq"} {
		m.ConsumerOutcome.Add(0, v)
	}
	m.AuthzDenied.Add(0, "claim")
	m.Panics.Add(0)
	m.SpansDropped.Add(0)
}

var def atomic.Pointer[Metrics]

func init() { def.Store(NewMetrics()) }

// M returns the process-wide metrics.
func M() *Metrics { return def.Load() }

// SetMetrics swaps the process-wide metrics (tests). It returns the old one.
func SetMetrics(m *Metrics) *Metrics { return def.Swap(m) }

// Since converts a start time to seconds for histograms.
func Since(t time.Time) float64 { return time.Since(t).Seconds() }
