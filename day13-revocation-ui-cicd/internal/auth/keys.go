package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var kidRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,32}$`)

// ValidKID reports whether kid is safe to put in a file name and a JWT header.
func ValidKID(kid string) bool { return kidRe.MatchString(kid) }

// KeySet is one signing key plus every public key the server still accepts.
// Rotation = add a new pair, point the signing kid at it, keep the old public
// key until tokens signed with it have expired, then delete it.
type KeySet struct {
	signKid string
	priv    *rsa.PrivateKey
	pubs    map[string]*rsa.PublicKey
}

// NewKeySet builds a set. extra holds the other accepted public keys (may be nil).
func NewKeySet(signKid string, priv *rsa.PrivateKey, extra map[string]*rsa.PublicKey) (*KeySet, error) {
	if !ValidKID(signKid) {
		return nil, fmt.Errorf("invalid signing kid %q", signKid)
	}
	if priv == nil {
		return nil, errors.New("a signing key is required")
	}
	if priv.N.BitLen() < MinRSABits {
		return nil, fmt.Errorf("signing key must be at least %d bits", MinRSABits)
	}
	pubs := map[string]*rsa.PublicKey{signKid: &priv.PublicKey}
	for kid, p := range extra {
		if kid == signKid {
			continue // the private key's own public half wins
		}
		if !ValidKID(kid) {
			return nil, fmt.Errorf("invalid kid %q", kid)
		}
		if p == nil || p.N.BitLen() < MinRSABits {
			return nil, fmt.Errorf("public key %q must be at least %d bits", kid, MinRSABits)
		}
		pubs[kid] = p
	}
	return &KeySet{signKid: signKid, priv: priv, pubs: pubs}, nil
}

// GenerateKey makes a new 2048-bit RSA key.
func GenerateKey() (*rsa.PrivateKey, error) { return rsa.GenerateKey(rand.Reader, MinRSABits) }

// SigningKID is the kid written into new tokens.
func (k *KeySet) SigningKID() string { return k.signKid }

// KIDs lists every accepted kid, sorted.
func (k *KeySet) KIDs() []string {
	out := make([]string, 0, len(k.pubs))
	for kid := range k.pubs {
		out = append(out, kid)
	}
	sort.Strings(out)
	return out
}

// WriteKeyPair creates <dir>/<kid>.key (PKCS#8) and <dir>/<kid>.pub (PKIX).
// It refuses to overwrite: an existing kid must never change its key.
func WriteKeyPair(dir, kid string) error {
	if !ValidKID(kid) {
		return fmt.Errorf("invalid kid %q (use letters, digits, - and _, up to 32 characters)", kid)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	keyPath, pubPath := filepath.Join(dir, kid+".key"), filepath.Join(dir, kid+".pub")
	for _, p := range []string{keyPath, pubPath} {
		if _, err := os.Stat(p); err == nil {
			return fmt.Errorf("%s already exists; pick a new kid instead of overwriting", p)
		}
	}
	priv, err := GenerateKey()
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return err
	}
	// 0644: the API container runs as a different (non-root) user and must read it.
	// The directory is git-ignored and bound read-only into the container.
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0o644); err != nil {
		return err
	}
	return os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}), 0o644)
}

// LoadKeyDir reads <signKid>.key as the signing key and every *.pub as an accepted public key.
func LoadKeyDir(dir, signKid string) (*KeySet, error) {
	if !ValidKID(signKid) {
		return nil, fmt.Errorf("invalid signing kid %q", signKid)
	}
	raw, err := os.ReadFile(filepath.Join(dir, signKid+".key"))
	if err != nil {
		return nil, fmt.Errorf("signing key: %w", err)
	}
	blk, _ := pem.Decode(raw)
	if blk == nil || blk.Type != "PRIVATE KEY" {
		return nil, errors.New("signing key must be a PKCS#8 PEM (BEGIN PRIVATE KEY)")
	}
	pk, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		return nil, fmt.Errorf("signing key: %w", err)
	}
	priv, ok := pk.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("signing key must be RSA")
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.pub"))
	if err != nil {
		return nil, err
	}
	extra := map[string]*rsa.PublicKey{}
	for _, f := range files {
		kid := strings.TrimSuffix(filepath.Base(f), ".pub")
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		blk, _ := pem.Decode(b)
		if blk == nil || blk.Type != "PUBLIC KEY" {
			return nil, fmt.Errorf("%s: not a PKIX public key PEM", filepath.Base(f))
		}
		pub, err := x509.ParsePKIXPublicKey(blk.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(f), err)
		}
		rp, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("%s: not an RSA key", filepath.Base(f))
		}
		extra[kid] = rp
	}
	if sp, ok := extra[signKid]; ok && !sp.Equal(&priv.PublicKey) {
		return nil, fmt.Errorf("%s.pub does not match %s.key", signKid, signKid)
	}
	return NewKeySet(signKid, priv, extra)
}
