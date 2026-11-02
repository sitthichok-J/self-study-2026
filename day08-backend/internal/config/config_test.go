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
