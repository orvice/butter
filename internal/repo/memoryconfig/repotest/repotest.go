// Package repotest is a conformance suite every memoryconfig.Repository
// implementation must pass.
package repotest

import (
	"errors"
	"testing"

	memoryconfigrepo "go.orx.me/apps/butter/internal/repo/memoryconfig"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Factory builds a fresh, empty repository per test.
type Factory func(t *testing.T) memoryconfigrepo.Repository

func config(baseURL string, enabled bool) *agentsv1.WorkspaceMemoryConfig {
	return &agentsv1.WorkspaceMemoryConfig{BaseUrl: baseURL, Enabled: enabled}
}

func cred(ciphertext string) *memoryconfigrepo.Credential {
	return &memoryconfigrepo.Credential{Ciphertext: ciphertext, KeyID: "k1"}
}

// Run exercises the conformance suite against the factory's repository.
func Run(t *testing.T, factory Factory) {
	t.Run("MissingConfig", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		if _, err := repo.Get(ctx, "ws1"); !errors.Is(err, memoryconfigrepo.ErrNotFound) {
			t.Fatalf("Get = %v, want ErrNotFound", err)
		}
		if _, err := repo.GetCredential(ctx, "ws1"); !errors.Is(err, memoryconfigrepo.ErrNotFound) {
			t.Fatalf("GetCredential = %v, want ErrNotFound", err)
		}
		if err := repo.Delete(ctx, "ws1"); !errors.Is(err, memoryconfigrepo.ErrNotFound) {
			t.Fatalf("Delete = %v, want ErrNotFound", err)
		}
	})

	t.Run("PutIsAPerWorkspaceSingleton", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()

		created, err := repo.Put(ctx, "ws1", config("https://a.example.com", true), nil)
		if err != nil {
			t.Fatalf("Put: %v", err)
		}
		if created.GetWorkspaceId() != "ws1" || created.GetBaseUrl() != "https://a.example.com" || !created.GetEnabled() {
			t.Fatalf("created = %v", created)
		}
		if created.GetCreatedAt() == nil || created.GetUpdatedAt() == nil {
			t.Fatal("timestamps not stamped")
		}
		if created.GetCredentialSet() || created.GetCredentialUpdatedAt() != nil {
			t.Fatal("credential reported without one")
		}

		replaced, err := repo.Put(ctx, "ws1", config("https://b.example.com", false), nil)
		if err != nil {
			t.Fatalf("second Put: %v", err)
		}
		if replaced.GetBaseUrl() != "https://b.example.com" || replaced.GetEnabled() {
			t.Fatalf("replaced = %v", replaced)
		}
		if !replaced.GetCreatedAt().AsTime().Equal(created.GetCreatedAt().AsTime()) {
			t.Fatal("created_at changed on replace")
		}

		if _, err := repo.Get(ctx, "ws2"); !errors.Is(err, memoryconfigrepo.ErrNotFound) {
			t.Fatalf("other workspace Get = %v, want ErrNotFound", err)
		}
		if err := repo.Delete(ctx, "ws2"); !errors.Is(err, memoryconfigrepo.ErrNotFound) {
			t.Fatalf("other workspace Delete = %v, want ErrNotFound", err)
		}
	})

	t.Run("CredentialKeepSetRotateClear", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()

		got, err := repo.Put(ctx, "ws1", config("https://a.example.com", true), cred("c1"))
		if err != nil {
			t.Fatalf("Put with credential: %v", err)
		}
		if !got.GetCredentialSet() || got.GetCredentialUpdatedAt() == nil {
			t.Fatalf("credential not reported: %v", got)
		}
		stored, err := repo.GetCredential(ctx, "ws1")
		if err != nil || stored.Ciphertext != "c1" || stored.KeyID != "k1" {
			t.Fatalf("GetCredential = %+v, %v", stored, err)
		}

		// nil keeps the stored credential.
		if got, err = repo.Put(ctx, "ws1", config("https://b.example.com", true), nil); err != nil || !got.GetCredentialSet() {
			t.Fatalf("keep Put = %v, %v", got, err)
		}
		if stored, _ = repo.GetCredential(ctx, "ws1"); stored.Ciphertext != "c1" {
			t.Fatalf("credential not kept: %+v", stored)
		}

		// A set credential rotates.
		if _, err = repo.Put(ctx, "ws1", config("https://b.example.com", true), cred("c2")); err != nil {
			t.Fatalf("rotate Put: %v", err)
		}
		if stored, _ = repo.GetCredential(ctx, "ws1"); stored.Ciphertext != "c2" {
			t.Fatalf("credential not rotated: %+v", stored)
		}

		// An unset credential clears.
		got, err = repo.Put(ctx, "ws1", config("https://b.example.com", true), &memoryconfigrepo.Credential{})
		if err != nil {
			t.Fatalf("clear Put: %v", err)
		}
		if got.GetCredentialSet() || got.GetCredentialUpdatedAt() != nil {
			t.Fatalf("credential still reported: %v", got)
		}
		if _, err := repo.GetCredential(ctx, "ws1"); !errors.Is(err, memoryconfigrepo.ErrNoCredential) {
			t.Fatalf("GetCredential after clear = %v, want ErrNoCredential", err)
		}
	})

	t.Run("DeleteRemovesConfigAndCredential", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()

		if _, err := repo.Put(ctx, "ws1", config("https://a.example.com", true), cred("c1")); err != nil {
			t.Fatalf("Put: %v", err)
		}
		if err := repo.Delete(ctx, "ws1"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.Get(ctx, "ws1"); !errors.Is(err, memoryconfigrepo.ErrNotFound) {
			t.Fatalf("Get after delete = %v, want ErrNotFound", err)
		}

		// Recreating does not resurrect the old credential.
		got, err := repo.Put(ctx, "ws1", config("https://a.example.com", true), nil)
		if err != nil {
			t.Fatalf("recreate Put: %v", err)
		}
		if got.GetCredentialSet() {
			t.Fatal("credential survived delete")
		}
	})
}
