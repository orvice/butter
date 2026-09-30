package linear

// Orchestrator tests at the primary seam (ADR-0015, spec #358): one
// accepted event in, Linear activities and Agent turns out.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/linearapi/lineartest"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	linearmemory "go.orx.me/apps/butter/internal/repo/linear/memory"
	"go.orx.me/apps/butter/internal/runtime/linearconn"
	"go.orx.me/apps/butter/internal/runtime/memoryhook"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	testToken     = "lin-access-token"
	linearUserOne = "8f7e6d5c-4b3a-4a1b-9c8d-7e6f5a4b3c2d"
	linearUserTwo = "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d"
)

type runCall struct {
	agent       string
	sessionID   string
	userID      string
	channel     string
	channelType string
	text        string
	principal   string
	deadline    time.Duration
}

// fakeRunner records turns instead of running ADK. Sessions live in ADK's
// in-memory service so history behaves like the real runner's.
type fakeRunner struct {
	mu       sync.Mutex
	known    map[string]string
	calls    []runCall
	output   string
	err      error
	block    chan struct{}
	started  chan struct{}
	sessions session.Service
	// emit, when set, plays events into the turn's callbacks.
	emit func(onEvent runner.EventCallback, onCompaction runner.CompactionCallback)
	// ran is how long the last RunTurnSSE took.
	ran time.Duration
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{
		known:    map[string]string{"support": "Support Agent"},
		output:   "Fixed the login bug.",
		sessions: session.InMemoryService(),
	}
}

func (r *fakeRunner) ResolveAgentRef(_ string, agentID string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name, ok := r.known[agentID]
	return name, ok
}

func (r *fakeRunner) GetSession(ctx context.Context, channelName, sessionID, userID string) (session.Session, error) {
	resp, err := r.sessions.Get(ctx, &session.GetRequest{AppName: channelName, UserID: userID, SessionID: sessionID})
	if err != nil {
		return nil, err
	}
	return resp.Session, nil
}

func (r *fakeRunner) RunTurnSSE(ctx context.Context, agentName string, parts []*genai.Part, _ string,
	info *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (*runner.TurnResult, error) {
	began := time.Now()
	defer func() {
		r.mu.Lock()
		r.ran = time.Since(began)
		r.mu.Unlock()
	}()
	call := runCall{
		agent: agentName, sessionID: info.GetSessionId(), userID: info.GetUserId(),
		channel: info.GetChannelName(), channelType: info.GetChannelType(),
		text: parts[0].Text, principal: info.GetMetadata()[memoryhook.PrincipalMetadataKey],
	}
	if deadline, ok := ctx.Deadline(); ok {
		call.deadline = time.Until(deadline).Round(time.Minute)
	}
	r.mu.Lock()
	r.calls = append(r.calls, call)
	block, started, output, err, emit := r.block, r.started, r.output, r.err, r.emit
	r.mu.Unlock()
	if started != nil {
		started <- struct{}{}
	}
	if emit != nil {
		emit(onEvent, onCompaction)
	}
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
			return &runner.TurnResult{}, ctx.Err()
		}
	}
	if err != nil {
		return &runner.TurnResult{}, err
	}
	created, getErr := r.sessions.Get(ctx, &session.GetRequest{AppName: info.GetChannelName(), UserID: info.GetUserId(), SessionID: info.GetSessionId()})
	var sess session.Session
	if getErr == nil {
		sess = created.Session
	} else {
		resp, createErr := r.sessions.Create(ctx, &session.CreateRequest{AppName: info.GetChannelName(), UserID: info.GetUserId(), SessionID: info.GetSessionId()})
		if createErr != nil {
			return nil, createErr
		}
		sess = resp.Session
	}
	ev := session.NewEvent(ctx, "inv")
	ev.Content = genai.NewContentFromText(output, genai.RoleModel)
	if err := r.sessions.AppendEvent(ctx, sess, ev); err != nil {
		return nil, err
	}
	return &runner.TurnResult{Output: output}, nil
}

func (r *fakeRunner) turns() []runCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]runCall(nil), r.calls...)
}

type orchestratorFixture struct {
	repo   *linearmemory.Store
	linear *lineartest.Fake
	runner *fakeRunner
	orch   *Orchestrator
	coord  *MemoryCoordinator
	app    *agentsv1.LinearApp
}

func newOrchestratorFixture(t *testing.T, mutate func(*agentsv1.LinearApp)) *orchestratorFixture {
	t.Helper()
	ctx := t.Context()
	repo := linearmemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	fake := lineartest.New(t)
	fake.AddToken(testToken, linearapi.Identity{AppUserID: "app-user", OrganizationID: "org-1"})

	app := &agentsv1.LinearApp{Id: "app-1", ClientId: "client-1", AgentId: "support", DisplayName: "Support", InboundEnabled: true}
	if mutate != nil {
		mutate(app)
	}
	stored, err := repo.CreateApp(ctx, "ws-a", app, linearrepo.AppCredentials{})
	if err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	ciphertext, keyID, err := keyring.Encrypt(ctx, []byte(testToken))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := repo.UpsertInstallation(ctx, "ws-a", &agentsv1.LinearInstallation{
		Id: "inst-1", AppId: "app-1", OrganizationId: "org-1", OrganizationName: "Acme",
	}, linearrepo.InstallationTokens{AccessToken: linearrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}

	agents := newFakeRunner()
	orch := NewOrchestrator(repo, agents, linearconn.NewTokenSource(repo, keyring), fake.Client())
	coord := NewMemoryCoordinator()
	orch.SetSessionCoordinator(coord)
	orch.SetExternalBaseURL(func(context.Context) string { return "https://butter.test" })
	return &orchestratorFixture{repo: repo, linear: fake, runner: agents, orch: orch, coord: coord, app: stored}
}

func (fx *orchestratorFixture) event(action string) *Event {
	return &Event{
		WorkspaceID: "ws-a", AppID: "app-1", AppRevision: fx.app.GetRevision(),
		InstallationID: "inst-1", OrganizationID: "org-1",
		AgentSessionID: "session-1", Action: action, PromptingUserID: linearUserOne,
		PromptContext: "<issue identifier=\"ENG-1\">Fix login</issue>",
		Issue:         Issue{Identifier: "ENG-1", Title: "Fix login", URL: "https://linear.app/acme/issue/ENG-1"},
		DeliveryID:    "d-" + action,
	}
}

func (fx *orchestratorFixture) prompted(text string) *Event {
	ev := fx.event(ActionPrompted)
	ev.PromptText = text
	ev.DeliveryID = "d-" + text
	return ev
}

func (fx *orchestratorFixture) handle(t *testing.T, ev *Event) {
	t.Helper()
	if err := fx.orch.Handle(t.Context(), ev); err != nil {
		t.Fatalf("Handle(%s): %v", ev.Action, err)
	}
}

func activityTypes(acts []lineartest.Activity) []string {
	out := make([]string, len(acts))
	for i, a := range acts {
		out[i] = a.Type
	}
	return out
}

func TestACreatedSessionIsAcknowledgedFirstThenAnswered(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	var runsAtAck int
	fx.linear.WhenActivity(func(a lineartest.Activity) {
		if a.Type == linearapi.ActivityThought {
			runsAtAck = len(fx.runner.turns())
		}
	})
	fx.handle(t, fx.event(ActionCreated))

	acts := fx.linear.Activities()
	if got := strings.Join(activityTypes(acts), ","); got != "thought,response" {
		t.Fatalf("activities = %s, want thought,response", got)
	}
	if runsAtAck != 0 {
		t.Fatal("the acknowledgement was posted after the Agent started")
	}
	if !strings.Contains(acts[0].Body, "ENG-1: Fix login") || !strings.Contains(acts[0].Body, "Support Agent") {
		t.Errorf("acknowledgement = %q", acts[0].Body)
	}
	if acts[1].Body != "Fixed the login bug." || acts[1].AgentSessionID != "session-1" || acts[1].Token != testToken {
		t.Errorf("response = %+v", acts[1])
	}

	updates := fx.linear.SessionUpdates()
	if len(updates) != 1 || len(updates[0].ExternalURLs) != 1 {
		t.Fatalf("session updates = %+v; want one external URL", updates)
	}
	link, _ := updates[0].ExternalURLs[0]["url"].(string)
	if !strings.HasPrefix(link, "https://butter.test/sessions/detail?") || !strings.Contains(link, "app=linear") {
		t.Errorf("external URL = %q", link)
	}
}

func TestTheTurnRunsInTheDerivedSessionWithLinearContext(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.handle(t, fx.event(ActionCreated))

	turns := fx.runner.turns()
	if len(turns) != 1 {
		t.Fatalf("turns = %d", len(turns))
	}
	got := turns[0]
	want := runCall{
		agent: "Support Agent", sessionID: "linear:app-1:session-1:support", userID: "linear:app-1:org-1",
		channel: "linear", channelType: "linear", principal: "linear:" + linearUserOne, deadline: 30 * time.Minute,
		text: "Linear context:\n<issue identifier=\"ENG-1\">Fix login</issue>",
	}
	if got != want {
		t.Fatalf("turn = %+v\nwant   %+v", got, want)
	}
}

func TestAFollowUpContinuesTheSessionWithTheMessageOnly(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.handle(t, fx.event(ActionCreated))
	fx.handle(t, fx.prompted("also add a test"))

	turns := fx.runner.turns()
	if len(turns) != 2 || turns[1].sessionID != turns[0].sessionID {
		t.Fatalf("turns = %+v; want two turns in one session", turns)
	}
	if turns[1].text != "also add a test" {
		t.Fatalf("follow-up input = %q, want the message only", turns[1].text)
	}
	// A follow-up is no new session: it is not acknowledged again.
	if got := strings.Join(activityTypes(fx.linear.Activities()), ","); got != "thought,response,response" {
		t.Fatalf("activities = %s", got)
	}
}

func TestAPromptOnASessionWithoutHistoryCarriesTheContextAgain(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.handle(t, fx.prompted("what's the status?"))
	text := fx.runner.turns()[0].text
	if !strings.HasPrefix(text, "Linear context:\n") || !strings.HasSuffix(text, "Message from the Linear user:\nwhat's the status?") {
		t.Fatalf("input = %q; want the context and the message", text)
	}
}

func TestRepointingTheAppStartsAFreshSession(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.runner.known["research"] = "Research Agent"
	fx.handle(t, fx.event(ActionCreated))
	app := proto.Clone(fx.app).(*agentsv1.LinearApp)
	app.AgentId = "research"
	if _, err := fx.repo.UpdateApp(t.Context(), "ws-a", app, fx.app.GetRevision()); err != nil {
		t.Fatalf("UpdateApp: %v", err)
	}
	fx.handle(t, fx.prompted("continue"))
	turns := fx.runner.turns()
	if turns[1].sessionID != "linear:app-1:session-1:research" || !strings.HasPrefix(turns[1].text, "Linear context:") {
		t.Fatalf("turn after repoint = %+v; want a fresh session with context", turns[1])
	}
}

func TestAdmission(t *testing.T) {
	for _, tc := range []struct {
		name      string
		allowlist []string
		user      string
		admitted  bool
	}{
		{"empty allowlist admits everyone", nil, linearUserTwo, true},
		{"listed user", []string{linearUserOne}, linearUserOne, true},
		{"unlisted user", []string{linearUserOne}, linearUserTwo, false},
		{"unknown user with an allowlist", []string{linearUserOne}, "", false},
		{"unknown user without an allowlist", nil, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newOrchestratorFixture(t, func(app *agentsv1.LinearApp) { app.AllowedUserIds = tc.allowlist })
			ev := fx.event(ActionCreated)
			ev.PromptingUserID = tc.user
			fx.handle(t, ev)
			ran := len(fx.runner.turns()) == 1
			if ran != tc.admitted {
				t.Fatalf("ran = %v, want %v", ran, tc.admitted)
			}
			if !tc.admitted {
				acts := fx.linear.Activities()
				if len(acts) != 1 || acts[0].Type != linearapi.ActivityError || !strings.Contains(acts[0].Body, "allowlist") {
					t.Fatalf("activities = %+v; want one allowlist error", acts)
				}
			}
		})
	}
}

func TestADisabledAppOrMissingAgentAnswersWithAnErrorAndRunsNothing(t *testing.T) {
	disabled := newOrchestratorFixture(t, func(app *agentsv1.LinearApp) { app.InboundEnabled = false })
	disabled.handle(t, disabled.event(ActionCreated))
	if acts := disabled.linear.Activities(); len(acts) != 1 || acts[0].Type != linearapi.ActivityError || !strings.Contains(acts[0].Body, "disabled") {
		t.Fatalf("disabled app activities = %+v", acts)
	}

	missing := newOrchestratorFixture(t, func(app *agentsv1.LinearApp) { app.AgentId = "unloaded" })
	missing.handle(t, missing.event(ActionCreated))
	if acts := missing.linear.Activities(); len(acts) != 1 || acts[0].Type != linearapi.ActivityError || !strings.Contains(acts[0].Body, "not available") {
		t.Fatalf("missing agent activities = %+v", acts)
	}
	if len(disabled.runner.turns())+len(missing.runner.turns()) != 0 {
		t.Fatal("an Agent ran")
	}
}

func TestADeletedAppIsSkippedSilently(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	if err := fx.repo.DeleteApp(t.Context(), "ws-a", "app-1"); err != nil {
		t.Fatalf("DeleteApp: %v", err)
	}
	fx.handle(t, fx.event(ActionCreated))
	if len(fx.linear.Activities()) != 0 || len(fx.runner.turns()) != 0 {
		t.Fatal("a deleted App produced activity")
	}
}

func TestFailuresAreReportedAsErrors(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.runner.err = errors.New("model exploded with key sk-ant-api03-abcdefghijklmnopqrstuvwxyz0123")
	fx.handle(t, fx.event(ActionCreated))
	acts := fx.linear.Activities()
	last := acts[len(acts)-1]
	if last.Type != linearapi.ActivityError || !strings.Contains(last.Body, "model exploded") {
		t.Fatalf("last activity = %+v; want the failure", last)
	}
	if strings.Contains(last.Body, "abcdefghijklmnop") {
		t.Fatalf("the error leaked a credential: %q", last.Body)
	}
}

func TestATurnPastItsMaxRunTimeIsReportedAsTimedOut(t *testing.T) {
	fx := newOrchestratorFixture(t, func(app *agentsv1.LinearApp) { app.MaxRunSeconds = proto.Int32(1) })
	fx.runner.block = make(chan struct{}) // never released
	fx.handle(t, fx.event(ActionCreated))
	acts := fx.linear.Activities()
	last := acts[len(acts)-1]
	if last.Type != linearapi.ActivityError || !strings.Contains(last.Body, "timed out") {
		t.Fatalf("last activity = %+v; want a timeout", last)
	}
}

func TestZeroMaxRunIsUnlimited(t *testing.T) {
	fx := newOrchestratorFixture(t, func(app *agentsv1.LinearApp) { app.MaxRunSeconds = proto.Int32(0) })
	fx.handle(t, fx.event(ActionCreated))
	if d := fx.runner.turns()[0].deadline; d != 0 {
		t.Fatalf("deadline = %v, want none", d)
	}
}

func TestLongResponsesAreTruncatedWithALinkAndCredentialsRedacted(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.runner.output = "token ghp_abcdefghijklmnopqrstuvwxyz0123456789 " + strings.Repeat("x", 9000)
	fx.handle(t, fx.event(ActionCreated))
	acts := fx.linear.Activities()
	body := acts[len(acts)-1].Body
	if n := len([]rune(body)); n > 8000 {
		t.Fatalf("response is %d runes, want at most 8000", n)
	}
	if !strings.Contains(body, "truncated") || !strings.Contains(body, "https://butter.test/sessions/detail") {
		t.Fatalf("truncated response should point at the full reply: %q", body[len(body)-200:])
	}
	if strings.Contains(body, "ghp_abcdef") {
		t.Fatal("the response leaked a credential")
	}
}

func TestAMarkedInstallationIsNotAnswered(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	if err := fx.repo.MarkInstallationNeedsReinstall(t.Context(), "ws-a", "inst-1", "revoked"); err != nil {
		t.Fatalf("mark: %v", err)
	}
	fx.handle(t, fx.event(ActionCreated))
	if len(fx.linear.Activities()) != 0 || len(fx.runner.turns()) != 0 {
		t.Fatal("a marked installation produced activity")
	}
}

func TestARejectedTokenMarksTheInstallationForReinstall(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.linear.OnActivity(func(lineartest.Activity) *lineartest.Failure {
		return &lineartest.Failure{GraphQLCode: "AUTHENTICATION_ERROR"}
	})
	fx.handle(t, fx.event(ActionCreated))
	inst, err := fx.repo.GetInstallation(t.Context(), "ws-a", "inst-1")
	if err != nil {
		t.Fatalf("GetInstallation: %v", err)
	}
	if inst.GetCredentialState() != agentsv1.LinearInstallationCredentialState_LINEAR_INSTALLATION_CREDENTIAL_STATE_NEEDS_REINSTALL {
		t.Fatalf("credential_state = %v, want NEEDS_REINSTALL", inst.GetCredentialState())
	}
}
