package runner

import (
	"encoding/json"
	"fmt"
)

func renderEventOutput(output any) string {
	if text, ok := output.(string); ok {
		return text
	}
	if encoded, err := json.Marshal(output); err == nil {
		return string(encoded)
	}
	return fmt.Sprint(output)
}
