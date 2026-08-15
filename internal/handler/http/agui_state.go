package http

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"google.golang.org/adk/v2/session"
)

// AG-UI shared state is the session's state map, server-authoritative
// (ADR-0002: the session is the single source of truth). State flows
// server → client only: STATE_SNAPSHOT re-baselines the client's mirror and
// STATE_DELTA carries incremental changes; a client-modified state is never
// adopted — the next run answers it with a corrective snapshot instead.

// aguiStateHiddenPrefixes are the ADK state scopes that never leave the
// server: app: and user: state is shared beyond this session, and temp: keys
// are invocation-scoped bookkeeping.
var aguiStateHiddenPrefixes = []string{
	session.KeyPrefixApp,
	session.KeyPrefixUser,
	session.KeyPrefixTemp,
}

// aguiVisibleState filters a session state map down to the keys a client may
// mirror, normalizing every value to its JSON shape so later comparisons and
// diffs are representation-independent (Mongo and the JSON decoder disagree
// about number and array types).
func aguiVisibleState(all map[string]any) map[string]any {
	visible := make(map[string]any, len(all))
	for k, v := range all {
		hidden := false
		for _, prefix := range aguiStateHiddenPrefixes {
			if strings.HasPrefix(k, prefix) {
				hidden = true
				break
			}
		}
		if !hidden {
			visible[k] = normalizeJSONValue(v)
		}
	}
	return visible
}

// sessionStateMap materializes a session's state into a plain map. A nil
// session or state reads as empty.
func sessionStateMap(sess session.Session) map[string]any {
	out := map[string]any{}
	if sess == nil {
		return out
	}
	st := sess.State()
	if st == nil {
		return out
	}
	for k, v := range st.All() {
		out[k] = v
	}
	return out
}

// normalizeJSONValue round-trips a value through JSON so bson-decoded and
// JSON-decoded representations of the same document compare equal. A value
// that cannot be marshalled normalizes to nil.
func normalizeJSONValue(v any) any {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

// aguiStatesEqual compares two JSON-normalized state maps.
func aguiStatesEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	return reflect.DeepEqual(normalizeJSONValue(a), normalizeJSONValue(b))
}

// aguiStateDiffOps computes the RFC 6902 operations that transform prev into
// next. Keys are top-level session state keys, so every path is a single
// escaped JSON Pointer token. Ops are key-sorted for deterministic output.
func aguiStateDiffOps(prev, next map[string]any) []aguievents.JSONPatchOperation {
	var ops []aguievents.JSONPatchOperation
	keys := make([]string, 0, len(next))
	for k := range next {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		nv := normalizeJSONValue(next[k])
		pv, existed := prev[k]
		if existed && reflect.DeepEqual(normalizeJSONValue(pv), nv) {
			continue
		}
		op := "add"
		if existed {
			op = "replace"
		}
		ops = append(ops, aguievents.JSONPatchOperation{Op: op, Path: jsonPointerToken(k), Value: nv})
	}
	removed := make([]string, 0)
	for k := range prev {
		if _, ok := next[k]; !ok {
			removed = append(removed, k)
		}
	}
	sort.Strings(removed)
	for _, k := range removed {
		ops = append(ops, aguievents.JSONPatchOperation{Op: "remove", Path: jsonPointerToken(k)})
	}
	return ops
}

// jsonPointerToken escapes one state key as an RFC 6901 pointer path.
func jsonPointerToken(key string) string {
	key = strings.ReplaceAll(key, "~", "~0")
	key = strings.ReplaceAll(key, "/", "~1")
	return "/" + key
}
