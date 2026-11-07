// Package tracing ties an ADK invocation to one OpenTelemetry trace.
//
// By default the OTel SDK gives every root span a random trace ID, and ADK sees
// each HTTP turn as its own invocation. Two things in Butter depend on the
// mapping instead:
//
//   - SessionEvent.trace_url is a Langfuse deep link built from the event's
//     invocation ID, so it addresses a trace only when the trace ID is derived
//     from that invocation ID.
//   - A workflow that pauses on a Human Input node resumes in a later HTTP
//     request. ADK reuses the paused event's invocation ID for that
//     continuation (runner.resolveInvocationID), so a trace ID derived from the
//     invocation ID keeps one logical invocation in one Langfuse trace instead
//     of splitting it across two.
//
// The trace ID is decided when the first span starts, which is before ADK
// creates the invocation context that carries the invocation ID. So Butter
// pins the invocation ID on the run's context instead of learning it later:
// StartInvocation mints one for a fresh run and hands it to ADK through
// platform.WithUUIDProvider, and WithInvocationID pins the ID of a run that
// continues an existing invocation. Either way the invocation ID is known
// before the run and both the trace ID and the session events agree on it.
package tracing

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"strings"
	"sync"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/adk/v2/platform"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

// invocationPrefix is what ADK puts in front of a minted invocation ID
// (adk/v2/internal/context: "e-" + platform.NewUUID).
const invocationPrefix = "e-"

// hexTraceIDLength is the number of hex characters in a 16-byte trace ID.
const hexTraceIDLength = 32

// hashDomain tags the hash input for invocation IDs that are not UUID-shaped,
// so the fallback cannot collide with the direct UUID mapping.
const hashDomain = "butter-trace-id:"

// invocationKey is the context key under which the pinned invocation ID is
// stored.
type invocationKey struct{}

// WithInvocationID pins the invocation ID of the run ctx belongs to.
//
// Use it for a run that continues an existing invocation — one whose ADK runner
// will reuse that invocation's ID — so its spans land in the same trace. A run
// that starts a new invocation wants StartInvocation instead.
func WithInvocationID(ctx context.Context, invocationID string) context.Context {
	id := strings.TrimSpace(invocationID)
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, invocationKey{}, id)
}

// InvocationIDFromContext returns the invocation ID pinned with
// WithInvocationID or StartInvocation, or "" when ctx carries none.
func InvocationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(invocationKey{}).(string)
	return id
}

// StartInvocation mints the invocation ID of a fresh invocation, pins it on the
// returned context, and makes ADK mint the invocation with that same ID.
//
// Every ID ADK generates for the run comes from platform.NewUUID, which reads
// its provider from the context, so the provider is the only way to name the
// invocation before it exists. The invocation ID is the first UUID ADK asks for
// in a run (it is created up front in Runner.run, before any event, function
// call or tool call ID), so the provider hands the minted ID to that first call
// and leaves every later ID random — event IDs and function call IDs must stay
// unique.
func StartInvocation(ctx context.Context) (context.Context, string) {
	id := invocationPrefix + uuid.NewString()
	provider := &seededUUIDProvider{seed: strings.TrimPrefix(id, invocationPrefix)}
	ctx = WithInvocationID(ctx, id)
	ctx = platform.WithUUIDProvider(ctx, provider.next)
	return ctx, id
}

// InvocationIDForMessage reports the invocation ID the ADK runner will reuse
// for the run of this message, or "" when the message starts a new invocation.
//
// It mirrors the runner's own continuation rule (adk/v2/runner.resolveInvocationID):
// a message that answers a FunctionCall already in the session — the implicit
// resume of a paused workflow, or a precisely addressed tool reply — continues
// the invocation that emitted that call.
func InvocationIDForMessage(sess session.Session, parts []*genai.Part) string {
	if sess == nil {
		return ""
	}
	events := sess.Events()
	for _, part := range parts {
		if part == nil || part.FunctionResponse == nil {
			continue
		}
		id := part.FunctionResponse.ID
		if id == "" {
			continue
		}
		// Reverse scan: the most recent event carrying the call is the one the
		// runner continues, and a call ID is only unique per session.
		for i := events.Len() - 1; i >= 0; i-- {
			ev := events.At(i)
			if ev == nil || ev.Content == nil || ev.InvocationID == "" {
				continue
			}
			for _, p := range ev.Content.Parts {
				if p != nil && p.FunctionCall != nil && p.FunctionCall.ID == id {
					return ev.InvocationID
				}
			}
		}
	}
	return ""
}

// TraceIDForInvocation maps an invocation ID onto the trace ID that names it.
//
// ADK mints "e-<uuid>": the UUID's 16 bytes are the trace ID, so the common case
// is the UUID with its prefix and dashes removed. Any other invocation ID (an
// older record, a caller-supplied one) is hashed, so the mapping stays stable
// across processes. An empty invocation ID has no trace and returns the invalid
// zero trace ID.
func TraceIDForInvocation(invocationID string) trace.TraceID {
	id := strings.TrimSpace(invocationID)
	if id == "" {
		return trace.TraceID{}
	}
	if compact, ok := compactUUID(id); ok {
		if tid, err := trace.TraceIDFromHex(compact); err == nil && tid.IsValid() {
			return tid
		}
	}
	sum := sha256.Sum256([]byte(hashDomain + id))
	var tid trace.TraceID
	copy(tid[:], sum[:len(tid)])
	if !tid.IsValid() {
		tid[len(tid)-1] = 1
	}
	return tid
}

// TraceIDString renders the trace ID of an invocation the way Langfuse shows it
// in its URLs: 32 lowercase hex characters. It returns "" for an empty
// invocation ID so callers can skip building a link.
func TraceIDString(invocationID string) string {
	if strings.TrimSpace(invocationID) == "" {
		return ""
	}
	return TraceIDForInvocation(invocationID).String()
}

// IDGenerator is the [go.opentelemetry.io/otel/sdk/trace.IDGenerator] that
// derives the trace ID of an invocation from the invocation ID pinned on the
// context and keeps span IDs random.
//
// Install it through the Langfuse plugin, which builds the trace provider ADK
// reports through:
//
//	cfg.Langfuse.TracerProviderOptions = []sdktrace.TracerProviderOption{
//	    sdktrace.WithIDGenerator(tracing.NewIDGenerator()),
//	}
//
// Span IDs stay random because a resumed workflow reuses the trace: a fixed span
// ID would collide with the spans the execution that paused already reported.
type IDGenerator struct{}

// NewIDGenerator returns the generator to install on the trace provider.
func NewIDGenerator() IDGenerator { return IDGenerator{} }

// NewIDs returns the trace ID named after the invocation pinned on ctx and a
// random span ID. A context that carries no invocation ID — a span started
// outside a Butter run — falls back to a random trace ID, like the SDK default.
func (IDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	if id := invocationIDFor(ctx); id != "" {
		return TraceIDForInvocation(id), randomSpanID()
	}
	return randomTraceID(), randomSpanID()
}

// NewSpanID returns a random span ID, in this trace and in any other.
func (IDGenerator) NewSpanID(context.Context, trace.TraceID) trace.SpanID {
	return randomSpanID()
}

// invocationIDFor reads the invocation ID off ctx: the one Butter pinned, or —
// for a span ADK starts from the invocation context itself — the ID that
// context reports.
func invocationIDFor(ctx context.Context) string {
	if id := InvocationIDFromContext(ctx); id != "" {
		return id
	}
	if carrier, ok := ctx.(interface{ InvocationID() string }); ok {
		return strings.TrimSpace(carrier.InvocationID())
	}
	return ""
}

// seededUUIDProvider hands seed to the first caller and a random UUID to every
// caller after it: the ADK runner asks for the invocation ID first, and every
// later ID it mints (events, function calls) has to stay unique.
type seededUUIDProvider struct {
	seed string

	mu   sync.Mutex
	used bool
}

func (p *seededUUIDProvider) next() string {
	p.mu.Lock()
	if !p.used {
		p.used = true
		p.mu.Unlock()
		return p.seed
	}
	p.mu.Unlock()
	return uuid.NewString()
}

// compactUUID strips the ADK prefix and the dashes of a UUID-shaped invocation
// ID, reporting whether what is left is exactly a trace ID worth of hex.
func compactUUID(invocationID string) (string, bool) {
	var b strings.Builder
	b.Grow(hexTraceIDLength)
	for _, r := range strings.TrimPrefix(invocationID, invocationPrefix) {
		if r == '-' {
			continue
		}
		if !isHexDigit(r) {
			return "", false
		}
		b.WriteRune(r)
	}
	if b.Len() != hexTraceIDLength {
		return "", false
	}
	return strings.ToLower(b.String()), true
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// randomTraceID returns a random, valid (non-zero) trace ID.
func randomTraceID() trace.TraceID {
	for {
		var tid trace.TraceID
		if _, err := rand.Read(tid[:]); err != nil {
			// crypto/rand does not fail on the platforms Butter runs on; a
			// zero trace ID is better than a panic in an ID generator.
			return trace.TraceID{}
		}
		if tid.IsValid() {
			return tid
		}
	}
}

// randomSpanID returns a random, valid (non-zero) span ID.
func randomSpanID() trace.SpanID {
	for {
		var sid trace.SpanID
		if _, err := rand.Read(sid[:]); err != nil {
			return trace.SpanID{}
		}
		if sid.IsValid() {
			return sid
		}
	}
}
