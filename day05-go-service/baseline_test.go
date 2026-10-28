package day05

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/handler"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/repository"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/service"
)

// Baseline: เก็บ "พฤติกรรมของ API ตอนนี้" ไว้เป็น golden (สถานะ + body ของแต่ละ request ในสถานการณ์ตายตัว)
// เวลาใน service ถูกตรึง จึงเทียบ JSON ตรงๆ ได้ ถ้า API เปลี่ยนโดยไม่ตั้งใจ เทสต์นี้จะแดง
// เปลี่ยนโดยตั้งใจ: go test -run Baseline -update แล้ว review diff ของ testdata/baseline.golden
var update = flag.Bool("update", false, "เขียน golden ใหม่")

func snapshot() string {
	fixed := time.Date(2026, 10, 28, 9, 0, 0, 0, time.UTC)
	svc := service.New(repository.NewMemory(), func() time.Time { return fixed })
	h := handler.New(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))

	var b strings.Builder
	call := func(method, path, body string) {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		fmt.Fprintf(&b, "%s %s -> %d\n%s", method, path, rec.Code, rec.Body.String())
		if rec.Body.Len() == 0 || !strings.HasSuffix(rec.Body.String(), "\n") {
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	call("POST", "/claims", `{"patient_ref":"P-001","hospital":"รพ.A","amount_satang":125000}`)
	call("POST", "/claims", `{"patient_ref":"P-002","hospital":"รพ.B","amount_satang":8250,"status":"APPROVED"}`)
	call("POST", "/claims", `{"patient_ref":"","hospital":"รพ.A","amount_satang":1}`)
	call("GET", "/claims/1", "")
	call("GET", "/claims/99", "")
	call("GET", "/claims?status=APPROVED", "")
	call("PUT", "/claims/1", `{"patient_ref":"P-001","hospital":"รพ.A","amount_satang":130000,"status":"REJECTED"}`)
	call("POST", "/claims/approve", `{"ids":[1,2]}`)
	call("GET", "/claims/report", "")
	call("DELETE", "/claims/2", "")
	call("GET", "/claims", "")
	return b.String()
}

func TestBaseline(t *testing.T) {
	const path = "testdata/baseline.golden"
	got := snapshot()
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ไม่มี golden (รัน -update ก่อน): %v", err)
	}
	if got != string(want) {
		t.Fatalf("baseline เปลี่ยน\n--- want\n%s--- got\n%s", want, got)
	}
}
