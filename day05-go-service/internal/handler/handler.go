// Package handler: ชั้น HTTP แปลง request/response และ error -> status code เท่านั้น ไม่มี business logic
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/report"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/service"
)

// Claims คือสิ่งที่ handler ต้องการจาก service (interface เล็ก ฝั่งผู้ใช้ ทำให้เทสต์ด้วย fake ได้)
type Claims interface {
	Create(ctx context.Context, in service.Input) (model.Claim, error)
	Get(ctx context.Context, id int64) (model.Claim, error)
	List(ctx context.Context, f model.ListFilter) ([]model.Claim, error)
	Update(ctx context.Context, id int64, in service.Input) (model.Claim, error)
	Delete(ctx context.Context, id int64) error
	ApproveMany(ctx context.Context, ids []int64) error
	Report(ctx context.Context) (string, error)
	All(ctx context.Context) ([]model.Claim, error)
}

const maxBody = 1 << 20 // 1 MiB

type errorBody struct {
	Error string `json:"error"`
}

type claimBody struct {
	PatientRef   string       `json:"patient_ref"`
	Hospital     string       `json:"hospital"`
	AmountSatang int64        `json:"amount_satang"`
	Status       model.Status `json:"status"`
}

func (b claimBody) input() service.Input {
	return service.Input{PatientRef: b.PatientRef, Hospital: b.Hospital, AmountSatang: b.AmountSatang, Status: b.Status}
}

type api struct {
	svc Claims
	log *slog.Logger
}

// New สร้าง http.Handler ของ API ทั้งหมด พร้อม middleware: requestID -> accessLog -> recover
func New(svc Claims, log *slog.Logger) http.Handler {
	a := &api{svc: svc, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /claims", a.create)
	mux.HandleFunc("GET /claims", a.list)
	mux.HandleFunc("POST /claims/approve", a.approve) // pattern ที่เจาะจงกว่า (literal) ชนะ {id} เสมอ ไม่ขึ้นกับลำดับที่ลงทะเบียน
	mux.HandleFunc("GET /claims/report", a.report)
	mux.HandleFunc("GET /claims/{id}", a.get)
	mux.HandleFunc("PUT /claims/{id}", a.update)
	mux.HandleFunc("DELETE /claims/{id}", a.delete)
	return requestID(accessLog(log, recoverPanic(log, mux)))
}

// Debug คือ mux สำหรับ dev เท่านั้น (pprof + endpoint ช้าที่ตั้งใจ) ต้อง bind ที่ localhost และไม่เปิดใน production
func Debug(svc Claims) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("GET /debug/report-slow", func(w http.ResponseWriter, r *http.Request) {
		all, err := svc.All(r.Context())
		if err != nil {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, report.BuildSlow(all))
	})
	return mux
}

func (a *api) create(w http.ResponseWriter, r *http.Request) {
	var b claimBody
	if !a.decode(w, r, &b) {
		return
	}
	c, err := a.svc.Create(r.Context(), b.input())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/claims/"+strconv.FormatInt(c.ID, 10))
	writeJSON(w, http.StatusCreated, c)
}

func (a *api) get(w http.ResponseWriter, r *http.Request) {
	id, ok := a.id(w, r)
	if !ok {
		return
	}
	c, err := a.svc.Get(r.Context(), id)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (a *api) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := model.ListFilter{Status: model.Status(q.Get("status"))}
	var err error
	if f.Limit, err = intParam(q.Get("limit")); err != nil {
		a.fail(w, r, &service.ValidationError{Field: "limit", Reason: "must be an integer"})
		return
	}
	if f.Offset, err = intParam(q.Get("offset")); err != nil {
		a.fail(w, r, &service.ValidationError{Field: "offset", Reason: "must be an integer"})
		return
	}
	list, err := a.svc.List(r.Context(), f)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list, "count": len(list)})
}

func (a *api) update(w http.ResponseWriter, r *http.Request) {
	id, ok := a.id(w, r)
	if !ok {
		return
	}
	var b claimBody
	if !a.decode(w, r, &b) {
		return
	}
	c, err := a.svc.Update(r.Context(), id, b.input())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (a *api) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := a.id(w, r)
	if !ok {
		return
	}
	if err := a.svc.Delete(r.Context(), id); err != nil {
		a.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) approve(w http.ResponseWriter, r *http.Request) {
	var b struct {
		IDs []int64 `json:"ids"`
	}
	if !a.decode(w, r, &b) {
		return
	}
	if err := a.svc.ApproveMany(r.Context(), b.IDs); err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"approved": len(b.IDs)})
}

func (a *api) report(w http.ResponseWriter, r *http.Request) {
	text, err := a.svc.Report(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, text)
}

func intParam(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func (a *api) id(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		a.fail(w, r, &service.ValidationError{Field: "id", Reason: "must be a positive integer"})
		return 0, false
	}
	return id, true
}

// decode อ่าน JSON แบบเข้มงวด: จำกัดขนาด, ห้าม field แปลก, ห้ามมีข้อมูลต่อท้าย
func (a *api) decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeJSON(w, http.StatusRequestEntityTooLarge, errorBody{Error: "request body too large"})
			return false
		}
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return false
	}
	if dec.More() {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return false
	}
	return true
}

// fail แปลง error เป็น status code ข้อความที่ส่งให้ client มีเฉพาะที่ปลอดภัย
// error ภายใน (DB ฯลฯ) ถูก log พร้อม request_id แล้วตอบแค่ "internal error"
func (a *api) fail(w http.ResponseWriter, r *http.Request, err error) {
	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeJSON(w, http.StatusBadRequest, errorBody{Error: ve.Error()})
	case errors.Is(err, model.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorBody{Error: "claim not found"})
	default:
		a.log.ErrorContext(r.Context(), "internal error", "err", err, "request_id", RequestIDFrom(r.Context()))
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: "internal error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
