package obs

import (
	"context"
	"encoding/hex"
	"io"
	"log/slog"
	"strings"
)

// WithRequestID / RequestID carry the correlation id through ctx.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, reqIDKey, id)
}

func RequestID(ctx context.Context) string {
	s, _ := ctx.Value(reqIDKey).(string)
	return s
}

var sensitiveExact = map[string]bool{
	"authorization": true, "cookie": true, "set-cookie": true, "jwt": true,
	"patient_ref": true, "patient": true, "api_key": true, "apikey": true,
}

var sensitiveSub = []string{"password", "passwd", "secret", "token"}

func isSensitive(key string) bool {
	k := strings.ToLower(key)
	if sensitiveExact[k] {
		return true
	}
	for _, s := range sensitiveSub {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}

func redact(_ []string, a slog.Attr) slog.Attr {
	if isSensitive(a.Key) {
		return slog.String(a.Key, "[redacted]")
	}
	return a
}

type ctxHandler struct{ slog.Handler }

func (h ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if s := SpanFromContext(ctx); s != nil {
		r.AddAttrs(
			slog.String("trace_id", hex.EncodeToString(s.d.Context.TraceID[:])),
			slog.String("span_id", hex.EncodeToString(s.d.Context.SpanID[:])),
		)
	}
	if id := RequestID(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h ctxHandler) WithAttrs(a []slog.Attr) slog.Handler { return ctxHandler{h.Handler.WithAttrs(a)} }
func (h ctxHandler) WithGroup(n string) slog.Handler      { return ctxHandler{h.Handler.WithGroup(n)} }

// NewLogger returns a JSON logger that adds trace_id/span_id/request_id from
// the context passed to the *Context methods and redacts sensitive keys.
func NewLogger(w io.Writer, level slog.Leveler) *slog.Logger {
	return slog.New(ctxHandler{slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level, ReplaceAttr: redact})})
}
