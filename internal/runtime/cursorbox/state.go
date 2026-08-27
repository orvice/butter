package cursorbox

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	adksession "google.golang.org/adk/v2/session"
)

// stateKeyPrefix scopes the per-agent Cursor session binding inside ADK
// session state: one butter session holds one Cursor session per Cursor agent,
// so switching agents and back resumes the earlier Cursor session.
const stateKeyPrefix = "cursorbox:"

func stateKey(agentID string) string { return stateKeyPrefix + agentID }

// binding records which Cursor session a (butter session × agent) pair is
// attached to, together with the box and directory it was created against.
// A repointed agent (different box or directory) invalidates the binding.
type binding struct {
	CursorSessionID string
	ButterboxID     string
	WorkingDir      string
}

func (b binding) stateValue() map[string]any {
	return map[string]any{
		"cursor_session_id": b.CursorSessionID,
		"butterbox_id":      b.ButterboxID,
		"working_dir":       b.WorkingDir,
	}
}

func (b binding) matches(butterboxID, workingDir string) bool {
	return b.ButterboxID == butterboxID && b.WorkingDir == workingDir
}

func readBinding(state adksession.State, agentID string) (binding, bool) {
	v, err := state.Get(stateKey(agentID))
	if err != nil {
		return binding{}, false
	}
	return bindingFromValue(v)
}

func bindingFromValue(v any) (binding, bool) {
	m, ok := bindingValueMap(v)
	if !ok {
		return binding{}, false
	}
	b := binding{
		CursorSessionID: stringField(m, "cursor_session_id"),
		ButterboxID:     stringField(m, "butterbox_id"),
		WorkingDir:      stringField(m, "working_dir"),
	}
	if strings.TrimSpace(b.CursorSessionID) == "" {
		return binding{}, false
	}
	return b, true
}

func bindingValueMap(v any) (map[string]any, bool) {
	switch value := v.(type) {
	case map[string]any:
		return value, true
	case bson.M:
		return map[string]any(value), true
	case bson.D:
		m := make(map[string]any, len(value))
		for _, elem := range value {
			m[elem.Key] = elem.Value
		}
		return m, true
	default:
		return nil, false
	}
}

func stringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
