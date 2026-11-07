package planprogressmemory

import (
	"context"
	mongoprogress "go.orx.me/apps/butter/internal/repo/planeprogress/mongo"
	plane "go.orx.me/apps/butter/internal/runtime/plane"
	"sort"
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	rows map[string]mongoprogress.Record
}

func New() *Store { return &Store{rows: map[string]mongoprogress.Record{}} }
func (s *Store) EmitProgress(_ context.Context, e *plane.Event, p plane.ProgressEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.rows[p.Key]
	if ok && old.WorkspaceID == e.WorkspaceID && old.Sequence >= p.Sequence {
		return nil
	}
	s.rows[p.Key] = mongoprogress.Record{Key: p.Key, WorkspaceID: e.WorkspaceID, AppID: e.AppID, InstallationID: e.InstallationID, RunID: e.AgentRunID, ActivityID: e.ActivityID, WorkflowID: p.WorkflowID, DeliveryID: p.DeliveryID, Sequence: p.Sequence, State: string(p.State), Error: p.Error, UpdatedAt: time.Now().UTC()}
	return nil
}
func (s *Store) List(_ context.Context, q mongoprogress.Query) ([]mongoprogress.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]mongoprogress.Record, 0)
	for _, r := range s.rows {
		if r.WorkspaceID != q.WorkspaceID || (q.AppID != "" && r.AppID != q.AppID) || (q.InstallationID != "" && r.InstallationID != q.InstallationID) || (q.RunID != "" && r.RunID != q.RunID) || (q.ActivityID != "" && r.ActivityID != q.ActivityID) {
			continue
		}
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].Sequence > out[j].Sequence
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out, nil
}
