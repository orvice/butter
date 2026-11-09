package plane

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"butterfly.orx.me/core/log"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// RecordWebhookRejection is a terminal diagnostic receipt, not a work inbox.
// The HTTP adapter calls it only after Deliver has authenticated the body and
// returned a permanent validation error. It never stores raw bodies/headers or
// schedules execution. Valid work must still be durably queued by Deliver.
func (r *Receiver) RecordWebhookRejection(ctx context.Context, app *agentsv1.PlaneApp, body []byte, rejection error) (string, error) {
	if r == nil || r.diagnostics == nil || app.GetWorkspaceId() == "" || app.GetId() == "" {
		return "", ErrRuntimeDependency
	}
	if !errors.Is(rejection, ErrIdentity) && !errors.Is(rejection, ErrMalformed) {
		return "", errors.New("only authenticated permanent validation failures can be recorded")
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte(app.GetWorkspaceId()))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write([]byte(app.GetId()))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write(body)
	key := "webhook:" + hex.EncodeToString(digest.Sum(nil))
	// Advisory only; match the existing decoder and never expose unknown
	// identifier values. The fingerprint can be compared with public event
	// names without storing the body or guessing routing from those names.
	var envelope struct {
		Event string `json:"event"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		envelope.Event = ""
	}
	eventHash := identityHash(envelope.Event)
	now := r.now()
	record := &DiagnosticRecord{
		Key: key, WorkspaceID: app.GetWorkspaceId(), AppID: app.GetId(),
		Class: FailurePermanent, Summary: SafeDiagnosticSummary(fmt.Errorf("webhook event_hash=%s: %w", eventHash, rejection)),
		Recoverable: false, CreatedAt: now, ExpiresAt: now.Add(30 * 24 * time.Hour),
	}
	if err := r.diagnostics.Put(ctx, record); err != nil {
		return "", err
	}
	log.FromContext(ctx).Info("plane webhook rejection recorded", "app_hash", identityHash(app.GetId()), "receipt_hash", identityHash(key), "event_hash", eventHash)
	return key, nil
}
