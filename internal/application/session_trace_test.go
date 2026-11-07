package application

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/runtime/tracing"
)

// TestEventToProtoWithTraceAddressesTheTrace: the deep link and the trace ID
// handed to the UI are the OTel trace ID derived from the event's invocation ID,
// not the invocation ID itself — Langfuse has no trace named after the latter.
func TestEventToProtoWithTraceAddressesTheTrace(t *testing.T) {
	const host = "https://cloud.langfuse.com"
	const invocationID = "e-0f3d4c3b-4c8c-4a1f-9a1a-8d8a5e6f7a01"

	evt := session.NewEvent(context.Background(), invocationID)
	evt.Author = "root-agent"

	got := eventToProtoWithTrace(evt, host)

	wantTraceID := "0f3d4c3b4c8c4a1f9a1a8d8a5e6f7a01"
	if got.TraceId != wantTraceID {
		t.Fatalf("trace_id = %q, want %q", got.TraceId, wantTraceID)
	}
	if want := host + "/trace/" + wantTraceID; got.TraceUrl != want {
		t.Fatalf("trace_url = %q, want %q", got.TraceUrl, want)
	}
	if got.InvocationId != invocationID {
		t.Fatalf("invocation_id = %q, want %q", got.InvocationId, invocationID)
	}
}

// TestEventToProtoWithTraceWithoutHost: with no Langfuse host wired there is no
// link, but the trace ID is still the derived one.
func TestEventToProtoWithTraceWithoutHost(t *testing.T) {
	evt := session.NewEvent(context.Background(), "e-0f3d4c3b-4c8c-4a1f-9a1a-8d8a5e6f7a01")

	got := eventToProtoWithTrace(evt, "")

	if got.TraceUrl != "" {
		t.Fatalf("trace_url = %q, want empty", got.TraceUrl)
	}
	if got.TraceId != "0f3d4c3b4c8c4a1f9a1a8d8a5e6f7a01" {
		t.Fatalf("trace_id = %q, want the derived trace ID", got.TraceId)
	}
}

// TestEventToProtoWithTraceLegacyInvocation: an invocation ID that is not an
// ADK-minted UUID still maps onto a stable trace ID, so a session recorded
// before the mapping existed keeps a well-formed link.
func TestEventToProtoWithTraceLegacyInvocation(t *testing.T) {
	evt := session.NewEvent(context.Background(), "legacy-invocation")

	got := eventToProtoWithTrace(evt, "https://langfuse.example.com")

	want := tracing.TraceIDString("legacy-invocation")
	if got.TraceId != want {
		t.Fatalf("trace_id = %q, want %q", got.TraceId, want)
	}
	if len(want) != 32 || strings.ContainsAny(want, "-") {
		t.Fatalf("trace ID %q is not 32 hex characters", want)
	}
	if wantURL := "https://langfuse.example.com/trace/" + want; got.TraceUrl != wantURL {
		t.Fatalf("trace_url = %q, want %q", got.TraceUrl, wantURL)
	}
}

// TestEventToProtoWithTraceNoInvocation: an event with no invocation belongs to
// no trace and gets neither a trace ID nor a link.
func TestEventToProtoWithTraceNoInvocation(t *testing.T) {
	got := eventToProtoWithTrace(session.NewEvent(context.Background(), ""), "https://cloud.langfuse.com")

	if got.TraceId != "" || got.TraceUrl != "" {
		t.Fatalf("trace_id = %q, trace_url = %q, want both empty", got.TraceId, got.TraceUrl)
	}
}
