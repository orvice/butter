package main

// The mobile Workflow editor (#452) saves graphs through AgentService and
// leaves running them to the server. These tests save editor-shaped Workflows
// through the production Connect handlers (auth and Workspace middleware
// included), read the stored records back, and run them on the production
// runner against a fake model backend. The configuration the editor writes
// must therefore produce the events it promises.
//
// The configurations mirror what mobile/lib/features/agents/workflow_draft.dart
// writes: unique node names, AGENT nodes referencing independent Agents by
// agent_id with every referenced Agent listed in child_agent_ids (in node
// order), route labels and default flags on edges, and a Human Input form
// whose fields keep their order.

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const workflowWorkspace = "alpha"

func (f *providerFixture) agentClient(workspaceID string) agentsv1connect.AgentServiceClient {
	interceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+f.token)
			req.Header().Set("X-Workspace-ID", workspaceID)
			return next(ctx, req)
		}
	})
	return agentsv1connect.NewAgentServiceClient(http.DefaultClient, f.url+"/api", connect.WithInterceptors(interceptor))
}

// savedWorkflow is a Workflow Agent and the independent Agents it runs, as
// the editor sends them.
type savedWorkflow struct {
	workflow *agentsv1.Agent
	children []*agentsv1.Agent
}

func llmChild(id, model string) *agentsv1.Agent {
	return &agentsv1.Agent{
		AgentId: id, Name: id, Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		Config: &agentsv1.AgentConfig{Model: model},
	}
}

func agentNode(name, agentID string) *agentsv1.WorkflowNode {
	return &agentsv1.WorkflowNode{
		Name: name, Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: agentID,
	}
}

func workflowAgent(id string, childIDs []string, nodes []*agentsv1.WorkflowNode, edges []*agentsv1.WorkflowEdge) *agentsv1.Agent {
	return &agentsv1.Agent{
		AgentId: id, Name: id, Type: agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		ChildAgentIds: childIDs,
		Config:        &agentsv1.AgentConfig{Workflow: &agentsv1.WorkflowConfig{Nodes: nodes, Edges: edges}},
	}
}

func branchWorkflow(prefix string) savedWorkflow {
	id := prefix + "-review"
	return savedWorkflow{
		workflow: workflowAgent(id,
			[]string{prefix + "-classify", prefix + "-approver", prefix + "-rejecter"},
			[]*agentsv1.WorkflowNode{
				agentNode("classify", prefix+"-classify"),
				{Name: "decide", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_ROUTER},
				agentNode("approver", prefix+"-approver"),
				agentNode("rejecter", prefix+"-rejecter"),
			},
			[]*agentsv1.WorkflowEdge{
				{From: "START", To: "classify"},
				{From: "classify", To: "decide"},
				{From: "decide", To: "approver", Route: "approve"},
				{From: "decide", To: "rejecter", IsDefault: true},
			}),
		children: []*agentsv1.Agent{
			llmChild(prefix+"-classify", "classifier"),
			llmChild(prefix+"-approver", "approver"),
			llmChild(prefix+"-rejecter", "rejecter"),
		},
	}
}

func joinWorkflow(prefix string) savedWorkflow {
	seed := agentNode("seed", prefix+"-seed")
	left := agentNode("left", prefix+"-left")
	left.TimeoutSeconds = 30
	left.Retry = &agentsv1.WorkflowRetryConfig{MaxAttempts: 2, InitialDelaySeconds: 1}
	return savedWorkflow{
		workflow: workflowAgent(prefix+"-fanout",
			[]string{prefix + "-seed", prefix + "-left", prefix + "-right", prefix + "-summarize"},
			[]*agentsv1.WorkflowNode{
				seed, left,
				agentNode("right", prefix+"-right"),
				{Name: "gather", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_JOIN},
				agentNode("summarize", prefix+"-summarize"),
			},
			[]*agentsv1.WorkflowEdge{
				{From: "START", To: "seed"},
				{From: "seed", To: "left"},
				{From: "seed", To: "right"},
				{From: "left", To: "gather"},
				{From: "right", To: "gather"},
				{From: "gather", To: "summarize"},
			}),
		children: []*agentsv1.Agent{
			llmChild(prefix+"-seed", "seeder"),
			llmChild(prefix+"-left", "left"),
			llmChild(prefix+"-right", "right"),
			llmChild(prefix+"-summarize", "summarizer"),
		},
	}
}

func parallelWorkerWorkflow(prefix string) savedWorkflow {
	work := agentNode("work", prefix+"-work")
	work.ParallelWorker = true
	return savedWorkflow{
		workflow: workflowAgent(prefix+"-batch",
			[]string{prefix + "-list", prefix + "-work", prefix + "-collect"},
			[]*agentsv1.WorkflowNode{agentNode("list", prefix+"-list"), work, agentNode("collect", prefix+"-collect")},
			[]*agentsv1.WorkflowEdge{
				{From: "START", To: "list"},
				{From: "list", To: "work"},
				{From: "work", To: "collect"},
			}),
		children: []*agentsv1.Agent{
			llmChild(prefix+"-list", "lister"),
			llmChild(prefix+"-work", "worker"),
			llmChild(prefix+"-collect", "collector"),
		},
	}
}

func humanInputWorkflow(prefix string) savedWorkflow {
	ask := &agentsv1.WorkflowNode{
		Name: "ask", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT,
		Question: "Deploy this draft?",
		Form: &agentsv1.HumanInputForm{
			Title: "Deploy approval",
			Fields: []*agentsv1.HumanInputFormField{
				{
					Name: "env", Label: "Environment", Required: true,
					Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE,
					Options: []*agentsv1.HumanInputFormOption{
						{Value: "prod", Label: "Production"},
						{Value: "staging", Label: "Staging"},
					},
				},
				{
					Name: "reason", Label: "Reason", Hint: "Why now?", MaxLength: 200,
					Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
				},
			},
		},
	}
	return savedWorkflow{
		workflow: workflowAgent(prefix+"-approval",
			[]string{prefix + "-draft", prefix + "-publish"},
			[]*agentsv1.WorkflowNode{
				agentNode("draft", prefix+"-draft"),
				ask,
				agentNode("publish", prefix+"-publish"),
			},
			[]*agentsv1.WorkflowEdge{
				{From: "START", To: "draft"},
				{From: "draft", To: "ask"},
				{From: "ask", To: "publish"},
			}),
		children: []*agentsv1.Agent{
			llmChild(prefix+"-draft", "drafter"),
			llmChild(prefix+"-publish", "publisher"),
		},
	}
}

// save creates the independent Agents, then the Workflow, through the real
// service, and returns the stored records as the runner would load them.
func (f *providerFixture) save(t *testing.T, flows ...savedWorkflow) []agentsv1.Agent {
	t.Helper()
	ctx := context.Background()
	client := f.agentClient(workflowWorkspace)
	var ids []string
	create := func(agent *agentsv1.Agent) {
		t.Helper()
		_, err := client.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
			Agent: agent,
			InitialContent: &agentsv1.AgentContentInput{
				Description: agent.GetDescription(),
				Prompt:      agent.GetConfig().GetInstruction(),
			},
		}))
		if err != nil {
			t.Fatalf("CreateAgent(%s): %v", agent.GetAgentId(), err)
		}
		ids = append(ids, agent.GetAgentId())
	}
	for _, flow := range flows {
		for _, child := range flow.children {
			create(child)
		}
		create(flow.workflow)
	}
	listed, err := client.ListAgents(ctx, connect.NewRequest(&agentsv1.ListAgentsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	stored := make(map[string]*agentsv1.Agent)
	for _, agent := range listed.Msg.GetAgents() {
		stored[agent.GetAgentId()] = agent
	}
	agents := make([]agentsv1.Agent, len(ids))
	for i, id := range ids {
		src, ok := stored[id]
		if !ok {
			t.Fatalf("saved Agent %q is not listed", id)
		}
		proto.Merge(&agents[i], src)
	}
	return agents
}

// runtime builds the production runner over saved Agents whose models are
// served by the fake backend.
func runtimeFor(t *testing.T, backend *openaifake.Backend, agents []agentsv1.Agent, models ...string) *runner.Service {
	t.Helper()
	configs := make([]*agentsv1.ModelConfig, len(models))
	for i, model := range models {
		configs[i] = &agentsv1.ModelConfig{Name: model}
	}
	providers := []agentsv1.ModelProvider{{
		Name: "fake", Type: "openai", BaseUrl: backend.URL(), Models: configs,
	}}
	svc, err := runner.NewService(context.Background(), agents, providers,
		nil, nil, nil, session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	return svc
}

func turnInfo(sessionID string) *agentsv1.ContextInfo {
	return &agentsv1.ContextInfo{
		Uuid: "uuid-" + sessionID, SessionId: sessionID, UserId: "mobile-test",
		ChannelName: "contract", Source: agentsv1.ContextSource_CONTEXT_SOURCE_API,
		WorkspaceId: workflowWorkspace,
	}
}

func text(s string) []*genai.Part { return []*genai.Part{{Text: s}} }

// A saved Router graph steers on the route label and falls to the default
// edge; the branch not taken never runs.
func TestSavedWorkflowBranchesOnRouteAndDefault(t *testing.T) {
	f := newProviderFixture(t)
	agents := f.save(t, branchWorkflow("br"))

	for _, tc := range []struct {
		name, classifierSays, ran, skipped, want string
	}{
		{"label matches case-insensitively", " APPROVE ", "approver", "rejecter", "approver( APPROVE )"},
		{"unmatched input takes the default edge", "needs work", "rejecter", "approver", "rejecter(needs work)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			backend := openaifake.New(t)
			backend.Answer("classifier", tc.classifierSays)
			svc := runtimeFor(t, backend, agents, "classifier", "approver", "rejecter")
			out, err := svc.Run(context.Background(), "br-review", text("please review"), "", turnInfo("branch"), nil, nil)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("reply %q does not contain %q", out, tc.want)
			}
			if got := backend.CallCount(tc.skipped); got != 0 {
				t.Errorf("%s ran %d times, want 0", tc.skipped, got)
			}
			if got := backend.CallCount(tc.ran); got != 1 {
				t.Errorf("%s ran %d times, want 1", tc.ran, got)
			}
		})
	}
}

// A saved fan-out reaches a Join, whose aggregate carries both branches. The
// timeout and retry the editor wrote on a branch node are accepted and the
// run still completes.
func TestSavedWorkflowFansOutAndJoins(t *testing.T) {
	f := newProviderFixture(t)
	agents := f.save(t, joinWorkflow("fj"))
	backend := openaifake.New(t)
	svc := runtimeFor(t, backend, agents, "seeder", "left", "right", "summarizer")

	out, err := svc.Run(context.Background(), "fj-fanout", text("topic"), "", turnInfo("join"), nil, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "summarizer(") {
		t.Fatalf("reply %q has no summarizer output", out)
	}
	joined := backend.LastInput("summarizer")
	for _, branch := range []string{"left(seeder(topic))", "right(seeder(topic))"} {
		if !strings.Contains(joined, branch) {
			t.Errorf("summarizer input %q is missing branch output %q", joined, branch)
		}
	}
	for _, model := range []string{"left", "right"} {
		if got := backend.CallCount(model); got != 1 {
			t.Errorf("%s ran %d times, want 1", model, got)
		}
	}
}

// A saved Parallel Worker node runs once per item of its list input.
func TestSavedWorkflowParallelWorkerRunsPerItem(t *testing.T) {
	f := newProviderFixture(t)
	agents := f.save(t, parallelWorkerWorkflow("pw"))
	backend := openaifake.New(t)
	backend.Answer("lister", `["red", "blue"]`)
	svc := runtimeFor(t, backend, agents, "lister", "worker", "collector")

	if _, err := svc.Run(context.Background(), "pw-batch", text("colors"), "", turnInfo("parallel"), nil, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := backend.CallCount("worker"); got != 2 {
		t.Fatalf("worker ran %d times, want once per item (2)", got)
	}
	collected := backend.LastInput("collector")
	for _, item := range []string{"worker(red)", "worker(blue)"} {
		if !strings.Contains(collected, item) {
			t.Errorf("collector input %q is missing %q", collected, item)
		}
	}
}

// A saved Human Input node pauses the run with its question and a form whose
// fields, order, options and limits are exactly what was configured; a plain
// reply then resumes the graph and reaches the successor.
func TestSavedHumanInputWorkflowPausesWithItsFormAndResumes(t *testing.T) {
	f := newProviderFixture(t)
	agents := f.save(t, humanInputWorkflow("hi"))
	backend := openaifake.New(t)
	svc := runtimeFor(t, backend, agents, "drafter", "publisher")
	info := turnInfo("approval")

	var requested []*session.Event
	onEvent := func(ev *session.Event) {
		if ev.RequestedInput != nil {
			requested = append(requested, ev)
		}
	}
	paused, err := svc.RunTurn(context.Background(), "hi-approval", text("write a post"), "", info, onEvent, nil)
	if err != nil {
		t.Fatalf("pause turn: %v", err)
	}
	if len(paused.Pending) != 1 || paused.Pending[0].Question == "" {
		t.Fatalf("pending = %+v, want one Interrupt with a question", paused.Pending)
	}
	if got := backend.CallCount("publisher"); got != 0 {
		t.Fatalf("publisher ran %d times before the human answered", got)
	}
	if len(requested) != 1 {
		t.Fatalf("request-input events = %d, want 1", len(requested))
	}
	if !strings.HasPrefix(requested[0].RequestedInput.Message, "Deploy this draft?") {
		t.Errorf("question = %q", requested[0].RequestedInput.Message)
	}
	form, ok := a2ui.FormOf(requested[0])
	if !ok {
		t.Fatal("the request-input event carries no form")
	}
	if form.Title != "Deploy approval" || len(form.Fields) != 2 {
		t.Fatalf("form = %+v", form)
	}
	env, reason := form.Fields[0], form.Fields[1]
	if env.Name != "env" || !env.Required || env.Type != a2ui.FieldSingleChoice ||
		len(env.Options) != 2 || env.Options[0].Value != "prod" || env.Options[1].Label != "Staging" {
		t.Errorf("first field = %+v", env)
	}
	if reason.Name != "reason" || reason.Required || reason.Type != a2ui.FieldText ||
		reason.MaxLength != 200 || reason.Hint != "Why now?" {
		t.Errorf("second field = %+v", reason)
	}

	resumed, err := svc.RunTurn(context.Background(), "hi-approval", text("yes, ship it"), "", info, nil, nil)
	if err != nil {
		t.Fatalf("resume turn: %v", err)
	}
	if got := backend.LastInput("publisher"); got != "yes, ship it" {
		t.Errorf("publisher input = %q, want the human's answer", got)
	}
	if !strings.Contains(resumed.Output, "publisher(yes, ship it)") || resumed.Interrupted() {
		t.Errorf("resumed turn = %+v", resumed)
	}
}

// The graph checks the editor performs locally are the server's own: saving
// what the editor flags is refused with InvalidArgument and stores nothing.
func TestServiceRefusesWorkflowsTheEditorFlags(t *testing.T) {
	f := newProviderFixture(t)
	client := f.agentClient(workflowWorkspace)
	ctx := context.Background()
	setup := branchWorkflow("bad")
	for _, child := range setup.children {
		if _, err := client.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: child})); err != nil {
			t.Fatal(err)
		}
	}
	for name, mutate := range map[string]func(*agentsv1.Agent){
		"router without a default edge": func(a *agentsv1.Agent) {
			edges := a.Config.Workflow.Edges
			a.Config.Workflow.Edges = edges[:len(edges)-1]
		},
		"no entry edge from START": func(a *agentsv1.Agent) {
			a.Config.Workflow.Edges = a.Config.Workflow.Edges[1:]
		},
		"route and default on one edge": func(a *agentsv1.Agent) {
			a.Config.Workflow.Edges[2].IsDefault = true
		},
		"route labels equal after folding": func(a *agentsv1.Agent) {
			a.Config.Workflow.Edges[3].IsDefault = false
			a.Config.Workflow.Edges[3].Route = "  APPROVE "
		},
		"duplicate node name": func(a *agentsv1.Agent) {
			a.Config.Workflow.Nodes[3].Name = "approver"
		},
		"agent node without an agent": func(a *agentsv1.Agent) {
			a.Config.Workflow.Nodes[0].AgentId = ""
		},
		"agent node naming a missing agent": func(a *agentsv1.Agent) {
			a.Config.Workflow.Nodes[0].AgentId = "does-not-exist"
		},
	} {
		t.Run(name, func(t *testing.T) {
			bad := branchWorkflow("bad").workflow
			bad.AgentId, bad.Name = "bad-"+strings.ReplaceAll(name, " ", "-"), "bad-"+strings.ReplaceAll(name, " ", "-")
			mutate(bad)
			_, err := client.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: bad}))
			wantCode(t, err, connect.CodeInvalidArgument)
			if _, err := client.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: bad.AgentId})); err == nil {
				t.Error("a refused Workflow was stored")
			}
		})
	}
}
