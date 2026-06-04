package cron

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	robfigcron "github.com/robfig/cron/v3"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/notify"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var notifyTargetTimeout = 10 * time.Second

// ErrAgentNotInWorkspace is returned when a cron job references an agent
// that does not exist in the job's workspace. The application layer maps
// this to connect.CodeInvalidArgument so callers see "this agent doesn't
// belong to your workspace" instead of a generic internal error, and the
// job is never persisted in the first place.
var ErrAgentNotInWorkspace = errors.New("cron job agent not found in workspace")

// Scheduler manages cron-based agent execution.
type Scheduler struct {
	cron      *robfigcron.Cron
	runner    runnerService
	execRepo  ExecutionRepo
	jobRepo   JobRepo
	groupRepo configrepo.NotifyGroupRepository
	notifier  *notify.Sender
	ctx       context.Context
	cancelFn  context.CancelFunc

	mu       sync.Mutex
	entryIDs map[string]robfigcron.EntryID // composite "workspace_id:name" -> cron entry ID
}

type runnerService interface {
	HasAgentInWorkspace(workspaceID, name string) bool
	RunSSE(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error)
}

func jobKey(workspaceID, name string) string { return workspaceID + ":" + name }

// NewScheduler creates a scheduler, loads jobs from the repo, and registers them.
func NewScheduler(ctx context.Context, runnerSvc *runner.Service, jobRepo JobRepo, execRepo ExecutionRepo, groupRepo configrepo.NotifyGroupRepository) (*Scheduler, error) {
	logger := log.FromContext(ctx)
	schedCtx, cancel := context.WithCancel(ctx)

	c := robfigcron.New(robfigcron.WithChain(robfigcron.SkipIfStillRunning(robfigcron.DefaultLogger)))

	s := &Scheduler{
		cron:      c,
		runner:    runnerSvc,
		execRepo:  execRepo,
		jobRepo:   jobRepo,
		groupRepo: groupRepo,
		notifier:  notify.NewSender(nil),
		ctx:       schedCtx,
		cancelFn:  cancel,
		entryIDs:  make(map[string]robfigcron.EntryID),
	}

	// Load existing jobs from MongoDB across every workspace.
	jobs, err := jobRepo.ListAll(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("loading cron jobs: %w", err)
	}

	for _, job := range jobs {
		if err := s.registerJob(job); err != nil {
			logger.Error("failed to register cron job", "job", job.GetName(), "err", err)
		}
	}

	logger.Info("cron scheduler initialized", "job_count", len(jobs))
	return s, nil
}

// Start begins the cron scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop gracefully stops the cron scheduler and waits for running jobs to finish.
func (s *Scheduler) Stop() context.Context {
	s.cancelFn()
	return s.cron.Stop()
}

// AddJob persists and schedules a new cron job. Job.WorkspaceId must be set.
func (s *Scheduler) AddJob(ctx context.Context, job *agentsv1.CronJob) error {
	if err := s.validateAgentScope(job); err != nil {
		return err
	}
	if err := s.jobRepo.Create(ctx, job); err != nil {
		return err
	}
	return s.registerJob(job)
}

// UpdateJob updates a persisted cron job and reschedules it.
func (s *Scheduler) UpdateJob(ctx context.Context, job *agentsv1.CronJob) error {
	if err := s.validateAgentScope(job); err != nil {
		return err
	}
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return err
	}
	s.unregisterJob(job.GetWorkspaceId(), job.GetName())
	return s.registerJob(job)
}

// validateAgentScope checks that the cron job's target agent exists inside
// the job's workspace. Called from AddJob/UpdateJob before persistence so a
// caller in workspace A cannot leave a database record referencing
// workspace B's agent, and so the user gets a typed error instead of a
// silent registration failure after the row is already written.
func (s *Scheduler) validateAgentScope(job *agentsv1.CronJob) error {
	if s.runner == nil {
		return nil
	}
	if !s.runner.HasAgentInWorkspace(job.GetWorkspaceId(), job.GetAgentName()) {
		return fmt.Errorf("%w: agent %q in workspace %q", ErrAgentNotInWorkspace, job.GetAgentName(), job.GetWorkspaceId())
	}
	return nil
}

// RemoveJob removes a cron job from persistence and unschedules it.
func (s *Scheduler) RemoveJob(ctx context.Context, workspaceID, name string) error {
	if err := s.jobRepo.Delete(ctx, workspaceID, name); err != nil {
		return err
	}
	s.unregisterJob(workspaceID, name)
	return nil
}

// GetJob returns a cron job by workspace and name.
func (s *Scheduler) GetJob(ctx context.Context, workspaceID, name string) (*agentsv1.CronJob, error) {
	return s.jobRepo.Get(ctx, workspaceID, name)
}

// ListJobs returns all cron jobs in the given workspace.
func (s *Scheduler) ListJobs(ctx context.Context, workspaceID string) ([]*agentsv1.CronJob, error) {
	return s.jobRepo.List(ctx, workspaceID)
}

// ListAllJobs returns every cron job across workspaces.
func (s *Scheduler) ListAllJobs(ctx context.Context) ([]*agentsv1.CronJob, error) {
	return s.jobRepo.ListAll(ctx)
}

// registerJob adds a job to the cron scheduler (in-memory).
func (s *Scheduler) registerJob(job *agentsv1.CronJob) error {
	logger := log.FromContext(s.ctx)

	if !job.GetEnabled() {
		logger.Info("skipping disabled cron job",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"agent", job.GetAgentName(),
		)
		return nil
	}

	if !s.runner.HasAgentInWorkspace(job.GetWorkspaceId(), job.GetAgentName()) {
		return fmt.Errorf("cron job %q references unknown agent %q in workspace %q", job.GetName(), job.GetAgentName(), job.GetWorkspaceId())
	}

	schedule := job.GetSchedule()
	if tz := job.GetTimezone(); tz != "" {
		if _, err := loadTimezone(tz); err != nil {
			return fmt.Errorf("invalid timezone %q: %w", tz, err)
		}
		schedule = fmt.Sprintf("CRON_TZ=%s %s", tz, schedule)
	}

	sched, err := parseSchedule(schedule)
	if err != nil {
		return fmt.Errorf("invalid cron expression %q: %w", job.GetSchedule(), err)
	}

	// Capture job name + workspace for the closure — the proto pointer is safe since we store it in MongoDB.
	jobName := job.GetName()
	workspaceID := job.GetWorkspaceId()
	entryID := s.cron.Schedule(sched, robfigcron.FuncJob(func() {
		// Re-read the job from repo to get the latest config.
		current, err := s.jobRepo.Get(s.ctx, workspaceID, jobName)
		if err != nil {
			logger.Error("failed to load cron job for execution", "job", jobName, "workspace_id", workspaceID, "err", err)
			return
		}
		s.executeJob(current)
	}))

	s.mu.Lock()
	s.entryIDs[jobKey(workspaceID, jobName)] = entryID
	s.mu.Unlock()

	logger.Info("registered cron job",
		"job", job.GetName(),
		"workspace_id", job.GetWorkspaceId(),
		"schedule", job.GetSchedule(),
		"agent", job.GetAgentName(),
		"timezone", job.GetTimezone(),
		"delivery_type", job.GetDelivery().GetType().String(),
	)
	return nil
}

// unregisterJob removes a job from the cron scheduler (in-memory).
func (s *Scheduler) unregisterJob(workspaceID, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := jobKey(workspaceID, name)
	if entryID, ok := s.entryIDs[key]; ok {
		s.cron.Remove(entryID)
		delete(s.entryIDs, key)
	}
}

// RunJobNow loads the named job and executes it immediately, bypassing the
// schedule. The returned execution record is the same shape that scheduled
// runs produce and is persisted to the execution repo.
func (s *Scheduler) RunJobNow(ctx context.Context, workspaceID, name string) (*agentsv1.CronExecution, error) {
	job, err := s.jobRepo.Get(ctx, workspaceID, name)
	if err != nil {
		return nil, err
	}
	if !s.runner.HasAgentInWorkspace(job.GetWorkspaceId(), job.GetAgentName()) {
		return nil, fmt.Errorf("cron job %q references unknown agent %q in workspace %q", name, job.GetAgentName(), job.GetWorkspaceId())
	}
	return s.executeJob(job), nil
}

func (s *Scheduler) executeJob(job *agentsv1.CronJob) *agentsv1.CronExecution {
	logger := log.FromContext(s.ctx)
	startTime := time.Now()

	execID := uuid.New().String()
	userID := fmt.Sprintf("cron:%s", job.GetName())
	sessionID := fmt.Sprintf("cron:%s", job.GetName())
	channelName := fmt.Sprintf("cron:%s", job.GetName())

	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        execID,
		SessionId:   sessionID,
		UserId:      userID,
		ChannelName: channelName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChannelType: "cron",
		WorkspaceId: job.GetWorkspaceId(),
	}

	input := job.GetInput()
	if input == "" {
		input = "execute"
	}

	logger.Info("executing cron job",
		"job", job.GetName(),
		"agent", job.GetAgentName(),
		"workspace_id", job.GetWorkspaceId(),
		"exec_id", execID,
		"input_len", len(input),
		"delivery_type", job.GetDelivery().GetType().String(),
	)

	parts := []*genai.Part{genai.NewPartFromText(input)}
	output, err := s.runner.RunSSE(s.ctx, job.GetAgentName(), parts, "", ctxInfo, nil, nil)

	finishTime := time.Now()

	exec := &agentsv1.CronExecution{
		Id:          execID,
		JobName:     job.GetName(),
		AgentName:   job.GetAgentName(),
		Status:      agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS,
		Input:       job.GetInput(),
		Output:      output,
		StartedAt:   timestamppb.New(startTime),
		FinishedAt:  timestamppb.New(finishTime),
		WorkspaceId: job.GetWorkspaceId(),
	}

	if err != nil {
		exec.Status = agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR
		exec.Output = err.Error()
		logger.Error("cron job execution failed",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", execID,
			"duration", finishTime.Sub(startTime),
			"err", err,
		)
	} else {
		logger.Info("cron job execution succeeded",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", execID,
			"duration", finishTime.Sub(startTime),
			"output_len", len(output),
		)
	}

	// Persist execution record.
	if saveErr := s.execRepo.Save(s.ctx, exec); saveErr != nil {
		logger.Error("failed to save cron execution record",
			"job", job.GetName(),
			"exec_id", execID,
			"err", saveErr,
		)
	}

	// Deliver result.
	s.deliver(job, exec)
	return exec
}

func (s *Scheduler) deliver(job *agentsv1.CronJob, exec *agentsv1.CronExecution) {
	log.FromContext(s.ctx).Debug("delivering cron job result",
		"job", job.GetName(),
		"workspace_id", job.GetWorkspaceId(),
		"exec_id", exec.GetId(),
		"status", exec.GetStatus().String(),
		"delivery_type", job.GetDelivery().GetType().String(),
	)
	delivery := job.GetDelivery()
	if delivery == nil {
		s.deliverLog(job, exec)
		return
	}

	switch delivery.GetType() {
	case agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_WEBHOOK:
		s.deliverWebhook(job, exec)
	case agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_CHANNEL:
		s.deliverChannel(job, exec)
	case agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_NOTIFY_GROUP:
		s.deliverNotifyGroup(job, exec)
	default:
		s.deliverLog(job, exec)
	}
}

func (s *Scheduler) deliverLog(job *agentsv1.CronJob, exec *agentsv1.CronExecution) {
	logger := log.FromContext(s.ctx)
	logger.Info("cron job result",
		"job", job.GetName(),
		"workspace_id", job.GetWorkspaceId(),
		"exec_id", exec.GetId(),
		"status", exec.GetStatus().String(),
		"output", exec.GetOutput(),
		"executed_at", exec.GetStartedAt().AsTime().Format(time.RFC3339),
	)
}

func (s *Scheduler) deliverWebhook(job *agentsv1.CronJob, exec *agentsv1.CronExecution) {
	logger := log.FromContext(s.ctx)
	url := job.GetDelivery().GetWebhookUrl()
	if url == "" {
		logger.Error("webhook delivery configured but no webhook_url set", "job", job.GetName(), "workspace_id", job.GetWorkspaceId(), "exec_id", exec.GetId())
		return
	}

	status := "success"
	if exec.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR {
		status = "error"
	}

	payload := map[string]string{
		"job_name":    exec.GetJobName(),
		"output":      exec.GetOutput(),
		"executed_at": exec.GetStartedAt().AsTime().Format(time.RFC3339),
		"status":      status,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logger.Error("failed to marshal webhook payload", "job", job.GetName(), "workspace_id", job.GetWorkspaceId(), "exec_id", exec.GetId(), "err", err)
		return
	}

	logger.Debug("sending cron webhook delivery",
		"job", job.GetName(),
		"workspace_id", job.GetWorkspaceId(),
		"exec_id", exec.GetId(),
		"payload_len", len(body),
	)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		logger.Error("webhook delivery failed", "job", job.GetName(), "workspace_id", job.GetWorkspaceId(), "exec_id", exec.GetId(), "url", url, "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		logger.Error("webhook delivery returned non-success status",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", exec.GetId(),
			"url", url,
			"status_code", resp.StatusCode,
		)
		return
	}
	logger.Info("webhook delivery succeeded",
		"job", job.GetName(),
		"workspace_id", job.GetWorkspaceId(),
		"exec_id", exec.GetId(),
		"status_code", resp.StatusCode,
	)
}

func (s *Scheduler) deliverChannel(_ *agentsv1.CronJob, _ *agentsv1.CronExecution) {
	// TODO: implement channel delivery when channel send interface is available.
	logger := log.FromContext(s.ctx)
	logger.Warn("channel delivery not yet implemented, falling back to log delivery")
}

func (s *Scheduler) deliverNotifyGroup(job *agentsv1.CronJob, exec *agentsv1.CronExecution) {
	logger := log.FromContext(s.ctx)
	delivery := job.GetDelivery()
	if delivery.GetNotifyGroupName() == "" {
		logger.Error("notify group delivery configured but no notify_group_name set", "job", job.GetName(), "workspace_id", job.GetWorkspaceId())
		return
	}
	if s.groupRepo == nil || s.notifier == nil {
		logger.Error("notify group delivery configured but notifier is not available", "job", job.GetName(), "workspace_id", job.GetWorkspaceId())
		return
	}

	group, err := s.groupRepo.GetNotifyGroup(s.ctx, job.GetWorkspaceId(), delivery.GetNotifyGroupName())
	if err != nil {
		logger.Error("failed to load notify group for cron delivery",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"notify_group", delivery.GetNotifyGroupName(),
			"err", err,
		)
		return
	}
	if !group.GetEnabled() {
		logger.Info("skipping disabled notify group",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"notify_group", group.GetName(),
		)
		return
	}

	message := notify.Message{
		Title: fmt.Sprintf("Cron job %s: %s", exec.GetJobName(), cronStatus(exec)),
		Text:  exec.GetOutput(),
	}
	for _, target := range group.GetTargets() {
		if target == nil || !target.GetEnabled() {
			continue
		}
		ctx, cancel := context.WithTimeout(s.ctx, notifyTargetTimeout)
		err := s.notifier.Send(ctx, target, message)
		cancel()
		if err != nil {
			logger.Error("notify group target delivery failed",
				"job", job.GetName(),
				"workspace_id", job.GetWorkspaceId(),
				"notify_group", group.GetName(),
				"target", target.GetName(),
				"type", target.GetType().String(),
				"err", err,
			)
			continue
		}
		logger.Info("notify group target delivered",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"notify_group", group.GetName(),
			"target", target.GetName(),
			"type", target.GetType().String(),
		)
	}
}

func cronStatus(exec *agentsv1.CronExecution) string {
	if exec.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR {
		return "error"
	}
	return "success"
}
