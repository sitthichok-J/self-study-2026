package auth

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

var secret = []byte("0123456789abcdef0123456789abcdef")

func newV(t *testing.T) (*Verifier, *time.Time) {
	t.Helper()
	v, err := NewVerifier(secret, "claim-tracker", "claims-api")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 4, 9, 0, 0, 0, time.UTC)
	v.Now = func() time.Time { return now }
	return v, &now
}

// craft signs arbitrary header/payload JSON with the real key.
func craft(v *Verifier, hdr, payload string) string {
	in := b64.EncodeToString([]byte(hdr)) + "." + b64.EncodeToString([]byte(payload))
	return in + "." + b64.EncodeToString(v.sign(in))
}

func payload(v *Verifier, over map[string]any) string {
	m := map[string]any{"sub": "alice", "iss": "claim-tracker", "aud": "claims-api",
		"exp": v.Now().Add(10 * time.Minute).Unix()}
	for k, x := range over {
		if x == nil {
			delete(m, k)
		} else {
			m[k] = x
		}
	}
	b, _ := json.Marshal(m)
	return string(b)
}

const hs = `{"alg":"HS256","typ":"JWT"}`

func TestMintVerifyRoundTrip(t *testing.T) {
	v, _ := newV(t)
	tok, _ := v.Mint("alice", RoleUser, 10*time.Minute)
	p, err := v.Verify(tok)
	if err != nil || p.Subject != "alice" || p.Staff {
		t.Fatalf("%v %+v", err, p)
	}
	tok, _ = v.Mint("bob", RoleStaff, 10*time.Minute)
	if p, _ := v.Verify(tok); !p.Staff {
		t.Fatal("staff role lost")
	}
}

func TestRejects(t *testing.T) {
	v, _ := newV(t)
	good, _ := v.Mint("alice", "", 10*time.Minute)
	parts := strings.Split(good, ".")
	cases := []struct {
		name string
		tok  string
		want error
	}{
		{"empty", "", ErrMissing},
		{"two parts", "a.b", ErrMalformed},
		{"four parts", good + ".x", ErrMalformed},
		{"not base64", "!!!.###.$$$", ErrMalformed},
		{"too long", strings.Repeat("a", 5000), ErrMalformed},
		{"alg none", b64.EncodeToString([]byte(`{"alg":"none"}`)) + "." + parts[1] + ".", ErrAlg},
		{"alg none with sig", craft(v, `{"alg":"none"}`, payload(v, nil)), ErrAlg},
		{"alg lowercase", craft(v, `{"alg":"hs256"}`, payload(v, nil)), ErrAlg},
		{"alg RS256", craft(v, `{"alg":"RS256"}`, payload(v, nil)), ErrAlg},
		{"alg HS512", craft(v, `{"alg":"HS512"}`, payload(v, nil)), ErrAlg},
		{"no alg", craft(v, `{}`, payload(v, nil)), ErrAlg},
		{"tampered payload", parts[0] + "." + b64.EncodeToString([]byte(payload(v, map[string]any{"sub": "mallory"}))) + "." + parts[2], ErrSignature},
		{"empty signature", parts[0] + "." + parts[1] + ".", ErrSignature},
		{"wrong key", func() string {
			o, _ := NewVerifier([]byte(strings.Repeat("z", 32)), "claim-tracker", "claims-api")
			o.Now = v.Now
			s, _ := o.Mint("alice", "", time.Minute)
			return s
		}(), ErrSignature},
		{"expired", craft(v, hs, payload(v, map[string]any{"exp": v.Now().Add(-time.Hour).Unix()})), ErrExpired},
		{"no exp", craft(v, hs, payload(v, map[string]any{"exp": nil})), ErrClaims},
		{"lifetime 1 day", craft(v, hs, payload(v, map[string]any{"exp": v.Now().Add(24 * time.Hour).Unix()})), ErrTooLong},
		{"nbf in future", craft(v, hs, payload(v, map[string]any{"nbf": v.Now().Add(time.Hour).Unix()})), ErrNotYet},
		{"wrong issuer", craft(v, hs, payload(v, map[string]any{"iss": "evil"})), ErrIssuer},
		{"no issuer", craft(v, hs, payload(v, map[string]any{"iss": nil})), ErrIssuer},
		{"wrong audience", craft(v, hs, payload(v, map[string]any{"aud": "other"})), ErrAudience},
		{"audience list without ours", craft(v, hs, payload(v, map[string]any{"aud": []string{"a", "b"}})), ErrAudience},
		{"no sub", craft(v, hs, payload(v, map[string]any{"sub": nil})), ErrClaims},
		{"sub with spaces", craft(v, hs, payload(v, map[string]any{"sub": "a b"})), ErrClaims},
		{"sub too long", craft(v, hs, payload(v, map[string]any{"sub": strings.Repeat("a", 65)})), ErrClaims},
		{"unknown role", craft(v, hs, payload(v, map[string]any{"role": "admin"})), ErrClaims},
		{"payload not json", craft(v, hs, "nope"), ErrMalformed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := v.Verify(c.tok)
			if !errors.Is(err, c.want) {
				t.Fatalf("want %v, got %v", c.want, err)
			}
		})
	}
}

func TestAudienceListContainingOursIsAccepted(t *testing.T) {
	v, _ := newV(t)
	tok := craft(v, hs, payload(v, map[string]any{"aud": []string{"other", "claims-api"}}))
	if _, err := v.Verify(tok); err != nil {
		t.Fatal(err)
	}
}

func TestLeewayAndExpiryBoundary(t *testing.T) {
	v, now := newV(t)
	tok, _ := v.Mint("alice", "", 10*time.Minute)
	*now = now.Add(10*time.Minute + 29*time.Second) // inside 30s leeway
	if _, err := v.Verify(tok); err != nil {
		t.Fatalf("within leeway: %v", err)
	}
	*now = now.Add(2 * time.Second) // past leeway
	if _, err := v.Verify(tok); !errors.Is(err, ErrExpired) {
		t.Fatalf("after leeway: %v", err)
	}
}

func TestNewVerifierRejectsWeakConfig(t *testing.T) {
	if _, err := NewVerifier([]byte("short"), "i", "a"); err == nil {
		t.Fatal("short secret must be refused")
	}
	if _, err := NewVerifier(secret, "", "a"); err == nil {
		t.Fatal("empty issuer must be refused")
	}
	if _, err := NewVerifier(secret, "i", ""); err == nil {
		t.Fatal("empty audience must be refused")
	}
}

func TestBearerToken(t *testing.T) {
	for in, want := range map[string]string{
		"Bearer abc": "abc", "bearer abc": "abc", "BEARER  abc ": "abc",
		"": "", "Bearer": "", "Basic abc": "", "Bearerabc": "", "Bearer ": "",
	} {
		if got := BearerToken(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestReasonsListMatchesWhatReasonReturns(t *testing.T) {
	known := map[string]bool{}
	for _, r := range Reasons {
		known[r] = true
	}
	for _, e := range []error{ErrMissing, ErrMalformed, ErrAlg, ErrSignature, ErrExpired, ErrNotYet, ErrIssuer, ErrAudience, ErrClaims, ErrTooLong} {
		if !known[Reason(e)] {
			t.Errorf("%v -> %q is not in Reasons", e, Reason(e))
		}
	}
	if len(known) != len(Reasons) {
		t.Fatal("duplicate in Reasons")
	}
}

func TestReasonLabelsAreBounded(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range []error{ErrMissing, ErrMalformed, ErrAlg, ErrSignature, ErrExpired, ErrNotYet, ErrIssuer, ErrAudience, ErrClaims, ErrTooLong, errors.New("x")} {
		seen[Reason(e)] = true
	}
	if len(seen) != 10 { // the unknown error shares the "claims" label
		t.Fatalf("got %d distinct labels", len(seen))
	}
}
