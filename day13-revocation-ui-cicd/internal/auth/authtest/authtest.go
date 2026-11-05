// Package authtest gives other packages' tests ready-made RSA key sets.
// Test code only: nothing in cmd/ imports it.
package authtest

import (
	"sync"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth"
)

var (
	mu    sync.Mutex
	cache = map[string]*auth.KeySet{}
)

// KeySet returns one cached key set per kid (RSA generation is the slow part of a test run).
func KeySet(kid string) *auth.KeySet {
	mu.Lock()
	defer mu.Unlock()
	if ks, ok := cache[kid]; ok {
		return ks
	}
	k, err := auth.GenerateKey()
	if err != nil {
		panic(err)
	}
	ks, err := auth.NewKeySet(kid, k, nil)
	if err != nil {
		panic(err)
	}
	cache[kid] = ks
	return ks
}
