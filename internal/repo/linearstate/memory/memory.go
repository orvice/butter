// Package memory implements linearstate.Repository in process memory.
package memory

import (
	"context"
	"sync"
	"time"

	"go.orx.me/apps/butter/internal/repo/linearstate"
)

// Store implements linearstate.Repository.
type Store struct {
	mu      sync.Mutex
	entries map[string]linearstate.Entry // hash -> entry without State
}

var _ linearstate.Repository = (*Store)(nil)

func New() *Store { return &Store{entries: map[string]linearstate.Entry{}} }

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) Create(_ context.Context, entry *linearstate.Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := *entry
	stored.State = ""
	s.entries[linearstate.Hash(entry.State)] = stored
	return nil
}

func (s *Store) Consume(_ context.Context, state string, now time.Time) (*linearstate.Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := linearstate.Hash(state)
	entry, ok := s.entries[key]
	if !ok {
		return nil, linearstate.ErrNotFound
	}
	delete(s.entries, key)
	if !now.Before(entry.ExpiresAt) {
		return nil, linearstate.ErrNotFound
	}
	entry.State = state
	return &entry, nil
}
