package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/application"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const (
	cronWorkspace        = "cron"
	cronAgentID          = "cron-agent"
	cronSeededJob        = "seeded-daily"
	cronWaitingExecID    = "exec-waiting"
	cronNotifyGroup      = "ops"
	cronTelegramDestID   = "cron-destination"
	cronTelegramChannel  = "cron-channel"
)

type memoryCronJobs struct {
	mu   sync.Mutex
	jobs map[string]*agentsv1.CronJob
}

func (r *memoryCronJobs) key(ws, name string) string { return ws + "/" + name }

func (r *memoryCronJobs) List(_ context.Context, workspaceID string) ([]*agentsv1.CronJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*agentsv1.CronJob
	for _, j := range r.jobs {
		if j.GetWorkspaceId() == workspaceID {
			out = append(out, proto.Clone(j).(*agentsv1.CronJob))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetName() < out[j].GetName() })
	return out, nil
}

func (r *memoryCronJobs) ListAll(_ context.Context) ([]*agentsv1.CronJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*agentsv1.CronJob, 0, len(r.jobs))
	for _, j := range r.jobs {
		out = append(out, proto.Clone(j).(*agentsv1.CronJob))
	}
	return out, nil
}

func (r *memoryCronJobs) Get(_ context.Context, workspaceID, name string) (*agentsv1.CronJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[r.key(workspaceID, name)]
	if !ok {
		return nil, fmt.Errorf("cron job not found")
	}
	return proto.Clone(j).(*agentsv1.CronJob), nil
}

func (r *memoryCronJobs) Create(_ context.Context, job *agentsv1.CronJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := r.key(job.GetWorkspaceId(), job.GetName())
	if _, ok := r.jobs[k]; ok {
		return fmt.Errorf("already exists")
	}
	r.jobs[k] = proto.Clone(job).(*agentsv1.CronJob)
	return nil
}

func (r *memoryCronJobs) Update(_ context.Context, job *agentsv1.CronJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := r.key(job.GetWorkspaceId(), job.GetName())
	if _, ok := r.jobs[k]; !ok {
		return fmt.Errorf("cron job not found")
	}
	r.jobs[k] = proto.Clone(job).(*agentsv1.CronJob)
	return nil
}

func (r *memoryCronJobs) Delete(_ context.Context, workspaceID, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := r.key(workspaceID, name)
	if _, ok := r.jobs[k]; !ok {
		return fmt.Errorf("cron job not found")
	}
	delete(r.jobs, k)
	return nil
}

type memoryCronExecs struct {
	mu    sync.Mutex
	execs map[string]*agentsv1.CronExecution
	order []string
}

func (r *memoryCronExecs) Save(_ context.Context, exec *agentsv1.CronExecution) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.execs[exec.GetId()]; !ok {
		r.order = append([]string{exec.GetId()}, r.order...)
	}
	r.execs[exec.GetId()] = proto.Clone(exec).(*agentsv1.CronExecution)
	return nil
}

func (r *memoryCronExecs) List(_ context.Context, workspaceID, jobName string, pageSize int32, pageToken string) ([]*agentsv1.CronExecution, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if pageSize <= 0 {
		pageSize = 20
	}
	start, _ := strconv.Atoi(pageToken)
	var filtered []*agentsv1.CronExecution
	for _, id := range r.order {
		e := r.execs[id]
		if e.GetWorkspaceId() != workspaceID {
			continue
		}
		if jobName != "" && e.GetJobName() != jobName {
			continue
		}
		filtered = append(filtered, proto.Clone(e).(*agentsv1.CronExecution))
	}
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + int(pageSize)
	if end > len(filtered) {
		end = len(filtered)
	}
	next := ""
	if end < len(filtered) {
		next = strconv.Itoa(end)
	}
	return filtered[start:end], next, nil
}

func (r *memoryCronExecs) GetByID(_ context.Context, id string) (*agentsv1.CronExecution, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.execs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return proto.Clone(e).(*agentsv1.CronExecution), nil
}

func (r *memoryCronExecs) ListWaitingBySessionAcrossWorkspaces(_ context.Context, appName, userID, sessionID string) ([]*agentsv1.CronExecution, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*agentsv1.CronExecution
	for _, e := range r.execs {
		if e.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
			continue
		}
		if e.GetSessionAppName() == appName && e.GetSessionUserId() == userID && e.GetSessionId() == sessionID {
			out = append(out, proto.Clone(e).(*agentsv1.CronExecution))
		}
	}
	return out, nil
}

func (r *memoryCronExecs) CountWaitingByJob(_ context.Context, workspaceID, jobName string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var n int64
	for _, e := range r.execs {
		if e.GetWorkspaceId() == workspaceID && e.GetJobName() == jobName &&
			e.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
			n++
		}
	}
	return n, nil
}

func (r *memoryCronExecs) ListByTimeRange(_ context.Context, _, _ string, _, _ time.Time) ([]*agentsv1.CronExecution, error) {
	return nil, nil
}

func registerCronFixture(
	ctx context.Context,
	r *gin.Engine,
	opts []connect.HandlerOption,
	agentRepo *configmemory.Store,
	runnerSvc *runner.Service,
	sessions *chatSessions,
	telegramRepo telegramrepo.Repository,
) error {
	if _, err := agentRepo.CreateAgent(ctx, cronWorkspace, &agentsv1.Agent{
		AgentId: cronAgentID, Name: "cron-agent", WorkspaceId: cronWorkspace,
		Config: &agentsv1.AgentConfig{Model: chatModel},
	}); err != nil {
		return err
	}
	_ = runnerSvc // scheduler invokes the shared chat runner after reload.

	jobs := &memoryCronJobs{jobs: map[string]*agentsv1.CronJob{}}
	execs := &memoryCronExecs{execs: map[string]*agentsv1.CronExecution{}, order: []string{}}
	sched, err := internalcron.NewScheduler(ctx, runnerSvc, jobs, execs, agentRepo, nil)
	if err != nil {
		return err
	}
	sched.Start()

	cronSvc := application.NewCronJobServiceServer()
	cronSvc.SetScheduler(sched)
	cronSvc.SetExecutionRepo(execs)
	cronSvc.SetTelegramRepo(telegramRepo)

	if err := jobs.Create(ctx, &agentsv1.CronJob{
		Name: cronSeededJob, Schedule: "@daily", AgentId: cronAgentID,
		WorkspaceId: cronWorkspace, Enabled: true,
		Delivery: &agentsv1.CronDelivery{Type: agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_LOG},
	}); err != nil {
		return err
	}

	appName := "cron:" + cronSeededJob
	userID := appName
	sessionID := "sess-wait"
	if err := createSessionWithEvents(ctx, sessions, cronWorkspace, appName, userID, sessionID, "waiting"); err != nil {
		return err
	}

	waiting := &agentsv1.CronExecution{
		Id: cronWaitingExecID, JobName: cronSeededJob, AgentId: cronAgentID,
		Status:         agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT,
		Output:         "Approve deploy?",
		StartedAt:      timestamppb.Now(),
		SessionAppName: appName, SessionUserId: userID,
		SessionId: sessionID, WorkspaceId: cronWorkspace,
	}
	if err := execs.Save(ctx, waiting); err != nil {
		return err
	}

	if _, err := agentRepo.CreateNotifyGroup(ctx, cronWorkspace, &agentsv1.NotifyGroup{
		Name: cronNotifyGroup, WorkspaceId: cronWorkspace, Enabled: true,
	}); err != nil {
		return err
	}

	if _, err := telegramRepo.CreateChannel(ctx, cronWorkspace,
		&agentsv1.TelegramChannel{
			Id: cronTelegramChannel, Key: "cron", Name: "cron", BotId: "999001",
		},
		telegramrepo.ChannelCredentials{BotToken: telegramrepo.Credential{Ciphertext: "unused", KeyID: "contract"}}); err != nil {
		return err
	}
	if _, err := telegramRepo.CreateDestination(ctx, cronWorkspace, &agentsv1.TelegramDestination{
		Id: cronTelegramDestID, Key: "cron-topic", ChannelId: cronTelegramChannel, ChatId: "-10099",
		OutboundEnabled: true,
		Config:          &agentsv1.TelegramDestinationConfig{AgentId: cronAgentID},
	}); err != nil {
		return err
	}

	cronPath, cronHandler := agentsv1connect.NewCronJobServiceHandler(cronSvc, opts...)
	r.Any("/api"+cronPath+"*path", gin.WrapH(http.StripPrefix("/api", cronHandler)))

	notifySvc := application.NewNotifyGroupServiceServer(agentRepo)
	notifySvc.SetTelegramRepo(telegramRepo)
	notifyPath, notifyHandler := agentsv1connect.NewNotifyGroupServiceHandler(notifySvc, opts...)
	r.Any("/api"+notifyPath+"*path", gin.WrapH(http.StripPrefix("/api", notifyHandler)))

	tgSvc := application.NewTelegramDestinationServiceServer(telegramRepo)
	tgPath, tgHandler := agentsv1connect.NewTelegramDestinationServiceHandler(tgSvc, opts...)
	r.Any("/api"+tgPath+"*path", gin.WrapH(http.StripPrefix("/api", tgHandler)))

	return nil
}
