package plane

import (
	"context"
	"errors"
	"fmt"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
	"strings"
	"sync"
	"time"
)

var ErrPromptRequired = errors.New("plane activity prompt is required")
var ErrRunnerUnavailable = errors.New("plane runner is unavailable")
var ErrReplyUnavailable = errors.New("plane reply publisher is unavailable")

type Runner interface {
	Run(context.Context, string, []*genai.Part, string, *agentsv1.ContextInfo, runner.EventCallback, runner.CompactionCallback) (string, error)
}
type ReplyPublisher interface {
	PublishReply(context.Context, *Event, string) error
}
type RunState string

const (
	RunStateRunning   RunState = "running"
	RunStateStopping  RunState = "stopping"
	RunStateStopped   RunState = "stopped"
	RunStateSucceeded RunState = "succeeded"
	RunStateFailed    RunState = "failed"
)

type ProgressEvent struct {
	Key                    string
	Sequence               int64
	State                  RunState
	Error                  string
	WorkflowID, DeliveryID string
}
type ProgressSink interface {
	EmitProgress(context.Context, *Event, ProgressEvent) error
}
type StopAudit struct {
	Key      string
	State    RunState
	Attempts int
}
type Processor struct {
	runner      Runner
	publisher   ReplyPublisher
	progress    ProgressSink
	diagnostics DiagnosticStore
	mu          sync.Mutex
	active      map[string]struct{}
	cancels     map[string]context.CancelFunc
	states      map[string]RunState
	seen        map[string]struct{}
	maxReply    int
	claimRepo   planeprocessing.ClaimRepository
	claimOwner  string
}

func NewProcessor(r Runner, p ReplyPublisher) *Processor {
	return &Processor{runner: r, publisher: p, active: map[string]struct{}{}, cancels: map[string]context.CancelFunc{}, states: map[string]RunState{}, seen: map[string]struct{}{}, maxReply: 500}
}
func (p *Processor) SetProgressSink(s ProgressSink)           { p.progress = s }
func (p *Processor) SetDiagnosticStore(store DiagnosticStore) { p.diagnostics = store }
func (p *Processor) SetClaimRepository(r planeprocessing.ClaimRepository, owner string) {
	p.claimRepo = r
	p.claimOwner = owner
}
func NewProcessorWithDurableDelivery(r Runner, target ReplyPublisher, store DeliveryStore) *Processor {
	return NewProcessor(r, NewDurablePublisher(target, store))
}

func (p *Processor) recordDiagnostic(ctx context.Context, e *Event, err error) {
	if p.diagnostics != nil {
		_ = p.diagnostics.Put(ctx, BuildDiagnostic(e, "", err))
	}
}
func (p *Processor) emit(ctx context.Context, e *Event, k string, n int64, s RunState, err string) {
	if p.progress != nil {
		_ = p.progress.EmitProgress(ctx, e, ProgressEvent{Key: k, Sequence: n, State: s, Error: err})
	}
}
func (p *Processor) Process(ctx context.Context, app *agentsv1.PlaneApp, e *Event) (string, error) {
	if p == nil || p.runner == nil {
		return "", ErrRunnerUnavailable
	}
	if p.publisher == nil {
		return "", ErrReplyUnavailable
	}
	if app == nil || app.GetAgentId() == "" {
		return "", errors.New("plane app agent is not configured")
	}
	prompt := strings.TrimSpace(e.Prompt)
	if prompt == "" {
		return "", ErrPromptRequired
	}
	key := e.InstallationID + ":" + e.AgentRunID + ":" + e.ActivityID
	claimKey := e.WorkspaceID + ":" + e.AppID + ":" + e.InstallationID + ":" + e.AgentRunID
	p.mu.Lock()
	if _, ok := p.seen[key]; ok {
		p.mu.Unlock()
		return "", ErrDuplicate
	}
	if _, ok := p.active[key]; ok {
		p.mu.Unlock()
		return "", ErrDuplicate
	}
	var claim planeprocessing.Claim
	claimed := false
	if p.claimRepo != nil {
		if p.claimOwner == "" {
			return "", planeprocessing.ErrClaimConflict
		}
		var claimErr error
		claim, claimErr = p.claimRepo.Claim(ctx, e.WorkspaceID, e.AppID, e.InstallationID, e.AgentRunID, p.claimOwner, time.Now().Add(5*time.Minute))
		if claimErr != nil {
			return "", claimErr
		}
		claimed = true
	}
	p.active[key] = struct{}{}
	child, cancel := context.WithCancel(ctx)
	p.cancels[key] = cancel
	p.states[key] = RunStateRunning
	p.mu.Unlock()
	defer func() { cancel(); p.mu.Lock(); delete(p.cancels, key); delete(p.active, key); p.mu.Unlock() }()
	out, err := p.runner.Run(child, app.GetAgentId(), []*genai.Part{{Text: prompt}}, "", &agentsv1.ContextInfo{UserId: e.BotUserID, SessionId: e.AgentRunID, ChannelName: "plane"}, nil, nil)
	if claimed {
		if completeErr := p.claimRepo.Complete(context.Background(), claimKey, claim); completeErr != nil {
			return "", errors.New("plane processing result was discarded after claim completion failure")
		}
	}
	p.mu.Lock()
	stopWon := p.states[key] == RunStateStopping
	if stopWon {
		p.states[key] = RunStateStopped
	}
	p.mu.Unlock()
	if stopWon {
		p.emit(ctx, e, key, 3, RunStateStopped, "stopped before completion")
		return "", context.Canceled
	}
	p.emit(ctx, e, key, 1, RunStateRunning, "")
	if err != nil {
		state := RunStateFailed
		if errors.Is(err, context.Canceled) {
			state = RunStateStopped
		}
		p.mu.Lock()
		p.states[key] = state
		p.mu.Unlock()
		seq := int64(2)
		if state == RunStateStopped {
			seq = 3
		}
		p.emit(ctx, e, key, seq, state, err.Error())
		p.recordDiagnostic(ctx, e, err)
		return "", fmt.Errorf("plane runner: %w", err)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		p.mu.Lock()
		p.states[key] = RunStateFailed
		p.mu.Unlock()
		p.emit(ctx, e, key, 2, RunStateFailed, "empty reply")
		p.recordDiagnostic(ctx, e, errors.New("empty reply"))
		return "", errors.New("plane runner returned empty reply")
	}
	p.mu.Lock()
	if p.states[key] == RunStateStopping {
		p.states[key] = RunStateStopped
		p.mu.Unlock()
		p.emit(ctx, e, key, 3, RunStateStopped, "stopped before reply")
		return "", context.Canceled
	}
	// Reserve the successful execution outcome before the external reply
	// write. A later Stop can no longer rewrite a turn already completed.
	p.states[key] = RunStateSucceeded
	p.mu.Unlock()
	if len(out) > p.maxReply {
		out = out[:p.maxReply]
	}
	if err = p.publisher.PublishReply(ctx, e, out); err != nil {
		p.mu.Lock()
		p.states[key] = RunStateFailed
		p.mu.Unlock()
		p.emit(ctx, e, key, 2, RunStateFailed, err.Error())
		p.recordDiagnostic(ctx, e, err)
		return "", fmt.Errorf("plane reply: %w", err)
	}
	p.mu.Lock()
	p.seen[key] = struct{}{}
	p.mu.Unlock()
	p.emit(ctx, e, key, 2, RunStateSucceeded, "")
	return out, nil
}
func (p *Processor) Stop(ctx context.Context, e *Event) (StopAudit, error) {
	key := e.InstallationID + ":" + e.AgentRunID + ":" + e.ActivityID
	p.mu.Lock()
	defer p.mu.Unlock()
	if s, ok := p.states[key]; ok && s != RunStateRunning && s != RunStateStopping {
		return StopAudit{Key: key, State: s, Attempts: 1}, nil
	}
	if c, ok := p.cancels[key]; ok {
		p.states[key] = RunStateStopping
		c()
		p.emit(ctx, e, key, 2, RunStateStopping, "")
		return StopAudit{Key: key, State: RunStateStopping, Attempts: 1}, nil
	}
	s, ok := p.states[key]
	if !ok {
		s = RunStateStopped
	}
	return StopAudit{Key: key, State: s, Attempts: 1}, nil
}
