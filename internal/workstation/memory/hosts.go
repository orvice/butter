package memory

import (
	"context"
	ws "go.orx.me/apps/butter/internal/workstation"
)

func (s *Store) ListHosts(context.Context) ([]ws.Host, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []ws.Host{ws.CloudHost()}
	for _, h := range s.hosts {
		out = append(out, h)
	}
	return out, nil
}
func (s *Store) GetHost(_ context.Context, id string) (ws.Host, error) {
	if id == ws.PlaneCloudHost {
		return ws.CloudHost(), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.hosts[id]
	if !ok {
		return h, ws.ErrNotFound
	}
	return h, nil
}
func (s *Store) SaveHost(_ context.Context, h ws.Host, revision int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.hosts[h.ID]
	if h.ID == ws.PlaneCloudHost || (ok && old.Revision != revision) || (!ok && revision != 0) {
		return ws.ErrConflict
	}
	s.hosts[h.ID] = h
	return nil
}
func (s *Store) DeleteHost(_ context.Context, id string, revision int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == ws.PlaneCloudHost {
		return ws.ErrInvalid
	}
	h, ok := s.hosts[id]
	if !ok {
		return nil
	}
	if h.Revision != revision {
		return ws.ErrConflict
	}
	for _, r := range s.records {
		if r.HostID == id {
			return ws.ErrConflict
		}
	}
	delete(s.hosts, id)
	return nil
}
