// Package repotest is a conformance suite every plane.Repository
// implementation must pass.
package repotest

import (
	"errors"
	"testing"

	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Factory builds a fresh, empty repository per test.
type Factory func(t *testing.T) planerepo.Repository

func app(id, clientID, name string) *agentsv1.PlaneApp {
	return &agentsv1.PlaneApp{
		Id:              id,
		ProviderProfile: "cloud",
		ClientId:        clientID,
		DisplayName:     name,
		AgentId:         "support",
	}
}

func cred(v string) planerepo.Credential {
	return planerepo.Credential{Ciphertext: "cipher-" + v, KeyID: "key-1"}
}

// Run exercises the conformance suite against the factory's repository.
func Run(t *testing.T, factory Factory) {
	t.Run("AppCRUDAndWorkspaceIsolation", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()

		created, err := repo.CreateApp(ctx, "ws-a", app("app-1", "client-1", "Support"), planerepo.AppCredentials{})
		if err != nil {
			t.Fatalf("CreateApp: %v", err)
		}
		if created.GetWorkspaceId() != "ws-a" || created.GetRevision() != 1 {
			t.Fatalf("created = %+v; want workspace ws-a, revision 1", created)
		}
		if created.GetProviderProfile() != "cloud" || created.GetClientId() != "client-1" {
			t.Fatalf("created = %+v; want the stored provider profile and client ID", created)
		}
		if created.GetCreatedAt() == nil || created.GetUpdatedAt() == nil {
			t.Fatal("timestamps not stamped")
		}
		if created.GetCredentialState() != agentsv1.PlaneAppCredentialState_PLANE_APP_CREDENTIAL_STATE_INCOMPLETE {
			t.Fatalf("credential_state = %v without secrets", created.GetCredentialState())
		}

		if _, err := repo.GetApp(ctx, "ws-b", "app-1"); !errors.Is(err, planerepo.ErrNotFound) {
			t.Fatalf("cross-workspace GetApp = %v, want ErrNotFound", err)
		}
		if err := repo.DeleteApp(ctx, "ws-b", "app-1"); !errors.Is(err, planerepo.ErrNotFound) {
			t.Fatalf("cross-workspace DeleteApp = %v, want ErrNotFound", err)
		}
		if _, err := repo.UpdateApp(ctx, "ws-b", app("app-1", "client-1", "Hijacked"), 1); !errors.Is(err, planerepo.ErrNotFound) {
			t.Fatalf("cross-workspace UpdateApp = %v, want ErrNotFound", err)
		}
		if _, err := repo.SetAppCredentials(ctx, "ws-b", "app-1", planerepo.CredentialChange{ClientSecret: ptr(cred("x"))}); !errors.Is(err, planerepo.ErrNotFound) {
			t.Fatalf("cross-workspace SetAppCredentials = %v, want ErrNotFound", err)
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
		if _, err := repo.GetApp(ctx, "ws-a", "app-1"); !errors.Is(err, planerepo.ErrNotFound) {
			t.Fatalf("GetApp after delete = %v, want ErrNotFound", err)
		}
	})

	t.Run("ClientIDIsUniqueAcrossWorkspaces", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		if _, err := repo.CreateApp(ctx, "ws-a", app("app-1", "client-1", "A"), planerepo.AppCredentials{}); err != nil {
			t.Fatalf("CreateApp: %v", err)
		}
		if _, err := repo.CreateApp(ctx, "ws-b", app("app-2", "client-1", "B"), planerepo.AppCredentials{}); !errors.Is(err, planerepo.ErrClientIDExists) {
			t.Fatalf("second CreateApp with the same client ID = %v, want ErrClientIDExists", err)
		}
		// A different provider profile is a different Plane app namespace.
		other := app("app-3", "client-1", "C")
		other.ProviderProfile = "selfhosted"
		if _, err := repo.CreateApp(ctx, "ws-b", other, planerepo.AppCredentials{}); err != nil {
			t.Fatalf("CreateApp under another provider profile: %v", err)
		}
	})

	t.Run("UpdateIsGuardedByRevisionAndKeepsIdentity", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		created, err := repo.CreateApp(ctx, "ws-a", app("app-1", "client-1", "Support"), planerepo.AppCredentials{})
		if err != nil {
			t.Fatalf("CreateApp: %v", err)
		}
		change := app("app-1", "client-other", "Renamed")
		change.ProviderProfile = "selfhosted"
		change.AllowedUserIds = []string{"user-1"}
		updated, err := repo.UpdateApp(ctx, "ws-a", change, created.GetRevision())
		if err != nil {
			t.Fatalf("UpdateApp: %v", err)
		}
		if updated.GetDisplayName() != "Renamed" || len(updated.GetAllowedUserIds()) != 1 {
			t.Fatalf("updated = %+v; want the new name and allowed users", updated)
		}
		if updated.GetClientId() != "client-1" || updated.GetProviderProfile() != "cloud" {
			t.Fatalf("updated = %+v; identity and provider must be immutable", updated)
		}
		if updated.GetRevision() != created.GetRevision()+1 {
			t.Fatalf("revision = %d, want %d", updated.GetRevision(), created.GetRevision()+1)
		}
		if !updated.GetCreatedAt().AsTime().Equal(created.GetCreatedAt().AsTime()) {
			t.Fatal("UpdateApp changed created_at")
		}
		if updated.GetUpdatedAt().AsTime().Before(created.GetUpdatedAt().AsTime()) {
			t.Fatal("UpdateApp moved updated_at backwards")
		}
		if _, err := repo.UpdateApp(ctx, "ws-a", change, created.GetRevision()); !errors.Is(err, planerepo.ErrRevisionConflict) {
			t.Fatalf("stale UpdateApp = %v, want ErrRevisionConflict", err)
		}
	})

	t.Run("CredentialsAreKeptOutOfTheModel", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		created, err := repo.CreateApp(ctx, "ws-a", app("app-1", "client-1", "Support"),
			planerepo.AppCredentials{ClientSecret: cred("client")})
		if err != nil {
			t.Fatalf("CreateApp: %v", err)
		}
		if !created.GetClientSecretSet() || created.GetWebhookSecretSet() {
			t.Fatalf("created secrets set = %v/%v; want client only", created.GetClientSecretSet(), created.GetWebhookSecretSet())
		}
		if created.GetCredentialState() != agentsv1.PlaneAppCredentialState_PLANE_APP_CREDENTIAL_STATE_INCOMPLETE {
			t.Fatalf("credential_state = %v with one secret", created.GetCredentialState())
		}

		webhook := cred("webhook")
		both, err := repo.SetAppCredentials(ctx, "ws-a", "app-1", planerepo.CredentialChange{WebhookSecret: &webhook})
		if err != nil {
			t.Fatalf("SetAppCredentials: %v", err)
		}
		if both.GetCredentialState() != agentsv1.PlaneAppCredentialState_PLANE_APP_CREDENTIAL_STATE_COMPLETE {
			t.Fatalf("credential_state = %v with both secrets", both.GetCredentialState())
		}
		stored, err := repo.GetAppCredentials(ctx, "ws-a", "app-1")
		if err != nil {
			t.Fatalf("GetAppCredentials: %v", err)
		}
		if stored.ClientSecret != cred("client") || stored.WebhookSecret != webhook {
			t.Fatalf("stored = %+v; the untouched secret must be kept", stored)
		}

		cleared, err := repo.SetAppCredentials(ctx, "ws-a", "app-1", planerepo.CredentialChange{ClientSecret: &planerepo.Credential{}})
		if err != nil {
			t.Fatalf("clear client secret: %v", err)
		}
		if cleared.GetClientSecretSet() || !cleared.GetWebhookSecretSet() {
			t.Fatalf("after clearing = %v/%v; want webhook only", cleared.GetClientSecretSet(), cleared.GetWebhookSecretSet())
		}
		if _, err := repo.GetAppCredentials(ctx, "ws-b", "app-1"); !errors.Is(err, planerepo.ErrNotFound) {
			t.Fatalf("cross-workspace GetAppCredentials = %v, want ErrNotFound", err)
		}
		// Clearing one secret leaves the App's data model unchanged apart
		// from the derived flags.
		after, err := repo.GetApp(ctx, "ws-a", "app-1")
		if err != nil || after.GetDisplayName() != "Support" || after.GetRevision() != cleared.GetRevision() {
			t.Fatalf("GetApp after credential change = %+v, %v", after, err)
		}
	})
}

func ptr(c planerepo.Credential) *planerepo.Credential { return &c }
