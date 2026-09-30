package linear

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"butterfly.orx.me/core/log"

	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// webhookMaxSkew rejects replayed deliveries: Linear stamps each one.
const webhookMaxSkew = time.Minute

var (
	// ErrAppNotFound means no App has that ID.
	ErrAppNotFound = errors.New("unknown linear app")
	// ErrUnauthorized means the signature did not verify.
	ErrUnauthorized = errors.New("invalid linear webhook signature")
	// ErrMalformed means the payload can never be processed.
	ErrMalformed = errors.New("malformed linear webhook")
	// ErrStale means the delivery's timestamp is outside the replay window.
	ErrStale = errors.New("stale linear webhook")
)

// Decision is what happened to one authenticated delivery.
type Decision string

const (
	DecisionAccepted  Decision = "accepted"
	DecisionDuplicate Decision = "duplicate"
	DecisionIgnored   Decision = "ignored"
)

// acceptQueue is the slice of the durable queue the receiver writes to.
type acceptQueue interface {
	Accept(ctx context.Context, event *Event) (string, error)
}

// Receiver authenticates and queues inbound Linear deliveries.
//
// Every delivery re-reads the App and its credentials: an App disabled
// seconds ago must stop accepting, which a configuration cache would not.
type Receiver struct {
	repo    linearrepo.Repository
	keyring *secretbox.Keyring
	queue   acceptQueue
	clock   func() time.Time
}

func NewReceiver(repo linearrepo.Repository, keyring *secretbox.Keyring, queue acceptQueue) *Receiver {
	return &Receiver{repo: repo, keyring: keyring, queue: queue, clock: time.Now}
}

// SetClock overrides the clock. Used by tests.
func (r *Receiver) SetClock(now func() time.Time) { r.clock = now }

// FindApp resolves the App the public route names. It runs before the body
// is read, so an unknown App costs nothing.
func (r *Receiver) FindApp(ctx context.Context, appID string) (*agentsv1.LinearApp, error) {
	app, err := r.repo.FindApp(ctx, appID)
	if errors.Is(err, linearrepo.ErrNotFound) {
		return nil, ErrAppNotFound
	}
	return app, err
}

// Deliver authenticates one raw delivery for app and, when it is an agent
// session event for one of the App's installations, queues it. Any error
// other than ErrUnauthorized, ErrMalformed and ErrStale means the delivery
// could not be accepted right now and Linear should redeliver.
func (r *Receiver) Deliver(ctx context.Context, app *agentsv1.LinearApp, header http.Header, body []byte) (Decision, error) {
	creds, err := r.repo.GetAppCredentials(ctx, app.GetWorkspaceId(), app.GetId())
	if err != nil {
		return "", fmt.Errorf("read linear app credentials: %w", err)
	}
	if !creds.WebhookSecret.Set() {
		// No secret means no authenticated sender is possible.
		return "", ErrUnauthorized
	}
	secret, err := r.keyring.Decrypt(ctx, creds.WebhookSecret.Ciphertext, creds.WebhookSecret.KeyID)
	if err != nil {
		return "", fmt.Errorf("decrypt linear webhook secret: %w", err)
	}
	if !validSignature(secret, header.Get(SignatureHeader), body) {
		return "", ErrUnauthorized
	}

	var payload webhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	now := r.clock()
	if !freshTimestamp(payload.WebhookTimestamp, now) {
		return "", ErrStale
	}

	logger := log.FromContext(ctx).With("app_id", app.GetId(), "workspace_id", app.GetWorkspaceId())
	if payload.Type != "AgentSessionEvent" {
		logger.Debug("ignoring linear webhook", "type", payload.Type)
		return DecisionIgnored, nil
	}
	if !app.GetInboundEnabled() {
		logger.Info("ignoring linear webhook for an app that is not receiving")
		return DecisionIgnored, nil
	}
	if payload.AgentSession.ID == "" {
		return "", fmt.Errorf("%w: agent session event without a session id", ErrMalformed)
	}
	inst, err := r.installationFor(ctx, app, payload.OrganizationID)
	if err != nil {
		if errors.Is(err, linearrepo.ErrNotFound) {
			logger.Warn("ignoring linear webhook for an organization the app is not installed in",
				"organization_id", payload.OrganizationID)
			return DecisionIgnored, nil
		}
		return "", err
	}

	event := &Event{
		WorkspaceID:      app.GetWorkspaceId(),
		AppID:            app.GetId(),
		AppRevision:      app.GetRevision(),
		InstallationID:   inst.GetId(),
		OrganizationID:   inst.GetOrganizationId(),
		AgentSessionID:   payload.AgentSession.ID,
		Action:           payload.Action,
		PromptingUserID:  payload.promptingUserID(),
		PromptContext:    payload.PromptContext,
		DeliveryID:       deliveryID(header, body),
		ReceivedAtUnixMs: now.UnixMilli(),
	}
	if a := payload.AgentActivity; a != nil {
		event.Stop = a.Signal == "stop"
		event.PromptText = strings.TrimSpace(a.Content.Body)
	}
	if issue := payload.AgentSession.Issue; issue != nil {
		event.Issue = Issue{ID: issue.ID, Identifier: issue.Identifier, Title: issue.Title, URL: issue.URL}
	}
	logger.Debug("linear agent session event",
		"action", event.Action, "agent_session_id", event.AgentSessionID, "stop", event.Stop,
		"organization_id", payload.OrganizationID, "prompting_user_id", event.PromptingUserID,
		"delivery_header", header.Get(DeliveryHeader) != "")

	if _, err := r.queue.Accept(ctx, event); err != nil {
		if errors.Is(err, ErrDuplicate) {
			return DecisionDuplicate, nil
		}
		return "", err
	}
	return DecisionAccepted, nil
}

// installationFor resolves the App's installation for the delivery's
// organization. A payload without one falls back to the App's only
// installation; with several there is no honest choice.
func (r *Receiver) installationFor(ctx context.Context, app *agentsv1.LinearApp, organizationID string) (*agentsv1.LinearInstallation, error) {
	if organizationID != "" {
		return r.repo.FindInstallation(ctx, app.GetWorkspaceId(), app.GetId(), organizationID)
	}
	installs, err := r.repo.ListInstallations(ctx, app.GetWorkspaceId(), app.GetId())
	if err != nil {
		return nil, err
	}
	if len(installs) != 1 {
		return nil, fmt.Errorf("delivery names no organization and the app has %d installations: %w", len(installs), linearrepo.ErrNotFound)
	}
	return installs[0], nil
}

// deliveryID is the dedupe key: Linear's delivery ID, or the hash of the
// raw body when the header is missing — a redelivery repeats the body.
func deliveryID(header http.Header, body []byte) string {
	if id := strings.TrimSpace(header.Get(DeliveryHeader)); id != "" {
		return id
	}
	sum := sha256.Sum256(body)
	return "sha256-" + hex.EncodeToString(sum[:])
}

// validSignature checks the hex HMAC-SHA256 of the raw body under the
// webhook signing secret, in constant time.
func validSignature(secret []byte, signature string, body []byte) bool {
	provided, err := hex.DecodeString(strings.TrimSpace(signature))
	if err != nil || len(provided) == 0 {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return hmac.Equal(provided, mac.Sum(nil))
}

func freshTimestamp(unixMs int64, now time.Time) bool {
	if unixMs <= 0 {
		return false
	}
	skew := now.Sub(time.UnixMilli(unixMs))
	return skew <= webhookMaxSkew && skew >= -webhookMaxSkew
}
