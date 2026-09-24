package pibox

import (
	"strings"

	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/runtime/butterboxconn"
)

// stateKeyPrefix scopes the per-agent pi session binding inside ADK session
// state: one butter session holds one pi session per PI agent (ADR-0011 §3),
// so switching agents and back resumes the earlier pi session.
const stateKeyPrefix = "pibox:"

func stateKey(agentID string) string { return stateKeyPrefix + agentID }

// binding records which pi session a (butter session × agent) pair is
// attached to, together with the box and directory it was created against.
// A repointed agent (different box or directory) invalidates the binding:
// sessions are abandoned and recreated, never migrated, because a session's
// file references are only honest in the directory it ran in.
type binding struct {
	PiSessionID string
	ButterboxID string
	WorkingDir  string
}

// stateValue renders the binding as the plain map that goes into an event's
// StateDelta. Only JSON/BSON-friendly primitives, so it round-trips through
// any session backend.
func (b binding) stateValue() map[string]any {
	return map[string]any{
		"pi_session_id": b.PiSessionID,
		"butterbox_id":  b.ButterboxID,
		"working_dir":   b.WorkingDir,
	}
}

// matches reports whether the binding still points at the agent's current
// box and working directory.
func (b binding) matches(butterboxID, workingDir string) bool {
	return b.ButterboxID == butterboxID && b.WorkingDir == workingDir
}

// readBinding loads the agent's pi session binding from session state.
// A missing key or an unparseable value reads as "no binding" — the bridge
// then creates a fresh pi session, which is always safe.
func readBinding(state adksession.State, agentID string) (binding, bool) {
	v, err := state.Get(stateKey(agentID))
	if err != nil {
		return binding{}, false
	}
	return bindingFromValue(v)
}

// bindingFromValue parses a stored state value (map or Mongo-decoded bson.D).
func bindingFromValue(v any) (binding, bool) {
	m, ok := butterboxconn.StateMap(v)
	if !ok {
		return binding{}, false
	}
	b := binding{
		PiSessionID: butterboxconn.StringField(m, "pi_session_id"),
		ButterboxID: butterboxconn.StringField(m, "butterbox_id"),
		WorkingDir:  butterboxconn.StringField(m, "working_dir"),
	}
	if strings.TrimSpace(b.PiSessionID) == "" {
		return binding{}, false
	}
	return b, true
}
