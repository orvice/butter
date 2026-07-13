package cron

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	robfigcron "github.com/robfig/cron/v3"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/notify"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var notifyTargetTimeout = 10 * time.Second
var channelDeliveryTimeout = 10 * time.Second

// ErrAgentNotInWorkspace is returned when a cron job references an agent
// that does not exist in the job's workspace. The application layer maps
// this to connect.CodeInvalidArgument so callers see "this agent doesn't
// belong to your workspace" instead of a generic internal error, and the
// job is never persisted in the first place.
var ErrAgentNotInWorkspace = errors.New("cron job agent not found in workspace")

// Scheduler manages cron-based agent execution.
type Scheduler struct {
	cron          *robfigcron.Cron
	runner        runnerService
	execRepo      ExecutionRepo
	jobRepo       JobRepo
	groupRepo     configrepo.NotifyGroupRepository
	channelRepo   configrepo.ChannelRepository
	notifier      *notify.Sender
	channelSender channelSender
	ctx           context.Context
	cancelFn      context.CancelFunc

	mu       sync.Mutex
	entryIDs map[string]robfigcron.EntryID // composite "workspace_id:name" -> cron entry ID
	running  map[string]*runningJob        // composite "workspace_id:name" -> active execution
}

type runnerService interface {
	HasAgentInWorkspace(workspaceID, name string) bool
	RunTurnSSE(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (*runner.TurnResult, error)
}

type channelSender interface {
	Send(ctx context.Context, channel *agentsv1.AgentChannel, chatID, text string) error
}

type runningJob struct {
	cancel context.CancelFunc
	done   chan struct{}
}

func jobKey(workspaceID, name string) string { return workspaceID + ":" + name }

// cronSessionPrefix marks every session coordinate the scheduler creates;
// HandleTurn relies on it to skip ordinary chat traffic.
const cronSessionPrefix = "cron:"

// cronJobScope is the app-name/user coordinate shared by every run of a job.
func cronJobScope(jobName string) string { return cronSessionPrefix + jobName }

// cronSessionID is the session coordinate of one execution. The session is
// per-execution, not per-job: while an execution waits on a Human Input node,
// the job keeps firing on schedule, and a rerun posting its input onto the
// same session would answer the pending Interrupt in the human's place (ADR
// 0002 answers FIFO) and close the waiting record with unrelated output.
func cronSessionID(jobName, execID string) string {
	return cronJobScope(jobName) + ":" + execID
}

func validateJobConfig(job *agentsv1.CronJob) error {
	if job == nil {
		return errors.New("cron job is required")
	}
	if job.GetName() == "" {
		return errors.New("cron job name is required")
	}
	if job.GetSchedule() == "" {
		return errors.New("cron job schedule is required")
	}
	if job.GetAgentName() == "" {
		return errors.New("cron job agent_name is required")
	}
	if _, err := parseJobSchedule(job); err != nil {
		return err
	}
	if timeout := job.GetTimeout(); timeout != nil && timeout.AsDuration() < 0 {
		return errors.New("cron job timeout must be non-negative")
	}
	if retry := job.GetRetry(); retry != nil {
		if retry.GetMaxAttempts() < 0 {
			return errors.New("cron retry max_attempts must be non-negative")
		}
		if retry.GetBackoff() != nil && retry.GetBackoff().AsDuration() < 0 {
			return errors.New("cron retry backoff must be non-negative")
		}
	}
	switch job.GetConcurrencyPolicy() {
	case agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_UNSPECIFIED,
		agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_SKIP,
		agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_QUEUE,
		agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_REPLACE,
		agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_ALLOW:
	default:
		return fmt.Errorf("unsupported cron concurrency policy %s", job.GetConcurrencyPolicy())
	}
	switch job.GetNotifyOn() {
	case agentsv1.CronNotifyOn_CRON_NOTIFY_ON_UNSPECIFIED,
		agentsv1.CronNotifyOn_CRON_NOTIFY_ON_ALWAYS,
		agentsv1.CronNotifyOn_CRON_NOTIFY_ON_FAILURE,
		agentsv1.CronNotifyOn_CRON_NOTIFY_ON_SUCCESS:
	default:
		return fmt.Errorf("unsupported cron notify_on policy %s", job.GetNotifyOn())
	}
	if job.GetMaxOutputBytes() < 0 {
		return errors.New("cron max_output_bytes must be non-negative")
	}
	return nil
}

func parseJobSchedule(job *agentsv1.CronJob) (robfigcron.Schedule, error) {
	schedule := job.GetSchedule()
	if tz := job.GetTimezone(); tz != "" {
		if _, err := loadTimezone(tz); err != nil {
			return nil, fmt.Errorf("invalid timezone %q: %w", tz, err)
		}
		schedule = fmt.Sprintf("CRON_TZ=%s %s", tz, schedule)
	}
	sched, err := parseSchedule(schedule)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression %q: %w", job.GetSchedule(), err)
	}
	return sched, nil
}

func effectiveConcurrencyPolicy(job *agentsv1.CronJob) agentsv1.CronConcurrencyPolicy {
	if job.GetConcurrencyPolicy() == agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_UNSPECIFIED {
		return agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_SKIP
	}
	return job.GetConcurrencyPolicy()
}

func effectiveInput(job *agentsv1.CronJob) string {
	if job.GetInput() == "" {
		return "execute"
	}
	return job.GetInput()
}

// NewScheduler creates a scheduler, loads jobs from the repo, and registers them.
func NewScheduler(ctx context.Context, runnerSvc *runner.Service, jobRepo JobRepo, execRepo ExecutionRepo, groupRepo configrepo.NotifyGroupRepository, channelRepo configrepo.ChannelRepository) (*Scheduler, error) {
	logger := log.FromContext(ctx)
	schedCtx, cancel := context.WithCancel(ctx)

	c := robfigcron.New()

	s := &Scheduler{
		cron:          c,
		runner:        runnerSvc,
		execRepo:      execRepo,
		jobRepo:       jobRepo,
		groupRepo:     groupRepo,
		channelRepo:   channelRepo,
		notifier:      notify.NewSender(nil),
		channelSender: channel.NewSender(),
		ctx:           schedCtx,
		cancelFn:      cancel,
		entryIDs:      make(map[string]robfigcron.EntryID),
		running:       make(map[string]*runningJob),
	}

	// Close the ADR 0003 loop: the scheduler observes every runner turn so a
	// session reply that completes a paused workflow also completes the
	// session's WAITING_INPUT executions, whatever entry point the reply used.
	if runnerSvc != nil {
		runnerSvc.AddTurnListener(s.HandleTurn)
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
	if err := validateJobConfig(job); err != nil {
		return err
	}
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
	if err := validateJobConfig(job); err != nil {
		return err
	}
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

	sched, err := parseJobSchedule(job)
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
		s.executeJobWithTrigger(current, agentsv1.CronExecutionTriggerType_CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE)
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
	return s.executeJobWithTrigger(job, agentsv1.CronExecutionTriggerType_CRON_EXECUTION_TRIGGER_TYPE_MANUAL), nil
}

func (s *Scheduler) executeJob(job *agentsv1.CronJob) *agentsv1.CronExecution {
	return s.executeJobWithTrigger(job, agentsv1.CronExecutionTriggerType_CRON_EXECUTION_TRIGGER_TYPE_UNSPECIFIED)
}

func (s *Scheduler) executeJobWithTrigger(job *agentsv1.CronJob, trigger agentsv1.CronExecutionTriggerType) *agentsv1.CronExecution {
	policy := effectiveConcurrencyPolicy(job)
	if policy == agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_ALLOW {
		return s.runJob(job, trigger, s.ctx)
	}

	key := jobKey(job.GetWorkspaceId(), job.GetName())
	for {
		s.mu.Lock()
		if s.running == nil {
			s.running = make(map[string]*runningJob)
		}
		current := s.running[key]
		if current == nil {
			runCtx, cancel := context.WithCancel(s.ctx)
			active := &runningJob{cancel: cancel, done: make(chan struct{})}
			s.running[key] = active
			s.mu.Unlock()

			defer func() {
				cancel()
				s.mu.Lock()
				if s.running[key] == active {
					delete(s.running, key)
				}
				s.mu.Unlock()
				close(active.done)
			}()
			return s.runJob(job, trigger, runCtx)
		}

		switch policy {
		case agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_REPLACE:
			current.cancel()
			done := current.done
			s.mu.Unlock()
			select {
			case <-done:
				continue
			case <-s.ctx.Done():
				return s.recordSkipped(job, trigger, "scheduler stopped before replacing running execution")
			}
		case agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_QUEUE:
			done := current.done
			s.mu.Unlock()
			select {
			case <-done:
				continue
			case <-s.ctx.Done():
				return s.recordSkipped(job, trigger, "scheduler stopped before queued execution started")
			}
		default:
			s.mu.Unlock()
			return s.recordSkipped(job, trigger, "previous execution still running")
		}
	}
}

func (s *Scheduler) runJob(job *agentsv1.CronJob, trigger agentsv1.CronExecutionTriggerType, baseCtx context.Context) *agentsv1.CronExecution {
	logger := log.FromContext(s.ctx)
	startTime := time.Now()

	execID := uuid.New().String()
	userID := cronJobScope(job.GetName())
	sessionID := cronSessionID(job.GetName(), execID)
	channelName := cronJobScope(job.GetName())

	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        execID,
		SessionId:   sessionID,
		UserId:      userID,
		ChannelName: channelName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChannelType: "cron",
		WorkspaceId: job.GetWorkspaceId(),
	}

	input := effectiveInput(job)

	logger.Info("executing cron job",
		"job", job.GetName(),
		"agent", job.GetAgentName(),
		"workspace_id", job.GetWorkspaceId(),
		"exec_id", execID,
		"input_len", len(input),
		"delivery_type", job.GetDelivery().GetType().String(),
	)

	parts := []*genai.Part{genai.NewPartFromText(input)}
	runCtx := baseCtx
	cancel := func() {}
	if timeout := job.GetTimeout(); timeout != nil && timeout.AsDuration() > 0 {
		runCtx, cancel = context.WithTimeout(baseCtx, timeout.AsDuration())
	}
	defer cancel()

	turn, err, attemptCount := s.runWithRetry(runCtx, job, parts, ctxInfo)

	finishTime := time.Now()
	duration := finishTime.Sub(startTime)
	storedOutput, truncated := truncateOutput(turn.Output, job.GetMaxOutputBytes())
	errText := ""
	if err != nil {
		errText = err.Error()
		if storedOutput == "" {
			storedOutput, truncated = truncateOutput(errText, job.GetMaxOutputBytes())
		}
	}

	exec := &agentsv1.CronExecution{
		Id:           execID,
		JobName:      job.GetName(),
		AgentName:    job.GetAgentName(),
		Status:       agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS,
		Input:        input,
		Output:       storedOutput,
		StartedAt:    timestamppb.New(startTime),
		FinishedAt:   timestamppb.New(finishTime),
		Error:        errText,
		DurationMs:   duration.Milliseconds(),
		AttemptCount: attemptCount,
		TriggerType:  trigger,
		Truncated:    truncated,
		WorkspaceId:  job.GetWorkspaceId(),
	}

	switch {
	case err != nil:
		if errors.Is(err, context.Canceled) {
			exec.Status = agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_CANCELLED
		} else {
			exec.Status = agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR
		}
		logger.Error("cron job execution failed",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", execID,
			"duration", duration,
			"attempt_count", attemptCount,
			"err", err,
		)
	case turn.Interrupted():
		// The workflow paused on a Human Input node (ADR 0003): the execution
		// is not finished — it waits for a human answer on the run's session.
		// The session coordinates recorded here are what ReplySession needs.
		exec.Status = agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT
		exec.FinishedAt = nil
		exec.DurationMs = 0
		exec.SessionAppName = channelName
		exec.SessionUserId = userID
		exec.SessionId = sessionID
		logger.Info("cron job execution waiting for human input",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", execID,
			"attempt_count", attemptCount,
			"pending_interrupts", len(turn.Pending),
		)
	default:
		logger.Info("cron job execution succeeded",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", execID,
			"duration", duration,
			"attempt_count", attemptCount,
			"output_len", len(turn.Output),
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

func (s *Scheduler) runWithRetry(ctx context.Context, job *agentsv1.CronJob, parts []*genai.Part, ctxInfo *agentsv1.ContextInfo) (*runner.TurnResult, error, int32) {
	totalAttempts := int32(1)
	if retry := job.GetRetry(); retry != nil && retry.GetMaxAttempts() > 0 {
		totalAttempts += retry.GetMaxAttempts()
	}
	var (
		turn *runner.TurnResult
		err  error
	)
	for attempt := int32(1); attempt <= totalAttempts; attempt++ {
		turn, err = s.runner.RunTurnSSE(ctx, job.GetAgentName(), parts, "", ctxInfo, nil, nil)
		if turn == nil {
			turn = &runner.TurnResult{}
		}
		if err == nil {
			return turn, nil, attempt
		}
		if attempt == totalAttempts || ctx.Err() != nil {
			if ctx.Err() != nil {
				return turn, ctx.Err(), attempt
			}
			return turn, err, attempt
		}
		backoff := time.Duration(0)
		if retry := job.GetRetry(); retry != nil && retry.GetBackoff() != nil {
			backoff = retry.GetBackoff().AsDuration()
		}
		if backoff <= 0 {
			continue
		}
		timer := time.NewTimer(backoff)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return turn, ctx.Err(), attempt
		}
	}
	return turn, err, totalAttempts
}

// HandleTurn observes the outcome of every agent turn (registered as a
// runner turn listener). It closes ADR 0003's loop: a human answers a paused
// cron workflow by messaging its session — through ReplySession or any other
// entry point — and when a turn on that session ends with no pending
// Interrupt, the session's WAITING_INPUT executions reach their terminal
// state and the final output is delivered through the job's targets.
//
// The initial pausing run is invisible here by construction: its listener
// fires before runJob persists the WAITING_INPUT record, so no record
// matches. A turn that errored or still has pending Interrupts leaves the
// records waiting.
func (s *Scheduler) HandleTurn(ctxInfo *agentsv1.ContextInfo, turn *runner.TurnResult, runErr error) {
	if ctxInfo == nil || turn == nil || runErr != nil || turn.Interrupted() {
		return
	}
	// Cron sessions are the only ones that can carry waiting executions;
	// skip the repo lookup for ordinary chat traffic.
	if !strings.HasPrefix(ctxInfo.GetSessionId(), cronSessionPrefix) {
		return
	}
	logger := log.FromContext(s.ctx)
	waiting, err := s.execRepo.ListWaitingBySessionAcrossWorkspaces(s.ctx, ctxInfo.GetChannelName(), ctxInfo.GetUserId(), ctxInfo.GetSessionId())
	if err != nil {
		logger.Error("failed to look up waiting cron executions",
			"session_id", ctxInfo.GetSessionId(),
			"err", err,
		)
		return
	}

	now := time.Now()
	for _, exec := range waiting {
		job, jobErr := s.jobRepo.Get(s.ctx, exec.GetWorkspaceId(), exec.GetJobName())
		var maxOutputBytes int32
		if jobErr == nil {
			maxOutputBytes = job.GetMaxOutputBytes()
		}
		exec.Output, exec.Truncated = truncateOutput(turn.Output, maxOutputBytes)
		exec.Status = agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS
		exec.FinishedAt = timestamppb.New(now)
		exec.DurationMs = now.Sub(exec.GetStartedAt().AsTime()).Milliseconds()

		if saveErr := s.execRepo.Save(s.ctx, exec); saveErr != nil {
			logger.Error("failed to complete waiting cron execution",
				"job", exec.GetJobName(),
				"workspace_id", exec.GetWorkspaceId(),
				"exec_id", exec.GetId(),
				"err", saveErr,
			)
			continue
		}
		logger.Info("waiting cron execution completed by session reply",
			"job", exec.GetJobName(),
			"workspace_id", exec.GetWorkspaceId(),
			"exec_id", exec.GetId(),
			"duration_ms", exec.GetDurationMs(),
		)
		if jobErr != nil {
			// The job was deleted while its execution waited; the record is
			// closed but there is no delivery config left to notify with.
			logger.Warn("completed waiting execution without delivery: job no longer exists",
				"job", exec.GetJobName(),
				"workspace_id", exec.GetWorkspaceId(),
				"exec_id", exec.GetId(),
				"err", jobErr,
			)
			continue
		}
		s.deliver(job, exec)
	}
}

// HandleSessionDeleted observes session deletions (registered as a session
// delete listener on the SessionService RPC). Deleting a paused session is
// the documented way to abandon its workflow (ADR 0002), so the session's
// WAITING_INPUT executions transition to CANCELLED with a reason and the
// cancellation is delivered through the job's targets — otherwise the record
// waits forever on a session that no longer exists (issue #132).
func (s *Scheduler) HandleSessionDeleted(appName, userID, sessionID string) {
	if !strings.HasPrefix(sessionID, cronSessionPrefix) {
		return
	}
	logger := log.FromContext(s.ctx)
	waiting, err := s.execRepo.ListWaitingBySessionAcrossWorkspaces(s.ctx, appName, userID, sessionID)
	if err != nil {
		logger.Error("failed to look up waiting cron executions for deleted session",
			"session_id", sessionID,
			"err", err,
		)
		return
	}

	now := time.Now()
	for _, exec := range waiting {
		exec.Status = agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_CANCELLED
		exec.Error = "session deleted before a human answered"
		exec.FinishedAt = timestamppb.New(now)
		exec.DurationMs = now.Sub(exec.GetStartedAt().AsTime()).Milliseconds()

		if saveErr := s.execRepo.Save(s.ctx, exec); saveErr != nil {
			logger.Error("failed to cancel waiting cron execution",
				"job", exec.GetJobName(),
				"workspace_id", exec.GetWorkspaceId(),
				"exec_id", exec.GetId(),
				"err", saveErr,
			)
			continue
		}
		logger.Info("waiting cron execution cancelled by session deletion",
			"job", exec.GetJobName(),
			"workspace_id", exec.GetWorkspaceId(),
			"exec_id", exec.GetId(),
		)
		job, jobErr := s.jobRepo.Get(s.ctx, exec.GetWorkspaceId(), exec.GetJobName())
		if jobErr != nil {
			logger.Warn("cancelled waiting execution without delivery: job no longer exists",
				"job", exec.GetJobName(),
				"workspace_id", exec.GetWorkspaceId(),
				"exec_id", exec.GetId(),
				"err", jobErr,
			)
			continue
		}
		s.deliver(job, exec)
	}
}

func (s *Scheduler) recordSkipped(job *agentsv1.CronJob, trigger agentsv1.CronExecutionTriggerType, reason string) *agentsv1.CronExecution {
	now := time.Now()
	exec := &agentsv1.CronExecution{
		Id:            uuid.New().String(),
		JobName:       job.GetName(),
		AgentName:     job.GetAgentName(),
		Status:        agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SKIPPED,
		Input:         effectiveInput(job),
		StartedAt:     timestamppb.New(now),
		FinishedAt:    timestamppb.New(now),
		AttemptCount:  0,
		TriggerType:   trigger,
		SkippedReason: reason,
		WorkspaceId:   job.GetWorkspaceId(),
	}
	if s.execRepo != nil {
		if err := s.execRepo.Save(s.ctx, exec); err != nil {
			log.FromContext(s.ctx).Error("failed to save skipped cron execution record",
				"job", job.GetName(),
				"workspace_id", job.GetWorkspaceId(),
				"exec_id", exec.GetId(),
				"err", err,
			)
		}
	}
	log.FromContext(s.ctx).Warn("skipping cron job execution",
		"job", job.GetName(),
		"workspace_id", job.GetWorkspaceId(),
		"reason", reason,
	)
	s.deliver(job, exec)
	return exec
}

func truncateOutput(s string, maxBytes int32) (string, bool) {
	if maxBytes <= 0 || len(s) <= int(maxBytes) {
		return s, false
	}
	max := int(maxBytes)
	if max <= 0 {
		return "", true
	}
	for max > 0 && !utf8.ValidString(s[:max]) {
		max--
	}
	return s[:max], true
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
	if !shouldDeliver(job, exec) {
		log.FromContext(s.ctx).Debug("skipping cron delivery due to notify policy",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", exec.GetId(),
			"status", exec.GetStatus().String(),
			"notify_on", job.GetNotifyOn().String(),
		)
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

	payload := map[string]string{
		"execution_id": exec.GetId(),
		"job_name":     exec.GetJobName(),
		"output":       exec.GetOutput(),
		"error":        exec.GetError(),
		"executed_at":  exec.GetStartedAt().AsTime().Format(time.RFC3339),
		"status":       cronStatus(exec),
	}
	if exec.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
		payload["agent_name"] = exec.GetAgentName()
		payload["session_app_name"] = exec.GetSessionAppName()
		payload["session_user_id"] = exec.GetSessionUserId()
		payload["session_id"] = exec.GetSessionId()
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

func (s *Scheduler) deliverChannel(job *agentsv1.CronJob, exec *agentsv1.CronExecution) {
	logger := log.FromContext(s.ctx)
	delivery := job.GetDelivery()
	if delivery.GetChannelName() == "" || delivery.GetChatId() == "" {
		logger.Error("channel delivery configured but channel_name or chat_id is empty",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", exec.GetId(),
		)
		return
	}
	if s.channelRepo == nil || s.channelSender == nil {
		logger.Error("channel delivery configured but channel sender is not available",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", exec.GetId(),
			"channel", delivery.GetChannelName(),
		)
		return
	}

	ch, err := s.channelRepo.GetChannel(s.ctx, job.GetWorkspaceId(), delivery.GetChannelName())
	if err != nil {
		logger.Error("failed to load channel for cron delivery",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", exec.GetId(),
			"channel", delivery.GetChannelName(),
			"err", err,
		)
		return
	}
	if !ch.GetEnabled() {
		logger.Info("skipping disabled channel delivery",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", exec.GetId(),
			"channel", ch.GetName(),
		)
		return
	}

	ctx, cancel := context.WithTimeout(s.ctx, channelDeliveryTimeout)
	err = s.channelSender.Send(ctx, ch, delivery.GetChatId(), cronDeliveryMessage(exec))
	cancel()
	if err != nil {
		logger.Error("channel delivery failed",
			"job", job.GetName(),
			"workspace_id", job.GetWorkspaceId(),
			"exec_id", exec.GetId(),
			"channel", ch.GetName(),
			"chat_id", delivery.GetChatId(),
			"err", err,
		)
		return
	}
	logger.Info("channel delivery succeeded",
		"job", job.GetName(),
		"workspace_id", job.GetWorkspaceId(),
		"exec_id", exec.GetId(),
		"channel", ch.GetName(),
		"chat_id", delivery.GetChatId(),
	)
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

	text := exec.GetOutput()
	if coords := replyCoordinates(exec); coords != "" {
		text += "\n" + coords
	}
	message := notify.Message{
		Title: fmt.Sprintf("Cron job %s: %s", exec.GetJobName(), cronStatus(exec)),
		Text:  text,
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
	switch exec.GetStatus() {
	case agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR:
		return "error"
	case agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SKIPPED:
		return "skipped"
	case agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_CANCELLED:
		return "cancelled"
	case agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT:
		return "waiting_input"
	default:
		return "success"
	}
}

func cronDeliveryMessage(exec *agentsv1.CronExecution) string {
	status := cronStatus(exec)
	body := exec.GetOutput()
	if body == "" && exec.GetError() != "" {
		body = exec.GetError()
	}
	if body == "" && exec.GetSkippedReason() != "" {
		body = exec.GetSkippedReason()
	}
	msg := fmt.Sprintf("Cron job %s: %s\nExecution: %s\n%s", exec.GetJobName(), status, exec.GetId(), body)
	if coords := replyCoordinates(exec); coords != "" {
		msg += "\n" + coords
	}
	return msg
}

// replyCoordinates renders the ReplySession coordinates a human needs to
// answer a WAITING_INPUT execution; empty for every other status.
func replyCoordinates(exec *agentsv1.CronExecution) string {
	if exec.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
		return ""
	}
	return fmt.Sprintf(
		"Answer via SessionService.ReplySession: agent_name=%s app_name=%s user_id=%s session_id=%s",
		exec.GetAgentName(), exec.GetSessionAppName(), exec.GetSessionUserId(), exec.GetSessionId(),
	)
}

func shouldDeliver(job *agentsv1.CronJob, exec *agentsv1.CronExecution) bool {
	// A paused workflow's question must always reach a human — suppressing it
	// by notify policy would leave the execution waiting with nobody told.
	if exec.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
		return true
	}
	switch job.GetNotifyOn() {
	case agentsv1.CronNotifyOn_CRON_NOTIFY_ON_FAILURE:
		return exec.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR ||
			exec.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_CANCELLED
	case agentsv1.CronNotifyOn_CRON_NOTIFY_ON_SUCCESS:
		return exec.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS
	default:
		return true
	}
}
