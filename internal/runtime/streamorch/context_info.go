// Package streamorch is the shared agent-run streaming orchestration used by
// StreamAgent and (future) adapters: ContextInfo construction and the
// event-loop-to-sink dispatch that turns ADK session.Events into ordered
// frames.
package streamorch

import (
	"errors"

	"github.com/google/uuid"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ContextInfoInput carries the caller-supplied fields needed to build a
// ContextInfo for a streaming run, plus the defaults/invariants each
// streaming adapter must apply identically.
type ContextInfoInput struct {
	AppName       string
	UserID        string
	SessionID     string
	SessionPrefix string
	WorkspaceID   string
	HasWorkspace  bool
	IsAdmin       bool
	Source        agentsv1.ContextSource
	ChatType      agentsv1.ChatType
}

// NewContextInfo builds a ContextInfo, defaulting AppName/UserID to "api"
// and generating a SessionID with the given prefix when none is supplied.
// A non-admin caller with no workspace in context is rejected — an empty
// WorkspaceId makes the runner treat the call as a system path.
func NewContextInfo(in ContextInfoInput) (*agentsv1.ContextInfo, error) {
	if !in.HasWorkspace && !in.IsAdmin {
		return nil, errors.New("workspace required (set X-Workspace-ID header)")
	}
	appName := in.AppName
	if appName == "" {
		appName = "api"
	}
	userID := in.UserID
	if userID == "" {
		userID = "api"
	}
	sessionID := in.SessionID
	if sessionID == "" {
		sessionID = in.SessionPrefix + uuid.NewString()
	}
	invocationID := uuid.NewString()
	if id, err := uuid.NewV7(); err == nil {
		invocationID = id.String()
	}
	return &agentsv1.ContextInfo{
		Uuid:        invocationID,
		ChannelName: appName,
		UserId:      userID,
		SessionId:   sessionID,
		WorkspaceId: in.WorkspaceID,
		Source:      in.Source,
		ChatType:    in.ChatType,
	}, nil
}
