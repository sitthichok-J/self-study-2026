// Package config reads settings from the environment (12-factor).
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	RedisAddr   string
	Brokers     []string
	Group       string
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr:    get("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisAddr:   get("REDIS_ADDR", "valkey:6379"),
		Group:       get("KAFKA_GROUP", "claims-worker"),
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
