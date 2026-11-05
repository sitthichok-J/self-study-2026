package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Login and refresh sit on top of the Day 10 verifier.
//
//   - access token: short-lived HS256 JWT (the same one every /claims route checks)
//   - refresh token: 32 random bytes, only its SHA-256 is stored
//   - refresh tokens rotate: each one works once. Using one a second time means
//     it was copied, so the whole family (every token from that login) is revoked.

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	// ErrRefreshReused wraps ErrInvalidRefresh: callers may log it as a security event.
	ErrRefreshReused = fmt.Errorf("%w: reuse detected", ErrInvalidRefresh)
	ErrNoUser        = errors.New("no such user")
)

type User struct {
	Username     string
	PasswordHash string
	Role         string
}

type RefreshRecord struct {
	Hash     string
	Family   string
	Username string
	Expires  time.Time
}

// UseStatus is what Store.UseRefresh found.
type UseStatus int

const (
	UseOK      UseStatus = iota // was valid; now marked used
	UseUnknown                  // no such token
	UseReused                   // already used once
	UseRevoked                  // family revoked (logout or earlier reuse)
	UseExpired
)

// Store is the persistence the session service needs.
type Store interface {
	GetUser(ctx context.Context, username string) (User, error) // ErrNoUser if missing
	UpsertUser(ctx context.Context, u User) error
	SaveRefresh(ctx context.Context, r RefreshRecord) error
	// UseRefresh must be atomic: of two concurrent calls with the same hash,
	// exactly one may get UseOK. Check order: unknown, reused, revoked, expired.
	UseRefresh(ctx context.Context, hash string, now time.Time) (RefreshRecord, UseStatus, error)
	// FamilyOf finds the family for logout; ok=false if the hash is unknown.
	FamilyOf(ctx context.Context, hash string) (family string, ok bool, err error)
	RevokeFamily(ctx context.Context, family string, now time.Time) error
}

// PasswordChecker is what the service needs from a Hasher (an interface so a test can count calls).
type PasswordChecker interface {
	Verify(password, stored string) bool
}

type Tokens struct {
	Subject        string
	Access         string
	ExpiresIn      int // seconds until the access token expires
	Refresh        string
	RefreshExpires time.Time
	RefreshMaxAge  int // seconds the refresh token lives (cookie Max-Age): the TTL, not "time until now+TTL"
}

type Service struct {
	store      Store
	ver        *Verifier
	checker    PasswordChecker
	dummy      string // hash checked when the user does not exist, so timing does not reveal it
	accessTTL  time.Duration
	refreshTTL time.Duration
	Now        func() time.Time
	revoker    Revoker // optional: set with SetRevoker
}

// SetRevoker lets the service cut off a user's access tokens when it detects a stolen refresh token.
func (s *Service) SetRevoker(r Revoker) { s.revoker = r }

func NewService(store Store, ver *Verifier, h Hasher, accessTTL, refreshTTL time.Duration) (*Service, error) {
	return newService(store, ver, h, h, accessTTL, refreshTTL)
}

func newService(store Store, ver *Verifier, h Hasher, checker PasswordChecker, accessTTL, refreshTTL time.Duration) (*Service, error) {
	if store == nil || ver == nil {
		return nil, errors.New("store and verifier are required")
	}
	if accessTTL <= 0 || accessTTL > ver.MaxTTL {
		return nil, fmt.Errorf("access TTL must be within (0, %s]", ver.MaxTTL)
	}
	if refreshTTL <= accessTTL {
		return nil, errors.New("refresh TTL must be longer than access TTL")
	}
	filler, err := randomToken(16)
	if err != nil {
		return nil, err
	}
	dummy, err := h.Hash("dummy-" + filler)
	if err != nil {
		return nil, err
	}
	return &Service{store: store, ver: ver, checker: checker, dummy: dummy, accessTTL: accessTTL, refreshTTL: refreshTTL, Now: time.Now}, nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashRefresh is the value stored in place of the raw token.
func HashRefresh(raw string) string {
	s := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(s[:])
}

// plausible rejects obviously wrong input before it reaches the store.
func plausible(raw string) bool { return len(raw) >= 20 && len(raw) <= 128 }

func (s *Service) issue(ctx context.Context, u User, family string) (Tokens, error) {
	now := s.Now()
	access, err := s.ver.MintAt(u.Username, u.Role, s.accessTTL, now)
	if err != nil {
		return Tokens{}, err
	}
	raw, err := randomToken(32)
	if err != nil {
		return Tokens{}, err
	}
	exp := now.Add(s.refreshTTL)
	if err := s.store.SaveRefresh(ctx, RefreshRecord{Hash: HashRefresh(raw), Family: family, Username: u.Username, Expires: exp}); err != nil {
		return Tokens{}, err
	}
	return Tokens{Subject: u.Username, Access: access, ExpiresIn: int(s.accessTTL.Seconds()), Refresh: raw, RefreshExpires: exp, RefreshMaxAge: int(s.refreshTTL.Seconds())}, nil
}

// Login checks the password and starts a new token family. Every failure is
// the same ErrInvalidCredentials, and an unknown user costs one hash check
// like a wrong password does.
func (s *Service) Login(ctx context.Context, username, password string) (Tokens, error) {
	if !subRe.MatchString(username) || len(password) > maxPasswordLen {
		s.checker.Verify(password, s.dummy)
		return Tokens{}, ErrInvalidCredentials
	}
	u, err := s.store.GetUser(ctx, username)
	if errors.Is(err, ErrNoUser) {
		s.checker.Verify(password, s.dummy)
		return Tokens{}, ErrInvalidCredentials
	}
	if err != nil {
		return Tokens{}, err
	}
	if !s.checker.Verify(password, u.PasswordHash) {
		return Tokens{}, ErrInvalidCredentials
	}
	family, err := randomToken(16)
	if err != nil {
		return Tokens{}, err
	}
	return s.issue(ctx, u, family)
}

// Refresh trades a refresh token for a new pair in the same family.
func (s *Service) Refresh(ctx context.Context, raw string) (Tokens, error) {
	if !plausible(raw) {
		return Tokens{}, ErrInvalidRefresh
	}
	rec, st, err := s.store.UseRefresh(ctx, HashRefresh(raw), s.Now())
	if err != nil {
		return Tokens{}, err
	}
	switch st {
	case UseOK:
	case UseReused:
		if err := s.store.RevokeFamily(ctx, rec.Family, s.Now()); err != nil {
			return Tokens{}, err
		}
		// The copy of the token is in someone's hands: access tokens already issued to this user die now too.
		if s.revoker != nil {
			if err := s.revoker.RevokeUser(ctx, rec.Username, s.Now(), s.ver.MaxTokenLife()); err != nil {
				return Tokens{}, err
			}
		}
		return Tokens{}, ErrRefreshReused
	default:
		return Tokens{}, ErrInvalidRefresh
	}
	u, err := s.store.GetUser(ctx, rec.Username)
	if errors.Is(err, ErrNoUser) { // account removed: nothing from this login may live on
		if rerr := s.store.RevokeFamily(ctx, rec.Family, s.Now()); rerr != nil {
			return Tokens{}, rerr
		}
		return Tokens{}, ErrInvalidRefresh
	}
	if err != nil {
		return Tokens{}, err
	}
	return s.issue(ctx, u, rec.Family)
}

// Logout revokes the whole family of the given refresh token. An unknown or
// malformed token is not an error: logging out twice must be harmless.
func (s *Service) Logout(ctx context.Context, raw string) error {
	if !plausible(raw) {
		return nil
	}
	family, ok, err := s.store.FamilyOf(ctx, HashRefresh(raw))
	if err != nil || !ok {
		return err
	}
	return s.store.RevokeFamily(ctx, family, s.Now())
}

// SetUser creates or replaces a user (seeding and tests).
func (s *Service) SetUser(ctx context.Context, h Hasher, username, password, role string) error {
	if !subRe.MatchString(username) {
		return errors.New("invalid username")
	}
	if role != RoleUser && role != RoleStaff {
		return errors.New("invalid role")
	}
	hash, err := h.Hash(password)
	if err != nil {
		return err
	}
	return s.store.UpsertUser(ctx, User{Username: username, PasswordHash: hash, Role: role})
}
