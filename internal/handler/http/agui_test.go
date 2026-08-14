package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/adk/v2/workflow"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
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
			name:    "client tools not supported",
			agentID: "writer",
			body: withAGUIField(valid, "tools", []map[string]any{
				{"name": "confirm", "description": "ask", "parameters": map[string]any{}},
			}),
			runner: &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "tools",
		},
		{
			name:    "shared state not supported",
			agentID: "writer",
			body:    withAGUIField(valid, "state", map[string]any{"count": 1}),
			runner:  &mockRunner{}, wantStatus: http.StatusBadRequest, wantError: "state",
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
