package memory

import (
	"context"
	"github.com/google/uuid"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
	"sync"
	"time"
)

type Store struct {
	mu      sync.RWMutex
	records map[string]*agentsv1.PlaneProcessingRecord
	keys    map[string]string
}

func New() *Store {
	return &Store{records: map[string]*agentsv1.PlaneProcessingRecord{}, keys: map[string]string{}}
}
func (s *Store) EnsureIndexes(context.Context) error { return nil }
func (s *Store) Create(_ context.Context, r *agentsv1.PlaneProcessingRecord) (*agentsv1.PlaneProcessingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := r.GetInstallationId() + ":" + r.GetRunId() + ":" + r.GetActivityId() + ":" + r.GetEventAction()
	if _, ok := s.keys[key]; ok {
		return nil, planeprocessing.ErrDuplicate
	}
	c := proto.Clone(r).(*agentsv1.PlaneProcessingRecord)
	if c.Id == "" {
		c.Id = uuid.NewString()
	}
	now := timestamppb.New(time.Now().UTC())
	if c.ReceivedAt == nil {
		c.ReceivedAt = now
	}
	c.UpdatedAt = now
	s.records[c.Id] = c
	s.keys[key] = c.Id
	return proto.Clone(c).(*agentsv1.PlaneProcessingRecord), nil
}
func (s *Store) Get(_ context.Context, ws, id string) (*agentsv1.PlaneProcessingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.records[id]
	if !ok || r.GetWorkspaceId() != ws {
		return nil, planeprocessing.ErrNotFound
	}
	return proto.Clone(r).(*agentsv1.PlaneProcessingRecord), nil
}
func (s *Store) List(_ context.Context, f planeprocessing.Filter) ([]*agentsv1.PlaneProcessingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*agentsv1.PlaneProcessingRecord{}
	for _, r := range s.records {
		if f.WorkspaceID != "" && r.GetWorkspaceId() != f.WorkspaceID || f.AppID != "" && r.GetAppId() != f.AppID || f.Status != 0 && r.GetStatus() != f.Status {
			continue
		}
		out = append(out, proto.Clone(r).(*agentsv1.PlaneProcessingRecord))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetReceivedAt().AsTime().After(out[j].GetReceivedAt().AsTime()) })
	if f.Limit > 0 && len(out) > f.Limit {
		out = out[:f.Limit]
	}
	return out, nil
}
