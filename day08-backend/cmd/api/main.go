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

	apihttp "github.com/sitthichok-J/self-study-2026/day08-backend/internal/adapter/http"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/adapter/kafka"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/adapter/postgres"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/adapter/redis"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/config"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/service"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repo, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer repo.Close()
	cache := redis.New(cfg.RedisAddr)
	defer cache.Close()
	pub := kafka.NewPublisher(cfg.Brokers)
	defer pub.Close()

	svc := service.NewClaims(repo, cache, log)
	relay := service.NewRelay(repo, pub, log)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); relay.Run(ctx) }()

	srv := &nethttp.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           apihttp.New(svc, repo.Ping, log),
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
