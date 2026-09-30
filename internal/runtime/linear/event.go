// Package linear runs Linear Agent Sessions (ADR-0015): the receive path
// that authenticates and queues Linear's AgentSessionEvent webhooks, the
// worker that drains the queue on every Pod, and the orchestrator that
// turns one event into an Agent turn answered in the session.
package linear

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Linear webhook headers.
const (
	SignatureHeader = "Linear-Signature"
	DeliveryHeader  = "Linear-Delivery"
)

// Session event actions.
const (
	ActionCreated  = "created"
	ActionPrompted = "prompted"
)

// webhookPayload is the part of an AgentSessionEvent webhook Butter reads.
// Linear's payload is richer; unknown fields are ignored.
type webhookPayload struct {
	Type             string `json:"type"`
	Action           string `json:"action"`
	OrganizationID   string `json:"organizationId"`
	WebhookTimestamp int64  `json:"webhookTimestamp"` // unix ms
	// PromptContext is Linear's formatted issue, comment and guidance
	// context for the session.
	PromptContext string `json:"promptContext"`
	AgentSession  struct {
		ID        string    `json:"id"`
		CreatorID string    `json:"creatorId"`
		Creator   *userRef  `json:"creator"`
		Issue     *issueRef `json:"issue"`
		Comment   *struct {
			Body   string   `json:"body"`
			UserID string   `json:"userId"`
			User   *userRef `json:"user"`
		} `json:"comment"`
	} `json:"agentSession"`
	// AgentActivity is the user's prompt on a prompted event.
	AgentActivity *struct {
		ID      string   `json:"id"`
		Signal  string   `json:"signal"`
		UserID  string   `json:"userId"`
		User    *userRef `json:"user"`
		Content struct {
			Body string `json:"body"`
		} `json:"content"`
	} `json:"agentActivity"`
}

type userRef struct {
	ID string `json:"id"`
}

type issueRef struct {
	ID          string `json:"id"`
	Identifier  string `json:"identifier"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// promptingUserID finds who drove the event. Linear's payload names the
// user in several places depending on the action; the first present wins.
func (p *webhookPayload) promptingUserID() string {
	var candidates []string
	if a := p.AgentActivity; a != nil {
		candidates = append(candidates, a.UserID)
		if a.User != nil {
			candidates = append(candidates, a.User.ID)
		}
	}
	candidates = append(candidates, p.AgentSession.CreatorID)
	if p.AgentSession.Creator != nil {
		candidates = append(candidates, p.AgentSession.Creator.ID)
	}
	if c := p.AgentSession.Comment; c != nil {
		candidates = append(candidates, c.UserID)
		if c.User != nil {
			candidates = append(candidates, c.User.ID)
		}
	}
	for _, id := range candidates {
		if id = strings.TrimSpace(id); id != "" {
			return id
		}
	}
	return ""
}

// Issue identifies the issue a session is about.
type Issue struct {
	ID         string `json:"id,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	Title      string `json:"title,omitempty"`
	URL        string `json:"url,omitempty"`
}

// Label names the issue for people, e.g. "ENG-123: Fix login".
func (i Issue) Label() string {
	switch {
	case i.Identifier == "":
		return i.Title
	case i.Title == "":
		return i.Identifier
	default:
		return i.Identifier + ": " + i.Title
	}
}

// Event is one accepted Linear delivery: a frozen snapshot of the routing
// decision and of the parts of the payload the worker acts on. The public
// route carries only an App ID, so tenancy is frozen here rather than
// re-derived from the payload.
type Event struct {
	WorkspaceID    string `json:"workspace_id"`
	AppID          string `json:"app_id"`
	AppRevision    int64  `json:"app_revision"`
	InstallationID string `json:"installation_id"`
	OrganizationID string `json:"organization_id"`

	AgentSessionID  string `json:"agent_session_id"`
	Action          string `json:"action"`
	Stop            bool   `json:"stop,omitempty"`
	PromptText      string `json:"prompt_text,omitempty"`
	PromptingUserID string `json:"prompting_user_id,omitempty"`
	PromptContext   string `json:"prompt_context,omitempty"`
	Issue           Issue  `json:"issue,omitzero"`

	DeliveryID       string `json:"delivery_id"`
	ReceivedAtUnixMs int64  `json:"received_at_unix_ms"`
}

// Encode renders the event for the Stream payload.
func (e *Event) Encode() (string, error) {
	raw, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("encode linear event: %w", err)
	}
	return string(raw), nil
}

// DecodeEvent parses a Stream payload.
func DecodeEvent(payload string) (*Event, error) {
	event := &Event{}
	if err := json.Unmarshal([]byte(payload), event); err != nil {
		return nil, fmt.Errorf("decode linear event: %w", err)
	}
	return event, nil
}
