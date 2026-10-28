// API ของ claims: ADDR=:8080 DATABASE_URL=... go run -tags pgx ./cmd/api
// ไม่ตั้ง DATABASE_URL = ใช้ in-memory (ข้อมูลหายเมื่อปิด)
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/app"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/config"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/handler"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/repository"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err) // log/แสดง error ที่ชั้นบนสุดครั้งเดียว
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, closeRepo, err := openRepo(ctx, cfg, log)
	if err != nil {
		return err
	}
	defer closeRepo()
	svc := service.New(repo, nil)

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.Addr, err)
	}
	srv := &http.Server{
		Handler:           handler.New(svc, log),
		ReadHeaderTimeout: 5 * time.Second, // กัน slowloris
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Info("listening", "addr", ln.Addr().String(), "db", cfg.DatabaseURL != "")

	if cfg.PprofAddr != "" { // dev เท่านั้น: เปิด listener แยก ไม่ปน port หลัก
		dbg := &http.Server{Addr: cfg.PprofAddr, Handler: handler.Debug(svc), ReadHeaderTimeout: 5 * time.Second}
		go func() {
			log.Warn("pprof enabled (dev only)", "addr", cfg.PprofAddr)
			if err := dbg.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error("pprof server", "err", err)
			}
		}()
		defer dbg.Close()
	}

	return app.Serve(ctx, srv, ln, cfg.ShutdownTimeout, log)
}

func openRepo(ctx context.Context, cfg config.Config, log *slog.Logger) (service.Repository, func(), error) {
	if cfg.DatabaseURL == "" {
		log.Warn("DATABASE_URL ว่าง: ใช้ in-memory repository")
		return repository.NewMemory(), func() {}, nil
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL) // driver ถูก register โดย driver_pgx.go (build tag pgx)
	if err != nil {
		return nil, nil, fmt.Errorf("open db (ลืม -tags pgx หรือยังไม่ได้ go get pgx?): %w", err)
	}
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnLifetime)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("ping db: %w", err)
	}
	return repository.NewPostgres(db), func() { db.Close() }, nil
}
