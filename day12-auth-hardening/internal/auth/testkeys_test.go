package auth

import (
	"crypto/x509"
	"encoding/pem"
	"sync"
	"testing"
)

var (
	keyMu    sync.Mutex
	keyCache = map[string]*KeySet{}
)

// mustKeySet returns one cached RSA key set per kid (generating 2048-bit keys is the slow part).
func mustKeySet(kid string) *KeySet {
	keyMu.Lock()
	defer keyMu.Unlock()
	if ks, ok := keyCache[kid]; ok {
		return ks
	}
	k, err := GenerateKey()
	if err != nil {
		panic(err)
	}
	ks, err := NewKeySet(kid, k, nil)
	if err != nil {
		panic(err)
	}
	keyCache[kid] = ks
	return ks
}

func publicPEM(t *testing.T, ks *KeySet, kid string) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(ks.pubs[kid])
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}
