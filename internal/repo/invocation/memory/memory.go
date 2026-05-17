package memory

import (
	"context"
	"sort"
	"strconv"
	"sync"

	"go.orx.me/apps/butter/internal/repo/invocation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

// Store is a thread-safe in-memory implementation of invocation.Repository.
type Store struct {
	mu      sync.RWMutex
	byID    map[string]*agentsv1.Invocation
	ordered []string // insertion order (oldest first); newest at end
}

func New() *Store {
	return &Store{byID: make(map[string]*agentsv1.Invocation)}
}

func (s *Store) Save(_ context.Context, inv *agentsv1.Invocation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := inv.GetId()
	if _, exists := s.byID[id]; !exists {
		s.ordered = append(s.ordered, id)
	}
	s.byID[id] = proto.Clone(inv).(*agentsv1.Invocation)
	return nil
}

func (s *Store) Get(_ context.Context, id string) (*agentsv1.Invocation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inv, ok := s.byID[id]
	if !ok {
		return nil, invocation.ErrNotFound
	}
	return proto.Clone(inv).(*agentsv1.Invocation), nil
}

func (s *Store) List(_ context.Context, filter invocation.ListFilter, pageSize int32, pageToken string) ([]*agentsv1.Invocation, string, int32, error) {
	all := s.snapshotDesc()
	matched := make([]*agentsv1.Invocation, 0, len(all))
	for _, inv := range all {
		if filter.WorkspaceID != "" && inv.GetWorkspaceId() != filter.WorkspaceID {
			continue
		}
		if filter.AgentName != "" && inv.GetAgentName() != filter.AgentName {
			continue
		}
		if filter.SessionID != "" && inv.GetSessionId() != filter.SessionID {
			continue
		}
		matched = append(matched, inv)
	}
	page, next := paginate(matched, pageSize, pageToken)
	return page, next, int32(len(matched)), nil
}

func (s *Store) ListRecent(_ context.Context, limit int32, pageToken string) ([]*agentsv1.Invocation, string, error) {
	all := s.snapshotDesc()
	page, next := paginate(all, limit, pageToken)
	return page, next, nil
}

func (s *Store) snapshotDesc() []*agentsv1.Invocation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.Invocation, 0, len(s.ordered))
	for i := len(s.ordered) - 1; i >= 0; i-- {
		inv, ok := s.byID[s.ordered[i]]
		if !ok {
			continue
		}
		out = append(out, proto.Clone(inv).(*agentsv1.Invocation))
	}
	// Defensive sort by started_at desc in case ordering drifted.
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].GetStartedAt().AsTime().After(out[j].GetStartedAt().AsTime())
	})
	return out
}

func paginate(items []*agentsv1.Invocation, pageSize int32, pageToken string) ([]*agentsv1.Invocation, string) {
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := 0
	if pageToken != "" {
		if n, err := strconv.Atoi(pageToken); err == nil && n >= 0 {
			offset = n
		}
	}
	if offset >= len(items) {
		return nil, ""
	}
	end := offset + int(pageSize)
	if end > len(items) {
		end = len(items)
	}
	next := ""
	if end < len(items) {
		next = strconv.Itoa(end)
	}
	return items[offset:end], next
}
