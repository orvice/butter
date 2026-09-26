package mem0memory

import (
	"context"
	"fmt"
)

// Target names one of the two memory scopes (ADR-0013 §3).
type Target int

const (
	// TargetWorkspace is Workspace Memory: one pool shared by every member,
	// entry point, and memory-enabled agent of the workspace.
	TargetWorkspace Target = iota + 1
	// TargetAgent is Agent Memory: private to one agent, written only
	// explicitly.
	TargetAgent
)

func (t Target) String() string {
	switch t {
	case TargetWorkspace:
		return "workspace"
	case TargetAgent:
		return "agent"
	default:
		return fmt.Sprintf("Target(%d)", int(t))
	}
}

// WorkspaceUserID is the mem0 `user_id` Workspace Memory is stored under.
// The workspace lives in the identity field itself, not only in metadata,
// so a forgotten filter cannot cross tenants.
func WorkspaceUserID(workspaceID string) string {
	return "ws:" + workspaceID
}

// AgentMemoryID is the mem0 `agent_id` one agent's Agent Memory is stored
// under.
func AgentMemoryID(workspaceID, agentID string) string {
	return "ws:" + workspaceID + ":agent:" + agentID
}

// Scope identifies whose memory a call reads and writes, plus the
// provenance a write records. The runner builds one per turn from the
// invocation's root agent.
type Scope struct {
	WorkspaceID string
	// AgentID is the root agent's Agent ID (not its runtime name). Empty
	// disables Agent Memory for the call.
	AgentID string
	// SessionID and InvocationID are the ADK session and invocation of the
	// turn. Capture selects the turn's events by InvocationID.
	SessionID    string
	InvocationID string
	// Channel and Principal record where the turn came from and who sent
	// it, so a later per-person scope can filter on them without migrating
	// data.
	Channel   string
	Principal string
}

type scopeKey struct{}

// WithScope returns a context carrying the memory scope, for callers that
// reach the service through the ADK memory.Service interface.
func WithScope(ctx context.Context, scope Scope) context.Context {
	return context.WithValue(ctx, scopeKey{}, scope)
}

// ScopeFromContext returns the scope set by WithScope.
func ScopeFromContext(ctx context.Context) (Scope, bool) {
	scope, ok := ctx.Value(scopeKey{}).(Scope)
	return scope, ok
}

// identity returns the mem0 identity fields for one target. Workspace
// Memory carries only `user_id` and Agent Memory only `agent_id`: adding the
// other field, or `run_id`, would narrow mem0's add-time dedup.
func (s Scope) identity(target Target) (userID, agentID string, err error) {
	if s.WorkspaceID == "" {
		return "", "", fmt.Errorf("memory scope has no workspace")
	}
	switch target {
	case TargetWorkspace:
		return WorkspaceUserID(s.WorkspaceID), "", nil
	case TargetAgent:
		if s.AgentID == "" {
			return "", "", fmt.Errorf("memory scope has no agent for agent memory")
		}
		return "", AgentMemoryID(s.WorkspaceID, s.AgentID), nil
	default:
		return "", "", fmt.Errorf("unknown memory target %v", target)
	}
}

// filters returns the mem0 search filter for one target.
func (s Scope) filters(target Target) (map[string]any, error) {
	userID, agentID, err := s.identity(target)
	if err != nil {
		return nil, err
	}
	if userID != "" {
		return map[string]any{"user_id": userID}, nil
	}
	return map[string]any{"agent_id": agentID}, nil
}

// metadata returns the provenance recorded on a write. Keys are namespaced
// so they cannot collide with mem0's own payload fields (`agent_id` is an
// identity key there).
func (s Scope) metadata() map[string]any {
	meta := map[string]any{"butter_workspace_id": s.WorkspaceID}
	for k, v := range map[string]string{
		"butter_agent_id":      s.AgentID,
		"butter_session_id":    s.SessionID,
		"butter_invocation_id": s.InvocationID,
		"butter_channel":       s.Channel,
		"butter_principal":     s.Principal,
	} {
		if v != "" {
			meta[k] = v
		}
	}
	return meta
}
