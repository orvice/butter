// Package memory implements memoryconfig.Repository in process memory for
// local development and tests.
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	memoryconfigrepo "go.orx.me/apps/butter/internal/repo/memoryconfig"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type record struct {
	baseURL           string
	enabled           bool
	createdAt         time.Time
	updatedAt         time.Time
	credential        memoryconfigrepo.Credential
	credentialUpdated time.Time
}

// Store implements memoryconfig.Repository.
type Store struct {
	mu      sync.RWMutex
	configs map[string]*record // workspaceID -> record
}

var _ memoryconfigrepo.Repository = (*Store)(nil)

func New() *Store {
	return &Store{configs: map[string]*record{}}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

// materialize stamps the derived fields, mirroring the mongo decoder.
func (r *record) materialize(workspaceID string) *agentsv1.WorkspaceMemoryConfig {
	cfg := &agentsv1.WorkspaceMemoryConfig{
		WorkspaceId:   workspaceID,
		BaseUrl:       r.baseURL,
		Enabled:       r.enabled,
		CredentialSet: r.credential.Set(),
		CreatedAt:     timestamppb.New(r.createdAt),
		UpdatedAt:     timestamppb.New(r.updatedAt),
	}
	if r.credential.Set() && !r.credentialUpdated.IsZero() {
		cfg.CredentialUpdatedAt = timestamppb.New(r.credentialUpdated)
	}
	return cfg
}

func (s *Store) Get(_ context.Context, workspaceID string) (*agentsv1.WorkspaceMemoryConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.configs[workspaceID]
	if !ok {
		return nil, fmt.Errorf("workspace memory config %q: %w", workspaceID, memoryconfigrepo.ErrNotFound)
	}
	return r.materialize(workspaceID), nil
}

func (s *Store) Put(_ context.Context, workspaceID string, cfg *agentsv1.WorkspaceMemoryConfig, cred *memoryconfigrepo.Credential) (*agentsv1.WorkspaceMemoryConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	r, ok := s.configs[workspaceID]
	if !ok {
		r = &record{createdAt: now}
		s.configs[workspaceID] = r
	}
	r.baseURL = cfg.GetBaseUrl()
	r.enabled = cfg.GetEnabled()
	r.updatedAt = now
	switch {
	case cred == nil:
	case cred.Set():
		r.credential = *cred
		r.credentialUpdated = now
	default:
		r.credential = memoryconfigrepo.Credential{}
		r.credentialUpdated = time.Time{}
	}
	return r.materialize(workspaceID), nil
}

func (s *Store) Delete(_ context.Context, workspaceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.configs[workspaceID]; !ok {
		return fmt.Errorf("workspace memory config %q: %w", workspaceID, memoryconfigrepo.ErrNotFound)
	}
	delete(s.configs, workspaceID)
	return nil
}

func (s *Store) GetCredential(_ context.Context, workspaceID string) (memoryconfigrepo.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.configs[workspaceID]
	if !ok {
		return memoryconfigrepo.Credential{}, fmt.Errorf("workspace memory config %q: %w", workspaceID, memoryconfigrepo.ErrNotFound)
	}
	if !r.credential.Set() {
		return memoryconfigrepo.Credential{}, fmt.Errorf("workspace memory config %q: %w", workspaceID, memoryconfigrepo.ErrNoCredential)
	}
	return r.credential, nil
}
