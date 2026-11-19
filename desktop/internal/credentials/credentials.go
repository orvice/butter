// Package credentials keeps session tokens in the system's credential
// store, behind an interface small enough to fake in tests and to fall
// back from on a machine without one.
package credentials

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrNotFound reports a key with no stored value.
	ErrNotFound = errors.New("credentials: not found")
	// ErrUnsupported reports a system without a credential store, such as
	// Linux without a Secret Service. The client must not fall back to a
	// plain file: it asks the user to sign in again instead.
	ErrUnsupported = errors.New("credentials: no system credential store")
)

// Store is one secret store.
type Store interface {
	// Get returns the value of key, or ErrNotFound.
	Get(key string) ([]byte, error)
	// Set stores value under key, replacing what was there.
	Set(key string, value []byte) error
	// Delete removes key; deleting an absent key is not an error.
	Delete(key string) error
}

// Memory is an in-process store, for tests and for a machine without a
// credential store where the app runs without persisting anything.
type Memory struct {
	mu     sync.RWMutex
	values map[string][]byte
}

// NewMemory returns an empty Memory store.
func NewMemory() *Memory { return &Memory{values: map[string][]byte{}} }

// Get implements Store.
func (m *Memory) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.values[key]
	if !ok {
		return nil, ErrNotFound
	}
	return append([]byte(nil), v...), nil
}

// Set implements Store.
func (m *Memory) Set(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[key] = append([]byte(nil), value...)
	return nil
}

// Delete implements Store.
func (m *Memory) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.values, key)
	return nil
}

// Keys returns the stored keys, sorted by nothing in particular; it keeps
// tests that assert what was cleaned up readable.
func (m *Memory) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.values))
	for k := range m.values {
		out = append(out, k)
	}
	return out
}

// Unsupported is a store for a system without one; every call fails with
// ErrUnsupported so the caller can tell the user instead of writing a token
// to disk.
type Unsupported struct{}

// Get implements Store.
func (Unsupported) Get(string) ([]byte, error) { return nil, ErrUnsupported }

// Set implements Store.
func (Unsupported) Set(string, []byte) error { return ErrUnsupported }

// Delete implements Store.
func (Unsupported) Delete(string) error { return ErrUnsupported }

// ReadToken is the Get-and-classify step every caller of a token store
// needs: ErrNotFound and ErrUnsupported stay distinguishable.
func ReadToken(s Store, key string) (string, error) {
	v, err := s.Get(key)
	if err != nil {
		return "", fmt.Errorf("credentials: read %q: %w", key, err)
	}
	return string(v), nil
}
