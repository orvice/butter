// Package memory implements butterbox.Repository in process memory for local
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

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type record struct {
	box               *agentsv1.ButterBox
	credential        butterboxrepo.Credential
	credentialUpdated time.Time
}

// Store implements butterbox.Repository.
type Store struct {
	mu    sync.RWMutex
	boxes map[string]map[string]*record // workspaceID -> id -> record
}

var _ butterboxrepo.Repository = (*Store)(nil)

func New() *Store {
	return &Store{boxes: map[string]map[string]*record{}}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

// materialize clones the stored box and stamps the derived fields, mirroring
// what the mongo decoder does on every read.
func (r *record) materialize(workspaceID string) *agentsv1.ButterBox {
	box := proto.Clone(r.box).(*agentsv1.ButterBox)
	box.WorkspaceId = workspaceID
	box.CredentialSet = r.credential.Set()
	box.CredentialUpdatedAt = nil
	if r.credential.Set() && !r.credentialUpdated.IsZero() {
		box.CredentialUpdatedAt = timestamppb.New(r.credentialUpdated)
	}
	return box
}

// sanitize strips derived fields before storage so they can never contradict
// the credential record.
func sanitize(box *agentsv1.ButterBox) *agentsv1.ButterBox {
	stored := proto.Clone(box).(*agentsv1.ButterBox)
	stored.WorkspaceId = ""
	stored.CredentialSet = false
	stored.CredentialUpdatedAt = nil
	return stored
}

func (s *Store) List(_ context.Context, workspaceID string) ([]*agentsv1.ButterBox, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.ButterBox, 0, len(s.boxes[workspaceID]))
	for _, r := range s.boxes[workspaceID] {
		out = append(out, r.materialize(workspaceID))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetName() < out[j].GetName() })
	return out, nil
}

func (s *Store) Get(_ context.Context, workspaceID, id string) (*agentsv1.ButterBox, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.boxes[workspaceID][id]
	if !ok {
		return nil, fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNotFound)
	}
	return r.materialize(workspaceID), nil
}

func (s *Store) Create(_ context.Context, workspaceID string, box *agentsv1.ButterBox, cred butterboxrepo.Credential) (*agentsv1.ButterBox, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := box.GetId()
	if _, ok := s.boxes[workspaceID][id]; ok {
		return nil, fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrAlreadyExists)
	}
	for _, r := range s.boxes[workspaceID] {
		if r.box.GetName() == box.GetName() {
			return nil, fmt.Errorf("butterbox name %q: %w", box.GetName(), butterboxrepo.ErrAlreadyExists)
		}
	}
	stored := sanitize(box)
	now := timestamppb.New(time.Now().UTC())
	stored.CreatedAt = now
	stored.UpdatedAt = now
	r := &record{box: stored}
	if cred.Set() {
		r.credential = cred
		r.credentialUpdated = time.Now().UTC()
	}
	if s.boxes[workspaceID] == nil {
		s.boxes[workspaceID] = map[string]*record{}
	}
	s.boxes[workspaceID][id] = r
	return r.materialize(workspaceID), nil
}

func (s *Store) Update(_ context.Context, workspaceID string, box *agentsv1.ButterBox) (*agentsv1.ButterBox, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := box.GetId()
	prev, ok := s.boxes[workspaceID][id]
	if !ok {
		return nil, fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNotFound)
	}
	for otherID, r := range s.boxes[workspaceID] {
		if otherID != id && r.box.GetName() == box.GetName() {
			return nil, fmt.Errorf("butterbox name %q: %w", box.GetName(), butterboxrepo.ErrAlreadyExists)
		}
	}
	stored := sanitize(box)
	stored.CreatedAt = prev.box.GetCreatedAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	prev.box = stored
	return prev.materialize(workspaceID), nil
}

func (s *Store) Delete(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.boxes[workspaceID][id]; !ok {
		return fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNotFound)
	}
	delete(s.boxes[workspaceID], id)
	return nil
}

func (s *Store) SetCredential(_ context.Context, workspaceID, id string, cred butterboxrepo.Credential) (*agentsv1.ButterBox, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.boxes[workspaceID][id]
	if !ok {
		return nil, fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNotFound)
	}
	if cred.Set() {
		r.credential = cred
		r.credentialUpdated = time.Now().UTC()
	} else {
		r.credential = butterboxrepo.Credential{}
		r.credentialUpdated = time.Time{}
	}
	return r.materialize(workspaceID), nil
}

func (s *Store) GetCredential(_ context.Context, workspaceID, id string) (butterboxrepo.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.boxes[workspaceID][id]
	if !ok {
		return butterboxrepo.Credential{}, fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNotFound)
	}
	if !r.credential.Set() {
		return butterboxrepo.Credential{}, fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNoCredential)
	}
	return r.credential, nil
}
