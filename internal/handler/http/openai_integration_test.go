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
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/runtime/runner"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeUpstream mimics an OpenAI-compatible model provider, supporting both
// JSON completions and SSE streaming, so the full handler → runner → ADK →
// provider path can be exercised without a live model.
type fakeUpstream struct {
	srv *httptest.Server
}

func newFakeUpstream(t *testing.T, reply string) *fakeUpstream {
	t.Helper()
	f := &fakeUpstream{}
	mux := http.NewServeMux()
	mux.HandleFunc("/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Model  string `json:"model"`
			Stream bool   `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !req.Stream {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"id": "cmpl-test",
				"object": "chat.completion",
				"created": 1,
				"model": %q,
				"choices": [{"index": 0, "message": {"role": "assistant", "content": %q}, "finish_reason": "stop"}]
			}`, req.Model, reply)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		// Stream the reply one rune at a time, then a finish chunk.
		for _, r := range reply {
			chunk := fmt.Sprintf(`{"id":"cmpl-test","object":"chat.completion.chunk","created":1,"model":%q,"choices":[{"index":0,"delta":{"content":%q}}]}`, req.Model, string(r))
			fmt.Fprintf(w, "data: %s\n\n", chunk)
		}
		final := fmt.Sprintf(`{"id":"cmpl-test","object":"chat.completion.chunk","created":1,"model":%q,"choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`, req.Model)
		fmt.Fprintf(w, "data: %s\n\n", final)
		fmt.Fprint(w, "data: [DONE]\n\n")
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

// buildIntegrationRouter wires the OpenAI handler to a real runner.Service
// backed by the fake upstream.
func buildIntegrationRouter(t *testing.T, upstream *fakeUpstream) *gin.Engine {
	t.Helper()

	agents := []agentsv1.Agent{{
		Name:            "echo-agent",
		Description:     "test agent",
		WorkspaceId:     "ws-a",
		EnableOpenaiApi: true,
		Config:          &agentsv1.AgentConfig{Model: "fake-model"},
	}}
	providers := []agentsv1.ModelProvider{{
		Name:    "fake",
		Type:    "openai",
		BaseUrl: upstream.srv.URL,
		Models:  []*agentsv1.ModelConfig{{Name: "fake-model"}},
	}}

	svc, err := runner.NewService(context.Background(), agents, providers,
		nil, nil, nil, session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	repo := &stubAgentRepo{agents: []*agentsv1.Agent{&agents[0]}}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := wsctx.WithID(c.Request.Context(), "ws-a")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	h := NewOpenAIHandler(repo)
	h.SetRunnerService(svc)
	h.Register(r)
	return r
}

func TestChatCompletions_Integration_NonStreaming(t *testing.T) {
	upstream := newFakeUpstream(t, "hello from model")
	router := buildIntegrationRouter(t, upstream)

	body := `{"model":"echo-agent","messages":[{"role":"user","content":"hi"}],"stream":false}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp chatCompletionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if got := resp.Choices[0].Message.Content; got != "hello from model" {
		t.Errorf("content = %q, want %q", got, "hello from model")
	}
}

func TestChatCompletions_Integration_Streaming(t *testing.T) {
	upstream := newFakeUpstream(t, "streamed reply")
	router := buildIntegrationRouter(t, upstream)

	body := `{"model":"echo-agent","messages":[{"role":"user","content":"hi"}],"stream":true}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/chat/completions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var content strings.Builder
	sawDone := false
	for _, line := range strings.Split(w.Body.String(), "\n") {
		data, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		}
		if data == "[DONE]" {
			sawDone = true
			continue
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			t.Fatalf("bad chunk %q: %v", data, err)
		}
		for _, ch := range chunk.Choices {
			content.WriteString(ch.Delta.Content)
		}
	}
	if !sawDone {
		t.Error("missing [DONE] terminator")
	}
	if got := content.String(); got != "streamed reply" {
		t.Errorf("streamed content = %q, want %q", got, "streamed reply")
	}
}
