// Package http is the REST adapter (stdlib net/http, Go 1.22+ mux patterns).
package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	nethttp "net/http"
	"strconv"

	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/service"
)

type Handler struct {
	svc   *service.Claims
	ready func(context.Context) error
	log   *slog.Logger
}

// New returns the router. ready may be nil (always ready).
func New(svc *service.Claims, ready func(context.Context) error, log *slog.Logger) nethttp.Handler {
	if log == nil {
		log = slog.Default()
	}
	h := &Handler{svc: svc, ready: ready, log: log}
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /claims", h.submit)
	mux.HandleFunc("GET /claims/{id}", h.get)
	mux.HandleFunc("GET /claims", h.list)
	mux.HandleFunc("GET /healthz", func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", h.readyz)
	return mux
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
		writeErr(w, h.log, domain.ErrInvalid, "request body must be valid JSON with patient_ref and amount")
		return
	}
	c, created, err := h.svc.Submit(r.Context(), service.SubmitInput{
		IdempotencyKey: r.Header.Get("Idempotency-Key"), PatientRef: b.PatientRef, Amount: b.Amount,
	})
	if err != nil {
		writeErr(w, h.log, err, "")
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
	c, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, h.log, err, "")
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
			writeErr(w, h.log, err, "")
			return
		}
		st = &v
	}
	limit, err1 := atoiDefault(q.Get("limit"), 0)
	offset, err2 := atoiDefault(q.Get("offset"), 0)
	if err1 != nil || err2 != nil {
		writeErr(w, h.log, domain.ErrInvalid, "limit and offset must be integers")
		return
	}
	items, err := h.svc.List(r.Context(), st, limit, offset)
	if err != nil {
		writeErr(w, h.log, err, "")
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

func writeErr(w nethttp.ResponseWriter, log *slog.Logger, err error, override string) {
	code, msg := 500, "internal error"
	switch {
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
		log.Error("unhandled error", "err", err) // details stay in the log, not the response
	}
	writeJSON(w, code, errBody{Error: msg})
}
