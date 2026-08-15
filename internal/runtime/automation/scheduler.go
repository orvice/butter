package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	robfigcron "github.com/robfig/cron/v3"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var automationCronParser = robfigcron.NewParser(
	robfigcron.Minute | robfigcron.Hour | robfigcron.Dom | robfigcron.Month | robfigcron.Dow | robfigcron.Descriptor,
)

// SchedulerLeaseKey elects the single Pod whose cron fires execute. Every Pod
// registers every schedule (so failover needs no re-registration); leadership
// decides which Pod's fire runs.
const SchedulerLeaseKey = "butter:automation:lease:scheduler"

// SchedulerLeaseTTL bounds how long a crashed leader blocks scheduled fires.
// The renewal tick must be comfortably shorter than this.
const SchedulerLeaseTTL = 60 * time.Second

// schedulerLeaderTick is how often the leader loop re-acquires (renews) the
// lease. TTL/4 keeps two missed ticks from losing leadership.
const schedulerLeaderTick = SchedulerLeaseTTL / 4

// SchedulerLeader elects the single instance whose cron fires execute.
// Optional: without one, every instance executes its own fires (single-Pod
// deployments and tests).
type SchedulerLeader interface {
	// Acquire takes or keeps the lease; it doubles as the renewal.
	Acquire(ctx context.Context) (bool, error)
	// Release gives the lease up immediately on graceful shutdown.
	Release(ctx context.Context) error
}

// Scheduler registers schedule-triggered automations and delegates execution
// to Engine. Definitions remain persisted through DefinitionRepo.
type Scheduler struct {
	cron   *robfigcron.Cron
	repo   DefinitionRepo
	engine *Engine
	ctx    context.Context
	cancel context.CancelFunc

	mu      sync.Mutex
	entryID map[string]robfigcron.EntryID

	leaderMu   sync.RWMutex
	leader     SchedulerLeader
	holdsLease bool
	leaderDone chan struct{}
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

// SetLeader wires leader election. Must be called before Start.
func (s *Scheduler) SetLeader(leader SchedulerLeader) {
	s.leaderMu.Lock()
	defer s.leaderMu.Unlock()
	s.leader = leader
}

func (s *Scheduler) Start() {
	s.startLeaderLoop()
	s.cron.Start()
}

func (s *Scheduler) Stop() context.Context {
	s.cancel()
	s.leaderMu.RLock()
	leader := s.leader
	done := s.leaderDone
	s.leaderMu.RUnlock()
	if done != nil {
		<-done
	}
	if leader != nil {
		// Release on a detached context so the next leader does not wait out
		// the TTL after a graceful shutdown.
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), 5*time.Second)
		_ = leader.Release(releaseCtx)
		cancel()
	}
	return s.cron.Stop()
}

// startLeaderLoop keeps leadership fresh in the background so a fire only has
// to consult the cached verdict. Without a configured leader this Pod always
// executes.
func (s *Scheduler) startLeaderLoop() {
	s.leaderMu.Lock()
	leader := s.leader
	if leader == nil {
		s.holdsLease = true
		s.leaderMu.Unlock()
		return
	}
	done := make(chan struct{})
	s.leaderDone = done
	s.leaderMu.Unlock()

	s.tickLeader(leader)
	go func() {
		defer close(done)
		ticker := time.NewTicker(schedulerLeaderTick)
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.tickLeader(leader)
			}
		}
	}()
}

func (s *Scheduler) tickLeader(leader SchedulerLeader) {
	held, err := leader.Acquire(s.ctx)
	if err != nil {
		// Redis being unreachable must not silently duplicate fires across the
		// fleet: keep the last verdict a leader had, and let a non-leader stay
		// passive until leadership can be evaluated again.
		log.FromContext(s.ctx).Warn("automation scheduler could not evaluate leadership", "err", err)
		return
	}
	s.leaderMu.Lock()
	s.holdsLease = held
	s.leaderMu.Unlock()
}

// HoldsLease reports whether this instance currently executes scheduled fires.
func (s *Scheduler) HoldsLease() bool {
	s.leaderMu.RLock()
	defer s.leaderMu.RUnlock()
	return s.holdsLease
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
		if !s.HoldsLease() {
			return
		}
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
	s.mu.Lock()
	s.entryID[automationID(workspaceID, name)] = entryID
	s.mu.Unlock()
	return nil
}

func (s *Scheduler) Reschedule(a *agentsv1.Automation) error {
	s.Unregister(a.GetWorkspaceId(), a.GetName())
	return s.Register(a)
}

func (s *Scheduler) Unregister(workspaceID, name string) {
	key := automationID(workspaceID, name)
	s.mu.Lock()
	id, ok := s.entryID[key]
	if ok {
		delete(s.entryID, key)
	}
	s.mu.Unlock()
	if ok {
		s.cron.Remove(id)
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
