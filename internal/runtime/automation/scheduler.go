package automation

import (
	"context"
	"fmt"
	"time"

	"butterfly.orx.me/core/log"
	robfigcron "github.com/robfig/cron/v3"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var automationCronParser = robfigcron.NewParser(
	robfigcron.Minute | robfigcron.Hour | robfigcron.Dom | robfigcron.Month | robfigcron.Dow | robfigcron.Descriptor,
)

// Scheduler registers schedule-triggered automations and delegates execution
// to Engine. Definitions remain persisted through DefinitionRepo.
type Scheduler struct {
	cron    *robfigcron.Cron
	repo    DefinitionRepo
	engine  *Engine
	ctx     context.Context
	cancel  context.CancelFunc
	entryID map[string]robfigcron.EntryID
}

func NewScheduler(ctx context.Context, repo DefinitionRepo, engine *Engine) (*Scheduler, error) {
	schedCtx, cancel := context.WithCancel(ctx)
	s := &Scheduler{
		cron:    robfigcron.New(),
		repo:    repo,
		engine:  engine,
		ctx:     schedCtx,
		cancel:  cancel,
		entryID: make(map[string]robfigcron.EntryID),
	}
	automations, err := repo.ListAll(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("loading automations: %w", err)
	}
	for _, a := range automations {
		if err := s.Register(a); err != nil {
			log.FromContext(ctx).Error("failed to register scheduled automation", "automation", a.GetName(), "workspace_id", a.GetWorkspaceId(), "err", err)
		}
	}
	return s, nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() context.Context {
	s.cancel()
	return s.cron.Stop()
}

func (s *Scheduler) Register(a *agentsv1.Automation) error {
	if !shouldScheduleAutomation(a) {
		return nil
	}
	schedule, err := parseAutomationSchedule(a)
	if err != nil {
		return err
	}
	workspaceID := a.GetWorkspaceId()
	name := a.GetName()
	entryID := s.cron.Schedule(schedule, robfigcron.FuncJob(func() {
		current, err := s.repo.Get(s.ctx, workspaceID, name)
		if err != nil {
			log.FromContext(s.ctx).Error("failed to load automation for scheduled execution", "automation", name, "workspace_id", workspaceID, "err", err)
			return
		}
		if !shouldScheduleAutomation(current) {
			return
		}
		if _, err := s.engine.Execute(s.ctx, current, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_SCHEDULE, "{}"); err != nil {
			log.FromContext(s.ctx).Error("scheduled automation execution failed", "automation", name, "workspace_id", workspaceID, "err", err)
		}
	}))
	s.entryID[automationID(workspaceID, name)] = entryID
	return nil
}

func (s *Scheduler) Reschedule(a *agentsv1.Automation) error {
	s.Unregister(a.GetWorkspaceId(), a.GetName())
	return s.Register(a)
}

func (s *Scheduler) Unregister(workspaceID, name string) {
	key := automationID(workspaceID, name)
	if id, ok := s.entryID[key]; ok {
		s.cron.Remove(id)
		delete(s.entryID, key)
	}
}

func shouldScheduleAutomation(a *agentsv1.Automation) bool {
	return a != nil &&
		a.GetEnabled() &&
		a.GetTrigger().GetType() == agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_SCHEDULE &&
		a.GetTrigger().GetSchedule().GetSchedule() != ""
}

func parseAutomationSchedule(a *agentsv1.Automation) (robfigcron.Schedule, error) {
	schedule := a.GetTrigger().GetSchedule().GetSchedule()
	if tz := a.GetTrigger().GetSchedule().GetTimezone(); tz != "" {
		if _, err := time.LoadLocation(tz); err != nil {
			return nil, fmt.Errorf("invalid timezone %q: %w", tz, err)
		}
		schedule = fmt.Sprintf("CRON_TZ=%s %s", tz, schedule)
	}
	parsed, err := automationCronParser.Parse(schedule)
	if err != nil {
		return nil, fmt.Errorf("invalid automation schedule %q: %w", a.GetTrigger().GetSchedule().GetSchedule(), err)
	}
	return parsed, nil
}
