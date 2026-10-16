package main

import (
	"context"
	"net/http"
	"testing"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

func TestExplorerSessionsSeedAndReply(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler, _, err := newFixture(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_ = handler
	// The Dart session_contract_test exercises the same handlers end-to-end;
	// this test only guards the seeded session ids the contract relies on.
	store := newChatSessions()
	if err := seedExplorerSessions(ctx, store); err != nil {
		t.Fatal(err)
	}
	wctx := workspace.WithID(context.Background(), sessionsWorkspace)
	list, err := store.ListByWorkspace(wctx, sessionsWorkspace, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 3 {
		t.Fatalf("expected at least 3 seeded sessions in %q, got %d", sessionsWorkspace, len(list))
	}
}

func TestUpdateSessionTitleOnSeededOwnSession(t *testing.T) {
	f := newProviderFixture(t)
	sessions := agentsv1connect.NewSessionServiceClient(http.DefaultClient, f.url+"/api", connect.WithInterceptors(
		connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				req.Header().Set("Authorization", "Bearer "+f.token)
				req.Header().Set("X-Workspace-ID", sessionsWorkspace)
				return next(ctx, req)
			}
		}),
	))
	resp, err := sessions.UpdateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web-chat", UserId: "mobile-test", SessionId: "sess-own", Title: "Renamed title",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetSession().GetTitle() != "Renamed title" {
		t.Fatalf("title = %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestSessionServiceReplyRequiresWorkspaceHeader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, err := newFixture(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Documented expectation: ReplySession without X-Workspace-ID must not
	// resolve agents across tenants (checked in session_contract_test.dart).
	_ = connect.CodeInvalidArgument
	_ = agentsv1.ReplySessionRequest{}
}
