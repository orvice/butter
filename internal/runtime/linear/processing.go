package linear

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"butterfly.orx.me/core/log"

	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/repo/linearprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// processingLeaseTTL keeps a second delivery or an operator resend out
	// of a record while one owner works on it; renewed while it runs.
	processingLeaseTTL = 5 * time.Minute
	// preAgentAttempts bounds retries before Agent work starts. Only that
	// window is safely retryable.
	preAgentAttempts       = 3
	defaultPreAgentBackoff = 500 * time.Millisecond
)

// ErrNotResendable means a record has no persisted, undelivered reply.
var ErrNotResendable = errors.New("linear processing record has no reply to resend")

// permanentError marks a pre-Agent failure retrying cannot fix.
type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

func permanent(err error) error {
	if err == nil {
		return nil
	}
	return permanentError{err: err}
}

func isPermanent(err error) bool {
	var p permanentError
	return errors.As(err, &p)
}

// retryPreAgent runs a transient pre-Agent step with bounded backoff.
func (o *Orchestrator) retryPreAgent(ctx context.Context, step func() error) error {
	var lastErr error
	for attempt := 1; attempt <= preAgentAttempts; attempt++ {
		lastErr = step()
		if lastErr == nil || isPermanent(lastErr) {
			return lastErr
		}
		if attempt == preAgentAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt) * o.preAgentBackoff):
		}
	}
	return lastErr
}

// newRecord is the record of a freshly claimed delivery.
func newRecord(event *Event) *agentsv1.LinearProcessingRecord {
	return &agentsv1.LinearProcessingRecord{
		WorkspaceId:     event.WorkspaceID,
		AppId:           event.AppID,
		InstallationId:  event.InstallationID,
		AgentSessionId:  event.AgentSessionID,
		Action:          event.Action,
		DeliveryId:      event.DeliveryID,
		AppRevision:     event.AppRevision,
		PromptingUserId: event.PromptingUserID,
		IssueIdentifier: event.Issue.Identifier,
		Status:          agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_RECEIVED,
	}
}

// claim creates or re-claims the delivery's record and derives the only
// safe action from its persisted state.
func (o *Orchestrator) claim(ctx context.Context, t *turn) (linearprocessing.ClaimAction, error) {
	if o.processing == nil {
		return linearprocessing.ClaimRunAgent, nil
	}
	now := time.Now().UTC()
	lease := uuid.NewString()
	record, action, err := o.processing.Claim(ctx, newRecord(t.event), lease, now, now.Add(processingLeaseTTL))
	if err != nil {
		return linearprocessing.ClaimAcknowledge, err
	}
	t.record = record
	if action != linearprocessing.ClaimAcknowledge {
		t.lease = lease
	}
	return action, nil
}

func (o *Orchestrator) release(ctx context.Context, t *turn) {
	if o.processing == nil || t.record == nil || t.lease == "" {
		return
	}
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := o.processing.ReleaseClaim(releaseCtx, t.record.GetWorkspaceId(), t.record.GetId(), t.lease); err != nil &&
		!errors.Is(err, linearprocessing.ErrLeaseLost) {
		t.logger(ctx).Warn("could not release linear processing claim", "record_id", t.record.GetId(), "err", err)
	}
}

// heartbeat renews the claim while work runs, cancelling the work if the
// claim is lost to another owner.
func (o *Orchestrator) heartbeat(ctx context.Context, t *turn) (context.Context, func()) {
	if o.processing == nil || t.record == nil || t.lease == "" {
		return ctx, func() {}
	}
	leaseCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(processingLeaseTTL / 3)
		defer ticker.Stop()
		for {
			select {
			case <-leaseCtx.Done():
				return
			case <-ticker.C:
				if err := o.processing.RenewClaim(leaseCtx, t.record.GetWorkspaceId(), t.record.GetId(), t.lease,
					time.Now().UTC().Add(processingLeaseTTL)); err != nil {
					log.FromContext(ctx).Error("linear processing claim lost", "record_id", t.record.GetId(), "err", err)
					cancel()
					return
				}
			}
		}
	}()
	var once sync.Once
	return leaseCtx, func() {
		once.Do(func() {
			cancel()
			<-done
		})
	}
}

// save writes the record under the claim.
func (o *Orchestrator) save(ctx context.Context, t *turn) error {
	if o.processing == nil || t.record == nil {
		return nil
	}
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	var stored *agentsv1.LinearProcessingRecord
	var err error
	if t.lease != "" {
		stored, err = o.processing.UpdateClaimed(saveCtx, t.record, t.lease)
	} else {
		stored, err = o.processing.Update(saveCtx, t.record)
	}
	if err != nil {
		return fmt.Errorf("record linear processing state %s: %w", t.record.GetStatus(), err)
	}
	t.record = stored
	return nil
}

func (o *Orchestrator) recordStatus(ctx context.Context, t *turn, status agentsv1.LinearProcessingStatus, errText string) error {
	if o.processing == nil || t.record == nil {
		return nil
	}
	t.record.Status = status
	t.record.Error = errText
	return o.save(ctx, t)
}

// recordUncertain dead-letters a turn whose Agent may have run.
func (o *Orchestrator) recordUncertain(ctx context.Context, t *turn, cause error) error {
	if o.processing == nil || t.record == nil {
		return nil
	}
	t.record.Status = agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN
	t.record.DeadLettered = true
	t.record.Error = sanitizeError(cause)
	if err := o.save(ctx, t); err != nil {
		t.logger(ctx).Error("could not dead-letter linear processing record", "err", err)
		return err
	}
	return nil
}

// persistReply stores the reply before it is posted.
func (o *Orchestrator) persistReply(ctx context.Context, t *turn, a linearapi.Activity) error {
	if o.processing == nil || t.record == nil {
		// Without records the reply lives only in memory until posted.
		t.record = &agentsv1.LinearProcessingRecord{}
	}
	t.record.Output = a.Body
	t.record.OutputType = a.Type
	t.record.OutputSignal = a.Signal
	t.record.OutputSignalMetadata = ""
	if len(a.SignalMetadata) > 0 {
		raw, err := json.Marshal(a.SignalMetadata)
		if err != nil {
			return fmt.Errorf("encode signal metadata: %w", err)
		}
		t.record.OutputSignalMetadata = string(raw)
	}
	t.record.Status = agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_READY_TO_DELIVER
	t.record.Error = ""
	return o.save(ctx, t)
}

// replyActivity rebuilds the persisted reply.
func replyActivity(record *agentsv1.LinearProcessingRecord) linearapi.Activity {
	a := linearapi.Activity{Type: record.GetOutputType(), Body: record.GetOutput(), Signal: record.GetOutputSignal()}
	if a.Type == "" {
		a.Type = linearapi.ActivityResponse
	}
	if raw := record.GetOutputSignalMetadata(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &a.SignalMetadata)
	}
	return a
}

// deliver posts the persisted reply and records the outcome: SUCCEEDED, or
// FAILED with the reply kept for a resend.
func (o *Orchestrator) deliver(ctx context.Context, t *turn) error {
	postErr := t.post(ctx, replyActivity(t.record))
	if o.processing == nil {
		return postErr
	}
	if postErr != nil {
		t.record.Status = agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED
		t.record.Error = sanitizeError(postErr)
	} else {
		t.record.Status = agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED
		t.record.Delivered = true
		t.record.DeadLettered = false
		t.record.Error = ""
	}
	if err := o.save(ctx, t); err != nil {
		t.logger(ctx).Error("could not record linear delivery", "err", err)
	}
	return postErr
}

// answer replies without running the Agent: the reply is persisted, then
// posted, like any other.
func (o *Orchestrator) answer(ctx context.Context, t *turn, activityType, body string) error {
	if err := o.persistReply(ctx, t, linearapi.Activity{Type: activityType, Body: body}); err != nil {
		return err
	}
	_ = o.deliver(ctx, t)
	return nil
}

// Resend posts the persisted reply of a FAILED record again. It never
// invokes the Agent.
func (o *Orchestrator) Resend(ctx context.Context, workspaceID, recordID string) (*agentsv1.LinearProcessingRecord, error) {
	if o.processing == nil {
		return nil, errors.New("linear processing records are not configured")
	}
	now := time.Now().UTC()
	lease := uuid.NewString()
	record, err := o.processing.ClaimForResend(ctx, workspaceID, recordID, lease, now, now.Add(processingLeaseTTL))
	if err != nil {
		return nil, err
	}
	event := &Event{
		WorkspaceID: record.GetWorkspaceId(), AppID: record.GetAppId(), InstallationID: record.GetInstallationId(),
		AgentSessionID: record.GetAgentSessionId(), DeliveryID: record.GetDeliveryId(),
	}
	t := &turn{o: o, event: event, record: record, lease: lease}
	defer o.release(ctx, t)
	if record.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED || record.GetOutput() == "" || record.GetDelivered() {
		return record, ErrNotResendable
	}
	token, err := o.tokens.AccessToken(ctx, record.GetWorkspaceId(), record.GetInstallationId())
	if err != nil {
		return record, err
	}
	t.token = token
	if err := o.deliver(ctx, t); err != nil {
		return t.record, err
	}
	return t.record, nil
}
