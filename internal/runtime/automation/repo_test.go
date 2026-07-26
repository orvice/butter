package automation

import (
	"context"
	"testing"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMemoryDefinitionRepoWorkspaceIsolation(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryDefinitionRepo()

	for _, a := range []*agentsv1.Automation{
		{Name: "daily", WorkspaceId: "ws-a", Enabled: true},
		{Name: "weekly", WorkspaceId: "ws-a", Enabled: true},
		{Name: "daily", WorkspaceId: "ws-b", Enabled: true},
	} {
		if err := repo.Create(ctx, a); err != nil {
			t.Fatalf("Create(%s/%s): %v", a.GetWorkspaceId(), a.GetName(), err)
		}
	}

	got, err := repo.List(ctx, "ws-a")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List(ws-a) len = %d, want 2", len(got))
	}
	for _, a := range got {
		if a.GetWorkspaceId() != "ws-a" {
			t.Fatalf("List(ws-a) returned workspace %q", a.GetWorkspaceId())
		}
	}

	if _, err := repo.Get(ctx, "ws-a", "missing"); err != ErrAutomationNotFound {
		t.Fatalf("Get missing error = %v, want ErrAutomationNotFound", err)
	}
}

func TestMemoryRunRepoPaginationAndStatusUpdate(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRunRepo()
	base := time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)

	runs := []*agentsv1.AutomationRun{
		{Id: "old", WorkspaceId: "ws-a", AutomationName: "daily", Status: agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING, StartedAt: timestamppb.New(base)},
		{Id: "new", WorkspaceId: "ws-a", AutomationName: "daily", Status: agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING, StartedAt: timestamppb.New(base.Add(time.Hour))},
		{Id: "other-automation", WorkspaceId: "ws-a", AutomationName: "weekly", Status: agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING, StartedAt: timestamppb.New(base.Add(2 * time.Hour))},
		{Id: "other-workspace", WorkspaceId: "ws-b", AutomationName: "daily", Status: agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING, StartedAt: timestamppb.New(base.Add(3 * time.Hour))},
	}
	for _, run := range runs {
		if err := repo.Save(ctx, run); err != nil {
			t.Fatalf("Save(%s): %v", run.GetId(), err)
		}
	}

	page, next, err := repo.List(ctx, "ws-a", "daily", 1, "")
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if len(page) != 1 || page[0].GetId() != "new" {
		t.Fatalf("page 1 = %v, want newest daily run", ids(page))
	}
	if next == "" {
		t.Fatal("next token empty, want another page")
	}

	page, next, err = repo.List(ctx, "ws-a", "daily", 1, next)
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if len(page) != 1 || page[0].GetId() != "old" {
		t.Fatalf("page 2 = %v, want old daily run", ids(page))
	}
	if next != "" {
		t.Fatalf("next token = %q, want empty", next)
	}

	updated := proto.Clone(runs[1]).(*agentsv1.AutomationRun)
	updated.Status = agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED
	if err := repo.Save(ctx, updated); err != nil {
		t.Fatalf("Save update: %v", err)
	}
	got, err := repo.Get(ctx, "ws-a", "new")
	if err != nil {
		t.Fatalf("Get updated: %v", err)
	}
	if got.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED {
		t.Fatalf("status = %s, want succeeded", got.GetStatus())
	}
}

func TestMemoryRunRepoListWaitingBySession(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRunRepo()

	waiting := &agentsv1.AutomationRun{
		Id: "waiting", WorkspaceId: "ws-a", AutomationName: "approval",
		Status:         agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT,
		SessionAppName: "automation:approval", SessionUserId: "automation:ws-a", SessionId: "automation:waiting",
	}
	for _, run := range []*agentsv1.AutomationRun{
		waiting,
		// Same session coords but already succeeded — must be excluded.
		{Id: "done", WorkspaceId: "ws-a", AutomationName: "approval", Status: agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED,
			SessionAppName: "automation:approval", SessionUserId: "automation:ws-a", SessionId: "automation:done"},
		// Waiting run on a different session — must be excluded from the lookup.
		{Id: "other", WorkspaceId: "ws-a", AutomationName: "approval", Status: agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT,
			SessionAppName: "automation:approval", SessionUserId: "automation:ws-a", SessionId: "automation:other"},
	} {
		if err := repo.Save(ctx, run); err != nil {
			t.Fatalf("Save(%s): %v", run.GetId(), err)
		}
	}

	got, err := repo.ListWaitingBySession(ctx, "automation:approval", "automation:ws-a", "automation:waiting")
	if err != nil {
		t.Fatalf("ListWaitingBySession: %v", err)
	}
	if len(got) != 1 || got[0].GetId() != "waiting" {
		t.Fatalf("ListWaitingBySession = %v, want [waiting]", ids(got))
	}

	none, err := repo.ListWaitingBySession(ctx, "automation:approval", "automation:ws-a", "automation:done")
	if err != nil {
		t.Fatalf("ListWaitingBySession(done): %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("ListWaitingBySession(done) = %v, want empty", ids(none))
	}
}

func TestMemoryStepRunRepoListByRunOrdering(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryStepRunRepo()
	base := time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)

	for _, stepRun := range []*agentsv1.AutomationStepRun{
		{Id: "second", WorkspaceId: "ws-a", RunId: "run-1", StepName: "notify", Order: 2, StartedAt: timestamppb.New(base.Add(time.Minute))},
		{Id: "first", WorkspaceId: "ws-a", RunId: "run-1", StepName: "summarize", Order: 1, StartedAt: timestamppb.New(base)},
		{Id: "other-run", WorkspaceId: "ws-a", RunId: "run-2", StepName: "ignore", Order: 1, StartedAt: timestamppb.New(base)},
		{Id: "other-workspace", WorkspaceId: "ws-b", RunId: "run-1", StepName: "ignore", Order: 1, StartedAt: timestamppb.New(base)},
	} {
		if err := repo.Save(ctx, stepRun); err != nil {
			t.Fatalf("Save(%s): %v", stepRun.GetId(), err)
		}
	}

	got, err := repo.ListByRun(ctx, "ws-a", "run-1")
	if err != nil {
		t.Fatalf("ListByRun: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListByRun len = %d, want 2", len(got))
	}
	if got[0].GetId() != "first" || got[1].GetId() != "second" {
		t.Fatalf("ListByRun order = %v, want [first second]", stepIDs(got))
	}
}

func ids(runs []*agentsv1.AutomationRun) []string {
	out := make([]string, len(runs))
	for i, run := range runs {
		out[i] = run.GetId()
	}
	return out
}

func stepIDs(stepRuns []*agentsv1.AutomationStepRun) []string {
	out := make([]string, len(stepRuns))
	for i, stepRun := range stepRuns {
		out[i] = stepRun.GetId()
	}
	return out
}
