package plane

import (
	"context"
	"errors"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
	"strings"
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
type InvocationRecovery interface {
	FailActive(context.Context, string, string, string, string) error
}
type Processor struct {
	invocations InvocationRecovery
	runner      Runner
	publisher   ReplyPublisher
	progress    ProgressSink
	diagnostics DiagnosticStore
	maxReply    int
	claimRepo   planeprocessing.ClaimRepository
	processing  planeprocessing.Repository
	claimOwner  string
	control     RunControl
	controlPoll time.Duration
	batches     *redisBatchSource
}

func NewProcessor(r Runner, p ReplyPublisher) *Processor {
	return &Processor{runner: r, publisher: p, control: NewMemoryRunControl(), maxReply: 500}
}
func (p *Processor) SetProgressSink(s ProgressSink)           { p.progress = s }
func (p *Processor) SetDiagnosticStore(store DiagnosticStore) { p.diagnostics = store }
func (p *Processor) SetRunControl(c RunControl)               { p.control = c }
func (p *Processor) SetClaimRepository(r planeprocessing.ClaimRepository, owner string) {
	p.claimRepo = r
	p.claimOwner = owner
}
func (p *Processor) SetProcessingRepository(r planeprocessing.Repository) { p.processing = r }
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
	if p.control == nil {
		return "", ErrRuntimeDependency
	}
	if p.batches != nil {
		if err := p.batches.prepare(ctx, e); err != nil {
			return "", err
		}
		if batching, ok := p.runner.(interface{ SupportsPromptBatching(string, string) bool }); ok && !batching.SupportsPromptBatching(e.WorkspaceID, e.AgentID) {
			// A Workflow answer addresses the oldest Interrupt exactly once.
			// Keep each reply intact rather than joining several answers.
			e.Members = e.Members[:1]
		}
		// Re-check mutable policy against every frozen member. Stop and
		// terminal recovery remain available after inbound is disabled.
		for i, member := range e.Members {
			if !app.GetInboundEnabled() || member.AgentID == "" || member.AgentID != app.GetAgentId() ||
				!allowedStopUser(app.GetAllowedUserIds(), member.ActorUserID) || strings.EqualFold(member.ActorUserID, member.BotUserID) {
				if i == 0 {
					return "", ErrPromptRejected
				}
				e.Members = e.Members[:i]
				break
			}
		}
	}
	return p.processControlled(ctx, app, e)
}
func (p *Processor) Stop(ctx context.Context, e *Event) (StopAudit, error) {
	if p == nil || p.control == nil {
		return StopAudit{}, ErrRuntimeDependency
	}
	token, err := p.control.RequestStop(ctx, e)
	state := RunStateStopped
	if token != "" {
		state = RunStateStopping
	}
	return StopAudit{Key: planeRunKey(e), State: state, Attempts: 1}, err
}

func planeRunKey(e *Event) string {
	// A Stop has its own Activity ID; the shared Run identity locates the
	// current Prompt. Keep the workspace/App/installation boundaries intact.
	return e.WorkspaceID + ":" + e.AppID + ":" + e.InstallationID + ":" + e.AgentRunID
}
