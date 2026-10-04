package application

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/runtime/sessionshare"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// allowAllTurns admits every turn, for tests about what a turn carries
// rather than who may run it.
type allowAllTurns struct{}

func (allowAllTurns) AuthorizeTurn(context.Context, string, string, string) error { return nil }

// turnRecordingRunner records the ContextInfo of every turn and can fail
// them all with err.
type turnRecordingRunner struct {
	invokeTestRunner
	got []*agentsv1.ContextInfo
	err error
}

func (r *turnRecordingRunner) Run(_ context.Context, _ string, _ []*genai.Part, _ string, ctxInfo *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.got = append(r.got, ctxInfo)
	if r.err != nil {
		return "", r.err
	}
	return "ok", nil
}

func (r *turnRecordingRunner) RunSSE(ctx context.Context, agentName string, parts []*genai.Part, model string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error) {
	return r.Run(ctx, agentName, parts, model, ctxInfo, onEvent, onCompaction)
}

// newTurnPolicyService is an AgentService that applies the real session
// policy over the sessions of newSessionPolicyFixture.
func newTurnPolicyService(t *testing.T) (*AgentServiceServer, *turnRecordingRunner) {
	t.Helper()
	fx := newSessionPolicyFixture(t)
	r := &turnRecordingRunner{invokeTestRunner: invokeTestRunner{idToName: map[string]string{"helper": "helper"}}}
	svc := NewAgentServiceServer(memory.New())
	svc.runnerSvc = r
	svc.SetSessionAuthorizer(fx.svc)
	return svc, r
}

// turnPolicyCases cover who may run a turn on which session; they mirror
// the ReplySession rules.
var turnPolicyCases = []struct {
	name               string
	caller             sessionCaller
	app, user, session string
	want               connect.Code // 0: the turn runs
}{
	{"member on own session", aliceInA, "web-chat", "alice", "s-alice", 0},
	{"member on another member's session", aliceInA, "web-chat", "bob", "s-bob", connect.CodeNotFound},
	{"member starting a session for someone else", aliceInA, "web-chat", "bob", "s-new", connect.CodePermissionDenied},
	{"member on own session from another workspace", aliceInA, "web-chat", "alice", "s-alice-b", connect.CodeNotFound},
	{"member on the workspace's cron session", aliceInA, "cron:nightly", "cron:nightly", "s-cron", 0},
	{"owner on a member's session", olgaInA, "web-chat", "bob", "s-bob", 0},
	{"API token on a member's session", tokenA, "web-chat", "bob", "s-bob", 0},
	{"API token on another workspace's session", tokenA, "web-chat", "alice", "s-alice-b", connect.CodeNotFound},
}

func TestInvokeAgent_FollowsTheSessionAccessPolicy(t *testing.T) {
	for _, tc := range turnPolicyCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, r := newTurnPolicyService(t)
			_, err := svc.InvokeAgent(tc.caller.ctx(), connect.NewRequest(&agentsv1.InvokeAgentRequest{
				AgentId: "helper", Input: "hi", AppName: tc.app, UserId: tc.user, SessionId: tc.session,
			}))
			if tc.want == 0 {
				if err != nil || len(r.got) != 1 {
					t.Fatalf("err = %v, turns = %d; want the turn to run", err, len(r.got))
				}
				return
			}
			if got := connect.CodeOf(err); err == nil || got != tc.want {
				t.Fatalf("err = %v, want code %v", err, tc.want)
			}
			if len(r.got) != 0 {
				t.Fatal("a refused turn must not run")
			}
		})
	}
}

func TestStreamAgent_FollowsTheSessionAccessPolicy(t *testing.T) {
	for _, tc := range turnPolicyCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, r := newTurnPolicyService(t)
			client := newAgentClientAs(t, svc, tc.caller)
			stream, err := client.StreamAgent(context.Background(), connect.NewRequest(&agentsv1.StreamAgentRequest{
				AgentId: "helper", Message: "hi", AppName: tc.app, UserId: tc.user, SessionId: tc.session,
			}))
			if err == nil {
				for stream.Receive() {
				}
				err = stream.Err()
				_ = stream.Close()
			}
			if tc.want == 0 {
				if err != nil || len(r.got) != 1 {
					t.Fatalf("err = %v, turns = %d; want the turn to run", err, len(r.got))
				}
				return
			}
			if got := connect.CodeOf(err); err == nil || got != tc.want {
				t.Fatalf("err = %v, want code %v", err, tc.want)
			}
			if len(r.got) != 0 {
				t.Fatal("a refused turn must not run")
			}
		})
	}
}

// newAgentClientAs serves svc to a client whose every request is
// authenticated as caller.
func newAgentClientAs(t *testing.T, svc *AgentServiceServer, caller sessionCaller) agentsv1connect.AgentServiceClient {
	t.Helper()
	path, handler := agentsv1connect.NewAgentServiceHandler(svc, connectx.HandlerOptions()...)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(w, req.WithContext(caller.ctx()))
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return agentsv1connect.NewAgentServiceClient(srv.Client(), srv.URL)
}

// A person who names no user runs as themselves, inside their own sessions,
// rather than as the "api" user every such caller would share.
func TestTurnsWithoutAUserRunAsTheSignedInPerson(t *testing.T) {
	svc, r := newTurnPolicyService(t)
	if _, err := svc.InvokeAgent(aliceInA.ctx(), connect.NewRequest(&agentsv1.InvokeAgentRequest{AgentId: "helper", Input: "hi"})); err != nil {
		t.Fatalf("person: %v", err)
	}
	if _, err := svc.InvokeAgent(tokenA.ctx(), connect.NewRequest(&agentsv1.InvokeAgentRequest{AgentId: "helper", Input: "hi"})); err != nil {
		t.Fatalf("API token: %v", err)
	}
	if got := r.got[0].GetUserId(); got != "alice" {
		t.Errorf("person ran as %q, want alice", got)
	}
	if got := r.got[1].GetUserId(); got != "api" {
		t.Errorf("API token ran as %q, want api", got)
	}
}

// Without a policy wired, only a global admin may run a turn.
func TestUnwiredTurnPolicyFailsClosed(t *testing.T) {
	r := &turnRecordingRunner{invokeTestRunner: invokeTestRunner{idToName: map[string]string{"helper": "helper"}}}
	svc := NewAgentServiceServer(memory.New())
	svc.runnerSvc = r
	req := &agentsv1.InvokeAgentRequest{AgentId: "helper", Input: "hi"}
	if _, err := svc.InvokeAgent(tokenA.ctx(), connect.NewRequest(req)); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("API token: err = %v, want failed_precondition", err)
	}
	if _, err := svc.InvokeAgent(globalAdmin.ctx(), connect.NewRequest(req)); err != nil {
		t.Fatalf("global admin: %v", err)
	}
}

// A turn whose session ID another user holds is the caller's mistake, not
// an internal error, on every turn entry point.
func TestTurnsOnASessionIDAnotherUserHoldsAreRefused(t *testing.T) {
	taken := fmt.Errorf("creating session: %w", sessionshare.ErrIDTaken)

	svc, r := newTurnPolicyService(t)
	r.err = taken
	_, err := svc.InvokeAgent(aliceInA.ctx(), connect.NewRequest(&agentsv1.InvokeAgentRequest{
		AgentId: "helper", Input: "hi", AppName: "web-chat", SessionId: "s-new",
	}))
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Errorf("InvokeAgent: err = %v, want already_exists", err)
	}

	if code := streamAgentError(taken).Code(); code != connect.CodeAlreadyExists {
		t.Errorf("StreamAgent: code = %v, want already_exists", code)
	}

	fx := newSessionPolicyFixture(t)
	fx.svc.runnerSvc = &failingReplyRunner{err: taken}
	_, err = fx.svc.ReplySession(aliceInA.ctx(), connect.NewRequest(&agentsv1.ReplySessionRequest{
		AgentId: "helper", Message: "hi", AppName: "web-chat", UserId: "alice", SessionId: "s-new",
	}))
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Errorf("ReplySession: err = %v, want already_exists", err)
	}
}

func TestCreateSession_RefusesASessionIDAnotherUserHolds(t *testing.T) {
	fx := newSessionPolicyFixture(t)
	fx.svc.SetSessionService(&takenSessionService{fx.sessions})
	_, err := fx.svc.CreateSession(aliceInA.ctx(), connect.NewRequest(&agentsv1.CreateSessionRequest{
		AppName: "web-chat", UserId: "alice", SessionId: "s-bob",
	}))
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("err = %v, want already_exists", err)
	}
}

// failingReplyRunner fails every turn with err.
type failingReplyRunner struct{ err error }

func (r *failingReplyRunner) Run(context.Context, string, []*genai.Part, string, *agentsv1.ContextInfo, runner.EventCallback, runner.CompactionCallback) (string, error) {
	return "", r.err
}

func (r *failingReplyRunner) ResolveAgentRef(_, agentID string) (string, bool) {
	return agentID, agentID != ""
}

// takenSessionService refuses every creation the way the store refuses a
// session ID another user holds.
type takenSessionService struct{ *policySessionService }

func (s *takenSessionService) Create(context.Context, *session.CreateRequest) (*session.CreateResponse, error) {
	return nil, fmt.Errorf("%w: web-chat/s-bob", sessionshare.ErrIDTaken)
}
