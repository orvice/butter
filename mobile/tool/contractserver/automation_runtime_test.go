package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

func automationClient(t *testing.T) agentsv1connect.AutomationServiceClient {
	t.Helper()
	f := newProviderFixture(t)
	interceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+f.token)
			req.Header().Set("X-Workspace-ID", automationWorkspace)
			return next(ctx, req)
		}
	})
	return agentsv1connect.NewAutomationServiceClient(
		http.DefaultClient, f.url+"/api", connect.WithInterceptors(interceptor),
	)
}

func TestAutomationManualRunAndStepRuns(t *testing.T) {
	ctx := context.Background()
	auto := automationClient(t)

	run, err := auto.RunAutomationNow(ctx, connect.NewRequest(&agentsv1.RunAutomationNowRequest{
		Name: automationSeededManual,
	}))
	if err != nil {
		t.Fatalf("RunAutomationNow: %v", err)
	}
	runID := run.Msg.GetRun().GetId()
	if runID == "" {
		t.Fatal("expected run id")
	}

	deadline := time.Now().Add(30 * time.Second)
	for {
		got, err := auto.GetAutomationRun(ctx, connect.NewRequest(&agentsv1.GetAutomationRunRequest{Id: runID}))
		if err != nil {
			t.Fatalf("GetAutomationRun: %v", err)
		}
		switch got.Msg.GetRun().GetStatus() {
		case agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED,
			agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED,
			agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SKIPPED:
			steps, err := auto.ListAutomationStepRuns(ctx, connect.NewRequest(&agentsv1.ListAutomationStepRunsRequest{RunId: runID}))
			if err != nil {
				t.Fatalf("ListAutomationStepRuns: %v", err)
			}
			if len(steps.Msg.GetStepRuns()) < 2 {
				t.Fatalf("expected notify step runs, got %d", len(steps.Msg.GetStepRuns()))
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("run %s did not settle", runID)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestAutomationConditionPolicySkipsRun(t *testing.T) {
	ctx := context.Background()
	auto := automationClient(t)
	run, err := auto.RunAutomationNow(ctx, connect.NewRequest(&agentsv1.RunAutomationNowRequest{
		Name: automationCondFail, TriggerPayloadJson: `{}`,
	}))
	if err != nil {
		t.Fatalf("RunAutomationNow: %v", err)
	}
	runID := run.Msg.GetRun().GetId()
	deadline := time.Now().Add(15 * time.Second)
	for {
		got, err := auto.GetAutomationRun(ctx, connect.NewRequest(&agentsv1.GetAutomationRunRequest{Id: runID}))
		if err != nil {
			t.Fatalf("GetAutomationRun: %v", err)
		}
		if got.Msg.GetRun().GetStatus() == agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SKIPPED {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("status = %v, want SKIPPED", got.Msg.GetRun().GetStatus())
		}
		time.Sleep(100 * time.Millisecond)
	}
}
