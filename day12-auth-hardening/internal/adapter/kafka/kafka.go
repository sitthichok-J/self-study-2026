// Package kafka wraps segmentio/kafka-go: a publisher and a consumer loop.
package kafka

import (
	"context"
	"errors"
	"log/slog"
	"time"

	kg "github.com/segmentio/kafka-go"

	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/port"
)

type Publisher struct{ w *kg.Writer }

func NewPublisher(brokers []string) *Publisher {
	return &Publisher{w: &kg.Writer{
		Addr:         kg.TCP(brokers...),
		Balancer:     &kg.Hash{}, // same key -> same partition -> per-claim ordering
		RequiredAcks: kg.RequireAll,
		WriteTimeout: 10 * time.Second,
		// Topic is set per message.
	}}
}

func (p *Publisher) Publish(ctx context.Context, topic, key string, value []byte) error {
	return p.w.WriteMessages(ctx, kg.Message{Topic: topic, Key: []byte(key), Value: value})
}

func (p *Publisher) Close() error { return p.w.Close() }

// Handler returns nil when the message is done (commit), error to keep it.
type Handler func(ctx context.Context, m port.Message) error

// Run consumes until ctx is cancelled. The offset is committed only after the
// handler returns nil. On handler error the same message is retried here
// (after a pause) so a later commit can never skip over it.
func Run(ctx context.Context, brokers []string, group, topic string, h Handler, log *slog.Logger) error {
	r := kg.NewReader(kg.ReaderConfig{
		Brokers: brokers, GroupID: group, Topic: topic,
		MinBytes: 1, MaxBytes: 10 << 20, MaxWait: time.Second,
	})
	defer r.Close()

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, context.Canceled) {
				return nil // shutting down
			}
			return err
		}
		pm := port.Message{Topic: m.Topic, Key: string(m.Key), Value: m.Value}
		for {
			// In-flight work is allowed to finish even after SIGTERM.
			hctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			err := h(hctx, pm)
			cancel()
			if err == nil {
				break
			}
			log.Error("handler failed, will retry same message", "key", pm.Key, "err", err)
			select {
			case <-ctx.Done():
				return nil // not committed -> redelivered on next start
			case <-time.After(2 * time.Second):
			}
		}
		cctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		err = r.CommitMessages(cctx, m)
		cancel()
		if err != nil {
			return err
		}
	}
}
