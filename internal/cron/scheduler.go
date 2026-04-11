package cron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	robfigcron "github.com/robfig/cron/v3"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Scheduler manages cron-based agent execution.
type Scheduler struct {
	cron      *robfigcron.Cron
	runner    *runner.Service
	repo      ExecutionRepo
	jobs      []agentsv1.CronJob
	ctx       context.Context
	cancelFn  context.CancelFunc
}

// NewScheduler creates a scheduler and registers all enabled cron jobs.
func NewScheduler(ctx context.Context, runnerSvc *runner.Service, jobs []agentsv1.CronJob, repo ExecutionRepo) (*Scheduler, error) {
	logger := log.FromContext(ctx)
	schedCtx, cancel := context.WithCancel(ctx)

	c := robfigcron.New(robfigcron.WithChain(robfigcron.SkipIfStillRunning(robfigcron.DefaultLogger)))

	s := &Scheduler{
		cron:     c,
		runner:   runnerSvc,
		repo:     repo,
		jobs:     jobs,
		ctx:      schedCtx,
		cancelFn: cancel,
	}

	for i := range jobs {
		job := &jobs[i]
		if !job.GetEnabled() {
			logger.Info("skipping disabled cron job", "job", job.GetName())
			continue
		}

		if !runnerSvc.HasAgent(job.GetAgentName()) {
			logger.Error("cron job references unknown agent, skipping",
				"job", job.GetName(),
				"agent", job.GetAgentName(),
			)
			continue
		}

		schedule := job.GetSchedule()

		if tz := job.GetTimezone(); tz != "" {
			if _, err := loadTimezone(tz); err != nil {
				logger.Error("invalid timezone for cron job, skipping",
					"job", job.GetName(),
					"timezone", tz,
					"err", err,
				)
				continue
			}
			schedule = fmt.Sprintf("CRON_TZ=%s %s", tz, schedule)
		}

		sched, err := parseSchedule(schedule)
		if err != nil {
			logger.Error("invalid cron expression, skipping",
				"job", job.GetName(),
				"schedule", job.GetSchedule(),
				"err", err,
			)
			continue
		}

		c.Schedule(sched, robfigcron.FuncJob(func() {
			s.executeJob(job)
		}))

		logger.Info("registered cron job",
			"job", job.GetName(),
			"schedule", job.GetSchedule(),
			"agent", job.GetAgentName(),
			"timezone", job.GetTimezone(),
		)
	}

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

// Jobs returns the configured cron jobs.
func (s *Scheduler) Jobs() []agentsv1.CronJob {
	return s.jobs
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
	if saveErr := s.repo.Save(s.ctx, exec); saveErr != nil {
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
