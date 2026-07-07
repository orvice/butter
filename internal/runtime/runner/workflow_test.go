package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeOpenAIServer serves an OpenAI-compatible chat completions endpoint.
// Each completion echoes "<model>(<last user message>)" so a test can
// observe which model ran and what input it received — the expected chain
// output is an independent literal, not recomputed from the code under test.
func fakeOpenAIServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		lastUser := ""
		for _, m := range req.Messages {
			if m.Role != "user" {
				continue
			}
			// Content is either a JSON string or an array of typed parts.
			var s string
			if err := json.Unmarshal(m.Content, &s); err == nil {
				lastUser = s
				continue
			}
			var parts []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}
			if err := json.Unmarshal(m.Content, &parts); err == nil {
				var b strings.Builder
				for _, p := range parts {
					b.WriteString(p.Text)
				}
				lastUser = b.String()
			}
		}
		reply := fmt.Sprintf("%s(%s)", req.Model, lastUser)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"id": "cmpl-test",
			"object": "chat.completion",
			"created": 1,
			"model": %q,
			"choices": [{"index": 0, "message": {"role": "assistant", "content": %q}, "finish_reason": "stop"}]
		}`, req.Model, reply)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// TestRun_WorkflowLinearChain drives a message through the runner seam: a
// WORKFLOW agent with a linear two-AGENT-node graph runs to completion and
// the chain's final output comes back as the reply. step_b's input is
// step_a's output, proving the edge actually carried the value.
func TestRun_WorkflowLinearChain(t *testing.T) {
	srv := fakeOpenAIServer(t)

	providers := []agentsv1.ModelProvider{{
		Name:    "fake",
		Type:    "openai",
		BaseUrl: srv.URL,
		Models: []*agentsv1.ModelConfig{
			{Name: "model-a"},
			{Name: "model-b"},
		},
	}}

	agents := []agentsv1.Agent{{
		Name:        "wf",
		Description: "linear workflow",
		Type:        agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		WorkspaceId: "ws-a",
		SubAgents: []*agentsv1.Agent{
			{Name: "step_a", Config: &agentsv1.AgentConfig{Model: "model-a"}},
			{Name: "step_b", Config: &agentsv1.AgentConfig{Model: "model-b"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "step_a", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "step_a"},
					{Name: "step_b", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "step_b"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "step_a"},
					{From: "step_a", To: "step_b"},
				},
			},
		},
	}}

	svc, err := NewService(context.Background(), agents, providers,
		nil, nil, nil, session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        "test-uuid",
		SessionId:   "s1",
		UserId:      "u1",
		ChannelName: "test-app",
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		WorkspaceId: "ws-a",
	}

	out, err := svc.Run(context.Background(), "wf", []*genai.Part{{Text: "hello"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// step_a sees the user's message; step_b sees step_a's output.
	want := "model-b(model-a(hello))"
	if !strings.Contains(out, want) {
		t.Fatalf("reply %q does not contain the chain's final output %q", out, want)
	}
}
