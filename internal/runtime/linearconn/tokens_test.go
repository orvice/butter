package linearconn

import (
	"errors"
	"testing"

	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	linearmemory "go.orx.me/apps/butter/internal/repo/linear/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestAccessTokenIsTheInstallationsDecryptedToken(t *testing.T) {
	ctx := t.Context()
	repo := linearmemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	if _, err := repo.CreateApp(ctx, "ws-a", &agentsv1.LinearApp{Id: "app-1", ClientId: "c", AgentId: "a"}, linearrepo.AppCredentials{}); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	ciphertext, keyID, err := keyring.Encrypt(ctx, []byte("access-1"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := repo.UpsertInstallation(ctx, "ws-a", &agentsv1.LinearInstallation{Id: "inst-1", AppId: "app-1", OrganizationId: "org-1"},
		linearrepo.InstallationTokens{AccessToken: linearrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}
	if _, err := repo.UpsertInstallation(ctx, "ws-a", &agentsv1.LinearInstallation{Id: "inst-2", AppId: "app-1", OrganizationId: "org-2"},
		linearrepo.InstallationTokens{}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}

	source := NewTokenSource(repo, keyring)
	token, err := source.AccessToken(ctx, "ws-a", "inst-1")
	if err != nil || token != "access-1" {
		t.Fatalf("AccessToken = %q, %v; want the decrypted token", token, err)
	}
	if _, err := source.AccessToken(ctx, "ws-a", "inst-2"); !errors.Is(err, ErrNoToken) {
		t.Fatalf("AccessToken without a token = %v, want ErrNoToken", err)
	}
	if _, err := source.AccessToken(ctx, "ws-b", "inst-1"); !errors.Is(err, linearrepo.ErrNotFound) {
		t.Fatalf("cross-workspace AccessToken = %v, want ErrNotFound", err)
	}
}
