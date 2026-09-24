package cursorbox

import (
	"strings"

	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/runtime/butterboxconn"
)

// stateKeyPrefix scopes the per-agent Cursor session binding inside ADK
// session state: one butter session holds one Cursor session per CURSOR
// agent, so switching agents and back resumes the earlier Cursor session.
const stateKeyPrefix = "cursorbox:"

func stateKey(agentID string) string { return stateKeyPrefix + agentID }

// binding records which Cursor session a (butter session × agent) pair is
// attached to, together with the box and directory it was created against.
// A repointed agent invalidates the binding: sessions are abandoned and
// recreated, never migrated.
type binding struct {
	CursorSessionID string
	ButterboxID     string
	WorkingDir      string
}

// stateValue renders the binding as the plain map that goes into an event's
// StateDelta.
func (b binding) stateValue() map[string]any {
	return map[string]any{
		"cursor_session_id": b.CursorSessionID,
		"butterbox_id":      b.ButterboxID,
		"working_dir":       b.WorkingDir,
	}
}

// matches reports whether the binding still points at the agent's current
// box and working directory.
func (b binding) matches(butterboxID, workingDir string) bool {
	return b.ButterboxID == butterboxID && b.WorkingDir == workingDir
}

// readBinding loads the agent's Cursor session binding from session state.
// A missing key or an unparseable value reads as "no binding" — the bridge
// then creates a fresh session, which is always safe.
func readBinding(state adksession.State, agentID string) (binding, bool) {
	v, err := state.Get(stateKey(agentID))
	if err != nil {
		return binding{}, false
	}
	m, ok := butterboxconn.StateMap(v)
	if !ok {
		return binding{}, false
	}
	b := binding{
		CursorSessionID: butterboxconn.StringField(m, "cursor_session_id"),
		ButterboxID:     butterboxconn.StringField(m, "butterbox_id"),
		WorkingDir:      butterboxconn.StringField(m, "working_dir"),
	}
	if strings.TrimSpace(b.CursorSessionID) == "" {
		return binding{}, false
	}
	return b, true
}
