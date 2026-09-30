// Package memory implements linear.Repository in process memory for local
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

	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type appRecord struct {
	workspaceID string
	app         *agentsv1.LinearApp
	creds       linearrepo.AppCredentials
}

// Store implements linear.Repository.
type Store struct {
	mu   sync.RWMutex
	apps map[string]*appRecord // id -> record
}

var _ linearrepo.Repository = (*Store)(nil)

func New() *Store {
	return &Store{apps: map[string]*appRecord{}}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (r *appRecord) materialize() *agentsv1.LinearApp {
	app := proto.Clone(r.app).(*agentsv1.LinearApp)
	app.WorkspaceId = r.workspaceID
	linearrepo.StampCredentialState(app, r.creds)
	return app
}

func (s *Store) lookup(workspaceID, id string) (*appRecord, error) {
	r, ok := s.apps[id]
	if !ok || r.workspaceID != workspaceID {
		return nil, fmt.Errorf("linear app %q: %w", id, linearrepo.ErrNotFound)
	}
	return r, nil
}

func (s *Store) ListApps(_ context.Context, workspaceID string) ([]*agentsv1.LinearApp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*agentsv1.LinearApp{}
	for _, r := range s.apps {
		if r.workspaceID == workspaceID {
			out = append(out, r.materialize())
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].GetDisplayName() != out[j].GetDisplayName() {
			return out[i].GetDisplayName() < out[j].GetDisplayName()
		}
		return out[i].GetId() < out[j].GetId()
	})
	return out, nil
}

func (s *Store) GetApp(_ context.Context, workspaceID, id string) (*agentsv1.LinearApp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, err := s.lookup(workspaceID, id)
	if err != nil {
		return nil, err
	}
	return r.materialize(), nil
}

func (s *Store) FindApp(_ context.Context, id string) (*agentsv1.LinearApp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.apps[id]
	if !ok {
		return nil, fmt.Errorf("linear app %q: %w", id, linearrepo.ErrNotFound)
	}
	return r.materialize(), nil
}

func (s *Store) CreateApp(_ context.Context, workspaceID string, app *agentsv1.LinearApp, creds linearrepo.AppCredentials) (*agentsv1.LinearApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.apps[app.GetId()]; ok {
		return nil, fmt.Errorf("linear app %q already exists", app.GetId())
	}
	for _, r := range s.apps {
		if r.app.GetClientId() == app.GetClientId() {
			return nil, fmt.Errorf("linear client id %q: %w", app.GetClientId(), linearrepo.ErrClientIDExists)
		}
	}
	stored := proto.Clone(app).(*agentsv1.LinearApp)
	linearrepo.StripDerived(stored)
	now := timestamppb.New(time.Now().UTC())
	stored.CreatedAt = now
	stored.UpdatedAt = now
	stored.Revision = 1
	r := &appRecord{workspaceID: workspaceID, app: stored, creds: creds}
	s.apps[stored.GetId()] = r
	return r.materialize(), nil
}

func (s *Store) UpdateApp(_ context.Context, workspaceID string, app *agentsv1.LinearApp, expectedRevision int64) (*agentsv1.LinearApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.lookup(workspaceID, app.GetId())
	if err != nil {
		return nil, err
	}
	if r.app.GetRevision() != expectedRevision {
		return nil, fmt.Errorf("linear app %q: %w", app.GetId(), linearrepo.ErrRevisionConflict)
	}
	stored := proto.Clone(app).(*agentsv1.LinearApp)
	linearrepo.StripDerived(stored)
	stored.ClientId = r.app.GetClientId()
	stored.CreatedAt = r.app.GetCreatedAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	stored.Revision = r.app.GetRevision() + 1
	r.app = stored
	return r.materialize(), nil
}

func (s *Store) SetAppCredentials(_ context.Context, workspaceID, id string, change linearrepo.CredentialChange) (*agentsv1.LinearApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.lookup(workspaceID, id)
	if err != nil {
		return nil, err
	}
	if change.ClientSecret != nil {
		r.creds.ClientSecret = *change.ClientSecret
	}
	if change.WebhookSecret != nil {
		r.creds.WebhookSecret = *change.WebhookSecret
	}
	return r.materialize(), nil
}

func (s *Store) GetAppCredentials(_ context.Context, workspaceID, id string) (linearrepo.AppCredentials, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, err := s.lookup(workspaceID, id)
	if err != nil {
		return linearrepo.AppCredentials{}, err
	}
	return r.creds, nil
}

func (s *Store) DeleteApp(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.lookup(workspaceID, id); err != nil {
		return err
	}
	delete(s.apps, id)
	return nil
}
