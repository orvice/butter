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
	claims  map[string]planeprocessing.Claim
}

func New() *Store {
	return &Store{records: map[string]*agentsv1.PlaneProcessingRecord{}, keys: map[string]string{}, claims: map[string]planeprocessing.Claim{}}
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

func (s *Store) Claim(_ context.Context, ws, app, installation, run, owner string, expiry time.Time) (planeprocessing.Claim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := ws + ":" + app + ":" + installation + ":" + run
	c, ok := s.claims[key]
	now := time.Now()
	if ok && c.ExpiresAt.After(now) && c.Owner != owner {
		return planeprocessing.Claim{}, planeprocessing.ErrClaimConflict
	}
	c.Owner = owner
	c.Revision++
	c.Fence++
	c.ExpiresAt = expiry
	s.claims[key] = c
	return c, nil
}
func (s *Store) Renew(_ context.Context, key string, in planeprocessing.Claim, expiry time.Time) (planeprocessing.Claim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.claims[key]
	if !ok || c.Owner != in.Owner || c.Revision != in.Revision || c.Fence != in.Fence || !c.ExpiresAt.After(time.Now()) {
		return planeprocessing.Claim{}, planeprocessing.ErrClaimExpired
	}
	c.ExpiresAt = expiry
	s.claims[key] = c
	return c, nil
}
func (s *Store) Complete(_ context.Context, key string, in planeprocessing.Claim) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.claims[key]
	if !ok || c.Owner != in.Owner || c.Revision != in.Revision || c.Fence != in.Fence || !c.ExpiresAt.After(time.Now()) {
		return planeprocessing.ErrClaimExpired
	}
	delete(s.claims, key)
	return nil
}
func (s *Store) ReclaimExpired(_ context.Context, ws, app, installation, run, owner string, expiry time.Time) (planeprocessing.Claim, error) {
	return s.Claim(context.Background(), ws, app, installation, run, owner, expiry)
}
func (s *Store) RecoverExpired(_ context.Context, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, c := range s.claims {
		if !c.ExpiresAt.After(now) {
			delete(s.claims, k)
		}
	}
	return nil
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

func (s *Store) UpdateStatus(_ context.Context, ws, app, installation, run, activity, action string, status agentsv1.PlaneProcessingStatus, failure string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := installation + ":" + run + ":" + activity + ":" + action
	id, ok := s.keys[key]
	if !ok {
		return planeprocessing.ErrNotFound
	}
	r := s.records[id]
	if r.GetWorkspaceId() != ws || r.GetAppId() != app {
		return planeprocessing.ErrNotFound
	}
	r = proto.Clone(r).(*agentsv1.PlaneProcessingRecord)
	r.Status = status
	r.FailureReason = failure
	r.UpdatedAt = timestamppb.Now()
	s.records[id] = r
	return nil
}
