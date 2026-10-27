package fetch

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newServer(t *testing.T, cur, peak *atomic.Int64) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := cur.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		defer cur.Add(-1)
		time.Sleep(20 * time.Millisecond)
		if r.URL.Path == "/boom" {
			http.Error(w, "boom", 500)
			return
		}
		fmt.Fprint(w, strings.Repeat("x", len(r.URL.Path)))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestAllRespectsLimitAndKeepsOrder(t *testing.T) {
	var cur, peak atomic.Int64
	srv := newServer(t, &cur, &peak)
	urls := make([]string, 20)
	for i := range urls {
		urls[i] = fmt.Sprintf("%s/%d", srv.URL, i)
	}
	got, err := All(context.Background(), srv.Client(), urls, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 20 {
		t.Fatalf("got %d", len(got))
	}
	for i, r := range got {
		if r.URL != urls[i] || r.Status != 200 || r.Bytes != len(fmt.Sprintf("/%d", i)) {
			t.Fatalf("result %d = %+v", i, r)
		}
	}
	if p := peak.Load(); p > 5 || p < 2 {
		t.Fatalf("peak=%d ต้องอยู่ระหว่าง 2..5", p)
	}
}

func TestAllReturnsErrorOnBadURL(t *testing.T) {
	var cur, peak atomic.Int64
	srv := newServer(t, &cur, &peak)
	_, err := All(context.Background(), srv.Client(), []string{srv.URL + "/ok", "http://127.0.0.1:1/unreachable"}, 2)
	if err == nil || !strings.Contains(err.Error(), "fetch") {
		t.Fatalf("err=%v", err)
	}
}

func TestAllCancelledContext(t *testing.T) {
	var cur, peak atomic.Int64
	srv := newServer(t, &cur, &peak)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := All(ctx, srv.Client(), []string{srv.URL + "/a", srv.URL + "/b"}, 2)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestAllEmpty(t *testing.T) {
	got, err := All(context.Background(), http.DefaultClient, nil, 3)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v err %v", got, err)
	}
}
