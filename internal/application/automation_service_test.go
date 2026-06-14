package application

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/genai"

	runtimeautomation "go.orx.me/apps/butter/internal/runtime/automation"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type automationTestRunner struct {
	calls int
}

func (r *automationTestRunner) HasAgentInWorkspace(workspaceID, name string) bool {
	return workspaceID == wsTest && name == "agent1"
}

func (r *automationTestRunner) RunSSE(context.Context, string, []*genai.Part, string, *agentsv1.ContextInfo, runner.EventCallback, runner.CompactionCallback) (string, error) {
	r.calls++
	return "agent done", nil
}

func newAutomationTestService() (*AutomationServiceServer, *runtimeautomation.MemoryDefinitionRepo, *runtimeautomation.MemoryRunRepo, *runtimeautomation.MemoryStepRunRepo, *automationTestRunner) {
	defRepo := runtimeautomation.NewMemoryDefinitionRepo()
	runRepo := runtimeautomation.NewMemoryRunRepo()
	stepRepo := runtimeautomation.NewMemoryStepRunRepo()
	runnerSvc := &automationTestRunner{}
	engine := runtimeautomation.NewEngine(defRepo, runRepo, stepRepo, runtimeautomation.EngineOptions{Runner: runnerSvc})
	svc := NewAutomationServiceServer()
	svc.SetRepos(defRepo, runRepo, stepRepo)
	svc.SetEngine(engine)
	svc.SetAgentValidator(runnerSvc)
	return svc, defRepo, runRepo, stepRepo, runnerSvc
}

func validAutomation(name string) *agentsv1.Automation {
	return &agentsv1.Automation{
		Name:    name,
		Enabled: true,
		Trigger: &agentsv1.AutomationTrigger{
			Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL,
		},
		Steps: []*agentsv1.AutomationStep{
			{
				Name:        "invoke",
				Type:        agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT,
				InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "agent1", Input: "run"},
			},
		},
	}
}

func TestAutomationServiceCRUDWorkspaceIsolation(t *testing.T) {
	svc, _, _, _, _ := newAutomationTestService()
	ctx := testCtx()

	createResp, err := svc.CreateAutomation(ctx, connect.NewRequest(&agentsv1.CreateAutomationRequest{Automation: validAutomation("daily")}))
	if err != nil {
		t.Fatalf("CreateAutomation: %v", err)
	}
	if createResp.Msg.GetAutomation().GetWorkspaceId() != wsTest {
		t.Fatalf("workspace_id = %q, want %q", createResp.Msg.GetAutomation().GetWorkspaceId(), wsTest)
	}

	listResp, err := svc.ListAutomations(ctx, connect.NewRequest(&agentsv1.ListAutomationsRequest{}))
	if err != nil {
		t.Fatalf("ListAutomations: %v", err)
	}
	if len(listResp.Msg.GetAutomations()) != 1 {
		t.Fatalf("list len = %d, want 1", len(listResp.Msg.GetAutomations()))
	}

	otherCtx := workspace.WithID(context.Background(), "ws-other")
	otherList, err := svc.ListAutomations(otherCtx, connect.NewRequest(&agentsv1.ListAutomationsRequest{}))
	if err != nil {
		t.Fatalf("ListAutomations other: %v", err)
	}
	if len(otherList.Msg.GetAutomations()) != 0 {
		t.Fatalf("other workspace list len = %d, want 0", len(otherList.Msg.GetAutomations()))
	}

	updated := validAutomation("daily")
	updated.Enabled = false
	updateResp, err := svc.UpdateAutomation(ctx, connect.NewRequest(&agentsv1.UpdateAutomationRequest{Automation: updated}))
	if err != nil {
		t.Fatalf("UpdateAutomation: %v", err)
	}
	if updateResp.Msg.GetAutomation().GetEnabled() {
		t.Fatal("expected automation to be disabled after update")
	}

	deleteResp, err := svc.DeleteAutomation(ctx, connect.NewRequest(&agentsv1.DeleteAutomationRequest{Name: "daily"}))
	if err != nil {
		t.Fatalf("DeleteAutomation: %v", err)
	}
	if deleteResp.Msg.GetAutomation().GetName() != "daily" {
		t.Fatalf("deleted name = %q, want daily", deleteResp.Msg.GetAutomation().GetName())
	}
}

func TestAutomationServiceManualRunAndHistory(t *testing.T) {
	svc, _, _, _, runnerSvc := newAutomationTestService()
	ctx := testCtx()
	if _, err := svc.CreateAutomation(ctx, connect.NewRequest(&agentsv1.CreateAutomationRequest{Automation: validAutomation("daily")})); err != nil {
		t.Fatalf("CreateAutomation: %v", err)
	}

	runResp, err := svc.RunAutomationNow(ctx, connect.NewRequest(&agentsv1.RunAutomationNowRequest{Name: "daily", TriggerPayloadJson: `{"kind":"manual"}`}))
	if err != nil {
		t.Fatalf("RunAutomationNow: %v", err)
	}
	run := runResp.Msg.GetRun()
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED {
		t.Fatalf("run status = %s, want succeeded", run.GetStatus())
	}
	if runnerSvc.calls != 1 {
		t.Fatalf("runner calls = %d, want 1", runnerSvc.calls)
	}

	runsResp, err := svc.ListAutomationRuns(ctx, connect.NewRequest(&agentsv1.ListAutomationRunsRequest{AutomationName: "daily"}))
	if err != nil {
		t.Fatalf("ListAutomationRuns: %v", err)
	}
	if len(runsResp.Msg.GetRuns()) != 1 {
		t.Fatalf("runs len = %d, want 1", len(runsResp.Msg.GetRuns()))
	}
	getRunResp, err := svc.GetAutomationRun(ctx, connect.NewRequest(&agentsv1.GetAutomationRunRequest{Id: run.GetId()}))
	if err != nil {
		t.Fatalf("GetAutomationRun: %v", err)
	}
	if getRunResp.Msg.GetRun().GetId() != run.GetId() {
		t.Fatalf("get run id = %q, want %q", getRunResp.Msg.GetRun().GetId(), run.GetId())
	}
	stepsResp, err := svc.ListAutomationStepRuns(ctx, connect.NewRequest(&agentsv1.ListAutomationStepRunsRequest{RunId: run.GetId()}))
	if err != nil {
		t.Fatalf("ListAutomationStepRuns: %v", err)
	}
	if len(stepsResp.Msg.GetStepRuns()) != 1 {
		t.Fatalf("step runs len = %d, want 1", len(stepsResp.Msg.GetStepRuns()))
	}
}

func TestAutomationServiceValidationAndErrorMapping(t *testing.T) {
	svc, _, _, _, _ := newAutomationTestService()
	ctx := testCtx()
	_, err := svc.CreateAutomation(ctx, connect.NewRequest(&agentsv1.CreateAutomationRequest{Automation: &agentsv1.Automation{Name: "bad"}}))
	if codeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid argument, got %v", err)
	}
	_, err = svc.GetAutomation(ctx, connect.NewRequest(&agentsv1.GetAutomationRequest{Name: "missing"}))
	if codeOf(err) != connect.CodeNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
	_, err = svc.ListAutomations(context.Background(), connect.NewRequest(&agentsv1.ListAutomationsRequest{}))
	if codeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("expected failed precondition for missing workspace, got %v", err)
	}
}

func TestAutomationServiceRejectsInvokeAgentOutsideWorkspace(t *testing.T) {
	svc, _, _, _, _ := newAutomationTestService()
	ctx := testCtx()
	automation := validAutomation("bad-agent")
	automation.Steps[0].InvokeAgent.AgentName = "missing-agent"

	_, err := svc.CreateAutomation(ctx, connect.NewRequest(&agentsv1.CreateAutomationRequest{Automation: automation}))
	if codeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("CreateAutomation code = %v, want invalid argument", err)
	}

	if _, err := svc.CreateAutomation(ctx, connect.NewRequest(&agentsv1.CreateAutomationRequest{Automation: validAutomation("daily")})); err != nil {
		t.Fatalf("CreateAutomation valid: %v", err)
	}
	update := validAutomation("daily")
	update.Steps[0].InvokeAgent.AgentName = "missing-agent"
	_, err = svc.UpdateAutomation(ctx, connect.NewRequest(&agentsv1.UpdateAutomationRequest{Automation: update}))
	if codeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("UpdateAutomation code = %v, want invalid argument", err)
	}
}

func codeOf(err error) connect.Code {
	var cerr *connect.Error
	if errors.As(err, &cerr) {
		return cerr.Code()
	}
	return connect.CodeUnknown
}
