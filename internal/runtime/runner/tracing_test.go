package runner

import (
	"context"
	"strings"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/tracing"
)

// TestRun_InvocationIsPinnedBeforeTheRun: the runner decides the invocation ID
// before ADK creates the invocation, so the ID the run reports is the one the
// trace ID is derived from — that is what makes SessionEvent.trace_url address a
// trace that exists.
func TestRun_InvocationIsPinnedBeforeTheRun(t *testing.T) {
	exporter := installTraceProvider(t)
	backend := newFakeBackend(t)
	agents := approvalAgents()
	svc := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, session.InMemoryService())

	result, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", turnCtxInfo(&agents[0]), nil, nil)
	if err != nil {
		t.Fatalf("RunTurn: %v", err)
	}
	if result.InvocationID == "" {
		t.Fatal("turn reported no invocation ID")
	}
	if !strings.HasPrefix(result.InvocationID, "e-") {
		t.Fatalf("invocation ID %q lacks the ADK prefix", result.InvocationID)
	}

	spans := exporter.spans()
	if len(spans) == 0 {
		t.Fatal("no spans exported")
	}
	want := tracing.TraceIDForInvocation(result.InvocationID)
	for _, span := range spans {
		if got := span.SpanContext().TraceID(); got != want {
			t.Fatalf("span %q trace ID = %s, want %s (invocation %s)",
				span.Name(), got, want, result.InvocationID)
		}
	}
}

// TestRun_ResumedWorkflowStaysInOneTrace is the reason the invocation ID is
// pinned at all: a workflow pauses on a Human Input node in one request and
// resumes in another, and ADK reuses the paused invocation's ID for the
// continuation. Both legs must report spans in one trace, or the pause and its
// answer show up as two unrelated Langfuse traces.
func TestRun_ResumedWorkflowStaysInOneTrace(t *testing.T) {
	exporter := installTraceProvider(t)
	backend := newFakeBackend(t)
	agents := approvalAgents()
	svc := buildWorkflowService(t, backend, agents, []string{"drafter", "publisher"}, session.InMemoryService())
	ctxInfo := turnCtxInfo(&agents[0])

	pause, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "write a post"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("pause turn: %v", err)
	}
	if !pause.Interrupted() {
		t.Fatal("workflow did not pause on the Human Input node")
	}
	pauseSpans := exporter.spans()
	if len(pauseSpans) == 0 {
		t.Fatal("the pause turn exported no spans")
	}

	resume, err := svc.RunTurn(context.Background(), "approval",
		[]*genai.Part{{Text: "approved"}}, "", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("resume turn: %v", err)
	}
	if resume.Interrupted() {
		t.Fatalf("resume turn still reports pending Interrupts: %+v", resume.Pending)
	}
	allSpans := exporter.spans()
	if len(allSpans) == len(pauseSpans) {
		t.Fatal("the resume turn exported no spans")
	}

	// The continuation reuses the paused invocation, so both legs keep one ID
	// and one trace.
	if resume.InvocationID != pause.InvocationID {
		t.Fatalf("resume invocation = %q, want the paused one %q", resume.InvocationID, pause.InvocationID)
	}
	want := tracing.TraceIDForInvocation(pause.InvocationID)
	for _, span := range allSpans {
		if got := span.SpanContext().TraceID(); got != want {
			t.Fatalf("span %q trace ID = %s, want %s (invocation %s)",
				span.Name(), got, want, pause.InvocationID)
		}
	}
}

var (
	tracerProviderOnce sync.Once
	spanRecorder       = &traceRecorder{}
)

// installTraceProvider routes the spans of the test through a recorder with the
// package's ID generator installed, and empties it first.
//
// The global provider is installed once per process and never restored: OTel's
// global is a delegating proxy, and putting that proxy back into the global slot
// after a test leaves the next one exporting into a shut-down provider.
func installTraceProvider(t *testing.T) *traceRecorder {
	t.Helper()
	tracerProviderOnce.Do(func() {
		otel.SetTracerProvider(sdktrace.NewTracerProvider(
			sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(spanRecorder)),
			sdktrace.WithIDGenerator(tracing.NewIDGenerator()),
		))
	})
	spanRecorder.reset()
	return spanRecorder
}

// traceRecorder is a span exporter that keeps what it exported.
type traceRecorder struct {
	mu    sync.Mutex
	saved []sdktrace.ReadOnlySpan
}

func (r *traceRecorder) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saved = append(r.saved, spans...)
	return nil
}

func (r *traceRecorder) Shutdown(context.Context) error { return nil }

func (r *traceRecorder) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saved = nil
}

func (r *traceRecorder) spans() []sdktrace.ReadOnlySpan {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]sdktrace.ReadOnlySpan(nil), r.saved...)
}
