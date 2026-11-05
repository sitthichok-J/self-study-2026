package obs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// SpanContext is the part of a span that crosses process boundaries
// (W3C Trace Context, version 00).
type SpanContext struct {
	TraceID [16]byte
	SpanID  [8]byte
	Sampled bool
}

func (sc SpanContext) Valid() bool {
	return sc.TraceID != [16]byte{} && sc.SpanID != [8]byte{}
}

// Traceparent formats "00-<trace>-<span>-<flags>".
func (sc SpanContext) Traceparent() string {
	fl := "00"
	if sc.Sampled {
		fl = "01"
	}
	return "00-" + hex.EncodeToString(sc.TraceID[:]) + "-" + hex.EncodeToString(sc.SpanID[:]) + "-" + fl
}

func isLowerHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// ParseTraceparent is strict: only version 00, exact lengths, lowercase hex,
// non-zero ids. Anything else is ignored by callers (a new trace is started).
func ParseTraceparent(s string) (SpanContext, bool) {
	var sc SpanContext
	if len(s) != 55 || s[2] != '-' || s[35] != '-' || s[52] != '-' {
		return sc, false
	}
	ver, tid, sid, fl := s[0:2], s[3:35], s[36:52], s[53:55]
	if ver != "00" || !isLowerHex(tid) || !isLowerHex(sid) || !isLowerHex(fl) {
		return sc, false
	}
	b, _ := hex.DecodeString(tid)
	copy(sc.TraceID[:], b)
	b, _ = hex.DecodeString(sid)
	copy(sc.SpanID[:], b)
	fb, _ := hex.DecodeString(fl)
	sc.Sampled = fb[0]&1 == 1
	return sc, sc.Valid()
}

type Attr struct {
	Key string
	Val any // string, int, int64, bool
}

type SpanKind int

const (
	KindInternal SpanKind = 1
	KindServer   SpanKind = 2
	KindClient   SpanKind = 3
	KindProducer SpanKind = 4
	KindConsumer SpanKind = 5
)

// SpanData is what exporters receive once a span has ended.
type SpanData struct {
	Name    string
	Kind    SpanKind
	Context SpanContext
	Parent  [8]byte
	Start   time.Time
	End     time.Time
	Attrs   []Attr
	ErrMsg  string // non-empty => status ERROR
	Service string
}

func (d SpanData) Duration() time.Duration { return d.End.Sub(d.Start) }

type Exporter interface{ Export(SpanData) }

type Tracer struct {
	service string
	exp     Exporter // may be nil: ids still exist for logs, nothing is shipped
	now     func() time.Time
}

func NewTracer(service string, exp Exporter) *Tracer {
	return &Tracer{service: service, exp: exp, now: time.Now}
}

type Span struct {
	t     *Tracer
	mu    sync.Mutex
	d     SpanData
	ended bool
}

type ctxKey int

const (
	spanKey ctxKey = iota
	remoteKey
	reqIDKey
)

func newID(b []byte) {
	for {
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}
		for _, x := range b {
			if x != 0 {
				return
			}
		}
	}
}

// ContextWithRemote sets the parent that came in over the wire.
func ContextWithRemote(ctx context.Context, sc SpanContext) context.Context {
	return context.WithValue(ctx, remoteKey, sc)
}

// Start begins a span. Parent = current span in ctx, else remote parent, else new trace.
func (t *Tracer) Start(ctx context.Context, name string, kind SpanKind, attrs ...Attr) (context.Context, *Span) {
	s := &Span{t: t}
	s.d = SpanData{Name: name, Kind: kind, Start: t.now(), Attrs: attrs, Service: t.service}
	switch {
	case SpanFromContext(ctx) != nil:
		p := SpanFromContext(ctx).d.Context
		s.d.Context.TraceID, s.d.Parent, s.d.Context.Sampled = p.TraceID, p.SpanID, p.Sampled
	default:
		if rc, ok := ctx.Value(remoteKey).(SpanContext); ok && rc.Valid() {
			s.d.Context.TraceID, s.d.Parent, s.d.Context.Sampled = rc.TraceID, rc.SpanID, rc.Sampled
		} else {
			newID(s.d.Context.TraceID[:])
			s.d.Context.Sampled = true
		}
	}
	newID(s.d.Context.SpanID[:])
	return context.WithValue(ctx, spanKey, s), s
}

func SpanFromContext(ctx context.Context) *Span {
	s, _ := ctx.Value(spanKey).(*Span)
	return s
}

func (s *Span) Context() SpanContext { return s.d.Context }

func (s *Span) SetName(n string) { s.mu.Lock(); s.d.Name = n; s.mu.Unlock() }

func (s *Span) SetAttr(k string, v any) {
	s.mu.Lock()
	s.d.Attrs = append(s.d.Attrs, Attr{k, v})
	s.mu.Unlock()
}

// SetError marks the span failed. Only the message is kept; callers must not
// pass errors that embed personal data or tokens.
func (s *Span) SetError(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	s.d.ErrMsg = err.Error()
	s.mu.Unlock()
}

func (s *Span) MarkError(msg string) { s.mu.Lock(); s.d.ErrMsg = msg; s.mu.Unlock() }

func (s *Span) End() {
	s.mu.Lock()
	if s.ended {
		s.mu.Unlock()
		return
	}
	s.ended = true
	s.d.End = s.t.now()
	d := s.d
	s.mu.Unlock()
	if s.t.exp != nil && d.Context.Sampled {
		s.t.exp.Export(d)
	}
}

// ---- process-wide tracer, so service code does not need it injected ----

var (
	gmu     sync.RWMutex
	gTracer = NewTracer("app", nil)
)

func SetTracer(t *Tracer) *Tracer {
	gmu.Lock()
	defer gmu.Unlock()
	old := gTracer
	gTracer = t
	return old
}

func globalTracer() *Tracer { gmu.RLock(); defer gmu.RUnlock(); return gTracer }

// Start uses the process-wide tracer.
func Start(ctx context.Context, name string, attrs ...Attr) (context.Context, *Span) {
	return globalTracer().Start(ctx, name, KindInternal, attrs...)
}

func StartKind(ctx context.Context, name string, kind SpanKind, attrs ...Attr) (context.Context, *Span) {
	return globalTracer().Start(ctx, name, kind, attrs...)
}

// TraceIDFromContext returns the hex trace id of the current span ("" if none).
func TraceIDFromContext(ctx context.Context) string {
	if s := SpanFromContext(ctx); s != nil {
		return hex.EncodeToString(s.d.Context.TraceID[:])
	}
	return ""
}

// TraceparentFromContext is what gets copied into message payloads.
func TraceparentFromContext(ctx context.Context) string {
	if s := SpanFromContext(ctx); s != nil {
		return s.d.Context.Traceparent()
	}
	return ""
}

// MemExporter keeps spans in memory (tests).
type MemExporter struct {
	mu    sync.Mutex
	spans []SpanData
}

func (m *MemExporter) Export(d SpanData) { m.mu.Lock(); m.spans = append(m.spans, d); m.mu.Unlock() }

func (m *MemExporter) Spans() []SpanData {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]SpanData(nil), m.spans...)
}
