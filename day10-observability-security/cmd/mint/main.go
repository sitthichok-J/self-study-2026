// Command mint prints a short-lived dev JWT. It reads the same JWT_* variables
// as the API. Demo tool only: a real deployment gets tokens from an IdP.
//
//	JWT_SECRET=... go run ./cmd/mint -sub alice [-role staff] [-ttl 15m]
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/auth"
)

func main() {
	sub := flag.String("sub", "", "subject (user id)")
	role := flag.String("role", "user", "user or staff")
	ttl := flag.Duration("ttl", 15*time.Minute, "token lifetime (max 1h)")
	flag.Parse()
	get := func(k, d string) string {
		if v := os.Getenv(k); v != "" {
			return v
		}
		return d
	}
	v, err := auth.NewVerifier([]byte(os.Getenv("JWT_SECRET")), get("JWT_ISSUER", "claim-tracker"), get("JWT_AUDIENCE", "claims-api"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	if *sub == "" || *ttl <= 0 || *ttl > time.Hour {
		fmt.Fprintln(os.Stderr, "error: -sub is required and -ttl must be within (0, 1h]")
		os.Exit(2)
	}
	tok, err := v.Mint(*sub, *role, *ttl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println(tok)
}
