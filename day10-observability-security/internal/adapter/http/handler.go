// Package http is the REST adapter (stdlib net/http, Go 1.22+ mux patterns).
package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	nethttp "net/http"
	"strconv"

	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/auth"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/service"
)

type Handler struct {
	svc   *service.Claims
	ready func(context.Context) error
	log   *slog.Logger
	ver   *auth.Verifier
}

// New returns the router. ready may be nil (always ready). Every /claims
// route needs a valid Bearer token; /healthz and /readyz stay open for probes.
// Metrics are NOT served here: they live on a separate listener (see cmd).
func New(svc *service.Claims, ready func(context.Context) error, log *slog.Logger, ver *auth.Verifier) nethttp.Handler {
	if log == nil {
		log = slog.Default()
	}
	if ver == nil {
		panic("http.New: a JWT verifier is required (fail closed)")
	}
	for _, r := range auth.Reasons {
		obs.M().AuthFailures.Add(0, r) // see obs.initSeries: first increment must be visible to increase()
	}
	h := &Handler{svc: svc, ready: ready, log: log, ver: ver}
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /claims", h.authed(h.submit))
	mux.HandleFunc("GET /claims/{id}", h.authed(h.get))
	mux.HandleFunc("GET /claims", h.authed(h.list))
	mux.HandleFunc("GET /healthz", func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", h.readyz)
	// obs.HTTP must see the mux directly (route label comes from r.Pattern).
	return obs.HTTP(securityHeaders(mux), obs.M(), log)
}

// securityHeaders: this API only returns JSON, so lock the browser down hard.
func securityHeaders(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		hd := w.Header()
		hd.Set("X-Content-Type-Options", "nosniff")
		hd.Set("Cache-Control", "no-store")
		hd.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		hd.Set("X-Frame-Options", "DENY")
		hd.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// authed verifies the Bearer token and puts the Principal in the context.
func (h *Handler) authed(next nethttp.HandlerFunc) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		p, err := h.ver.Verify(auth.BearerToken(r.Header.Get("Authorization")))
		if err != nil {
			reason := auth.Reason(err)
			obs.M().AuthFailures.Inc(reason)
			h.log.WarnContext(r.Context(), "authentication failed", "security_event", "auth_failed", "reason", reason)
			w.Header().Set("WWW-Authenticate", `Bearer realm="claims"`)
			writeJSON(w, nethttp.StatusUnauthorized, errBody{Error: "unauthorized"})
			return
		}
		next(w, r.WithContext(auth.WithPrincipal(r.Context(), p)))
	}
}

func viewer(r *nethttp.Request) domain.Viewer {
	p, _ := auth.FromContext(r.Context()) // authed() guarantees it; zero value = no access
	return domain.Viewer{ID: p.Subject, Staff: p.Staff}
}

type submitBody struct {
	PatientRef string `json:"patient_ref"`
	Amount     int64  `json:"amount"`
}

func (h *Handler) submit(w nethttp.ResponseWriter, r *nethttp.Request) {
	r.Body = nethttp.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var b submitBody
	if err := dec.Decode(&b); err != nil {
		writeErr(w, r, h.log, domain.ErrInvalid, "request body must be valid JSON with patient_ref and amount")
		return
	}
	c, created, err := h.svc.Submit(r.Context(), service.SubmitInput{
		Owner: viewer(r).ID, IdempotencyKey: r.Header.Get("Idempotency-Key"), PatientRef: b.PatientRef, Amount: b.Amount,
	})
	if err != nil {
		writeErr(w, r, h.log, err, "")
		return
	}
	w.Header().Set("Location", "/claims/"+c.ID)
	if !created {
		w.Header().Set("Idempotent-Replayed", "true")
		writeJSON(w, nethttp.StatusOK, c)
		return
	}
	writeJSON(w, nethttp.StatusCreated, c)
}

func (h *Handler) get(w nethttp.ResponseWriter, r *nethttp.Request) {
	c, err := h.svc.Get(r.Context(), viewer(r), r.PathValue("id"))
	if errors.Is(err, domain.ErrForbidden) {
		// Someone else's claim looks exactly like a missing one.
		obs.M().AuthzDenied.Inc("claim")
		h.log.WarnContext(r.Context(), "access denied to claim", "security_event", "access_denied",
			"sub", viewer(r).ID, "claim_id", r.PathValue("id"))
		writeErr(w, r, h.log, domain.ErrNotFound, "")
		return
	}
	if err != nil {
		writeErr(w, r, h.log, err, "")
		return
	}
	writeJSON(w, 200, c)
}

func (h *Handler) list(w nethttp.ResponseWriter, r *nethttp.Request) {
	q := r.URL.Query()
	var st *domain.Status
	if s := q.Get("status"); s != "" {
		v, err := domain.ParseStatus(s)
		if err != nil {
			writeErr(w, r, h.log, err, "")
			return
		}
		st = &v
	}
	limit, err1 := atoiDefault(q.Get("limit"), 0)
	offset, err2 := atoiDefault(q.Get("offset"), 0)
	if err1 != nil || err2 != nil {
		writeErr(w, r, h.log, domain.ErrInvalid, "limit and offset must be integers")
		return
	}
	items, err := h.svc.List(r.Context(), viewer(r), st, limit, offset)
	if err != nil {
		writeErr(w, r, h.log, err, "")
		return
	}
	writeJSON(w, 200, map[string]any{"items": items, "limit": effLimit(limit), "offset": offset})
}

func (h *Handler) readyz(w nethttp.ResponseWriter, r *nethttp.Request) {
	if h.ready != nil {
		if err := h.ready(r.Context()); err != nil {
			writeJSON(w, 503, map[string]string{"status": "not ready"})
			return
		}
	}
	writeJSON(w, 200, map[string]string{"status": "ready"})
}

func effLimit(l int) int {
	if l == 0 {
		return service.DefaultLimit
	}
	return l
}

func atoiDefault(s string, def int) (int, error) {
	if s == "" {
		return def, nil
	}
	return strconv.Atoi(s)
}

func writeJSON(w nethttp.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type errBody struct {
	Error string `json:"error"`
}

func writeErr(w nethttp.ResponseWriter, r *nethttp.Request, log *slog.Logger, err error, override string) {
	code, msg := 500, "internal error"
	switch {
	case errors.Is(err, domain.ErrForbidden):
		code, msg = 403, "forbidden"
	case errors.Is(err, domain.ErrNotFound):
		code, msg = 404, "claim not found"
	case errors.Is(err, domain.ErrIdempotencyRequired):
		code, msg = 400, "Idempotency-Key header is required"
	case errors.Is(err, domain.ErrIdempotencyConflict):
		code, msg = 409, "Idempotency-Key was already used with a different request"
	case errors.Is(err, domain.ErrInvalid):
		code, msg = 400, err.Error()
		if override != "" {
			msg = override
		}
	default:
		log.ErrorContext(r.Context(), "unhandled error", "err", err) // details stay in the log, not the response
	}
	writeJSON(w, code, errBody{Error: msg})
}
