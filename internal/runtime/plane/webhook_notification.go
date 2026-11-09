package plane

import (
	"bytes"
	"encoding/json"
)

// Use the same v1 field matching as the Agent decoder: a case-folded Agent
// discriminator must not be hidden by a stricter/different notification probe.
// This is a bounded filter for the ordinary event observed in production, not
// a catch-all for unknown events or incomplete Agent payloads.
func isIssueCommentNotification(body []byte) bool {
	var envelope struct {
		notificationAgentClaims
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
	}
	if json.Unmarshal(body, &envelope) != nil || envelope.Event != "issue_comment" || envelope.hasAgentClaim(false) {
		return false
	}
	data := bytes.TrimSpace(envelope.Data)
	if len(data) == 0 || data[0] != '{' {
		return false
	}
	var nested notificationAgentClaims
	return json.Unmarshal(data, &nested) == nil && !nested.hasAgentClaim(true)
}

type notificationAgentClaims struct {
	Type             string          `json:"type"`
	AgentRun         json.RawMessage `json:"agent_run"`
	Run              json.RawMessage `json:"run"`
	AgentRunActivity json.RawMessage `json:"agent_run_activity"`
	AgentActivity    json.RawMessage `json:"agent_activity"`
	Activity         json.RawMessage `json:"activity"`
}

func (c notificationAgentClaims) hasAgentClaim(nested bool) bool {
	if c.Type != "" && c.Type != "issue_comment" {
		return true
	}
	for _, claim := range []json.RawMessage{c.AgentRun, c.Run, c.AgentRunActivity, c.AgentActivity} {
		if notificationClaimPresent(claim) {
			return true
		}
	}
	// At the root this is opaque transport metadata. Inside data it can be a
	// legacy business alias and must not be silently discarded.
	return nested && notificationClaimPresent(c.Activity)
}

func notificationClaimPresent(claim json.RawMessage) bool {
	claim = bytes.TrimSpace(claim)
	return len(claim) != 0 && !bytes.Equal(claim, []byte("null"))
}
