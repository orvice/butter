package planestate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("plane install state not found")

type Entry struct {
	State, WorkspaceID, AppID, ClientID, UserID, RedirectURI, ReturnURL string
	CreatedAt, ExpiresAt                                                time.Time
}
type Repository interface {
	Create(context.Context, *Entry) error
	Consume(context.Context, string, time.Time) (*Entry, error)
}
type Memory struct {
	mu      sync.Mutex
	entries map[string]Entry
}

func NewMemory() *Memory   { return &Memory{entries: map[string]Entry{}} }
func Hash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func (s *Memory) Create(_ context.Context, e *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[Hash(e.State)] = *e
	return nil
}
func (s *Memory) Consume(_ context.Context, state string, now time.Time) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := Hash(state)
	e, ok := s.entries[k]
	if !ok || !e.ExpiresAt.After(now) {
		delete(s.entries, k)
		return nil, ErrNotFound
	}
	delete(s.entries, k)
	return &e, nil
}
