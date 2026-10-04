package invocation_test

// The stale sweep's rule (#390, ADR-0016 decision 3), run against the
// in-process liveness registry and — when REDIS_ADDR is set — against Redis.

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/liveness"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type registryFactory func(t *testing.T, ttl time.Duration) liveness.Registry

func memoryRegistry(_ *testing.T, ttl time.Duration) liveness.Registry {
	return liveness.NewMemory(ttl)
}

func redisRegistry(t *testing.T, ttl time.Duration) liveness.Registry {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is required for the stale sweep against Redis")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(t.Context()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	// Instance IDs are fresh UUIDs and every key carries a short TTL, so the
	// test's keys neither collide with others nor outlive it for long.
	return liveness.NewRedis(rdb, ttl)
}

func newInstance(name string) string { return name + "-" + uuid.NewString() }

const (
	running = agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING
	queued  = agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED
	failed  = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
	done    = agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED
)

func save(t *testing.T, repo invocation.Repository, id string, status agentsv1.InvocationStatus, started time.Time) {
	t.Helper()
	if err := repo.Save(t.Context(), &agentsv1.Invocation{
		Id:          id,
		WorkspaceId: "ws-1",
		SessionId:   "session-" + id,
		Status:      status,
		StartedAt:   timestamppb.New(started),
	}); err != nil {
		t.Fatalf("Save(%s): %v", id, err)
	}
}

func status(t *testing.T, repo invocation.Repository, id string) *agentsv1.Invocation {
	t.Helper()
	inv, err := repo.GetAcrossWorkspaces(t.Context(), id)
	if err != nil {
		t.Fatalf("Get(%s): %v", id, err)
	}
	return inv
}

func keep(t *testing.T, ctx context.Context, r liveness.Registry, id string) {
	t.Helper()
	if err := liveness.Keep(ctx, r, id); err != nil {
		t.Fatalf("Keep(%s): %v", id, err)
	}
}

func TestStaleSweep(t *testing.T) {
	for name, factory := range map[string]registryFactory{"memory": memoryRegistry, "redis": redisRegistry} {
		t.Run(name, func(t *testing.T) { runStaleSweep(t, factory) })
	}
}

func runStaleSweep(t *testing.T, factory registryFactory) {
	const ttl = 300 * time.Millisecond

	// Acceptance: process B starts while process A has a RUNNING record and
	// a live liveness key. The record stays RUNNING.
	t.Run("AnotherProcessStartingLeavesALiveOwnersRunsAlone", func(t *testing.T) {
		registry := factory(t, ttl)
		podA, podB := newInstance("pod-a"), newInstance("pod-b")
		keep(t, t.Context(), registry, podA)
		store := memory.New()
		save(t, store.WithOwner(podA), "a-running", running, time.Now())
		save(t, store.WithOwner(podA), "a-queued", queued, time.Now())

		keep(t, t.Context(), registry, podB)
		sweeper := &invocation.StaleSweeper{Repo: store.WithOwner(podB), Liveness: registry, Self: podB, LegacyAge: invocation.LegacyStaleAge}
		// Sweep at B's startup and again over several of A's renewals.
		for range 3 {
			result, err := sweeper.Sweep(t.Context())
			if err != nil {
				t.Fatalf("Sweep: %v", err)
			}
			if result.Failed != 0 || len(result.LostOwners) != 0 {
				t.Fatalf("Sweep = %+v, want nothing failed while A renews", result)
			}
			time.Sleep(ttl / 2)
		}
		if got := status(t, store, "a-running").GetStatus(); got != running {
			t.Fatalf("A's run is %v, want RUNNING", got)
		}
		if got := status(t, store, "a-queued").GetStatus(); got != queued {
			t.Fatalf("A's queued run is %v, want QUEUED", got)
		}
		// The one-active-run check still sees A's run.
		if active, err := store.FindActiveBySession(t.Context(), "ws-1", "session-a-running"); err != nil || active.GetId() != "a-running" {
			t.Fatalf("FindActiveBySession = %v, %v; want A's run", active, err)
		}
	})

	// Acceptance: A stops renewing, as a crashed process does. A later sweep
	// fails A's QUEUED and RUNNING records, naming A.
	t.Run("ARunIsFailedOnceItsOwnerStopsRenewing", func(t *testing.T) {
		registry := factory(t, ttl)
		podA, podB := newInstance("pod-a"), newInstance("pod-b")
		ctxA, crashA := context.WithCancel(t.Context())
		defer crashA()
		keep(t, ctxA, registry, podA)
		keep(t, t.Context(), registry, podB)
		store := memory.New()
		a := store.WithOwner(podA)
		save(t, a, "a-running", running, time.Now())
		save(t, a, "a-queued", queued, time.Now())
		save(t, a, "a-done", done, time.Now())
		save(t, store.WithOwner(podB), "b-running", running, time.Now())
		sweeper := &invocation.StaleSweeper{Repo: store.WithOwner(podB), Liveness: registry, Self: podB, LegacyAge: invocation.LegacyStaleAge}

		if result, err := sweeper.Sweep(t.Context()); err != nil || result.Failed != 0 {
			t.Fatalf("Sweep before the crash = %+v, %v; want nothing failed", result, err)
		}

		crashA()
		deadline := time.Now().Add(3 * time.Second)
		var result invocation.SweepResult
		for {
			var err error
			if result, err = sweeper.Sweep(t.Context()); err != nil {
				t.Fatalf("Sweep: %v", err)
			}
			if result.Failed > 0 || time.Now().After(deadline) {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if result.Failed != 2 || len(result.LostOwners) != 1 || result.LostOwners[0] != podA {
			t.Fatalf("Sweep after A stopped renewing = %+v, want A's 2 active records failed and A reported lost", result)
		}
		for _, id := range []string{"a-running", "a-queued"} {
			inv := status(t, store, id)
			if inv.GetStatus() != failed {
				t.Fatalf("%s is %v, want FAILED", id, inv.GetStatus())
			}
			if !strings.Contains(inv.GetError(), podA) || !strings.Contains(inv.GetError(), "resubmit") {
				t.Fatalf("%s error = %q, want a reason that names %s and points to resubmitting", id, inv.GetError(), podA)
			}
			if inv.GetFinishedAt() == nil {
				t.Fatalf("%s has no finished_at", id)
			}
		}
		if got := status(t, store, "a-done").GetStatus(); got != done {
			t.Fatalf("A's finished run is %v, want it untouched", got)
		}
		if got := status(t, store, "b-running").GetStatus(); got != running {
			t.Fatalf("B's own run is %v, want RUNNING", got)
		}
	})

	t.Run("AProcessNeverFailsItsOwnRuns", func(t *testing.T) {
		// Nothing is published: even this process's own key is missing, as
		// after Redis lost it. The process sweeping is alive regardless.
		registry := factory(t, ttl)
		podA, podB := newInstance("pod-a"), newInstance("pod-b")
		store := memory.New()
		save(t, store.WithOwner(podA), "a-running", running, time.Now())
		save(t, store.WithOwner(podB), "b-running", running, time.Now())
		sweeper := &invocation.StaleSweeper{Repo: store.WithOwner(podB), Liveness: registry, Self: podB, LegacyAge: invocation.LegacyStaleAge}
		result, err := sweeper.Sweep(t.Context())
		if err != nil {
			t.Fatalf("Sweep: %v", err)
		}
		if result.Failed != 1 {
			t.Fatalf("Sweep = %+v, want only A's run failed", result)
		}
		if got := status(t, store, "b-running").GetStatus(); got != running {
			t.Fatalf("the sweeping process failed its own run: %v", got)
		}
	})

	// Acceptance: a record without an owner, written before owners were
	// stamped, is failed only past the cutoff.
	t.Run("OwnerlessRecordsAreFailedOnlyPastTheCutoff", func(t *testing.T) {
		registry := factory(t, ttl)
		podB := newInstance("pod-b")
		keep(t, t.Context(), registry, podB)
		store := memory.New()
		now := time.Now()
		save(t, store, "legacy-old", running, now.Add(-invocation.LegacyStaleAge-time.Hour))
		save(t, store, "legacy-young", running, now.Add(-time.Hour))
		sweeper := &invocation.StaleSweeper{Repo: store.WithOwner(podB), Liveness: registry, Self: podB, LegacyAge: invocation.LegacyStaleAge}
		result, err := sweeper.Sweep(t.Context())
		if err != nil {
			t.Fatalf("Sweep: %v", err)
		}
		if result.Failed != 1 || len(result.LostOwners) != 0 {
			t.Fatalf("Sweep = %+v, want only the record past the cutoff failed", result)
		}
		old := status(t, store, "legacy-old")
		if old.GetStatus() != failed || strings.Contains(old.GetError(), "instance") {
			t.Fatalf("legacy-old = %v / %q, want FAILED with the reason for a record without an owner", old.GetStatus(), old.GetError())
		}
		if got := status(t, store, "legacy-young").GetStatus(); got != running {
			t.Fatalf("legacy-young is %v, want RUNNING until the cutoff", got)
		}
	})
}

type unreachable struct{}

func (unreachable) Alive(context.Context, []string) (map[string]bool, error) {
	return nil, errors.New("redis: connection refused")
}

// Not knowing whether an owner is alive is not knowing that it is gone.
func TestStaleSweepFailsNothingWhenLivenessCannotBeRead(t *testing.T) {
	store := memory.New()
	save(t, store.WithOwner("pod-a"), "a-running", running, time.Now())
	save(t, store, "legacy-old", running, time.Now().Add(-48*time.Hour))
	sweeper := &invocation.StaleSweeper{Repo: store.WithOwner("pod-b"), Liveness: unreachable{}, Self: "pod-b", LegacyAge: invocation.LegacyStaleAge}
	if _, err := sweeper.Sweep(t.Context()); err == nil {
		t.Fatal("Sweep succeeded without an answer from the liveness registry")
	}
	for _, id := range []string{"a-running", "legacy-old"} {
		if got := status(t, store, id).GetStatus(); got != running {
			t.Fatalf("%s is %v, want RUNNING: a sweep without an answer fails nothing", id, got)
		}
	}
}

func TestStaleSweepNeedsARepositoryAndARegistry(t *testing.T) {
	if _, err := (&invocation.StaleSweeper{Repo: memory.New()}).Sweep(t.Context()); err == nil {
		t.Fatal("Sweep without a liveness registry succeeded")
	}
}
