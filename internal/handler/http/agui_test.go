package http

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/aguitool"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func setupAGUIRouter(repo configrepo.AgentRepository, runnerSvc AGUIRunnerService, authenticated bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if authenticated {
		// Inject workspace context as if the auth middleware ran.
		r.Use(func(c *gin.Context) {
			c.Request = c.Request.WithContext(wsctx.WithID(c.Request.Context(), "test-workspace"))
			c.Next()
		})
	}
	h := NewAGUIHandler(repo)
	if runnerSvc != nil {
		h.SetRunnerService(runnerSvc)
	}
	h.Register(r)
	return r
}

func aguiEnabledRepo() *stubAgentRepo {
	return &stubAgentRepo{agents: []*agentsv1.Agent{
		{Name: "Writer", AgentId: "writer", EnableAgui: true},
		{Name: "Hidden", AgentId: "hidden", EnableAgui: false},
	}}
}

// postAGUI issues a run request and returns the recorder.
func postAGUI(t *testing.T, router *gin.Engine, agentID string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	switch v := body.(type) {
	case string:
		payload = []byte(v)
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		payload = encoded
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/agui/"+agentID, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

// minimalAGUIBody is a valid Phase 1 request body.
func minimalAGUIBody(threadID, text string) map[string]any {
	return map[string]any{
		"threadId": threadID,
		"runId":    "run-1",
		"messages": []map[string]any{{"id": "m1", "role": "user", "content": text}},
		"tools":    []any{},
		"context":  []any{},
		"state":    nil,
	}
}

func TestAGUIRun_StreamsEventsForEnabledAgent(t *testing.T) {
	mock := &mockRunner{runResult: "hello there"}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	w := postAGUI(t, router, "writer", minimalAGUIBody("t-1", "hi"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}

	body := w.Body.String()
	for _, want := range []string{
		`"type":"RUN_STARTED"`, `"threadId":"t-1"`, `"runId":"run-1"`,
		`"type":"TEXT_MESSAGE_CONTENT"`, `"delta":"hello there"`,
		`"type":"RUN_FINISHED"`, `"outcome":{"type":"success"}`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("stream missing %s\n---\n%s", want, body)
		}
	}

	// The agent is addressed by agent_id but the runner is driven by name.
	if mock.lastAgentName != "Writer" {
		t.Errorf("runner agent = %q, want Writer", mock.lastAgentName)
	}
	// threadId must be namespaced, never used as a bare session ID.
	if got := mock.lastCtxInfo.GetSessionId(); got != "agui-t-1" {
		t.Errorf("session id = %q, want agui-t-1", got)
	}
	if got := mock.lastCtxInfo.GetWorkspaceId(); got != "test-workspace" {
		t.Errorf("workspace = %q", got)
	}
	if len(mock.lastParts) != 1 || mock.lastParts[0].Text != "hi" {
		t.Errorf("parts = %+v, want the trailing user message only", mock.lastParts)
	}
}

// Only the trailing user message is forwarded — the server-side session is
// authoritative, so replaying the client's history would duplicate it.
func TestAGUIRun_SendsOnlyTrailingUserMessage(t *testing.T) {
	mock := &mockRunner{runResult: "ok"}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	body := minimalAGUIBody("t-1", "ignored")
	body["messages"] = []map[string]any{
		{"id": "m1", "role": "user", "content": "first"},
		{"id": "m2", "role": "assistant", "content": "reply"},
		{"id": "m3", "role": "user", "content": "latest"},
	}
	if w := postAGUI(t, router, "writer", body); w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if len(mock.lastParts) != 1 || mock.lastParts[0].Text != "latest" {
		t.Fatalf("parts = %+v, want only the latest user message", mock.lastParts)
	}
}

// Multimodal content fragments are accepted; Phase 1 keeps the text.
func TestAGUIRun_AcceptsContentFragments(t *testing.T) {
	mock := &mockRunner{runResult: "ok"}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	body := minimalAGUIBody("t-1", "")
	body["messages"] = []map[string]any{{"id": "m1", "role": "user", "content": []map[string]any{
		{"type": "text", "text": "line one"},
		{"type": "text", "text": "line two"},
	}}}
	if w := postAGUI(t, router, "writer", body); w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if len(mock.lastParts) != 1 || mock.lastParts[0].Text != "line one\nline two" {
		t.Fatalf("parts = %+v", mock.lastParts)
	}
}

// A resume request is translated into the FunctionResponse the workflow engine
// resumes on, addressed by interruptId so it survives interrupt.Resume's
// oldest-first fallback untouched.
func TestAGUIRun_ResumeBecomesAddressedFunctionResponse(t *testing.T) {
	mock := &mockRunner{runResult: "resumed"}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	body := map[string]any{
		"threadId": "t-1",
		"runId":    "run-2",
		"messages": []any{},
		"tools":    []any{},
		"resume": []map[string]any{
			{"interruptId": "int-7", "status": "resolved", "payload": "approved"},
		},
	}
	if w := postAGUI(t, router, "writer", body); w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if len(mock.lastParts) != 1 {
		t.Fatalf("parts = %+v, want 1", mock.lastParts)
	}
	fr := mock.lastParts[0].FunctionResponse
	if fr == nil {
		t.Fatalf("part carries no FunctionResponse: %+v", mock.lastParts[0])
	}
	if fr.ID != "int-7" || fr.Name != workflow.WorkflowInputFunctionCallName {
		t.Errorf("function response = %+v", fr)
	}
	if got := fr.Response["payload"]; got != "approved" {
		t.Errorf("payload = %v, want approved", got)
	}
}

func TestAGUIRun_ReportsRunFailureAsRunError(t *testing.T) {
	mock := &mockRunner{runErr: errors.New("model exploded")}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	w := postAGUI(t, router, "writer", minimalAGUIBody("t-1", "hi"))
	// Headers are already committed when the run fails, so the failure is
	// in-band rather than an HTTP error status.
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"type":"RUN_ERROR"`) || !strings.Contains(body, "model exploded") {
		t.Fatalf("stream missing RUN_ERROR:\n%s", body)
	}
}

func TestAGUIRun_Rejections(t *testing.T) {
	valid := minimalAGUIBody("t-1", "hi")

	tests := []struct {
		name       string
		agentID    string
		body       any
		runner     AGUIRunnerService
		noWorkspce bool
		wantStatus int
		wantError  string
	}{
		{
			name: "unknown agent", agentID: "nope", body: valid,
			runner: &mockRunner{}, wantStatus: http.StatusNotFound,
		},
		{
			name: "agent not opted in", agentID: "hidden", body: valid,
			runner: &mockRunner{}, wantStatus: http.StatusNotFound,
		},
		{
			name: "no workspace", agentID: "writer", body: valid,
			runner: &mockRunner{}, noWorkspce: true, wantStatus: http.StatusUnauthorized,
		},
		{
			name: "runner unavailable", agentID: "writer", body: valid,
			runner: nil, wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "malformed body", agentID: "writer", body: "{not json",
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest,
		},
		{
			name:    "missing threadId",
			agentID: "writer",
			body: map[string]any{
				"runId":    "run-1",
				"messages": []map[string]any{{"id": "m1", "role": "user", "content": "hi"}},
			},
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "threadId",
		},
		{
			name:    "tool without a name",
			agentID: "writer",
			body: withAGUIField(valid, "tools", []map[string]any{
				{"name": "  ", "description": "ask", "parameters": map[string]any{}},
			}),
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "name",
		},
		{
			name:    "duplicate tool names",
			agentID: "writer",
			body: withAGUIField(valid, "tools", []map[string]any{
				{"name": "confirm", "description": "a", "parameters": map[string]any{}},
				{"name": "confirm", "description": "b", "parameters": map[string]any{}},
			}),
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "duplicate tool name",
		},
		{
			name:    "tool result without toolCallId",
			agentID: "writer",
			body: withAGUIField(valid, "messages", []map[string]any{
				{"id": "m1", "role": "user", "content": "hi"},
				{"id": "m2", "role": "tool", "content": "42"},
			}),
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "toolCallId",
		},
		{
			name:    "duplicate tool results",
			agentID: "writer",
			body: withAGUIField(valid, "messages", []map[string]any{
				{"id": "m2", "role": "tool", "toolCallId": "call-1", "content": "a"},
				{"id": "m3", "role": "tool", "toolCallId": "call-1", "content": "b"},
			}),
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "duplicate tool result",
		},
		{
			name:    "state must be an object",
			agentID: "writer",
			body:    withAGUIField(valid, "state", "not-an-object"),
			runner:  &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "state",
		},
		{
			name:    "state needs a session service",
			agentID: "writer",
			body:    withAGUIField(valid, "state", map[string]any{"count": 1}),
			runner:  &mockRunner{}, wantStatus: http.StatusServiceUnavailable, wantError: "state",
		},
		{
			name:    "cancelling an interrupt",
			agentID: "writer",
			body: withAGUIField(valid, "resume", []map[string]any{
				{"interruptId": "int-1", "status": "cancelled"},
			}),
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "cancel",
		},
		{
			name:    "resume without interruptId",
			agentID: "writer",
			body: withAGUIField(valid, "resume", []map[string]any{
				{"status": "resolved", "payload": "yes"},
			}),
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "interruptId",
		},
		{
			name:    "no user message",
			agentID: "writer",
			body: withAGUIField(valid, "messages", []map[string]any{
				{"id": "m1", "role": "assistant", "content": "hi"},
			}),
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "user message",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := setupAGUIRouter(aguiEnabledRepo(), tc.runner, !tc.noWorkspce)
			w := postAGUI(t, router, tc.agentID, tc.body)
			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantError != "" && !strings.Contains(w.Body.String(), tc.wantError) {
				t.Errorf("error body %s, want it to mention %q", w.Body.String(), tc.wantError)
			}
			// A rejection must never open a stream.
			if strings.Contains(w.Body.String(), "RUN_STARTED") {
				t.Errorf("rejection emitted stream events: %s", w.Body.String())
			}
		})
	}
}

// An empty state object means "no state", not "unsupported state".
func TestAGUIRun_EmptyStateObjectIsAccepted(t *testing.T) {
	router := setupAGUIRouter(aguiEnabledRepo(), &mockRunner{runResult: "ok"}, true)
	body := withAGUIField(minimalAGUIBody("t-1", "hi"), "state", map[string]any{})
	if w := postAGUI(t, router, "writer", body); w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

// A missing runId is filled in server-side so every event still correlates.
func TestAGUIRun_GeneratesMissingRunID(t *testing.T) {
	router := setupAGUIRouter(aguiEnabledRepo(), &mockRunner{runResult: "ok"}, true)
	body := minimalAGUIBody("t-1", "hi")
	delete(body, "runId")

	w := postAGUI(t, router, "writer", body)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), `"runId":""`) {
		t.Fatalf("empty runId in stream:\n%s", w.Body.String())
	}
}

// fakeSessionGuard records Acquire calls and simulates the cross-Pod lease.
type fakeSessionGuard struct {
	mu       sync.Mutex
	busy     bool
	err      error
	loseCtx  bool
	keys     []string
	releases int
}

func (g *fakeSessionGuard) Acquire(ctx context.Context, key string) (context.Context, func(), bool, error) {
	g.mu.Lock()
	g.keys = append(g.keys, key)
	g.mu.Unlock()
	if g.err != nil {
		return ctx, func() {}, false, g.err
	}
	if g.busy {
		return ctx, func() {}, false, nil
	}
	if g.loseCtx {
		lost, cancel := context.WithCancel(ctx)
		cancel()
		ctx = lost
	}
	return ctx, func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.releases++
	}, true, nil
}

// ctxEchoRunner fails with the run context's error, simulating a run torn
// down by lease loss.
type ctxEchoRunner struct{ mockRunner }

func (r *ctxEchoRunner) RunSSE(ctx context.Context, agentName string, parts []*genai.Part, model string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return r.mockRunner.RunSSE(ctx, agentName, parts, model, ctxInfo, onEvent, onCompaction)
}

func setupGuardedAGUIRouter(runnerSvc AGUIRunnerService, guard *fakeSessionGuard) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(wsctx.WithID(c.Request.Context(), "test-workspace"))
		c.Next()
	})
	h := NewAGUIHandler(aguiEnabledRepo())
	h.SetRunnerService(runnerSvc)
	h.SetSessionGuard(guard)
	h.Register(r)
	return r
}

// A held session lease is a pre-stream 409, never a stream, and the runner is
// not invoked.
func TestAGUIRun_BusyThreadIsConflict(t *testing.T) {
	mock := &mockRunner{runResult: "should not run"}
	guard := &fakeSessionGuard{busy: true}
	router := setupGuardedAGUIRouter(mock, guard)

	w := postAGUI(t, router, "writer", minimalAGUIBody("t-1", "hi"))
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (body %s)", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "RUN_STARTED") {
		t.Fatalf("busy rejection opened a stream: %s", w.Body.String())
	}
	if mock.lastAgentName != "" {
		t.Fatal("runner was invoked despite a held lease")
	}
	// The lease key covers the caller and the namespaced session, so two users
	// sharing a threadId never serialize against each other.
	if len(guard.keys) != 1 || guard.keys[0] != "agui-user:agui-t-1" {
		t.Fatalf("lease keys = %v", guard.keys)
	}
}

// A guard failure (Redis down) refuses the run rather than silently dropping
// cross-Pod serialization.
func TestAGUIRun_GuardErrorIsServiceUnavailable(t *testing.T) {
	mock := &mockRunner{runResult: "should not run"}
	guard := &fakeSessionGuard{err: errors.New("redis down")}
	router := setupGuardedAGUIRouter(mock, guard)

	w := postAGUI(t, router, "writer", minimalAGUIBody("t-1", "hi"))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 (body %s)", w.Code, w.Body.String())
	}
	if mock.lastAgentName != "" {
		t.Fatal("runner was invoked despite guard failure")
	}
}

// The lease is released after a successful run and after a failed one.
func TestAGUIRun_ReleasesLease(t *testing.T) {
	for name, mock := range map[string]*mockRunner{
		"success": {runResult: "done"},
		"failure": {runErr: errors.New("model exploded")},
	} {
		t.Run(name, func(t *testing.T) {
			guard := &fakeSessionGuard{}
			router := setupGuardedAGUIRouter(mock, guard)
			if w := postAGUI(t, router, "writer", minimalAGUIBody("t-1", "hi")); w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			if guard.releases != 1 {
				t.Fatalf("releases = %d, want 1", guard.releases)
			}
		})
	}
}

// Losing the lease mid-run cancels the run and is reported in-band as a named
// RUN_ERROR, not a bare context cancellation.
func TestAGUIRun_LeaseLossBecomesRunError(t *testing.T) {
	guard := &fakeSessionGuard{loseCtx: true}
	router := setupGuardedAGUIRouter(&ctxEchoRunner{}, guard)

	w := postAGUI(t, router, "writer", minimalAGUIBody("t-1", "hi"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (stream already open)", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"type":"RUN_ERROR"`) || !strings.Contains(body, "session lease lost") {
		t.Fatalf("stream missing lease-loss RUN_ERROR:\n%s", body)
	}
	if guard.releases != 1 {
		t.Fatalf("releases = %d, want 1", guard.releases)
	}
}

// --- frontend tools ---

// ctxCaptureRunner records the run context so tests can assert what rode it.
type ctxCaptureRunner struct {
	mockRunner
	lastCtx context.Context
}

func (r *ctxCaptureRunner) RunSSE(ctx context.Context, agentName string, parts []*genai.Part, model string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error) {
	r.lastCtx = ctx
	return r.mockRunner.RunSSE(ctx, agentName, parts, model, ctxInfo, onEvent, onCompaction)
}

// fakeADKSession is a minimal session.Session carrying a fixed event list and
// state map.
type fakeADKSession struct {
	events []*adksession.Event
	state  map[string]any
}

func (f *fakeADKSession) ID() string                { return "s1" }
func (f *fakeADKSession) AppName() string           { return "agui" }
func (f *fakeADKSession) UserID() string            { return "agui-user" }
func (f *fakeADKSession) State() adksession.State   { return fakeADKState(f.state) }
func (f *fakeADKSession) LastUpdateTime() time.Time { return time.Time{} }
func (f *fakeADKSession) Events() adksession.Events { return fakeADKEvents(f.events) }

type fakeADKState map[string]any

func (s fakeADKState) Get(k string) (any, error) { return s[k], nil }
func (s fakeADKState) Set(k string, v any) error { s[k] = v; return nil }
func (s fakeADKState) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range s {
			if !yield(k, v) {
				return
			}
		}
	}
}

type fakeADKEvents []*adksession.Event

func (e fakeADKEvents) Len() int                   { return len(e) }
func (e fakeADKEvents) At(i int) *adksession.Event { return e[i] }
func (e fakeADKEvents) All() iter.Seq[*adksession.Event] {
	return func(yield func(*adksession.Event) bool) {
		for _, ev := range e {
			if !yield(ev) {
				return
			}
		}
	}
}

// fakeSessionService serves one fixed session; every other method panics via
// the embedded nil interface, keeping the fake honest about what runs.
type fakeSessionService struct {
	adksession.Service
	sess   adksession.Session
	getErr error
}

func (f *fakeSessionService) Get(_ context.Context, _ *adksession.GetRequest) (*adksession.GetResponse, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &adksession.GetResponse{Session: f.sess}, nil
}

// pendingToolCallSession returns a session whose last run paused on frontend
// tool call-1 ("confirm").
func pendingToolCallSession() *fakeADKSession {
	ev := &adksession.Event{}
	ev.LongRunningToolIDs = []string{"call-1"}
	ev.Content = &genai.Content{Parts: []*genai.Part{{
		FunctionCall: &genai.FunctionCall{Name: "confirm", ID: "call-1", Args: map[string]any{"q": "sure?"}},
	}}}
	return &fakeADKSession{events: []*adksession.Event{ev}}
}

func toolResultBody(threadID, toolCallID, content string) map[string]any {
	return map[string]any{
		"threadId": threadID,
		"runId":    "run-2",
		"messages": []map[string]any{
			{"id": "m1", "role": "user", "content": "hi"},
			{"id": "m2", "role": "assistant", "toolCalls": []map[string]any{
				{"id": toolCallID, "type": "function", "function": map[string]any{"name": "confirm", "arguments": "{}"}},
			}},
			{"id": "m3", "role": "tool", "toolCallId": toolCallID, "content": content},
		},
	}
}

// Client tool declarations ride the run context for the aguitool toolset.
func TestAGUIRun_ClientToolsReachTheRun(t *testing.T) {
	mock := &ctxCaptureRunner{mockRunner: mockRunner{runResult: "ok"}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(wsctx.WithID(c.Request.Context(), "test-workspace"))
		c.Next()
	})
	h := NewAGUIHandler(aguiEnabledRepo())
	h.SetRunnerService(mock)
	h.Register(r)

	body := withAGUIField(minimalAGUIBody("t-1", "hi"), "tools", []map[string]any{
		{"name": "confirm", "description": "ask the user", "parameters": map[string]any{
			"type": "object", "properties": map[string]any{"q": map[string]any{"type": "string"}},
		}},
	})
	if w := postAGUI(t, r, "writer", body); w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	decls := aguitool.ClientToolsFrom(mock.lastCtx)
	if len(decls) != 1 || decls[0].Name != "confirm" || decls[0].Description != "ask the user" {
		t.Fatalf("declarations on run context = %+v", decls)
	}
	if decls[0].Parameters == nil {
		t.Fatal("tool parameters schema was dropped")
	}
}

// A trailing tool-role message answers the pending call: it becomes the
// FunctionResponse ADK resumes on, named from the session's record.
func TestAGUIRun_ToolResultBecomesFunctionResponse(t *testing.T) {
	mock := &mockRunner{runResult: "done"}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(wsctx.WithID(c.Request.Context(), "test-workspace"))
		c.Next()
	})
	h := NewAGUIHandler(aguiEnabledRepo())
	h.SetRunnerService(mock)
	h.SetSessionService(&fakeSessionService{sess: pendingToolCallSession()})
	h.Register(r)

	w := postAGUI(t, r, "writer", toolResultBody("t-1", "call-1", `{"approved":true}`))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if len(mock.lastParts) != 1 {
		t.Fatalf("parts = %+v, want 1", mock.lastParts)
	}
	fr := mock.lastParts[0].FunctionResponse
	if fr == nil || fr.ID != "call-1" || fr.Name != "confirm" {
		t.Fatalf("function response = %+v", fr)
	}
	if got := fr.Response["approved"]; got != true {
		t.Errorf("response = %v, want the decoded JSON object", fr.Response)
	}

	// A plain-string result is wrapped rather than dropped.
	mock.lastParts = nil
	w = postAGUI(t, r, "writer", toolResultBody("t-1", "call-1", "looks good"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if fr := mock.lastParts[0].FunctionResponse; fr.Response["result"] != "looks good" {
		t.Fatalf("wrapped response = %+v", fr.Response)
	}
}

// A result for a call the session is not waiting on is rejected before the
// runner is invoked — nothing is rerun.
func TestAGUIRun_UnknownToolResultRejected(t *testing.T) {
	mock := &mockRunner{runResult: "should not run"}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(wsctx.WithID(c.Request.Context(), "test-workspace"))
		c.Next()
	})
	h := NewAGUIHandler(aguiEnabledRepo())
	h.SetRunnerService(mock)
	h.SetSessionService(&fakeSessionService{sess: &fakeADKSession{}})
	h.Register(r)

	w := postAGUI(t, r, "writer", toolResultBody("t-1", "call-404", "42"))
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "call-404") {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if mock.lastAgentName != "" {
		t.Fatal("runner was invoked for an unknown tool result")
	}
}

// Without a session store there is no way to validate results, so they are
// refused rather than half-trusted.
func TestAGUIRun_ToolResultsNeedSessionService(t *testing.T) {
	router := setupAGUIRouter(aguiEnabledRepo(), &mockRunner{}, true)
	w := postAGUI(t, router, "writer", toolResultBody("t-1", "call-1", "42"))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 (body %s)", w.Code, w.Body.String())
	}
}

// --- shared state ---

func setupStatefulAGUIRouter(runnerSvc AGUIRunnerService, svc adksession.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(wsctx.WithID(c.Request.Context(), "test-workspace"))
		c.Next()
	})
	h := NewAGUIHandler(aguiEnabledRepo())
	h.SetRunnerService(runnerSvc)
	h.SetSessionService(svc)
	h.Register(r)
	return r
}

// The client's mirror is validated against the authoritative state: absent or
// diverged mirrors get a corrective STATE_SNAPSHOT right after RUN_STARTED; a
// matching mirror gets none. Hidden ADK scopes never leave the server.
func TestAGUIRun_StateSnapshotOnBaselineAndConflict(t *testing.T) {
	sess := &fakeADKSession{state: map[string]any{
		"draft":       "v1",
		"temp:cursor": 3,
		"user:theme":  "dark",
	}}

	cases := []struct {
		name         string
		state        any
		wantSnapshot bool
	}{
		{"absent mirror is baselined", nil, true},
		{"diverged mirror is corrected", map[string]any{"draft": "edited"}, true},
		{"matching mirror stays quiet", map[string]any{"draft": "v1"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := setupStatefulAGUIRouter(&mockRunner{runResult: "ok"},
				&fakeSessionService{sess: sess})
			w := postAGUI(t, router, "writer", withAGUIField(minimalAGUIBody("t-1", "hi"), "state", tc.state))
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			body := w.Body.String()
			if got := strings.Contains(body, `"type":"STATE_SNAPSHOT"`); got != tc.wantSnapshot {
				t.Fatalf("snapshot emitted = %v, want %v\n%s", got, tc.wantSnapshot, body)
			}
			if tc.wantSnapshot {
				if !strings.Contains(body, `"draft":"v1"`) {
					t.Errorf("snapshot missing authoritative value:\n%s", body)
				}
				if strings.Contains(body, "temp:cursor") || strings.Contains(body, "user:theme") {
					t.Errorf("hidden state scope leaked:\n%s", body)
				}
			}
		})
	}
}

// State written during the run that the stream never carried (output_key
// lands on final events) surfaces as a trailing STATE_DELTA before
// RUN_FINISHED, read back from the authoritative session.
func TestAGUIRun_TrailingStateDelta(t *testing.T) {
	svc := &fakeSessionService{sess: &fakeADKSession{state: map[string]any{}}}
	// The runner mutates state as a side effect of the turn, as output_key
	// would; the second session Get sees it.
	mock := &mockRunner{runResult: "ok", onEventFn: func(runner.EventCallback) {
		svc.sess = &fakeADKSession{state: map[string]any{"summary": "done"}}
	}}
	router := setupStatefulAGUIRouter(mock, svc)

	w := postAGUI(t, router, "writer", minimalAGUIBody("t-1", "hi"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"type":"STATE_DELTA"`) ||
		!strings.Contains(body, `{"op":"add","path":"/summary","value":"done"}`) {
		t.Fatalf("stream missing trailing STATE_DELTA:\n%s", body)
	}
	// The delta must precede RUN_FINISHED so the client applies it in-run.
	if strings.Index(body, "STATE_DELTA") > strings.Index(body, "RUN_FINISHED") {
		t.Fatalf("STATE_DELTA emitted after RUN_FINISHED:\n%s", body)
	}
}

// withAGUIField returns a copy of body with one field replaced, so table cases
// do not mutate a shared map.
func withAGUIField(body map[string]any, key string, value any) map[string]any {
	out := make(map[string]any, len(body)+1)
	for k, v := range body {
		out[k] = v
	}
	out[key] = value
	return out
}
