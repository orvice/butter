package memory

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"go.orx.me/apps/butter/internal/repo/invocation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Store is a thread-safe in-memory implementation of invocation.Repository.
type Store struct {
	*records
	// owner is the instance ID stamped on the records this store creates.
	owner string
}

// records is the data every view of one Store shares.
type records struct {
	mu      sync.RWMutex
	byID    map[string]*agentsv1.Invocation
	owners  map[string]string // invocation ID → owner stamp
	ordered []string          // insertion order (oldest first); newest at end
}

var _ invocation.Repository = (*Store)(nil)

// New returns an empty store whose records carry no owner stamp.
func New() *Store {
	return &Store{records: &records{
		byID:   make(map[string]*agentsv1.Invocation),
		owners: make(map[string]string),
	}}
}

// WithOwner returns a view of the same records that stamps each record it
// creates with owner: the instance ID of the process that runs it. Each
// process opens one; tests open several over one store to stand in for
// processes that share a database.
func (s *Store) WithOwner(owner string) *Store {
	return &Store{records: s.records, owner: owner}
}

func (s *Store) Save(_ context.Context, inv *agentsv1.Invocation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := inv.GetId()
	if _, exists := s.byID[id]; !exists {
		s.ordered = append(s.ordered, id)
		if s.owner != "" {
			s.owners[id] = s.owner
		}
	}
	s.byID[id] = proto.Clone(inv).(*agentsv1.Invocation)
	return nil
}

func (s *Store) Get(_ context.Context, workspaceID, id string) (*agentsv1.Invocation, error) {
	inv, err := s.get(id)
	if err != nil || inv.GetWorkspaceId() != workspaceID {
		return nil, invocation.ErrNotFound
	}
	return inv, nil
}

func (s *Store) GetAcrossWorkspaces(_ context.Context, id string) (*agentsv1.Invocation, error) {
	return s.get(id)
}

func (s *Store) get(id string) (*agentsv1.Invocation, error) {
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
		if filter.AgentID != "" && inv.GetAgentId() != filter.AgentID {
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

func (s *Store) StatusSummaries(_ context.Context, workspaceID string, agentNames []string) (map[string]invocation.StatusSummary, error) {
	wanted := make(map[string]bool, len(agentNames))
	for _, n := range agentNames {
		wanted[n] = true
	}
	out := make(map[string]invocation.StatusSummary, len(agentNames))
	for _, inv := range s.snapshotDesc() {
		if inv.GetWorkspaceId() != workspaceID || !wanted[inv.GetAgentName()] {
			continue
		}
		sum := out[inv.GetAgentName()]
		if sum.Latest == nil {
			sum.Latest = inv
		}
		if inv.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
			sum.Running++
		}
		out[inv.GetAgentName()] = sum
	}
	return out, nil
}

func (s *Store) CountByTimeRange(_ context.Context, start, end time.Time) (int64, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var total, failed int64
	for _, inv := range s.byID {
		ts := inv.GetStartedAt()
		if ts == nil {
			continue
		}
		t := ts.AsTime()
		if t.Before(start) || !t.Before(end) {
			continue
		}
		total++
		if inv.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
			failed++
		}
	}
	return total, failed, nil
}

func (s *Store) FindByRequestID(_ context.Context, workspaceID, requestID string) (*agentsv1.Invocation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, inv := range s.byID {
		if inv.GetWorkspaceId() == workspaceID && inv.GetRequestId() == requestID {
			return proto.Clone(inv).(*agentsv1.Invocation), nil
		}
	}
	return nil, invocation.ErrNotFound
}

func (s *Store) FindActiveBySession(_ context.Context, workspaceID, sessionID string) (*agentsv1.Invocation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, inv := range s.byID {
		if inv.GetWorkspaceId() != workspaceID || inv.GetSessionId() != sessionID {
			continue
		}
		st := inv.GetStatus()
		if st == agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED ||
			st == agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
			return proto.Clone(inv).(*agentsv1.Invocation), nil
		}
	}
	return nil, invocation.ErrNotFound
}

func (s *Store) ActiveOwners(_ context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := make(map[string]bool)
	owners := []string{}
	for id, inv := range s.byID {
		owner := s.owners[id]
		if owner == "" || seen[owner] || !isActive(inv.GetStatus()) {
			continue
		}
		seen[owner] = true
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	return owners, nil
}

func (s *Store) MarkStaleRunning(_ context.Context, sel invocation.StaleSelection) (int64, error) {
	lost := make(map[string]bool, len(sel.LostOwners))
	for _, owner := range sel.LostOwners {
		lost[owner] = true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var count int64
	for id, inv := range s.byID {
		if !isActive(inv.GetStatus()) {
			continue
		}
		owner := s.owners[id]
		if owner != "" && !lost[owner] {
			continue
		}
		if owner == "" && !startedBefore(inv, sel.LegacyBefore) {
			continue
		}
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
		inv.Error = ""
		if sel.Reason != nil {
			inv.Error = sel.Reason(owner)
		}
		inv.FinishedAt = timestamppb.Now()
		count++
	}
	return count, nil
}

func isActive(st agentsv1.InvocationStatus) bool {
	return st == agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED ||
		st == agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING
}

// startedBefore reports whether inv started before cutoff. A record with no
// start time counts as older than any cutoff; the zero cutoff matches none.
func startedBefore(inv *agentsv1.Invocation, cutoff time.Time) bool {
	if cutoff.IsZero() {
		return false
	}
	return inv.GetStartedAt() == nil || inv.GetStartedAt().AsTime().Before(cutoff)
}

func (s *Store) FindLatestBySession(_ context.Context, workspaceID, sessionID string) (*agentsv1.Invocation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var latest *agentsv1.Invocation
	for _, inv := range s.byID {
		if inv.GetWorkspaceId() != workspaceID || inv.GetSessionId() != sessionID {
			continue
		}
		// Ties on started_at break by descending ID, matching the mongo
		// implementation's (started_at desc, _id desc) sort; invocation IDs
		// are v7 UUIDs, so a higher ID is the later-created record.
		if latest == nil ||
			inv.GetStartedAt().AsTime().After(latest.GetStartedAt().AsTime()) ||
			(inv.GetStartedAt().AsTime().Equal(latest.GetStartedAt().AsTime()) && inv.GetId() > latest.GetId()) {
			latest = inv
		}
	}
	if latest == nil {
		return nil, invocation.ErrNotFound
	}
	return proto.Clone(latest).(*agentsv1.Invocation), nil
}

func (s *Store) ListBySession(_ context.Context, workspaceID, sessionID string) ([]*agentsv1.Invocation, error) {
	all := s.snapshotDesc()
	var out []*agentsv1.Invocation
	for _, inv := range all {
		if inv.GetWorkspaceId() == workspaceID && inv.GetSessionId() == sessionID {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (s *Store) RedactContent(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.byID[id]
	if !ok || inv.GetWorkspaceId() != workspaceID {
		return invocation.ErrNotFound
	}
	inv.Input = ""
	inv.Output = ""
	inv.Error = ""
	return nil
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
