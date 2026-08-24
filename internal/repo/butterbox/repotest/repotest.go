// Package repotest is a conformance suite every butterbox.Repository
// implementation must pass.
package repotest

import (
	"errors"
	"testing"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Factory builds a fresh, empty repository per test.
type Factory func(t *testing.T) butterboxrepo.Repository

func box(id, name string) *agentsv1.ButterBox {
	return &agentsv1.ButterBox{Id: id, Name: name, BaseUrl: "https://" + name + ".example.com", Enabled: true}
}

// Run exercises the conformance suite against the factory's repository.
func Run(t *testing.T, factory Factory) {
	t.Run("CRUDAndWorkspaceIsolation", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()

		created, err := repo.Create(ctx, "ws1", box("b1", "alpha"), butterboxrepo.Credential{})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if created.GetCreatedAt() == nil || created.GetUpdatedAt() == nil {
			t.Fatal("timestamps not stamped")
		}
		if created.GetWorkspaceId() != "ws1" {
			t.Fatalf("workspace_id = %q", created.GetWorkspaceId())
		}
		if created.GetCredentialSet() {
			t.Fatal("credential_set true without credential")
		}

		if _, err := repo.Get(ctx, "ws2", "b1"); !errors.Is(err, butterboxrepo.ErrNotFound) {
			t.Fatalf("cross-workspace Get = %v, want ErrNotFound", err)
		}
		if err := repo.Delete(ctx, "ws2", "b1"); !errors.Is(err, butterboxrepo.ErrNotFound) {
			t.Fatalf("cross-workspace Delete = %v, want ErrNotFound", err)
		}

		updated := box("b1", "alpha-renamed")
		got, err := repo.Update(ctx, "ws1", updated)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if got.GetName() != "alpha-renamed" {
			t.Fatalf("name = %q", got.GetName())
		}
		if got.GetCreatedAt().AsTime() != created.GetCreatedAt().AsTime() {
			t.Fatal("Update changed created_at")
		}

		list, err := repo.List(ctx, "ws1")
		if err != nil || len(list) != 1 {
			t.Fatalf("List = %v, %v", list, err)
		}
		otherWs, err := repo.List(ctx, "ws2")
		if err != nil || len(otherWs) != 0 {
			t.Fatalf("List other workspace = %v, %v", otherWs, err)
		}

		if err := repo.Delete(ctx, "ws1", "b1"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.Get(ctx, "ws1", "b1"); !errors.Is(err, butterboxrepo.ErrNotFound) {
			t.Fatalf("Get after delete = %v, want ErrNotFound", err)
		}
	})

	t.Run("NameUniquePerWorkspace", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()

		if _, err := repo.Create(ctx, "ws1", box("b1", "alpha"), butterboxrepo.Credential{}); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if _, err := repo.Create(ctx, "ws1", box("b2", "alpha"), butterboxrepo.Credential{}); !errors.Is(err, butterboxrepo.ErrAlreadyExists) {
			t.Fatalf("duplicate name = %v, want ErrAlreadyExists", err)
		}
		// The same name in another workspace is fine.
		if _, err := repo.Create(ctx, "ws2", box("b3", "alpha"), butterboxrepo.Credential{}); err != nil {
			t.Fatalf("same name other workspace: %v", err)
		}
		// Renaming onto an existing name is also rejected.
		if _, err := repo.Create(ctx, "ws1", box("b4", "beta"), butterboxrepo.Credential{}); err != nil {
			t.Fatalf("Create beta: %v", err)
		}
		if _, err := repo.Update(ctx, "ws1", box("b4", "alpha")); !errors.Is(err, butterboxrepo.ErrAlreadyExists) {
			t.Fatalf("rename onto duplicate = %v, want ErrAlreadyExists", err)
		}
	})

	t.Run("CredentialLifecycle", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()

		// Atomic create with credential.
		created, err := repo.Create(ctx, "ws1", box("b1", "alpha"), butterboxrepo.Credential{Ciphertext: "ct1", KeyID: "k1"})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if !created.GetCredentialSet() || created.GetCredentialUpdatedAt() == nil {
			t.Fatalf("derived credential fields not set: %+v", created)
		}
		cred, err := repo.GetCredential(ctx, "ws1", "b1")
		if err != nil || cred.Ciphertext != "ct1" || cred.KeyID != "k1" {
			t.Fatalf("GetCredential = %+v, %v", cred, err)
		}

		// Update must not touch the credential.
		if _, err := repo.Update(ctx, "ws1", box("b1", "alpha-2")); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if cred, err = repo.GetCredential(ctx, "ws1", "b1"); err != nil || cred.Ciphertext != "ct1" {
			t.Fatalf("credential after Update = %+v, %v", cred, err)
		}

		// Rotate. Rotation stamps credential_updated_at but, like the
		// telegram credential seam, leaves the resource's updated_at alone.
		beforeRotate, err := repo.Get(ctx, "ws1", "b1")
		if err != nil {
			t.Fatalf("Get before rotate: %v", err)
		}
		rotated, err := repo.SetCredential(ctx, "ws1", "b1", butterboxrepo.Credential{Ciphertext: "ct2", KeyID: "k2"})
		if err != nil || !rotated.GetCredentialSet() {
			t.Fatalf("SetCredential = %+v, %v", rotated, err)
		}
		if rotated.GetUpdatedAt().AsTime() != beforeRotate.GetUpdatedAt().AsTime() {
			t.Fatal("SetCredential changed the resource updated_at")
		}
		if cred, err = repo.GetCredential(ctx, "ws1", "b1"); err != nil || cred.Ciphertext != "ct2" || cred.KeyID != "k2" {
			t.Fatalf("credential after rotate = %+v, %v", cred, err)
		}

		// Clear.
		cleared, err := repo.SetCredential(ctx, "ws1", "b1", butterboxrepo.Credential{})
		if err != nil {
			t.Fatalf("clear credential: %v", err)
		}
		if cleared.GetCredentialSet() || cleared.GetCredentialUpdatedAt() != nil {
			t.Fatalf("derived fields after clear: %+v", cleared)
		}
		if _, err := repo.GetCredential(ctx, "ws1", "b1"); !errors.Is(err, butterboxrepo.ErrNoCredential) {
			t.Fatalf("GetCredential after clear = %v, want ErrNoCredential", err)
		}

		// Cross-workspace credential access is not found.
		if _, err := repo.GetCredential(ctx, "ws2", "b1"); !errors.Is(err, butterboxrepo.ErrNotFound) {
			t.Fatalf("cross-workspace GetCredential = %v, want ErrNotFound", err)
		}
	})

	t.Run("DerivedFieldsIgnoredOnWrite", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()

		lying := box("b1", "alpha")
		lying.CredentialSet = true
		lying.WorkspaceId = "ws-spoofed"
		created, err := repo.Create(ctx, "ws1", lying, butterboxrepo.Credential{})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if created.GetCredentialSet() {
			t.Fatal("client-supplied credential_set leaked through Create")
		}
		if created.GetWorkspaceId() != "ws1" {
			t.Fatalf("workspace_id = %q, want stamped ws1", created.GetWorkspaceId())
		}
	})
}
