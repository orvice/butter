package cursorbox

import (
	"context"
	"strings"
	"testing"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	butterboxmemory "go.orx.me/apps/butter/internal/repo/butterbox/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestFactory_ClientForUnknownBox(t *testing.T) {
	f := NewFactory(butterboxmemory.New(), secretbox.NewKeyring(cryptokeymemory.New()))
	_, err := f.ClientFor(context.Background(), "ws-1", "missing-box")
	if err == nil || !strings.Contains(err.Error(), "no longer exists") {
		t.Fatalf("expected pointed not-found error, got %v", err)
	}
}

func TestFactory_ClientForReturnsClient(t *testing.T) {
	ctx := context.Background()
	repo := butterboxmemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	ciphertext, keyID, err := keyring.Encrypt(ctx, []byte("box-token"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	_, err = repo.Create(ctx, "ws-1", &agentsv1.ButterBox{
		Id: "box-1", Name: "dev box", BaseUrl: "http://localhost:9999", Enabled: true,
	}, butterboxrepo.Credential{Ciphertext: ciphertext, KeyID: keyID})
	if err != nil {
		t.Fatalf("create box: %v", err)
	}

	client, err := NewFactory(repo, keyring).ClientFor(ctx, "ws-1", "box-1")
	if err != nil {
		t.Fatalf("ClientFor: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestFactory_NilRepoReturnsError(t *testing.T) {
	f := &Factory{}
	_, err := f.ClientFor(context.Background(), "ws-1", "box-1")
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected config error, got %v", err)
	}
}
