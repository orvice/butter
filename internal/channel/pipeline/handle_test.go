package pipeline

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// --- fakes ---------------------------------------------------------------

type runCall struct {
	agentName     string
	parts         []*genai.Part
	modelOverride string
	ctxInfo       *agentsv1.ContextInfo
	onEvent       runner.EventCallback
	onCompaction  runner.CompactionCallback
}

type fakeRunner struct {
	runResult *runner.TurnResult
	runErr    error
	runCalls  []runCall
	runHook   func(runCall)

	knownAgents  map[string]bool
	agentStatus  *runner.AgentStatus
	agentModel   string
	providers    []agentsv1.ModelProvider
	sess         session.Session
	sessErr      error
	clearErr     error
	clearedCalls int
}

func (f *fakeRunner) RunTurn(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (*runner.TurnResult, error) {
	call := runCall{agentName, parts, modelOverride, ctxInfo, onEvent, onCompaction}
	f.runCalls = append(f.runCalls, call)
	if f.runHook != nil {
		f.runHook(call)
	}
	return f.runResult, f.runErr
}

func (f *fakeRunner) HasAgentInWorkspace(workspaceID, name string) bool { return f.knownAgents[name] }
func (f *fakeRunner) GetAgentStatus(name string) *runner.AgentStatus    { return f.agentStatus }
func (f *fakeRunner) GetAgentModel(name string) string                  { return f.agentModel }
func (f *fakeRunner) ModelProviders() []agentsv1.ModelProvider          { return f.providers }
func (f *fakeRunner) GetSession(ctx context.Context, channelName, sessionID, userID string) (session.Session, error) {
	return f.sess, f.sessErr
}
func (f *fakeRunner) ClearSession(ctx context.Context, channelName, sessionID, userID string) error {
	f.clearedCalls++
	return f.clearErr
}

type fakeSelector struct {
	values  map[string]string
	getErr  error
	setErr  error
	setKeys []string
}

func newFakeSelector() *fakeSelector { return &fakeSelector{values: map[string]string{}} }

func (s *fakeSelector) Get(ctx context.Context, channelName, sessionID string) (string, error) {
	if s.getErr != nil {
		return "", s.getErr
	}
	return s.values[channelName+":"+sessionID], nil
}

func (s *fakeSelector) Set(ctx context.Context, channelName, sessionID, value string) error {
	if s.setErr != nil {
		return s.setErr
	}
	s.values[channelName+":"+sessionID] = value
	s.setKeys = append(s.setKeys, value)
	return nil
}

type fakeDebug struct {
	override    *bool
	toggleState bool
	toggleErr   error
	toggleCalls int
}

func (d *fakeDebug) Get(ctx context.Context, channelName, sessionID string) (*bool, error) {
	return d.override, nil
}

func (d *fakeDebug) Toggle(ctx context.Context, channelName, sessionID string, channelDefault bool) (bool, error) {
	d.toggleCalls++
	return d.toggleState, d.toggleErr
}

type fakeTransport struct {
	replies         []string
	typingCount     int
	debugStatus     []bool
	agentLists      [][]AgentChoice
	modelLists      [][]ModelChoice
	statusViews     []StatusView
	processingMsgs  []string
	processingDebug []*DebugSummary
	editedMsgs      []editedMsg
	debugEdits      []debugEdit
}

type editedMsg struct {
	messageID string
	agentName string
	text      string
	debug     *DebugSummary
}

type debugEdit struct {
	messageID string
	agentName string
	debug     DebugSummary
}

func (t *fakeTransport) SendReply(ctx context.Context, msg IncomingMessage, text string) {
	t.replies = append(t.replies, text)
}
func (t *fakeTransport) SendProcessing(ctx context.Context, msg IncomingMessage, agentName string, debug *DebugSummary) string {
	id := fmt.Sprintf("proc-%d", len(t.processingMsgs)+1)
	t.processingMsgs = append(t.processingMsgs, agentName)
	if debug == nil {
		t.processingDebug = append(t.processingDebug, nil)
	} else {
		copy := cloneDebugSummary(*debug)
		t.processingDebug = append(t.processingDebug, &copy)
	}
	return id
}
func (t *fakeTransport) EditDebug(ctx context.Context, msg IncomingMessage, messageID string, agentName string, debug DebugSummary) {
	t.debugEdits = append(t.debugEdits, debugEdit{messageID: messageID, agentName: agentName, debug: debug})
}
func (t *fakeTransport) EditReply(ctx context.Context, msg IncomingMessage, messageID string, agentName string, text string, debug *DebugSummary) {
	t.editedMsgs = append(t.editedMsgs, editedMsg{messageID: messageID, agentName: agentName, text: text, debug: debug})
}
func (t *fakeTransport) SendTyping(ctx context.Context, msg IncomingMessage) { t.typingCount++ }
func (t *fakeTransport) SendDebugStatus(ctx context.Context, msg IncomingMessage, active bool) {
	t.debugStatus = append(t.debugStatus, active)
}
func (t *fakeTransport) SendAgentList(ctx context.Context, msg IncomingMessage, choices []AgentChoice) {
	t.agentLists = append(t.agentLists, choices)
}
func (t *fakeTransport) SendModelList(ctx context.Context, msg IncomingMessage, choices []ModelChoice) {
	t.modelLists = append(t.modelLists, choices)
}
func (t *fakeTransport) SendStatus(ctx context.Context, msg IncomingMessage, view StatusView) {
	t.statusViews = append(t.statusViews, view)
}

// --- helpers -------------------------------------------------------------

func textParts(s string) func(context.Context) ([]*genai.Part, error) {
	return func(context.Context) ([]*genai.Part, error) {
		return []*genai.Part{genai.NewPartFromText(s)}, nil
	}
}

func newHarness(cfg Config) (*Handler, *fakeRunner, *fakeSelector, *fakeSelector, *fakeDebug, *fakeTransport) {
	r := &fakeRunner{knownAgents: map[string]bool{}}
	agentSel := newFakeSelector()
	modelSel := newFakeSelector()
	debug := &fakeDebug{}
	tr := &fakeTransport{}
	h := NewHandler(cfg, r, agentSel, modelSel, debug, tr)
	return h, r, agentSel, modelSel, debug, tr
}

func baseMsg() IncomingMessage {
	return IncomingMessage{
		SessionID:   "chat:1",
		UserID:      "42",
		ChatID:      "1",
		MessageID:   "7",
		Text:        "hello",
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
		ChannelType: "telegram",
		IsPrivate:   true,
		Metadata:    map[string]string{"username": "orx"},
		BuildParts:  textParts("hello"),
	}
}

// --- spine: plain message ------------------------------------------------

func TestHandle_PlainMessage_RunsAgentAndReplies(t *testing.T) {
	h, r, _, _, _, tr := newHarness(Config{
		ChannelName:  "tg",
		WorkspaceID:  "ws1",
		DefaultAgent: "assistant",
		DefaultModel: "gpt",
		ChannelType:  "telegram",
	})
	r.runResult = &runner.TurnResult{Output: "hi there"}

	h.Handle(context.Background(), baseMsg())

	if len(r.runCalls) != 1 {
		t.Fatalf("expected 1 runner.Run call, got %d", len(r.runCalls))
	}
	call := r.runCalls[0]
	if call.agentName != "assistant" {
		t.Errorf("agentName = %q, want assistant", call.agentName)
	}
	if call.modelOverride != "gpt" {
		t.Errorf("modelOverride = %q, want gpt", call.modelOverride)
	}
	if call.ctxInfo.GetWorkspaceId() != "ws1" || call.ctxInfo.GetSessionId() != "chat:1" {
		t.Errorf("ctxInfo mismatch: %+v", call.ctxInfo)
	}
	if call.ctxInfo.GetChannelType() != "telegram" || call.ctxInfo.GetChatId() != "1" {
		t.Errorf("ctxInfo channel/chat mismatch: %+v", call.ctxInfo)
	}
	if call.ctxInfo.GetMetadata()["username"] != "orx" {
		t.Errorf("ctxInfo metadata missing username: %+v", call.ctxInfo.GetMetadata())
	}
	if len(tr.processingMsgs) != 1 || tr.processingMsgs[0] != "assistant" {
		t.Errorf("processingMsgs = %v, want [assistant]", tr.processingMsgs)
	}
	if len(tr.editedMsgs) != 1 {
		t.Fatalf("expected 1 edited msg, got %d", len(tr.editedMsgs))
	}
	if tr.editedMsgs[0].text != "hi there" {
		t.Errorf("edited text = %q, want %q", tr.editedMsgs[0].text, "hi there")
	}
	if tr.editedMsgs[0].agentName != "assistant" {
		t.Errorf("edited agentName = %q, want %q", tr.editedMsgs[0].agentName, "assistant")
	}
}

func TestHandle_PlainMessage_RunnerErrorReplies(t *testing.T) {
	h, r, _, _, _, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	r.runErr = errors.New("boom")

	h.Handle(context.Background(), baseMsg())

	if len(tr.editedMsgs) != 1 {
		t.Fatalf("expected 1 edited error msg, got %d", len(tr.editedMsgs))
	}
	if tr.editedMsgs[0].text == "" {
		t.Errorf("expected non-empty error text in edited msg")
	}
}

func TestHandle_PlainMessage_EmptyResponseFallback(t *testing.T) {
	h, r, _, _, _, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	r.runResult = &runner.TurnResult{}

	h.Handle(context.Background(), baseMsg())

	if len(tr.editedMsgs) != 1 || tr.editedMsgs[0].text != emptyTurnResponse {
		t.Errorf("editedMsgs = %v, want text=%q", tr.editedMsgs, emptyTurnResponse)
	}
}

func TestHandle_PlainMessage_MaxTokensResponse(t *testing.T) {
	h, r, _, _, _, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	r.runResult = &runner.TurnResult{FinishReason: genai.FinishReasonMaxTokens}

	h.Handle(context.Background(), baseMsg())

	if len(tr.editedMsgs) != 1 || tr.editedMsgs[0].text != "The agent reached its response limit before producing a reply. Please try a shorter request." {
		t.Errorf("editedMsgs = %v", tr.editedMsgs)
	}
}

func TestHandle_PlainMessage_WhitespaceSafetyResponse(t *testing.T) {
	h, r, _, _, _, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	r.runResult = &runner.TurnResult{Output: "  \n", ErrorCode: "SAFETY"}

	h.Handle(context.Background(), baseMsg())

	if len(tr.editedMsgs) != 1 || tr.editedMsgs[0].text != "The agent could not provide a response because it was blocked by the model's safety checks. Please revise your message and try again." {
		t.Errorf("editedMsgs = %v", tr.editedMsgs)
	}
}

func TestActiveAgent_SelectionOverridesDefault(t *testing.T) {
	h, _, agentSel, _, _, _ := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	if got := h.ActiveAgent(context.Background(), "chat:1"); got != "assistant" {
		t.Errorf("default agent = %q, want assistant", got)
	}
	_ = agentSel.Set(context.Background(), "tg", "chat:1", "researcher")
	if got := h.ActiveAgent(context.Background(), "chat:1"); got != "researcher" {
		t.Errorf("selected agent = %q, want researcher", got)
	}
}
