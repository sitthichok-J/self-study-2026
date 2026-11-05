// Command worker consumes claims.submitted and decides each claim.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/adapter/kafka"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/adapter/postgres"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/adapter/redis"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/config"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/port"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/service"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	flushTraces := obs.SetupTracing("claims-worker", cfg.OTLPEndpoint)
	defer func() {
		fctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = flushTraces(fctx)
	}()
	go obs.ServeMetrics(ctx, cfg.MetricsAddr, obs.M().Reg, log)

	repo, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer repo.Close()
	cache := redis.New(cfg.RedisAddr)
	defer cache.Close()
	pub := kafka.NewPublisher(cfg.Brokers) // used for the DLQ
	defer pub.Close()

	svc := service.NewClaims(repo, cache, log)
	cons := service.NewConsumer(svc, pub, log)

	log.Info("worker started", "group", cfg.Group, "topic", port.TopicSubmitted)
	err = kafka.Run(ctx, cfg.Brokers, cfg.Group, port.TopicSubmitted, cons.Handle, log)
	log.Info("worker stopped")
	return err
}
