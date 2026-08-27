// Package cursorbox bridges Cursor SDK Bridge sessions hosted on a ButterBox
// into the ADK agent interface so an AGENT_TYPE_CURSOR leaf can be invoked
// like any other agent. Mirrors internal/runtime/pibox for PI agents.
//
// The bridge drives a synchronous turn API: SendMessage blocks until the
// Cursor agent finishes (the box's CursorService collects RunUpdate events
// internally), so there is no poll loop. One Cursor session exists per
// (butter session × agent), keyed in ADK session state; on a repointed agent
// or a session the box no longer knows, the bridge abandons and recreates.
package cursorbox

import "context"

// CursorClient abstracts the CursorService RPCs exposed by a ButterBox.
// The production implementation wraps the Connect client generated from
// butter-box's cursor.v1 proto (issue #315). Tests use a fake.
type CursorClient interface {
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error)
	SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)
	AbortSession(ctx context.Context, req *AbortSessionRequest) error
}

// CreateSessionRequest mirrors CursorService.CreateSession.
type CreateSessionRequest struct {
	Name       string // human-readable session name
	WorkingDir string // absolute or relative to box sandbox root
	Model      string // Cursor model ID, empty = default
	Mode       string // "agent" or "plan", empty = "agent"
}

// CreateSessionResponse is the box's answer to CreateSession.
type CreateSessionResponse struct {
	SessionID string // opaque session ID on the box
}

// SendMessageRequest mirrors CursorService.SendMessage.
type SendMessageRequest struct {
	SessionID string
	Message   string
	Images    []ImageContent
}

// ImageContent passes one inline image to the Cursor agent.
type ImageContent struct {
	MIMEType string
	Data     []byte
}

// SendMessageResponse is the box's answer after the Cursor agent finishes a turn.
type SendMessageResponse struct {
	Text string // the agent's final text output
}

// AbortSessionRequest mirrors CursorService.AbortSession.
type AbortSessionRequest struct {
	SessionID string
}
