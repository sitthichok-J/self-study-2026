// Command mint prints a short-lived dev JWT. It reads the same JWT_* variables
// as the API. Demo tool only: a real deployment gets tokens from an IdP.
//
//	JWT_KEYS_DIR=keys JWT_SIGNING_KID=k1 go run ./cmd/mint -sub alice [-role staff] [-ttl 15m]
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth"
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
	ks, err := auth.LoadKeyDir(os.Getenv("JWT_KEYS_DIR"), get("JWT_SIGNING_KID", "k1"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	v, err := auth.NewVerifier(ks, get("JWT_ISSUER", "claim-tracker"), get("JWT_AUDIENCE", "claims-api"))
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
