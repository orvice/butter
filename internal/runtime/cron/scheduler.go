package cron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	robfigcron "github.com/robfig/cron/v3"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Scheduler manages cron-based agent execution.
type Scheduler struct {
	cron     *robfigcron.Cron
	runner   *runner.Service
	execRepo ExecutionRepo
	jobRepo  JobRepo
	ctx      context.Context
	cancelFn context.CancelFunc

	mu       sync.Mutex
	entryIDs map[string]robfigcron.EntryID // job name -> cron entry ID
}

// NewScheduler creates a scheduler, loads jobs from the repo, and registers them.
func NewScheduler(ctx context.Context, runnerSvc *runner.Service, jobRepo JobRepo, execRepo ExecutionRepo) (*Scheduler, error) {
	logger := log.FromContext(ctx)
	schedCtx, cancel := context.WithCancel(ctx)

	c := robfigcron.New(robfigcron.WithChain(robfigcron.SkipIfStillRunning(robfigcron.DefaultLogger)))

	s := &Scheduler{
		cron:     c,
		runner:   runnerSvc,
		execRepo: execRepo,
		jobRepo:  jobRepo,
		ctx:      schedCtx,
		cancelFn: cancel,
		entryIDs: make(map[string]robfigcron.EntryID),
	}

	// Load existing jobs from MongoDB.
	jobs, err := jobRepo.List(ctx)
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

// AddJob persists and schedules a new cron job.
func (s *Scheduler) AddJob(ctx context.Context, job *agentsv1.CronJob) error {
	if err := s.jobRepo.Create(ctx, job); err != nil {
		return err
	}
	return s.registerJob(job)
}

// UpdateJob updates a persisted cron job and reschedules it.
func (s *Scheduler) UpdateJob(ctx context.Context, job *agentsv1.CronJob) error {
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return err
	}
	s.unregisterJob(job.GetName())
	return s.registerJob(job)
}

// RemoveJob removes a cron job from persistence and unschedules it.
func (s *Scheduler) RemoveJob(ctx context.Context, name string) error {
	if err := s.jobRepo.Delete(ctx, name); err != nil {
		return err
	}
	s.unregisterJob(name)
	return nil
}

// GetJob returns a cron job by name.
func (s *Scheduler) GetJob(ctx context.Context, name string) (*agentsv1.CronJob, error) {
	return s.jobRepo.Get(ctx, name)
}

// ListJobs returns all cron jobs.
func (s *Scheduler) ListJobs(ctx context.Context) ([]*agentsv1.CronJob, error) {
	return s.jobRepo.List(ctx)
}

// registerJob adds a job to the cron scheduler (in-memory).
func (s *Scheduler) registerJob(job *agentsv1.CronJob) error {
	logger := log.FromContext(s.ctx)

	if !job.GetEnabled() {
		logger.Info("skipping disabled cron job", "job", job.GetName())
		return nil
	}

	if !s.runner.HasAgent(job.GetAgentName()) {
		return fmt.Errorf("cron job %q references unknown agent %q", job.GetName(), job.GetAgentName())
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

	// Capture job name for the closure — the proto pointer is safe since we store it in MongoDB.
	jobName := job.GetName()
	entryID := s.cron.Schedule(sched, robfigcron.FuncJob(func() {
		// Re-read the job from repo to get the latest config.
		current, err := s.jobRepo.Get(s.ctx, jobName)
		if err != nil {
			logger.Error("failed to load cron job for execution", "job", jobName, "err", err)
			return
		}
		s.executeJob(current)
	}))

	s.mu.Lock()
	s.entryIDs[job.GetName()] = entryID
	s.mu.Unlock()

	logger.Info("registered cron job",
		"job", job.GetName(),
		"schedule", job.GetSchedule(),
		"agent", job.GetAgentName(),
		"timezone", job.GetTimezone(),
	)
	return nil
}

// unregisterJob removes a job from the cron scheduler (in-memory).
func (s *Scheduler) unregisterJob(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entryID, ok := s.entryIDs[name]; ok {
		s.cron.Remove(entryID)
		delete(s.entryIDs, name)
	}
}

func (s *Scheduler) executeJob(job *agentsv1.CronJob) {
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
	}

	input := job.GetInput()
	if input == "" {
		input = "execute"
	}

	logger.Info("executing cron job",
		"job", job.GetName(),
		"agent", job.GetAgentName(),
		"exec_id", execID,
	)

	parts := []*genai.Part{genai.NewPartFromText(input)}
	output, err := s.runner.Run(s.ctx, job.GetAgentName(), parts, "", ctxInfo, nil, nil)

	finishTime := time.Now()

	exec := &agentsv1.CronExecution{
		Id:         execID,
		JobName:    job.GetName(),
		AgentName:  job.GetAgentName(),
		Status:     agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS,
		Input:      job.GetInput(),
		Output:     output,
		StartedAt:  timestamppb.New(startTime),
		FinishedAt: timestamppb.New(finishTime),
	}

	if err != nil {
		exec.Status = agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR
		exec.Output = err.Error()
		logger.Error("cron job execution failed",
			"job", job.GetName(),
			"exec_id", execID,
			"duration", finishTime.Sub(startTime),
			"err", err,
		)
	} else {
		logger.Info("cron job execution succeeded",
			"job", job.GetName(),
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
}

func (s *Scheduler) deliver(job *agentsv1.CronJob, exec *agentsv1.CronExecution) {
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
	default:
		s.deliverLog(job, exec)
	}
}

func (s *Scheduler) deliverLog(job *agentsv1.CronJob, exec *agentsv1.CronExecution) {
	logger := log.FromContext(s.ctx)
	logger.Info("cron job result",
		"job", job.GetName(),
		"status", exec.GetStatus().String(),
		"output", exec.GetOutput(),
		"executed_at", exec.GetStartedAt().AsTime().Format(time.RFC3339),
	)
}

func (s *Scheduler) deliverWebhook(job *agentsv1.CronJob, exec *agentsv1.CronExecution) {
	logger := log.FromContext(s.ctx)
	url := job.GetDelivery().GetWebhookUrl()
	if url == "" {
		logger.Error("webhook delivery configured but no webhook_url set", "job", job.GetName())
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
		logger.Error("failed to marshal webhook payload", "job", job.GetName(), "err", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		logger.Error("webhook delivery failed", "job", job.GetName(), "url", url, "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		logger.Error("webhook delivery returned non-success status",
			"job", job.GetName(),
			"url", url,
			"status_code", resp.StatusCode,
		)
	}
}

func (s *Scheduler) deliverChannel(_ *agentsv1.CronJob, _ *agentsv1.CronExecution) {
	// TODO: implement channel delivery when channel send interface is available.
	logger := log.FromContext(s.ctx)
	logger.Warn("channel delivery not yet implemented, falling back to log delivery")
}
