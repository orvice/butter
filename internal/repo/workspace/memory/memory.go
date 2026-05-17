package memory

import (
	"context"
	"sort"
	"sync"

	"google.golang.org/protobuf/proto"

	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Store is an in-memory Repository implementation suitable for tests and
// the default storage backend.
type Store struct {
	mu         sync.RWMutex
	workspaces map[string]*agentsv1.Workspace          // id -> workspace
	bySlug     map[string]string                       // slug -> id
	members    map[string]map[string]*agentsv1.WorkspaceMember // workspaceID -> userID -> member
}

func New() *Store {
	return &Store{
		workspaces: make(map[string]*agentsv1.Workspace),
		bySlug:     make(map[string]string),
		members:    make(map[string]map[string]*agentsv1.WorkspaceMember),
	}
}

func (s *Store) EnsureIndexes(_ context.Context) error { return nil }

func (s *Store) ListWorkspaces(_ context.Context) ([]*agentsv1.Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.Workspace, 0, len(s.workspaces))
	for _, ws := range s.workspaces {
		out = append(out, proto.Clone(ws).(*agentsv1.Workspace))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetSlug() < out[j].GetSlug() })
	return out, nil
}

func (s *Store) GetWorkspace(_ context.Context, id string) (*agentsv1.Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ws, ok := s.workspaces[id]
	if !ok {
		return nil, workspacerepo.ErrNotFound
	}
	return proto.Clone(ws).(*agentsv1.Workspace), nil
}

func (s *Store) GetWorkspaceBySlug(_ context.Context, slug string) (*agentsv1.Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.bySlug[slug]
	if !ok {
		return nil, workspacerepo.ErrNotFound
	}
	return proto.Clone(s.workspaces[id]).(*agentsv1.Workspace), nil
}

func (s *Store) CreateWorkspace(_ context.Context, ws *agentsv1.Workspace) (*agentsv1.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workspaces[ws.GetId()]; ok {
		return nil, workspacerepo.ErrAlreadyExists
	}
	if _, ok := s.bySlug[ws.GetSlug()]; ok {
		return nil, workspacerepo.ErrAlreadyExists
	}
	stored := proto.Clone(ws).(*agentsv1.Workspace)
	s.workspaces[stored.GetId()] = stored
	s.bySlug[stored.GetSlug()] = stored.GetId()
	return proto.Clone(stored).(*agentsv1.Workspace), nil
}

func (s *Store) UpdateWorkspace(_ context.Context, ws *agentsv1.Workspace) (*agentsv1.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.workspaces[ws.GetId()]
	if !ok {
		return nil, workspacerepo.ErrNotFound
	}
	if existing.GetSlug() != ws.GetSlug() {
		if _, taken := s.bySlug[ws.GetSlug()]; taken {
			return nil, workspacerepo.ErrAlreadyExists
		}
		delete(s.bySlug, existing.GetSlug())
		s.bySlug[ws.GetSlug()] = ws.GetId()
	}
	stored := proto.Clone(ws).(*agentsv1.Workspace)
	s.workspaces[stored.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.Workspace), nil
}

func (s *Store) DeleteWorkspace(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ws, ok := s.workspaces[id]
	if !ok {
		return workspacerepo.ErrNotFound
	}
	delete(s.workspaces, id)
	delete(s.bySlug, ws.GetSlug())
	delete(s.members, id)
	return nil
}

func (s *Store) CountWorkspaces(_ context.Context) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.workspaces)), nil
}

func (s *Store) ListMembers(_ context.Context, workspaceID string) ([]*agentsv1.WorkspaceMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.members[workspaceID]
	if !ok {
		return nil, nil
	}
	out := make([]*agentsv1.WorkspaceMember, 0, len(m))
	for _, member := range m {
		out = append(out, proto.Clone(member).(*agentsv1.WorkspaceMember))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetUserId() < out[j].GetUserId() })
	return out, nil
}

func (s *Store) ListMembershipsForUser(_ context.Context, userID string) ([]*agentsv1.WorkspaceMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.WorkspaceMember
	for _, members := range s.members {
		if member, ok := members[userID]; ok {
			out = append(out, proto.Clone(member).(*agentsv1.WorkspaceMember))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetWorkspaceId() < out[j].GetWorkspaceId() })
	return out, nil
}

func (s *Store) IsMember(_ context.Context, workspaceID, userID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.members[workspaceID]
	if !ok {
		return false, nil
	}
	_, ok = m[userID]
	return ok, nil
}

func (s *Store) GetMember(_ context.Context, workspaceID, userID string) (*agentsv1.WorkspaceMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.members[workspaceID]
	if !ok {
		return nil, workspacerepo.ErrNotFound
	}
	member, ok := m[userID]
	if !ok {
		return nil, workspacerepo.ErrNotFound
	}
	return proto.Clone(member).(*agentsv1.WorkspaceMember), nil
}

func (s *Store) AddMember(_ context.Context, member *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workspaces[member.GetWorkspaceId()]; !ok {
		return nil, workspacerepo.ErrNotFound
	}
	bucket := s.members[member.GetWorkspaceId()]
	if bucket == nil {
		bucket = make(map[string]*agentsv1.WorkspaceMember)
		s.members[member.GetWorkspaceId()] = bucket
	}
	if _, ok := bucket[member.GetUserId()]; ok {
		return nil, workspacerepo.ErrAlreadyExists
	}
	stored := proto.Clone(member).(*agentsv1.WorkspaceMember)
	bucket[stored.GetUserId()] = stored
	return proto.Clone(stored).(*agentsv1.WorkspaceMember), nil
}

func (s *Store) UpdateMember(_ context.Context, member *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.members[member.GetWorkspaceId()]
	if !ok {
		return nil, workspacerepo.ErrNotFound
	}
	if _, ok := bucket[member.GetUserId()]; !ok {
		return nil, workspacerepo.ErrNotFound
	}
	stored := proto.Clone(member).(*agentsv1.WorkspaceMember)
	bucket[stored.GetUserId()] = stored
	return proto.Clone(stored).(*agentsv1.WorkspaceMember), nil
}

func (s *Store) RemoveMember(_ context.Context, workspaceID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.members[workspaceID]
	if !ok {
		return workspacerepo.ErrNotFound
	}
	if _, ok := bucket[userID]; !ok {
		return workspacerepo.ErrNotFound
	}
	delete(bucket, userID)
	return nil
}
