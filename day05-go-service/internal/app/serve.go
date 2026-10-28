// Package app: เริ่ม/หยุด HTTP server แบบ graceful (แยกจาก main เพื่อให้เทสต์ได้)
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Serve เปิด srv บน ln จนกว่า ctx จะถูก cancel แล้วเรียก Shutdown (รอ request ที่ค้างอยู่ไม่เกิน timeout)
// คืน nil เมื่อปิดเรียบร้อย; คืน error ถ้า server ล้มเองหรือปิดไม่ทันเวลา
func Serve(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration, log *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		err := srv.Serve(ln)
		if !errors.Is(err, http.ErrServerClosed) { // ErrServerClosed คือการปิดตามปกติ ไม่ใช่ error
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh: // server ล้มเองก่อนถูกสั่งปิด
		if err != nil {
			return fmt.Errorf("serve: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	log.Info("shutting down", "timeout", timeout.String())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout) // ctx เดิมถูก cancel แล้ว ต้องสร้างใหม่
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close() // หมดเวลา: ตัด connection ที่เหลือ
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errCh; err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	log.Info("shutdown complete")
	return nil
}
