// Package memory implements the same publication contract for single-process
// development. Production uses mongostore, including its server-time fences.
package memory

import (
	"context"
	"slices"
	"sync"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

type Store struct {
	mu           sync.Mutex
	records      map[string]ws.Record
	observations []ws.Observation
	Clock        func() time.Time
	hosts        map[string]ws.Host
	budgets      map[string]time.Time
}

func New() *Store {
	return &Store{records: map[string]ws.Record{}, hosts: map[string]ws.Host{}, budgets: map[string]time.Time{}, Clock: time.Now}
}
func key(w, id string) string { return w + "\x00" + id }
func clone(r ws.Record) ws.Record {
	r.Scopes = ws.CloneScopes(r.Scopes)
	return r
}

func (s *Store) EnsureIndexes(context.Context) error    { return nil }
func (s *Store) Now(context.Context) (time.Time, error) { return s.Clock().UTC(), nil }
func (s *Store) List(_ context.Context, w string) ([]ws.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []ws.Record{}
	for _, r := range s.records {
		if r.WorkspaceID == w {
			out = append(out, clone(r))
		}
	}
	return out, nil
}
func (s *Store) Due(context.Context) ([]ws.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []ws.Record{}
	for _, r := range s.records {
		if r.Paused || r.Credential.Ciphertext == "" {
			continue
		}
		for _, sc := range r.Scopes {
			if !sc.Hidden && !sc.Due.After(s.Clock()) && !sc.LeaseUntil.After(s.Clock()) {
				out = append(out, clone(r))
				break
			}
		}
	}
	return out, nil
}
func (s *Store) Get(_ context.Context, w, id string) (ws.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.records[key(w, id)]
	if !ok {
		return r, ws.ErrNotFound
	}
	return clone(r), nil
}
func (s *Store) Create(_ context.Context, r ws.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := key(r.WorkspaceID, r.ID)
	if _, ok := s.records[k]; ok {
		return ws.ErrConflict
	}
	s.records[k] = clone(r)
	return nil
}
func (s *Store) CAS(_ context.Context, before ws.Record, after *ws.Record, f *ws.Fence) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := key(before.WorkspaceID, before.ID)
	r, ok := s.records[k]
	if !ok || r.Version != before.Version {
		return false, nil
	}
	if f != nil && (!s.Clock().Before(f.Until) || !hasClaim(r, f.Token)) {
		return false, nil
	}
	if after == nil {
		delete(s.records, k)
	} else {
		s.records[k] = clone(*after)
	}
	return true, nil
}
func (s *Store) Stage(_ context.Context, observations []ws.Observation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, o := range observations {
		o.Item = proto.Clone(o.Item).(*v1.WorkStationItem)
		s.observations = append(s.observations, o)
	}
	return nil
}
func (s *Store) Query(_ context.Context, w string, q ws.Query) ([]*v1.WorkStationItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rs := []ws.Record{}
	for _, r := range s.records {
		if r.WorkspaceID == w {
			rs = append(rs, r)
		}
	}
	obs := append([]ws.Observation(nil), s.observations...)
	return ws.Project(rs, obs, q), nil
}
func (s *Store) HostInUse(_ context.Context, id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.records {
		if r.HostID == id {
			return true, nil
		}
	}
	return false, nil
}
func (s *Store) PruneSnapshot(_ context.Context, w, connection, scope, snapshot string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.observations[:0]
	for _, o := range s.observations {
		if o.WorkspaceID == w && o.ConnectionID == connection && o.ScopeID == scope && o.Snapshot == snapshot {
			continue
		}
		kept = append(kept, o)
	}
	s.observations = kept
	return nil
}
func (s *Store) Collect(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.observations[:0]
	for _, o := range s.observations {
		r, ok := s.records[key(o.WorkspaceID, o.ConnectionID)]
		current := false
		for _, sc := range r.Scopes {
			if sc.Snapshot == o.Snapshot || (sc.Pending != nil && slices.Contains(sc.Pending.Batches, o.Snapshot)) {
				current = true
			}
		}
		if (ok && current) || o.Item.ObservedAt.AsTime().After(s.Clock().Add(-ws.OrphanGrace)) {
			kept = append(kept, o)
		}
	}
	s.observations = kept
	return nil
}

func hasClaim(r ws.Record, token string) bool {
	for _, sc := range r.Scopes {
		if sc.Claim == token {
			return true
		}
	}
	return false
}

func (s *Store) Facets(ctx context.Context, w string, q ws.Query, dimension, search, after string, limit int) ([]*v1.WorkStationFilterOption, error) {
	q.Limit = 0
	items, err := s.Query(ctx, w, q)
	if err != nil {
		return nil, err
	}
	return ws.ProjectOptions(items, dimension, search, after, limit), nil
}

func (s *Store) Staged(_ context.Context, w, connection, scope string, tokens []string) ([]*v1.WorkStationItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []*v1.WorkStationItem{}
	for _, o := range s.observations {
		if o.WorkspaceID == w && o.ConnectionID == connection && o.ScopeID == scope && slices.Contains(tokens, o.Snapshot) {
			out = append(out, proto.Clone(o.Item).(*v1.WorkStationItem))
		}
	}
	return out, nil
}
