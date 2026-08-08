// Package memory implements githost.Repository in process memory for local
// development and tests.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	githostrepo "go.orx.me/apps/butter/internal/repo/githost"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Store implements githost.Repository.
type Store struct {
	mu    sync.RWMutex
	hosts map[string]*agentsv1.GitHost
}

var _ githostrepo.Repository = (*Store)(nil)

func New() *Store {
	return &Store{hosts: map[string]*agentsv1.GitHost{}}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func clone(h *agentsv1.GitHost) *agentsv1.GitHost {
	return proto.Clone(h).(*agentsv1.GitHost)
}

func (s *Store) List(context.Context) ([]*agentsv1.GitHost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.GitHost, 0, len(s.hosts))
	for _, h := range s.hosts {
		out = append(out, clone(h))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetName() < out[j].GetName() })
	return out, nil
}

func (s *Store) Get(_ context.Context, id string) (*agentsv1.GitHost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.hosts[id]
	if !ok {
		return nil, fmt.Errorf("git host %q: %w", id, githostrepo.ErrNotFound)
	}
	return clone(h), nil
}

func (s *Store) Create(_ context.Context, host *agentsv1.GitHost) (*agentsv1.GitHost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := host.GetId()
	if _, ok := s.hosts[id]; ok {
		return nil, fmt.Errorf("git host %q: %w", id, githostrepo.ErrAlreadyExists)
	}
	stored := clone(host)
	now := timestamppb.New(time.Now().UTC())
	stored.CreatedAt = now
	stored.UpdatedAt = now
	s.hosts[id] = stored
	return clone(stored), nil
}

func (s *Store) Update(_ context.Context, host *agentsv1.GitHost) (*agentsv1.GitHost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := host.GetId()
	prev, ok := s.hosts[id]
	if !ok {
		return nil, fmt.Errorf("git host %q: %w", id, githostrepo.ErrNotFound)
	}
	stored := clone(host)
	stored.CreatedAt = prev.GetCreatedAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	s.hosts[id] = stored
	return clone(stored), nil
}

func (s *Store) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.hosts[id]; !ok {
		return fmt.Errorf("git host %q: %w", id, githostrepo.ErrNotFound)
	}
	delete(s.hosts, id)
	return nil
}
