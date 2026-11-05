// Package auth signs and verifies RS256 JWTs with the standard library only.
//
// Rules (OWASP A07): the algorithm is fixed by the server, never read from the
// token; "none", HS256 and every other alg are rejected; the key is picked by
// the kid header from a server-side set (so keys can be rotated); exp, iss,
// aud and sub are all required; lifetime is capped.
package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrMissing   = errors.New("missing token")
	ErrMalformed = errors.New("malformed token")
	ErrAlg       = errors.New("unsupported algorithm")
	ErrSignature = errors.New("bad signature")
	ErrExpired   = errors.New("token expired")
	ErrNotYet    = errors.New("token not valid yet")
	ErrIssuer    = errors.New("wrong issuer")
	ErrAudience  = errors.New("wrong audience")
	ErrClaims    = errors.New("invalid claims")
	ErrTooLong   = errors.New("token lifetime too long")
	ErrRevoked   = errors.New("token revoked")
)

// Reasons is every value Reason can return (used to pre-create metric series).
var Reasons = []string{"missing", "malformed", "alg", "signature", "expired", "not_yet_valid", "issuer", "audience", "too_long", "claims", "revoked"}

// Reason is a short, bounded label for metrics and logs.
func Reason(err error) string {
	switch {
	case errors.Is(err, ErrMissing):
		return "missing"
	case errors.Is(err, ErrMalformed):
		return "malformed"
	case errors.Is(err, ErrAlg):
		return "alg"
	case errors.Is(err, ErrSignature):
		return "signature"
	case errors.Is(err, ErrExpired):
		return "expired"
	case errors.Is(err, ErrNotYet):
		return "not_yet_valid"
	case errors.Is(err, ErrIssuer):
		return "issuer"
	case errors.Is(err, ErrAudience):
		return "audience"
	case errors.Is(err, ErrTooLong):
		return "too_long"
	case errors.Is(err, ErrRevoked):
		return "revoked"
	default:
		return "claims"
	}
}

const (
	RoleUser  = "user"
	RoleStaff = "staff"

	MinRSABits  = 2048
	maxTokenLen = 4096
)

// Principal is who is calling. Subject is the owner id used for access checks.
type Principal struct {
	Subject  string
	Staff    bool
	JTI      string    // token id: what a logout puts on the blocklist
	IssuedAt time.Time // zero if the token has no iat
	Expires  time.Time
}

type ctxKey struct{}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}

type audience []string

func (a *audience) UnmarshalJSON(b []byte) error {
	var one string
	if json.Unmarshal(b, &one) == nil {
		*a = audience{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

type claims struct {
	Sub  string   `json:"sub"`
	Iss  string   `json:"iss"`
	Aud  audience `json:"aud"`
	Exp  *int64   `json:"exp"`
	Nbf  *int64   `json:"nbf"`
	Iat  *int64   `json:"iat"`
	Role string   `json:"role"`
	Jti  string   `json:"jti"`
}

var subRe = regexp.MustCompile(`^[A-Za-z0-9._:@-]{1,64}$`)

// jtiRe: every token we mint carries a random jti (16 bytes, base64url = 22 chars). It is required,
// because a token without an id cannot be put on the blocklist.
var jtiRe = regexp.MustCompile(`^[A-Za-z0-9_-]{16,64}$`)

func newJTI() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b64.EncodeToString(b)
}

type Verifier struct {
	keys     *KeySet
	iss, aud string
	Now      func() time.Time
	Leeway   time.Duration
	MaxTTL   time.Duration
}

func NewVerifier(keys *KeySet, iss, aud string) (*Verifier, error) {
	if keys == nil {
		return nil, errors.New("a key set is required")
	}
	if iss == "" || aud == "" {
		return nil, errors.New("issuer and audience are required")
	}
	return &Verifier{keys: keys, iss: iss, aud: aud, Now: time.Now, Leeway: 30 * time.Second, MaxTTL: time.Hour}, nil
}

var b64 = base64.RawURLEncoding

// sign signs with the current signing key (RSASSA-PKCS1-v1_5 over SHA-256).
func (v *Verifier) sign(signingInput string) []byte {
	sum := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, v.keys.priv, crypto.SHA256, sum[:])
	if err != nil {
		panic(err) // only possible with a broken key; NewKeySet rejects those
	}
	return sig
}

// Mint issues a token (dev tool and tests; the real issuer would be an IdP).
func (v *Verifier) Mint(sub, role string, ttl time.Duration) (string, error) {
	return v.MintAt(sub, role, ttl, v.Now())
}

// MintAt is Mint with an explicit issue time (the session service passes its own clock).
func (v *Verifier) MintAt(sub, role string, ttl time.Duration, now time.Time) (string, error) {
	h, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT", "kid": v.keys.signKid})
	p, _ := json.Marshal(map[string]any{
		"sub": sub, "iss": v.iss, "aud": v.aud, "role": role, "jti": newJTI(),
		"iat": now.Unix(), "exp": now.Add(ttl).Unix(),
	})
	in := b64.EncodeToString(h) + "." + b64.EncodeToString(p)
	return in + "." + b64.EncodeToString(v.sign(in)), nil
}

func (v *Verifier) Verify(token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrMissing
	}
	if len(token) > maxTokenLen {
		return Principal{}, ErrMalformed
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Principal{}, ErrMalformed
	}
	hb, err := b64.DecodeString(parts[0])
	if err != nil {
		return Principal{}, ErrMalformed
	}
	var hdr struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(hb, &hdr); err != nil {
		return Principal{}, ErrMalformed
	}
	if hdr.Alg != "RS256" { // exact match: no "none", no HS256 (key-confusion attack), no "rs256"
		return Principal{}, ErrAlg
	}
	pub, ok := v.keys.pubs[hdr.Kid] // unknown or missing kid looks like a bad signature to the caller
	if !ok {
		return Principal{}, ErrSignature
	}
	sig, err := b64.DecodeString(parts[2])
	if err != nil {
		return Principal{}, ErrMalformed
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig) != nil {
		return Principal{}, ErrSignature
	}
	pb, err := b64.DecodeString(parts[1])
	if err != nil {
		return Principal{}, ErrMalformed
	}
	var c claims
	if err := json.Unmarshal(pb, &c); err != nil {
		return Principal{}, ErrMalformed
	}
	now := v.Now()
	if c.Exp == nil {
		return Principal{}, fmt.Errorf("%w: exp is required", ErrClaims)
	}
	exp := time.Unix(*c.Exp, 0)
	if !now.Before(exp.Add(v.Leeway)) {
		return Principal{}, ErrExpired
	}
	if exp.After(now.Add(v.MaxTTL + v.Leeway)) {
		return Principal{}, ErrTooLong
	}
	if c.Nbf != nil && now.Add(v.Leeway).Before(time.Unix(*c.Nbf, 0)) {
		return Principal{}, ErrNotYet
	}
	if c.Iss != v.iss {
		return Principal{}, ErrIssuer
	}
	okAud := false
	for _, a := range c.Aud {
		if a == v.aud {
			okAud = true
		}
	}
	if !okAud {
		return Principal{}, ErrAudience
	}
	if !subRe.MatchString(c.Sub) {
		return Principal{}, fmt.Errorf("%w: sub", ErrClaims)
	}
	if !jtiRe.MatchString(c.Jti) {
		return Principal{}, fmt.Errorf("%w: jti", ErrClaims)
	}
	p := Principal{Subject: c.Sub, JTI: c.Jti, Expires: exp}
	if c.Iat != nil {
		p.IssuedAt = time.Unix(*c.Iat, 0)
	}
	switch c.Role {
	case "", RoleUser:
		return p, nil
	case RoleStaff:
		p.Staff = true
		return p, nil
	}
	return Principal{}, fmt.Errorf("%w: role", ErrClaims)
}

// BearerToken extracts the token from an Authorization header value.
func BearerToken(h string) string {
	if len(h) < 8 || !strings.EqualFold(h[:7], "bearer ") {
		return ""
	}
	return strings.TrimSpace(h[7:])
}

// SigningKID is the kid written into new tokens (for start-up logs and rotation checks).
func (v *Verifier) SigningKID() string { return v.keys.signKid }

// AcceptedKIDs lists the kids whose signatures are still accepted, sorted.
func (v *Verifier) AcceptedKIDs() []string { return v.keys.KIDs() }
