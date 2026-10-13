// Package memory is an in-process OAuth handoff store for tests and local
// development; production uses the mongo store so a handoff survives a
// redeploy and is visible to every Pod.
package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.orx.me/apps/butter/internal/repo/oauthhandoff"
)

type Store struct {
	mu      sync.Mutex
	entries map[string]oauthhandoff.Entry
}

var _ oauthhandoff.Repository = (*Store)(nil)

func New() *Store {
	return &Store{entries: make(map[string]oauthhandoff.Entry)}
}

func (s *Store) EnsureIndexes(_ context.Context) error { return nil }

func (s *Store) Create(_ context.Context, entry *oauthhandoff.Entry) error {
	if entry == nil || entry.CodeHash == "" {
		return errors.New("oauth handoff entry invalid")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[entry.CodeHash]; exists {
		return errors.New("oauth handoff already exists")
	}
	s.entries[entry.CodeHash] = *entry
	return nil
}

func (s *Store) Consume(_ context.Context, codeHash string, now time.Time) (*oauthhandoff.Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[codeHash]
	if !ok {
		return nil, oauthhandoff.ErrNotFound
	}
	delete(s.entries, codeHash)
	if !entry.ExpiresAt.After(now) {
		return nil, oauthhandoff.ErrNotFound
	}
	return &entry, nil
}
