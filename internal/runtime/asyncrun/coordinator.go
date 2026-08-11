// Package asyncrun manages background Agent execution for dashboard async
// chat. The Coordinator accepts submitted invocations, transitions them
// through QUEUED→RUNNING→terminal, and exposes cancellation — all in-process
// and independently of any observing browser connection.
package asyncrun

import (
	"context"
	"errors"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/inputpart"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/runtime/streamorch"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Runner is the subset of runner.Service needed for background execution.
type Runner interface {
	RunSSE(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error)
	ResolveAgentRef(workspaceID, agentID string) (string, bool)
	GetAgentIdentity(name string) (agentID, displayName string, ok bool)
	CancelInvocation(id, workspaceID string) bool
}

// TurnCompleteFunc is called after a successful async turn completes. The
// coordinator uses this to trigger title generation without blocking the
// invocation.
type TurnCompleteFunc func(ctx context.Context, inv *agentsv1.Invocation)

// Config holds coordinator-level settings.
type Config struct {
	MaxRunDuration time.Duration
}

// Coordinator manages the lifecycle of async dashboard invocations.
type Coordinator struct {
	invRepo       invocation.Repository
	inputPartRepo inputpart.Repository
	runner        Runner
	cfg           Config

	mu           sync.Mutex
	running      map[string]*runEntry // invocationID → active run
	shuttingDown bool
	wg           sync.WaitGroup

	hub *watchHub

	onTurnComplete TurnCompleteFunc
}

type runEntry struct {
	cancel      context.CancelFunc
	workspaceID string
	// cancelRequested marks an explicit user Stop; the run ends CANCELLED.
	cancelRequested bool
	// shutdownRequested marks a graceful process shutdown; the run ends
	// FAILED with an honest operational reason, never CANCELLED, so user
	// intent stays distinguishable from a system interruption.
	shutdownRequested bool
	finished          bool
}

// New creates a Coordinator. The runner and invRepo must be non-nil.
func New(invRepo invocation.Repository, inputPartRepo inputpart.Repository, r Runner, cfg Config) *Coordinator {
	if cfg.MaxRunDuration == 0 {
		cfg.MaxRunDuration = 30 * time.Minute
	}
	return &Coordinator{
		invRepo:       invRepo,
		inputPartRepo: inputPartRepo,
		runner:        r,
		cfg:           cfg,
		running:       make(map[string]*runEntry),
		hub:           newWatchHub(),
	}
}

// SetTurnCompleteCallback sets the function called after a successful turn.
func (c *Coordinator) SetTurnCompleteCallback(fn TurnCompleteFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onTurnComplete = fn
}

func (c *Coordinator) getTurnComplete() TurnCompleteFunc {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.onTurnComplete
}

// Enqueue transitions the invocation from QUEUED to RUNNING and starts
// background execution. It must be called after both the invocation and its
// Input Parts are durably persisted. Parts are loaded from the InputPart
// repository at execution time so they survive process restarts between
// accept and run.
func (c *Coordinator) Enqueue(inv *agentsv1.Invocation, agentName string, modelOverride string) {
	invID := inv.GetId()
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.MaxRunDuration)

	c.mu.Lock()
	if c.shuttingDown {
		c.mu.Unlock()
		cancel()
		c.failInvocation(inv, shutdownFailureReason)
		return
	}
	c.running[invID] = &runEntry{cancel: cancel, workspaceID: inv.GetWorkspaceId()}
	c.wg.Add(1)
	c.mu.Unlock()

	go func() {
		defer c.wg.Done()
		c.execute(ctx, cancel, inv, agentName, modelOverride)
	}()
}

func (c *Coordinator) execute(ctx context.Context, cancel context.CancelFunc, inv *agentsv1.Invocation, agentName string, modelOverride string) {
	invID := inv.GetId()
	defer func() {
		cancel()
		c.mu.Lock()
		delete(c.running, invID)
		c.mu.Unlock()
	}()

	logger := log.FromContext(ctx)
	logger.Info("async run starting",
		"invocation_id", invID,
		"agent", agentName,
		"session_id", inv.GetSessionId(),
		"workspace_id", inv.GetWorkspaceId(),
	)

	// Load persisted input parts.
	parts, loadErr := c.loadParts(ctx, invID)
	if loadErr != nil {
		logger.Error("async run: failed to load input parts", "invocation_id", invID, "err", loadErr)
		c.failInvocation(inv, "failed to load input parts: "+loadErr.Error())
		return
	}

	// Transition to RUNNING.
	inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING
	inv.StartedAt = timestamppb.Now()
	if err := c.invRepo.Save(context.Background(), inv); err != nil {
		logger.Error("async run: failed to persist RUNNING status", "invocation_id", invID, "err", err)
		c.claimTerminal(invID)
		// Observers must still see one terminal state frame — otherwise
		// close-without-terminal reads as observer lag and clients would
		// re-attach to a dead invocation.
		c.failInvocation(inv, "failed to persist RUNNING status: "+err.Error())
		return
	}
	c.publishState(inv)

	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        invID,
		ChannelName: inv.GetAppName(),
		UserId:      inv.GetUserId(),
		SessionId:   inv.GetSessionId(),
		WorkspaceId: inv.GetWorkspaceId(),
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
	}

	var response string
	var runErr error
	if c.cancelRequested(invID) {
		runErr = context.Canceled
	} else {
		// Drive the run through the shared streaming orchestration so
		// observers attached via Watch receive the same frame sequence
		// StreamAgent emits.
		sink := &hubSink{hub: c.hub, invocationID: invID}
		runErr = streamorch.Run(runner.WithoutInvocationRecording(ctx), c.runner, streamorch.AgentRef{Name: agentName, ID: inv.GetAgentId()}, parts, modelOverride, ctxInfo, sink)
		response = sink.response
	}

	now := timestamppb.Now()
	claim := c.claimTerminal(invID)
	switch {
	case claim.cancelled:
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED
		inv.Error = "cancelled by user"
		inv.Output = ""
	case runErr == nil:
		// The runner finished cleanly, so the output is complete — record the
		// honest SUCCEEDED even when a shutdown or deadline signal raced the
		// finish line.
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED
		inv.Output = response
	case claim.shutdown:
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
		inv.Error = shutdownFailureReason
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
		inv.Error = deadlineFailureReason(c.cfg.MaxRunDuration)
	default:
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
		inv.Error = runErr.Error()
	}
	inv.FinishedAt = now
	if inv.GetStartedAt() != nil {
		inv.LatencyMs = now.AsTime().Sub(inv.GetStartedAt().AsTime()).Milliseconds()
	}

	if err := c.invRepo.Save(context.Background(), inv); err != nil {
		logger.Error("async run: failed to persist terminal status",
			"invocation_id", invID, "status", inv.GetStatus().String(), "err", err)
	}

	// Publish the single terminal state frame, then close every observer so
	// watchers see terminal-then-close in order.
	c.publishState(inv)
	c.hub.closeAll(invID)

	// Clean up temporary input parts after successful session event commit.
	if inv.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		if delErr := c.inputPartRepo.Delete(context.Background(), invID); delErr != nil {
			logger.Warn("async run: failed to clean up input parts", "invocation_id", invID, "err", delErr)
		}
	}

	logger.Info("async run finished",
		"invocation_id", invID,
		"status", inv.GetStatus().String(),
		"latency_ms", inv.GetLatencyMs(),
	)

	if inv.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		if fn := c.getTurnComplete(); fn != nil {
			fn(context.Background(), inv)
		}
	}
}

// loadParts reads persisted InputPart protos and converts them to genai.Part
// for the runner.
func (c *Coordinator) loadParts(ctx context.Context, invocationID string) ([]*genai.Part, error) {
	stored, err := c.inputPartRepo.Load(ctx, invocationID)
	if err != nil {
		return nil, err
	}
	parts := make([]*genai.Part, 0, len(stored))
	for _, sp := range stored {
		switch p := sp.GetPart().(type) {
		case *agentsv1.InputPart_Text:
			parts = append(parts, genai.NewPartFromText(p.Text))
		case *agentsv1.InputPart_InlineData:
			parts = append(parts, genai.NewPartFromBytes(p.InlineData.GetData(), p.InlineData.GetMimeType()))
		}
	}
	return parts, nil
}

// failInvocation persists a FAILED status without entering RUNNING, and
// delivers the terminal state frame to any attached observers so a run that
// never starts still ends their streams with terminal-then-close.
func (c *Coordinator) failInvocation(inv *agentsv1.Invocation, reason string) {
	inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
	inv.Error = reason
	inv.FinishedAt = timestamppb.Now()
	_ = c.invRepo.Save(context.Background(), inv)
	c.publishState(inv)
	c.hub.closeAll(inv.GetId())
}

func (c *Coordinator) cancelRequested(invocationID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.running[invocationID]
	return ok && entry.cancelRequested
}

// terminalClaim reports why a run's context was interrupted, if it was.
// Explicit user cancellation takes precedence over shutdown so a Stop
// pressed just before shutdown still records user intent.
type terminalClaim struct {
	cancelled bool
	shutdown  bool
}

// claimTerminal atomically closes the cancellation window. A Cancel that
// returns true always sets cancelRequested before this point; once finished is
// set, later cancellation reports false and cannot overwrite the chosen state.
func (c *Coordinator) claimTerminal(invocationID string) terminalClaim {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.running[invocationID]
	if !ok {
		return terminalClaim{}
	}
	entry.finished = true
	return terminalClaim{
		cancelled: entry.cancelRequested,
		shutdown:  entry.shutdownRequested && !entry.cancelRequested,
	}
}

// Cancel stops a running invocation. Returns true if found and cancelled.
func (c *Coordinator) Cancel(invocationID, workspaceID string) bool {
	c.mu.Lock()
	entry, ok := c.running[invocationID]
	if !ok || entry.finished {
		c.mu.Unlock()
		return false
	}
	if workspaceID != "" && entry.workspaceID != workspaceID {
		c.mu.Unlock()
		return false
	}
	entry.cancelRequested = true
	entry.cancel()
	c.mu.Unlock()
	// Also propagate to the runner's cancel registry for context guard.
	c.runner.CancelInvocation(invocationID, workspaceID)
	return true
}

// noReplaySuffix closes every operational failure reason: work is never
// replayed automatically, and the user's recovery path is an explicit
// resubmission. Reasons are surfaced verbatim in the dashboard beside the
// submitted turn.
const noReplaySuffix = "; no work was replayed automatically. Review your message and resubmit to retry."

// shutdownFailureReason is the honest terminal error recorded when a graceful
// process shutdown interrupts an async run (or rejects one submitted while
// shutting down).
const shutdownFailureReason = "interrupted by a service shutdown" + noReplaySuffix

// deadlineFailureReason is the honest terminal error recorded when a run
// exceeds the configured maximum duration.
func deadlineFailureReason(max time.Duration) string {
	return "exceeded the maximum run duration (" + max.String() + ") and was stopped" + noReplaySuffix
}

// Shutdown stops all process-owned async work for a graceful process exit.
// Every in-flight run is cancelled and persists FAILED with an honest
// shutdown reason; Shutdown blocks until those terminal writes complete or
// ctx expires. New Enqueue calls after Shutdown fail immediately.
func (c *Coordinator) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	c.shuttingDown = true
	for _, entry := range c.running {
		if !entry.finished {
			entry.shutdownRequested = true
			entry.cancel()
		}
	}
	c.mu.Unlock()

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsActive reports whether the given invocation is tracked as in-flight.
func (c *Coordinator) IsActive(invocationID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.running[invocationID]
	return ok && !entry.finished
}

// NewInvocationID generates a v7 UUID for a new invocation.
func NewInvocationID() string {
	if id, err := uuid.NewV7(); err == nil {
		return id.String()
	}
	return uuid.NewString()
}

// BuildContextInfo creates a ContextInfo for an async invocation, reusing the
// shared streamorch validation but in a minimal form.
func BuildContextInfo(workspaceID, userID, sessionID, invocationID string) *agentsv1.ContextInfo {
	return &agentsv1.ContextInfo{
		Uuid:        invocationID,
		ChannelName: "web-chat",
		UserId:      userID,
		SessionId:   sessionID,
		WorkspaceId: workspaceID,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
	}
}

// staleFailureReason is the honest terminal error recorded for QUEUED/RUNNING
// invocations orphaned by a previous process exit. Startup reconciliation
// only marks records — it never re-invokes the Agent or repeats tool side
// effects.
const staleFailureReason = "interrupted by a service restart before it could finish" + noReplaySuffix

// ReconcileStale marks orphaned QUEUED/RUNNING invocations as FAILED on
// startup. Returns the number of affected records.
func ReconcileStale(ctx context.Context, repo invocation.Repository) (int64, error) {
	return repo.MarkStaleRunning(ctx, staleFailureReason)
}

// Ensure Coordinator satisfies a minimal interface for type checking.
var _ streamorch.Runner = (Runner)(nil)
