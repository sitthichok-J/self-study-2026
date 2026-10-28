// Package config: อ่านค่าตั้งจาก environment (รับฟังก์ชัน getenv เพื่อให้เทสต์ได้โดยไม่แตะ env จริง)
package config

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr            string        // HTTP หลัก
	DatabaseURL     string        // ว่าง = ใช้ in-memory
	LogLevel        slog.Level    // DEBUG/INFO/WARN/ERROR
	ShutdownTimeout time.Duration // เวลารอให้ request ที่ค้างอยู่จบตอนปิด
	PprofAddr       string        // ว่าง = ปิด pprof (เปิดเฉพาะ dev และควรเป็น localhost)
	DBMaxOpenConns  int
	DBMaxIdleConns  int
	DBConnLifetime  time.Duration
}

func Load(getenv func(string) string) (Config, error) {
	c := Config{
		Addr:            or(getenv("ADDR"), ":8080"),
		DatabaseURL:     getenv("DATABASE_URL"),
		PprofAddr:       getenv("PPROF_ADDR"),
		ShutdownTimeout: 10 * time.Second,
		DBMaxOpenConns:  10,
		DBMaxIdleConns:  5,
		DBConnLifetime:  30 * time.Minute,
	}
	if v := getenv("LOG_LEVEL"); v != "" {
		if err := c.LogLevel.UnmarshalText([]byte(strings.ToUpper(v))); err != nil {
			return Config{}, fmt.Errorf("LOG_LEVEL %q: %w", v, err)
		}
	}
	var err error
	if c.ShutdownTimeout, err = duration(getenv, "SHUTDOWN_TIMEOUT", c.ShutdownTimeout); err != nil {
		return Config{}, err
	}
	if c.DBConnLifetime, err = duration(getenv, "DB_CONN_LIFETIME", c.DBConnLifetime); err != nil {
		return Config{}, err
	}
	if c.DBMaxOpenConns, err = integer(getenv, "DB_MAX_OPEN_CONNS", c.DBMaxOpenConns); err != nil {
		return Config{}, err
	}
	if c.DBMaxIdleConns, err = integer(getenv, "DB_MAX_IDLE_CONNS", c.DBMaxIdleConns); err != nil {
		return Config{}, err
	}
	return c, nil
}

func or(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func duration(getenv func(string) string, key string, def time.Duration) (time.Duration, error) {
	v := getenv(key)
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("%s %q: ต้องเป็นช่วงเวลาที่มากกว่า 0 เช่น 10s", key, v)
	}
	return d, nil
}

func integer(getenv func(string) string, key string, def int) (int, error) {
	v := getenv(key)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s %q: ต้องเป็นจำนวนเต็มที่มากกว่า 0", key, v)
	}
	return n, nil
}
