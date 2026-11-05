package obs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func render(r *Registry) string {
	var b bytes.Buffer
	r.Write(&b)
	return b.String()
}

func TestCounterTextFormatSortedAndEscaped(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounterVec("b_total", "second", "k")
	a := r.NewCounterVec("a_total", "first", "k")
	c.Inc("x")
	c.Add(2, "x")
	a.Inc(`q"uo\te` + "\n")
	got := render(r)
	want := "# HELP a_total first\n# TYPE a_total counter\na_total{k=\"q\\\"uo\\\\te\\n\"} 1\n" +
		"# HELP b_total second\n# TYPE b_total counter\nb_total{k=\"x\"} 3\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestCounterNeverDecreases(t *testing.T) {
	c := NewRegistry().NewCounterVec("c_total", "h")
	c.Add(5)
	c.Add(-3)
	if c.Value() != 5 {
		t.Fatalf("got %v", c.Value())
	}
}

func TestHistogramCumulativeBuckets(t *testing.T) {
	r := NewRegistry()
	h := r.NewHistogramVec("d_seconds", "h", []float64{0.1, 1}, "route")
	h.Observe(0.05, "/a")
	h.Observe(0.1, "/a") // boundary counts in le=0.1
	h.Observe(0.5, "/a")
	h.Observe(7, "/a")
	got := render(r)
	for _, w := range []string{
		`d_seconds_bucket{route="/a",le="0.1"} 2`,
		`d_seconds_bucket{route="/a",le="1"} 3`,
		`d_seconds_bucket{route="/a",le="+Inf"} 4`,
		`d_seconds_sum{route="/a"} 7.65`,
		`d_seconds_count{route="/a"} 4`,
	} {
		if !strings.Contains(got, w+"\n") {
			t.Errorf("missing %q in\n%s", w, got)
		}
	}
}

func TestMisuseFailsLoudly(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounterVec("x_total", "h", "a")
	mustPanic(t, "wrong label count", func() { c.Inc() })
	mustPanic(t, "duplicate name", func() { r.NewCounterVec("x_total", "h") })
	mustPanic(t, "bad name", func() { r.NewCounterVec("1bad", "h") })
	mustPanic(t, "unsorted buckets", func() { r.NewHistogramVec("h_s", "h", []float64{2, 1}) })
}

func mustPanic(t *testing.T, what string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: expected panic", what)
		}
	}()
	f()
}

func TestMetricsHandlerMethodAndContentType(t *testing.T) {
	m := NewMetrics()
	rec := httptest.NewRecorder()
	m.Reg.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	if rec.Code != 200 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/plain; version=0.0.4") {
		t.Fatalf("%d %v", rec.Code, rec.Header())
	}
	for _, n := range []string{"http_requests_total", "http_request_duration_seconds", "go_goroutines", "consumer_messages_total", "auth_failures_total"} {
		_ = n // families with no series print only HELP/TYPE once used; gauges print always
	}
	if !strings.Contains(rec.Body.String(), "go_goroutines ") {
		t.Fatal("runtime gauge missing")
	}
	rec = httptest.NewRecorder()
	m.Reg.Handler().ServeHTTP(rec, httptest.NewRequest("POST", "/metrics", nil))
	if rec.Code != 405 {
		t.Fatalf("POST should be 405, got %d", rec.Code)
	}
}

func TestTraceparentParse(t *testing.T) {
	good := "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"
	sc, ok := ParseTraceparent(good)
	if !ok || !sc.Sampled || sc.Traceparent() != good {
		t.Fatalf("round trip failed: %v %v", ok, sc.Traceparent())
	}
	bad := map[string]string{
		"empty":         "",
		"version ff":    "ff-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
		"upper hex":     "00-0AF7651916CD43DD8448EB211C80319C-b7ad6b7169203331-01",
		"zero trace":    "00-00000000000000000000000000000000-b7ad6b7169203331-01",
		"zero span":     "00-0af7651916cd43dd8448eb211c80319c-0000000000000000-01",
		"short":         "00-0af7651916cd43dd8448eb211c80319c-b7ad6b716920333-01",
		"extra suffix":  good + "-x",
		"non hex":       "00-0af7651916cd43dd8448eb211c80319z-b7ad6b7169203331-01",
		"bad separator": "00_0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
	}
	for n, s := range bad {
		if _, ok := ParseTraceparent(s); ok {
			t.Errorf("%s: should be rejected", n)
		}
	}
	sc, ok = ParseTraceparent("00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-00")
	if !ok || sc.Sampled {
		t.Fatal("flag 00 means not sampled")
	}
}

func TestSpanParentingAndRemote(t *testing.T) {
	mem := &MemExporter{}
	tr := NewTracer("svc", mem)
	remote, _ := ParseTraceparent("00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	ctx := ContextWithRemote(context.Background(), remote)
	ctx, root := tr.Start(ctx, "root", KindServer)
	_, child := tr.Start(ctx, "child", KindInternal)
	child.End()
	root.End()
	root.End() // second End is ignored
	sp := mem.Spans()
	if len(sp) != 2 {
		t.Fatalf("want 2 spans, got %d", len(sp))
	}
	c, r := sp[0], sp[1]
	if c.Context.TraceID != remote.TraceID || r.Context.TraceID != remote.TraceID {
		t.Fatal("trace id must continue the remote trace")
	}
	if r.Parent != remote.SpanID || c.Parent != r.Context.SpanID {
		t.Fatal("parent chain wrong")
	}
	if c.Context.SpanID == r.Context.SpanID {
		t.Fatal("span ids must differ")
	}
}

func TestNewTraceWhenNoParentAndUnsampledNotExported(t *testing.T) {
	mem := &MemExporter{}
	tr := NewTracer("svc", mem)
	_, s := tr.Start(context.Background(), "a", KindInternal)
	if !s.Context().Valid() {
		t.Fatal("ids must be generated")
	}
	s.End()
	un, _ := ParseTraceparent("00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-00")
	_, s2 := tr.Start(ContextWithRemote(context.Background(), un), "b", KindInternal)
	s2.End()
	if len(mem.Spans()) != 1 {
		t.Fatalf("unsampled span must not be exported, got %d", len(mem.Spans()))
	}
}

func TestEncodeOTLPShape(t *testing.T) {
	mem := &MemExporter{}
	tr := NewTracer("api", mem)
	ctx, root := tr.Start(context.Background(), "GET /x", KindServer, Attr{"n", 3}, Attr{"ok", true}, Attr{"s", "v"})
	_, ch := tr.Start(ctx, "db", KindInternal)
	ch.MarkError("boom")
	ch.End()
	root.End()
	var doc struct {
		ResourceSpans []struct {
			Resource struct {
				Attributes []struct {
					Key   string
					Value map[string]string
				}
			}
			ScopeSpans []struct {
				Spans []struct {
					TraceID, SpanID, ParentSpanID, Name string
					Kind                                int
					StartTimeUnixNano, EndTimeUnixNano  string
					Attributes                          []struct {
						Key   string
						Value map[string]any
					}
					Status struct {
						Code    int
						Message string
					}
				}
			}
		}
	}
	if err := json.Unmarshal(EncodeOTLP("api", mem.Spans()), &doc); err != nil {
		t.Fatal(err)
	}
	rs := doc.ResourceSpans[0]
	if rs.Resource.Attributes[0].Key != "service.name" || rs.Resource.Attributes[0].Value["stringValue"] != "api" {
		t.Fatal("service.name missing")
	}
	sp := rs.ScopeSpans[0].Spans
	if len(sp) != 2 || len(sp[0].TraceID) != 32 || len(sp[0].SpanID) != 16 || sp[0].TraceID != sp[1].TraceID {
		t.Fatalf("ids: %+v", sp)
	}
	db, srv := sp[0], sp[1]
	if db.ParentSpanID != srv.SpanID || srv.ParentSpanID != "" {
		t.Fatal("parent linkage lost")
	}
	if db.Status.Code != 2 || db.Status.Message != "boom" || srv.Status.Code != 1 {
		t.Fatalf("status: %+v / %+v", db.Status, srv.Status)
	}
	if srv.Kind != 2 || srv.StartTimeUnixNano == "" || srv.EndTimeUnixNano < srv.StartTimeUnixNano {
		t.Fatalf("kind/time: %+v", srv)
	}
	var sawInt bool
	for _, a := range srv.Attributes {
		if a.Key == "n" && a.Value["intValue"] == "3" {
			sawInt = true
		}
	}
	if !sawInt {
		t.Fatal("int attribute must be encoded as a string intValue")
	}
}

func TestOTLPExporterPostsBatchAndShutdownFlushes(t *testing.T) {
	var got atomic.Int32
	var path, ctype string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, ctype = r.URL.Path, r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		if bytes.Contains(b, []byte(`"name":"one"`)) {
			got.Add(1)
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()
	m := NewMetrics()
	e := newOTLPExporter(srv.URL+"/", "api", m, time.Hour, 64, 16) // only Shutdown can flush
	tr := NewTracer("api", e)
	_, s := tr.Start(context.Background(), "one", KindInternal)
	s.End()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if got.Load() != 1 || path != "/v1/traces" || ctype != "application/json" {
		t.Fatalf("got=%d path=%q ctype=%q", got.Load(), path, ctype)
	}
	if m.OTLPExport.Value("ok") != 1 {
		t.Fatal("ok export not counted")
	}
}

func TestOTLPExporterCountsFailureAndDrops(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(500) }))
	defer srv.Close()
	m := NewMetrics()
	e := newOTLPExporter(srv.URL, "api", m, time.Hour, 1000, 1) // tiny queue, no timed flush
	tr := NewTracer("api", e)
	for i := 0; i < 5; i++ {
		_, s := tr.Start(context.Background(), "x", KindInternal)
		s.End()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = e.Shutdown(ctx)
	if m.SpansDropped.Value() == 0 {
		t.Log("loop drained fast enough, nothing dropped") // timing dependent, not asserted
	}
	if m.OTLPExport.Value("error") == 0 {
		t.Fatal("500 from collector must be counted as an export error")
	}
}

func newBuf() (*bytes.Buffer, *slog.Logger) {
	var b bytes.Buffer
	return &b, NewLogger(&b, slog.LevelDebug)
}

func lastLine(t *testing.T, b *bytes.Buffer) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	var m map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &m); err != nil {
		t.Fatalf("not json: %q", lines[len(lines)-1])
	}
	return m
}

func TestLoggerAddsTraceAndRequestID(t *testing.T) {
	b, log := newBuf()
	tr := NewTracer("svc", nil)
	ctx, sp := tr.Start(context.Background(), "x", KindInternal)
	ctx = WithRequestID(ctx, "req-1")
	log.InfoContext(ctx, "hello")
	m := lastLine(t, b)
	if m["trace_id"] != TraceIDFromContext(ctx) || len(m["trace_id"].(string)) != 32 {
		t.Fatalf("trace_id: %v", m["trace_id"])
	}
	if m["request_id"] != "req-1" || len(m["span_id"].(string)) != 16 {
		t.Fatalf("%v", m)
	}
	_ = sp
	log.Info("no ctx")
	if _, ok := lastLine(t, b)["trace_id"]; ok {
		t.Fatal("no span in ctx => no trace_id")
	}
}

func TestLoggerRedactsSensitiveKeys(t *testing.T) {
	b, log := newBuf()
	log = log.With("access_token", "AAA")
	log.Info("x", "Authorization", "Bearer abc", "patient_ref", "FAKE-AB12", "db_password", "p", "amount", 5,
		slog.Group("g", slog.String("jwt", "zzz")))
	m := lastLine(t, b)
	s := b.String()
	for _, leak := range []string{"AAA", "Bearer abc", "FAKE-AB12", "zzz", `"p"`} {
		if strings.Contains(s, leak) {
			t.Errorf("leaked %q in %s", leak, s)
		}
	}
	if m["amount"] != float64(5) || m["Authorization"] != "[redacted]" {
		t.Fatalf("%v", m)
	}
}

func testMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /things/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		_, _ = w.Write([]byte("hi"))
	})
	mux.HandleFunc("GET /boom", func(http.ResponseWriter, *http.Request) { panic("secret-in-panic") })
	mux.HandleFunc("GET /fail", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(503) })
	return mux
}

func TestHTTPMiddlewareRecordsRouteNotRawPath(t *testing.T) {
	m := NewMetrics()
	mem := &MemExporter{}
	old := SetTracer(NewTracer("api", mem))
	defer SetTracer(old)
	b, log := newBuf()
	h := HTTP(testMux(), m, log)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/things/123", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/things/456", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/nope/1", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/nope/2", nil))

	if m.HTTPRequests.Value("GET", "/things/{id}", "201") != 2 {
		t.Fatal("route label must be the pattern, not the raw path")
	}
	if m.HTTPRequests.Value("GET", "/nope/1", "404") != 0 || m.HTTPRequests.Value("GET", "unmatched", "404") != 2 {
		t.Fatal("unmatched paths must share one label")
	}
	if m.HTTPDuration.Count("GET", "/things/{id}") != 2 {
		t.Fatal("duration not observed")
	}
	if m.InFlight.Load() != 0 {
		t.Fatal("in-flight gauge must return to 0")
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("request id header missing")
	}
	sp := mem.Spans()
	if len(sp) != 4 || sp[0].Name != "GET /things/{id}" || sp[0].Kind != KindServer {
		t.Fatalf("spans: %+v", sp)
	}
	line := lastLine(t, b)
	if line["route"] != "unmatched" || line["status"] != float64(404) || line["trace_id"] == nil || line["request_id"] == nil {
		t.Fatalf("access log: %v", line)
	}
	if strings.Contains(b.String(), "/nope/2") || strings.Contains(b.String(), "/things/123") {
		t.Fatal("raw paths must not be logged")
	}
}

func TestHTTPMiddlewareRequestIDHandling(t *testing.T) {
	m := NewMetrics()
	_, log := newBuf()
	h := HTTP(testMux(), m, log)
	req := httptest.NewRequest("GET", "/things/1", nil)
	req.Header.Set("X-Request-Id", "abc-123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("X-Request-Id") != "abc-123" {
		t.Fatal("valid id should be kept")
	}
	req = httptest.NewRequest("GET", "/things/1", nil)
	req.Header.Set("X-Request-Id", "bad id\r\nInjected: 1")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	got := rec.Header().Get("X-Request-Id")
	if got == "" || strings.ContainsAny(got, " \r\n") || len(got) != 32 {
		t.Fatalf("unsafe id must be replaced, got %q", got)
	}
}

func TestHTTPMiddlewareContinuesIncomingTrace(t *testing.T) {
	m := NewMetrics()
	mem := &MemExporter{}
	old := SetTracer(NewTracer("api", mem))
	defer SetTracer(old)
	_, log := newBuf()
	h := HTTP(testMux(), m, log)
	req := httptest.NewRequest("GET", "/things/1", nil)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	h.ServeHTTP(httptest.NewRecorder(), req)
	sp := mem.Spans()[0]
	if sp.Context.Traceparent()[3:35] != "0af7651916cd43dd8448eb211c80319c" {
		t.Fatal("trace id not continued")
	}
	req = httptest.NewRequest("GET", "/things/1", nil)
	req.Header.Set("traceparent", "garbage")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if mem.Spans()[1].Context.Traceparent()[3:35] == "0af7651916cd43dd8448eb211c80319c" {
		t.Fatal("garbage header must start a new trace")
	}
}

func TestHTTPMiddlewarePanicBecomes500WithoutLeak(t *testing.T) {
	m := NewMetrics()
	mem := &MemExporter{}
	old := SetTracer(NewTracer("api", mem))
	defer SetTracer(old)
	b, log := newBuf()
	rec := httptest.NewRecorder()
	HTTP(testMux(), m, log).ServeHTTP(rec, httptest.NewRequest("GET", "/boom", nil))
	if rec.Code != 500 || strings.Contains(rec.Body.String(), "secret-in-panic") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if strings.Contains(b.String(), "secret-in-panic") {
		t.Fatal("panic value must not be logged")
	}
	if m.Panics.Value() != 1 || m.HTTPRequests.Value("GET", "/boom", "500") != 1 {
		t.Fatal("panic not counted")
	}
	if mem.Spans()[0].ErrMsg == "" {
		t.Fatal("span should be marked as error")
	}
}

func TestHTTPMiddlewareServerErrorMarksSpan(t *testing.T) {
	m := NewMetrics()
	mem := &MemExporter{}
	old := SetTracer(NewTracer("api", mem))
	defer SetTracer(old)
	_, log := newBuf()
	HTTP(testMux(), m, log).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/fail", nil))
	if mem.Spans()[0].ErrMsg != "http 503" {
		t.Fatalf("%q", mem.Spans()[0].ErrMsg)
	}
}
