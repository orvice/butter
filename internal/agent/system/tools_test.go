package system

import (
	"context"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/runtime/cron"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestNewListAgentsTool(t *testing.T) {
	store := memory.New()
	store.Seed(context.Background(), []agentsv1.Agent{
		{Name: "agent-a", Description: "First agent", Type: agentsv1.AgentType_AGENT_TYPE_LLM},
		{Name: "agent-b", Description: "Second agent", Type: agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL},
	}, nil, nil, nil)

	tool, err := newListAgentsTool(store)
	if err != nil {
		t.Fatalf("newListAgentsTool: %v", err)
	}
	if tool.Name() != "list_agents" {
		t.Errorf("expected name list_agents, got %s", tool.Name())
	}
}

func TestNewGetAgentTool(t *testing.T) {
	store := memory.New()
	store.Seed(context.Background(), []agentsv1.Agent{
		{Name: "test-agent", Description: "A test agent"},
	}, nil, nil, nil)

	tool, err := newGetAgentTool(store)
	if err != nil {
		t.Fatalf("newGetAgentTool: %v", err)
	}
	if tool.Name() != "get_agent" {
		t.Errorf("expected name get_agent, got %s", tool.Name())
	}
}

func newTestScheduler(t *testing.T, jobRepo *mockJobRepo) *cron.Scheduler {
	t.Helper()
	// Create scheduler with empty job repo and nil runner.
	// This works because no jobs need to be registered at startup.
	s, err := cron.NewScheduler(context.Background(), nil, jobRepo, &mockExecRepo{})
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}
	t.Cleanup(func() { s.Stop() })
	return s
}

func TestNewListCronJobsTool(t *testing.T) {
	scheduler := newTestScheduler(t, &mockJobRepo{})

	tool, err := newListCronJobsTool(scheduler)
	if err != nil {
		t.Fatalf("newListCronJobsTool: %v", err)
	}
	if tool.Name() != "list_cron_jobs" {
		t.Errorf("expected name list_cron_jobs, got %s", tool.Name())
	}
}

func TestNewCreateCronJobTool(t *testing.T) {
	scheduler := newTestScheduler(t, &mockJobRepo{})

	tool, err := newCreateCronJobTool(scheduler)
	if err != nil {
		t.Fatalf("newCreateCronJobTool: %v", err)
	}
	if tool.Name() != "create_cron_job" {
		t.Errorf("expected name create_cron_job, got %s", tool.Name())
	}
}

func TestNewUpdateCronJobTool(t *testing.T) {
	scheduler := newTestScheduler(t, &mockJobRepo{})

	tool, err := newUpdateCronJobTool(scheduler)
	if err != nil {
		t.Fatalf("newUpdateCronJobTool: %v", err)
	}
	if tool.Name() != "update_cron_job" {
		t.Errorf("expected name update_cron_job, got %s", tool.Name())
	}
}

func TestNewDeleteCronJobTool(t *testing.T) {
	scheduler := newTestScheduler(t, &mockJobRepo{})

	tool, err := newDeleteCronJobTool(scheduler)
	if err != nil {
		t.Fatalf("newDeleteCronJobTool: %v", err)
	}
	if tool.Name() != "delete_cron_job" {
		t.Errorf("expected name delete_cron_job, got %s", tool.Name())
	}
}

func TestNewListCronExecutionsTool(t *testing.T) {
	execRepo := &mockExecRepo{
		execs: []*agentsv1.CronExecution{
			{
				Id:        "exec-1",
				JobName:   "job1",
				AgentName: "agent-a",
				Status:    agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS,
				StartedAt: timestamppb.New(time.Now()),
			},
		},
	}

	tool, err := newListCronExecutionsTool(execRepo)
	if err != nil {
		t.Fatalf("newListCronExecutionsTool: %v", err)
	}
	if tool.Name() != "list_cron_executions" {
		t.Errorf("expected name list_cron_executions, got %s", tool.Name())
	}
}

func TestBuildTools(t *testing.T) {
	store := memory.New()
	scheduler := newTestScheduler(t, &mockJobRepo{})
	execRepo := &mockExecRepo{}

	tools, err := buildTools(store, scheduler, execRepo)
	if err != nil {
		t.Fatalf("buildTools: %v", err)
	}
	if len(tools) != 7 {
		t.Errorf("expected 7 tools, got %d", len(tools))
	}

	expectedNames := map[string]bool{
		"list_agents":           true,
		"get_agent":             true,
		"list_cron_jobs":        true,
		"create_cron_job":       true,
		"update_cron_job":       true,
		"delete_cron_job":       true,
		"list_cron_executions":  true,
	}
	for _, tool := range tools {
		if !expectedNames[tool.Name()] {
			t.Errorf("unexpected tool name: %s", tool.Name())
		}
	}
}

// --- Mock types ---

type mockJobRepo struct {
	jobs []*agentsv1.CronJob
}

func (r *mockJobRepo) List(_ context.Context) ([]*agentsv1.CronJob, error) {
	return r.jobs, nil
}

func (r *mockJobRepo) Get(_ context.Context, name string) (*agentsv1.CronJob, error) {
	for _, j := range r.jobs {
		if j.GetName() == name {
			return j, nil
		}
	}
	return nil, context.DeadlineExceeded
}

func (r *mockJobRepo) Create(_ context.Context, job *agentsv1.CronJob) error {
	r.jobs = append(r.jobs, job)
	return nil
}

func (r *mockJobRepo) Update(_ context.Context, job *agentsv1.CronJob) error {
	for i, j := range r.jobs {
		if j.GetName() == job.GetName() {
			r.jobs[i] = job
			return nil
		}
	}
	return context.DeadlineExceeded
}

func (r *mockJobRepo) Delete(_ context.Context, name string) error {
	for i, j := range r.jobs {
		if j.GetName() == name {
			r.jobs = append(r.jobs[:i], r.jobs[i+1:]...)
			return nil
		}
	}
	return context.DeadlineExceeded
}

type mockExecRepo struct {
	execs []*agentsv1.CronExecution
}

func (r *mockExecRepo) Save(_ context.Context, exec *agentsv1.CronExecution) error {
	r.execs = append(r.execs, exec)
	return nil
}

func (r *mockExecRepo) List(_ context.Context, jobName string, pageSize int32, pageToken string) ([]*agentsv1.CronExecution, string, error) {
	if jobName == "" {
		return r.execs, "", nil
	}
	var filtered []*agentsv1.CronExecution
	for _, e := range r.execs {
		if e.GetJobName() == jobName {
			filtered = append(filtered, e)
		}
	}
	return filtered, "", nil
}

func (r *mockExecRepo) GetByID(_ context.Context, id string) (*agentsv1.CronExecution, error) {
	for _, e := range r.execs {
		if e.GetId() == id {
			return e, nil
		}
	}
	return nil, context.DeadlineExceeded
}
