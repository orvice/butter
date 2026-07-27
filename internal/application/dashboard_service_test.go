package application

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/repo/auth"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDashboard_RequiresAdmin(t *testing.T) {
	svc := NewDashboardServiceServer(nil, nil)

	t.Run("non-admin user is rejected", func(t *testing.T) {
		ctx := auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "u-1", Role: "member"}, &auth.Session{})

		if _, err := svc.GetOverview(ctx, connect.NewRequest(&agentsv1.GetOverviewRequest{})); !isPermissionDenied(err) {
			t.Errorf("GetOverview: expected PermissionDenied, got %v", err)
		}
		if _, err := svc.GetActivityFeed(ctx, connect.NewRequest(&agentsv1.GetActivityFeedRequest{})); !isPermissionDenied(err) {
			t.Errorf("GetActivityFeed: expected PermissionDenied, got %v", err)
		}
		if _, err := svc.GetCronExecutionTimeseries(ctx, connect.NewRequest(&agentsv1.GetCronExecutionTimeseriesRequest{})); !isPermissionDenied(err) {
			t.Errorf("GetCronExecutionTimeseries: expected PermissionDenied, got %v", err)
		}
		if _, err := svc.GetActivityMetrics(ctx, connect.NewRequest(&agentsv1.GetActivityMetricsRequest{})); !isPermissionDenied(err) {
			t.Errorf("GetActivityMetrics: expected PermissionDenied, got %v", err)
		}
	})

	t.Run("admin user is allowed", func(t *testing.T) {
		ctx := auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "u-1", Role: "admin"}, &auth.Session{})

		// GetActivityFeed short-circuits when invRepo is nil, so admin is enough to
		// confirm the guard does not reject them.
		if _, err := svc.GetActivityFeed(ctx, connect.NewRequest(&agentsv1.GetActivityFeedRequest{})); err != nil {
			t.Errorf("GetActivityFeed admin: unexpected error %v", err)
		}
	})

	t.Run("WithAdmin context is allowed", func(t *testing.T) {
		ctx := auth.WithAdmin(context.Background())

		if _, err := svc.GetActivityFeed(ctx, connect.NewRequest(&agentsv1.GetActivityFeedRequest{})); err != nil {
			t.Errorf("GetActivityFeed admin-tagged: unexpected error %v", err)
		}
	})
}

func TestDashboard_ActivityMetrics(t *testing.T) {
	inv := invocationmemory.New()
	ctx := auth.WithAdmin(context.Background())
	now := time.Now().UTC()

	save := func(id string, started time.Time, status agentsv1.InvocationStatus) {
		if err := inv.Save(ctx, &agentsv1.Invocation{
			Id:        id,
			AgentName: "agent-a",
			StartedAt: timestamppb.New(started),
			Status:    status,
		}); err != nil {
			t.Fatalf("save invocation %s: %v", id, err)
		}
	}
	// Within a 7d window: one succeeded, one failed.
	save("in-1", now.Add(-2*time.Hour), agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED)
	save("in-2", now.Add(-48*time.Hour), agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED)
	// Outside the 7d window: must not be counted.
	save("out-1", now.Add(-10*24*time.Hour), agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED)

	svc := NewDashboardServiceServer(nil, nil)
	svc.SetInvocationRepo(inv)

	resp, err := svc.GetActivityMetrics(ctx, connect.NewRequest(&agentsv1.GetActivityMetricsRequest{
		Range: agentsv1.GetActivityMetricsRequest_RANGE_7D,
	}))
	if err != nil {
		t.Fatalf("GetActivityMetrics: %v", err)
	}
	msg := resp.Msg
	if msg.GetAgentRuns() != 2 {
		t.Errorf("agent_runs = %d, want 2", msg.GetAgentRuns())
	}
	if msg.GetAgentRunsFailed() != 1 {
		t.Errorf("agent_runs_failed = %d, want 1", msg.GetAgentRunsFailed())
	}
	if msg.GetAutomationRuns() != 0 {
		t.Errorf("automation_runs = %d, want 0 (no cron repo wired)", msg.GetAutomationRuns())
	}
	if msg.GetWindowStart() == nil || msg.GetWindowEnd() == nil {
		t.Errorf("window bounds should be populated: start=%v end=%v", msg.GetWindowStart(), msg.GetWindowEnd())
	}
}

func isPermissionDenied(err error) bool {
	twerr, ok := err.(*connect.Error)
	return ok && twerr.Code() == connect.CodePermissionDenied
}
