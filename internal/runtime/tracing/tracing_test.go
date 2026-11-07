package tracing

import (
	"context"
	"iter"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/platform"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

func TestTraceIDForInvocationMapsADKUUID(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   string
	}{
		{name: "adk prefix", id: invocationPrefix + uuid.NewString()},
		{name: "bare uuid", id: uuid.NewString()},
		{name: "uppercase uuid", id: strings.ToUpper(uuid.NewString())},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := TraceIDForInvocation(tc.id)
			if !got.IsValid() {
				t.Fatalf("TraceIDForInvocation(%q) is invalid", tc.id)
			}
			compact := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(tc.id, invocationPrefix, ""), "-", ""))
			if got.String() != compact {
				t.Fatalf("TraceIDForInvocation(%q) = %s, want %s", tc.id, got, compact)
			}
			if TraceIDString(tc.id) != got.String() {
				t.Fatalf("TraceIDString(%q) = %s, TraceIDForInvocation = %s", tc.id, TraceIDString(tc.id), got)
			}
		})
	}
}

func TestTraceIDForInvocationIsStableAndDistinct(t *testing.T) {
	seen := map[trace.TraceID]string{}
	for i := 0; i < 64; i++ {
		id := invocationPrefix + uuid.NewString()
		tid := TraceIDForInvocation(id)
		if tid != TraceIDForInvocation(id) {
			t.Fatalf("TraceIDForInvocation(%q) is not deterministic", id)
		}
		// The rendered trace ID maps back to itself, which is what a Langfuse
		// URL carrying it relies on.
		if tid != TraceIDForInvocation(tid.String()) {
			t.Fatalf("trace ID %s does not round-trip through its own hex form", tid)
		}
		if other, dup := seen[tid]; dup {
			t.Fatalf("trace ID collision between %q and %q: %s", id, other, tid)
		}
		seen[tid] = id
	}
}

func TestTraceIDForInvocationHashesNonUUID(t *testing.T) {
	for _, id := range []string{"inv-1", "legacy-invocation", "e-not-a-uuid", "e-1234"} {
		tid := TraceIDForInvocation(id)
		if !tid.IsValid() {
			t.Fatalf("TraceIDForInvocation(%q) is invalid", id)
		}
		if tid != TraceIDForInvocation(id) {
			t.Fatalf("TraceIDForInvocation(%q) is not deterministic", id)
		}
	}
	if TraceIDForInvocation("inv-1") == TraceIDForInvocation("inv-2") {
		t.Fatal("distinct non-UUID invocation IDs share a trace ID")
	}
	if got := TraceIDString("   "); got != "" {
		t.Fatalf("TraceIDString(blank) = %q, want empty", got)
	}
	if got := TraceIDForInvocation(""); got.IsValid() {
		t.Fatalf("TraceIDForInvocation(\"\") = %s, want invalid", got)
	}
}

// TestStartInvocationSeedsADKUUIDProvider: the invocation ID Butter mints is the
// one ADK uses, and only that first ID is seeded — the events and function calls
// ADK mints afterwards keep random IDs.
func TestStartInvocationSeedsADKUUIDProvider(t *testing.T) {
	ctx, id := StartInvocation(context.Background())

	if !strings.HasPrefix(id, invocationPrefix) {
		t.Fatalf("invocation ID %q lacks the ADK prefix %q", id, invocationPrefix)
	}
	if _, err := uuid.Parse(strings.TrimPrefix(id, invocationPrefix)); err != nil {
		t.Fatalf("invocation ID %q does not carry a UUID: %v", id, err)
	}
	if got := InvocationIDFromContext(ctx); got != id {
		t.Fatalf("InvocationIDFromContext = %q, want %q", got, id)
	}

	// ADK itself prepends the prefix, so the provider hands out the bare UUID.
	first := platform.NewUUID(ctx)
	if want := strings.TrimPrefix(id, invocationPrefix); first != want {
		t.Fatalf("first UUID ADK minted = %q, want %q", first, want)
	}
	second := platform.NewUUID(ctx)
	if second == first {
		t.Fatal("the UUID provider reuses the invocation ID for later IDs")
	}
	if _, err := uuid.Parse(second); err != nil {
		t.Fatalf("later UUID %q is not a UUID: %v", second, err)
	}
}

func TestWithInvocationIDIgnoresBlank(t *testing.T) {
	ctx := WithInvocationID(context.Background(), "  ")
	if got := InvocationIDFromContext(ctx); got != "" {
		t.Fatalf("InvocationIDFromContext = %q, want empty", got)
	}
}

func TestInvocationIDForMessage(t *testing.T) {
	const (
		pausedInvocation = "e-11111111-1111-1111-1111-111111111111"
		laterInvocation  = "e-22222222-2222-2222-2222-222222222222"
	)
	sess := newSession(t,
		eventWithCall(pausedInvocation, "call-1"),
		eventWithCall(laterInvocation, "call-2"),
	)

	for _, tc := range []struct {
		name  string
		parts []*genai.Part
		want  string
	}{
		{
			name:  "answers an older call",
			parts: []*genai.Part{answer("call-1")},
			want:  pausedInvocation,
		},
		{
			name:  "answers the latest call",
			parts: []*genai.Part{answer("call-2")},
			want:  laterInvocation,
		},
		{
			name:  "unknown call starts a new invocation",
			parts: []*genai.Part{answer("call-3")},
			want:  "",
		},
		{
			name:  "plain text starts a new invocation",
			parts: []*genai.Part{{Text: "hello"}},
			want:  "",
		},
		{
			name:  "no parts starts a new invocation",
			parts: nil,
			want:  "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := InvocationIDForMessage(sess, tc.parts); got != tc.want {
				t.Fatalf("InvocationIDForMessage = %q, want %q", got, tc.want)
			}
		})
	}

	if got := InvocationIDForMessage(nil, []*genai.Part{answer("call-1")}); got != "" {
		t.Fatalf("InvocationIDForMessage(nil session) = %q, want empty", got)
	}
}

func TestGeneratorPinsTraceToInvocation(t *testing.T) {
	gen := NewIDGenerator()
	ctx, id := StartInvocation(context.Background())

	tid, sid := gen.NewIDs(ctx)
	if want := TraceIDForInvocation(id); tid != want {
		t.Fatalf("NewIDs trace ID = %s, want %s", tid, want)
	}
	if !sid.IsValid() {
		t.Fatal("NewIDs span ID is invalid")
	}

	// Span IDs stay random: a resumed execution reuses the trace, so a fixed
	// span ID would collide with the spans the paused execution reported.
	if _, otherSID := gen.NewIDs(ctx); otherSID == sid {
		t.Fatal("NewIDs reused a span ID inside the same trace")
	}
	if got := gen.NewSpanID(ctx, tid); !got.IsValid() || got == sid {
		t.Fatalf("NewSpanID = %s, want a fresh valid span ID", got)
	}

	// A continuation pins the invocation it continues.
	continued := WithInvocationID(context.Background(), id)
	if got, _ := gen.NewIDs(continued); got != tid {
		t.Fatalf("continuation trace ID = %s, want %s", got, tid)
	}
}

func TestGeneratorWithoutInvocationIDIsRandom(t *testing.T) {
	gen := NewIDGenerator()
	tid, _ := gen.NewIDs(context.Background())
	if !tid.IsValid() {
		t.Fatal("random trace ID is invalid")
	}
	if other, _ := gen.NewIDs(context.Background()); other == tid {
		t.Fatalf("two random trace IDs are equal: %s", tid)
	}
}

// TestGeneratorPinsRealADKInvocation runs an agent through the ADK runner with
// the generator installed on the tracer provider and checks that the invocation
// reports the pinned ID and that every span of the run carries the trace ID
// derived from it. That is the mapping SessionEvent.trace_url is built from.
func TestGeneratorPinsRealADKInvocation(t *testing.T) {
	exporter := &spanRecorder{}
	installTracerProvider(t, exporter)

	root, err := llmagent.New(llmagent.Config{Name: "root-agent", Model: stubModel{}})
	if err != nil {
		t.Fatalf("llmagent.New: %v", err)
	}
	r, err := adkrunner.New(adkrunner.Config{
		AppName:           "app",
		Agent:             root,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}

	ctx, id := StartInvocation(context.Background())
	var invocationID string
	for evt, err := range r.Run(ctx, "u-1", "s-1",
		genai.NewContentFromText("hello", genai.RoleUser), agent.RunConfig{}) {
		if err != nil {
			t.Fatalf("runner.Run: %v", err)
		}
		if evt != nil && evt.InvocationID != "" {
			invocationID = evt.InvocationID
		}
	}
	if invocationID != id {
		t.Fatalf("run reported invocation %q, pinned %q", invocationID, id)
	}

	want := TraceIDForInvocation(id)
	spans := exporter.spans()
	if len(spans) == 0 {
		t.Fatal("no spans exported")
	}
	var rootNames []string
	for _, span := range spans {
		if got := span.SpanContext().TraceID(); got != want {
			t.Fatalf("span %q trace ID = %s, want %s (invocation %s)", span.Name(), got, want, id)
		}
		if span.Parent().IsValid() {
			continue
		}
		rootNames = append(rootNames, span.Name())
	}
	// The invocation reports one root span (invoke_workflow, or invoke_agent
	// when ADK treats the root agent as the workflow), and everything else hangs
	// off it.
	if len(rootNames) != 1 {
		t.Fatalf("root spans = %v, want exactly one", rootNames)
	}
	if !strings.HasPrefix(rootNames[0], "invoke_") {
		t.Fatalf("root span %q is not an invocation span", rootNames[0])
	}
}

// installTracerProvider routes the spans of the test through exporter with the
// package's ID generator installed.
//
// The global provider is installed once per process and never restored: OTel's
// global is a delegating proxy, and putting that proxy back into the global slot
// after a test leaves the next one exporting into a shut-down provider.
func installTracerProvider(t *testing.T, exporter sdktrace.SpanExporter) {
	t.Helper()
	once.Do(func() {
		otel.SetTracerProvider(sdktrace.NewTracerProvider(
			sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
			sdktrace.WithIDGenerator(NewIDGenerator()),
		))
	})
}

var once sync.Once

func newSession(t *testing.T, events ...*session.Event) session.Session {
	t.Helper()
	svc := session.InMemoryService()
	resp, err := svc.Create(context.Background(), &session.CreateRequest{
		AppName:   "app",
		UserID:    "u-1",
		SessionID: "s-1",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	for _, ev := range events {
		if err := svc.AppendEvent(context.Background(), resp.Session, ev); err != nil {
			t.Fatalf("append event: %v", err)
		}
	}
	return resp.Session
}

func eventWithCall(invocationID, callID string) *session.Event {
	ev := session.NewEvent(context.Background(), invocationID)
	ev.Author = "root-agent"
	ev.Content = &genai.Content{
		Role: genai.RoleModel,
		Parts: []*genai.Part{{
			FunctionCall: &genai.FunctionCall{ID: callID, Name: "adk_request_input"},
		}},
	}
	return ev
}

func answer(callID string) *genai.Part {
	return &genai.Part{
		FunctionResponse: &genai.FunctionResponse{ID: callID, Name: "adk_request_input"},
	}
}

type stubModel struct{}

func (stubModel) Name() string { return "stub-model" }

func (stubModel) GenerateContent(context.Context, *model.LLMRequest, bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(&model.LLMResponse{
			Content:      genai.NewContentFromText("hi there", genai.RoleModel),
			FinishReason: genai.FinishReasonStop,
		}, nil)
	}
}

// spanRecorder is a span exporter that keeps what it exported.
type spanRecorder struct {
	mu    sync.Mutex
	saved []sdktrace.ReadOnlySpan
}

func (r *spanRecorder) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saved = append(r.saved, spans...)
	return nil
}

func (r *spanRecorder) Shutdown(context.Context) error { return nil }

func (r *spanRecorder) spans() []sdktrace.ReadOnlySpan {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]sdktrace.ReadOnlySpan(nil), r.saved...)
}
