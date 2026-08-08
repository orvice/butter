package cron

import (
	"errors"
	"testing"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestSchedulerValidateAgentScope(t *testing.T) {
	svc := runner.NewServiceForTestAgents(
		&agentsv1.Agent{Name: "a-shared", AgentId: "a-shared", WorkspaceId: "ws-a"},
		&agentsv1.Agent{Name: "a-shared-other", AgentId: "a-shared-other", WorkspaceId: "ws-b"},
		&agentsv1.Agent{Name: "b-only", AgentId: "b-only", WorkspaceId: "ws-b"},
	)
	s := &Scheduler{runner: svc}

	cases := []struct {
		name    string
		job     *agentsv1.CronJob
		wantErr bool
	}{
		{
			name: "agent in same workspace",
			job:  &agentsv1.CronJob{Name: "j1", WorkspaceId: "ws-a", AgentId: "a-shared"},
		},
		{
			name:    "agent in different workspace",
			job:     &agentsv1.CronJob{Name: "j2", WorkspaceId: "ws-a", AgentId: "b-only"},
			wantErr: true,
		},
		{
			name:    "agent does not exist",
			job:     &agentsv1.CronJob{Name: "j3", WorkspaceId: "ws-a", AgentId: "ghost"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.validateAgentScope(tc.job)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, ErrAgentNotInWorkspace) {
					t.Fatalf("expected ErrAgentNotInWorkspace, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// On success the job's agent_name is backfilled with the runtime name.
			if tc.job.GetAgentName() != tc.job.GetAgentId() {
				t.Fatalf("agent_name = %q, want resolved runtime name %q", tc.job.GetAgentName(), tc.job.GetAgentId())
			}
		})
	}
}

func TestSchedulerValidateAgentScopeWithoutRunner(t *testing.T) {
	// Runner not wired (e.g. tests with stub schedulers) must not panic.
	s := &Scheduler{}
	if err := s.validateAgentScope(&agentsv1.CronJob{Name: "j", WorkspaceId: "ws-a", AgentId: "x"}); err != nil {
		t.Fatalf("expected nil error when runner is nil, got %v", err)
	}
}
