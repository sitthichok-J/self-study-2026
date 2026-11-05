// Package http is the REST adapter (stdlib net/http, Go 1.22+ mux patterns).
package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	nethttp "net/http"
	"strconv"
	"time"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/ratelimit"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/service"
)

type Handler struct {
	svc    *service.Claims
	ready  func(context.Context) error
	log    *slog.Logger
	ver    *auth.Verifier
	sess   *auth.Service // nil = no /auth routes (Day 10 behaviour)
	cookie CookieConfig
	opt    Options
}

// Options are the Day 13 guards. Zero value = off (Day 12 behaviour).
type Options struct {
	Revoker    auth.Revoker      // nil = no blocklist check
	Limiter    ratelimit.Limiter // nil = no per-user limit
	UserLimit  int               // requests per UserWindow per signed-in user
	UserWindow time.Duration
}

// New returns the router. ready may be nil (always ready). Every /claims
// route needs a valid Bearer token; /healthz and /readyz stay open for probes.
// Metrics are NOT served here: they live on a separate listener (see cmd).
func New(svc *service.Claims, ready func(context.Context) error, log *slog.Logger, ver *auth.Verifier) nethttp.Handler {
	return NewWithSessions(svc, ready, log, ver, nil, CookieConfig{})
}

// NewWithSessions is New plus the /auth/login, /auth/refresh and /auth/logout
// routes. sess may be nil (no /auth routes). The /auth routes are the only ones
// that do not need a Bearer token: they are how you get one.
func NewWithSessions(svc *service.Claims, ready func(context.Context) error, log *slog.Logger, ver *auth.Verifier, sess *auth.Service, cookie CookieConfig) nethttp.Handler {
	return NewFull(svc, ready, log, ver, sess, cookie, Options{})
}

// NewFull is NewWithSessions plus the blocklist and per-user rate limit.
func NewFull(svc *service.Claims, ready func(context.Context) error, log *slog.Logger, ver *auth.Verifier, sess *auth.Service, cookie CookieConfig, opt Options) nethttp.Handler {
	if log == nil {
		log = slog.Default()
	}
	if ver == nil {
		panic("http.New: a JWT verifier is required (fail closed)")
	}
	for _, r := range auth.Reasons {
		obs.M().AuthFailures.Add(0, r) // see obs.initSeries: first increment must be visible to increase()
	}
	h := &Handler{svc: svc, ready: ready, log: log, ver: ver, sess: sess, cookie: cookie.withDefaults(), opt: opt}
	mux := nethttp.NewServeMux()
	if sess != nil {
		mux.HandleFunc("POST /auth/login", h.login)
		mux.HandleFunc("POST /auth/refresh", h.refresh)
		mux.HandleFunc("POST /auth/logout", h.logout)
	}
	mux.HandleFunc("POST /claims", h.authed(h.submit))
	mux.HandleFunc("GET /claims/{id}", h.authed(h.get))
	mux.HandleFunc("GET /claims", h.authed(h.list))
	mux.HandleFunc("POST /claims/{id}/approve", h.authed(h.decide(domain.StatusApproved, "manual_approved")))
	mux.HandleFunc("POST /claims/{id}/reject", h.authed(h.decide(domain.StatusRejected, "manual_rejected")))
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

// authed verifies the Bearer token, then the blocklist, then the per-user limit, and puts the Principal in the context.
func (h *Handler) authed(next nethttp.HandlerFunc) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		p, err := h.ver.Verify(auth.BearerToken(r.Header.Get("Authorization")))
		if err == nil && h.opt.Revoker != nil {
			revoked, cerr := h.opt.Revoker.Check(r.Context(), p)
			switch {
			case cerr != nil:
				// Cannot tell whether the token was revoked: fail closed (503, not 401, so the client does not log the user out).
				h.log.ErrorContext(r.Context(), "revocation check failed", "security_event", "revocation_check_failed", "err", cerr)
				w.Header().Set("Retry-After", "5")
				writeJSON(w, nethttp.StatusServiceUnavailable, errBody{Error: "temporarily unavailable"})
				return
			case revoked:
				err = auth.ErrRevoked
			}
		}
		if err != nil {
			reason := auth.Reason(err)
			obs.M().AuthFailures.Inc(reason)
			h.log.WarnContext(r.Context(), "authentication failed", "security_event", "auth_failed", "reason", reason)
			w.Header().Set("WWW-Authenticate", `Bearer realm="claims"`)
			writeJSON(w, nethttp.StatusUnauthorized, errBody{Error: "unauthorized"})
			return
		}
		if !h.allowUser(w, r, p) {
			return
		}
		next(w, r.WithContext(auth.WithPrincipal(r.Context(), p)))
	}
}

// allowUser applies the per-user limit. If the counter store is down it lets the request through
// (availability over a limit; Kong's per-IP limit still applies) and logs it.
func (h *Handler) allowUser(w nethttp.ResponseWriter, r *nethttp.Request, p auth.Principal) bool {
	if h.opt.Limiter == nil || h.opt.UserLimit <= 0 {
		return true
	}
	res, err := h.opt.Limiter.Allow(r.Context(), "user:"+p.Subject, h.opt.UserLimit, h.opt.UserWindow)
	if err != nil {
		h.log.WarnContext(r.Context(), "user rate limit unavailable, allowing", "security_event", "rate_limit_unavailable", "err", err)
		return true
	}
	if res.Allowed {
		w.Header().Set("X-RateLimit-Remaining-User", strconv.Itoa(res.Remaining))
		return true
	}
	secs := int(math.Ceil(res.RetryAfter.Seconds()))
	if secs < 1 {
		secs = 1
	}
	h.log.WarnContext(r.Context(), "user rate limit exceeded", "security_event", "user_rate_limited")
	w.Header().Set("Retry-After", strconv.Itoa(secs))
	writeJSON(w, nethttp.StatusTooManyRequests, errBody{Error: "too many requests"})
	return false
}

// decide is the staff-only manual review: approve or reject a claim that is still PENDING.
func (h *Handler) decide(to domain.Status, reason string) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		c, err := h.svc.Decide(r.Context(), viewer(r), r.PathValue("id"), to, reason)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidTransition) {
				writeJSON(w, nethttp.StatusConflict, errBody{Error: "claim is no longer pending"})
				return
			}
			writeErr(w, r, h.log, err, "")
			return
		}
		writeJSON(w, nethttp.StatusOK, c)
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
