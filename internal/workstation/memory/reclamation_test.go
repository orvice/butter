package memory

import (
	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/repotest"
	"testing"
)

func TestReclamation(t *testing.T) {
	repotest.RunReclamation(t, func(*testing.T) (ws.Repository, func() int64) {
		repo := New()
		return repo, func() int64 { repo.mu.Lock(); defer repo.mu.Unlock(); return int64(len(repo.observations)) }
	})
}
