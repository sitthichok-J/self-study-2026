package auth

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

var fast = Hasher{Iterations: MinIterations} // real default is 600k: too slow for unit tests

func TestHashVerifyRoundTrip(t *testing.T) {
	h, err := fast.Hash("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "pbkdf2-sha256$1000$") {
		t.Fatalf("unexpected format %q", h)
	}
	if !fast.Verify("correct horse battery", h) {
		t.Fatal("right password rejected")
	}
	for _, bad := range []string{"correct horse batterY", "", "correct horse battery ", strings.Repeat("x", 300)} {
		if fast.Verify(bad, h) {
			t.Fatalf("wrong password %q accepted", bad)
		}
	}
}

func TestHashIsSalted(t *testing.T) {
	a, _ := fast.Hash("same-password-123")
	b, _ := fast.Hash("same-password-123")
	if a == b {
		t.Fatal("two hashes of one password must differ (random salt)")
	}
	if strings.Contains(a, "same-password-123") {
		t.Fatal("hash contains the password")
	}
}

func TestHashRejectsBadInput(t *testing.T) {
	if _, err := fast.Hash("short"); err != ErrPasswordShort {
		t.Fatalf("short password: %v", err)
	}
	if _, err := fast.Hash(strings.Repeat("x", 257)); err != ErrPasswordLong {
		t.Fatalf("long password: %v", err)
	}
	if _, err := (Hasher{Iterations: 10}).Hash("long-enough-password"); err == nil {
		t.Fatal("iteration count below the floor must be refused")
	}
}

func TestDefaultIterations(t *testing.T) {
	if (Hasher{}).iters() != DefaultIterations || DefaultIterations < 600_000 {
		t.Fatal("default iterations must follow OWASP (>= 600000)")
	}
}

func TestVerifyMalformedStoredHash(t *testing.T) {
	good, _ := fast.Hash("correct horse battery")
	parts := strings.Split(good, "$")
	cases := map[string]string{
		"empty":           "",
		"plain password":  "correct horse battery",
		"wrong scheme":    "bcrypt$" + strings.Join(parts[1:], "$"),
		"too few parts":   strings.Join(parts[:3], "$"),
		"too many parts":  good + "$x",
		"iterations NaN":  parts[0] + "$abc$" + parts[2] + "$" + parts[3],
		"iterations low":  parts[0] + "$10$" + parts[2] + "$" + parts[3],
		"iterations huge": parts[0] + "$999999999$" + parts[2] + "$" + parts[3],
		"salt not b64":    parts[0] + "$1000$!!!$" + parts[3],
		"salt too short":  parts[0] + "$1000$AAAA$" + parts[3],
		"hash not b64":    parts[0] + "$1000$" + parts[2] + "$!!!",
		"hash too short":  parts[0] + "$1000$" + parts[2] + "$AAAA",
	}
	for name, stored := range cases {
		if fast.Verify("correct horse battery", stored) {
			t.Errorf("%s: malformed hash accepted", name)
		}
	}
}

func TestIterationsComeFromTheStoredHash(t *testing.T) {
	old, _ := (Hasher{Iterations: 2000}).Hash("correct horse battery")
	if !(Hasher{Iterations: MinIterations}).Verify("correct horse battery", old) {
		t.Fatal("a hash made with a different iteration count must still verify")
	}
}

func TestVerifyRefusesAWeakStoredIterationCount(t *testing.T) {
	// A well-formed hash that really was made with 10 iterations: the maths matches,
	// but a count below the floor must still be refused (an attacker who can write
	// to the table could otherwise plant cheap-to-crack hashes).
	salt := []byte("0123456789abcdef")
	key, err := pbkdf2.Key(sha256.New, "correct horse battery", salt, 10, keyLen)
	if err != nil {
		t.Fatal(err)
	}
	weak := fmt.Sprintf("pbkdf2-sha256$10$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key))
	if fast.Verify("correct horse battery", weak) {
		t.Fatal("hash with 10 iterations accepted")
	}
}
