// Package auth verifies HS256 JWTs with the standard library only.
//
// Rules (OWASP A07): the algorithm is fixed by the server, never read from the
// token; "none" and every other alg are rejected; signature compare is
// constant time; exp, iss, aud and sub are all required; lifetime is capped.
package auth

import (
	"context"
	"crypto/hmac"
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
)

// Reasons is every value Reason can return (used to pre-create metric series).
var Reasons = []string{"missing", "malformed", "alg", "signature", "expired", "not_yet_valid", "issuer", "audience", "too_long", "claims"}

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
	default:
		return "claims"
	}
}

const (
	RoleUser  = "user"
	RoleStaff = "staff"

	MinSecretLen = 32
	maxTokenLen  = 4096
)

// Principal is who is calling. Subject is the owner id used for access checks.
type Principal struct {
	Subject string
	Staff   bool
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
}

var subRe = regexp.MustCompile(`^[A-Za-z0-9._:@-]{1,64}$`)

type Verifier struct {
	secret   []byte
	iss, aud string
	Now      func() time.Time
	Leeway   time.Duration
	MaxTTL   time.Duration
}

func NewVerifier(secret []byte, iss, aud string) (*Verifier, error) {
	if len(secret) < MinSecretLen {
		return nil, fmt.Errorf("JWT secret must be at least %d bytes", MinSecretLen)
	}
	if iss == "" || aud == "" {
		return nil, errors.New("issuer and audience are required")
	}
	return &Verifier{secret: secret, iss: iss, aud: aud, Now: time.Now, Leeway: 30 * time.Second, MaxTTL: time.Hour}, nil
}

var b64 = base64.RawURLEncoding

func (v *Verifier) sign(signingInput string) []byte {
	m := hmac.New(sha256.New, v.secret)
	m.Write([]byte(signingInput))
	return m.Sum(nil)
}

// Mint issues a token (dev tool and tests; the real issuer would be an IdP).
func (v *Verifier) Mint(sub, role string, ttl time.Duration) (string, error) {
	return v.MintAt(sub, role, ttl, v.Now())
}

// MintAt is Mint with an explicit issue time (the session service passes its own clock).
func (v *Verifier) MintAt(sub, role string, ttl time.Duration, now time.Time) (string, error) {
	h, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	p, _ := json.Marshal(map[string]any{
		"sub": sub, "iss": v.iss, "aud": v.aud, "role": role,
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
	}
	if err := json.Unmarshal(hb, &hdr); err != nil {
		return Principal{}, ErrMalformed
	}
	if hdr.Alg != "HS256" { // exact match: no "none", no "hs256", no RS256 confusion
		return Principal{}, ErrAlg
	}
	sig, err := b64.DecodeString(parts[2])
	if err != nil {
		return Principal{}, ErrMalformed
	}
	if !hmac.Equal(sig, v.sign(parts[0]+"."+parts[1])) {
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
	switch c.Role {
	case "", RoleUser:
		return Principal{Subject: c.Sub}, nil
	case RoleStaff:
		return Principal{Subject: c.Sub, Staff: true}, nil
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
