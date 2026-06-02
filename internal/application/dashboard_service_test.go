package application

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/repo/auth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
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

func isPermissionDenied(err error) bool {
	twerr, ok := err.(*connect.Error)
	return ok && twerr.Code() == connect.CodePermissionDenied
}
