// Package repotest is a conformance suite for repobinding.Repository
// implementations: the memory and mongo backends must behave identically
// through this seam (issue #214).
package repotest

import (
	"context"
	"errors"
	"testing"

	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Factory returns a fresh, empty repository for one subtest.
type Factory func(t *testing.T) repobindingrepo.Repository

func newBinding() *agentsv1.WorkspaceRepoBinding {
	return &agentsv1.WorkspaceRepoBinding{
		GitHostId:            "gh-1",
		Repository:           "acme/agents",
		Branch:               "main",
		RootPath:             "butter",
		WriteMode:            agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_DIRECT_COMMIT,
		ContentSchemaVersion: 1,
		Status: &agentsv1.RepoBindingStatus{
			State: agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_UNVALIDATED,
		},
	}
}

func put(t *testing.T, repo repobindingrepo.Repository, ws string) *agentsv1.WorkspaceRepoBinding {
	t.Helper()
	stored, err := repo.Put(context.Background(), ws, newBinding())
	if err != nil {
		t.Fatalf("Put %s: %v", ws, err)
	}
	return stored
}

// Run executes the conformance suite against the given repository factory.
func Run(t *testing.T, factory Factory) {
	ctx := context.Background()

	t.Run("PutThenGetRoundTrip", func(t *testing.T) {
		repo := factory(t)
		stored := put(t, repo, "ws-a")
		if stored.GetWorkspaceId() != "ws-a" {
			t.Fatalf("expected workspace stamped, got %q", stored.GetWorkspaceId())
		}
		if stored.GetCreatedAt() == nil || stored.GetUpdatedAt() == nil {
			t.Fatalf("expected timestamps set, got %v / %v", stored.GetCreatedAt(), stored.GetUpdatedAt())
		}
		got, err := repo.Get(ctx, "ws-a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.GetGitHostId() != "gh-1" || got.GetRepository() != "acme/agents" ||
			got.GetBranch() != "main" || got.GetRootPath() != "butter" {
			t.Fatalf("binding did not round-trip: %v", got)
		}
		if got.GetWriteMode() != agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_DIRECT_COMMIT {
			t.Fatalf("write mode did not round-trip: %v", got.GetWriteMode())
		}
		if got.GetContentSchemaVersion() != 1 {
			t.Fatalf("schema version did not round-trip: %v", got.GetContentSchemaVersion())
		}
		if got.GetStatus().GetState() != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_UNVALIDATED {
			t.Fatalf("status did not round-trip: %v", got.GetStatus())
		}
		if got.GetCredentialSet() {
			t.Fatal("fresh binding must not report a credential")
		}
	})

	t.Run("GetMissingIsNotFound", func(t *testing.T) {
		repo := factory(t)
		if _, err := repo.Get(ctx, "ws-none"); !errors.Is(err, repobindingrepo.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("PutUpsertsKeepingCreatedAt", func(t *testing.T) {
		repo := factory(t)
		first := put(t, repo, "ws-a")
		mod := newBinding()
		mod.Branch = "develop"
		second, err := repo.Put(ctx, "ws-a", mod)
		if err != nil {
			t.Fatalf("second Put: %v", err)
		}
		if second.GetBranch() != "develop" {
			t.Fatalf("upsert did not apply: %v", second)
		}
		if !second.GetCreatedAt().AsTime().Equal(first.GetCreatedAt().AsTime()) {
			t.Fatalf("created_at changed on upsert: %v -> %v", first.GetCreatedAt(), second.GetCreatedAt())
		}
		all, err := repo.ListAcrossWorkspaces(ctx)
		if err != nil {
			t.Fatalf("ListAcrossWorkspaces: %v", err)
		}
		if len(all) != 1 {
			t.Fatalf("workspace must hold at most one binding, got %d", len(all))
		}
	})

	t.Run("CredentialLifecycle", func(t *testing.T) {
		repo := factory(t)
		if err := repo.SetCredential(ctx, "ws-a", "ct-1"); !errors.Is(err, repobindingrepo.ErrNotFound) {
			t.Fatalf("SetCredential without binding: err = %v, want ErrNotFound", err)
		}
		put(t, repo, "ws-a")
		if _, err := repo.GetCredential(ctx, "ws-a"); !errors.Is(err, repobindingrepo.ErrNoCredential) {
			t.Fatalf("GetCredential before set: err = %v, want ErrNoCredential", err)
		}
		if err := repo.SetCredential(ctx, "ws-a", "ct-1"); err != nil {
			t.Fatalf("SetCredential: %v", err)
		}
		ct, err := repo.GetCredential(ctx, "ws-a")
		if err != nil {
			t.Fatalf("GetCredential: %v", err)
		}
		if ct != "ct-1" {
			t.Fatalf("ciphertext did not round-trip: %q", ct)
		}
		got, err := repo.Get(ctx, "ws-a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if !got.GetCredentialSet() || got.GetCredentialUpdatedAt() == nil {
			t.Fatalf("credential fields not derived: set=%v at=%v", got.GetCredentialSet(), got.GetCredentialUpdatedAt())
		}
		// Replace.
		if err := repo.SetCredential(ctx, "ws-a", "ct-2"); err != nil {
			t.Fatalf("replace credential: %v", err)
		}
		if ct, _ := repo.GetCredential(ctx, "ws-a"); ct != "ct-2" {
			t.Fatalf("replacement did not apply: %q", ct)
		}
	})

	t.Run("PutPreservesStoredCredential", func(t *testing.T) {
		repo := factory(t)
		put(t, repo, "ws-a")
		if err := repo.SetCredential(ctx, "ws-a", "ct-1"); err != nil {
			t.Fatalf("SetCredential: %v", err)
		}
		mod := newBinding()
		mod.Branch = "develop"
		stored, err := repo.Put(ctx, "ws-a", mod)
		if err != nil {
			t.Fatalf("Put: %v", err)
		}
		if !stored.GetCredentialSet() {
			t.Fatal("Put dropped the stored credential flag")
		}
		if ct, err := repo.GetCredential(ctx, "ws-a"); err != nil || ct != "ct-1" {
			t.Fatalf("Put dropped the stored credential: %q, %v", ct, err)
		}
	})

	t.Run("PutCannotForgeCredentialFields", func(t *testing.T) {
		repo := factory(t)
		forged := newBinding()
		forged.CredentialSet = true
		stored, err := repo.Put(ctx, "ws-a", forged)
		if err != nil {
			t.Fatalf("Put: %v", err)
		}
		if stored.GetCredentialSet() {
			t.Fatal("credential_set must be derived from stored state, not input")
		}
	})

	t.Run("DeleteRemovesBindingAndCredential", func(t *testing.T) {
		repo := factory(t)
		put(t, repo, "ws-a")
		if err := repo.SetCredential(ctx, "ws-a", "ct-1"); err != nil {
			t.Fatalf("SetCredential: %v", err)
		}
		if err := repo.Delete(ctx, "ws-a"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.Get(ctx, "ws-a"); !errors.Is(err, repobindingrepo.ErrNotFound) {
			t.Fatalf("Get after delete: err = %v, want ErrNotFound", err)
		}
		if _, err := repo.GetCredential(ctx, "ws-a"); !errors.Is(err, repobindingrepo.ErrNotFound) {
			t.Fatalf("GetCredential after delete: err = %v, want ErrNotFound", err)
		}
		// Re-creating the binding must not resurrect the old credential.
		put(t, repo, "ws-a")
		if _, err := repo.GetCredential(ctx, "ws-a"); !errors.Is(err, repobindingrepo.ErrNoCredential) {
			t.Fatalf("credential resurrected after delete+put: err = %v, want ErrNoCredential", err)
		}
		if err := repo.Delete(ctx, "ws-none"); !errors.Is(err, repobindingrepo.ErrNotFound) {
			t.Fatalf("Delete missing: err = %v, want ErrNotFound", err)
		}
	})

	t.Run("WorkspaceIsolationAndOverlapListing", func(t *testing.T) {
		repo := factory(t)
		put(t, repo, "ws-a")
		put(t, repo, "ws-b") // same location — overlap is allowed
		if _, err := repo.Get(ctx, "ws-c"); !errors.Is(err, repobindingrepo.ErrNotFound) {
			t.Fatalf("Get ws-c: err = %v, want ErrNotFound", err)
		}
		all, err := repo.ListAcrossWorkspaces(ctx)
		if err != nil {
			t.Fatalf("ListAcrossWorkspaces: %v", err)
		}
		if len(all) != 2 {
			t.Fatalf("expected 2 bindings, got %d", len(all))
		}
		seen := map[string]bool{}
		for _, b := range all {
			seen[b.GetWorkspaceId()] = true
		}
		if !seen["ws-a"] || !seen["ws-b"] {
			t.Fatalf("workspace ids not stamped in flat view: %v", seen)
		}
	})

	t.Run("MutatingResultDoesNotAffectStore", func(t *testing.T) {
		repo := factory(t)
		stored := put(t, repo, "ws-a")
		stored.Branch = "mutated"
		got, err := repo.Get(ctx, "ws-a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.GetBranch() != "main" {
			t.Fatalf("store was mutated through returned pointer: %v", got)
		}
	})
}
