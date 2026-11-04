// Command api serves the REST API and runs the outbox relay.
package main

import (
	"context"
	"errors"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	apihttp "github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/http"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/kafka"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/postgres"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/redis"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/config"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/service"
)

func main() {
	log := obs.NewLogger(os.Stdout, slog.LevelInfo)
	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ver, err := cfg.Verifier()
	if err != nil {
		return err // no JWT secret / too short: refuse to start unauthenticated
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	flushTraces := obs.SetupTracing("claims-api", cfg.OTLPEndpoint)
	defer func() {
		fctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = flushTraces(fctx)
	}()

	repo, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer repo.Close()
	cache := redis.New(cfg.RedisAddr)
	defer cache.Close()
	pub := kafka.NewPublisher(cfg.Brokers)
	defer pub.Close()

	registerPoolGauges(repo)
	go obs.ServeMetrics(ctx, cfg.MetricsAddr, obs.M().Reg, log)

	svc := service.NewClaims(repo, cache, log)
	relay := service.NewRelay(repo, pub, log)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); relay.Run(ctx) }()

	srv := &nethttp.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           apihttp.New(svc, repo.Ping, log, ver),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Info("api listening", "addr", cfg.HTTPAddr)

	select {
	case err := <-errc:
		if !errors.Is(err, nethttp.ErrServerClosed) {
			stop()
			wg.Wait()
			return err
		}
	case <-ctx.Done():
		log.Info("shutting down")
	}
	sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = srv.Shutdown(sctx) // stop accepting, finish in-flight
	wg.Wait()                // relay returns once ctx is cancelled
	return err
}

// registerPoolGauges exposes the connection pool (USE method: utilization and
// saturation of the one shared resource every request needs).
func registerPoolGauges(repo *postgres.Repo) {
	r := obs.M().Reg
	r.NewGaugeFunc("db_pool_acquired_conns", "Connections in use.", func() float64 { a, _, _, _ := repo.PoolStats(); return float64(a) })
	r.NewGaugeFunc("db_pool_idle_conns", "Idle connections.", func() float64 { _, i, _, _ := repo.PoolStats(); return float64(i) })
	r.NewGaugeFunc("db_pool_total_conns", "Open connections.", func() float64 { _, _, t, _ := repo.PoolStats(); return float64(t) })
	r.NewGaugeFunc("db_pool_max_conns", "Pool size limit.", func() float64 { _, _, _, m := repo.PoolStats(); return float64(m) })
}
