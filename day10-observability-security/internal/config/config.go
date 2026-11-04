// Package config reads settings from the environment (12-factor).
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/auth"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	RedisAddr   string
	Brokers     []string
	Group       string

	MetricsAddr  string // private listener for /metrics (never published to the host)
	OTLPEndpoint string // e.g. http://jaeger:4318 ; empty = no export

	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
}

// Verifier builds the JWT verifier. The API calls it at start-up and refuses
// to run without a strong secret; the worker does not need it.
func (c Config) Verifier() (*auth.Verifier, error) {
	return auth.NewVerifier([]byte(c.JWTSecret), c.JWTIssuer, c.JWTAudience)
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr:    get("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisAddr:   get("REDIS_ADDR", "valkey:6379"),
		Group:       get("KAFKA_GROUP", "claims-worker"),

		MetricsAddr:  get("METRICS_ADDR", ":9090"),
		OTLPEndpoint: os.Getenv("OTLP_ENDPOINT"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		JWTIssuer:    get("JWT_ISSUER", "claim-tracker"),
		JWTAudience:  get("JWT_AUDIENCE", "claims-api"),
	}
	for _, b := range strings.Split(get("KAFKA_BROKERS", "kafka:9092"), ",") {
		if b = strings.TrimSpace(b); b != "" {
			c.Brokers = append(c.Brokers, b)
		}
	}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	if len(c.Brokers) == 0 {
		return c, fmt.Errorf("KAFKA_BROKERS is empty")
	}
	return c, nil
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
