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
	invRepo invocation.Repository
	runner  Runner
	cfg     Config

	mu      sync.Mutex
	running map[string]*runEntry // invocationID → active run

	onTurnComplete TurnCompleteFunc
}

type runEntry struct {
	cancel          context.CancelFunc
	workspaceID     string
	cancelRequested bool
	finished        bool
}

// New creates a Coordinator. The runner and invRepo must be non-nil.
func New(invRepo invocation.Repository, r Runner, cfg Config) *Coordinator {
	if cfg.MaxRunDuration == 0 {
		cfg.MaxRunDuration = 30 * time.Minute
	}
	return &Coordinator{
		invRepo: invRepo,
		runner:  r,
		cfg:     cfg,
		running: make(map[string]*runEntry),
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
// background execution. It must be called after the invocation is durably
// persisted. The caller passes the pre-resolved agentName and parts.
func (c *Coordinator) Enqueue(inv *agentsv1.Invocation, agentName string, parts []*genai.Part, modelOverride string) {
	invID := inv.GetId()
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.MaxRunDuration)

	c.mu.Lock()
	c.running[invID] = &runEntry{cancel: cancel, workspaceID: inv.GetWorkspaceId()}
	c.mu.Unlock()

	go c.execute(ctx, cancel, inv, agentName, parts, modelOverride)
}

func (c *Coordinator) execute(ctx context.Context, cancel context.CancelFunc, inv *agentsv1.Invocation, agentName string, parts []*genai.Part, modelOverride string) {
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

	// Transition to RUNNING.
	inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING
	inv.StartedAt = timestamppb.Now()
	if err := c.invRepo.Save(context.Background(), inv); err != nil {
		logger.Error("async run: failed to persist RUNNING status", "invocation_id", invID, "err", err)
		c.claimTerminal(invID)
		return
	}

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
		response, runErr = c.runner.RunSSE(runner.WithoutInvocationRecording(ctx), agentName, parts, modelOverride, ctxInfo, nil, nil)
	}

	now := timestamppb.Now()
	if c.claimTerminal(invID) {
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED
		inv.Error = "cancelled by user"
		inv.Output = ""
	} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
		inv.Error = "exceeded maximum run duration"
	} else if runErr != nil {
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
		inv.Error = runErr.Error()
	} else {
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED
		inv.Output = response
	}
	inv.FinishedAt = now
	if inv.GetStartedAt() != nil {
		inv.LatencyMs = now.AsTime().Sub(inv.GetStartedAt().AsTime()).Milliseconds()
	}

	if err := c.invRepo.Save(context.Background(), inv); err != nil {
		logger.Error("async run: failed to persist terminal status",
			"invocation_id", invID, "status", inv.GetStatus().String(), "err", err)
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

func (c *Coordinator) cancelRequested(invocationID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.running[invocationID]
	return ok && entry.cancelRequested
}

// claimTerminal atomically closes the cancellation window. A Cancel that
// returns true always sets cancelRequested before this point; once finished is
// set, later cancellation reports false and cannot overwrite the chosen state.
func (c *Coordinator) claimTerminal(invocationID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.running[invocationID]
	if !ok {
		return false
	}
	entry.finished = true
	return entry.cancelRequested
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

// ReconcileStale marks orphaned QUEUED/RUNNING invocations as FAILED on
// startup. Returns the number of affected records.
func ReconcileStale(ctx context.Context, repo invocation.Repository) (int64, error) {
	return repo.MarkStaleRunning(ctx, "process restarted: invocation orphaned")
}

// Ensure Coordinator satisfies a minimal interface for type checking.
var _ streamorch.Runner = (Runner)(nil)
