package pibox

import (
	"fmt"

	"google.golang.org/adk/v2/agent"

	"go.orx.me/apps/butter/internal/a2ui"
	internalagent "go.orx.me/apps/butter/internal/agent"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// AgentBuilder builds each Pi leaf with the Card Policy of its place in the
// agent tree, just as the factory does for an LLM leaf.
func AgentBuilder(factory ClientFactory) internalagent.PiAgentBuilder {
	return func(pb *agentsv1.Agent, policy a2ui.CardPolicy) (agent.Agent, error) {
		b := NewBridge(pb, factory)
		b.cardPolicy = policy
		description := pb.GetDescription()
		if description == "" {
			description = fmt.Sprintf("Pi agent on ButterBox %s", b.butterboxID)
		}
		return b.BuildAgent(pb.GetName(), description)
	}
}
