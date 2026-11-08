package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type record struct {
	app   *agentsv1.PlaneApp
	creds planerepo.AppCredentials
}

type clientKey struct {
	profile  string
	clientID string
}

// Store is the in-memory Plane repository used by local runs and tests.
type Store struct {
	mu      sync.RWMutex
	records map[string]*record
	clients map[clientKey]string // -> app ID
}

func New() *Store {
	return &Store{
		records: make(map[string]*record),
		clients: make(map[clientKey]string),
	}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) ListApps(_ context.Context, workspaceID string) ([]*agentsv1.PlaneApp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.PlaneApp, 0, len(s.records))
	for _, rec := range s.records {
		if rec.app.GetWorkspaceId() != workspaceID {
			continue
		}
		out = append(out, materialize(rec))
	}
	sortApps(out)
	return out, nil
}

func (s *Store) FindAppByID(_ context.Context, id string) (*agentsv1.PlaneApp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.records[id]
	if !ok {
		return nil, planerepo.ErrNotFound
	}
	return materialize(rec), nil
}

func (s *Store) GetApp(_ context.Context, workspaceID, id string) (*agentsv1.PlaneApp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.records[id]
	if !ok || rec.app.GetWorkspaceId() != workspaceID {
		return nil, planerepo.ErrNotFound
	}
	return materialize(rec), nil
}

func (s *Store) CreateApp(_ context.Context, workspaceID string, app *agentsv1.PlaneApp, creds planerepo.AppCredentials) (*agentsv1.PlaneApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := clientKey{profile: app.GetProviderProfile(), clientID: strings.TrimSpace(app.GetClientId())}
	if _, ok := s.clients[key]; ok {
		return nil, planerepo.ErrClientIDExists
	}
	now := timestamppb.Now()
	stored := proto.Clone(app).(*agentsv1.PlaneApp)
	planerepo.StripDerived(stored)
	stored.WorkspaceId = workspaceID
	stored.Revision = 1
	stored.CreatedAt = now
	stored.UpdatedAt = now
	s.records[stored.Id] = &record{app: stored, creds: creds}
	s.clients[key] = stored.Id
	return materialize(s.records[stored.Id]), nil
}

func (s *Store) UpdateApp(_ context.Context, workspaceID string, app *agentsv1.PlaneApp, expectedRevision int64) (*agentsv1.PlaneApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[app.GetId()]
	if !ok || rec.app.GetWorkspaceId() != workspaceID {
		return nil, planerepo.ErrNotFound
	}
	if rec.app.GetRevision() != expectedRevision {
		return nil, planerepo.ErrRevisionConflict
	}
	next := proto.Clone(app).(*agentsv1.PlaneApp)
	planerepo.StripDerived(next)
	next.WorkspaceId = workspaceID
	next.ProviderProfile = rec.app.GetProviderProfile()
	next.ClientId = rec.app.GetClientId()
	next.Revision = rec.app.GetRevision() + 1
	next.CreatedAt = rec.app.GetCreatedAt()
	next.UpdatedAt = timestamppb.Now()
	rec.app = next
	return materialize(rec), nil
}

func (s *Store) SetAppCredentials(_ context.Context, workspaceID, id string, change planerepo.CredentialChange) (*agentsv1.PlaneApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[id]
	if !ok || rec.app.GetWorkspaceId() != workspaceID {
		return nil, planerepo.ErrNotFound
	}
	applyCredentials(&rec.creds, change)
	rec.app.UpdatedAt = timestamppb.Now()
	return materialize(rec), nil
}

func (s *Store) GetAppCredentials(_ context.Context, workspaceID, id string) (planerepo.AppCredentials, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.records[id]
	if !ok || rec.app.GetWorkspaceId() != workspaceID {
		return planerepo.AppCredentials{}, planerepo.ErrNotFound
	}
	return rec.creds, nil
}

func (s *Store) DeleteApp(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[id]
	if !ok || rec.app.GetWorkspaceId() != workspaceID {
		return planerepo.ErrNotFound
	}
	delete(s.records, id)
	delete(s.clients, clientKey{profile: rec.app.GetProviderProfile(), clientID: rec.app.GetClientId()})
	return nil
}

func applyCredentials(creds *planerepo.AppCredentials, change planerepo.CredentialChange) {
	if change.ClientSecret != nil {
		creds.ClientSecret = *change.ClientSecret
	}
	if change.WebhookSecret != nil {
		creds.WebhookSecret = *change.WebhookSecret
	}
}

func materialize(rec *record) *agentsv1.PlaneApp {
	app := proto.Clone(rec.app).(*agentsv1.PlaneApp)
	planerepo.StampCredentialState(app, rec.creds)
	return app
}

func sortApps(apps []*agentsv1.PlaneApp) {
	// Display name, then ID, keeps list output stable without relying on map
	// iteration order.
	sort.Slice(apps, func(i, j int) bool {
		if apps[i].GetDisplayName() != apps[j].GetDisplayName() {
			return apps[i].GetDisplayName() < apps[j].GetDisplayName()
		}
		return apps[i].GetId() < apps[j].GetId()
	})
}
