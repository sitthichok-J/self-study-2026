package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/repository"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/service"
)

var fixed = time.Date(2026, 10, 28, 9, 0, 0, 0, time.UTC)

// safeBuf ให้ logger เขียนจากหลาย goroutine ได้โดยไม่ race
type safeBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *safeBuf) Write(p []byte) (int, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.b.Write(p) }
func (s *safeBuf) String() string              { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

func setup(t *testing.T) (http.Handler, *safeBuf) {
	t.Helper()
	logs := &safeBuf{}
	log := slog.New(slog.NewJSONHandler(logs, nil))
	svc := service.New(repository.NewMemory(), func() time.Time { return fixed })
	return New(svc, log), logs
}

func do(h http.Handler, method, path, body string, hdr ...string) *httptest.ResponseRecorder {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

const validBody = `{"patient_ref":"P-001","hospital":"รพ.ตัวอย่าง","amount_satang":125000}`

func decodeClaim(t *testing.T, rec *httptest.ResponseRecorder) model.Claim {
	t.Helper()
	var c model.Claim
	if err := json.Unmarshal(rec.Body.Bytes(), &c); err != nil {
		t.Fatalf("body=%s err=%v", rec.Body.String(), err)
	}
	return c
}

func TestHealthz(t *testing.T) {
	h, _ := setup(t)
	rec := do(h, "GET", "/healthz", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

func TestCRUDFlow(t *testing.T) {
	h, _ := setup(t)

	rec := do(h, "POST", "/claims", validBody)
	if rec.Code != 201 || rec.Header().Get("Location") != "/claims/1" {
		t.Fatalf("create: %d loc=%q body=%s", rec.Code, rec.Header().Get("Location"), rec.Body.String())
	}
	c := decodeClaim(t, rec)
	if c.ID != 1 || c.Status != model.StatusPending || c.AmountSatang != 125000 || !c.CreatedAt.Equal(fixed) {
		t.Fatalf("%+v", c)
	}

	if rec = do(h, "GET", "/claims/1", ""); rec.Code != 200 || decodeClaim(t, rec) != c {
		t.Fatalf("get: %d %s", rec.Code, rec.Body.String())
	}

	rec = do(h, "PUT", "/claims/1", `{"patient_ref":"P-009","hospital":"H2","amount_satang":5,"status":"APPROVED"}`)
	if rec.Code != 200 {
		t.Fatalf("put: %d %s", rec.Code, rec.Body.String())
	}
	if u := decodeClaim(t, rec); u.Status != model.StatusApproved || u.PatientRef != "P-009" {
		t.Fatalf("%+v", u)
	}

	rec = do(h, "GET", "/claims?status=APPROVED", "")
	var list struct {
		Items []model.Claim `json:"items"`
		Count int           `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || list.Count != 1 || len(list.Items) != 1 {
		t.Fatalf("list: %s", rec.Body.String())
	}

	if rec = do(h, "DELETE", "/claims/1", ""); rec.Code != 204 || rec.Body.Len() != 0 {
		t.Fatalf("delete: %d", rec.Code)
	}
	if rec = do(h, "GET", "/claims/1", ""); rec.Code != 404 {
		t.Fatalf("หลังลบต้อง 404 ได้ %d", rec.Code)
	}
}

func TestEmptyListIsArrayNotNull(t *testing.T) {
	h, _ := setup(t)
	if body := do(h, "GET", "/claims", "").Body.String(); !strings.Contains(body, `"items":[]`) {
		t.Fatalf("body=%s", body)
	}
}

func TestErrorResponses(t *testing.T) {
	h, _ := setup(t)
	do(h, "POST", "/claims", validBody)
	tests := []struct {
		name         string
		method, path string
		body         string
		want         int
		contains     string
	}{
		{"JSON พัง", "POST", "/claims", `{`, 400, "invalid JSON"},
		{"field แปลก", "POST", "/claims", `{"patient_ref":"P","hospital":"H","amount_satang":1,"admin":true}`, 400, "invalid JSON"},
		{"ข้อมูลต่อท้าย", "POST", "/claims", validBody + `{}`, 400, "invalid JSON"},
		{"ชนิดผิด", "POST", "/claims", `{"patient_ref":"P","hospital":"H","amount_satang":"1"}`, 400, "invalid JSON"},
		{"validation", "POST", "/claims", `{"patient_ref":"","hospital":"H","amount_satang":1}`, 400, "patient_ref"},
		{"status ผิด", "POST", "/claims", `{"patient_ref":"P","hospital":"H","amount_satang":1,"status":"X"}`, 400, "status"},
		{"id ไม่ใช่เลข", "GET", "/claims/abc", "", 400, "id"},
		{"id ติดลบ", "GET", "/claims/-1", "", 400, "id"},
		{"ไม่พบ", "GET", "/claims/999", "", 404, "not found"},
		{"PUT ไม่พบ", "PUT", "/claims/999", `{"patient_ref":"P","hospital":"H","amount_satang":1,"status":"PENDING"}`, 404, "not found"},
		{"DELETE ไม่พบ", "DELETE", "/claims/999", "", 404, "not found"},
		{"limit ไม่ใช่เลข", "GET", "/claims?limit=x", "", 400, "limit"},
		{"offset ไม่ใช่เลข", "GET", "/claims?offset=x", "", 400, "offset"},
		{"status filter ผิด", "GET", "/claims?status=X", "", 400, "status"},
		{"approve ว่าง", "POST", "/claims/approve", `{"ids":[]}`, 400, "ids"},
		{"approve ไม่พบ", "POST", "/claims/approve", `{"ids":[1,999]}`, 404, "not found"},
		{"method ผิด", "PATCH", "/claims/1", "", 405, ""},
		{"path ไม่มี", "GET", "/nope", "", 404, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(h, tc.method, tc.path, tc.body)
			if rec.Code != tc.want || !strings.Contains(rec.Body.String(), tc.contains) {
				t.Fatalf("got %d %q want %d contains %q", rec.Code, rec.Body.String(), tc.want, tc.contains)
			}
		})
	}
}

func TestBodyTooLarge(t *testing.T) {
	h, _ := setup(t)
	big := `{"patient_ref":"` + strings.Repeat("a", maxBody+10) + `"}`
	if rec := do(h, "POST", "/claims", big); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestApproveAndReport(t *testing.T) {
	h, _ := setup(t)
	do(h, "POST", "/claims", `{"patient_ref":"A","hospital":"H1","amount_satang":100}`)
	do(h, "POST", "/claims", `{"patient_ref":"B","hospital":"H1","amount_satang":50}`)
	if rec := do(h, "POST", "/claims/approve", `{"ids":[1,2]}`); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"approved":2`) {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if c := decodeClaim(t, do(h, "GET", "/claims/2", "")); c.Status != model.StatusApproved {
		t.Fatalf("%+v", c)
	}
	rec := do(h, "GET", "/claims/report", "") // literal path ต้องชนะ {id}
	if rec.Code != 200 || rec.Body.String() != "H1\t2\t150\n" || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("%d %q", rec.Code, rec.Body.String())
	}
}

func TestRequestID(t *testing.T) {
	h, _ := setup(t)
	if got := do(h, "GET", "/healthz", "", HeaderRequestID, "abc-123").Header().Get(HeaderRequestID); got != "abc-123" {
		t.Fatalf("ต้องใช้ id จาก client: %q", got)
	}
	gen1 := do(h, "GET", "/healthz", "").Header().Get(HeaderRequestID)
	gen2 := do(h, "GET", "/healthz", "").Header().Get(HeaderRequestID)
	if len(gen1) != 16 || gen1 == gen2 {
		t.Fatalf("สร้างใหม่ต้องยาว 16 และไม่ซ้ำ: %q %q", gen1, gen2)
	}
	long := strings.Repeat("x", 65)
	if got := do(h, "GET", "/healthz", "", HeaderRequestID, long).Header().Get(HeaderRequestID); got == long {
		t.Fatal("id ยาวเกินต้องไม่ถูกใช้")
	}
	if RequestIDFrom(context.Background()) != "" {
		t.Fatal("ไม่มีใน context ต้องว่าง")
	}
}

func lastLog(t *testing.T, logs *safeBuf) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
	var m map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &m); err != nil {
		t.Fatalf("log=%q err=%v", logs.String(), err)
	}
	return m
}

func TestAccessLogHasFieldsAndNoSensitiveData(t *testing.T) {
	h, logs := setup(t)
	do(h, "POST", "/claims", validBody, HeaderRequestID, "rid-1")
	do(h, "GET", "/claims?status=PENDING", "", HeaderRequestID, "rid-2")

	m := lastLog(t, logs)
	if m["msg"] != "request" || m["request_id"] != "rid-2" || m["method"] != "GET" || m["path"] != "/claims" || m["status"] != float64(200) {
		t.Fatalf("%v", m)
	}
	if _, ok := m["duration_ms"]; !ok {
		t.Fatal("ต้องมี duration_ms")
	}
	all := logs.String()
	for _, secret := range []string{"P-001", "รพ.ตัวอย่าง", "125000", "status=PENDING"} {
		if strings.Contains(all, secret) {
			t.Fatalf("log ต้องไม่มี %q:\n%s", secret, all)
		}
	}
}

// brokenClaims: service ที่พังด้วย error ภายใน
type brokenClaims struct{ Claims }

func (brokenClaims) Get(context.Context, int64) (model.Claim, error) {
	return model.Claim{}, errors.New("pq: password authentication failed for user admin")
}
func (brokenClaims) Report(context.Context) (string, error) { return "", errors.New("secret detail") }
func (brokenClaims) All(context.Context) ([]model.Claim, error) {
	return nil, errors.New("secret detail")
}

type panicClaims struct{ Claims }

func (panicClaims) Get(context.Context, int64) (model.Claim, error) { panic("boom") }

func TestInternalErrorDoesNotLeakDetails(t *testing.T) {
	logs := &safeBuf{}
	h := New(brokenClaims{}, slog.New(slog.NewJSONHandler(logs, nil)))
	rec := do(h, "GET", "/claims/1", "", HeaderRequestID, "rid-9")
	if rec.Code != 500 || strings.TrimSpace(rec.Body.String()) != `{"error":"internal error"}` {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Fatal("ห้ามส่งรายละเอียดภายในให้ client")
	}
	if !strings.Contains(logs.String(), "password authentication failed") || !strings.Contains(logs.String(), "rid-9") {
		t.Fatalf("ต้อง log สาเหตุจริงพร้อม request_id:\n%s", logs.String())
	}
	if rec := do(h, "GET", "/claims/report", ""); rec.Code != 500 {
		t.Fatalf("report error: %d", rec.Code)
	}
}

func TestPanicBecomes500AndServerSurvives(t *testing.T) {
	logs := &safeBuf{}
	h := New(panicClaims{}, slog.New(slog.NewJSONHandler(logs, nil)))
	rec := do(h, "GET", "/claims/1", "")
	if rec.Code != 500 || !strings.Contains(rec.Body.String(), "internal error") || strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(logs.String(), `"msg":"panic"`) {
		t.Fatalf("ต้อง log panic:\n%s", logs.String())
	}
	if m := lastLog(t, logs); m["status"] != float64(500) {
		t.Fatalf("access log ต้องบันทึก 500: %v", m)
	}
	if rec := do(h, "GET", "/healthz", ""); rec.Code != 200 {
		t.Fatal("request ถัดไปต้องยังทำงาน")
	}
}

func TestDebugMux(t *testing.T) {
	svc := service.New(repository.NewMemory(), func() time.Time { return fixed })
	svc.Create(context.Background(), service.Input{PatientRef: "P", Hospital: "H", AmountSatang: 5})
	d := Debug(svc)
	if rec := do(d, "GET", "/debug/report-slow", ""); rec.Code != 200 || rec.Body.String() != "H\t1\t5\n" {
		t.Fatalf("%d %q", rec.Code, rec.Body.String())
	}
	if rec := do(d, "GET", "/debug/pprof/", ""); rec.Code != 200 || !strings.Contains(rec.Body.String(), "goroutine") {
		t.Fatalf("pprof index: %d", rec.Code)
	}
	if rec := do(Debug(brokenClaims{}), "GET", "/debug/report-slow", ""); rec.Code != 500 || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("service error ต้องเป็น 500 โดยไม่เผยรายละเอียด: %d %s", rec.Code, rec.Body.String())
	}
}
