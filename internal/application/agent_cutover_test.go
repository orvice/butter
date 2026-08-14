package application

import (
	"context"
	"testing"

	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	forummemory "go.orx.me/apps/butter/internal/repo/forum/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	internalautomation "go.orx.me/apps/butter/internal/runtime/automation"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeCronJobRepo struct {
	internalcron.JobRepo
	jobs []*agentsv1.CronJob
}

func (f *fakeCronJobRepo) ListAll(context.Context) ([]*agentsv1.CronJob, error) { return f.jobs, nil }

type fakeAutomationRepo struct {
	internalautomation.DefinitionRepo
	automations []*agentsv1.Automation
}

func (f *fakeAutomationRepo) ListAll(context.Context) ([]*agentsv1.Automation, error) {
	return f.automations, nil
}

func seedCutoverAgent(t *testing.T, store *configmemory.Store, ws string, a *agentsv1.Agent) {
	t.Helper()
	if _, err := store.CreateAgent(context.Background(), ws, a); err != nil {
		t.Fatalf("seed agent %s/%s: %v", ws, a.GetAgentId(), err)
	}
}

func checksOf(findings []*agentsv1.AgentIDCutoverFinding) map[string]int {
	out := make(map[string]int)
	for _, f := range findings {
		out[f.GetCheck()]++
	}
	return out
}

func TestAgentIDCutoverVerifier_PassesOnCleanState(t *testing.T) {
	store := configmemory.New()
	seedCutoverAgent(t, store, "ws-a", &agentsv1.Agent{
		AgentId: "parent", Name: "parent",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
		ChildAgentIds:   []string{"child"},
	})
	seedCutoverAgent(t, store, "ws-a", &agentsv1.Agent{
		AgentId: "child", Name: "child",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
	})
	seedCutoverAgent(t, store, "ws-a", &agentsv1.Agent{
		AgentId: "wf", Name: "wf",
		Type:            agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
		Config: &agentsv1.AgentConfig{Workflow: &agentsv1.WorkflowConfig{
			Nodes: []*agentsv1.WorkflowNode{
				{Name: "step", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "child"},
			},
			Edges: []*agentsv1.WorkflowEdge{{From: "START", To: "step"}},
		}},
	})

	findings, err := RunAgentIDCutoverVerifier(context.Background(), AgentCutoverSources{
		Agents:      store,
		Channels:    store,
		CronJobs:    &fakeCronJobRepo{jobs: []*agentsv1.CronJob{{Name: "daily", WorkspaceId: "ws-a", AgentId: "child"}}},
		Automations: &fakeAutomationRepo{},
	})
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %v", findings)
	}
}

func TestAgentIDCutoverVerifier_FlagsAgentViolations(t *testing.T) {
	store := configmemory.New()
	// Runtime-name conflict across workspaces (runner constraint).
	seedCutoverAgent(t, store, "ws-a", &agentsv1.Agent{
		AgentId: "writer-a", Name: "Writer",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
	})
	seedCutoverAgent(t, store, "ws-b", &agentsv1.Agent{
		AgentId: "writer-b", Name: "Writer",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
	})
	// Invalid slug, MIGRATION_REQUIRED, unresolved child, legacy workflow ref.
	seedCutoverAgent(t, store, "ws-a", &agentsv1.Agent{
		AgentId: "Bad_Slug", Name: "bad-slug",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_MIGRATION_REQUIRED,
		ChildAgentIds:   []string{"ghost"},
	})
	seedCutoverAgent(t, store, "ws-a", &agentsv1.Agent{
		AgentId: "wf", Name: "wf",
		Type:            agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
		Config: &agentsv1.AgentConfig{Workflow: &agentsv1.WorkflowConfig{
			Nodes: []*agentsv1.WorkflowNode{
				{Name: "legacy", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "Writer"},
				{Name: "dangling", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "ghost"},
			},
			Edges: []*agentsv1.WorkflowEdge{{From: "START", To: "legacy"}},
		}},
	})

	findings, err := RunAgentIDCutoverVerifier(context.Background(), AgentCutoverSources{Agents: store})
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	checks := checksOf(findings)
	for check, want := range map[string]int{
		"agent_id_invalid":             1,
		"lifecycle_migration_required": 1,
		"child_agent_id_unresolved":    1,
		"workflow_legacy_agent_ref":    1,
		"workflow_agent_id_unresolved": 1,
		"runtime_name_conflict":        2,
	} {
		if checks[check] != want {
			t.Errorf("check %q: got %d findings, want %d (all: %v)", check, checks[check], want, checks)
		}
	}
}

func TestAgentIDCutoverVerifier_SkipsTombstonesForNameConflicts(t *testing.T) {
	store := configmemory.New()
	seedCutoverAgent(t, store, "ws-a", &agentsv1.Agent{
		AgentId: "writer-a", Name: "Writer",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
	})
	// A tombstone with the same runtime name in another workspace never
	// registers with the runner, so it must not count as a conflict.
	seedCutoverAgent(t, store, "ws-b", &agentsv1.Agent{
		AgentId: "writer-b", Name: "Writer",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED,
	})

	findings, err := RunAgentIDCutoverVerifier(context.Background(), AgentCutoverSources{Agents: store})
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %v", findings)
	}
}

func TestAgentIDCutoverVerifier_FlagsConsumerRecords(t *testing.T) {
	store := configmemory.New()
	seedCutoverAgent(t, store, "ws-a", &agentsv1.Agent{
		AgentId: "writer", Name: "Writer",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
	})
	if _, err := store.CreateChannel(context.Background(), "ws-a", &agentsv1.AgentChannel{
		Name: "legacy-channel", AgentName: "Writer",
	}); err != nil {
		t.Fatalf("seed channel: %v", err)
	}

	wsRepo := workspacememory.New()
	if _, err := wsRepo.CreateWorkspace(context.Background(), &agentsv1.Workspace{Id: "ws-a", Name: "Workspace A"}); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	forumRepo := forummemory.New()
	if err := forumRepo.CreateThread(context.Background(), &agentsv1.ForumThread{
		Id: "t1", WorkspaceId: "ws-a", Title: "Legacy thread", AgentNames: []string{"Writer"},
	}); err != nil {
		t.Fatalf("seed thread: %v", err)
	}

	findings, err := RunAgentIDCutoverVerifier(context.Background(), AgentCutoverSources{
		Agents:   store,
		Channels: store,
		CronJobs: &fakeCronJobRepo{jobs: []*agentsv1.CronJob{
			{Name: "daily", WorkspaceId: "ws-a", AgentName: "Writer"},
		}},
		Automations: &fakeAutomationRepo{automations: []*agentsv1.Automation{
			{Name: "auto", WorkspaceId: "ws-a", Steps: []*agentsv1.AutomationStep{
				{Name: "invoke", InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "Writer"}},
			}},
		}},
		Forum:      forumRepo,
		Workspaces: wsRepo,
	})
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	checks := checksOf(findings)
	for _, check := range []string{
		"channel_agent_id_missing",
		"cron_agent_id_missing",
		"automation_agent_id_missing",
		"forum_agent_id_missing",
	} {
		if checks[check] != 1 {
			t.Errorf("check %q: got %d findings, want 1 (all: %v)", check, checks[check], checks)
		}
	}
}
