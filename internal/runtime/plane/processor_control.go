package plane

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
)

// processControlled keeps execution's lease through delivery and terminal
// progress. Stop and Finish arbitrate on the shared RunControl before any
// terminal writes; the Stop observer confirms only after End.
func (p *Processor) processControlled(ctx context.Context, app *agentsv1.PlaneApp, e *Event) (out string, resultErr error) {
	if p.runner == nil {
		return "", ErrRunnerUnavailable
	}
	if p.publisher == nil {
		return "", ErrReplyUnavailable
	}
	if app == nil || app.GetAgentId() == "" {
		return "", errors.New("plane app agent is not configured")
	}
	if strings.TrimSpace(e.Prompt) == "" {
		return "", ErrPromptRequired
	}
	agentName := app.GetAgentId()
	if resolver, ok := p.runner.(interface {
		ResolveAgentRef(string, string) (string, bool)
	}); ok {
		var found bool
		agentName, found = resolver.ResolveAgentRef(e.WorkspaceID, app.GetAgentId())
		if !found {
			return "", errors.New("plane agent is not available in this workspace")
		}
	}
	var claim planeprocessing.Claim
	if p.claimRepo != nil {
		if p.claimOwner == "" {
			return "", planeprocessing.ErrClaimConflict
		}
		var err error
		claim, err = p.claimRepo.Claim(ctx, e.WorkspaceID, e.AppID, e.InstallationID, e.AgentRunID, p.claimOwner, time.Now().Add(controlTTL))
		if err != nil {
			return "", err
		}
		defer func() {
			finishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := p.claimRepo.Complete(finishCtx, planeRunKey(e), claim); err != nil {
				resultErr = errors.Join(resultErr, err)
			}
		}()
	}
	token := uuid.NewString()
	if err := p.control.Begin(ctx, e, token); err != nil {
		if errors.Is(err, ErrTurnConsumed) {
			return "", errors.Join(ErrDuplicate, err)
		}
		return "", err
	}
	// Once Begin succeeds this event must never run the model again, even
	// when the Worker loses its delivery or the Pod dies during the turn.
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(ErrTurnConsumed, resultErr)
		}
	}()
	runCtx, cancelRun := context.WithCancelCause(ctx)
	defer cancelRun(nil)
	keepCtx, stopKeeping := context.WithCancel(context.WithoutCancel(ctx))
	kept := make(chan struct{})
	var leaseLost atomic.Bool
	go func() {
		defer close(kept)
		p.keepControlled(keepCtx, e, token, claim, &leaseLost, cancelRun)
	}()
	defer func() { stopKeeping(); <-kept }()

	// Runner-owned Invocations must use the same Stop/Finish arbitration
	// before their terminal Save. Cache it for runners without a recorder
	// and failures before an Invocation was created.
	var finishOnce sync.Once
	var state RunState
	var finishErr error
	finish := func(ctx context.Context, proposed RunState) (RunState, error) {
		finishOnce.Do(func() {
			if leaseLost.Load() {
				finishErr = ErrRunLeaseLost
				return
			}
			if p.claimRepo != nil {
				if _, err := p.claimRepo.Renew(ctx, planeRunKey(e), claim, time.Now().Add(controlTTL)); err != nil {
					finishErr = errors.Join(ErrRunLeaseLost, err)
					return
				}
			}
			state, finishErr = p.control.Finish(ctx, e, token, proposed)
		})
		return state, finishErr
	}
	runCtx = runner.WithInvocationFinalizer(runCtx, func(ctx context.Context, inv *agentsv1.Invocation) error {
		proposed := RunStateSucceeded
		if inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED || runCtx.Err() != nil || strings.TrimSpace(inv.GetOutput()) == "" {
			proposed = RunStateFailed
		}
		result, err := finish(ctx, proposed)
		if err != nil {
			return err
		}
		switch result {
		case RunStateStopped:
			inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED
			inv.Error = "stopped by Plane"
			inv.Output = ""
		case RunStateFailed:
			inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
			if inv.GetError() == "" {
				inv.Error = "plane runner returned empty reply or execution was interrupted"
			}
		}
		return nil
	})
	out, runErr := p.runner.Run(runCtx, agentName, []*genai.Part{{Text: strings.TrimSpace(e.Prompt)}}, "",
		&agentsv1.ContextInfo{WorkspaceId: e.WorkspaceID, UserId: e.BotUserID, SessionId: e.AgentRunID, ChannelName: "plane"}, nil, nil)
	proposed := RunStateSucceeded
	out = strings.TrimSpace(out)
	if runErr != nil || runCtx.Err() != nil || out == "" {
		proposed = RunStateFailed
	}
	finishCtx, finishCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finishCancel()
	if errors.Is(runErr, runner.ErrInvocationFinalization) {
		return "", runErr
	}
	// Lease loss suppresses all subsequent reply/progress/diagnostic writes.
	if leaseLost.Load() {
		return "", ErrRunLeaseLost
	}
	state, err := finish(finishCtx, proposed)
	if err != nil {
		return "", err
	}
	if leaseLost.Load() {
		return "", ErrRunLeaseLost
	}
	defer func() {
		if leaseLost.Load() {
			return
		}
		endCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := p.control.End(endCtx, e, token, state); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}()
	key := planeRunKey(e) + ":" + e.ActivityID
	if state == RunStateStopped {
		p.emit(finishCtx, e, key, 3, RunStateStopped, "")
		return "", context.Canceled
	}
	p.emit(finishCtx, e, key, 1, RunStateRunning, "")
	if state != RunStateSucceeded {
		if runErr == nil {
			runErr = context.Cause(runCtx)
		}
		if runErr == nil {
			runErr = errors.New("plane runner returned empty reply")
		}
		p.emit(finishCtx, e, key, 2, state, runErr.Error())
		p.recordDiagnostic(finishCtx, e, runErr)
		return "", runErr
	}
	if len(out) > p.maxReply {
		out = out[:p.maxReply]
	}
	deliveryCtx, deliveryCancel := context.WithTimeout(runCtx, 30*time.Second)
	defer deliveryCancel()
	if err := p.publisher.PublishReply(deliveryCtx, e, out); err != nil {
		// The model succeeded; a failed delivery cannot rewrite execution
		// into FAILED or cause a later Stop to win over its completed turn.
		if !leaseLost.Load() {
			diagnosticCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			p.recordDiagnostic(diagnosticCtx, e, err)
			done()
		}
		return "", fmt.Errorf("plane reply: %w", err)
	}
	if leaseLost.Load() {
		return "", ErrRunLeaseLost
	}
	progressCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer done()
	p.emit(progressCtx, e, key, 2, RunStateSucceeded, "")
	return out, nil
}

func (p *Processor) keepControlled(ctx context.Context, e *Event, token string, claim planeprocessing.Claim, lost *atomic.Bool, cancel context.CancelCauseFunc) {
	poll := p.controlPoll
	if poll <= 0 {
		poll = 250 * time.Millisecond
	}
	tick := time.NewTicker(poll)
	defer tick.Stop()
	lastClaimRenewal := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
		attempt, done := context.WithTimeout(ctx, 2*time.Second)
		stopped, err := p.control.Renew(attempt, e, token)
		if err == nil && p.claimRepo != nil && time.Since(lastClaimRenewal) >= controlTTL/3 {
			_, err = p.claimRepo.Renew(attempt, planeRunKey(e), claim, time.Now().Add(controlTTL))
			if err == nil {
				lastClaimRenewal = time.Now()
			}
		}
		done()
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			lost.Store(true)
			cancel(errors.Join(ErrRunLeaseLost, err))
			return
		}
		if stopped {
			cancel(context.Canceled)
			// Keep renewing while cancellation converges. Never release the
			// holder while Pi/Cursor is still taking its AbortSession path.
		}
	}
}
