package runner

import (
	"encoding/json"
	"fmt"
)

// RenderEventOutput is the text of a workflow Event.Output: a string as is,
// anything else as JSON. A turn with no text answer shows this instead.
func RenderEventOutput(output any) string {
	if text, ok := output.(string); ok {
		return text
	}
	if encoded, err := json.Marshal(output); err == nil {
		return string(encoded)
	}
	return fmt.Sprint(output)
}
