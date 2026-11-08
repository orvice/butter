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
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
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
	processing    planeprocessing.Repository
	keyring       *secretbox.Keyring
	queue         Queue
	now           func() time.Time
}

func NewReceiver(apps planerepo.Repository, installations planeinstallation.Repository, keyring *secretbox.Keyring, queue Queue) *Receiver {
	return &Receiver{apps: apps, installations: installations, keyring: keyring, queue: queue, now: time.Now}
}

func (r *Receiver) SetProcessingRepo(repo planeprocessing.Repository) { r.processing = repo }

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
	e := &Event{WorkspaceID: app.GetWorkspaceId(), AppID: app.GetId(), InstallationID: inst.GetId(), PlaneInstallationID: inst.GetAppInstallationId(), PlaneWorkspaceID: inst.GetPlaneWorkspaceId(), BotUserID: inst.GetAppBotUserId(), AgentRunID: p.Run.ID, ActivityID: p.Activity.ID, Prompt: p.Prompt, Signal: p.Activity.Signal, ActorUserID: p.Activity.UserID, Type: p.Type, Action: p.Action, DeliveryID: headers.Get(DeliveryHeader), EventHeader: headers.Get(EventHeader), ReceivedAt: r.now().UnixMilli(), Raw: append([]byte(nil), body...)}
	e.AgentID = app.GetAgentId()
	e.Stop = strings.EqualFold(strings.TrimSpace(e.Signal), "stop")
	if (e.Stop || e.Type != "agent_run") && (!allowedStopUser(app.GetAllowedUserIds(), e.ActorUserID) || strings.EqualFold(e.ActorUserID, e.BotUserID)) {
		return "", ErrIdentity
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
	if r.processing != nil {
		// The Stream already owns the event; the Worker repairs this projection
		// if Mongo is unavailable or it races the first processing attempt.
		_ = (&Processor{processing: r.processing}).project(ctx, e, agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_RECEIVED, "", "", "")
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

type webhookActivity struct {
	ID      string `json:"id"`
	Signal  string `json:"signal"`
	UserID  string `json:"user_id"`
	Content struct {
		Body string `json:"body"`
	} `json:"content"`
}

type webhookPayload struct {
	Prompt, Type, Action, AppClientID, AppInstallationID, AgentUserID, WorkspaceID string
	Run                                                                            struct{ ID, WorkspaceID string }
	Activity                                                                       webhookActivity
}

func (p *webhookPayload) UnmarshalJSON(data []byte) error {
	var raw struct {
		Prompt            string                           `json:"prompt"`
		Type              string                           `json:"type"`
		Action            string                           `json:"action"`
		AppClientID       string                           `json:"app_client_id"`
		AppInstallationID string                           `json:"app_installation_id"`
		AgentUserID       string                           `json:"agent_user_id"`
		WorkspaceID       string                           `json:"workspace_id"`
		AgentRun          struct{ ID, WorkspaceID string } `json:"agent_run"`
		Run               struct{ ID, WorkspaceID string } `json:"run"`
		AgentActivity     webhookActivity                  `json:"agent_activity"`
		Activity          webhookActivity                  `json:"activity"`
		Data              json.RawMessage                  `json:"data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	aliases := []webhookActivity{raw.AgentActivity, raw.Activity}
	if len(raw.Data) != 0 && string(raw.Data) != "null" {
		var nested struct {
			Prompt            string                           `json:"prompt"`
			Type              string                           `json:"type"`
			Action            string                           `json:"action"`
			AppClientID       string                           `json:"app_client_id"`
			AppInstallationID string                           `json:"app_installation_id"`
			AgentUserID       string                           `json:"agent_user_id"`
			WorkspaceID       string                           `json:"workspace_id"`
			AgentRun          struct{ ID, WorkspaceID string } `json:"agent_run"`
			Run               struct{ ID, WorkspaceID string } `json:"run"`
			AgentActivity     webhookActivity                  `json:"agent_run_activity"`
			Activity          webhookActivity                  `json:"activity"`
		}
		if err := json.Unmarshal(raw.Data, &nested); err != nil {
			return err
		}
		aliases = append(aliases, nested.AgentActivity, nested.Activity)
		if raw.Type == "" {
			raw.Type = nested.Type
		}
		if raw.Action == "" {
			raw.Action = nested.Action
		}
		if raw.AppClientID == "" {
			raw.AppClientID = nested.AppClientID
		}
		if raw.AppInstallationID == "" {
			raw.AppInstallationID = nested.AppInstallationID
		}
		if raw.AgentUserID == "" {
			raw.AgentUserID = nested.AgentUserID
		}
		if raw.WorkspaceID == "" {
			raw.WorkspaceID = nested.WorkspaceID
		}
		if raw.Run.ID == "" {
			raw.Run = nested.Run
		}
		if raw.Run.ID == "" {
			raw.Run = nested.AgentRun
		}
		if raw.Activity.ID == "" {
			raw.Activity.ID = nested.Activity.ID
		}
		if raw.Activity.ID == "" {
			raw.Activity = nested.AgentActivity
		}
		if raw.Prompt == "" {
			raw.Prompt = nested.AgentActivity.Content.Body
		}
		if raw.Prompt == "" {
			raw.Prompt = nested.Activity.Content.Body
		}
		if raw.AgentActivity.Signal == "" {
			raw.AgentActivity.Signal = nested.AgentActivity.Signal
		}
		if raw.AgentActivity.UserID == "" {
			raw.AgentActivity.UserID = nested.AgentActivity.UserID
		}
		if raw.Activity.Signal == "" {
			raw.Activity.Signal = nested.Activity.Signal
		}
		if raw.Activity.UserID == "" {
			raw.Activity.UserID = nested.Activity.UserID
		}
	}
	var actor, activityID, signal string
	for _, alias := range aliases {
		for _, pair := range []struct {
			value  string
			target *string
		}{
			{alias.UserID, &actor}, {alias.ID, &activityID}, {alias.Signal, &signal},
		} {
			value := strings.TrimSpace(pair.value)
			if value == "" {
				continue
			}
			if *pair.target != "" && !strings.EqualFold(*pair.target, value) {
				return ErrIdentity
			}
			*pair.target = value
		}
	}
	p.Type, p.Action, p.AppClientID, p.AppInstallationID, p.AgentUserID, p.WorkspaceID = raw.Type, raw.Action, raw.AppClientID, raw.AppInstallationID, raw.AgentUserID, raw.WorkspaceID
	p.Prompt, p.Type, p.Action, p.AppClientID, p.AppInstallationID, p.AgentUserID, p.WorkspaceID = raw.Prompt, raw.Type, raw.Action, raw.AppClientID, raw.AppInstallationID, raw.AgentUserID, raw.WorkspaceID
	if p.Run.ID == "" {
		p.Run = raw.Run
	}
	p.Activity = raw.AgentActivity
	if p.Activity.ID == "" {
		p.Activity = raw.Activity
	}
	p.Activity.UserID, p.Activity.Signal = actor, signal
	return nil
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
