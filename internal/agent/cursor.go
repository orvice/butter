package agent

import (
	"fmt"
	"strings"

	"google.golang.org/adk/v2/agent"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// CursorAgentBuilder builds the ADK agent for one AGENT_TYPE_CURSOR leaf. It
// is implemented by the runtime layer (internal/runtime/cursorbox), which
// owns the ButterBox resolution and the CursorService bridge; nil means
// CURSOR agents cannot be built in this deployment.
type CursorAgentBuilder func(pb *agentsv1.Agent) (agent.Agent, error)

// BoxAgentBuilders carries the runtime-provided builders for box-backed leaf
// agents. A nil *BoxAgentBuilders (or a nil field) means that agent type is
// not available in this deployment.
type BoxAgentBuilders struct {
	Pi     PiAgentBuilder
	Cursor CursorAgentBuilder
}

func (b *BoxAgentBuilders) pi() PiAgentBuilder {
	if b == nil {
		return nil
	}
	return b.Pi
}

func (b *BoxAgentBuilders) cursor() CursorAgentBuilder {
	if b == nil {
		return nil
	}
	return b.Cursor
}

// Cursor agent modes accepted by the box's CursorService. Empty means
// CursorModeAgent.
const (
	CursorModeAgent = "agent"
	CursorModePlan  = "plan"
)

// ValidateCursorAgent checks an AGENT_TYPE_CURSOR config. Like
// ValidatePiAgent it is pure proto validation — the box-existence check
// lives with the repositories. Non-CURSOR agents pass through.
//
// A CURSOR agent is a leaf whose behavior surface lives on the box: rules
// come from the working directory's .cursor/rules, tools from Cursor's
// built-ins and the box's mcp.json. Every butter-side behavior field is
// therefore rejected on write.
func ValidateCursorAgent(pb *agentsv1.Agent) error {
	if pb == nil || pb.GetType() != agentsv1.AgentType_AGENT_TYPE_CURSOR {
		return nil
	}
	if err := validateCursorConfig(pb); err != nil {
		return fmt.Errorf("agent %q: %w", pb.GetName(), err)
	}
	return nil
}

func validateCursorConfig(pb *agentsv1.Agent) error {
	c := pb.GetConfig().GetCursor()
	if c == nil {
		return fmt.Errorf("a cursor agent requires config.cursor with the ButterBox binding")
	}
	if strings.TrimSpace(c.GetButterboxId()) == "" {
		return fmt.Errorf("config.cursor.butterbox_id is required: a cursor agent binds one ButterBox")
	}
	switch strings.TrimSpace(c.GetMode()) {
	case "", CursorModeAgent, CursorModePlan:
	default:
		return fmt.Errorf("config.cursor.mode %q is not supported: use %q (default) or %q", c.GetMode(), CursorModeAgent, CursorModePlan)
	}
	if c.MaxRunSeconds != nil && c.GetMaxRunSeconds() < 0 {
		return fmt.Errorf("config.cursor.max_run_seconds must not be negative (unset defaults to 1800, 0 means unlimited)")
	}
	if len(pb.GetChildAgentIds()) > 0 {
		return fmt.Errorf("child_agent_ids is not supported: a cursor agent is a leaf — compose it as a child or workflow node of another agent instead")
	}
	return rejectBoxOwnedFields(pb.GetConfig(), "cursor", "Cursor's rules and tools are configured on the box, in the working directory's .cursor/rules and mcp.json")
}

// newCursorAgent validates the CURSOR config and delegates construction to
// the runtime-provided builder.
func newCursorAgent(pb *agentsv1.Agent, builder CursorAgentBuilder) (agent.Agent, error) {
	if err := ValidateCursorAgent(pb); err != nil {
		return nil, err
	}
	if builder == nil {
		return nil, fmt.Errorf("agent %q: cursor agents are not available: no ButterBox bridge is wired", pb.GetName())
	}
	return builder(pb)
}
