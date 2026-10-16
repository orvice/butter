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

func cronClients(t *testing.T) (agentsv1connect.CronJobServiceClient, agentsv1connect.SessionServiceClient) {
	t.Helper()
	f := newProviderFixture(t)
	interceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+f.token)
			req.Header().Set("X-Workspace-ID", cronWorkspace)
			return next(ctx, req)
		}
	})
	with := connect.WithInterceptors(interceptor)
	base := f.url + "/api"
	return agentsv1connect.NewCronJobServiceClient(http.DefaultClient, base, with),
		agentsv1connect.NewSessionServiceClient(http.DefaultClient, base, with)
}

func TestCronRunNowAndWaitingInputReply(t *testing.T) {
	ctx := context.Background()
	cron, sessions := cronClients(t)

	run, err := cron.RunCronJobNow(ctx, connect.NewRequest(&agentsv1.RunCronJobNowRequest{Name: cronInstantJob}))
	if err != nil {
		t.Fatalf("RunCronJobNow: %v", err)
	}
	execID := run.Msg.GetExecution().GetId()
	if execID == "" {
		t.Fatal("expected execution id")
	}

	deadline := time.Now().Add(45 * time.Second)
	for {
		list, err := cron.ListCronExecutions(ctx, connect.NewRequest(&agentsv1.ListCronExecutionsRequest{
			JobName: cronInstantJob, PageSize: 20,
		}))
		if err != nil {
			t.Fatalf("ListCronExecutions: %v", err)
		}
		for _, e := range list.Msg.GetExecutions() {
			if e.GetId() != execID {
				continue
			}
			switch e.GetStatus() {
			case agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS,
				agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR,
				agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SKIPPED,
				agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_CANCELLED:
				goto waiting
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("instant cron run did not finish")
		}
		time.Sleep(200 * time.Millisecond)
	}

waiting:
	reply, err := sessions.ReplySession(ctx, connect.NewRequest(&agentsv1.ReplySessionRequest{
		AgentId:   cronAgentID,
		AppName:   "cron:" + cronSeededJob,
		UserId:    "cron:" + cronSeededJob,
		SessionId: "cron:" + cronSeededJob + ":" + cronWaitingExecID,
		Message:   "approved",
	}))
	if err != nil {
		t.Fatalf("ReplySession: %v", err)
	}
	if reply.Msg.GetResponse() == "" {
		t.Fatal("expected reply text")
	}

	for i := 0; i < 50; i++ {
		list, err := cron.ListCronExecutions(ctx, connect.NewRequest(&agentsv1.ListCronExecutionsRequest{
			JobName: cronSeededJob, PageSize: 20,
		}))
		if err != nil {
			t.Fatalf("list waiting job: %v", err)
		}
		for _, e := range list.Msg.GetExecutions() {
			if e.GetId() == cronWaitingExecID &&
				e.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS {
				return
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("exec-waiting still not SUCCESS")
}
