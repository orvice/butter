# ADR-0018: One Langfuse trace per invocation

- Status: Accepted
- Date: 2026-11-08
- Issue: KON-111 (Langfuse integration version check)
- Builds on: ADR-0002 (Interrupt state derived from session events), ADR-0016
  (AG-UI runs outlive their request)

## Context

Langfuse tracing comes from `adk-utils-go`'s `plugin/langfuse`, which wires ADK
telemetry to an OTLP/HTTP exporter. Butter surfaces a link to it on every session
event: `SessionEvent.trace_url` is `<langfuse host>/trace/<id>`, rendered by the
Session detail page.

That `<id>` used to be the event's `invocation_id`, and the trace that the
invocation's spans ended up in had an unrelated, randomly generated ID. Both
halves were wrong for the same reason: nothing tied the two together.

- **The link addresses nothing.** Langfuse resolves `/trace/<id>` by OTel trace
  ID. An invocation ID is a UUID ADK minted for its own bookkeeping, and no
  trace carries it, so the link 404s for every event.
- **One logical run can become several traces.** A workflow that pauses on a
  Human Input node ends its HTTP turn and resumes in a later request (ADR-0002,
  ADR-0016). ADK reuses the paused event's invocation ID for the continuation
  (`runner.resolveInvocationID`), but a random trace ID per execution meant the
  question and its answer appeared as two unrelated traces, in two different
  requests, possibly on two different Pods with no shared parent.

The gap was latent rather than visible: `config.yaml` ships no Langfuse keys, so
the integration is disabled unless an operator turns it on, and the OTel
provider was never installed at all when it is.

## Decision

The trace ID of an invocation is derived from the invocation ID, and Butter
pins the invocation ID before the run starts.

1. **Derivation** (`internal/runtime/tracing.TraceIDForInvocation`). ADK mints
   `e-<uuid>`; the UUID's 16 bytes are the trace ID, so mapping is
   "strip the prefix and the dashes". Anything that is not UUID-shaped (an older
   record, a caller-supplied ID) is SHA-256 hashed with a domain tag, so the
   mapping is stable across processes and can never collide with the direct
   case. An empty invocation ID maps to no trace.
2. **Pinning** (`tracing.StartInvocation`). The trace ID is decided when the
   first span starts, which is before ADK creates the invocation context that
   carries the invocation ID. So Butter names the invocation itself: it mints
   `e-<uuid>` and installs it on the run's context through
   `platform.WithUUIDProvider`, the seam ADK documents for hosts that supply
   their own identifiers. ADK turns the `e-` prefix into the invocation ID it
   uses, and the same value seeds the trace (the ADK runner mints the invocation
   ID before any event, function call or tool call ID, so it is the first UUID
   of a run; the provider hands the seed to that first call and leaves every
   later ID random).
3. **Continuation** (`tracing.InvocationIDForMessage`). When the message answers
   a `FunctionCall` already in the session — the implicit resume of a paused
   workflow, or a precisely addressed tool reply — Butter pins that call's
   invocation ID instead, mirroring the runner's own continuation rule. The
   resumed leg lands in the trace the pause started.
4. **The generator** (`tracing.NewIDGenerator`) is what the SDK calls: it returns
   the derived trace ID for a context that carries a pinned invocation, a random
   one otherwise, and always a random span ID (a resumed execution reuses the
   trace, so a fixed span ID would collide with the spans the paused execution
   already reported). It is installed through the Langfuse plugin's
   `TracerProviderOptions`.
5. **The API** reports the truth: `SessionEvent.trace_id` is the derived trace ID
   and `trace_url` is built from it.

If ADK ever heads an invocation with an ID other than the pinned one, the
runner logs a warning the first time it sees the event, rather than silently
writing a link that points nowhere.

## Considered Options

- **Keep the invocation ID as the link ID and pass the invocation ID into the
  span attributes.** Langfuse resolves deep links by trace ID, so the URL would
  still have to be built from the trace ID; an attribute solves nothing.
- **Carry Butter's own run ID in the session event and map it later.** The
  events are the only source for the URL (ADR-0002 keeps no Butter-owned
  pending state), so a second ID would need its own persisted mapping, and the
  trace still would not be derivable from what the event carries.
- **One trace per session.** Cleaner for reading a long conversation in one
  place, but a session never ends, so the trace would never end either: Langfuse
  would hold one unbounded trace per session, and a run's spans would be
  attributed to no run boundary.
- **Pin the trace ID only for the first leg and accept a second trace on
  resume.** Half the value, and the split trace is exactly what makes a paused
  workflow hard to read.

## Consequences

- **A dependency on ADK's minting order.** The seeded provider is only correct
  while the invocation ID is the first UUID the runner asks for. This is
  internal to ADK, so it is guarded by the runner's mismatch warning and by
  tests that run a real ADK runner and assert the reported invocation ID and the
  exported spans agree.
- **Spans outside a Butter run stay random.** A context with no pinned
  invocation keeps the SDK default, so tracing from other entry points (tests,
  probes) behaves as before.
- **One trace per invocation, not per run.** A resumed workflow is one
  invocation and therefore one trace with two root spans: the pause's execution
  and the continuation's. That is the shape Langfuse already renders for
  retries, and it is what makes the question and its answer one unit.
- **Traces only exist when Langfuse is configured.** The provider is installed
  by `langfuse.Setup`, and the ID generator rides along with it.
- **The invocation ID is Butter's, not ADK's.** It keeps the `e-<uuid>` shape
  ADK would have produced, so nothing downstream can tell the difference, but
  any code that assumed ADK minted it inside the run is now wrong.
