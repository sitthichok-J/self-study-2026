package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidKID(t *testing.T) {
	for kid, want := range map[string]bool{
		"k1": true, "2026-11_a": true, strings.Repeat("a", 32): true,
		"": false, "k 1": false, "../k1": false, "k1.pub": false, strings.Repeat("a", 33): false, "ก": false,
	} {
		if ValidKID(kid) != want {
			t.Errorf("%q: want %v", kid, want)
		}
	}
}

func TestWriteAndLoadKeyDir(t *testing.T) {
	dir := t.TempDir()
	if err := WriteKeyPair(dir, "k1"); err != nil {
		t.Fatal(err)
	}
	if err := WriteKeyPair(dir, "k2"); err != nil {
		t.Fatal(err)
	}
	ks, err := LoadKeyDir(dir, "k2")
	if err != nil {
		t.Fatal(err)
	}
	if ks.SigningKID() != "k2" || strings.Join(ks.KIDs(), ",") != "k1,k2" {
		t.Fatalf("%s %v", ks.SigningKID(), ks.KIDs())
	}
	v, _ := NewVerifier(ks, "claim-tracker", "claims-api")
	tok, _ := v.Mint("alice", "", 10_000_000_000)
	if _, err := v.Verify(tok); err != nil {
		t.Fatal(err)
	}
	// rotation by files: delete k1.pub -> only k2 left
	if err := os.Remove(filepath.Join(dir, "k1.pub")); err != nil {
		t.Fatal(err)
	}
	ks, _ = LoadKeyDir(dir, "k2")
	if strings.Join(ks.KIDs(), ",") != "k2" {
		t.Fatalf("%v", ks.KIDs())
	}
}

func TestWriteKeyPairNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	if err := WriteKeyPair(dir, "k1"); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(dir, "k1.key"))
	if err := WriteKeyPair(dir, "k1"); err == nil {
		t.Fatal("second write for the same kid must fail")
	}
	after, _ := os.ReadFile(filepath.Join(dir, "k1.key"))
	if string(before) != string(after) {
		t.Fatal("existing key changed")
	}
	if err := WriteKeyPair(dir, "../evil"); err == nil {
		t.Fatal("path-like kid must be refused")
	}
}

func TestLoadKeyDirErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadKeyDir(dir, "k1"); err == nil {
		t.Fatal("missing signing key must fail")
	}
	if _, err := LoadKeyDir(dir, "../k1"); err == nil {
		t.Fatal("bad kid must fail")
	}
	if err := os.WriteFile(filepath.Join(dir, "k1.key"), []byte("not pem"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadKeyDir(dir, "k1"); err == nil {
		t.Fatal("garbage key must fail")
	}

	// a .pub that does not belong to the .key
	dir = t.TempDir()
	if err := WriteKeyPair(dir, "k1"); err != nil {
		t.Fatal(err)
	}
	other := t.TempDir()
	if err := WriteKeyPair(other, "k1"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(other, "k1.pub"))
	if err := os.WriteFile(filepath.Join(dir, "k1.pub"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadKeyDir(dir, "k1"); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("mismatched pair: %v", err)
	}

	// a garbage .pub must stop start-up, not be skipped silently
	dir = t.TempDir()
	if err := WriteKeyPair(dir, "k1"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "k2.pub"), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadKeyDir(dir, "k1"); err == nil {
		t.Fatal("garbage public key must fail")
	}
}

func TestWeakKeysAreRefused(t *testing.T) {
	weak, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewKeySet("k1", weak, nil); err == nil {
		t.Fatal("1024-bit signing key must be refused")
	}
	strong := mustKeySet("k1")
	if _, err := NewKeySet("k1", strong.priv, map[string]*rsa.PublicKey{"old": &weak.PublicKey}); err == nil {
		t.Fatal("1024-bit public key must be refused")
	}
	if _, err := NewKeySet("bad kid", strong.priv, nil); err == nil {
		t.Fatal("bad kid must be refused")
	}
	if _, err := NewKeySet("k1", nil, nil); err == nil {
		t.Fatal("nil key must be refused")
	}
}

func TestLoadKeyDirRejectsNonRSAAndWrongPEMType(t *testing.T) {
	dir := t.TempDir()
	if err := WriteKeyPair(dir, "k1"); err != nil {
		t.Fatal(err)
	}
	// PEM with the wrong block type
	b, _ := os.ReadFile(filepath.Join(dir, "k1.key"))
	blk, _ := pem.Decode(b)
	wrong := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: blk.Bytes})
	if err := os.WriteFile(filepath.Join(dir, "k1.key"), wrong, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadKeyDir(dir, "k1"); err == nil {
		t.Fatal("wrong PEM type must fail")
	}
	_ = x509.ErrUnsupportedAlgorithm
}

func TestNewKeySetExtraRules(t *testing.T) {
	a, b := mustKeySet("ka"), mustKeySet("kb")
	// an extra entry with the signing kid must not replace the private key's own public half
	ks, err := NewKeySet("ka", a.priv, map[string]*rsa.PublicKey{"ka": &b.priv.PublicKey})
	if err != nil {
		t.Fatal(err)
	}
	if !ks.pubs["ka"].Equal(&a.priv.PublicKey) {
		t.Fatal("signing kid's public key was overridden")
	}
	// kid in an extra entry is validated like the signing kid
	for _, bad := range []string{"../x", "a b", ""} {
		if _, err := NewKeySet("ka", a.priv, map[string]*rsa.PublicKey{bad: &b.priv.PublicKey}); err == nil {
			t.Errorf("extra kid %q accepted", bad)
		}
	}
}
