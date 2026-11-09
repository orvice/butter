package plane

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
	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	planeinstallation "go.orx.me/apps/butter/internal/repo/planeinstallation"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/userinput"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	SignatureHeader = "X-Plane-Signature"
	DeliveryHeader  = "X-Plane-Delivery"
	EventHeader     = "X-Plane-Event"
	MaxWebhookBytes = 1 << 20
)

var (
	ErrAppNotFound  = errors.New("unknown plane app")
	ErrUnauthorized = errors.New("invalid plane webhook signature")
	ErrMalformed    = errors.New("malformed plane webhook")
	ErrIdentity     = errors.New("plane webhook identity mismatch")
	ErrDuplicate    = errors.New("plane webhook already accepted")
)

// Plane creates the Run. This integration reads Runs and reads/writes their
// Activities; creating/updating Runs is neither used nor required.
var requiredAgentScopes = []string{"agents.runs:read", "agents.run_activities:read", "agents.run_activities:write"}

type Decision string

const (
	DecisionAccepted  Decision = "accepted"
	DecisionDuplicate Decision = "duplicate"
	DecisionIgnored   Decision = "ignored"
	// Rejected acknowledges a durably recorded permanent failure, not work.
	DecisionRejected Decision = "rejected"
)

type Queue interface {
	Accept(context.Context, *Event) (string, error)
}
type Receiver struct {
	apps          planerepo.Repository
	installations planeinstallation.Repository
	diagnostics   DiagnosticStore
	keyring       *secretbox.Keyring
	queue         Queue
	now           func() time.Time
}

func NewReceiver(apps planerepo.Repository, installations planeinstallation.Repository, keyring *secretbox.Keyring, queue Queue) *Receiver {
	return &Receiver{apps: apps, installations: installations, keyring: keyring, queue: queue, now: time.Now}
}

func (r *Receiver) SetDiagnosticStore(store DiagnosticStore) { r.diagnostics = store }

type appLookup interface {
	FindAppByID(context.Context, string) (*agentsv1.PlaneApp, error)
}

func (r *Receiver) FindApp(ctx context.Context, id string) (*agentsv1.PlaneApp, error) {
	lookup, ok := r.apps.(appLookup)
	if !ok {
		return nil, errors.New("plane app lookup is not configured")
	}
	app, err := lookup.FindAppByID(ctx, id)
	if errors.Is(err, planerepo.ErrNotFound) {
		return nil, ErrAppNotFound
	}
	return app, err
}
func (r *Receiver) FindScopedApp(ctx context.Context, workspaceID, appID, installationID string) (*agentsv1.PlaneApp, error) {
	if workspaceID == "" || appID == "" || installationID == "" {
		return nil, errors.New("plane app scope is incomplete")
	}
	lookup, ok := r.apps.(interface {
		GetApp(context.Context, string, string) (*agentsv1.PlaneApp, error)
	})
	if !ok {
		return nil, errors.New("plane app lookup is not configured")
	}
	app, err := lookup.GetApp(ctx, workspaceID, appID)
	if errors.Is(err, planerepo.ErrNotFound) {
		return nil, ErrAppNotFound
	}
	if err != nil {
		return nil, err
	}
	if app.GetWorkspaceId() != workspaceID {
		return nil, ErrAppNotFound
	}
	return app, nil
}
func (r *Receiver) Deliver(ctx context.Context, app *agentsv1.PlaneApp, headers http.Header, body []byte) (decision Decision, resultErr error) {
	if r == nil || r.apps == nil || r.installations == nil || r.keyring == nil || r.queue == nil {
		return "", errors.New("plane receive is not configured")
	}
	creds, err := r.apps.GetAppCredentials(ctx, app.GetWorkspaceId(), app.GetId())
	if err != nil {
		return "", err
	}
	if !creds.WebhookSecret.Set() {
		return "", ErrUnauthorized
	}
	secret, err := r.keyring.Decrypt(ctx, creds.WebhookSecret.Ciphertext, creds.WebhookSecret.KeyID)
	if err != nil {
		return "", err
	}
	defer clear(secret)
	if !validSignature(secret, headers.Get(SignatureHeader), body) {
		return "", ErrUnauthorized
	}
	// Ordinary comment notifications have no Agent identity block. Acknowledge
	// only the known non-Agent shape, after authentication and before applying
	// Agent-specific fields. Agent claims always keep the strict validation path.
	if isIssueCommentNotification(body) {
		log.FromContext(ctx).Info("plane notification ignored", "app_hash", identityHash(app.GetId()), "event_kind", "issue_comment")
		return DecisionIgnored, nil
	}
	var p webhookPayload
	defer func() {
		if errors.Is(resultErr, ErrIdentity) {
			resultErr = p.identityDiagnostic(resultErr)
		}
	}()
	if err := json.Unmarshal(body, &p); err != nil {
		if errors.Is(err, ErrIdentity) {
			return "", err
		}
		return "", webhookJSONDiagnostic(err)
	}
	for _, field := range []struct{ name, value string }{
		{"app_client_id", p.AppClientID}, {"workspace_id", p.WorkspaceID}, {"agent_user_id", p.AgentUserID},
	} {
		if field.value == "" {
			return "", rejectWebhookIdentity("missing", field.name, "", "")
		}
	}
	if p.AppClientID != app.GetClientId() {
		return "", rejectWebhookIdentity("app_client_mismatch", "app_client_id", app.GetClientId(), p.AppClientID)
	}
	if p.Type != "agent_run" && p.Type != "agent_run_activity" {
		return "", fmt.Errorf("%w: unsupported event type", ErrMalformed)
	}
	if (p.Type == "agent_run" && p.Action != "created") || (p.Type == "agent_run_activity" && p.Action != "prompted") {
		return "", fmt.Errorf("%w: unsupported event action", ErrMalformed)
	}
	if p.Run.ID == "" {
		return "", ErrMalformed
	}
	for _, field := range []struct{ name, expected, actual string }{
		{"agent_run.workspace", p.WorkspaceID, p.Run.WorkspaceID},
		{"agent_run.agent_user", p.AgentUserID, p.Run.AgentUserID},
		{"activity.workspace", p.WorkspaceID, p.Activity.WorkspaceID},
		{"activity.agent_run", p.Run.ID, p.Activity.AgentRunID},
	} {
		if field.actual != "" && field.actual != field.expected {
			return "", rejectWebhookIdentity("conflict", field.name, field.expected, field.actual)
		}
	}
	list, err := r.installations.List(ctx, app.GetWorkspaceId(), app.GetId())
	if err != nil {
		return "", err
	}
	inst, err := resolveWebhookInstallation(&p, list)
	if err != nil {
		return "", err
	}
	e := &Event{WorkspaceID: app.GetWorkspaceId(), AppID: app.GetId(), InstallationID: inst.GetId(), PlaneInstallationID: inst.GetAppInstallationId(), PlaneWorkspaceID: inst.GetPlaneWorkspaceId(), BotUserID: inst.GetAppBotUserId(), AgentRunID: p.Run.ID, ActivityID: p.Activity.ID, Prompt: p.Prompt, Signal: p.Activity.Signal, ActorUserID: p.Activity.UserID, Type: p.Type, Action: p.Action, DeliveryID: headers.Get(DeliveryHeader), EventHeader: headers.Get(EventHeader), ReceivedAt: r.now().UnixMilli(), Raw: append([]byte(nil), body...)}
	e.AgentID = app.GetAgentId()
	e.Stop = strings.EqualFold(strings.TrimSpace(e.Signal), "stop")
	if e.Stop || e.Type != "agent_run" {
		switch {
		case e.ActorUserID == "":
			return "", rejectWebhookIdentity("missing", "activity.actor", "", "")
		case strings.EqualFold(e.ActorUserID, e.BotUserID):
			return "", rejectWebhookIdentity("actor_is_bot", "activity.actor", "", e.ActorUserID)
		case !allowedStopUser(app.GetAllowedUserIds(), e.ActorUserID):
			return "", rejectWebhookIdentity("actor_not_allowed", "activity.actor", "", e.ActorUserID)
		}
	}
	if e.Type == "agent_run" {
		return DecisionIgnored, nil
	}
	if !e.Stop && !app.GetInboundEnabled() {
		return DecisionIgnored, nil
	}
	if !e.Stop {
		var input userinput.Turn
		if strings.TrimSpace(e.Prompt) == "" {
			return "", fmt.Errorf("%w: empty prompt", ErrMalformed)
		}
		if err := input.AddText(e.Prompt); err != nil {
			return "", fmt.Errorf("%w: %v", ErrMalformed, err)
		}
	}
	if e.ActivityID == "" {
		return "", ErrMalformed
	}
	if _, err := r.queue.Accept(ctx, e); err != nil {
		if errors.Is(err, ErrDuplicate) {
			return DecisionDuplicate, nil
		}
		return "", err
	}
	log.FromContext(ctx).Info("plane webhook queued", planeEventLogFields(e)...)
	// The worker owns the RECEIVED projection. A slow projection must never
	// delay acknowledging an event already owned by the durable Stream.
	return DecisionAccepted, nil
}
func validSignature(secret []byte, encoded string, body []byte) bool {
	expected, err := hex.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(body)
	return hmac.Equal(expected, mac.Sum(nil))
}

func allowedStopUser(allowed []string, user string) bool {
	user = strings.TrimSpace(user)
	if user == "" {
		return false
	}
	if len(allowed) == 0 {
		return true
	}
	for _, candidate := range allowed {
		if strings.EqualFold(strings.TrimSpace(candidate), user) {
			return true
		}
	}
	return false
}

type Event struct {
	// AcceptedOrder is assigned at durable admission, never by a Worker.
	AcceptedOrder       int64  `json:"accepted_order,omitempty"`
	StopTarget          string `json:"stop_target,omitempty"`
	WorkspaceID         string `json:"workspace_id"`
	AppID               string `json:"app_id"`
	InstallationID      string `json:"installation_id"`
	PlaneInstallationID string `json:"plane_installation_id"`
	PlaneWorkspaceID    string `json:"plane_workspace_id"`
	BotUserID           string `json:"bot_user_id"`
	AgentRunID          string `json:"agent_run_id"`
	ActivityID          string `json:"activity_id"`
	Type                string `json:"type"`
	Action              string `json:"action"`
	DeliveryID          string `json:"delivery_id"`
	EventHeader         string `json:"event_header"`
	ReceivedAt          int64  `json:"received_at"`
	Prompt              string `json:"prompt"`
	Signal              string `json:"signal,omitempty"`
	ActorUserID         string `json:"actor_user_id,omitempty"`
	Stop                bool   `json:"stop,omitempty"`
	Raw                 []byte `json:"raw"`
	AgentID             string `json:"agent_id,omitempty"`
	// The batch is reconstructed from the Stream, never serialized as a
	// second pending store.
	Members        []*Event     `json:"-"`
	InvocationID   string       `json:"-"`
	ExecutionFence uint64       `json:"-"`
	ActivityType   ActivityType `json:"activity_type,omitempty"`
}

func (e *Event) Encode() (string, error) { b, err := json.Marshal(e); return string(b), err }
func DecodeEvent(s string) (*Event, error) {
	var e Event
	if err := json.Unmarshal([]byte(s), &e); err != nil {
		return nil, err
	}
	return &e, nil
}
