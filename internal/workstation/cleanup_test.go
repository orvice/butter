package workstation_test

import (
	"context"
	"errors"
	"testing"

	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/repotest"
)

type retryCollectRepo struct {
	ws.Repository
	failures int
	calls    int
}

func (r *retryCollectRepo) Collect(ctx context.Context) error {
	r.calls++
	if r.calls <= r.failures {
		return errors.New("transient cleanup failure")
	}
	return r.Repository.Collect(ctx)
}

func TestCleanupRetriesTransientCollectionFailure(t *testing.T) {
	base := memory.New()
	repo := &retryCollectRepo{Repository: base, failures: 2}
	s := repotest.NewService(repo, &repotest.Provider{})
	if err := s.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if repo.calls != 3 {
		t.Fatalf("cleanup calls = %d, want 3", repo.calls)
	}
}

func TestCleanupStopsAfterBoundedRetries(t *testing.T) {
	repo := &retryCollectRepo{Repository: memory.New(), failures: 10}
	s := repotest.NewService(repo, &repotest.Provider{})
	if err := s.Cleanup(context.Background()); err == nil {
		t.Fatal("expected cleanup failure")
	}
	if repo.calls != 3 {
		t.Fatalf("cleanup calls = %d, want 3", repo.calls)
	}
}
