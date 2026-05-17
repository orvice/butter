package application

import (
	"context"
	"testing"

	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/repo/apitoken/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRevokeAPIToken_RejectsCrossWorkspace(t *testing.T) {
	store := memory.New()
	svc := NewAPITokenServiceServer(store)

	// Seed a token owned by ws-other.
	if err := store.Create(context.Background(), &agentsv1.APIToken{
		Id:          "tok-1",
		WorkspaceId: "ws-other",
	}, "hash-1"); err != nil {
		t.Fatalf("seed token: %v", err)
	}

	// Caller is in ws-self.
	ctx := workspace.WithID(context.Background(), "ws-self")

	_, err := svc.RevokeAPIToken(ctx, &agentsv1.RevokeAPITokenRequest{Id: "tok-1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected twirp.Error, got %T", err)
	}
	if twerr.Code() != twirp.NotFound {
		t.Fatalf("expected NotFound (to avoid leaking), got %s", twerr.Code())
	}

	// Token must remain un-revoked.
	got, err := store.Get(context.Background(), "tok-1")
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if got.GetRevoked() {
		t.Fatal("token was revoked across workspace boundary")
	}
}

func TestRevokeAPIToken_AllowsSameWorkspace(t *testing.T) {
	store := memory.New()
	svc := NewAPITokenServiceServer(store)

	if err := store.Create(context.Background(), &agentsv1.APIToken{
		Id:          "tok-1",
		WorkspaceId: "ws-self",
	}, "hash-1"); err != nil {
		t.Fatalf("seed token: %v", err)
	}

	ctx := workspace.WithID(context.Background(), "ws-self")
	resp, err := svc.RevokeAPIToken(ctx, &agentsv1.RevokeAPITokenRequest{Id: "tok-1"})
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !resp.GetToken().GetRevoked() {
		t.Fatal("expected token to be revoked")
	}
}

func TestRevokeAPIToken_RequiresWorkspaceContext(t *testing.T) {
	store := memory.New()
	svc := NewAPITokenServiceServer(store)

	_, err := svc.RevokeAPIToken(context.Background(), &agentsv1.RevokeAPITokenRequest{Id: "tok-1"})
	if err == nil {
		t.Fatal("expected error when workspace missing")
	}
	twerr, ok := err.(twirp.Error)
	if !ok {
		t.Fatalf("expected twirp.Error, got %T", err)
	}
	if twerr.Code() != twirp.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %s", twerr.Code())
	}
}
