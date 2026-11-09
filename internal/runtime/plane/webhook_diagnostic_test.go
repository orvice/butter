package plane

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestWebhookJSONDiagnosticSafeFallbacks(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"unknown error", errors.New("sensitive-decoder-message"), "reason=json_decode field=$ expected_type=unknown actual_kind=unknown offset=0"},
		{"unrecognized path and kind", &json.UnmarshalTypeError{Value: "sensitive-value", Type: reflect.TypeFor[string](), Field: "data.sensitive-field", Offset: 17}, "reason=json_type field=$ expected_type=string actual_kind=unknown offset=17"},
		{"wrapped decoder error", fmt.Errorf("sensitive-context: %w", &json.UnmarshalTypeError{Value: "number 9876543210123456789", Type: reflect.TypeFor[string](), Field: "prompt", Offset: 29}), "reason=json_type field=prompt expected_type=string actual_kind=number offset=29"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := webhookJSONDiagnostic(tc.err)
			if !errors.Is(err, ErrMalformed) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected safe diagnostic: %v", err)
			}
			if strings.Contains(err.Error(), "sensitive-") || strings.Contains(err.Error(), "9876543210123456789") {
				t.Fatal("diagnostic formatted untrusted decoder content")
			}
		})
	}
}

func TestReceiverAuthenticatesBeforeJSONDiagnostics(t *testing.T) {
	f := newReceiverFixture(t)
	_, err := f.receiver.Deliver(t.Context(), f.app, nil, []byte(`{"prompt":`))
	if !errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrMalformed) || strings.Contains(err.Error(), "reason=json_") || len(f.queue.events) != 0 {
		t.Fatalf("unsigned malformed input reached JSON diagnostics: %v", err)
	}
}

// These are diagnostic probes, not claims about the live Plane schema. They
// exercise signed receive and must remain rejected without exposing the value.
func TestReceiverJSONDiagnostics(t *testing.T) {
	for _, tc := range []struct{ name, body, reason, field, expected, actual string }{
		{"invalid envelope marker", `{"event":{"private":"sensitive-prompt"},"data":{}}`, "json_type", "event", "string", "object"},
		{"expanded actor", `{"agent_run_activity":{"actor":{"id":"sensitive-actor-id","name":"sensitive-prompt"}}}`, "json_type", "agent_run_activity.actor", "string", "object"},
		{"enveloped actor", `{"data":{"agent_run_activity":{"actor":{"id":"sensitive-actor-id"}}}}`, "json_type", "data.agent_run_activity.actor", "string", "object"},
		{"string resource", `{"agent_run":"sensitive-run-id"}`, "json_type", "agent_run", "object", "string"},
		{"array envelope", `{"data":[]}`, "json_type", "data", "object", "array"},
		{"root array", `[]`, "json_type", "$", "object", "array"},
		{"numeric prompt", `{"prompt":9876543210123456789}`, "json_type", "prompt", "string", "number"},
		{"boolean prompt", `{"prompt":true}`, "json_type", "prompt", "string", "boolean"},
		{"syntax", `{"prompt":"sensitive-prompt","extra":sensitive-invalid-token}`, "json_syntax", "$", "unknown", "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newReceiverFixture(t)
			_, err := f.deliverBody(t, []byte(tc.body))
			if !errors.Is(err, ErrMalformed) || errors.Is(err, ErrIdentity) || len(f.queue.events) != 0 {
				t.Fatalf("malformed input changed admission: err=%v queued=%d", err, len(f.queue.events))
			}
			for _, part := range []string{"reason=" + tc.reason, "field=" + tc.field + " ", "expected_type=" + tc.expected, "actual_kind=" + tc.actual} {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("diagnostic missing %q: %v", part, err)
				}
			}
			_, offset, ok := strings.Cut(err.Error(), "offset=")
			n, parseErr := strconv.ParseInt(offset, 10, 64)
			if !ok || parseErr != nil || n <= 0 {
				t.Errorf("missing decoder offset: %v", err)
			}
			want := fmt.Sprintf("%s: invalid JSON payload reason=%s field=%s expected_type=%s actual_kind=%s offset=%d",
				ErrMalformed, tc.reason, tc.field, tc.expected, tc.actual, n)
			if err.Error() != want {
				t.Errorf("diagnostic contains unexpected decoder details: %v", err)
			}
			for _, sensitive := range []string{"sensitive-actor-id", "sensitive-run-id", "sensitive-prompt", "sensitive-invalid-token", "9876543210123456789", "fixture-webhook-secret", tc.body} {
				if strings.Contains(err.Error(), sensitive) {
					t.Errorf("diagnostic leaks input value %q", sensitive)
				}
			}
		})
	}
}
