package memory

import (
	"context"
	"sync"
	"time"

	"go.orx.me/apps/butter/internal/repo/apitoken"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Store is an in-memory implementation of apitoken.Repository, primarily used
// for tests and when no MongoDB backend is configured.
type Store struct {
	mu      sync.RWMutex
	byID    map[string]*record
	byHash  map[string]string // secret hash → id
}

type record struct {
	token      *agentsv1.APIToken
	secretHash string
}

func New() *Store {
	return &Store{
		byID:   make(map[string]*record),
		byHash: make(map[string]string),
	}
}

func (s *Store) List(_ context.Context, workspaceID string) ([]*agentsv1.APIToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.APIToken, 0, len(s.byID))
	for _, r := range s.byID {
		if workspaceID != "" && r.token.GetWorkspaceId() != workspaceID {
			continue
		}
		out = append(out, proto.Clone(r.token).(*agentsv1.APIToken))
	}
	return out, nil
}

func (s *Store) Get(_ context.Context, id string) (*agentsv1.APIToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.byID[id]
	if !ok {
		return nil, apitoken.ErrNotFound
	}
	return proto.Clone(r.token).(*agentsv1.APIToken), nil
}

func (s *Store) Create(_ context.Context, token *agentsv1.APIToken, secretHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[token.GetId()] = &record{token: proto.Clone(token).(*agentsv1.APIToken), secretHash: secretHash}
	s.byHash[secretHash] = token.GetId()
	return nil
}

func (s *Store) Revoke(_ context.Context, id string) (*agentsv1.APIToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.byID[id]
	if !ok {
		return nil, apitoken.ErrNotFound
	}
	r.token.Revoked = true
	delete(s.byHash, r.secretHash)
	return proto.Clone(r.token).(*agentsv1.APIToken), nil
}

func (s *Store) Lookup(_ context.Context, secretHash string) (*agentsv1.APIToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byHash[secretHash]
	if !ok {
		return nil, apitoken.ErrNotFound
	}
	r := s.byID[id]
	if r == nil || r.token.GetRevoked() {
		return nil, apitoken.ErrNotFound
	}
	return proto.Clone(r.token).(*agentsv1.APIToken), nil
}

func (s *Store) TouchLastUsed(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.byID[id]
	if !ok {
		return apitoken.ErrNotFound
	}
	r.token.LastUsedAt = timestamppb.New(time.Now().UTC())
	return nil
}
