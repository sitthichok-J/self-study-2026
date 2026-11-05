// Package obs is a small, dependency-free observability kit: Prometheus text
// metrics, W3C trace context + OTLP/HTTP JSON export, and slog helpers.
//
// It is deliberately NOT prometheus/client_golang or the OpenTelemetry SDK
// (those modules could not be downloaded in the sandbox where this was
// written). The wire formats are the same, so Prometheus and Jaeger read it.
package obs

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type collector interface {
	name() string
	write(w io.Writer)
}

// Registry holds metric families and renders them in Prometheus text format.
type Registry struct {
	mu   sync.Mutex
	cols map[string]collector
}

func NewRegistry() *Registry { return &Registry{cols: map[string]collector{}} }

func (r *Registry) add(c collector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.cols[c.name()]; dup {
		panic("obs: duplicate metric " + c.name())
	}
	r.cols[c.name()] = c
}

// Write renders all families sorted by name (stable output).
func (r *Registry) Write(w io.Writer) {
	r.mu.Lock()
	names := make([]string, 0, len(r.cols))
	for n := range r.cols {
		names = append(names, n)
	}
	sort.Strings(names)
	cs := make([]collector, 0, len(names))
	for _, n := range names {
		cs = append(cs, r.cols[n])
	}
	r.mu.Unlock()
	for _, c := range cs {
		c.write(w)
	}
}

// Handler serves GET /metrics.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		r.Write(w)
	})
}

func validName(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		ok := c == '_' || c == ':' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (i > 0 && c >= '0' && c <= '9')
		if !ok {
			return false
		}
	}
	return true
}

func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return strings.ReplaceAll(s, `"`, `\"`)
}

func labelString(names, vals []string, extra ...string) string {
	if len(names) == 0 && len(extra) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for i, n := range names {
		if !first {
			b.WriteByte(',')
		}
		first = false
		fmt.Fprintf(&b, `%s="%s"`, n, escapeLabel(vals[i]))
	}
	for i := 0; i+1 < len(extra); i += 2 {
		if !first {
			b.WriteByte(',')
		}
		first = false
		fmt.Fprintf(&b, `%s="%s"`, extra[i], escapeLabel(extra[i+1]))
	}
	b.WriteByte('}')
	return b.String()
}

func fmtFloat(f float64) string {
	if math.IsInf(f, 1) {
		return "+Inf"
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// ---- counter ----

type CounterVec struct {
	n, help string
	labels  []string
	mu      sync.Mutex
	vals    map[string]*counterSeries
}

type counterSeries struct {
	lv []string
	v  float64
}

func (r *Registry) NewCounterVec(name, help string, labels ...string) *CounterVec {
	if !validName(name) {
		panic("obs: bad metric name " + name)
	}
	c := &CounterVec{n: name, help: help, labels: labels, vals: map[string]*counterSeries{}}
	r.add(c)
	return c
}

func (c *CounterVec) name() string { return c.n }

// Inc adds 1 for the given label values (must match the label count).
func (c *CounterVec) Inc(lv ...string) { c.Add(1, lv...) }

func (c *CounterVec) Add(d float64, lv ...string) {
	if len(lv) != len(c.labels) {
		panic(fmt.Sprintf("obs: %s wants %d label values, got %d", c.n, len(c.labels), len(lv)))
	}
	if d < 0 {
		return // counters never go down
	}
	k := strings.Join(lv, "\xff")
	c.mu.Lock()
	s, ok := c.vals[k]
	if !ok {
		s = &counterSeries{lv: append([]string(nil), lv...)}
		c.vals[k] = s
	}
	s.v += d
	c.mu.Unlock()
}

// Value is for tests.
func (c *CounterVec) Value(lv ...string) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s, ok := c.vals[strings.Join(lv, "\xff")]; ok {
		return s.v
	}
	return 0
}

func (c *CounterVec) write(w io.Writer) {
	fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s counter\n", c.n, c.help, c.n)
	c.mu.Lock()
	keys := make([]string, 0, len(c.vals))
	for k := range c.vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s := c.vals[k]
		fmt.Fprintf(w, "%s%s %s\n", c.n, labelString(c.labels, s.lv), fmtFloat(s.v))
	}
	c.mu.Unlock()
}

// ---- histogram ----

var DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

type HistogramVec struct {
	n, help string
	labels  []string
	buckets []float64
	mu      sync.Mutex
	vals    map[string]*histSeries
}

type histSeries struct {
	lv     []string
	counts []uint64 // per bucket, non-cumulative
	inf    uint64
	sum    float64
	n      uint64
}

func (r *Registry) NewHistogramVec(name, help string, buckets []float64, labels ...string) *HistogramVec {
	if !validName(name) {
		panic("obs: bad metric name " + name)
	}
	if len(buckets) == 0 {
		buckets = DefBuckets
	}
	if !sort.Float64sAreSorted(buckets) {
		panic("obs: buckets must be sorted")
	}
	h := &HistogramVec{n: name, help: help, labels: labels, buckets: buckets, vals: map[string]*histSeries{}}
	r.add(h)
	return h
}

func (h *HistogramVec) name() string { return h.n }

func (h *HistogramVec) Observe(v float64, lv ...string) {
	if len(lv) != len(h.labels) {
		panic(fmt.Sprintf("obs: %s wants %d label values, got %d", h.n, len(h.labels), len(lv)))
	}
	k := strings.Join(lv, "\xff")
	h.mu.Lock()
	s, ok := h.vals[k]
	if !ok {
		s = &histSeries{lv: append([]string(nil), lv...), counts: make([]uint64, len(h.buckets))}
		h.vals[k] = s
	}
	i := sort.SearchFloat64s(h.buckets, v) // first bucket with bound >= v
	if i < len(h.buckets) {
		s.counts[i]++
	} else {
		s.inf++
	}
	s.sum += v
	s.n++
	h.mu.Unlock()
}

// Count is for tests.
func (h *HistogramVec) Count(lv ...string) uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s, ok := h.vals[strings.Join(lv, "\xff")]; ok {
		return s.n
	}
	return 0
}

func (h *HistogramVec) write(w io.Writer) {
	fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s histogram\n", h.n, h.help, h.n)
	h.mu.Lock()
	keys := make([]string, 0, len(h.vals))
	for k := range h.vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s := h.vals[k]
		var cum uint64
		for i, b := range h.buckets {
			cum += s.counts[i]
			fmt.Fprintf(w, "%s_bucket%s %d\n", h.n, labelString(h.labels, s.lv, "le", fmtFloat(b)), cum)
		}
		cum += s.inf
		fmt.Fprintf(w, "%s_bucket%s %d\n", h.n, labelString(h.labels, s.lv, "le", "+Inf"), cum)
		fmt.Fprintf(w, "%s_sum%s %s\n", h.n, labelString(h.labels, s.lv), fmtFloat(s.sum))
		fmt.Fprintf(w, "%s_count%s %d\n", h.n, labelString(h.labels, s.lv), s.n)
	}
	h.mu.Unlock()
}

// ---- gauge from a function ----

type gaugeFunc struct {
	n, help string
	f       func() float64
}

func (g *gaugeFunc) name() string { return g.n }
func (g *gaugeFunc) write(w io.Writer) {
	fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s gauge\n%s %s\n", g.n, g.help, g.n, g.n, fmtFloat(g.f()))
}

func (r *Registry) NewGaugeFunc(name, help string, f func() float64) {
	if !validName(name) {
		panic("obs: bad metric name " + name)
	}
	r.add(&gaugeFunc{n: name, help: help, f: f})
}

// RegisterRuntime adds goroutine / heap gauges (the "U" of USE for the process).
func (r *Registry) RegisterRuntime() {
	r.NewGaugeFunc("go_goroutines", "Number of goroutines.", func() float64 { return float64(runtime.NumGoroutine()) })
	r.NewGaugeFunc("go_heap_alloc_bytes", "Bytes of allocated heap objects.", func() float64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return float64(m.HeapAlloc)
	})
}
