# ADR-0014: A2UI result cards and Human Input forms over AG-UI

- Status: Accepted
- Date: 2026-09-30
- Issue: #350 (spec), #351–#355 (slices)
- Builds on: ADR-0002 (Interrupt state derived from session events)

## Context

AG-UI Chat shows text, tool calls, shared state and Workflow Interrupts, but
structured results arrive as Markdown or raw tool JSON, and a Human Input node
offers only a question. A2UI v0.9.1 is a declarative UI format — the agent
sends JSON describing components from a catalog the client ships — and it is
complementary to AG-UI, which stays the transport. Several facts shape how the
two meet in Butter:

- The AG-UI session key is `(caller, "agui-" + threadId)`. Neither the
  workspace nor the agent is part of it, so the same caller reusing a
  `threadId` under another workspace or agent lands on the same session.
- The ADK runner appends every non-partial event to the session before it
  yields it, and a tool's `State().Set` becomes that event's `StateDelta`.
- Pending Interrupts are derived from session events alone (ADR-0002); there
  is no Butter-owned Interrupt store.
- A2UI v0.9.1 is prompt-first: model output must be validated after
  generation, never rendered as is.
- The assistant-ui AG-UI runtime and the AG-UI client assume one resume
  answers every open interrupt; Butter answers by ID and keeps the rest
  pending.

## Decision

1. **Transport.** Each A2UI message is one AG-UI `CUSTOM` event named
   `butter.a2ui` whose value carries the protocol version, a server-assigned
   `(revision, seq)`, the thread/run/message it belongs to, readable fallback
   text, and exactly one complete envelope. It is a Butter extension, not an
   A2UI-standard AG-UI binding. Clients opt in per run with
   `forwardedProps.butterA2UI = {version, catalogs}`; the declaration selects
   the server's built-in catalog `butter-basic-v1` and can never upload
   components or schemas. Without it nothing changes.
2. **Two kinds of Surface, one catalog.** Models create, update and delete
   read-only **Result Cards** through an invocation-scoped `render_ui` tool,
   offered only in negotiated runs to LLM agents. Inputs and buttons exist only
   in **Human Input Forms** the server builds; a model can neither collect
   input nor define an action. Every `render_ui` batch is validated whole
   (catalog, properties, references, lifecycle, limits) before anything is
   written.
3. **State lives in the session, nowhere else.** A card is a JSON-string record
   in a hidden namespace of ADK session state (`butter:a2ui:card:<id>`, deleted
   cards kept as tombstones so revisions stay monotonic), written through the
   tool's state delta. A form's binding — surface ID, submit token, revision,
   the Interrupt ID and the field rules — is frozen into the request-input
   event's `CustomMetadata` when the node pauses. There is no UI collection and
   no second pending store: which forms are open is `interrupt.Pending` joined
   with those bindings. The namespace never enters AG-UI shared state, so a
   client's `state` can neither read nor write it.
4. **Persist, then send.** The sink derives envelopes only from events the
   runner already stored, as the transition from what the client holds to what
   the session now holds. The UI Snapshot
   (`GET /api/agui/:agent_id/threads/:thread_id/ui`) rebuilds the same state
   after a refresh, a dropped stream or a restart, under the thread's session
   lease, without running the agent.
5. **UI Binding.** The AG-UI handler creates a new thread's session carrying
   `{principal, workspace, agent_id, thread_id}`. UI is exposed and accepted
   only when a request matches it; sessions created before A2UI (no binding)
   and reused threadIds in another context stay text-only.
6. **Submission is a resume.** A form is submitted through AG-UI `resume`
   (`status: "resolved"`, payload `{butterForm: {version, surfaceId,
   revision, token, values}}`). Under the lease the handler checks binding,
   token, revision, that the Interrupt is still pending, and the field rules,
   then delivers the configured fields as a JSON object text — the ordinary
   string payload of ADR-0002, not a typed resume. A rejection answers before
   the stream opens, with a machine-readable code, and never falls back to
   answering another Interrupt. For A2UI clients every run reports every
   Interrupt still open in its outcome and marks answered forms with an
   `updateDataModel` on `/status`; other clients keep the original outcome.

## Consequences

- Non-A2UI clients and every other entry point are unchanged; a Human Input
  node with a form appends field instructions to its question, and plain-text
  and FIFO answers keep working. The form's rules bind form submissions only.
- Surfaces cost one session-state write per card change and nothing else;
  deleting a session deletes its UI with it.
- The UI state is only as isolated as the binding: a pre-A2UI session never
  gets UI, even for its owner.
- ADK's workflow engine parses a JSON answer on resume and hands the successor
  the parsed object, so an AGENT successor sees the keys sorted and butter's
  Router matches the object as its JSON text; the stored payload keeps the
  configured order.
- The dashboard sends a form's resume through the assistant-ui runtime's
  "steer away" path and replaces the resume array the runtime and AG-UI client
  require (every open interrupt) with the form's single entry, so the
  submission appears as a readable reply in the conversation.
- A per-agent policy for card generation (disable, prefer) and pre-built card
  templates are left for later (#381). The policy has since been settled: see
  the Card Policy amendment below.

## Amendment: the thread history (#376)

The thread list (#375) made it possible to open any earlier thread, and the UI
Snapshot alone brought back its cards but none of the messages around them.
`GET /api/agui/:agent_id/threads/:thread_id/messages` now rebuilds the
conversation from the same session, under the same binding and lease rules:

- **The format is standard AG-UI `Message[]`**, so any AG-UI client can
  hydrate with it.
  - Each run is one assistant message, with real FunctionCall IDs and tool
    results.
  - What the live stream hides stays hidden.
  - Open Interrupts come back as `RUN_FINISHED` reported them.
  - Tool calls the server would not accept a result for are dropped, so
    restoring never makes a client cancel them.
- **Surfaces stay in the snapshot.** The history only says which reply each
  one belongs to. The dashboard places each surface in that reply and keeps a
  surface it cannot place in the "restored" block. There is still no second
  UI store.
- **An answered Human Input reads as its question and the user's answer.** A
  form's answer is formatted the way the dashboard showed the submission. An
  open one returns to the runtime as a pending Interrupt, so it is answered
  exactly as after a live run.

## Amendment: reads during a run (ADR-0016)

ADR-0016 lets an AG-UI run outlive its request, so a thread is often opened
while a run is still going. It takes the UI Snapshot (decision 4) and the
thread history (the amendment above) off the session lease:

- During a run they answer at once and report the run as `running`.
- They stop where the run began, keeping the turn that started it.
- The run's own Surfaces reach the client through the replay of its Run Log,
  so none of them shows twice.

## Amendment: the Card Policy (#381, #439)

Since #409 the dashboard's Chat negotiates A2UI on every run, so every LLM
agent opened from the dashboard is offered `render_ui`. An author's only
control was the instruction, which guarantees nothing and also reaches entry
points without cards. An Agent's `config.result_cards` now holds its Card
Policy:

- **Generation only narrows.** Unset inherits; `DISABLED` turns Result Cards
  off for the agent and every agent below it in the run, and nothing below can
  turn them back on. There is no `ENABLED`: under narrowing it would mean the
  same as unset. An agent run directly answers only to its own policy; run
  under a parent, to every agent on the path from the run's root.
- **Presentation is a hint.** `AUTO` or `PREFERRED`, inherited down the tree,
  the nearest explicit value winning; a run's root defaults to `AUTO`.
  `PREFERRED` is a sentence in `render_ui`'s description, so it reaches only
  model calls that can render a card and never enters the instruction.
- **The policy is resolved when the runner builds an agent tree.** Each LLM
  agent's `render_ui` toolset is built with the policy of its place in the
  tree. A run keeps the tree it started with, so offering the tool and running
  it read the same value, and a change applies from the next run. A disabled
  agent has no `render_ui`; a call copied from history gets ADK's "tool not
  found" error back and the run goes on. Resolving per run in the AG-UI
  handler was rejected: it would walk the tree a second time, outside the
  factory, keyed by agent name.
- **Forms and existing cards are outside it.** Human Input Forms, the UI
  Snapshot and the cards already in a thread are unchanged; a disabled agent
  only stops changing cards. PI and CURSOR agents reject the field on write;
  composite agents accept it to narrow their subtree.

Pre-built card templates (#381) stay deferred.
