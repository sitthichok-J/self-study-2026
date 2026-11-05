// Package config reads settings from the environment (12-factor).
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/auth"
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

	// Day 11: login / refresh
	AccessTTL           time.Duration // ACCESS_TTL, default 10m (never above the verifier's 1h cap)
	RefreshTTL          time.Duration // REFRESH_TTL, default 12h
	RefreshCookiePath   string        // path the browser sees behind the gateway, default /api/auth
	RefreshCookieSecure bool          // REFRESH_COOKIE_SECURE, default true; "false" only for plain-http local dev
	DemoPassword        string        // DEMO_PASSWORD: if set, the API seeds users alice, bob and ops(staff) with it
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

		RefreshCookiePath:   get("REFRESH_COOKIE_PATH", "/api/auth"),
		RefreshCookieSecure: true,
		DemoPassword:        os.Getenv("DEMO_PASSWORD"),
	}
	var err error
	if c.AccessTTL, err = duration("ACCESS_TTL", 10*time.Minute); err != nil {
		return c, err
	}
	if c.RefreshTTL, err = duration("REFRESH_TTL", 12*time.Hour); err != nil {
		return c, err
	}
	switch v := strings.ToLower(os.Getenv("REFRESH_COOKIE_SECURE")); v {
	case "", "true", "1":
	case "false", "0":
		c.RefreshCookieSecure = false
	default:
		return c, fmt.Errorf("REFRESH_COOKIE_SECURE must be true or false, got %q", v)
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

func duration(k string, def time.Duration) (time.Duration, error) {
	v := os.Getenv(k)
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration like 10m, got %q", k, v)
	}
	return d, nil
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
