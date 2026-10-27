package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunAllJobs(t *testing.T) {
	var buf bytes.Buffer
	done, total := run(context.Background(), &buf, 3, 10, time.Millisecond)
	if done != 10 || total != 10 {
		t.Fatalf("done=%d total=%d", done, total)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 10 || lines[0] != "job 1: 1" || lines[9] != "job 10: 100" {
		t.Fatalf("output=%q", buf.String())
	}
}

func TestRunCancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var buf bytes.Buffer
	done, total := run(ctx, &buf, 3, 10, time.Millisecond)
	if done != 0 || total != 10 || buf.Len() != 0 {
		t.Fatalf("done=%d total=%d out=%q", done, total, buf.String())
	}
}

func TestRunCancelMidwayStopsEarly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	var buf bytes.Buffer
	done, total := run(ctx, &buf, 2, 1000, 20*time.Millisecond)
	if done >= total {
		t.Fatalf("ควรหยุดก่อนครบ done=%d", done)
	}
}
