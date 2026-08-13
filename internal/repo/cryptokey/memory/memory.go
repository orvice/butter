// Package memory implements cryptokey.Repository in process memory for tests
// and single-process development. EnsureActive holds the lock across the
// generate call so concurrent goroutines converge on one key, matching the
// Mongo store's cross-Pod guarantee.
package memory

import (
	"context"
	"fmt"
	"sync"

	"go.orx.me/apps/butter/internal/repo/cryptokey"
)

// Store implements cryptokey.Repository in memory.
type Store struct {
	mu     sync.Mutex
	keys   map[string]cryptokey.MasterKey
	active string
}

var _ cryptokey.Repository = (*Store)(nil)

func New() *Store {
	return &Store{keys: make(map[string]cryptokey.MasterKey)}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) EnsureActive(_ context.Context, generate func() (cryptokey.MasterKey, error)) (cryptokey.MasterKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != "" {
		return s.keys[s.active], nil
	}
	key, err := generate()
	if err != nil {
		return cryptokey.MasterKey{}, fmt.Errorf("generate master key: %w", err)
	}
	s.keys[key.ID] = key
	s.active = key.ID
	return key, nil
}

func (s *Store) Get(_ context.Context, id string) (cryptokey.MasterKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key, ok := s.keys[id]
	if !ok {
		return cryptokey.MasterKey{}, fmt.Errorf("master key %q: %w", id, cryptokey.ErrNotFound)
	}
	return key, nil
}
