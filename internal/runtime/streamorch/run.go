package streamorch

import (
	"context"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Runner is the subset of runner.Service's streaming entry point the
// orchestrator needs — a seam so tests can supply a fake.
type Runner interface {
	RunSSE(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error)
}

// RunIdentity carries the identifiers that travel together across every
// frame of one streaming run, so Sink implementations and callers pass one
// value instead of the same strings repeated per method.
type RunIdentity struct {
	InvocationID string
	SessionID    string
	AgentName    string
	// AgentID is the resolved agent's immutable agent_id; empty when the
	// agent has not been assigned one yet.
	AgentID string
}

// AgentRef identifies the resolved agent for a streaming run: the registered
// runtime name plus its immutable agent_id (empty when not yet assigned).
type AgentRef struct {
	Name string
	ID   string
}

// Sink receives ordered frames for a streaming run. Implementations are not
// required to be concurrency-safe: Run calls them serially.
type Sink interface {
	Started(id RunIdentity) error
	TextDelta(id RunIdentity, text string) error
	RunEvent(id RunIdentity, evt *session.Event) error
	Final(id RunIdentity, response string) error
}

// Run drives a RunSSE-style call against r, translating ADK session.Events
// into ordered Sink frames, and returns the raw run error (or the first sink
// error) on failure. Callers map the returned error to their own transport's
// error scheme.
func Run(ctx context.Context, r Runner, agentRef AgentRef, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, sink Sink) error {
	agentName := agentRef.Name
	id := RunIdentity{
		InvocationID: ctxInfo.GetUuid(),
		SessionID:    ctxInfo.GetSessionId(),
		AgentName:    agentName,
		AgentID:      agentRef.ID,
	}

	if err := sink.Started(id); err != nil {
		return err
	}

	response, runErr := r.RunSSE(ctx, agentName, parts, modelOverride, ctxInfo, func(evt *session.Event) {
		deltas := textParts(evt)
		for _, text := range deltas {
			sink.TextDelta(id, text)
		}
		// Pure text-only partial events are surfaced as TextDelta only.
		// Mixed events (text + function call etc.) emit both a TextDelta for
		// each text chunk AND a RunEvent carrying the full event so adapters
		// that render tool calls (e.g. the dashboard) see them.
		if len(deltas) > 0 && hasOnlyTextParts(evt) {
			return
		}
		sink.RunEvent(id, evt)
	}, nil)

	if runErr != nil {
		return runErr
	}
	return sink.Final(id, response)
}
