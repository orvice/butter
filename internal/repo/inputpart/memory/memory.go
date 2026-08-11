package memory

import (
	"context"
	"sync"

	"go.orx.me/apps/butter/internal/repo/inputpart"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

// Store is a thread-safe in-memory implementation of inputpart.Repository.
type Store struct {
	mu   sync.RWMutex
	data map[string][]*agentsv1.InputPart // invocationID → ordered parts
}

func New() *Store {
	return &Store{data: make(map[string][]*agentsv1.InputPart)}
}

func (s *Store) SaveAll(_ context.Context, invocationID string, parts []*agentsv1.InputPart) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[invocationID]; exists {
		return nil // idempotent
	}
	cloned := make([]*agentsv1.InputPart, len(parts))
	for i, p := range parts {
		cloned[i] = proto.Clone(p).(*agentsv1.InputPart)
	}
	s.data[invocationID] = cloned
	return nil
}

func (s *Store) Load(_ context.Context, invocationID string) ([]*agentsv1.InputPart, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	parts, ok := s.data[invocationID]
	if !ok || len(parts) == 0 {
		return nil, inputpart.ErrNotFound
	}
	out := make([]*agentsv1.InputPart, len(parts))
	for i, p := range parts {
		out[i] = proto.Clone(p).(*agentsv1.InputPart)
	}
	return out, nil
}

func (s *Store) Delete(_ context.Context, invocationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, invocationID)
	return nil
}
