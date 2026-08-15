# AG-UI protocol integration

Research date: 2026-08-14 — verifying the proposal in [issue #286].

Every claim below was checked against the real dependencies at the versions
Butter would resolve, and the load-bearing ones were proved with a throwaway
compile-and-run spike (see [Spike](#spike-what-was-actually-run)).

## Conclusion

The proposal is **sound and the Phase 1 shape is right**, but three of its
statements do not survive contact with the code:

1. **ADK Go has no AG-UI support.** The issue says "AG-UI has been adopted by
   Google ADK". That is true of adk-python and of a *community* Go bridge — it
   is not true of `google.golang.org/adk/v2 v2.1.0`, which contains no `agui`
   package and no reference to the protocol. Butter writes this adapter itself.
2. **The AG-UI Go SDK is untagged.** It has no released versions at all; the
   only way to depend on it is a pseudo-version off `main`. This is a real
   supply-chain decision, not a footnote.
3. **`STEP_STARTED` / `STEP_FINISHED` has no source event.** The mapping table
   lists "Workflow node transition" as if ADK emits one. It does not — node
   boundaries exist only as OpenTelemetry spans. Any STEP events must be
   *synthesized*, and the honest signal is a change of `session.Event.Author`.

Against that, the two hardest-looking rows in the mapping table turn out to be
*easier* than the issue assumes: the interrupt outcome and the resume path both
fall out of seams Butter already has, with **no change to `streamorch`, the
runner, or the session layer**. The "purely additive" claim holds.

## Verified: dependency situation

`github.com/ag-ui-protocol/ag-ui/sdks/community/go` resolves to
`v0.0.0-20260813165816-6691ac50b34a`. The module proxy lists **zero tagged
versions**:

```
$ curl -s https://proxy.golang.org/github.com/ag-ui-protocol/ag-ui/sdks/community/go/@v/list
(empty)
```

Adding it to Butter is nonetheless nearly free. `go get` alone touches 3 lines,
and once the package is actually imported `go mod tidy` adds exactly one
transitive dependency:

```
+ github.com/ag-ui-protocol/ag-ui/sdks/community/go v0.0.0-20260813165816-6691ac50b34a
+ github.com/sirupsen/logrus v1.9.3 // indirect
```

`go build ./...` passes across the whole repo. Its other requirements (`uuid`,
`testify`) were already satisfied, and its `go 1.24.4` directive is below
Butter's `go 1.26.5`.

Note for whoever reviews the dependency diff: the same `go mod tidy` also
**removes** `github.com/bwmarrin/discordgo` and `github.com/go-telegram/bot`.
That is unrelated pre-existing drift — both became unused in the #273 legacy
channel cutover and no Go file imports them.

**Recommendation:** depend on the SDK **pinned to an exact pseudo-version**,
and treat upgrades as deliberate. Do not use `@latest` in CI.

### The community ADK bridge is not usable as-is

`github.com/ieshan/adk-go-pkg` (v1.0.16, 2026-07-23) does have an `agui`
package with a handler, emitter, and MCP middleware. But its `go.mod` pins
`google.golang.org/adk/v2 v2.0.0` — a minor version behind Butter — and an
*older* AG-UI SDK pseudo-version (`20260722`) than current. MVS would drag
Butter's whole ADK surface into a version pair nobody tests together.

**Recommendation:** read it as a reference implementation; do not depend on it.
The adapter Butter needs is ~150 lines (the spike is proof), which is cheaper
than owning that compatibility risk.

## Verified: the SDK has the types the design needs

All present in `pkg/core/types` and `pkg/core/events`:

- `RunAgentInput{ThreadID, RunID, State, Messages, Tools, Context, ForwardedProps, Resume}`
  — with `UnmarshalJSON` accepting **both** camelCase and snake_case, so it
  coexists with Butter's snake_case ConnectRPC codec without special handling.
- `types.Interrupt{ID, Reason, Message, ToolCallID, ResponseSchema, ExpiresAt, Metadata}`
- `types.ResumeEntry{InterruptID, Status, Payload}` with
  `ResumeStatusResolved` / `ResumeStatusCancelled`
- `RunFinishedEvent.Outcome` with `RunFinishedOutcomeTypeSuccess` /
  `RunFinishedOutcomeTypeInterrupt`, built via `WithSuccessOutcome()` /
  `WithInterruptOutcome([]types.Interrupt)`

What it does **not** have is a server: there is no HTTP handler, no
`RunAgentInput` routing, no run lifecycle. `pkg/client/sse` is a *client*.
Server-side, the SDK gives exactly two things — event structs and
`pkg/encoding/sse.SSEWriter.WriteEvent(ctx, io.Writer, Event)`. Everything
else in `internal/handler/http/agui.go` is Butter's own code.

One sharp edge: `SSEWriter.WriteEvent` calls `ctx.Err()` unguarded, so a `nil`
context **panics** rather than returning an error. Always pass
`c.Request.Context()`.

## Verified: the interrupt outcome needs no plumbing change

This is the row the issue under-specifies, and the finding is good news.

`streamorch.Sink` has four methods — `Started`, `TextDelta`, `RunEvent`,
`Final(id, response string)`. None carries `runner.TurnResult`, so `Pending`
interrupts appear to be out of reach: `streamorch.Run` calls `RunSSE`, which
returns only `(string, error)`.

But the pause is visible **in-stream**. `session.Event` (ADK v2.1.0) carries:

```go
// RequestedInput, when non-nil, signals that the workflow node
// emitting this event is asking for human input and is about to pause.
RequestedInput *RequestInput
```

and `RequestInput{InterruptID, Message, ResponseSchema, Payload}` maps almost
field-for-field onto `types.Interrupt`. The runner deliberately forwards these
events to the callback even though they terminate the agent loop
(`internal/runtime/runner/runner.go`, `run`'s event loop):

```go
// Request-input events count as final responses (they end the agent
// loop) but must still reach the callback: the dashboard chat stream
// renders them as the workflow's question instead of swallowing them.
if onEvent != nil && (!evt.IsFinalResponse() || evt.RequestedInput != nil) {
```

And `streamorch.Run` routes them to `Sink.RunEvent`: a request-input event is
not `Partial`, so `textParts` returns nothing and the `hasOnlyTextParts`
short-circuit cannot fire.

**So an `AGUISink` can collect interrupts during `RunEvent` and emit
`RUN_FINISHED{outcome:interrupt}` from `Final`, with no change to `Sink`,
`streamorch.Run`, or the runner.** Proved by `TestAGUI_InterruptOutcome`.

Two details the implementation must get right:

- **Filter the request-input `FunctionCall`.** The pause event *also* carries a
  `FunctionCall` named `workflow.WorkflowInputFunctionCallName`. Mapped
  naively it becomes a phantom `TOOL_CALL_START/ARGS/END` for a tool the user
  never called. Skip it in both the call and response directions.
- **The question is delivered twice.** `runner.run` appends each pending
  question to `turn.Output`, which becomes `Sink.Final`'s `response`. So the
  same text arrives once as `Interrupt.Message` and once as assistant message
  content. Decide deliberately which one the client renders.

## Verified: AG-UI's precise resume beats Butter's FIFO — for free

AG-UI resumes are addressed: `ResumeEntry` names an `interruptId`. Butter's
implicit resume is FIFO — plain text answers the *oldest* pending Interrupt
(ADR-0002). These look incompatible.

They are not. `interrupt.Resume` returns early, unchanged, when the inbound
parts already contain a `FunctionResponse`:

```go
for _, p := range parts {
    if p != nil && p.FunctionResponse != nil {
        return parts, false
    }
}
```

So the handler translates each `ResumeEntry` into

```go
&genai.Part{FunctionResponse: &genai.FunctionResponse{
    ID:       entry.InterruptID,
    Name:     workflow.WorkflowInputFunctionCallName,
    Response: map[string]any{"payload": entry.Payload},
}}
```

and the client's addressing wins over FIFO. Proved by
`TestAGUIResumeEntryBypassesFIFO`, which also pins the contrast: a plain-text
reply against the same two-interrupt session is forced onto the older one.

**Gap:** `ResumeStatusCancelled` has no Butter equivalent. There is no way to
abandon a pending Interrupt — the workflow stays paused forever. Phase 1
should reject `status: "cancelled"` with an explicit error rather than
silently treating it as resolved.

## Verified: STEP events must be synthesized

`google.golang.org/adk/v2/workflow` has no per-node session event. Node
boundaries exist only as tracing spans (`workflow/node_span.go`, which wraps
OpenTelemetry and never touches the session). Node identity reaches the sink
only through `session.Event.Author` and `Event.Branch`.

The workable approach is to emit `STEP_FINISHED`/`STEP_STARTED` on an `Author`
change and close the open step in `Final`. The spike does this and produces a
well-formed stream, but be clear about what it is: a derived approximation.
`Author` also changes for sub-agent transfers in a non-workflow agent, so
plain LLM agents will emit STEP events too.

**Recommendation:** ship Phase 1 without STEP events, or behind a flag. They
are the one row in the mapping table with no ground truth behind it, and a
wrong step boundary is worse for a client than no step boundary.

## Verified: routing and auth

`internal/app/routes.go` installs `AuthMiddleware` on the engine before any
handler registers, and it gates on an `isPublicPath` allowlist rather than an
opt-in list. A new `POST /api/agui/:agent_id` therefore inherits bearer auth
and `X-Workspace-ID` workspace resolution with no middleware work — the issue's
claim holds.

Three notes on the handler's own shape:

- **Follow `openai.go`, not `a2a.go`, for agent lookup.** The OpenAI adapter
  resolves by immutable `agent_id` via `configrepo.AgentRepository.GetAgent` and
  gates on an explicit per-agent opt-in (`enable_openai_api`). AG-UI mirrors that
  with its own `enable_agui` flag rather than exposing every agent in the
  workspace.
- **`streamorch` vs. direct runner.** `openai.go` bypasses `streamorch` and
  calls the runner directly; `agent_stream.go` and `asyncrun` go through it.
  AG-UI should go through `streamorch` — it is the seam that already
  normalizes partial-text vs. mixed events, and reusing it is what keeps the
  interrupt finding above true.
- **The sink belongs with the transport, not in `streamorch`.** Issue #286
  originally placed `AGUISink` in `internal/runtime/streamorch/`. Both existing
  `Sink` implementations live with their own transport (`streamAgentSink` in
  `internal/application/`, `hubSink` in `asyncrun/`), and `streamorch` has no
  wire-protocol dependency of its own. The shipped code follows that convention:
  `internal/handler/http/agui_sink.go`.

## Spike: what was actually run

Before any of this shipped, a `git worktree` off `8d7c989` with the AG-UI SDK
added, containing:

- an `AGUISink` implementing `streamorch.Sink`, ~150 lines
- a test driving it through the **real** `streamorch.Run`
- a test of the resume path against the **real** `interrupt.Resume`

All pass:

| Test | Asserts |
|---|---|
| `TestAGUI_TextRun` | `RUN_STARTED → TEXT_MESSAGE_START → CONTENT → END → RUN_FINISHED` |
| `TestAGUI_ToolCall` | `TOOL_CALL_START/ARGS/END` then `TOOL_CALL_RESULT` |
| `TestAGUI_InterruptOutcome` | `RUN_FINISHED.Outcome.Type == interrupt`, interrupt id/message preserved, **no** phantom `TOOL_CALL_*` |
| `TestAGUI_RunError` | `RUN_STARTED → RUN_ERROR` |
| `TestAGUI_SSEEncoding` | real SSE frames on the wire |
| `TestAGUIResumeEntryBypassesFIFO` | addressed resume passes through `interrupt.Resume` untouched; plain text still goes FIFO |

Actual wire output from `TestAGUI_SSEEncoding`:

```
id: RUN_STARTED_1786641062238
data: {"type":"RUN_STARTED","timestamp":1786641062238,"threadId":"thread-1","runId":"run-1"}

id: TEXT_MESSAGE_START_1786641062238
data: {"type":"TEXT_MESSAGE_START","timestamp":1786641062238,"messageId":"msg-1","role":"assistant"}

id: TEXT_MESSAGE_CONTENT_1786641062238
data: {"type":"TEXT_MESSAGE_CONTENT","timestamp":1786641062238,"messageId":"msg-1","delta":"hi"}

id: TEXT_MESSAGE_END_1786641062238
data: {"type":"TEXT_MESSAGE_END","timestamp":1786641062238,"messageId":"msg-1"}

id: RUN_FINISHED_1786641062238
data: {"type":"RUN_FINISHED","timestamp":1786641062238,"threadId":"thread-1","runId":"run-1","outcome":{"type":"success"}}
```

The worktree was discarded. The shipped implementation carries the same
assertions, plus cases the spike did not cover (thought-part filtering, closing
an open message on error, non-streaming responses, and the handler's rejection
table) — see `internal/handler/http/agui_sink_test.go` and `agui_test.go`.

## Corrected mapping table

| ADK signal | AG-UI event | Status |
|---|---|---|
| Run start | `RUN_STARTED` | ✅ `Sink.Started` |
| Partial text part | `TEXT_MESSAGE_START` / `_CONTENT` | ✅ `Sink.TextDelta`, lazily opening the message |
| `Part.FunctionCall` | `TOOL_CALL_START` + `_ARGS` + `_END` | ✅ — **must** skip `adk_request_input` |
| `Part.FunctionResponse` | `TOOL_CALL_RESULT` | ✅ — same exclusion |
| Run completion | `TEXT_MESSAGE_END` | ✅ `Sink.Final` (also emits `response` when nothing streamed) |
| `Event.RequestedInput` | `RUN_FINISHED{outcome:interrupt}` | ✅ in-stream via `Sink.RunEvent`; no plumbing change |
| Normal completion | `RUN_FINISHED{outcome:success}` | ✅ `Sink.Final` |
| Run error | `RUN_ERROR` | ⚠️ `Sink` has no error frame — handler emits it from `streamorch.Run`'s return |
| ~~Workflow node transition~~ | `STEP_STARTED` / `STEP_FINISHED` | ❌ no source event; synthesizable from `Event.Author` only |
| `ResumeEntry` → `FunctionResponse` | (inbound) | ✅ passes through `interrupt.Resume` untouched |
| `ResumeStatus: cancelled` | (inbound) | ❌ unsupported — reject explicitly |

## Phase 1 decisions taken from this research

1. **`threadId` is namespaced, not used directly.** The issue mapped it straight
   to a session ID, which would put a client-chosen string into the namespace
   shared with every other adapter (`chat-`, `openai-`, `tg:…`). The session is
   `agui-{threadId}`, and because the ADK session key also includes the user ID
   (derived from the authenticated caller), two users may reuse one `threadId`
   without sharing history.
2. **Only the trailing user message is forwarded.** `RunAgentInput.Messages` is
   the client's full history but the server-side session is authoritative;
   replaying it would duplicate the conversation.
3. **Unimplemented capabilities are rejected, not ignored** — non-empty `tools`
   (Phase 2), non-empty `state` (Phase 3), and `resume[].status: "cancelled"`.
   An empty `state` object counts as "none" and is accepted.
4. **No `STEP_*` events.** See above — there is no source event to derive an
   honest boundary from.

## Known limitation: per-session serialization is in-process only *(resolved)*

`runner.acquireSessionTurn` serializes turns per `appName+userID+sessionID`, but
it is an in-process lock. Butter runs multiple Pods, so two concurrent requests
carrying the same `threadId` that land on different Pods were **not** serialized
and their turns could interleave.

**Resolved by issue #291 item 1**: the AG-UI handler now takes a bounded,
renewable Redis lease per `(caller, threadId)` before opening the stream
(`internal/runtime/sessionguard`, on the `internal/redislease` primitive shared
with the Telegram pipeline). A busy thread is a pre-stream `409`; losing the
lease mid-run cancels the run and surfaces as `RUN_ERROR`.

## Follow-up: issue #291 (2026-08-15)

The Phase 1 rejections were lifted one contract at a time:

- **Frontend tools** (`tools`) map onto ADK long-running function tools: a
  toolset attached to every LLM agent resolves the request's declarations from
  the run context per invocation (`internal/aguitool`), a call ends the run
  after `TOOL_CALL_*` (the FunctionCall event is final via
  `Event.LongRunningToolIDs`, which the runner's callback gate now forwards),
  and the client's trailing tool-role message resumes ADK by `FunctionCall.ID`
  (`runner/run_node.go`'s `openLongRunningCallIDs` path). Pending calls are
  derived from session events (`interrupt.PendingToolCalls`, ADR-0002 style)
  to validate results before anything runs.
- **Shared state** (`state`) is server-authoritative: pre-run session state is
  validated against the client mirror (corrective `STATE_SNAPSHOT` after
  `RUN_STARTED`), in-run `EventActions.StateDelta` streams as RFC 6902
  `STATE_DELTA`, and a post-run session re-read emits the trailing delta for
  `output_key` writes the runner never streams. `app:`/`user:`/`temp:` scopes
  stay server-side.
- **Dashboard client**: `front/src/api/agui.ts` (fetch + ReadableStream SSE
  parser; no AG-UI npm dependency) and `front/src/features/agui-chat` render
  the stream — text, tool calls, shared state, `RUN_ERROR`, and Interrupts
  with addressed resume — beside the untouched classic chat.

`resume[].status: "cancelled"` remains rejected: Butter still has no way to
abandon a pending Interrupt or tool call.

## References

- AG-UI spec: <https://docs.ag-ui.com>
- AG-UI Go SDK: <https://github.com/ag-ui-protocol/ag-ui/tree/main/sdks/community/go>
- Community ADK-Go bridge: <https://github.com/ieshan/adk-go-pkg>
- ADR-0002 (interrupts derived from session events): `docs/adr/0002-interrupt-state-derived-from-session-events.md`

[issue #286]: https://github.com/orvice/butter/issues/286
