package plane

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Keep the existing v1 decoder's acceptance semantics. Never format its error
// or UnmarshalTypeError.Value: both can contain bytes from the signed body.
func webhookJSONDiagnostic(err error) error {
	reason, field, expected, actual := "json_decode", "$", "unknown", "unknown"
	var offset int64
	if mismatch, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		reason, field = "json_type", webhookJSONField(mismatch.Field)
		expected, offset = webhookJSONKind(mismatch.Type), mismatch.Offset
		switch mismatch.Value {
		case "object", "array", "string", "null":
			actual = mismatch.Value
		case "bool":
			actual = "boolean"
		default:
			if mismatch.Value == "number" || strings.HasPrefix(mismatch.Value, "number ") {
				actual = "number"
			}
		}
	} else if syntax, ok := errors.AsType[*json.SyntaxError](err); ok {
		reason, offset = "json_syntax", syntax.Offset
	}
	return fmt.Errorf("%w: invalid JSON payload reason=%s field=%s expected_type=%s actual_kind=%s offset=%d",
		ErrMalformed, reason, field, expected, actual, offset)
}

// Paths come from typed struct tags, not arbitrary JSON keys. Keep a fixed
// vocabulary as a second guard against input values entering diagnostics.
func webhookJSONField(field string) string {
	if field == "" {
		return "$"
	}
	for part := range strings.SplitSeq(field, ".") {
		switch part {
		case "event", "data", "prompt", "type", "action", "app_client_id", "app_installation_id", "agent_user_id", "workspace_id",
			"agent_run", "run", "agent_run_activity", "agent_activity", "activity", "id", "workspace", "agent_user",
			"signal", "user_id", "actor", "content", "body":
		default:
			return "$"
		}
	}
	return field
}

// Classification is diagnostic only, never routing or authorization. Unknown
// discriminator values might contain input text, so never format them verbatim.
func (p *webhookPayload) identityDiagnostic(err error) error {
	kind, action, event := "other", "other", "other"
	switch p.Type {
	case "":
		kind = "absent"
	case "agent_run", "agent_run_activity", "issue", "issue_comment", "work_item", "work_item_comment":
		kind = p.Type
	}
	switch p.Action {
	case "":
		action = "absent"
	case "created", "prompted", "updated", "deleted":
		action = p.Action
	}
	switch p.TransportEvent {
	case "":
		event = "absent"
	case "agent_run_create", "agent_run_user_prompt", "agent_run", "agent_run_activity", "issue", "issue_comment", "work_item", "work_item_comment":
		event = p.TransportEvent
	}
	return fmt.Errorf("%w payload_type=%s payload_action=%s event_kind=%s", err, kind, action, event)
}

func webhookJSONKind(t reflect.Type) string {
	if t == nil {
		return "unknown"
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Struct, reflect.Map:
		return "object"
	case reflect.Array, reflect.Slice:
		return "array"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	default:
		return "unknown"
	}
}
