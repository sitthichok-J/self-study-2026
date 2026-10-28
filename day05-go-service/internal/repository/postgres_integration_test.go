//go:build integration && pgx

package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/service"
)

// รัน: docker compose up -d db
//
//	TEST_DATABASE_URL=postgres://claims:claims@localhost:5433/claims?sslmode=disable \
//	  go test -tags "integration pgx" ./internal/repository
//
// (ใช้ docker compose + env แทน testcontainers-go เพราะตอนเขียนดึง testcontainers ไม่ได้)
func TestPostgresContract(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("ตั้ง TEST_DATABASE_URL เพื่อรัน integration test")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("เชื่อม DB ไม่ได้: %v", err)
	}
	up, err := os.ReadFile("../../migrations/000001_create_claims.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	down, err := os.ReadFile("../../migrations/000001_create_claims.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	// เริ่มจากตารางสะอาดทุกครั้ง และทิ้งไว้สะอาดหลังจบ
	db.ExecContext(ctx, string(down))
	if _, err := db.ExecContext(ctx, string(up)); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(context.Background(), string(down)) })

	runContract(t, func(t *testing.T) service.Repository {
		if _, err := db.ExecContext(ctx, "TRUNCATE claims RESTART IDENTITY"); err != nil {
			t.Fatal(err)
		}
		return NewPostgres(db)
	})
}
