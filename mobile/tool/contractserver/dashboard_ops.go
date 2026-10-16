package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/repo/invocation"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const opsWorkspace = "ops468"

const (
	opsRunningInvocation  = "ops-inv-running"
	opsAguiInvocation     = "ops-inv-agui"
	opsSucceededInvocation = "ops-inv-done"
)

type opsAguiStop struct {
	repo invocation.Repository
}

func (s *opsAguiStop) StopInvocation(_ context.Context, _, _, invocationID string) (bool, error) {
	inv, err := s.repo.GetAcrossWorkspaces(context.Background(), invocationID)
	if err != nil {
		return false, nil
	}
	inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED
	return true, s.repo.Save(context.Background(), inv)
}

func registerOpsFixture(
	ctx context.Context,
	r *gin.Engine,
	opts []connect.HandlerOption,
	agentSvc *application.AgentServiceServer,
	agentRepo *configmemory.Store,
	wsRepo workspacerepo.Repository,
	tokens apitoken.Repository,
	daemonReg *daemon.Registry,
	runnerReady func() bool,
	userID string,
) error {
	if _, err := wsRepo.CreateWorkspace(ctx, &agentsv1.Workspace{
		Id: opsWorkspace, Name: opsWorkspace, Slug: opsWorkspace,
	}); err != nil {
		return err
	}
	if _, err := wsRepo.AddMember(ctx, &agentsv1.WorkspaceMember{
		WorkspaceId: opsWorkspace, UserId: userID, Role: "owner",
	}); err != nil {
		return err
	}
	if _, err := agentRepo.CreateAgent(ctx, opsWorkspace, &agentsv1.Agent{
		AgentId: "ops-agent", Name: "ops-agent", WorkspaceId: opsWorkspace,
	}); err != nil {
		return err
	}
	invRepo := invocationmemory.New().WithOwner("ops-fixture")
	now := time.Now().UTC()
	seed := []*agentsv1.Invocation{
		{
			Id: opsRunningInvocation, WorkspaceId: opsWorkspace, AgentId: "ops-agent",
			AgentName: "ops-agent", SessionId: "sess-run", UserId: userID,
			Source: "stream", Status: agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
			StartedAt: timestamppb.New(now.Add(-2 * time.Minute)),
		},
		{
			Id: opsAguiInvocation, WorkspaceId: opsWorkspace, AgentId: "ops-agent",
			AgentName: "ops-agent", SessionId: "agui-thread", UserId: userID,
			Source: invocation.SourceAGUIDetached,
			Status: agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
			StartedAt: timestamppb.New(now.Add(-1 * time.Minute)),
		},
		{
			Id: opsSucceededInvocation, WorkspaceId: opsWorkspace, AgentId: "ops-agent",
			AgentName: "ops-agent", SessionId: "sess-done", UserId: userID,
			Source: "cron", Status: agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
			StartedAt: timestamppb.New(now.Add(-30 * time.Minute)),
		},
	}
	for _, inv := range seed {
		if err := invRepo.Save(ctx, inv); err != nil {
			return err
		}
	}
	agentSvc.SetInvocationRepo(invRepo)
	agentSvc.SetAGUIRunStopper(&opsAguiStop{repo: invRepo})

	dash := application.NewDashboardServiceServer(agentRepo, daemonReg)
	dash.SetInvocationRepo(invRepo)
	dash.SetRunnerReady(runnerReady)
	_, dashHandler := agentsv1connect.NewDashboardServiceHandler(dash, opts...)
	for _, procedure := range []string{
		agentsv1connect.DashboardServiceGetOverviewProcedure,
		agentsv1connect.DashboardServiceGetActivityFeedProcedure,
		agentsv1connect.DashboardServiceGetCronExecutionTimeseriesProcedure,
		agentsv1connect.DashboardServiceGetActivityMetricsProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", dashHandler)))
	}

	tokenSvc := application.NewAPITokenServiceServer(tokens)
	tokenSvc.SetWorkspaceRepo(wsRepo)
	_, tokenHandler := agentsv1connect.NewAPITokenServiceHandler(tokenSvc, opts...)
	r.POST("/api"+agentsv1connect.APITokenServiceCreateAPITokenProcedure,
		gin.WrapH(http.StripPrefix("/api", tokenHandler)))

	return nil
}
