package httpmw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecover(t *testing.T) {
	tests := []struct {
		name string
		h    http.HandlerFunc
		want int
	}{
		{"ปกติ", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }, 204},
		{"panic -> 500", func(w http.ResponseWriter, r *http.Request) { panic("boom") }, 500},
		{"nil map panic -> 500", func(w http.ResponseWriter, r *http.Request) {
			var m map[string]int
			m["x"] = 1
		}, 500},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Recover(tc.h).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
			if rec.Code != tc.want {
				t.Fatalf("got %d want %d", rec.Code, tc.want)
			}
		})
	}
}
