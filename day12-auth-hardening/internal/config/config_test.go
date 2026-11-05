package config

import (
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/auth"
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
	t.Setenv("JWT_KEYS_DIR", "")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.MetricsAddr != ":9090" || c.JWTIssuer != "claim-tracker" || c.JWTAudience != "claims-api" || c.OTLPEndpoint != "" || c.JWTSigningKID != "k1" {
		t.Fatalf("defaults %+v", c)
	}
	if _, err := c.Verifier(); err == nil {
		t.Fatal("api must not start without JWT_KEYS_DIR")
	}
	dir := t.TempDir()
	t.Setenv("JWT_KEYS_DIR", dir)
	c, _ = Load()
	if _, err := c.Verifier(); err == nil {
		t.Fatal("an empty key directory must be refused")
	}
	if err := auth.WriteKeyPair(dir, "k1"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JWT_ISSUER", "other-issuer")
	c, _ = Load()
	if _, err := c.Verifier(); err != nil || c.JWTIssuer != "other-issuer" {
		t.Fatalf("%v %+v", err, c)
	}
	t.Setenv("JWT_SIGNING_KID", "k2") // no k2.key in the directory
	c, _ = Load()
	if _, err := c.Verifier(); err == nil {
		t.Fatal("a signing kid without a key file must be refused")
	}
	if err := auth.WriteKeyPair(dir, "k2"); err != nil {
		t.Fatal(err)
	}
	c, _ = Load()
	v, err := c.Verifier()
	if err != nil {
		t.Fatal(err)
	}
	tok, _ := v.Mint("alice", "", time.Minute)
	if _, err := v.Verify(tok); err != nil {
		t.Fatalf("token signed by the new kid: %v", err)
	}
}

func TestWorkerLoadsWithoutJWTKeys(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_KEYS_DIR", "")
	if _, err := Load(); err != nil {
		t.Fatalf("worker does not need JWT keys: %v", err)
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
