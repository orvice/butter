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
	"sync/atomic"
	"testing"
	"time"

	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeBackend is an OpenAI-compatible chat completions endpoint. By default
// each completion echoes "<model>(<last user message>)" so a test can observe
// which model ran and what input it received — the expected chain output is
// an independent literal, not recomputed from the code under test. A model
// with a scripted handler answers through it instead. Calls per model are
// recorded.
type fakeBackend struct {
	srv      *httptest.Server
	scripted map[string]http.HandlerFunc

	mu    sync.Mutex
	calls map[string][]string // model -> inputs received
}

func newFakeBackend(t *testing.T) *fakeBackend {
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
	b.script(model, func(w http.ResponseWriter, _ *http.Request) {
		writeCompletion(w, model, reply)
	})
}

// script installs a handler for a model, replacing the echo default.
func (b *fakeBackend) script(model string, handler http.HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scripted[model] = handler
}

// requireConcurrent makes the model's completions block until n requests are
// in flight at once, so the run only succeeds if the caller really issues
// them concurrently. A request left waiting fails with HTTP 400 (not retried
// by the client), surfacing as a run error.
func (b *fakeBackend) requireConcurrent(model string, n int32) {
	var inFlight atomic.Int32
	proceed := make(chan struct{})
	var once sync.Once
	b.script(model, func(w http.ResponseWriter, r *http.Request) {
		if inFlight.Add(1) >= n {
			once.Do(func() { close(proceed) })
		}
		select {
		case <-proceed:
			writeCompletion(w, model, "done")
		case <-time.After(3 * time.Second):
			http.Error(w, `{"error": {"message": "items were not processed concurrently"}}`, http.StatusBadRequest)
		}
	})
}

// failFirstCall makes the model's first completion fail with HTTP 400
// (never retried by the OpenAI client); later calls echo as usual.
func (b *fakeBackend) failFirstCall(model string) {
	failed := false
	b.script(model, func(w http.ResponseWriter, _ *http.Request) {
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
	})
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
	backend := newFakeBackend(t)

	out, err := runWorkflowAgent(t, backend, []agentsv1.Agent{{
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
	backend := newFakeBackend(t)
	backend.answer("classifier", " APPROVE ")

	out, err := runWorkflowAgent(t, backend, approveRejectAgents(),
		[]string{"classifier", "approver", "rejecter"}, "please review")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The approve branch received the router's pass-through input.
	if !strings.Contains(out, "approver( APPROVE )") {
		t.Errorf("reply %q does not contain the approve branch output %q", out, "approver( APPROVE )")
	}
	if backend.callCount("rejecter") != 0 {
		t.Errorf("rejecter model was called %d times, want 0 (default branch must not run)", backend.callCount("rejecter"))
	}
}

// TestRun_WorkflowRouterUnmatchedTakesDefault: input matching no route label
// follows the default edge; the labeled branch never runs.
func TestRun_WorkflowRouterUnmatchedTakesDefault(t *testing.T) {
	backend := newFakeBackend(t)
	backend.answer("classifier", "needs more work")

	out, err := runWorkflowAgent(t, backend, approveRejectAgents(),
		[]string{"classifier", "approver", "rejecter"}, "please review")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !strings.Contains(out, "rejecter(needs more work)") {
		t.Errorf("reply %q does not contain the default branch output %q", out, "rejecter(needs more work)")
	}
	if backend.callCount("approver") != 0 {
		t.Errorf("approver model was called %d times, want 0 (labeled branch must not run)", backend.callCount("approver"))
	}
}

// TestRun_WorkflowFanOutJoin: one node fans out to two branches over
// unconditional edges; a Join node fans them back in and hands the
// aggregated outputs (a map keyed by predecessor) to its successor.
func TestRun_WorkflowFanOutJoin(t *testing.T) {
	backend := newFakeBackend(t)

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

	out, err := runWorkflowAgent(t, backend, agents,
		[]string{"seeder", "left", "right", "summarizer"}, "topic")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Both branches ran on the seed's output, and the summarizer received
	// the Join's aggregate containing both branch outputs.
	if !strings.Contains(out, "summarizer(") {
		t.Fatalf("reply %q does not contain the summarizer's output", out)
	}
	summarizerInput := backend.lastInput("summarizer")
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
	backend := newFakeBackend(t)
	backend.answer("lister", `["red", "blue"]`)

	_, err := runWorkflowAgent(t, backend, parallelWorkerAgents(nil),
		[]string{"lister", "worker", "collector"}, "colors")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := backend.callCount("worker"); got != 2 {
		t.Fatalf("worker model called %d times, want 2 (once per list item)", got)
	}
	collectorInput := backend.lastInput("collector")
	for _, itemOut := range []string{"worker(red)", "worker(blue)"} {
		if !strings.Contains(collectorInput, itemOut) {
			t.Errorf("collector input %q missing aggregated item output %q", collectorInput, itemOut)
		}
	}
}

// TestRun_WorkflowParallelWorkerConcurrent: list items are processed
// concurrently — the backend only answers once both items' requests are in
// flight at the same time, so a serialized worker would fail the run.
func TestRun_WorkflowParallelWorkerConcurrent(t *testing.T) {
	backend := newFakeBackend(t)
	backend.answer("lister", `["a", "b"]`)
	backend.requireConcurrent("worker", 2)

	_, err := runWorkflowAgent(t, backend, parallelWorkerAgents(nil),
		[]string{"lister", "worker", "collector"}, "items")
	if err != nil {
		t.Fatalf("Run: %v (items were not processed concurrently)", err)
	}
}

// TestRun_WorkflowParallelWorkerRetries: a Parallel Worker item whose first
// attempt fails is retried per the node's retry option and the run still
// succeeds. The failure is an HTTP 400, which the OpenAI client does not
// retry itself — recovery can only come from the workflow retry policy.
func TestRun_WorkflowParallelWorkerRetries(t *testing.T) {
	backend := newFakeBackend(t)
	backend.answer("lister", `["red"]`)
	backend.failFirstCall("worker")

	agents := parallelWorkerAgents(func(n *agentsv1.WorkflowNode) {
		n.Retry = &agentsv1.WorkflowRetryConfig{MaxAttempts: 2, InitialDelaySeconds: 1}
	})
	_, err := runWorkflowAgent(t, backend, agents,
		[]string{"lister", "worker", "collector"}, "colors")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := backend.callCount("worker"); got != 2 {
		t.Fatalf("worker model called %d times, want 2 (failed attempt + retry)", got)
	}
	if !strings.Contains(backend.lastInput("collector"), "worker(red)") {
		t.Errorf("collector input %q missing the retried item's output", backend.lastInput("collector"))
	}
}

// TestRun_WorkflowParallelWorkerTimeout: a node's timeout option bounds its
// activation — a backend that never answers fails the run instead of
// hanging it.
func TestRun_WorkflowParallelWorkerTimeout(t *testing.T) {
	backend := newFakeBackend(t)
	backend.answer("lister", `["red"]`)
	backend.script("worker", func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // never answer
	})

	agents := parallelWorkerAgents(func(n *agentsv1.WorkflowNode) {
		n.TimeoutSeconds = 1
	})
	_, err := runWorkflowAgent(t, backend, agents,
		[]string{"lister", "worker", "collector"}, "colors")
	if err == nil {
		t.Fatal("expected the run to fail once the node timeout elapsed, got nil")
	}
	if !strings.Contains(err.Error(), "deadline") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("error %q does not mention the timeout", err.Error())
	}
}

// approvalAgents returns the demo Human Input graph: draft -> ask (Human Input,
// pauses with a question) -> publish. The human's reply becomes publish's
// input (handoff resume).
func approvalAgents() []agentsv1.Agent {
	return []agentsv1.Agent{{
		Name:        "approval",
		Type:        agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		WorkspaceId: "ws-a",
		SubAgents: []*agentsv1.Agent{
			{Name: "draft", Config: &agentsv1.AgentConfig{Model: "drafter"}},
			{Name: "publish", Config: &agentsv1.AgentConfig{Model: "publisher"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "draft", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "draft"},
					{Name: "ask", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Approve this draft?"},
					{Name: "publish", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "publish"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "draft"},
					{From: "draft", To: "ask"},
					{From: "ask", To: "publish"},
				},
			},
		},
	}}
}

// TestRun_WorkflowHumanInputPausesWithQuestion: a workflow reaching a Human
// Input node ends the turn with the node's question as the reply, and the
// turn result carries the pending Interrupt so pause-aware callers (cron)
// can react. The successor must not run yet.
func TestRun_WorkflowHumanInputPausesWithQuestion(t *testing.T) {
	backend := newFakeBackend(t)
	agents := approvalAgents()
	svc := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, session.InMemoryService())

	result, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", turnCtxInfo(&agents[0]), nil, nil)
	if err != nil {
		t.Fatalf("RunTurn: %v", err)
	}

	if !strings.Contains(result.Output, "Approve this draft?") {
		t.Errorf("reply %q does not contain the question %q", result.Output, "Approve this draft?")
	}
	if len(result.Pending) != 1 {
		t.Fatalf("pending interrupts = %d, want 1", len(result.Pending))
	}
	if result.Pending[0].Question != "Approve this draft?" {
		t.Errorf("pending question = %q, want %q", result.Pending[0].Question, "Approve this draft?")
	}
	if result.Pending[0].InterruptID == "" {
		t.Error("pending interrupt has no ID")
	}
	if got := backend.callCount("publisher"); got != 0 {
		t.Errorf("publisher model called %d times before the human answered, want 0", got)
	}
}

// TestRun_WorkflowHumanInputResume: the next message on the session is
// implicitly taken as the answer — the workflow resumes and the reply
// reaches the paused node's successor as its input (handoff).
func TestRun_WorkflowHumanInputResume(t *testing.T) {
	backend := newFakeBackend(t)
	agents := approvalAgents()
	svc := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, session.InMemoryService())
	ctxInfo := turnCtxInfo(&agents[0])

	// Turn 1: pause on the question.
	if _, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", ctxInfo, nil, nil); err != nil {
		t.Fatalf("turn 1: %v", err)
	}

	// Turn 2: a plain-text reply resumes the workflow.
	result, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "yes, ship it"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("turn 2: %v", err)
	}

	if got := backend.lastInput("publisher"); got != "yes, ship it" {
		t.Errorf("publisher input = %q, want the human's answer %q", got, "yes, ship it")
	}
	if !strings.Contains(result.Output, "publisher(yes, ship it)") {
		t.Errorf("reply %q does not contain the successor's output", result.Output)
	}
	if result.Interrupted() {
		t.Errorf("turn 2 still reports pending interrupts: %+v", result.Pending)
	}
}

// TestRun_TurnListenerObservesPauseAndResume: registered turn listeners see
// every turn's outcome together with its context info — the cron scheduler
// relies on this to close a WAITING_INPUT execution when a reply on the
// session completes the workflow (ADR 0003), whatever the entry point.
func TestRun_TurnListenerObservesPauseAndResume(t *testing.T) {
	backend := newFakeBackend(t)
	agents := approvalAgents()
	svc := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, session.InMemoryService())
	ctxInfo := turnCtxInfo(&agents[0])

	type observed struct {
		sessionID   string
		interrupted bool
	}
	var turns []observed
	svc.AddTurnListener(func(ci *agentsv1.ContextInfo, turn *TurnResult, runErr error) {
		if runErr != nil {
			t.Errorf("listener saw unexpected error: %v", runErr)
		}
		turns = append(turns, observed{sessionID: ci.GetSessionId(), interrupted: turn.Interrupted()})
	})

	// Turn 1 pauses on the Human Input node; turn 2 resumes and completes.
	if _, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", ctxInfo, nil, nil); err != nil {
		t.Fatalf("turn 1: %v", err)
	}
	if _, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "yes, ship it"}}, "", ctxInfo, nil, nil); err != nil {
		t.Fatalf("turn 2: %v", err)
	}

	want := []observed{
		{sessionID: "s1", interrupted: true},
		{sessionID: "s1", interrupted: false},
	}
	if len(turns) != 2 || turns[0] != want[0] || turns[1] != want[1] {
		t.Fatalf("listener observed %+v, want %+v", turns, want)
	}
}

// TestRun_WorkflowHumanInputFIFO: with two pending Interrupts from parallel
// branches, each reply answers the oldest one first.
func TestRun_WorkflowHumanInputFIFO(t *testing.T) {
	backend := newFakeBackend(t)
	agents := []agentsv1.Agent{{
		Name:        "double",
		Type:        agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		WorkspaceId: "ws-a",
		SubAgents: []*agentsv1.Agent{
			{Name: "seed", Config: &agentsv1.AgentConfig{Model: "seeder"}},
			{Name: "handle_a", Config: &agentsv1.AgentConfig{Model: "handler-a"}},
			{Name: "handle_b", Config: &agentsv1.AgentConfig{Model: "handler-b"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "seed", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "seed"},
					{Name: "ask_a", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Question A?"},
					{Name: "ask_b", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Question B?"},
					{Name: "handle_a", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "handle_a"},
					{Name: "handle_b", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "handle_b"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "seed"},
					{From: "seed", To: "ask_a"},
					{From: "seed", To: "ask_b"},
					{From: "ask_a", To: "handle_a"},
					{From: "ask_b", To: "handle_b"},
				},
			},
		},
	}}
	svc := buildWorkflowService(t, backend, agents,
		[]string{"seeder", "handler-a", "handler-b"}, session.InMemoryService())
	ctxInfo := turnCtxInfo(&agents[0])

	// Turn 1: both branches pause; branch scheduling decides which asked
	// first, so read the order from the turn result.
	turn1, err := svc.RunTurn(context.Background(), "double",
		[]*genai.Part{{Text: "go"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("turn 1: %v", err)
	}
	if len(turn1.Pending) != 2 {
		t.Fatalf("pending after turn 1 = %d, want 2", len(turn1.Pending))
	}
	handlerFor := map[string]string{"Question A?": "handler-a", "Question B?": "handler-b"}
	firstHandler := handlerFor[turn1.Pending[0].Question]
	secondHandler := handlerFor[turn1.Pending[1].Question]
	if firstHandler == "" || secondHandler == "" || firstHandler == secondHandler {
		t.Fatalf("unexpected pending questions: %+v", turn1.Pending)
	}

	// Turn 2: the reply answers the OLDEST pending Interrupt.
	turn2, err := svc.RunTurn(context.Background(), "double",
		[]*genai.Part{{Text: "first answer"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("turn 2: %v", err)
	}
	if got := backend.lastInput(firstHandler); got != "first answer" {
		t.Errorf("%s input = %q, want %q", firstHandler, got, "first answer")
	}
	if got := backend.callCount(secondHandler); got != 0 {
		t.Errorf("%s ran %d times before its question was answered, want 0", secondHandler, got)
	}
	if !turn2.Interrupted() {
		t.Error("turn 2 should still report the second pending Interrupt")
	}

	// Turn 3: the next reply answers the remaining Interrupt.
	turn3, err := svc.RunTurn(context.Background(), "double",
		[]*genai.Part{{Text: "second answer"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("turn 3: %v", err)
	}
	if got := backend.lastInput(secondHandler); got != "second answer" {
		t.Errorf("%s input = %q, want %q", secondHandler, got, "second answer")
	}
	if turn3.Interrupted() {
		t.Errorf("turn 3 still reports pending interrupts: %+v", turn3.Pending)
	}
}

// TestRun_WorkflowHumanInputRestartResume: a paused workflow survives a
// process restart — a second service built over the same session store (a
// fresh agent instance) resumes correctly, because the run state lives in
// session state, not on the agent.
func TestRun_WorkflowHumanInputRestartResume(t *testing.T) {
	backend := newFakeBackend(t)
	agents := approvalAgents()
	sessions := session.InMemoryService()
	ctxInfo := turnCtxInfo(&agents[0])

	svc1 := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, sessions)
	if _, err := svc1.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", ctxInfo, nil, nil); err != nil {
		t.Fatalf("turn 1: %v", err)
	}

	// "Restart": a brand-new service (fresh ADK agents) over the same store.
	svc2 := buildWorkflowService(t, backend, approvalAgents(), []string{"drafter", "publisher"}, sessions)
	result, err := svc2.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "approved"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("turn 2 after restart: %v", err)
	}

	if got := backend.lastInput("publisher"); got != "approved" {
		t.Errorf("publisher input = %q, want %q", got, "approved")
	}
	if result.Interrupted() {
		t.Errorf("turn 2 still reports pending interrupts: %+v", result.Pending)
	}
}

// TestRun_WorkflowHumanInputClearSessionAbandons: clearing the session
// abandons the paused workflow — the next message starts a fresh run
// instead of resuming.
func TestRun_WorkflowHumanInputClearSessionAbandons(t *testing.T) {
	backend := newFakeBackend(t)
	agents := approvalAgents()
	svc := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, session.InMemoryService())
	ctxInfo := turnCtxInfo(&agents[0])

	if _, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", ctxInfo, nil, nil); err != nil {
		t.Fatalf("turn 1: %v", err)
	}

	if err := svc.ClearSession(context.Background(), ctxInfo.GetChannelName(), ctxInfo.GetSessionId(), ctxInfo.GetUserId()); err != nil {
		t.Fatalf("ClearSession: %v", err)
	}

	// The next message must start fresh: the drafter runs again and the
	// workflow pauses on a new question — nothing reaches the publisher.
	result, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "start over"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("turn 2 after clear: %v", err)
	}

	if got := backend.callCount("drafter"); got != 2 {
		t.Errorf("drafter ran %d times, want 2 (fresh run after clear)", got)
	}
	if got := backend.callCount("publisher"); got != 0 {
		t.Errorf("publisher ran %d times, want 0 (the old answer path was abandoned)", got)
	}
	if len(result.Pending) != 1 {
		t.Errorf("pending after fresh run = %d, want 1 (the new run's question)", len(result.Pending))
	}
}

// TestRun_WorkflowHumanInputEventReachesCallback: the request-for-input
// event is delivered to the onEvent callback (the dashboard chat stream
// forwards those as run events) instead of being swallowed as a final
// response.
func TestRun_WorkflowHumanInputEventReachesCallback(t *testing.T) {
	backend := newFakeBackend(t)
	agents := approvalAgents()
	svc := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, session.InMemoryService())

	var requested []*session.Event
	onEvent := func(evt *session.Event) {
		if evt.RequestedInput != nil {
			requested = append(requested, evt)
		}
	}
	if _, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", turnCtxInfo(&agents[0]), onEvent, nil); err != nil {
		t.Fatalf("RunTurn: %v", err)
	}

	if len(requested) != 1 {
		t.Fatalf("request-input events delivered to onEvent = %d, want 1", len(requested))
	}
	if got := requested[0].RequestedInput.Message; got != "Approve this draft?" {
		t.Errorf("event question = %q, want %q", got, "Approve this draft?")
	}
}

// TestRun_WorkflowResumeDoesNotHijackOtherAgents: a session may be reused
// with a different agent_name (InvokeAgent/StreamAgent allow it). A message
// addressed to an agent with no Workflow Agent in its tree must reach it as
// plain text even when another workflow left a pending Interrupt on the
// session — only workflow-bearing agents resume.
func TestRun_WorkflowResumeDoesNotHijackOtherAgents(t *testing.T) {
	backend := newFakeBackend(t)
	agents := append(approvalAgents(), agentsv1.Agent{
		Name:        "chatbot",
		Type:        agentsv1.AgentType_AGENT_TYPE_LLM,
		WorkspaceId: "ws-a",
		Config:      &agentsv1.AgentConfig{Model: "chat-model"},
	})
	svc := buildWorkflowService(t, backend, agents,
		[]string{"drafter", "publisher", "chat-model"}, session.InMemoryService())
	ctxInfo := turnCtxInfo(&agents[0])

	// Pause the workflow on the session.
	if _, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", ctxInfo, nil, nil); err != nil {
		t.Fatalf("turn 1: %v", err)
	}

	// Same session, different agent: the message must NOT be rewrapped as
	// the Interrupt's answer.
	out, err := svc.Run(context.Background(), "chatbot",
		[]*genai.Part{{Text: "unrelated question"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("chatbot turn: %v", err)
	}
	if got := backend.lastInput("chat-model"); got != "unrelated question" {
		t.Errorf("chatbot model input = %q, want the raw text %q", got, "unrelated question")
	}
	if !strings.Contains(out, "chat-model(unrelated question)") {
		t.Errorf("chatbot reply %q does not contain its own answer", out)
	}
	// The workflow's Interrupt is still pending and answerable.
	result, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "approved"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("resume turn: %v", err)
	}
	if got := backend.lastInput("publisher"); got != "approved" {
		t.Errorf("publisher input = %q, want %q", got, "approved")
	}
	if result.Interrupted() {
		t.Errorf("resume turn still reports pending interrupts: %+v", result.Pending)
	}
}

// TestRunTurnSSE_WorkflowHumanInputPauses: the streaming entry point (the
// path cron and automation actually call) exposes the same structured turn
// result as RunTurn.
func TestRunTurnSSE_WorkflowHumanInputPauses(t *testing.T) {
	backend := newFakeBackend(t)
	agents := approvalAgents()
	svc := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, session.InMemoryService())

	result, err := svc.RunTurnSSE(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", turnCtxInfo(&agents[0]), nil, nil)
	if err != nil {
		t.Fatalf("RunTurnSSE: %v", err)
	}
	if !result.Interrupted() {
		t.Fatal("turn did not report the pending Interrupt")
	}
	if result.Pending[0].Question != "Approve this draft?" {
		t.Errorf("pending question = %q, want %q", result.Pending[0].Question, "Approve this draft?")
	}
	if !strings.Contains(result.Output, "Approve this draft?") {
		t.Errorf("reply %q does not contain the question", result.Output)
	}
}

// TestResumeParts_PreservesNonTextParts: an answer sent with an attachment
// keeps the attachment — only the text is rewrapped as the Interrupt's
// answer; other parts ride along instead of being dropped.
func TestResumeParts_PreservesNonTextParts(t *testing.T) {
	backend := newFakeBackend(t)
	agents := approvalAgents()
	svc := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, session.InMemoryService())
	ctxInfo := turnCtxInfo(&agents[0])

	if _, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", ctxInfo, nil, nil); err != nil {
		t.Fatalf("turn 1: %v", err)
	}

	sess, err := svc.GetSession(context.Background(), ctxInfo.GetChannelName(), ctxInfo.GetSessionId(), ctxInfo.GetUserId())
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	image := &genai.Part{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte{1, 2, 3}}}
	resumed, ok := resumeParts(sess, []*genai.Part{{Text: "approved"}, image})
	if !ok {
		t.Fatal("resumeParts did not rewrap a text answer with a pending Interrupt")
	}
	if resumed[0].FunctionResponse == nil {
		t.Fatal("first part is not the Interrupt's FunctionResponse")
	}
	found := false
	for _, p := range resumed {
		if p.InlineData != nil {
			found = true
		}
	}
	if !found {
		t.Error("the image part was dropped from the resumed message")
	}
}

// runWorkflowAgent builds a runner service holding the given agents (models
// served by the fake backend) and drives one message through the first
// agent via the runner seam, returning the reply.
func runWorkflowAgent(t *testing.T, b *fakeBackend, agents []agentsv1.Agent, models []string, input string) (string, error) {
	t.Helper()
	svc := buildWorkflowService(t, b, agents, models, session.InMemoryService())
	return svc.Run(context.Background(), agents[0].GetName(), []*genai.Part{{Text: input}}, "", turnCtxInfo(&agents[0]), nil, nil)
}

// buildWorkflowService builds a runner service over the given session store;
// passing the same store to two services simulates a process restart.
func buildWorkflowService(t *testing.T, b *fakeBackend, agents []agentsv1.Agent, models []string, sessSvc session.Service) *Service {
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
		nil, nil, nil, sessSvc, nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	return svc
}

func turnCtxInfo(a *agentsv1.Agent) *agentsv1.ContextInfo {
	return &agentsv1.ContextInfo{
		Uuid:        "test-uuid",
		SessionId:   "s1",
		UserId:      "u1",
		ChannelName: "test-app",
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		WorkspaceId: a.GetWorkspaceId(),
	}
}
