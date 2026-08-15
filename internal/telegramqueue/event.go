// Package telegramqueue is the durable hand-off between Telegram receive
// (Webhook handler or Long Polling leader) and the workers that run Agents
// (issue #264).
//
// Redis Streams is the queue, not a cache. Webhook acknowledgement is defined
// as "the event is durably in the Stream": returning 200 before that would
// tell Telegram the update is handled while it exists only in one Pod's
// memory. This is why enabling a Webhook Channel requires Redis to be
// configured for persistence and no eviction.
//
// The event carries a frozen snapshot of the routing decision. A worker may
// pick it up seconds later, after the Destination was edited or disabled;
// carrying the revision lets the worker notice, and carrying the policy means
// a worker never re-derives routing from configuration that has since moved.
package telegramqueue

import (
	"encoding/json"
	"fmt"
)

// EventKind distinguishes what a worker is expected to do with an event.
type EventKind string

const (
	// KindDestinationUpdate is an update that matched an exact Destination.
	KindDestinationUpdate EventKind = "destination_update"
	// KindWhere is the transport-level `/where` command, which is accepted
	// at addresses no Destination covers and invokes no Agent.
	KindWhere EventKind = "where"
)

// DestinationPolicy is the frozen inbound policy of a Destination at the
// moment the update was accepted.
type DestinationPolicy struct {
	AgentID            string   `json:"agent_id,omitempty"`
	Model              string   `json:"model,omitempty"`
	SelectableAgentIDs []string `json:"selectable_agent_ids,omitempty"`
	SelectableModels   []string `json:"selectable_models,omitempty"`
	TriggerMode        string   `json:"trigger_mode,omitempty"`
	SessionPolicy      string   `json:"session_policy,omitempty"`
	AllowedUserIDs     []string `json:"allowed_user_ids,omitempty"`
	ControllerUserIDs  []string `json:"controller_user_ids,omitempty"`
	ReplyMode          string   `json:"reply_mode,omitempty"`
	// DebugDefault is tri-state: nil means the Destination never chose, which
	// resolves to debug enabled.
	DebugDefault *bool `json:"debug_default,omitempty"`
}

// Address is the exact Telegram address an event belongs to.
type Address struct {
	ChatID          string `json:"chat_id"`
	MessageThreadID string `json:"message_thread_id,omitempty"`
}

// Event is one accepted Telegram update.
type Event struct {
	Kind EventKind `json:"kind"`

	// WorkspaceID, ChannelID, and BotID identify the transport. They are
	// frozen because the public callback route carries only a Channel ID and
	// the worker must not re-derive tenancy.
	WorkspaceID string `json:"workspace_id"`
	ChannelID   string `json:"channel_id"`
	BotID       string `json:"bot_id"`
	BotUsername string `json:"bot_username,omitempty"`

	// DestinationID and DestinationRevision are empty for `/where`.
	DestinationID       string             `json:"destination_id,omitempty"`
	DestinationRevision int64              `json:"destination_revision,omitempty"`
	Policy              *DestinationPolicy `json:"policy,omitempty"`

	Address Address `json:"address"`

	// UpdateID is Telegram's own update identifier, used for deduplication
	// together with ChannelID.
	UpdateID int64 `json:"update_id"`

	// Update is the raw Telegram update JSON. Media is deliberately not
	// downloaded here — only metadata travels through Redis.
	Update json.RawMessage `json:"update"`

	// ReceivedAtUnixMs stamps acceptance so a worker can report queue delay.
	ReceivedAtUnixMs int64 `json:"received_at_unix_ms"`
}

// Encode renders the event for the Stream payload field.
func (e *Event) Encode() (string, error) {
	raw, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("encode telegram event: %w", err)
	}
	return string(raw), nil
}

// DecodeEvent parses a Stream payload.
func DecodeEvent(payload string) (*Event, error) {
	event := &Event{}
	if err := json.Unmarshal([]byte(payload), event); err != nil {
		return nil, fmt.Errorf("decode telegram event: %w", err)
	}
	return event, nil
}
