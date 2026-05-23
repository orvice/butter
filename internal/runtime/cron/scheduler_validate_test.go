package cron

import (
	"errors"
	"testing"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestSchedulerValidateAgentScope(t *testing.T) {
	svc := runner.NewServiceForTest(map[string]string{
		"a-shared":       "ws-a",
		"a-shared-other": "ws-b",
		"b-only":         "ws-b",
	})
	s := &Scheduler{runner: svc}

	cases := []struct {
		name    string
		job     *agentsv1.CronJob
		wantErr bool
	}{
		{
			name: "agent in same workspace",
			job:  &agentsv1.CronJob{Name: "j1", WorkspaceId: "ws-a", AgentName: "a-shared"},
		},
		{
			name:    "agent in different workspace",
			job:     &agentsv1.CronJob{Name: "j2", WorkspaceId: "ws-a", AgentName: "b-only"},
			wantErr: true,
		},
		{
			name:    "agent does not exist",
			job:     &agentsv1.CronJob{Name: "j3", WorkspaceId: "ws-a", AgentName: "ghost"},
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
		})
	}
}

func TestSchedulerValidateAgentScopeWithoutRunner(t *testing.T) {
	// Runner not wired (e.g. tests with stub schedulers) must not panic.
	s := &Scheduler{}
	if err := s.validateAgentScope(&agentsv1.CronJob{Name: "j", WorkspaceId: "ws-a", AgentName: "x"}); err != nil {
		t.Fatalf("expected nil error when runner is nil, got %v", err)
	}
}
