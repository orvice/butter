package pipeline

import (
	"slices"
	"strings"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Admit reports whether the message passes every admission rule. A rule passes
// when its allowlist is empty (no restriction), when it contains the rule's
// value, or when the value is empty and the rule is marked SkipWhenEmpty.
// It is exported so transport adapters can reuse it for out-of-band checks
// (e.g. validating callback-query senders against the same allowlists).
func Admit(rules []AdmissionRule) bool {
	for _, rule := range rules {
		if len(rule.Allowlist) == 0 {
			continue
		}
		if rule.Value == "" && rule.SkipWhenEmpty {
			continue
		}
		if !slices.Contains(rule.Allowlist, rule.Value) {
			return false
		}
	}
	return true
}

// matchesTrigger reports whether the message satisfies at least one configured
// trigger. No triggers means accept all.
func matchesTrigger(triggers []*agentsv1.AgentTrigger, msg IncomingMessage) bool {
	if len(triggers) == 0 {
		return true
	}
	for _, trigger := range triggers {
		switch trigger.GetType() {
		case agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_MESSAGE:
			return true
		case agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_PRIVATE_CHAT:
			if msg.IsPrivate {
				return true
			}
		case agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_COMMAND:
			if isCommand(msg.Text) {
				return true
			}
		}
	}
	return false
}

// isCommand reports whether text is a slash command.
func isCommand(text string) bool {
	return strings.HasPrefix(text, "/")
}
