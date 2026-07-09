package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeBackend is an OpenAI-compatible chat completions endpoint. By default
// each completion echoes "<model>(<last user message>)" so a test can observe
// which model ran and what input it received — the expected chain output is
// an independent literal, not recomputed from the code under test. A model
// listed in scripted answers with a fixed reply instead. Calls per model are
// recorded.
type fakeBackend struct {
	srv      *httptest.Server
	scripted map[string]http.HandlerFunc

	mu    sync.Mutex
	calls map[string][]string // model -> inputs received
}

func fakeOpenAIServer(t *testing.T) *fakeBackend {
	t.Helper()
	b := &fakeBackend{
		scripted: map[string]http.HandlerFunc{},
		calls:    map[string][]string{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"messages"`
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &req); err != nil {
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
				var sb strings.Builder
				for _, p := range parts {
					sb.WriteString(p.Text)
				}
				lastUser = sb.String()
			}
		}
		b.mu.Lock()
		b.calls[req.Model] = append(b.calls[req.Model], lastUser)
		handler := b.scripted[req.Model]
		b.mu.Unlock()

		if handler != nil {
			handler(w, r)
			return
		}
		writeCompletion(w, req.Model, fmt.Sprintf("%s(%s)", req.Model, lastUser))
	})
	b.srv = httptest.NewServer(mux)
	t.Cleanup(b.srv.Close)
	return b
}

// answer scripts a fixed reply for a model, overriding the echo default.
func (b *fakeBackend) answer(model, reply string) {
	b.scripted[model] = func(w http.ResponseWriter, _ *http.Request) {
		writeCompletion(w, model, reply)
	}
}

// failFirstCall makes the model's first completion fail with HTTP 400
// (never retried by the OpenAI client); later calls echo as usual.
func (b *fakeBackend) failFirstCall(model string) {
	failed := false
	b.scripted[model] = func(w http.ResponseWriter, _ *http.Request) {
		b.mu.Lock()
		first := !failed
		failed = true
		input := ""
		if inputs := b.calls[model]; len(inputs) > 0 {
			input = inputs[len(inputs)-1]
		}
		b.mu.Unlock()
		if first {
			http.Error(w, `{"error": {"message": "transient failure"}}`, http.StatusBadRequest)
			return
		}
		writeCompletion(w, model, fmt.Sprintf("%s(%s)", model, input))
	}
}

// lastInput returns the last user message the model was called with.
func (b *fakeBackend) lastInput(model string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	inputs := b.calls[model]
	if len(inputs) == 0 {
		return ""
	}
	return inputs[len(inputs)-1]
}

// callCount returns how many completions were requested for the model.
func (b *fakeBackend) callCount(model string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.calls[model])
}

func writeCompletion(w http.ResponseWriter, model, reply string) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"id": "cmpl-test",
		"object": "chat.completion",
		"created": 1,
		"model": %q,
		"choices": [{"index": 0, "message": {"role": "assistant", "content": %q}, "finish_reason": "stop"}]
	}`, model, reply)
}

// TestRun_WorkflowLinearChain drives a message through the runner seam: a
// WORKFLOW agent with a linear two-AGENT-node graph runs to completion and
// the chain's final output comes back as the reply. step_b's input is
// step_a's output, proving the edge actually carried the value.
func TestRun_WorkflowLinearChain(t *testing.T) {
	srv := fakeOpenAIServer(t)

	out, err := runWorkflowAgent(t, srv, []agentsv1.Agent{{
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
	}}, []string{"model-a", "model-b"}, "hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// step_a sees the user's message; step_b sees step_a's output.
	want := "model-b(model-a(hello))"
	if !strings.Contains(out, want) {
		t.Fatalf("reply %q does not contain the chain's final output %q", out, want)
	}
}

// approveRejectAgents returns the demo approve/reject graph: classify's
// answer flows into a Router that sends "approve" down one branch and
// everything else down the default branch.
func approveRejectAgents() []agentsv1.Agent {
	return []agentsv1.Agent{{
		Name:        "review",
		Type:        agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		WorkspaceId: "ws-a",
		SubAgents: []*agentsv1.Agent{
			{Name: "classify", Config: &agentsv1.AgentConfig{Model: "classifier"}},
			{Name: "approver", Config: &agentsv1.AgentConfig{Model: "approver"}},
			{Name: "rejecter", Config: &agentsv1.AgentConfig{Model: "rejecter"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "classify", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "classify"},
					{Name: "decide", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_ROUTER},
					{Name: "approver", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "approver"},
					{Name: "rejecter", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "rejecter"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "classify"},
					{From: "classify", To: "decide"},
					{From: "decide", To: "approver", Route: "approve"},
					{From: "decide", To: "rejecter", IsDefault: true},
				},
			},
		},
	}}
}

// TestRun_WorkflowRouterMatchesLabel: the classifier answers " APPROVE "
// (extra whitespace, different case) and the Router still takes the
// "approve" edge — matching is trimmed and case-insensitive. The rejected
// branch never runs.
func TestRun_WorkflowRouterMatchesLabel(t *testing.T) {
	srv := fakeOpenAIServer(t)
	srv.answer("classifier", " APPROVE ")

	out, err := runWorkflowAgent(t, srv, approveRejectAgents(),
		[]string{"classifier", "approver", "rejecter"}, "please review")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The approve branch received the router's pass-through input.
	if !strings.Contains(out, "approver( APPROVE )") {
		t.Errorf("reply %q does not contain the approve branch output %q", out, "approver( APPROVE )")
	}
	if srv.callCount("rejecter") != 0 {
		t.Errorf("rejecter model was called %d times, want 0 (default branch must not run)", srv.callCount("rejecter"))
	}
}

// TestRun_WorkflowRouterUnmatchedTakesDefault: input matching no route label
// follows the default edge; the labeled branch never runs.
func TestRun_WorkflowRouterUnmatchedTakesDefault(t *testing.T) {
	srv := fakeOpenAIServer(t)
	srv.answer("classifier", "needs more work")

	out, err := runWorkflowAgent(t, srv, approveRejectAgents(),
		[]string{"classifier", "approver", "rejecter"}, "please review")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !strings.Contains(out, "rejecter(needs more work)") {
		t.Errorf("reply %q does not contain the default branch output %q", out, "rejecter(needs more work)")
	}
	if srv.callCount("approver") != 0 {
		t.Errorf("approver model was called %d times, want 0 (labeled branch must not run)", srv.callCount("approver"))
	}
}

// TestRun_WorkflowFanOutJoin: one node fans out to two branches over
// unconditional edges; a Join node fans them back in and hands the
// aggregated outputs (a map keyed by predecessor) to its successor.
func TestRun_WorkflowFanOutJoin(t *testing.T) {
	srv := fakeOpenAIServer(t)

	agents := []agentsv1.Agent{{
		Name:        "fanout",
		Type:        agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		WorkspaceId: "ws-a",
		SubAgents: []*agentsv1.Agent{
			{Name: "seed", Config: &agentsv1.AgentConfig{Model: "seeder"}},
			{Name: "b1", Config: &agentsv1.AgentConfig{Model: "left"}},
			{Name: "b2", Config: &agentsv1.AgentConfig{Model: "right"}},
			{Name: "summarize", Config: &agentsv1.AgentConfig{Model: "summarizer"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "seed", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "seed"},
					{Name: "b1", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "b1"},
					{Name: "b2", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "b2"},
					{Name: "gather", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_JOIN},
					{Name: "summarize", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "summarize"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "seed"},
					{From: "seed", To: "b1"},
					{From: "seed", To: "b2"},
					{From: "b1", To: "gather"},
					{From: "b2", To: "gather"},
					{From: "gather", To: "summarize"},
				},
			},
		},
	}}

	out, err := runWorkflowAgent(t, srv, agents,
		[]string{"seeder", "left", "right", "summarizer"}, "topic")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Both branches ran on the seed's output, and the summarizer received
	// the Join's aggregate containing both branch outputs.
	if !strings.Contains(out, "summarizer(") {
		t.Fatalf("reply %q does not contain the summarizer's output", out)
	}
	summarizerInput := srv.lastInput("summarizer")
	for _, branchOut := range []string{"left(seeder(topic))", "right(seeder(topic))"} {
		if !strings.Contains(summarizerInput, branchOut) {
			t.Errorf("summarizer input %q missing joined branch output %q", summarizerInput, branchOut)
		}
	}
}

// parallelWorkerAgents returns a graph where "work" is marked as a Parallel
// Worker fed by an agent that answers a JSON list. mutateWork adjusts the
// worker node's options per test.
func parallelWorkerAgents(mutateWork func(n *agentsv1.WorkflowNode)) []agentsv1.Agent {
	work := &agentsv1.WorkflowNode{
		Name:           "work",
		Kind:           agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT,
		Agent:          "work",
		ParallelWorker: true,
	}
	if mutateWork != nil {
		mutateWork(work)
	}
	return []agentsv1.Agent{{
		Name:        "batch",
		Type:        agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		WorkspaceId: "ws-a",
		SubAgents: []*agentsv1.Agent{
			{Name: "list", Config: &agentsv1.AgentConfig{Model: "lister"}},
			{Name: "work", Config: &agentsv1.AgentConfig{Model: "worker"}},
			{Name: "collect", Config: &agentsv1.AgentConfig{Model: "collector"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "list", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "list"},
					work,
					{Name: "collect", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "collect"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "list"},
					{From: "list", To: "work"},
					{From: "work", To: "collect"},
				},
			},
		},
	}}
}

// TestRun_WorkflowParallelWorker: a Parallel Worker node runs once per item
// of its list-shaped input and the aggregated outputs flow to the successor.
func TestRun_WorkflowParallelWorker(t *testing.T) {
	srv := fakeOpenAIServer(t)
	srv.answer("lister", `["red", "blue"]`)

	_, err := runWorkflowAgent(t, srv, parallelWorkerAgents(nil),
		[]string{"lister", "worker", "collector"}, "colors")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := srv.callCount("worker"); got != 2 {
		t.Fatalf("worker model called %d times, want 2 (once per list item)", got)
	}
	collectorInput := srv.lastInput("collector")
	for _, itemOut := range []string{"worker(red)", "worker(blue)"} {
		if !strings.Contains(collectorInput, itemOut) {
			t.Errorf("collector input %q missing aggregated item output %q", collectorInput, itemOut)
		}
	}
}

// TestRun_WorkflowParallelWorkerRetries: a Parallel Worker item whose first
// attempt fails is retried per the node's retry option and the run still
// succeeds. The failure is an HTTP 400, which the OpenAI client does not
// retry itself — recovery can only come from the workflow retry policy.
func TestRun_WorkflowParallelWorkerRetries(t *testing.T) {
	srv := fakeOpenAIServer(t)
	srv.answer("lister", `["red"]`)
	srv.failFirstCall("worker")

	agents := parallelWorkerAgents(func(n *agentsv1.WorkflowNode) {
		n.Retry = &agentsv1.WorkflowRetryConfig{MaxAttempts: 2, InitialDelaySeconds: 1}
	})
	_, err := runWorkflowAgent(t, srv, agents,
		[]string{"lister", "worker", "collector"}, "colors")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := srv.callCount("worker"); got != 2 {
		t.Fatalf("worker model called %d times, want 2 (failed attempt + retry)", got)
	}
	if !strings.Contains(srv.lastInput("collector"), "worker(red)") {
		t.Errorf("collector input %q missing the retried item's output", srv.lastInput("collector"))
	}
}

// TestRun_WorkflowParallelWorkerTimeout: a node's timeout option bounds its
// activation — a backend that never answers fails the run instead of
// hanging it.
func TestRun_WorkflowParallelWorkerTimeout(t *testing.T) {
	srv := fakeOpenAIServer(t)
	srv.answer("lister", `["red"]`)
	srv.scripted["worker"] = func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // never answer
	}

	agents := parallelWorkerAgents(func(n *agentsv1.WorkflowNode) {
		n.TimeoutSeconds = 1
	})
	_, err := runWorkflowAgent(t, srv, agents,
		[]string{"lister", "worker", "collector"}, "colors")
	if err == nil {
		t.Fatal("expected the run to fail once the node timeout elapsed, got nil")
	}
	if !strings.Contains(err.Error(), "deadline") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("error %q does not mention the timeout", err.Error())
	}
}

// runWorkflowAgent builds a runner service holding the given agents (models
// served by the fake backend) and drives one message through the first
// agent via the runner seam, returning the reply.
func runWorkflowAgent(t *testing.T, b *fakeBackend, agents []agentsv1.Agent, models []string, input string) (string, error) {
	t.Helper()

	modelCfgs := make([]*agentsv1.ModelConfig, 0, len(models))
	for _, m := range models {
		modelCfgs = append(modelCfgs, &agentsv1.ModelConfig{Name: m})
	}
	providers := []agentsv1.ModelProvider{{
		Name:    "fake",
		Type:    "openai",
		BaseUrl: b.srv.URL,
		Models:  modelCfgs,
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
		WorkspaceId: agents[0].GetWorkspaceId(),
	}

	return svc.Run(context.Background(), agents[0].GetName(), []*genai.Part{{Text: input}}, "", ctxInfo, nil, nil)
}
