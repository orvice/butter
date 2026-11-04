package memory_test

import (
	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/repotest"
	"testing"
	"time"
)

func TestContract(t *testing.T) {
	repotest.Run(t, func(*testing.T) ws.Repository { return memory.New() })
}

func TestDurableBatchResume(t *testing.T) {
	now := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	repo := memory.New()
	repo.Clock = func() time.Time { return now }
	repotest.RunBatches(t, repo, func(d time.Duration) { now = now.Add(d) })
}
