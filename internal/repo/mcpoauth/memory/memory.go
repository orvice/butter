package memory

import (
	"context"
	"sync"
	"time"

	repo "go.orx.me/apps/butter/internal/repo/mcpoauth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Store is a thread-safe in-memory MCP OAuth connection repository.
type Store struct {
	mu          sync.RWMutex
	connections map[string]map[string]*repo.Connection
}

func New() *Store {
	return &Store{connections: make(map[string]map[string]*repo.Connection)}
}

func (s *Store) EnsureIndexes(context.Context) error {
	return nil
}

func (s *Store) Get(_ context.Context, workspaceID, serverID string) (*repo.Connection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	byServer := s.connections[workspaceID]
	if byServer == nil {
		return nil, repo.ErrNotFound
	}
	conn := byServer[serverID]
	if conn == nil {
		return nil, repo.ErrNotFound
	}
	return conn.Clone(), nil
}

func (s *Store) Save(_ context.Context, conn *repo.Connection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	byServer := s.connections[conn.WorkspaceID]
	if byServer == nil {
		byServer = make(map[string]*repo.Connection)
		s.connections[conn.WorkspaceID] = byServer
	}
	now := time.Now().UTC()
	stored := conn.Clone()
	if stored.CreatedAt.IsZero() {
		if current := byServer[stored.ServerID]; current != nil && !current.CreatedAt.IsZero() {
			stored.CreatedAt = current.CreatedAt
		} else {
			stored.CreatedAt = now
		}
	}
	if stored.UpdatedAt.IsZero() {
		stored.UpdatedAt = now
	}
	byServer[stored.ServerID] = stored
	return nil
}

func (s *Store) Delete(_ context.Context, workspaceID, serverID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	byServer := s.connections[workspaceID]
	if byServer == nil {
		return nil
	}
	delete(byServer, serverID)
	return nil
}

func (s *Store) MarkState(_ context.Context, workspaceID, serverID string, state agentsv1.MCPOAuthConnectionState, detail string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	byServer := s.connections[workspaceID]
	if byServer == nil || byServer[serverID] == nil {
		return repo.ErrNotFound
	}
	conn := byServer[serverID].Clone()
	conn.State = state
	conn.LastError = detail
	conn.LastCheckedAt = at
	conn.UpdatedAt = at
	conn.ReauthorizationRequired = state == agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED
	byServer[serverID] = conn
	return nil
}
