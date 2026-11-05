package obs

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var reqIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type statusWriter struct {
	http.ResponseWriter
	code  int
	bytes int
}

func (w *statusWriter) WriteHeader(c int) {
	if w.code == 0 {
		w.code = c
	}
	w.ResponseWriter.WriteHeader(c)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.code == 0 {
		w.code = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// routeLabel turns "GET /claims/{id}" into "/claims/{id}". Unmatched requests
// share one label so scanners cannot create unbounded series.
func routeLabel(pattern string) string {
	if pattern == "" {
		return "unmatched"
	}
	if i := strings.IndexByte(pattern, ' '); i >= 0 {
		return pattern[i+1:]
	}
	return pattern
}

// HTTP wraps the router (it must wrap the mux directly: the mux fills in
// r.Pattern on the request it receives, which is how the route label is known).
// It sets a request id, continues or starts a trace, records RED metrics,
// writes one access-log line and turns panics into a 500.
func HTTP(next http.Handler, m *Metrics, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := r.Header.Get("X-Request-Id")
		if !reqIDRe.MatchString(id) {
			id = NewRequestID()
		}
		w.Header().Set("X-Request-Id", id)

		ctx := WithRequestID(r.Context(), id)
		if sc, ok := ParseTraceparent(r.Header.Get("traceparent")); ok {
			ctx = ContextWithRemote(ctx, sc)
		}
		ctx, span := StartKind(ctx, r.Method, KindServer, Attr{"http.request.method", r.Method})
		r = r.WithContext(ctx)

		sw := &statusWriter{ResponseWriter: w}
		m.InFlight.Add(1)
		defer func() {
			m.InFlight.Add(-1)
			if rec := recover(); rec != nil {
				m.Panics.Inc()
				log.ErrorContext(ctx, "handler panic", "panic", "recovered")
				if sw.code == 0 {
					sw.Header().Set("Content-Type", "application/json")
					sw.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(sw).Encode(map[string]string{"error": "internal error"})
				}
				span.MarkError("panic")
			}
			code := sw.code
			if code == 0 {
				code = http.StatusOK
			}
			route := routeLabel(r.Pattern)
			dur := time.Since(start)
			m.HTTPRequests.Inc(r.Method, route, strconv.Itoa(code))
			m.HTTPDuration.Observe(dur.Seconds(), r.Method, route)
			span.SetName(r.Method + " " + route)
			span.SetAttr("http.route", route)
			span.SetAttr("http.response.status_code", code)
			if code >= 500 {
				span.MarkError("http " + strconv.Itoa(code))
			}
			span.End()
			log.LogAttrs(ctx, levelFor(code), "request",
				slog.String("method", r.Method), slog.String("route", route),
				slog.Int("status", code), slog.Float64("duration_ms", float64(dur.Microseconds())/1000),
				slog.Int("bytes", sw.bytes))
		}()
		next.ServeHTTP(sw, r)
	})
}

func levelFor(code int) slog.Level {
	if code >= 500 {
		return slog.LevelError
	}
	return slog.LevelInfo
}

// NewRequestID: 16 random bytes as hex, reusing the trace id generator.
func NewRequestID() string {
	var b [16]byte
	newID(b[:])
	const hexd = "0123456789abcdef"
	out := make([]byte, 32)
	for i, x := range b {
		out[i*2], out[i*2+1] = hexd[x>>4], hexd[x&15]
	}
	return string(out)
}
