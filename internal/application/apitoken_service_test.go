package application

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/repo/apitoken/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// newAPITokenTestService seeds ws-self with an owner, an admin and a member.
func newAPITokenTestService(t *testing.T) (*APITokenServiceServer, *memory.Store) {
	t.Helper()
	wsRepo := workspacememory.New()
	if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: "ws-self", Name: "ws-self", Slug: "ws-self"}); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	for _, m := range []struct{ user, role string }{{"owner-1", "owner"}, {"admin-1", "admin"}, {"member-1", "member"}} {
		if _, err := wsRepo.AddMember(t.Context(), &agentsv1.WorkspaceMember{WorkspaceId: "ws-self", UserId: m.user, Role: m.role}); err != nil {
			t.Fatalf("seed member: %v", err)
		}
	}
	store := memory.New()
	svc := NewAPITokenServiceServer(store)
	svc.SetWorkspaceRepo(wsRepo)
	return svc, store
}

// A token reaches every session in its workspace, so only owners and admins
// mint or revoke one.
func TestAPIToken_OnlyOwnersAndAdminsMintAndRevoke(t *testing.T) {
	cases := []struct {
		name    string
		ctx     context.Context
		allowed bool
	}{
		{"owner", testPersonContext("ws-self", "owner-1"), true},
		{"admin", testPersonContext("ws-self", "admin-1"), true},
		{"global admin", workspace.WithID(testGlobalAdminContext(), "ws-self"), true},
		{"member", testPersonContext("ws-self", "member-1"), false},
		{"another API token", testAPITokenContext("ws-self"), false},
	}
	for _, tc := range cases {
		svc, store := newAPITokenTestService(t)
		_, err := svc.CreateAPIToken(tc.ctx, connect.NewRequest(&agentsv1.CreateAPITokenRequest{Name: "ci"}))
		if tc.allowed != (err == nil) {
			t.Errorf("%s CreateAPIToken: err = %v, want allowed = %v", tc.name, err, tc.allowed)
		}

		if err := store.Create(context.Background(), &agentsv1.APIToken{
			Id: "tok-1", WorkspaceId: "ws-self", Kind: agentsv1.APITokenKind_API_TOKEN_KIND_USER, Scopes: []string{"api:*"},
		}, "hash-1"); err != nil {
			t.Fatalf("seed token: %v", err)
		}
		_, err = svc.RevokeAPIToken(tc.ctx, connect.NewRequest(&agentsv1.RevokeAPITokenRequest{Id: "tok-1"}))
		if tc.allowed != (err == nil) {
			t.Errorf("%s RevokeAPIToken: err = %v, want allowed = %v", tc.name, err, tc.allowed)
		}
		got, _ := store.Get(context.Background(), "tok-1")
		if got.GetRevoked() != tc.allowed {
			t.Errorf("%s RevokeAPIToken: revoked = %v, want %v", tc.name, got.GetRevoked(), tc.allowed)
		}
	}
}

func TestRevokeAPIToken_RejectsCrossWorkspace(t *testing.T) {
	svc, store := newAPITokenTestService(t)

	// Seed a token owned by ws-other.
	if err := store.Create(context.Background(), &agentsv1.APIToken{
		Id:          "tok-1",
		WorkspaceId: "ws-other",
		Kind:        agentsv1.APITokenKind_API_TOKEN_KIND_USER,
		Scopes:      []string{"api:*"},
	}, "hash-1"); err != nil {
		t.Fatalf("seed token: %v", err)
	}

	// Caller is an owner of ws-self.
	ctx := testPersonContext("ws-self", "owner-1")

	_, err := svc.RevokeAPIToken(ctx, connect.NewRequest(&agentsv1.RevokeAPITokenRequest{Id: "tok-1"}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	twerr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if twerr.Code() != connect.CodeNotFound {
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
	svc, store := newAPITokenTestService(t)

	if err := store.Create(context.Background(), &agentsv1.APIToken{
		Id:          "tok-1",
		WorkspaceId: "ws-self",
		Kind:        agentsv1.APITokenKind_API_TOKEN_KIND_USER,
		Scopes:      []string{"api:*"},
	}, "hash-1"); err != nil {
		t.Fatalf("seed token: %v", err)
	}

	ctx := testPersonContext("ws-self", "owner-1")
	resp, err := svc.RevokeAPIToken(ctx, connect.NewRequest(&agentsv1.RevokeAPITokenRequest{Id: "tok-1"}))
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !resp.Msg.GetToken().GetRevoked() {
		t.Fatal("expected token to be revoked")
	}
}

func TestRevokeAPIToken_RequiresWorkspaceContext(t *testing.T) {
	store := memory.New()
	svc := NewAPITokenServiceServer(store)

	_, err := svc.RevokeAPIToken(context.Background(), connect.NewRequest(&agentsv1.RevokeAPITokenRequest{Id: "tok-1"}))
	if err == nil {
		t.Fatal("expected error when workspace missing")
	}
	twerr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if twerr.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %s", twerr.Code())
	}
}
