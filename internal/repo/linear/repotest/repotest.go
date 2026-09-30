// Package repotest is a conformance suite every linear.Repository
// implementation must pass.
package repotest

import (
	"errors"
	"testing"

	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Factory builds a fresh, empty repository per test.
type Factory func(t *testing.T) linearrepo.Repository

func app(id, clientID, name string) *agentsv1.LinearApp {
	return &agentsv1.LinearApp{Id: id, ClientId: clientID, DisplayName: name, AgentId: "support"}
}

func cred(v string) linearrepo.Credential {
	return linearrepo.Credential{Ciphertext: "cipher-" + v, KeyID: "key-1"}
}

// Run exercises the conformance suite against the factory's repository.
func Run(t *testing.T, factory Factory) {
	t.Run("AppCRUDAndWorkspaceIsolation", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()

		created, err := repo.CreateApp(ctx, "ws-a", app("app-1", "client-1", "Support"), linearrepo.AppCredentials{})
		if err != nil {
			t.Fatalf("CreateApp: %v", err)
		}
		if created.GetWorkspaceId() != "ws-a" || created.GetRevision() != 1 {
			t.Fatalf("created = %+v; want workspace ws-a, revision 1", created)
		}
		if created.GetCreatedAt() == nil || created.GetUpdatedAt() == nil {
			t.Fatal("timestamps not stamped")
		}
		if created.GetCredentialState() != agentsv1.LinearAppCredentialState_LINEAR_APP_CREDENTIAL_STATE_INCOMPLETE {
			t.Fatalf("credential_state = %v without secrets", created.GetCredentialState())
		}

		if _, err := repo.GetApp(ctx, "ws-b", "app-1"); !errors.Is(err, linearrepo.ErrNotFound) {
			t.Fatalf("cross-workspace GetApp = %v, want ErrNotFound", err)
		}
		if err := repo.DeleteApp(ctx, "ws-b", "app-1"); !errors.Is(err, linearrepo.ErrNotFound) {
			t.Fatalf("cross-workspace DeleteApp = %v, want ErrNotFound", err)
		}
		found, err := repo.FindApp(ctx, "app-1")
		if err != nil || found.GetWorkspaceId() != "ws-a" {
			t.Fatalf("FindApp = %+v, %v; want the ws-a App", found, err)
		}

		list, err := repo.ListApps(ctx, "ws-a")
		if err != nil || len(list) != 1 {
			t.Fatalf("ListApps = %d apps, %v; want 1", len(list), err)
		}
		if other, _ := repo.ListApps(ctx, "ws-b"); len(other) != 0 {
			t.Fatalf("ListApps(ws-b) = %d apps, want 0", len(other))
		}

		if err := repo.DeleteApp(ctx, "ws-a", "app-1"); err != nil {
			t.Fatalf("DeleteApp: %v", err)
		}
		if _, err := repo.FindApp(ctx, "app-1"); !errors.Is(err, linearrepo.ErrNotFound) {
			t.Fatalf("FindApp after delete = %v, want ErrNotFound", err)
		}
	})

	t.Run("ClientIDIsGloballyUnique", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		if _, err := repo.CreateApp(ctx, "ws-a", app("app-1", "client-1", "A"), linearrepo.AppCredentials{}); err != nil {
			t.Fatalf("CreateApp: %v", err)
		}
		if _, err := repo.CreateApp(ctx, "ws-b", app("app-2", "client-1", "B"), linearrepo.AppCredentials{}); !errors.Is(err, linearrepo.ErrClientIDExists) {
			t.Fatalf("second CreateApp with the same client ID = %v, want ErrClientIDExists", err)
		}
	})

	t.Run("UpdateIsGuardedByRevisionAndKeepsIdentity", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		created, err := repo.CreateApp(ctx, "ws-a", app("app-1", "client-1", "Support"), linearrepo.AppCredentials{})
		if err != nil {
			t.Fatalf("CreateApp: %v", err)
		}
		change := app("app-1", "client-other", "Renamed")
		updated, err := repo.UpdateApp(ctx, "ws-a", change, created.GetRevision())
		if err != nil {
			t.Fatalf("UpdateApp: %v", err)
		}
		if updated.GetDisplayName() != "Renamed" || updated.GetClientId() != "client-1" {
			t.Fatalf("updated = %+v; want the new name and the original client ID", updated)
		}
		if updated.GetRevision() != created.GetRevision()+1 {
			t.Fatalf("revision = %d, want %d", updated.GetRevision(), created.GetRevision()+1)
		}
		if !updated.GetCreatedAt().AsTime().Equal(created.GetCreatedAt().AsTime()) {
			t.Fatal("UpdateApp changed created_at")
		}
		if _, err := repo.UpdateApp(ctx, "ws-a", change, created.GetRevision()); !errors.Is(err, linearrepo.ErrRevisionConflict) {
			t.Fatalf("stale UpdateApp = %v, want ErrRevisionConflict", err)
		}
	})

	t.Run("CredentialsAreKeptOutOfTheModel", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		created, err := repo.CreateApp(ctx, "ws-a", app("app-1", "client-1", "Support"),
			linearrepo.AppCredentials{ClientSecret: cred("client")})
		if err != nil {
			t.Fatalf("CreateApp: %v", err)
		}
		if !created.GetClientSecretSet() || created.GetWebhookSecretSet() {
			t.Fatalf("created secrets set = %v/%v; want client only", created.GetClientSecretSet(), created.GetWebhookSecretSet())
		}

		webhook := cred("webhook")
		both, err := repo.SetAppCredentials(ctx, "ws-a", "app-1", linearrepo.CredentialChange{WebhookSecret: &webhook})
		if err != nil {
			t.Fatalf("SetAppCredentials: %v", err)
		}
		if both.GetCredentialState() != agentsv1.LinearAppCredentialState_LINEAR_APP_CREDENTIAL_STATE_COMPLETE {
			t.Fatalf("credential_state = %v with both secrets", both.GetCredentialState())
		}
		stored, err := repo.GetAppCredentials(ctx, "ws-a", "app-1")
		if err != nil {
			t.Fatalf("GetAppCredentials: %v", err)
		}
		if stored.ClientSecret != cred("client") || stored.WebhookSecret != webhook {
			t.Fatalf("stored = %+v; the untouched secret must be kept", stored)
		}

		cleared, err := repo.SetAppCredentials(ctx, "ws-a", "app-1", linearrepo.CredentialChange{ClientSecret: &linearrepo.Credential{}})
		if err != nil {
			t.Fatalf("clear client secret: %v", err)
		}
		if cleared.GetClientSecretSet() || !cleared.GetWebhookSecretSet() {
			t.Fatalf("after clearing = %v/%v; want webhook only", cleared.GetClientSecretSet(), cleared.GetWebhookSecretSet())
		}
		if _, err := repo.GetAppCredentials(ctx, "ws-b", "app-1"); !errors.Is(err, linearrepo.ErrNotFound) {
			t.Fatalf("cross-workspace GetAppCredentials = %v, want ErrNotFound", err)
		}
	})
}
