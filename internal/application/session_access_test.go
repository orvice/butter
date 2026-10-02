package application

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/auth"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// testAPITokenContext is a request authenticated by a workspace API token.
func testAPITokenContext(wsID string) context.Context {
	return auth.WithAPIToken(workspace.WithID(context.Background(), wsID), "tok-test")
}

// testPersonContext is a signed-in person, with X-Workspace-ID when wsID is
// set.
func testPersonContext(wsID, userID string) context.Context {
	ctx := context.Background()
	if wsID != "" {
		ctx = workspace.WithID(ctx, wsID)
	}
	return auth.WithAuthenticated(ctx, &agentsv1.User{Id: userID, Role: "member"}, nil)
}

func testGlobalAdminContext() context.Context {
	return auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "root", Role: "admin"}, nil)
}

// policySessionService is a session.Service over a fixed set of sessions
// whose List honors the app and user filters like the Mongo backend.
type policySessionService struct {
	session.Service
	sessions []*fakeWSSession
	deleted  []string
	created  []string
}

func (f *policySessionService) find(appName, userID, sessionID string) *fakeWSSession {
	for _, s := range f.sessions {
		if s.appName == appName && s.userID == userID && s.id == sessionID {
			return s
		}
	}
	return nil
}

func (f *policySessionService) Get(_ context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	if s := f.find(req.AppName, req.UserID, req.SessionID); s != nil {
		return &session.GetResponse{Session: s}, nil
	}
	return nil, errors.New("session not found")
}

func (f *policySessionService) List(_ context.Context, req *session.ListRequest) (*session.ListResponse, error) {
	var out []session.Session
	for _, s := range f.sessions {
		if (req.AppName == "" || s.appName == req.AppName) && (req.UserID == "" || s.userID == req.UserID) {
			out = append(out, s)
		}
	}
	return &session.ListResponse{Sessions: out}, nil
}

func (f *policySessionService) Delete(_ context.Context, req *session.DeleteRequest) error {
	f.deleted = append(f.deleted, req.SessionID)
	return nil
}

func (f *policySessionService) Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error) {
	wsID, _ := workspace.FromContext(ctx)
	f.created = append(f.created, req.UserID+"/"+req.SessionID)
	return &session.CreateResponse{Session: &fakeWSSession{id: req.SessionID, appName: req.AppName, userID: req.UserID, wsID: wsID}}, nil
}

type sessionPolicyFixture struct {
	svc      *SessionServiceServer
	sessions *policySessionService
	runner   *policyRunner
}

type policyRunner struct{ ran []string }

func (r *policyRunner) Run(_ context.Context, _ string, _ []*genai.Part, _ string, ctxInfo *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.ran = append(r.ran, ctxInfo.GetSessionId())
	return "ok", nil
}

func (r *policyRunner) ResolveAgentRef(_, agentID string) (string, bool) {
	return agentID, agentID != ""
}

// newSessionPolicyFixture seeds two workspaces. ws-a: owner olga, admin
// adam, members alice and bob. ws-b: member alice.
func newSessionPolicyFixture(t *testing.T) *sessionPolicyFixture {
	t.Helper()
	wsRepo := workspacememory.New()
	for _, ws := range []string{"ws-a", "ws-b"} {
		if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: ws, Name: ws, Slug: ws}); err != nil {
			t.Fatalf("seed workspace: %v", err)
		}
	}
	for _, m := range []struct{ ws, user, role string }{
		{"ws-a", "olga", "owner"},
		{"ws-a", "adam", "admin"},
		{"ws-a", "alice", "member"},
		{"ws-a", "bob", "member"},
		{"ws-b", "alice", "member"},
	} {
		if _, err := wsRepo.AddMember(t.Context(), &agentsv1.WorkspaceMember{WorkspaceId: m.ws, UserId: m.user, Role: m.role}); err != nil {
			t.Fatalf("seed member: %v", err)
		}
	}

	now := time.Now()
	sessions := &policySessionService{sessions: []*fakeWSSession{
		{id: "s-alice", appName: "web-chat", userID: "alice", wsID: "ws-a", lastUpdate: now},
		{id: "s-bob", appName: "web-chat", userID: "bob", wsID: "ws-a", lastUpdate: now.Add(-time.Minute)},
		{id: "s-cron", appName: "cron:nightly", userID: "cron:nightly", wsID: "ws-a", lastUpdate: now.Add(-2 * time.Minute)},
		{id: "s-alice-b", appName: "web-chat", userID: "alice", wsID: "ws-b", lastUpdate: now.Add(-3 * time.Minute)},
		{id: "s-legacy", appName: "web-chat", userID: "alice", wsID: "", lastUpdate: now.Add(-4 * time.Minute)},
	}}
	wsStore := newFakeWSStore()
	for _, s := range sessions.sessions {
		wsStore.addSession(s.wsID, s)
	}

	r := &policyRunner{}
	svc := NewSessionServiceServer()
	svc.SetSessionService(sessions)
	svc.SetWorkspaceSessionStore(wsStore)
	svc.SetWorkspaceRepo(wsRepo)
	svc.runnerSvc = r
	return &sessionPolicyFixture{svc: svc, sessions: sessions, runner: r}
}

var sessionCoords = map[string][2]string{ // session ID -> (app, user)
	"s-alice":   {"web-chat", "alice"},
	"s-bob":     {"web-chat", "bob"},
	"s-cron":    {"cron:nightly", "cron:nightly"},
	"s-alice-b": {"web-chat", "alice"},
	"s-legacy":  {"web-chat", "alice"},
	"s-missing": {"web-chat", "alice"},
}

type sessionCaller struct {
	name string
	ctx  func() context.Context
}

var (
	aliceInA      = sessionCaller{"member alice in ws-a", func() context.Context { return testPersonContext("ws-a", "alice") }}
	aliceNoHeader = sessionCaller{"member alice without header", func() context.Context { return testPersonContext("", "alice") }}
	olgaInA       = sessionCaller{"owner olga in ws-a", func() context.Context { return testPersonContext("ws-a", "olga") }}
	adamInA       = sessionCaller{"admin adam in ws-a", func() context.Context { return testPersonContext("ws-a", "adam") }}
	olgaNoHeader  = sessionCaller{"owner olga without header", func() context.Context { return testPersonContext("", "olga") }}
	tokenA        = sessionCaller{"API token of ws-a", func() context.Context { return testAPITokenContext("ws-a") }}
	globalAdmin   = sessionCaller{"global admin", testGlobalAdminContext}
	anonymous     = sessionCaller{"anonymous", context.Background}
	// A workspace header alone, without a user or an API token, is not an
	// identity.
	headerOnly = sessionCaller{"workspace header only", func() context.Context { return workspace.WithID(context.Background(), "ws-a") }}
)

func TestSessionAccess_GetSession(t *testing.T) {
	ok, notFound, unauth := connect.Code(0), connect.CodeNotFound, connect.CodeUnauthenticated
	cases := []struct {
		caller  sessionCaller
		session string
		want    connect.Code
	}{
		{aliceInA, "s-alice", ok},
		{aliceInA, "s-bob", notFound},
		{aliceInA, "s-cron", ok},
		{aliceInA, "s-alice-b", notFound}, // own, but another workspace than the header
		{aliceInA, "s-legacy", notFound},
		{aliceInA, "s-missing", notFound},
		{aliceNoHeader, "s-alice-b", ok},
		{aliceNoHeader, "s-legacy", ok},
		{aliceNoHeader, "s-bob", notFound},
		{aliceNoHeader, "s-cron", notFound},
		{olgaInA, "s-bob", ok},
		{olgaInA, "s-cron", ok},
		{olgaInA, "s-alice-b", notFound},
		{adamInA, "s-bob", ok},
		{olgaNoHeader, "s-bob", notFound}, // managing needs the workspace named
		{tokenA, "s-bob", ok},
		{tokenA, "s-alice-b", notFound},
		{globalAdmin, "s-alice-b", ok},
		{anonymous, "s-bob", unauth},
		{anonymous, "s-missing", unauth},
		{headerOnly, "s-bob", unauth},
	}
	fx := newSessionPolicyFixture(t)
	for _, tc := range cases {
		coord := sessionCoords[tc.session]
		_, err := fx.svc.GetSession(tc.caller.ctx(), connect.NewRequest(&agentsv1.GetSessionRequest{
			AppName: coord[0], UserId: coord[1], SessionId: tc.session,
		}))
		if got := connect.CodeOf(err); err == nil && tc.want != ok || err != nil && got != tc.want {
			t.Errorf("%s GetSession(%s): err = %v, want code %v", tc.caller.name, tc.session, err, tc.want)
		}
	}
}

func TestSessionAccess_DeleteSession(t *testing.T) {
	cases := []struct {
		caller  sessionCaller
		session string
		allowed bool
	}{
		{aliceInA, "s-bob", false},
		{aliceInA, "s-alice", true},
		{aliceInA, "s-cron", true},
		{aliceNoHeader, "s-bob", false},
		{olgaInA, "s-bob", true},
		{tokenA, "s-alice-b", false},
		{headerOnly, "s-bob", false},
	}
	for _, tc := range cases {
		fx := newSessionPolicyFixture(t)
		coord := sessionCoords[tc.session]
		_, err := fx.svc.DeleteSession(tc.caller.ctx(), connect.NewRequest(&agentsv1.DeleteSessionRequest{
			AppName: coord[0], UserId: coord[1], SessionId: tc.session,
		}))
		deleted := slices.Contains(fx.sessions.deleted, tc.session)
		if tc.allowed && (err != nil || !deleted) {
			t.Errorf("%s DeleteSession(%s): err = %v, deleted = %v, want deleted", tc.caller.name, tc.session, err, deleted)
		}
		if !tc.allowed && (err == nil || deleted) {
			t.Errorf("%s DeleteSession(%s): err = %v, deleted = %v, want refused", tc.caller.name, tc.session, err, deleted)
		}
	}
}

func TestSessionAccess_ReplySession(t *testing.T) {
	cases := []struct {
		caller          sessionCaller
		app, user, sess string
		allowed         bool
	}{
		{aliceInA, "web-chat", "bob", "s-bob", false},
		{aliceInA, "cron:nightly", "cron:nightly", "s-cron", true}, // resuming a paused cron run (ADR-0003)
		{aliceInA, "web-chat", "alice", "s-new", true},
		{aliceInA, "web-chat", "bob", "s-new", false},
		{aliceNoHeader, "cron:nightly", "cron:nightly", "s-cron", false},
		{olgaInA, "web-chat", "bob", "s-bob", true},
		{olgaInA, "web-chat", "bob", "s-new", true},
		{tokenA, "api", "customer-42", "s-new", true},
		{tokenA, "web-chat", "alice", "s-alice-b", false},
		{headerOnly, "api", "customer-42", "s-new", false},
	}
	for _, tc := range cases {
		fx := newSessionPolicyFixture(t)
		_, err := fx.svc.ReplySession(tc.caller.ctx(), connect.NewRequest(&agentsv1.ReplySessionRequest{
			AgentId: "agent-1", AppName: tc.app, UserId: tc.user, SessionId: tc.sess, Message: "hi",
		}))
		ran := slices.Contains(fx.runner.ran, tc.sess)
		if tc.allowed && (err != nil || !ran) {
			t.Errorf("%s ReplySession(%s/%s): err = %v, ran = %v, want a turn", tc.caller.name, tc.user, tc.sess, err, ran)
		}
		if !tc.allowed && (err == nil || ran) {
			t.Errorf("%s ReplySession(%s/%s): err = %v, ran = %v, want refused before the runner", tc.caller.name, tc.user, tc.sess, err, ran)
		}
	}
}

func TestSessionAccess_CreateSession(t *testing.T) {
	cases := []struct {
		caller  sessionCaller
		user    string
		allowed bool
	}{
		{aliceInA, "alice", true},
		{aliceInA, "bob", false},
		{aliceNoHeader, "alice", true},
		{aliceNoHeader, "bob", false},
		{olgaInA, "bob", true},
		{tokenA, "customer-42", true},
		{globalAdmin, "bob", true},
		{headerOnly, "bob", false},
	}
	for _, tc := range cases {
		fx := newSessionPolicyFixture(t)
		_, err := fx.svc.CreateSession(tc.caller.ctx(), connect.NewRequest(&agentsv1.CreateSessionRequest{
			AppName: "web-chat", UserId: tc.user, SessionId: "s-new",
		}))
		created := len(fx.sessions.created) == 1
		if tc.allowed != (err == nil && created) {
			t.Errorf("%s CreateSession(for %s): err = %v, created = %v, want allowed = %v", tc.caller.name, tc.user, err, created, tc.allowed)
		}
	}
}

func TestSessionAccess_ListSessions(t *testing.T) {
	cases := []struct {
		caller    sessionCaller
		app, user string
		want      []string // session IDs, newest first; nil with denied
		denied    bool
	}{
		{caller: aliceInA, want: []string{"s-alice"}},
		{caller: aliceInA, user: "alice", want: []string{"s-alice"}},
		{caller: aliceInA, user: "bob", denied: true},
		{caller: aliceInA, app: "cron:nightly", want: []string{"s-cron"}},
		{caller: aliceNoHeader, want: []string{"s-alice", "s-alice-b", "s-legacy"}},
		{caller: aliceNoHeader, user: "alice", want: []string{"s-alice", "s-alice-b", "s-legacy"}},
		{caller: aliceNoHeader, user: "bob", denied: true},
		{caller: olgaInA, want: []string{"s-alice", "s-bob", "s-cron"}},
		{caller: olgaInA, user: "bob", want: []string{"s-bob"}},
		{caller: olgaNoHeader, user: "bob", denied: true},
		{caller: tokenA, want: []string{"s-alice", "s-bob", "s-cron"}},
		{caller: globalAdmin, want: []string{"s-alice", "s-bob", "s-cron", "s-alice-b", "s-legacy"}},
		{caller: anonymous, denied: true},
		{caller: headerOnly, denied: true},
	}
	fx := newSessionPolicyFixture(t)
	for _, tc := range cases {
		resp, err := fx.svc.ListSessions(tc.caller.ctx(), connect.NewRequest(&agentsv1.ListSessionsRequest{
			AppName: tc.app, UserId: tc.user, PageSize: 50,
		}))
		if tc.denied {
			if err == nil {
				t.Errorf("%s ListSessions(app=%q user=%q): want refused, got %d sessions", tc.caller.name, tc.app, tc.user, len(resp.Msg.GetSessions()))
			}
			continue
		}
		if err != nil {
			t.Errorf("%s ListSessions(app=%q user=%q): %v", tc.caller.name, tc.app, tc.user, err)
			continue
		}
		var got []string
		for _, s := range resp.Msg.GetSessions() {
			got = append(got, s.GetSessionId())
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s ListSessions(app=%q user=%q) = %v, want %v", tc.caller.name, tc.app, tc.user, got, tc.want)
		}
	}
}

func TestSessionAccess_WorkspaceScopedListForAPIToken(t *testing.T) {
	fx := newSessionPolicyFixture(t)
	resp, err := fx.svc.ListSessions(testAPITokenContext("ws-a"), connect.NewRequest(&agentsv1.ListSessionsRequest{
		WorkspaceScoped: true, PageSize: 50,
	}))
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	var got []string
	for _, s := range resp.Msg.GetSessions() {
		got = append(got, s.GetSessionId())
	}
	slices.Sort(got)
	if want := []string{"s-alice", "s-bob", "s-cron"}; !slices.Equal(got, want) {
		t.Fatalf("token's workspace-scoped listing = %v, want %v", got, want)
	}
}
