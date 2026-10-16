// Package repotest is a conformance suite every invocation.Repository
// implementation must pass. It covers the owner stamp and the stale
// selection (#390); the rest of the repository predates it.
package repotest

import (
	"slices"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/invocation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Open opens one store as the process with the given instance ID, the way
// several processes share one database. The empty owner opens it the way
// records were written before owner stamps existed.
type Open func(owner string) invocation.Repository

// Factory builds a fresh, empty store per test.
type Factory func(t *testing.T) Open

const (
	running = agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING
	queued  = agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED
	failed  = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
	done    = agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED
)

func record(id string, status agentsv1.InvocationStatus, started time.Time) *agentsv1.Invocation {
	inv := &agentsv1.Invocation{
		Id:          id,
		WorkspaceId: "ws-1",
		SessionId:   "session-" + id,
		AgentName:   "agent",
		Status:      status,
		Input:       "input of " + id,
	}
	if !started.IsZero() {
		inv.StartedAt = timestamppb.New(started)
	}
	return inv
}

func save(t *testing.T, repo invocation.Repository, inv *agentsv1.Invocation) {
	t.Helper()
	if err := repo.Save(t.Context(), inv); err != nil {
		t.Fatalf("Save(%s): %v", inv.GetId(), err)
	}
}

func get(t *testing.T, repo invocation.Repository, id string) *agentsv1.Invocation {
	t.Helper()
	inv, err := repo.GetAcrossWorkspaces(t.Context(), id)
	if err != nil {
		t.Fatalf("Get(%s): %v", id, err)
	}
	return inv
}

func owners(t *testing.T, repo invocation.Repository) []string {
	t.Helper()
	got, err := repo.ActiveOwners(t.Context())
	if err != nil {
		t.Fatalf("ActiveOwners: %v", err)
	}
	return got
}

func mark(t *testing.T, repo invocation.Repository, sel invocation.StaleSelection) int64 {
	t.Helper()
	if sel.Reason == nil {
		sel.Reason = func(owner string) string { return "stale, owner=" + owner }
	}
	n, err := repo.MarkStaleRunning(t.Context(), sel)
	if err != nil {
		t.Fatalf("MarkStaleRunning: %v", err)
	}
	return n
}

func wantStatus(t *testing.T, repo invocation.Repository, id string, want agentsv1.InvocationStatus) *agentsv1.Invocation {
	t.Helper()
	inv := get(t, repo, id)
	if inv.GetStatus() != want {
		t.Fatalf("%s status = %v, want %v", id, inv.GetStatus(), want)
	}
	return inv
}

// Run exercises the conformance suite against the factory's store.
func Run(t *testing.T, factory Factory) {
	now := time.Now().UTC()

	t.Run("ARecordCarriesTheOwnerOfTheProcessThatCreatedIt", func(t *testing.T) {
		open := factory(t)
		save(t, open("pod-a"), record("a1", running, now))
		save(t, open("pod-b"), record("b1", queued, now))
		save(t, open(""), record("legacy", running, now))
		if got := owners(t, open("")); !slices.Equal(got, []string{"pod-a", "pod-b"}) {
			t.Fatalf("ActiveOwners = %v, want [pod-a pod-b]: one entry per owner, none for a record without a stamp", got)
		}
	})

	t.Run("LaterSavesFromAnyProcessKeepTheStamp", func(t *testing.T) {
		open := factory(t)
		a, b := open("pod-a"), open("pod-b")
		save(t, a, record("a1", running, now))
		// Another process updates the record and redacts it, as a session
		// delete does; the record still belongs to the process running it.
		update := record("a1", running, now)
		update.Output = "partial"
		save(t, b, update)
		if err := b.RedactContent(t.Context(), "ws-1", "a1"); err != nil {
			t.Fatalf("RedactContent: %v", err)
		}
		if got := owners(t, b); !slices.Equal(got, []string{"pod-a"}) {
			t.Fatalf("ActiveOwners = %v, want [pod-a]", got)
		}
		if n := mark(t, b, invocation.StaleSelection{LostOwners: []string{"pod-b"}}); n != 0 {
			t.Fatalf("failing pod-b's records failed %d of pod-a's", n)
		}
		if n := mark(t, b, invocation.StaleSelection{LostOwners: []string{"pod-a"}}); n != 1 {
			t.Fatalf("failing pod-a's records failed %d, want 1", n)
		}
		if inv := wantStatus(t, b, "a1", failed); inv.GetInput() != "" {
			t.Fatalf("failing the record restored redacted input %q", inv.GetInput())
		}
	})

	t.Run("OnlyQueuedAndRunningRecordsReportAnOwner", func(t *testing.T) {
		open := factory(t)
		a := open("pod-a")
		save(t, a, record("a1", running, now))
		save(t, a, record("a1", done, now))
		save(t, open("pod-b"), record("b1", failed, now))
		if got := owners(t, a); len(got) != 0 {
			t.Fatalf("ActiveOwners = %v, want none once every record is terminal", got)
		}
	})

	// A process starting fails nothing by itself: only a lost owner, or an
	// owner-less record past the cutoff, is selected.
	t.Run("AnEmptySelectionFailsNothing", func(t *testing.T) {
		open := factory(t)
		save(t, open("pod-a"), record("a1", running, now))
		save(t, open(""), record("legacy", queued, now.Add(-48*time.Hour)))
		if n := mark(t, open("pod-b"), invocation.StaleSelection{}); n != 0 {
			t.Fatalf("an empty selection failed %d records", n)
		}
		wantStatus(t, open(""), "a1", running)
		wantStatus(t, open(""), "legacy", queued)
	})

	t.Run("OnlyTheLostOwnersQueuedAndRunningRecordsAreFailed", func(t *testing.T) {
		open := factory(t)
		a, b := open("pod-a"), open("pod-b")
		save(t, a, record("a-running", running, now))
		save(t, a, record("a-queued", queued, now))
		save(t, a, record("a-done", done, now))
		other := record("a-other-workspace", running, now)
		other.WorkspaceId = "ws-2"
		save(t, a, other)
		save(t, b, record("b-running", running, now))

		n := mark(t, b, invocation.StaleSelection{LostOwners: []string{"pod-a"}})
		if n != 3 {
			t.Fatalf("MarkStaleRunning failed %d records, want pod-a's 3 active ones", n)
		}
		for _, id := range []string{"a-running", "a-queued", "a-other-workspace"} {
			inv := wantStatus(t, b, id, failed)
			if inv.GetError() != "stale, owner=pod-a" {
				t.Fatalf("%s error = %q, want the reason for pod-a", id, inv.GetError())
			}
			if inv.GetFinishedAt() == nil {
				t.Fatalf("%s has no finished_at", id)
			}
		}
		wantStatus(t, b, "a-done", done)
		wantStatus(t, b, "b-running", running)
		if active, err := b.FindActiveBySession(t.Context(), "ws-1", "session-b-running"); err != nil || active.GetId() != "b-running" {
			t.Fatalf("FindActiveBySession = %v, %v; want pod-b's run still active", active, err)
		}
		if got := owners(t, b); !slices.Equal(got, []string{"pod-b"}) {
			t.Fatalf("ActiveOwners after the sweep = %v, want [pod-b]", got)
		}
		if n := mark(t, b, invocation.StaleSelection{LostOwners: []string{"pod-a"}}); n != 0 {
			t.Fatalf("a second sweep failed %d records again", n)
		}
	})

	t.Run("OwnerlessRecordsAreFailedOnlyPastTheCutoff", func(t *testing.T) {
		open := factory(t)
		legacy := open("")
		save(t, legacy, record("old", running, now.Add(-48*time.Hour)))
		save(t, legacy, record("young", running, now.Add(-time.Hour)))
		save(t, legacy, record("no-start", queued, time.Time{}))
		// An owned record is never judged by age, however old.
		save(t, open("pod-a"), record("owned-old", running, now.Add(-48*time.Hour)))

		n := mark(t, legacy, invocation.StaleSelection{LegacyBefore: now.Add(-24 * time.Hour)})
		if n != 2 {
			t.Fatalf("MarkStaleRunning failed %d records, want the old one and the one with no start time", n)
		}
		for _, id := range []string{"old", "no-start"} {
			if inv := wantStatus(t, legacy, id, failed); inv.GetError() != "stale, owner=" {
				t.Fatalf("%s error = %q, want the reason for a record without an owner", id, inv.GetError())
			}
		}
		wantStatus(t, legacy, "young", running)
		wantStatus(t, legacy, "owned-old", running)
	})

	t.Run("BothRulesApplyInOneSweep", func(t *testing.T) {
		open := factory(t)
		save(t, open("pod-a"), record("lost", running, now))
		save(t, open("pod-b"), record("alive", running, now))
		save(t, open(""), record("legacy", running, now.Add(-48*time.Hour)))
		n := mark(t, open("pod-b"), invocation.StaleSelection{
			LostOwners:   []string{"pod-a"},
			LegacyBefore: now.Add(-24 * time.Hour),
		})
		if n != 2 {
			t.Fatalf("MarkStaleRunning failed %d records, want 2", n)
		}
		wantStatus(t, open(""), "lost", failed)
		wantStatus(t, open(""), "legacy", failed)
		wantStatus(t, open(""), "alive", running)
	})

	t.Run("ListNarrowsToAUsersRecordsAndAppPrefixes", func(t *testing.T) {
		open := factory(t)
		repo := open("pod-a")
		for i, r := range []struct{ id, user, app, ws string }{
			{"alice", "alice", "agui", "ws-1"},
			{"bob", "bob", "agui", "ws-1"},
			{"cron", "cron:nightly", "cron:nightly", "ws-1"},
			{"auto", "automation:daily", "automation:daily", "ws-1"},
			{"other-ws", "alice", "agui", "ws-2"},
		} {
			inv := record(r.id, done, now.Add(-time.Duration(i)*time.Minute))
			inv.UserId, inv.AppName, inv.WorkspaceId = r.user, r.app, r.ws
			save(t, repo, inv)
		}
		list := func(f invocation.ListFilter) []string {
			t.Helper()
			f.WorkspaceID = "ws-1"
			// Two a page, so the total and the paging answer for the
			// narrowed set, not the workspace.
			var ids []string
			token := ""
			for {
				page, next, total, err := repo.List(t.Context(), f, 2, token)
				if err != nil {
					t.Fatalf("List(%+v): %v", f, err)
				}
				for _, inv := range page {
					ids = append(ids, inv.GetId())
				}
				if total != int32(len(ids))+countAfter(t, repo, f, next) {
					t.Fatalf("List(%+v) total = %d after %v", f, total, ids)
				}
				if token = next; token == "" {
					break
				}
			}
			slices.Sort(ids)
			return ids
		}
		if got := list(invocation.ListFilter{}); !slices.Equal(got, []string{"alice", "auto", "bob", "cron"}) {
			t.Fatalf("an unnarrowed List = %v, want the workspace's four", got)
		}
		if got := list(invocation.ListFilter{UserID: "alice"}); !slices.Equal(got, []string{"alice"}) {
			t.Fatalf("List by user = %v, want [alice]", got)
		}
		if got := list(invocation.ListFilter{UserID: "alice", OrAppPrefixes: []string{"cron:", "automation:"}}); !slices.Equal(got, []string{"alice", "auto", "cron"}) {
			t.Fatalf("List by user or app prefix = %v, want [alice auto cron]", got)
		}
		if got := list(invocation.ListFilter{OrAppPrefixes: []string{"cron:"}}); !slices.Equal(got, []string{"cron"}) {
			t.Fatalf("List by app prefix alone = %v, want [cron]", got)
		}
	})
}

// countAfter is how many records the page token leaves unread.
func countAfter(t *testing.T, repo invocation.Repository, f invocation.ListFilter, token string) int32 {
	t.Helper()
	if token == "" {
		return 0
	}
	var n int32
	for {
		page, next, _, err := repo.List(t.Context(), f, 2, token)
		if err != nil {
			t.Fatalf("List(%+v): %v", f, err)
		}
		n += int32(len(page))
		if token = next; token == "" {
			return n
		}
	}
}
