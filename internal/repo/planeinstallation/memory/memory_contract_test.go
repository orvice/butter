package memory

import (
	"context"
	repo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"sync"
	"testing"
	"time"
)

func TestRuntimeTokenOwnershipAndCAS(t *testing.T) {
	s := New()
	ctx := context.Background()
	in := &agentsv1.PlaneInstallation{AppInstallationId: "inst", PlaneWorkspaceId: "pw"}
	if _, e := s.Upsert(ctx, "w", "a", in, repo.Tokens{AccessToken: "token", ExpiresAt: time.Now().Add(time.Hour)}); e != nil {
		t.Fatal(e)
	}
	m, e := s.GetRuntimeToken(ctx, "w", "a", "inst")
	if e != nil || m.AccessToken != "token" {
		t.Fatalf("material=%+v err=%v", m, e)
	}
	if _, e = s.GetRuntimeToken(ctx, "other", "a", "inst"); e != repo.ErrNotFound {
		t.Fatalf("ownership err=%v", e)
	}
	if _, e = s.CASRuntimeToken(ctx, "w", "a", "inst", m.Revision+1, repo.Tokens{AccessToken: "x"}); e != repo.ErrTokenCAS {
		t.Fatalf("cas err=%v", e)
	}
	if _, e = s.CASRuntimeToken(ctx, "w", "a", "inst", m.Revision, repo.Tokens{AccessToken: "new", ExpiresAt: time.Now().Add(time.Hour)}); e != nil {
		t.Fatal(e)
	}
}
func TestRuntimeTokenConcurrentCASHasSingleWinner(t *testing.T) {
	s := New()
	ctx := context.Background()
	in := &agentsv1.PlaneInstallation{AppInstallationId: "inst"}
	s.Upsert(ctx, "w", "a", in, repo.Tokens{AccessToken: "old"})
	m, _ := s.GetRuntimeToken(ctx, "w", "a", "inst")
	var wg sync.WaitGroup
	var wins int
	var mu sync.Mutex
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, e := s.CASRuntimeToken(ctx, "w", "a", "inst", m.Revision, repo.Tokens{AccessToken: "fresh"}); e == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("wins=%d", wins)
	}
}
