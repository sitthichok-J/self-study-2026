package obs

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// SetupTracing installs the process-wide tracer. With an empty endpoint spans
// still get ids (so logs carry trace_id) but nothing is exported.
func SetupTracing(service, otlpEndpoint string) (shutdown func(context.Context) error) {
	if otlpEndpoint == "" {
		SetTracer(NewTracer(service, nil))
		return func(context.Context) error { return nil }
	}
	exp := NewOTLPExporter(otlpEndpoint, service, M())
	SetTracer(NewTracer(service, exp))
	return exp.Shutdown
}

// MetricsMux serves /metrics and a liveness probe. It is meant for a private
// listener (compose network / cluster), never for the public API port.
func MetricsMux(reg *Registry) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", reg.Handler())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
	})
	return mux
}

// ServeMetrics runs MetricsMux until ctx is cancelled.
func ServeMetrics(ctx context.Context, addr string, reg *Registry, log *slog.Logger) {
	srv := &http.Server{
		Addr: addr, Handler: MetricsMux(reg),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second,
	}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	log.Info("metrics listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("metrics server stopped", "err", err)
	}
}
