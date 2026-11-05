// Command keygen creates an RSA key pair for signing JWTs:
//
//	keygen -dir keys -kid k1      -> keys/k1.key (private) and keys/k1.pub (public)
//
// To rotate: run it again with a new kid (k2), set JWT_SIGNING_KID=k2, re-render
// the gateway config, and delete the old .pub once its tokens have expired.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/auth"
)

func main() {
	dir := flag.String("dir", "keys", "directory for the key files")
	kid := flag.String("kid", "", "key id, e.g. k1 (letters, digits, - and _)")
	flag.Parse()
	if *kid == "" {
		fmt.Fprintln(os.Stderr, "error: -kid is required")
		os.Exit(2)
	}
	if err := auth.WriteKeyPair(*dir, *kid); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("created %s/%s.key and %s/%s.pub\n", *dir, *kid, *dir, *kid)
}
