package application

import (
	"context"
	"slices"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Invocation access (#436): an invocation is reachable exactly when its
// session is, so the invocation RPCs follow the session access policy of
// session_access_test.go over the same callers.

type invocationPolicyFixture struct {
	svc     *AgentServiceServer
	runner  *cancelCountingRunner
	stopper *fakeAGUIStopper
}

// invocationPolicyRecords are one invocation per session of
// newSessionPolicyFixture, named after it.
var invocationPolicyRecords = []struct {
	id, app, user, session, ws, source string
}{
	{"inv-alice", "web-chat", "alice", "s-alice", "ws-a", "CONTEXT_SOURCE_API"},
	{"inv-bob", "agui", "bob", "agui-t-bob", "ws-a", invocation.SourceAGUIDetached},
	{"inv-cron", "cron:nightly", "cron:nightly", "s-cron", "ws-a", "CONTEXT_SOURCE_CRON"},
	{"inv-alice-b", "web-chat", "alice", "s-alice-b", "ws-b", "CONTEXT_SOURCE_API"},
}

func newInvocationPolicyFixture(t *testing.T) *invocationPolicyFixture {
	t.Helper()
	sessions := newSessionPolicyFixture(t)
	invRepo := invocationmemory.New()
	for _, r := range invocationPolicyRecords {
		if err := invRepo.Save(context.Background(), &agentsv1.Invocation{
			Id: r.id, WorkspaceId: r.ws, UserId: r.user, AppName: r.app, SessionId: r.session,
			Source: r.source, Status: agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
			StartedAt: timestamppb.Now(),
		}); err != nil {
			t.Fatalf("seed invocation: %v", err)
		}
	}
	runner := &cancelCountingRunner{}
	stopper := &fakeAGUIStopper{stopped: true}
	svc := &AgentServiceServer{runnerSvc: runner, invRepo: invRepo}
	svc.SetAGUIRunStopper(stopper)
	svc.SetSessionAuthorizer(sessions.svc)
	return &invocationPolicyFixture{svc: svc, runner: runner, stopper: stopper}
}

func okCode() connect.Code { return connect.Code(0) }

// testGlobalAdminContextIn is a global admin naming a workspace.
func testGlobalAdminContextIn(wsID string) context.Context {
	return workspace.WithID(testGlobalAdminContext(), wsID)
}

func TestInvocationAccess_GetAgentInvocation(t *testing.T) {
	ok, notFound, unauth := okCode(), connect.CodeNotFound, connect.CodeUnauthenticated
	cases := []struct {
		caller sessionCaller
		inv    string
		want   connect.Code
	}{
		{aliceInA, "inv-alice", ok},
		{aliceInA, "inv-bob", notFound},
		{aliceInA, "inv-cron", ok},
		{aliceInA, "inv-alice-b", notFound}, // own, but another workspace than the header
		{aliceInA, "inv-missing", notFound},
		{olgaInA, "inv-bob", ok},
		{olgaInA, "inv-cron", ok},
		{olgaInA, "inv-alice-b", notFound},
		{adamInA, "inv-bob", ok},
		{tokenA, "inv-bob", ok},
		{tokenA, "inv-alice-b", notFound},
		{globalAdmin, "inv-alice-b", ok},
		{globalAdmin, "inv-bob", ok},
		{headerOnly, "inv-bob", unauth},
	}
	fx := newInvocationPolicyFixture(t)
	for _, tc := range cases {
		_, err := fx.svc.GetAgentInvocation(tc.caller.ctx(), connect.NewRequest(&agentsv1.GetAgentInvocationRequest{InvocationId: tc.inv}))
		if got := connect.CodeOf(err); err == nil && tc.want != ok || err != nil && got != tc.want {
			t.Errorf("%s GetAgentInvocation(%s): err = %v, want code %v", tc.caller.name, tc.inv, err, tc.want)
		}
	}
}

func TestInvocationAccess_CancelAgentInvocation(t *testing.T) {
	ok, notFound, unauth := okCode(), connect.CodeNotFound, connect.CodeUnauthenticated
	cases := []struct {
		caller sessionCaller
		inv    string
		want   connect.Code
	}{
		{aliceInA, "inv-alice", ok},
		{aliceInA, "inv-bob", notFound}, // another member's AG-UI run
		{aliceInA, "inv-cron", ok},
		{aliceInA, "inv-alice-b", notFound},
		{olgaInA, "inv-bob", ok},
		{adamInA, "inv-bob", ok},
		{tokenA, "inv-bob", ok},
		{tokenA, "inv-alice-b", notFound},
		{globalAdmin, "inv-bob", ok},
		{headerOnly, "inv-bob", unauth},
	}
	for _, tc := range cases {
		fx := newInvocationPolicyFixture(t)
		_, err := fx.svc.CancelAgentInvocation(tc.caller.ctx(), connect.NewRequest(&agentsv1.CancelAgentInvocationRequest{InvocationId: tc.inv}))
		if got := connect.CodeOf(err); err == nil && tc.want != ok || err != nil && got != tc.want {
			t.Errorf("%s CancelAgentInvocation(%s): err = %v, want code %v", tc.caller.name, tc.inv, err, tc.want)
		}
		if tc.want != ok && (len(fx.stopper.calls) != 0 || fx.runner.cancels != 0) {
			t.Errorf("%s CancelAgentInvocation(%s) was refused but reached the run: stops %v, runner cancels %d",
				tc.caller.name, tc.inv, fx.stopper.calls, fx.runner.cancels)
		}
	}
}

func TestInvocationAccess_ListAgentInvocations(t *testing.T) {
	cases := []struct {
		caller sessionCaller
		want   []string
	}{
		{aliceInA, []string{"inv-alice", "inv-cron"}},
		{olgaInA, []string{"inv-alice", "inv-bob", "inv-cron"}},
		{adamInA, []string{"inv-alice", "inv-bob", "inv-cron"}},
		{tokenA, []string{"inv-alice", "inv-bob", "inv-cron"}},
		{globalAdmin, []string{"inv-alice", "inv-bob", "inv-cron"}},
	}
	fx := newInvocationPolicyFixture(t)
	for _, tc := range cases {
		// A global admin without a workspace header is not scoped; give it
		// the one the others use.
		ctx := tc.caller.ctx()
		if tc.caller.name == globalAdmin.name {
			ctx = testGlobalAdminContextIn("ws-a")
		}
		var got []string
		token := ""
		for {
			resp, err := fx.svc.ListAgentInvocations(ctx, connect.NewRequest(&agentsv1.ListAgentInvocationsRequest{PageSize: 1, PageToken: token}))
			if err != nil {
				t.Fatalf("%s ListAgentInvocations: %v", tc.caller.name, err)
			}
			if int(resp.Msg.GetTotal()) != len(tc.want) {
				t.Errorf("%s total = %d, want %d", tc.caller.name, resp.Msg.GetTotal(), len(tc.want))
			}
			for _, inv := range resp.Msg.GetInvocations() {
				got = append(got, inv.GetId())
			}
			if token = resp.Msg.GetNextPageToken(); token == "" {
				break
			}
		}
		slices.Sort(got)
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s lists %v, want %v", tc.caller.name, got, tc.want)
		}
	}

	for _, c := range []sessionCaller{headerOnly} {
		_, err := fx.svc.ListAgentInvocations(c.ctx(), connect.NewRequest(&agentsv1.ListAgentInvocationsRequest{}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("%s ListAgentInvocations: err = %v, want unauthenticated", c.name, err)
		}
	}
}
