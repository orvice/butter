package memory

import (
	"context"
	"github.com/google/uuid"
	repo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
	"sync"
	"time"
)

type rec struct {
	workspace, app string
	inst           *agentsv1.PlaneInstallation
	tokens         repo.Tokens
}
type Store struct {
	mu   sync.RWMutex
	rows map[string]*rec
}

func New() *Store                                    { return &Store{rows: map[string]*rec{}} }
func (s *Store) EnsureIndexes(context.Context) error { return nil }
func (s *Store) Upsert(_ context.Context, w, a string, in *agentsv1.PlaneInstallation, t repo.Tokens) (*agentsv1.PlaneInstallation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var old *rec
	for _, r := range s.rows {
		if r.workspace == w && r.app == a && r.inst.GetPlaneWorkspaceId() == in.GetPlaneWorkspaceId() {
			old = r
			break
		}
	}
	now := time.Now().UTC()
	x := proto.Clone(in).(*agentsv1.PlaneInstallation)
	if old != nil {
		x.Id = old.inst.Id
		x.InstalledAt = old.inst.InstalledAt
	} else {
		x.Id = uuid.NewString()
		x.InstalledAt = timestamppb.New(now)
	}
	x.UpdatedAt = timestamppb.New(now)
	x.TokenExpiresAt = timestamppb.New(t.ExpiresAt)
	r := &rec{workspace: w, app: a, inst: x, tokens: t}
	s.rows[x.Id] = r
	return proto.Clone(x).(*agentsv1.PlaneInstallation), nil
}
func (s *Store) List(_ context.Context, w, a string) ([]*agentsv1.PlaneInstallation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*agentsv1.PlaneInstallation{}
	for _, r := range s.rows {
		if r.workspace == w && r.app == a {
			out = append(out, proto.Clone(r.inst).(*agentsv1.PlaneInstallation))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetId() < out[j].GetId() })
	return out, nil
}
func (s *Store) Get(_ context.Context, w, id string) (*agentsv1.PlaneInstallation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rows[id]
	if !ok || r.workspace != w {
		return nil, repo.ErrNotFound
	}
	return proto.Clone(r.inst).(*agentsv1.PlaneInstallation), nil
}
