package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestDefaults(t *testing.T) {
	c, err := Load(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.Addr != ":8080" || c.DatabaseURL != "" || c.PprofAddr != "" || c.LogLevel != slog.LevelInfo ||
		c.ShutdownTimeout != 10*time.Second || c.DBMaxOpenConns != 10 || c.DBMaxIdleConns != 5 || c.DBConnLifetime != 30*time.Minute {
		t.Fatalf("%+v", c)
	}
}

func TestOverrides(t *testing.T) {
	c, err := Load(env(map[string]string{
		"ADDR": ":9000", "DATABASE_URL": "postgres://x", "LOG_LEVEL": "debug", "PPROF_ADDR": "127.0.0.1:6060",
		"SHUTDOWN_TIMEOUT": "3s", "DB_MAX_OPEN_CONNS": "20", "DB_MAX_IDLE_CONNS": "2", "DB_CONN_LIFETIME": "5m",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.Addr != ":9000" || c.DatabaseURL != "postgres://x" || c.LogLevel != slog.LevelDebug || c.PprofAddr != "127.0.0.1:6060" ||
		c.ShutdownTimeout != 3*time.Second || c.DBMaxOpenConns != 20 || c.DBMaxIdleConns != 2 || c.DBConnLifetime != 5*time.Minute {
		t.Fatalf("%+v", c)
	}
}

func TestInvalidValues(t *testing.T) {
	tests := []struct{ key, val, want string }{
		{"LOG_LEVEL", "loud", "LOG_LEVEL"},
		{"SHUTDOWN_TIMEOUT", "abc", "SHUTDOWN_TIMEOUT"},
		{"SHUTDOWN_TIMEOUT", "-5s", "SHUTDOWN_TIMEOUT"},
		{"DB_CONN_LIFETIME", "0s", "DB_CONN_LIFETIME"},
		{"DB_MAX_OPEN_CONNS", "x", "DB_MAX_OPEN_CONNS"},
		{"DB_MAX_OPEN_CONNS", "0", "DB_MAX_OPEN_CONNS"},
		{"DB_MAX_IDLE_CONNS", "-1", "DB_MAX_IDLE_CONNS"},
	}
	for _, tc := range tests {
		t.Run(tc.key+"="+tc.val, func(t *testing.T) {
			_, err := Load(env(map[string]string{tc.key: tc.val}))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}
