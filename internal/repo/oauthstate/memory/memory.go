// Package memory is an in-process OAuth state store. It is intended for
// tests and local dev; production should use the mongo store so state
// survives process restarts.
package memory

import (
	"context"
	"sync"
	"time"

	"go.orx.me/apps/butter/internal/repo/oauthstate"
)

type Store struct {
	mu      sync.Mutex
	entries map[string]oauthstate.Entry
}

var _ oauthstate.Repository = (*Store)(nil)

func New() *Store {
	return &Store{entries: make(map[string]oauthstate.Entry)}
}

func (s *Store) EnsureIndexes(_ context.Context) error { return nil }

func (s *Store) Create(_ context.Context, entry *oauthstate.Entry) error {
	if entry == nil || entry.State == "" {
		return oauthstate.ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[entry.State] = *entry
	return nil
}

func (s *Store) Consume(_ context.Context, state string, now time.Time) (*oauthstate.Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[state]
	if !ok {
		return nil, oauthstate.ErrNotFound
	}
	delete(s.entries, state)
	if !entry.ExpiresAt.After(now) {
		return nil, oauthstate.ErrNotFound
	}
	return &entry, nil
}
