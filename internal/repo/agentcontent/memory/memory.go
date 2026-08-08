// Package memory provides an in-memory implementation of the agentcontent
// repository (issue #216).
package memory

import (
	"context"
	"sync"

	"go.orx.me/apps/butter/internal/agentcontent"
	agentcontentrepo "go.orx.me/apps/butter/internal/repo/agentcontent"
)

// Store is a thread-safe in-memory agentcontent.Repository.
type Store struct {
	mu   sync.RWMutex
	data map[string]agentcontent.Snapshot
}

// New creates a new empty in-memory store.
func New() *Store {
	return &Store{data: make(map[string]agentcontent.Snapshot)}
}

func (s *Store) PutSnapshot(_ context.Context, workspaceID string, snapshot agentcontent.Snapshot) error {
	cp := agentcontent.Snapshot{
		CommitSHA: snapshot.CommitSHA,
		Entries:   make(map[string]agentcontent.AgentContent, len(snapshot.Entries)),
	}
	for k, v := range snapshot.Entries {
		cp.Entries[k] = v
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[workspaceID] = cp
	return nil
}

func (s *Store) GetSnapshot(_ context.Context, workspaceID string) (agentcontent.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.data[workspaceID]
	if !ok {
		return agentcontent.Snapshot{}, agentcontentrepo.ErrNotFound
	}
	cp := agentcontent.Snapshot{
		CommitSHA: snap.CommitSHA,
		Entries:   make(map[string]agentcontent.AgentContent, len(snap.Entries)),
	}
	for k, v := range snap.Entries {
		cp.Entries[k] = v
	}
	return cp, nil
}

func (s *Store) Delete(_ context.Context, workspaceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, workspaceID)
	return nil
}
