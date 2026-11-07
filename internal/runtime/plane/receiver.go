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
	"slices"
	"strings"
	"time"

	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	planeinstallation "go.orx.me/apps/butter/internal/repo/planeinstallation"
	"go.orx.me/apps/butter/internal/secretbox"
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

var requiredAgentScopes = []string{"agents.runs:write", "agents.runs:read", "agents.activities:write", "agents.activities:read"}

type Decision string

const (
	DecisionAccepted  Decision = "accepted"
	DecisionDuplicate Decision = "duplicate"
	DecisionIgnored   Decision = "ignored"
)

type Queue interface {
	Accept(context.Context, *Event) (string, error)
}
type Receiver struct {
	apps          planerepo.Repository
	installations planeinstallation.Repository
	keyring       *secretbox.Keyring
	queue         Queue
	now           func() time.Time
}

func NewReceiver(apps planerepo.Repository, installations planeinstallation.Repository, keyring *secretbox.Keyring, queue Queue) *Receiver {
	return &Receiver{apps: apps, installations: installations, keyring: keyring, queue: queue, now: time.Now}
}

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
func (r *Receiver) Deliver(ctx context.Context, app *agentsv1.PlaneApp, headers http.Header, body []byte) (Decision, error) {
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
	if !validSignature(secret, headers.Get(SignatureHeader), body) {
		return "", ErrUnauthorized
	}
	var p webhookPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return "", fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	if p.AppClientID == "" || p.AppClientID != app.GetClientId() || p.AppInstallationID == "" || p.WorkspaceID == "" || p.AgentUserID == "" {
		return "", ErrIdentity
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
	if p.Run.WorkspaceID != "" && p.WorkspaceID != p.Run.WorkspaceID {
		return "", ErrIdentity
	}
	list, err := r.installations.List(ctx, app.GetWorkspaceId(), app.GetId())
	if err != nil {
		return "", err
	}
	var inst *agentsv1.PlaneInstallation
	for _, candidate := range list {
		if candidate.GetAppInstallationId() == p.AppInstallationID {
			inst = candidate
			break
		}
	}
	if inst == nil || inst.GetPlaneWorkspaceId() != p.WorkspaceID || inst.GetAppBotUserId() != p.AgentUserID {
		return "", ErrIdentity
	}
	for _, required := range requiredAgentScopes {
		if !slices.Contains(inst.GetGrantedScopes(), required) {
			return "", ErrIdentity
		}
	}
	e := &Event{WorkspaceID: app.GetWorkspaceId(), AppID: app.GetId(), InstallationID: inst.GetId(), PlaneInstallationID: inst.GetAppInstallationId(), PlaneWorkspaceID: inst.GetPlaneWorkspaceId(), BotUserID: inst.GetAppBotUserId(), AgentRunID: p.Run.ID, ActivityID: p.Activity.ID, Type: p.Type, Action: p.Action, DeliveryID: headers.Get(DeliveryHeader), EventHeader: headers.Get(EventHeader), ReceivedAt: r.now().UnixMilli(), Raw: append([]byte(nil), body...)}
	if e.ActivityID == "" {
		e.ActivityID = e.AgentRunID + ":" + e.Action
	}
	if _, err := r.queue.Accept(ctx, e); err != nil {
		if errors.Is(err, ErrDuplicate) {
			return DecisionDuplicate, nil
		}
		return "", err
	}
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

type webhookPayload struct {
	Type, Action, AppClientID, AppInstallationID, AgentUserID, WorkspaceID string
	Run                                                                    struct{ ID, WorkspaceID string }
	Activity                                                               struct{ ID string }
}

func (p *webhookPayload) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type              string                           `json:"type"`
		Action            string                           `json:"action"`
		AppClientID       string                           `json:"app_client_id"`
		AppInstallationID string                           `json:"app_installation_id"`
		AgentUserID       string                           `json:"agent_user_id"`
		WorkspaceID       string                           `json:"workspace_id"`
		AgentRun          struct{ ID, WorkspaceID string } `json:"agent_run"`
		Run               struct{ ID, WorkspaceID string } `json:"run"`
		AgentActivity     struct{ ID string }              `json:"agent_activity"`
		Activity          struct{ ID string }              `json:"activity"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.Type, p.Action, p.AppClientID, p.AppInstallationID, p.AgentUserID, p.WorkspaceID = raw.Type, raw.Action, raw.AppClientID, raw.AppInstallationID, raw.AgentUserID, raw.WorkspaceID
	p.Run = raw.AgentRun
	if p.Run.ID == "" {
		p.Run = raw.Run
	}
	p.Activity = raw.AgentActivity
	if p.Activity.ID == "" {
		p.Activity = raw.Activity
	}
	return nil
}

type Event struct {
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
	Raw                 []byte `json:"raw"`
}

func (e *Event) Encode() (string, error) { b, err := json.Marshal(e); return string(b), err }
func DecodeEvent(s string) (*Event, error) {
	var e Event
	if err := json.Unmarshal([]byte(s), &e); err != nil {
		return nil, err
	}
	return &e, nil
}
