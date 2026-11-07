package telegram

// End-to-end orchestration tests (issue #264/#268): one accepted update
// reaches the Agent exactly once and produces a reply in the originating
// Forum Topic, using the real sender and repository.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	telegrammemory "go.orx.me/apps/butter/internal/repo/telegram/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/runtime/sessionshare"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramapi/telegramtest"
	"go.orx.me/apps/butter/internal/telegramqueue"
	"go.orx.me/apps/butter/internal/telegramsend"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeAgentRunner records invocations instead of running ADK.
type fakeAgentRunner struct {
	mu                  sync.Mutex
	calls               []fakeAgentCall
	output              string
	failErr             error
	onRun               func()
	known               map[string]string
	modelOverrideLocked map[string]bool
}

type fakeAgentCall struct {
	agentName string
	sessionID string
	shared    bool // the turn may join a session ID other users hold
	model     string
	text      string
}

func (r *fakeAgentRunner) ResolveAgentRef(_ string, agentID string) (string, bool) {
	name, ok := r.known[agentID]
	return name, ok
}

func (r *fakeAgentRunner) SupportsModelOverride(_ string, agentID string) (bool, bool) {
	if _, ok := r.known[agentID]; !ok {
		return false, false
	}
	return !r.modelOverrideLocked[agentID], true
}

func (r *fakeAgentRunner) RunTurnSSE(ctx context.Context, agentName string, parts []*genai.Part,
	model string, ctxInfo *agentsv1.ContextInfo, _ runner.EventCallback,
	_ runner.CompactionCallback) (*runner.TurnResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.onRun != nil {
		r.onRun()
	}
	text := ""
	if len(parts) > 0 {
		text = parts[0].Text
	}
	r.calls = append(r.calls, fakeAgentCall{
		agentName: agentName, sessionID: ctxInfo.GetSessionId(), model: model, text: text,
		shared: sessionshare.Allowed(ctx),
	})
	if r.failErr != nil {
		return &runner.TurnResult{}, r.failErr
	}
	return &runner.TurnResult{Output: r.output}, nil
}

// fakeSessionClearer records /clear.
type fakeSessionClearer struct {
	deleted []string
	err     error
}

func (c *fakeSessionClearer) Delete(_ context.Context, req *session.DeleteRequest) error {
	c.deleted = append(c.deleted, req.SessionID)
	return c.err
}

type orchestratorFixture struct {
	orchestrator *Orchestrator
	repo         *telegrammemory.Store
	bots         *telegramtest.Fake
	agents       *fakeAgentRunner
	sessions     *fakeSessionClearer
	prefs        *MemoryPreferenceStore
	dest         *agentsv1.TelegramDestination
}

func newOrchestratorFixture(t *testing.T, mutate func(*agentsv1.TelegramDestinationConfig)) *orchestratorFixture {
	t.Helper()
	repo := telegrammemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	ciphertext, keyID, err := keyring.Encrypt(t.Context(), []byte("111111:token"))
	if err != nil {
		t.Fatalf("encrypt token: %v", err)
	}
	if _, err := repo.CreateChannel(t.Context(), "ws-a", &agentsv1.TelegramChannel{
		Id: "ch-1", Key: "main", BotId: "111111", BotUsername: botUsername,
		InboundEnabled: true, OutboundEnabled: true,
	}, telegramrepo.ChannelCredentials{
		BotToken: telegramrepo.Credential{Ciphertext: ciphertext, KeyID: keyID},
	}); err != nil {
		t.Fatalf("create channel: %v", err)
	}
	seed := destination(mutate)
	stored, err := repo.CreateDestination(t.Context(), "ws-a", seed)
	if err != nil {
		t.Fatalf("create destination: %v", err)
	}

	bots := telegramtest.NewFake().Register("111111:token", telegramtest.Identity(111111, botUsername))
	sender := telegramsend.New(repo, keyring, bots.Factory())
	agents := &fakeAgentRunner{output: "agent reply", known: map[string]string{"support": "Support Agent"}}
	sessions := &fakeSessionClearer{}

	prefs := NewMemoryPreferenceStore()
	orchestrator := NewOrchestrator(repo, sender, agents)
	orchestrator.SetSessionClearer(sessions)
	orchestrator.SetPreferenceStore(prefs)
	return &orchestratorFixture{
		orchestrator: orchestrator, repo: repo, bots: bots,
		agents: agents, sessions: sessions, dest: stored, prefs: prefs,
	}
}

// eventForStored wraps an update against the stored Destination's revision.
func (fx *orchestratorFixture) eventForStored(raw string) *telegramqueue.Event {
	event := eventFor(raw)
	event.DestinationID = fx.dest.GetId()
	event.DestinationRevision = fx.dest.GetRevision()
	event.Policy = snapshotPolicy(fx.dest.GetConfig())
	return event
}

func TestOneAcceptedUpdateReachesTheAgentOnceAndRepliesInTheTopic(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.agents.onRun = func() {
		reactions := fx.bots.Reactions()
		if len(reactions) != 1 || reactions[0].Params.MessageID != "9" || reactions[0].Params.Emoji != "👀" {
			t.Errorf("expected a reaction on the inbound message before the Agent runs, got %+v", reactions)
		}
		if len(fx.bots.Sent()) != 0 {
			t.Error("no message should be sent before the Agent finishes")
		}
	}

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(message(realUser, "hello", ""))); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(fx.agents.calls) != 1 {
		t.Fatalf("agent invoked %d times, want exactly once", len(fx.agents.calls))
	}
	call := fx.agents.calls[0]
	if call.agentName != "Support Agent" || call.text != "hello" {
		t.Errorf("unexpected invocation: %+v", call)
	}
	// Every member of a Destination holds its session ID, so the turn must
	// be allowed to join a session another member started.
	if !call.shared {
		t.Error("the Telegram turn must be allowed to share its session ID")
	}

	// Only the completed answer is sent, as a fresh, non-silent reply.
	sent := fx.bots.Sent()
	if len(sent) != 1 {
		t.Fatalf("sent %d messages, want only the completed reply", len(sent))
	}
	if sent[0].Params.MessageThreadID != "42" {
		t.Errorf("reply left the topic: thread %q", sent[0].Params.MessageThreadID)
	}
	if sent[0].Params.ReplyToMessageID != "9" {
		t.Errorf("reply_to = %q, want the inbound message quoted", sent[0].Params.ReplyToMessageID)
	}
	if sent[0].Edit != nil || sent[0].Params.DisableNotification {
		t.Fatal("the answer must be a new, non-silent message")
	}
	if !strings.Contains(sent[0].Params.Text, "agent reply") {
		t.Errorf("reply text = %q, want the agent output", sent[0].Params.Text)
	}
}

// An agent failure must still be reported in the originating topic, not
// swallowed into silence.
func TestAgentFailureIsReportedInTheTopic(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.agents.failErr = errors.New("model unavailable")

	err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "hello", "")))
	if err == nil {
		t.Fatal("expected the failure to surface so the event is retried")
	}
	sent := fx.bots.Sent()
	if len(sent) != 1 {
		t.Fatalf("sent %d messages, want only a failure notice", len(sent))
	}
	if sent[0].Params.MessageThreadID != "42" || sent[0].Params.ReplyToMessageID != "9" {
		t.Errorf("failure notice lost the topic or reply target: %+v", sent[0].Params)
	}
	if sent[0].Edit != nil || sent[0].Params.DisableNotification || !strings.Contains(sent[0].Params.Text, "could not complete") {
		t.Errorf("failure notice = %+v", sent[0])
	}
}

func TestReactionFailureDoesNotPreventAgentReply(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	fx.bots.OnReaction(func(context.Context, telegramapi.ReactionParams) error {
		return &telegramapi.APIError{Code: 400, Description: "REACTION_INVALID"}
	})
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "hello", ""))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(fx.agents.calls) != 1 || len(fx.bots.Sent()) != 1 {
		t.Fatal("reaction rejection must not prevent invocation or delivery")
	}
	if got := fx.bots.Sent()[0]; got.Edit != nil || !strings.Contains(got.Params.Text, "agent reply") {
		t.Fatalf("expected a fresh answer despite reaction failure: %+v", got)
	}
}

func TestReactionTargetsInboundMessageWhenRepliesDoNotQuote(t *testing.T) {
	fx := newOrchestratorFixture(t, func(c *agentsv1.TelegramDestinationConfig) {
		c.ReplyMode = agentsv1.TelegramReplyMode_TELEGRAM_REPLY_MODE_NEW_MESSAGE
	})
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "hello", ""))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	reactions := fx.bots.Reactions()
	if len(reactions) != 1 || reactions[0].Params.MessageID != "9" {
		t.Fatalf("reaction must still target the inbound message: %+v", reactions)
	}
	sent := fx.bots.Sent()
	if len(sent) != 1 || sent[0].Params.ReplyToMessageID != "" || sent[0].Params.MessageThreadID != "42" {
		t.Fatalf("standalone reply lost its configured mode or topic: %+v", sent)
	}
}

// /status is available to any admitted user and invokes no Agent.
func TestStatusIsAvailableWithoutControllerRights(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)

	raw := message(realUser, "/status", `"entities":[{"type":"bot_command","offset":0,"length":7}]`)
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(raw)); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(fx.agents.calls) != 0 {
		t.Fatal("/status must not invoke an agent")
	}
	sent := fx.bots.Sent()
	if len(sent) != 1 {
		t.Fatalf("sent %d messages", len(sent))
	}
	if sent[0].Params.MessageThreadID != "42" {
		t.Errorf("status answered outside the topic: %q", sent[0].Params.MessageThreadID)
	}
}

func TestClearRequiresAController(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	raw := message(realUser, "/clear", `"entities":[{"type":"bot_command","offset":0,"length":6}]`)

	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(raw)); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(fx.sessions.deleted) != 0 {
		t.Fatal("a non-controller cleared the shared session")
	}
	// The refusal is still answered in the topic rather than silently dropped.
	if len(fx.bots.Sent()) != 1 {
		t.Fatalf("expected the refusal to be delivered")
	}
}

func TestControllerCanClearTheSession(t *testing.T) {
	fx := newOrchestratorFixture(t, func(c *agentsv1.TelegramDestinationConfig) {
		c.ControllerUserIds = []string{"7"}
	})
	raw := message(realUser, "/clear", `"entities":[{"type":"bot_command","offset":0,"length":6}]`)

	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(raw)); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(fx.sessions.deleted) != 1 {
		t.Fatalf("cleared %d sessions, want 1", len(fx.sessions.deleted))
	}
	if fx.sessions.deleted[0] != SessionID("ch-1", fx.dest.GetId(), "d"+fx.dest.GetId(), "support") {
		t.Errorf("cleared session %q", fx.sessions.deleted[0])
	}
}

// A Destination disabled after acceptance must not run an Agent, even though
// the queued snapshot said it was enabled.
func TestDisabledAfterAcceptanceDoesNotInvokeTheAgent(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	event := fx.eventForStored(message(realUser, "hello", ""))

	dest, err := fx.repo.GetDestination(t.Context(), "ws-a", fx.dest.GetId())
	if err != nil {
		t.Fatalf("GetDestination: %v", err)
	}
	dest.InboundEnabled = false
	if _, err := fx.repo.UpdateDestination(t.Context(), "ws-a", dest, dest.GetRevision()); err != nil {
		t.Fatalf("disable destination: %v", err)
	}

	if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(fx.agents.calls) != 0 {
		t.Fatal("a disabled destination invoked the agent")
	}
	if len(fx.bots.Sent()) != 0 {
		t.Fatal("a disabled destination produced a reply")
	}
}

// A deleted Destination is terminal: retrying cannot bring it back.
func TestDeletedDestinationIsAcknowledged(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	event := fx.eventForStored(message(realUser, "hello", ""))
	if err := fx.repo.DeleteDestination(t.Context(), "ws-a", fx.dest.GetId()); err != nil {
		t.Fatalf("DeleteDestination: %v", err)
	}

	if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
		t.Fatalf("Handle must acknowledge a deleted destination, got %v", err)
	}
	if len(fx.agents.calls) != 0 || len(fx.bots.Sent()) != 0 {
		t.Fatal("a deleted destination produced work")
	}
}

// The Destination's explicit model override reaches the runner.
func TestModelOverrideIsPassedThrough(t *testing.T) {
	fx := newOrchestratorFixture(t, func(c *agentsv1.TelegramDestinationConfig) { c.Model = "pro" })

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(message(realUser, "hello", ""))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if fx.agents.calls[0].model != "pro" {
		t.Errorf("model = %q", fx.agents.calls[0].model)
	}
}
