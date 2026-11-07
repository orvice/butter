// Package memory implements planestate.Repository in process memory.
package memory

import (
	"context"
	"sync"
	"time"

	"go.orx.me/apps/butter/internal/repo/planestate"
)

// Store implements planestate.Repository.
type Store struct {
	mu      sync.Mutex
	entries map[string]planestate.Entry // hash -> entry without State
}

var _ planestate.Repository = (*Store)(nil)

func New() *Store { return &Store{entries: map[string]planestate.Entry{}} }

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) Create(_ context.Context, entry *planestate.Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := *entry
	stored.State = ""
	s.entries[planestate.Hash(entry.State)] = stored
	return nil
}

func (s *Store) Consume(_ context.Context, state string, now time.Time) (*planestate.Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := planestate.Hash(state)
	entry, ok := s.entries[key]
	if !ok {
		return nil, planestate.ErrNotFound
	}
	delete(s.entries, key)
	if !now.Before(entry.ExpiresAt) {
		return nil, planestate.ErrNotFound
	}
	entry.State = state
	return &entry, nil
}
