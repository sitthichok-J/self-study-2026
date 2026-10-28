//go:build pgx

package main

// register driver "pgx" สำหรับ database/sql
// ต้อง `go get github.com/jackc/pgx/v5` ก่อน แล้ว build ด้วย -tags pgx
import _ "github.com/jackc/pgx/v5/stdlib"
