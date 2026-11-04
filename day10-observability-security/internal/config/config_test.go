package config

import "testing"

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
