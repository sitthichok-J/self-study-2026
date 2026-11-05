package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("DATABASE_URL must be required")
	}
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("KAFKA_BROKERS", "a:1, b:2 ,,")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Brokers) != 2 || c.Brokers[1] != "b:2" {
		t.Fatalf("brokers %v", c.Brokers)
	}
	if c.HTTPAddr != ":8080" || c.Group != "claims-worker" {
		t.Fatalf("defaults %+v", c)
	}
}

func TestSecurityDefaultsAndVerifier(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", "")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.MetricsAddr != ":9090" || c.JWTIssuer != "claim-tracker" || c.JWTAudience != "claims-api" || c.OTLPEndpoint != "" {
		t.Fatalf("defaults %+v", c)
	}
	if _, err := c.Verifier(); err == nil {
		t.Fatal("api must not start without a JWT secret")
	}
	t.Setenv("JWT_SECRET", "too-short")
	c, _ = Load()
	if _, err := c.Verifier(); err == nil {
		t.Fatal("a short secret must be refused")
	}
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("JWT_ISSUER", "other-issuer")
	c, _ = Load()
	if _, err := c.Verifier(); err != nil || c.JWTIssuer != "other-issuer" {
		t.Fatalf("%v %+v", err, c)
	}
}

func TestWorkerLoadsWithoutJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", "")
	if _, err := Load(); err != nil {
		t.Fatalf("worker does not need a JWT secret: %v", err)
	}
}

func TestSessionDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.AccessTTL != 10*time.Minute || c.RefreshTTL != 12*time.Hour {
		t.Fatalf("ttl defaults %v %v", c.AccessTTL, c.RefreshTTL)
	}
	if c.RefreshCookiePath != "/api/auth" || !c.RefreshCookieSecure || c.DemoPassword != "" {
		t.Fatalf("cookie defaults %+v", c)
	}
}

func TestSessionOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("ACCESS_TTL", "5m")
	t.Setenv("REFRESH_TTL", "36h")
	t.Setenv("REFRESH_COOKIE_PATH", "/auth")
	t.Setenv("REFRESH_COOKIE_SECURE", "false")
	t.Setenv("DEMO_PASSWORD", "demo-fake-password-1")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.AccessTTL != 5*time.Minute || c.RefreshTTL != 36*time.Hour || c.RefreshCookiePath != "/auth" || c.RefreshCookieSecure || c.DemoPassword == "" {
		t.Fatalf("%+v", c)
	}
}

func TestSessionBadValuesAreRefused(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	for _, kv := range [][2]string{
		{"ACCESS_TTL", "soon"}, {"ACCESS_TTL", "0"}, {"ACCESS_TTL", "-5m"}, {"REFRESH_TTL", "1 day"},
		{"REFRESH_COOKIE_SECURE", "maybe"},
	} {
		t.Run(kv[0]+"="+kv[1], func(t *testing.T) {
			t.Setenv(kv[0], kv[1])
			if _, err := Load(); err == nil {
				t.Fatalf("%s=%s accepted", kv[0], kv[1])
			}
		})
	}
}
