package app

import (
	"context"
	"testing"

	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	internalautomation "go.orx.me/apps/butter/internal/runtime/automation"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// memoryCronJobRepo is a minimal in-memory internalcron.JobRepo for the
// backfill test.
type memoryCronJobRepo struct {
	jobs []*agentsv1.CronJob
}

func (r *memoryCronJobRepo) List(context.Context, string) ([]*agentsv1.CronJob, error) {
	return r.jobs, nil
}
func (r *memoryCronJobRepo) ListAll(context.Context) ([]*agentsv1.CronJob, error) { return r.jobs, nil }
func (r *memoryCronJobRepo) Get(_ context.Context, ws, name string) (*agentsv1.CronJob, error) {
	for _, j := range r.jobs {
		if j.GetWorkspaceId() == ws && j.GetName() == name {
			return j, nil
		}
	}
	return nil, internalcron.ErrAgentNotInWorkspace
}
func (r *memoryCronJobRepo) Create(_ context.Context, job *agentsv1.CronJob) error {
	r.jobs = append(r.jobs, job)
	return nil
}
func (r *memoryCronJobRepo) Update(_ context.Context, job *agentsv1.CronJob) error {
	for i, j := range r.jobs {
		if j.GetWorkspaceId() == job.GetWorkspaceId() && j.GetName() == job.GetName() {
			r.jobs[i] = job
			return nil
		}
	}
	return nil
}
func (r *memoryCronJobRepo) Delete(context.Context, string, string) error { return nil }

func TestBackfillConsumerAgentIDs(t *testing.T) {
	ctx := context.Background()
	const ws = "ws-a"

	agentRepo := configmemory.New()
	if _, err := agentRepo.CreateAgent(ctx, ws, &agentsv1.Agent{Name: "helper", AgentId: "helper-v2", WorkspaceId: ws}); err != nil {
		t.Fatal(err)
	}
	// An agent still without an assigned id — records pointing at it stay
	// name-only and must be left untouched.
	if _, err := agentRepo.CreateAgent(ctx, ws, &agentsv1.Agent{Name: "legacy", WorkspaceId: ws}); err != nil {
		t.Fatal(err)
	}

	chanRepo := configmemory.New()
	if _, err := chanRepo.CreateChannel(ctx, ws, &agentsv1.AgentChannel{Name: "ch-1", AgentName: "helper", WorkspaceId: ws}); err != nil {
		t.Fatal(err)
	}
	if _, err := chanRepo.CreateChannel(ctx, ws, &agentsv1.AgentChannel{Name: "ch-legacy", AgentName: "legacy", WorkspaceId: ws}); err != nil {
		t.Fatal(err)
	}
	// Already migrated — must not be disturbed.
	if _, err := chanRepo.CreateChannel(ctx, ws, &agentsv1.AgentChannel{Name: "ch-2", AgentName: "helper", AgentId: "pinned", WorkspaceId: ws}); err != nil {
		t.Fatal(err)
	}

	jobRepo := &memoryCronJobRepo{}
	_ = jobRepo.Create(ctx, &agentsv1.CronJob{Name: "job-1", AgentName: "helper", WorkspaceId: ws})

	autoRepo := internalautomation.NewMemoryDefinitionRepo()
	if err := autoRepo.Create(ctx, &agentsv1.Automation{
		Name:        "auto-1",
		WorkspaceId: ws,
		Steps: []*agentsv1.AutomationStep{
			{Name: "s1", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "helper"}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	backfillConsumerAgentIDs(ctx, agentRepo, chanRepo, jobRepo, autoRepo)

	ch1, _ := chanRepo.GetChannel(ctx, ws, "ch-1")
	if ch1.GetAgentId() != "helper-v2" {
		t.Fatalf("ch-1 agent_id = %q, want helper-v2", ch1.GetAgentId())
	}
	chLegacy, _ := chanRepo.GetChannel(ctx, ws, "ch-legacy")
	if chLegacy.GetAgentId() != "" {
		t.Fatalf("ch-legacy agent_id = %q, want empty (agent has no id)", chLegacy.GetAgentId())
	}
	ch2, _ := chanRepo.GetChannel(ctx, ws, "ch-2")
	if ch2.GetAgentId() != "pinned" {
		t.Fatalf("ch-2 agent_id = %q, want pinned (untouched)", ch2.GetAgentId())
	}

	job1, _ := jobRepo.Get(ctx, ws, "job-1")
	if job1.GetAgentId() != "helper-v2" {
		t.Fatalf("job-1 agent_id = %q, want helper-v2", job1.GetAgentId())
	}

	auto1, _ := autoRepo.Get(ctx, ws, "auto-1")
	if got := auto1.GetSteps()[0].GetInvokeAgent().GetAgentId(); got != "helper-v2" {
		t.Fatalf("auto-1 step agent_id = %q, want helper-v2", got)
	}
}
