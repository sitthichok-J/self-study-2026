package auth

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Passwords are stored as PBKDF2-HMAC-SHA256 (stdlib crypto/pbkdf2, Go 1.24+):
//
//	pbkdf2-sha256$<iterations>$<salt b64>$<hash b64>
//
// The iteration count is part of the stored string, so it can be raised later
// without breaking old hashes.
const (
	// DefaultIterations follows the OWASP guidance for PBKDF2-HMAC-SHA256.
	DefaultIterations = 600_000
	// MinIterations is the floor Verify accepts; a lower stored count is treated as corrupt.
	MinIterations  = 1_000
	saltLen        = 16
	keyLen         = 32
	MinPasswordLen = 12
	maxPasswordLen = 256 // PBKDF2 cost grows with input size: cap it
	scheme         = "pbkdf2-sha256"
)

var (
	ErrPasswordShort = fmt.Errorf("password must be at least %d characters", MinPasswordLen)
	ErrPasswordLong  = fmt.Errorf("password must be at most %d characters", maxPasswordLen)
)

// Hasher hashes and checks passwords. Iterations is only lowered in tests.
type Hasher struct{ Iterations int }

func (h Hasher) iters() int {
	if h.Iterations <= 0 {
		return DefaultIterations
	}
	return h.Iterations
}

func (h Hasher) Hash(password string) (string, error) {
	if len(password) < MinPasswordLen {
		return "", ErrPasswordShort
	}
	if len(password) > maxPasswordLen {
		return "", ErrPasswordLong
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	n := h.iters()
	if n < MinIterations {
		return "", fmt.Errorf("iterations below minimum %d", MinIterations)
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, n, keyLen)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s$%d$%s$%s", scheme, n, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

// Verify reports whether password matches the stored hash. Any malformed
// stored value counts as "no match" (never an error that leaks which part failed).
func (h Hasher) Verify(password, stored string) bool {
	if len(password) > maxPasswordLen {
		return false
	}
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != scheme {
		return false
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil || n < MinIterations || n > 10_000_000 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) < 8 {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(want) != keyLen {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, n, keyLen)
	if err != nil {
		return false
	}
	return hmac.Equal(got, want)
}
