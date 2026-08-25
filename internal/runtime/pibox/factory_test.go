package pibox

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	butterboxmemory "go.orx.me/apps/butter/internal/repo/butterbox/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestFactory_ClientForCarriesDecryptedToken(t *testing.T) {
	ctx := context.Background()
	fake := newFakePi()
	url := serveFake(t, fake)

	repo := butterboxmemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	ciphertext, keyID, err := keyring.Encrypt(ctx, []byte("box-token"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	box, err := repo.Create(ctx, "ws-1", &agentsv1.ButterBox{
		Id: "box-1", Name: "dev box", BaseUrl: url, Enabled: true,
	}, butterboxrepo.Credential{Ciphertext: ciphertext, KeyID: keyID})
	if err != nil {
		t.Fatalf("create box: %v", err)
	}
	if !box.GetCredentialSet() {
		t.Fatal("expected credential_set on the stored box")
	}

	client, err := NewFactory(repo, keyring).ClientFor(ctx, "ws-1", "box-1")
	if err != nil {
		t.Fatalf("ClientFor: %v", err)
	}
	if _, err := client.CreateSession(ctx, connect.NewRequest(&piv1.CreateSessionRequest{})); err != nil {
		t.Fatalf("CreateSession through factory client: %v", err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.gotAuth != "Bearer box-token" {
		t.Fatalf("authorization header: got %q", fake.gotAuth)
	}
}

func TestFactory_ClientForUnknownBox(t *testing.T) {
	f := NewFactory(butterboxmemory.New(), secretbox.NewKeyring(cryptokeymemory.New()))
	_, err := f.ClientFor(context.Background(), "ws-1", "missing-box")
	if err == nil || !strings.Contains(err.Error(), "no longer exists") {
		t.Fatalf("expected pointed not-found error, got %v", err)
	}
}
