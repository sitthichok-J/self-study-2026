package auth

import (
	"context"
	"sync"
	"time"
)

// MemStore is an in-memory Store for tests and the baseline snapshot.
type MemStore struct {
	mu      sync.Mutex
	users   map[string]User
	tokens  map[string]*memToken
	FailGet error // set to make GetUser fail (tests)
}

type memToken struct {
	rec     RefreshRecord
	used    bool
	revoked bool
}

func NewMemStore() *MemStore {
	return &MemStore{users: map[string]User{}, tokens: map[string]*memToken{}}
}

func (m *MemStore) GetUser(_ context.Context, username string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FailGet != nil {
		return User{}, m.FailGet
	}
	u, ok := m.users[username]
	if !ok {
		return User{}, ErrNoUser
	}
	return u, nil
}

func (m *MemStore) UpsertUser(_ context.Context, u User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.Username] = u
	return nil
}

// DeleteUser removes a user (tests: account removed while a session is alive).
func (m *MemStore) DeleteUser(username string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, username)
}

func (m *MemStore) SaveRefresh(_ context.Context, r RefreshRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[r.Hash] = &memToken{rec: r}
	return nil
}

func (m *MemStore) UseRefresh(_ context.Context, hash string, now time.Time) (RefreshRecord, UseStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[hash]
	switch {
	case !ok:
		return RefreshRecord{}, UseUnknown, nil
	case t.used:
		return t.rec, UseReused, nil
	case t.revoked:
		return t.rec, UseRevoked, nil
	case !now.Before(t.rec.Expires):
		return t.rec, UseExpired, nil
	}
	t.used = true
	return t.rec, UseOK, nil
}

func (m *MemStore) FamilyOf(_ context.Context, hash string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[hash]
	if !ok {
		return "", false, nil
	}
	return t.rec.Family, true, nil
}

func (m *MemStore) RevokeFamily(_ context.Context, family string, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tokens {
		if t.rec.Family == family {
			t.revoked = true
		}
	}
	return nil
}

// Live counts tokens in a family that could still be used (tests).
func (m *MemStore) Live(family string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, t := range m.tokens {
		if t.rec.Family == family && !t.used && !t.revoked {
			n++
		}
	}
	return n
}
