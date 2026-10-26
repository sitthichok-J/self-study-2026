// Package httpmw: middleware เล็กๆ
package httpmw

import (
	"log"
	"net/http"
)

// Recover จับ panic ใน handler แล้วตอบ 500 (ไม่ทำให้ทั้ง server ล้ม)
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
