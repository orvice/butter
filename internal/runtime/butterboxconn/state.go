package butterboxconn

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// StateMap reads a session-binding value stored in ADK session state as a
// plain map. In-memory sessions retain the map written in StateDelta, while
// Mongo decodes the nested document as a bson.D when loading it into
// map[string]any state.
func StateMap(v any) (map[string]any, bool) {
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

// StringField reads one field of a StateMap as a string.
func StringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
