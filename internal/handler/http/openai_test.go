package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/config"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// stubAgentRepo is a minimal in-memory AgentRepository for tests.
type stubAgentRepo struct {
	configrepo.AgentRepository
	agents []*agentsv1.Agent
}

func (s *stubAgentRepo) ListAgents(_ context.Context, _ string) ([]*agentsv1.Agent, error) {
	return s.agents, nil
}

func (s *stubAgentRepo) GetAgent(_ context.Context, _ string, name string) (*agentsv1.Agent, error) {
	for _, ag := range s.agents {
		if ag.GetName() == name {
			return ag, nil
		}
	}
	return nil, configrepo.ErrNotFound
}

func setupOpenAIRouter(repo configrepo.AgentRepository, authenticated bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	if authenticated {
		// Inject workspace context as if auth middleware ran.
		r.Use(func(c *gin.Context) {
			ctx := wsctx.WithID(c.Request.Context(), "test-workspace")
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		})
	} else {
		// Use real auth middleware that requires a token.
		cfg := &config.AppConfig{APIToken: "secret-token"}
		r.Use(APITokenAuthMiddleware(cfg, nil))
	}

	h := NewOpenAIHandler(repo)
	h.Register(r)
	return r
}

func TestOpenAIModels_FiltersEnabledAgents(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "agent-enabled", EnableOpenaiApi: true},
			{Name: "agent-disabled", EnableOpenaiApi: false},
			{Name: "agent-unset"},
		},
	}
	router := setupOpenAIRouter(repo, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/models", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Object string `json:"object"`
		Data   []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("expected object 'list', got %q", resp.Object)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 model, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "agent-enabled" {
		t.Errorf("expected model id 'agent-enabled', got %q", resp.Data[0].ID)
	}
	if resp.Data[0].Object != "model" {
		t.Errorf("expected object 'model', got %q", resp.Data[0].Object)
	}
	if resp.Data[0].OwnedBy != "butter" {
		t.Errorf("expected owned_by 'butter', got %q", resp.Data[0].OwnedBy)
	}
}

func TestOpenAIModels_EmptyWhenNoneEnabled(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "agent-a", EnableOpenaiApi: false},
			{Name: "agent-b"},
		},
	}
	router := setupOpenAIRouter(repo, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/models", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Object string        `json:"object"`
		Data   []interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("expected object 'list', got %q", resp.Object)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected empty data array, got %d items", len(resp.Data))
	}
}

func TestOpenAIModels_Unauthenticated(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "agent-enabled", EnableOpenaiApi: true},
		},
	}
	router := setupOpenAIRouter(repo, false)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/models", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Chat Completions tests ---

// mockRunner implements OpenAIRunnerService for testing.
type mockRunner struct {
	runResult string
	runErr    error
	// Captures the last call for assertion.
	lastAgentName string
	lastParts     []*genai.Part
	lastCtxInfo   *agentsv1.ContextInfo
	// onEvent simulates streaming events when set.
	onEventFn func(onEvent runner.EventCallback)
}

func (m *mockRunner) Run(_ context.Context, agentName string, parts []*genai.Part, _ string, ctxInfo *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	m.lastAgentName = agentName
	m.lastParts = parts
	m.lastCtxInfo = ctxInfo
	return m.runResult, m.runErr
}

func (m *mockRunner) RunSSE(_ context.Context, agentName string, parts []*genai.Part, _ string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	m.lastAgentName = agentName
	m.lastParts = parts
	m.lastCtxInfo = ctxInfo
	if m.onEventFn != nil {
		m.onEventFn(onEvent)
	}
	return m.runResult, m.runErr
}

func setupChatRouter(repo configrepo.AgentRepository, runnerSvc OpenAIRunnerService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := wsctx.WithID(c.Request.Context(), "test-workspace")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	h := NewOpenAIHandler(repo)
	if runnerSvc != nil {
		h.SetRunnerService(runnerSvc)
	}
	h.Register(r)
	return r
}

func TestChatCompletions_NonStreaming_HappyPath(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{runResult: "Hello, world!"}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model": "my-agent",
		"messages": []map[string]string{
			{"role": "user", "content": "Hi"},
		},
		"stream": false,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Object != "chat.completion" {
		t.Errorf("expected object 'chat.completion', got %q", resp.Object)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", resp.Choices[0].Message.Role)
	}
	if resp.Choices[0].Message.Content != "Hello, world!" {
		t.Errorf("expected content 'Hello, world!', got %q", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Errorf("expected finish_reason 'stop', got %q", resp.Choices[0].FinishReason)
	}
	if resp.ID == "" {
		t.Error("expected non-empty id")
	}
	// Usage must be present but zero-valued.
	if resp.Usage.PromptTokens != 0 || resp.Usage.CompletionTokens != 0 || resp.Usage.TotalTokens != 0 {
		t.Errorf("expected zero usage, got %+v", resp.Usage)
	}
	// Verify the runner received the correct agent name.
	if mr.lastAgentName != "my-agent" {
		t.Errorf("expected runner called with agent 'my-agent', got %q", mr.lastAgentName)
	}
	// Verify ContextInfo includes workspace ID and source.
	if mr.lastCtxInfo == nil {
		t.Fatal("expected non-nil ContextInfo")
	}
	if mr.lastCtxInfo.GetWorkspaceId() != "test-workspace" {
		t.Errorf("expected workspace_id 'test-workspace', got %q", mr.lastCtxInfo.GetWorkspaceId())
	}
	if mr.lastCtxInfo.GetSource() != agentsv1.ContextSource_CONTEXT_SOURCE_API {
		t.Errorf("expected source CONTEXT_SOURCE_API, got %v", mr.lastCtxInfo.GetSource())
	}
	if mr.lastCtxInfo.GetUuid() == "" {
		t.Error("expected non-empty Uuid in ContextInfo")
	}
}

func TestChatCompletions_ModelNotFound(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "other-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{runResult: "should not be called"}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "nonexistent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChatCompletions_ModelDisabled(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "disabled-agent", EnableOpenaiApi: false},
		},
	}
	mr := &mockRunner{runResult: "should not be called"}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "disabled-agent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for disabled agent, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChatCompletions_EmptyMessages(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "my-agent",
		"messages": []map[string]string{},
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChatCompletions_RunnerUnavailable(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	// No runner set (nil).
	router := setupChatRouter(repo, nil)

	body := map[string]interface{}{
		"model":    "my-agent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChatCompletions_RunnerError(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{runErr: fmt.Errorf("model overloaded")}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "my-agent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Streaming tests ---

func TestChatCompletions_Streaming_SSEFormat(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{
		runResult: "Hello World",
		onEventFn: func(onEvent runner.EventCallback) {
			// Simulate two partial events.
			onEvent(makePartialEvent("Hello "))
			onEvent(makePartialEvent("World"))
		},
	}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "my-agent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
		"stream":   true,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Check Content-Type is text/event-stream.
	ct := w.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("expected Content-Type 'text/event-stream', got %q", ct)
	}

	// Parse the SSE stream.
	lines := strings.Split(w.Body.String(), "\n")
	var dataLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
		}
	}

	// Should have at least 2 chunk lines + 1 [DONE].
	if len(dataLines) < 3 {
		t.Fatalf("expected at least 3 data lines, got %d: %v", len(dataLines), dataLines)
	}

	// Last data line should be [DONE].
	if dataLines[len(dataLines)-1] != "[DONE]" {
		t.Errorf("expected last line '[DONE]', got %q", dataLines[len(dataLines)-1])
	}

	// First chunk should have role "assistant" in delta.
	var firstChunk struct {
		Choices []struct {
			Delta struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(dataLines[0]), &firstChunk); err != nil {
		t.Fatalf("failed to parse first chunk: %v", err)
	}
	if len(firstChunk.Choices) != 1 {
		t.Fatalf("expected 1 choice in first chunk, got %d", len(firstChunk.Choices))
	}
	if firstChunk.Choices[0].Delta.Role != "assistant" {
		t.Errorf("expected role 'assistant' in first chunk, got %q", firstChunk.Choices[0].Delta.Role)
	}
	if firstChunk.Choices[0].Delta.Content != "Hello " {
		t.Errorf("expected content 'Hello ' in first chunk, got %q", firstChunk.Choices[0].Delta.Content)
	}

	// Second chunk should NOT have role, only content.
	var secondChunk struct {
		Choices []struct {
			Delta struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(dataLines[1]), &secondChunk); err != nil {
		t.Fatalf("failed to parse second chunk: %v", err)
	}
	if secondChunk.Choices[0].Delta.Role != "" {
		t.Errorf("expected no role in second chunk, got %q", secondChunk.Choices[0].Delta.Role)
	}
	if secondChunk.Choices[0].Delta.Content != "World" {
		t.Errorf("expected content 'World' in second chunk, got %q", secondChunk.Choices[0].Delta.Content)
	}
}

func TestChatCompletions_Streaming_ErrorBeforeStream(t *testing.T) {
	// Error before streaming (model not found) should return JSON, not SSE.
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "other-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "nonexistent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
		"stream":   true,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}

	// Response should be JSON error, not SSE.
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected JSON content type for pre-stream error, got %q", ct)
	}
}

func TestChatCompletions_Streaming_EmptyMessages(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "my-agent",
		"messages": []map[string]string{},
		"stream":   true,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChatCompletions_Streaming_RunnerUnavailable(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	router := setupChatRouter(repo, nil)

	body := map[string]interface{}{
		"model":    "my-agent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
		"stream":   true,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChatCompletions_Streaming_RunnerError(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{runErr: fmt.Errorf("model overloaded")}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "my-agent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
		"stream":   true,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (headers already sent), got %d", w.Code)
	}

	// Should still contain [DONE] and an error chunk with finish_reason "error".
	body_str := w.Body.String()
	if !strings.Contains(body_str, "data: [DONE]") {
		t.Error("expected [DONE] terminator even on error")
	}
	if !strings.Contains(body_str, `"finish_reason":"error"`) {
		t.Error("expected error chunk with finish_reason 'error'")
	}
}

func TestChatCompletions_Streaming_IgnoresNonPartialEvents(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{
		runResult: "final",
		onEventFn: func(onEvent runner.EventCallback) {
			// Send a non-partial event (e.g. tool call) — should be ignored.
			onEvent(&session.Event{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{Text: "ignored"}},
						Role:  genai.RoleModel,
					},
					Partial: false,
				},
			})
			// Send a partial event — should be included.
			onEvent(makePartialEvent("included"))
		},
	}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "my-agent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
		"stream":   true,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	lines := strings.Split(w.Body.String(), "\n")
	var dataLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") && !strings.Contains(line, "[DONE]") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
		}
	}

	// Only the partial event should produce a chunk.
	if len(dataLines) != 1 {
		t.Fatalf("expected 1 data chunk (non-partial filtered out), got %d", len(dataLines))
	}
	if !strings.Contains(dataLines[0], "included") {
		t.Errorf("expected 'included' in chunk, got %q", dataLines[0])
	}
}

func TestChatCompletions_Streaming_ChunkObjectFormat(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{
		runResult: "test",
		onEventFn: func(onEvent runner.EventCallback) {
			onEvent(makePartialEvent("test"))
		},
	}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model":    "my-agent",
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
		"stream":   true,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Parse first chunk and check object field.
	lines := strings.Split(w.Body.String(), "\n")
	var firstData string
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") && !strings.Contains(line, "[DONE]") {
			firstData = strings.TrimPrefix(line, "data: ")
			break
		}
	}
	if firstData == "" {
		t.Fatal("no data chunk found")
	}

	var chunk struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Model   string `json:"model"`
		Choices []struct {
			Index int `json:"index"`
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(firstData), &chunk); err != nil {
		t.Fatalf("failed to parse chunk: %v", err)
	}
	if chunk.Object != "chat.completion.chunk" {
		t.Errorf("expected object 'chat.completion.chunk', got %q", chunk.Object)
	}
	if chunk.ID == "" {
		t.Error("expected non-empty chunk id")
	}
	if chunk.Model != "my-agent" {
		t.Errorf("expected model 'my-agent', got %q", chunk.Model)
	}
}

// makePartialEvent creates a session.Event with partial text content for testing.
func makePartialEvent(text string) *session.Event {
	return &session.Event{
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{{Text: text}},
				Role:  genai.RoleModel,
			},
			Partial: true,
		},
	}
}

func TestChatCompletions_MessagesFormattedCorrectly(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "my-agent", EnableOpenaiApi: true},
		},
	}
	mr := &mockRunner{runResult: "ok"}
	router := setupChatRouter(repo, mr)

	body := map[string]interface{}{
		"model": "my-agent",
		"messages": []map[string]string{
			{"role": "system", "content": "You are helpful"},
			{"role": "user", "content": "Hello"},
			{"role": "assistant", "content": "Hi there"},
			{"role": "user", "content": "How are you?"},
		},
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the text sent to the runner includes formatted messages.
	if len(mr.lastParts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(mr.lastParts))
	}
	text := mr.lastParts[0].Text
	expected := "[System] You are helpful\n[User] Hello\n[Assistant] Hi there\n[User] How are you?"
	if text != expected {
		t.Errorf("expected formatted text:\n%s\ngot:\n%s", expected, text)
	}
}
