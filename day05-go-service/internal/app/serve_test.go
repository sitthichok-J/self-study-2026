package app

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func listen(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return ln
}

// request ที่กำลังทำอยู่ตอนสั่งปิด ต้องทำให้เสร็จ (graceful) แล้ว Serve คืน nil
func TestInFlightRequestFinishesDuringShutdown(t *testing.T) {
	started := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte("done"))
	})}
	ln := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	serveErr := make(chan error, 1)
	go func() { serveErr <- Serve(ctx, srv, ln, 2*time.Second, quiet()) }()

	type result struct {
		body string
		err  error
	}
	res := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			res <- result{err: err}
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		res <- result{body: string(b)}
	}()

	<-started // request เข้ามาและกำลังทำงานอยู่
	cancel()  // สั่งปิด

	if r := <-res; r.err != nil || r.body != "done" {
		t.Fatalf("request ที่ค้างอยู่ต้องจบปกติ: %+v", r)
	}
	if err := <-serveErr; err != nil {
		t.Fatalf("Serve ต้องคืน nil: %v", err)
	}
	if _, err := http.Get("http://" + ln.Addr().String()); err == nil {
		t.Fatal("หลังปิดต้องไม่รับ request ใหม่")
	}
}

// ถ้า request ค้างนานเกิน timeout ต้องคืน error (ไม่รอตลอดกาล)
func TestShutdownTimeoutReturnsError(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
	})}
	ln := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	serveErr := make(chan error, 1)
	go func() { serveErr <- Serve(ctx, srv, ln, 50*time.Millisecond, quiet()) }()

	go http.Get("http://" + ln.Addr().String())
	<-started
	cancel()
	err := <-serveErr
	close(release)
	if err == nil || !strings.Contains(err.Error(), "shutdown") {
		t.Fatalf("err=%v", err)
	}
}

// server ล้มเองก่อนถูกสั่งปิด (เช่น listener ถูกปิด) ต้องคืน error ไม่ใช่ค้าง
func TestServeFailureIsReported(t *testing.T) {
	srv := &http.Server{Handler: http.NotFoundHandler()}
	ln := listen(t)
	ln.Close() // listener พังตั้งแต่ต้น -> srv.Serve คืน error ที่ไม่ใช่ ErrServerClosed
	err := Serve(context.Background(), srv, ln, time.Second, quiet())
	if err == nil || !strings.Contains(err.Error(), "serve") {
		t.Fatalf("err=%v", err)
	}
}

func TestShutdownWithNoTraffic(t *testing.T) {
	srv := &http.Server{Handler: http.NotFoundHandler()}
	ln := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Serve(ctx, srv, ln, time.Second, quiet()); err != nil {
		t.Fatalf("err=%v", err)
	}
}
