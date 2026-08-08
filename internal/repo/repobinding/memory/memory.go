// Package memory implements repobinding.Repository in process memory for
// local development and tests.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type entry struct {
	binding           *agentsv1.WorkspaceRepoBinding
	credential        string
	credentialUpdated *timestamppb.Timestamp
}

// Store implements repobinding.Repository.
type Store struct {
	mu       sync.RWMutex
	bindings map[string]*entry // workspaceID -> entry
}

var _ repobindingrepo.Repository = (*Store)(nil)

func New() *Store {
	return &Store{bindings: map[string]*entry{}}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

// view clones the stored binding and overlays the repo-owned credential
// fields so callers always see the derived truth.
func (e *entry) view() *agentsv1.WorkspaceRepoBinding {
	b := proto.Clone(e.binding).(*agentsv1.WorkspaceRepoBinding)
	b.CredentialSet = e.credential != ""
	b.CredentialUpdatedAt = e.credentialUpdated
	return b
}

func (s *Store) Get(_ context.Context, workspaceID string) (*agentsv1.WorkspaceRepoBinding, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.bindings[workspaceID]
	if !ok {
		return nil, fmt.Errorf("repo binding (workspace %q): %w", workspaceID, repobindingrepo.ErrNotFound)
	}
	return e.view(), nil
}

func (s *Store) Put(_ context.Context, workspaceID string, binding *agentsv1.WorkspaceRepoBinding) (*agentsv1.WorkspaceRepoBinding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := proto.Clone(binding).(*agentsv1.WorkspaceRepoBinding)
	stored.WorkspaceId = workspaceID
	stored.CredentialSet = false
	stored.CredentialUpdatedAt = nil
	now := timestamppb.New(time.Now().UTC())
	stored.UpdatedAt = now
	prev, ok := s.bindings[workspaceID]
	if ok {
		stored.CreatedAt = prev.binding.GetCreatedAt()
		prev.binding = stored
		return prev.view(), nil
	}
	stored.CreatedAt = now
	e := &entry{binding: stored}
	s.bindings[workspaceID] = e
	return e.view(), nil
}

func (s *Store) Delete(_ context.Context, workspaceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bindings[workspaceID]; !ok {
		return fmt.Errorf("repo binding (workspace %q): %w", workspaceID, repobindingrepo.ErrNotFound)
	}
	delete(s.bindings, workspaceID)
	return nil
}

func (s *Store) SetCredential(_ context.Context, workspaceID, ciphertext string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.bindings[workspaceID]
	if !ok {
		return fmt.Errorf("repo binding (workspace %q): %w", workspaceID, repobindingrepo.ErrNotFound)
	}
	e.credential = ciphertext
	e.credentialUpdated = timestamppb.New(time.Now().UTC())
	return nil
}

func (s *Store) GetCredential(_ context.Context, workspaceID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.bindings[workspaceID]
	if !ok {
		return "", fmt.Errorf("repo binding (workspace %q): %w", workspaceID, repobindingrepo.ErrNotFound)
	}
	if e.credential == "" {
		return "", fmt.Errorf("repo binding (workspace %q): %w", workspaceID, repobindingrepo.ErrNoCredential)
	}
	return e.credential, nil
}

func (s *Store) ListAcrossWorkspaces(context.Context) ([]*agentsv1.WorkspaceRepoBinding, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.WorkspaceRepoBinding, 0, len(s.bindings))
	for _, e := range s.bindings {
		out = append(out, e.view())
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetWorkspaceId() < out[j].GetWorkspaceId() })
	return out, nil
}
