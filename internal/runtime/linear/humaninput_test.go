package linear

// Human Input tests (ADR-0015 §9, #369): a Workflow Agent that pauses on a
// Human Input node asks in the Linear Agent Session, and the reply resumes
// it. These run the real runner against a fake OpenAI-compatible model.

import (
	"net/http"
	"strings"
	"testing"

	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/linearapi/lineartest"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// approvalWorkflow is draft → ask (Human Input with form) → publish.
func approvalWorkflow(form *agentsv1.HumanInputForm) []agentsv1.Agent {
	return []agentsv1.Agent{{
		Name: "approval", AgentId: "support", WorkspaceId: "ws-a",
		Type:          agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		ChildAgentIds: []string{"draft", "publish"},
		Config: &agentsv1.AgentConfig{Workflow: &agentsv1.WorkflowConfig{
			Nodes: []*agentsv1.WorkflowNode{
				{Name: "draft", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "draft"},
				{Name: "ask", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Approve this deploy?", Form: form},
				{Name: "publish", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "publish"},
			},
			Edges: []*agentsv1.WorkflowEdge{
				{From: "START", To: "draft"},
				{From: "draft", To: "ask"},
				{From: "ask", To: "publish"},
			},
		}},
	},
		{Name: "draft", AgentId: "draft", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "drafter"}},
		{Name: "publish", AgentId: "publish", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "publisher"}},
	}
}

func decisionForm() *agentsv1.HumanInputForm {
	return &agentsv1.HumanInputForm{Fields: []*agentsv1.HumanInputFormField{{
		Name: "decision", Label: "Decision", Required: true,
		Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE,
		Options: []*agentsv1.HumanInputFormOption{
			{Value: "approve", Label: "Approve"},
			{Value: "reject", Label: "Reject"},
		},
	}}}
}

// workflowFixture wires the orchestrator to the real runner running the
// approval workflow on echoing fake models.
func workflowFixture(t *testing.T, form *agentsv1.HumanInputForm) (*orchestratorFixture, *openaifake.Backend) {
	t.Helper()
	fx, _ := newRecoveryFixture(t)
	backend := openaifake.New(t)
	for _, model := range []string{"drafter", "publisher"} {
		backend.ScriptRequest(model, func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
			openaifake.WriteReply(w, req, model+"("+backend.LastInput(model)+")")
		})
	}
	providers := []agentsv1.ModelProvider{{Name: "fake", Type: "openai", BaseUrl: backend.URL(),
		Models: []*agentsv1.ModelConfig{{Name: "drafter"}, {Name: "publisher"}}}}
	svc, err := runner.NewService(t.Context(), approvalWorkflow(form), providers,
		nil, nil, nil, adksession.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	fx.orch.runner = svc
	return fx, backend
}

func lastActivity(fx *orchestratorFixture) lineartest.Activity {
	acts := fx.linear.Activities()
	return acts[len(acts)-1]
}

func TestAPausedWorkflowAsksInLinearWithSelectableChoicesAndResumesOnTheAnswer(t *testing.T) {
	fx, backend := workflowFixture(t, decisionForm())
	fx.handle(t, fx.event(ActionCreated))

	ask := lastActivity(fx)
	if ask.Type != linearapi.ActivityElicitation || ask.Body != "Approve this deploy?" || ask.Signal != "select" {
		t.Fatalf("pause activity = %+v; want a select elicitation with the question", ask)
	}
	options, _ := ask.SignalMetadata["options"].([]any)
	if len(options) != 2 {
		t.Fatalf("options = %+v; want the two choices", ask.SignalMetadata)
	}
	first, _ := options[0].(map[string]any)
	if first["label"] != "Approve" || first["value"] != "approve" {
		t.Fatalf("first option = %+v", first)
	}
	record := recordFor(t, fx.orch.processing, "d-created")
	if record.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED || record.GetOutputType() != linearapi.ActivityElicitation {
		t.Fatalf("pause record = %+v; want the elicitation delivered", record)
	}

	// Picking an option sends its value as the next prompt.
	fx.handle(t, fx.prompted("approve"))
	done := lastActivity(fx)
	if done.Type != linearapi.ActivityResponse || !strings.HasPrefix(done.Body, "publisher(") {
		t.Fatalf("after the answer = %+v; want the publisher's reply", done)
	}
	if !strings.Contains(backend.LastInput("publisher"), "approve") {
		t.Fatalf("publisher input = %q; want it to carry the answer", backend.LastInput("publisher"))
	}
}

func TestAFreeTextAnswerToASelectAlsoResumes(t *testing.T) {
	fx, backend := workflowFixture(t, decisionForm())
	fx.handle(t, fx.event(ActionCreated))
	fx.handle(t, fx.prompted("go ahead and ship it"))
	if done := lastActivity(fx); done.Type != linearapi.ActivityResponse {
		t.Fatalf("after a free-text answer = %+v; want the workflow to finish", done)
	}
	if !strings.Contains(backend.LastInput("publisher"), "go ahead and ship it") {
		t.Fatalf("publisher input = %q", backend.LastInput("publisher"))
	}
}

func TestAFormWithSeveralFieldsAsksWithFieldInstructions(t *testing.T) {
	form := decisionForm()
	form.Fields = append(form.Fields, &agentsv1.HumanInputFormField{
		Name: "reason", Label: "Reason", Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
	})
	fx, _ := workflowFixture(t, form)
	fx.handle(t, fx.event(ActionCreated))
	ask := lastActivity(fx)
	if ask.Type != linearapi.ActivityElicitation || ask.Signal != "" {
		t.Fatalf("pause activity = %+v; want a plain elicitation", ask)
	}
	if !strings.HasPrefix(ask.Body, "Approve this deploy?") || !strings.Contains(ask.Body, "Reason") || !strings.Contains(ask.Body, "Decision") {
		t.Fatalf("elicitation body = %q; want the question and the field instructions", ask.Body)
	}
}

func TestAQuestionWithoutAFormIsAPlainElicitation(t *testing.T) {
	fx, _ := workflowFixture(t, nil)
	fx.handle(t, fx.event(ActionCreated))
	ask := lastActivity(fx)
	if ask.Type != linearapi.ActivityElicitation || ask.Body != "Approve this deploy?" || ask.Signal != "" {
		t.Fatalf("pause activity = %+v", ask)
	}
	fx.handle(t, fx.prompted("yes"))
	if lastActivity(fx).Type != linearapi.ActivityResponse {
		t.Fatal("the answer did not resume the workflow")
	}
}
