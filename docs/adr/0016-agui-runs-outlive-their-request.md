# ADR-0016: AG-UI runs outlive their request

- Status: Accepted
- Date: 2026-10-04
- Issue: #389 (PRD), #400 (this decision), #401–#408 (implementation),
  #409–#411 (cutover)
- Builds on: ADR-0002 (Interrupt state derived from session events), ADR-0009
  (the retry boundary), ADR-0014 (A2UI over AG-UI), ADR-0015 §7 (cross-Pod
  stop)
- Amends: ADR-0014 §4 and its thread-history amendment (reads no longer take
  the session lease), ADR-0015 §7 (its stop is no longer scoped to Linear)

## Context

Today an AG-UI run lives inside its HTTP request:

- **Lease**: `RunAgent` takes the thread's Redis lease on the request context
  (`acquireThread`), so a client disconnect cancels the run and releases the
  lease. The first renewal that fails cancels the run as well
  (`sessionguard`).
- **Output**: the sink writes straight to the response, and the session read
  after the run uses the request context.
- **Reads**: the Thread History and the UI Snapshot take the same lease
  (`readThread`), so they answer 409 while a run is going.
- **Record**: the runner records the run's Invocation, with no CANCELLED
  outcome, no maximum duration and no shutdown handling.

AG-UI Chat is becoming the dashboard's only chat, at `/chat` (PRD #389). The
chat it replaces keeps its runs going in the background through `asyncrun`,
which is single-instance by design. Its runs, cancellation and observers live
in one process, a process-local mutex serializes its accept step, and its
design limits it to one replica (`docs/design-dashboard-chat-async.md`,
"Single-Instance Boundary"). Production runs several replicas. So an AG-UI run
has to survive navigation and reloads, be stoppable on purpose, and be
observable again later, on any Pod, from the first release.

Several facts shape the design:

- **Standard clients stop a run by aborting its request.** AG-UI has no stop
  call. `HttpAgent.abortRun()` aborts the fetch, and assistant-ui's Stop and
  unmount both end there. So do AG-UI Chat's thread switch, Agent switch and
  New thread.
- **The stream is stateful from `RUN_STARTED`.** The sink tracks the open
  message, the shared-state mirror and the A2UI card baseline. A late observer
  is only correct with every event since the start. Those events cannot be
  rebuilt from the session: text deltas are partial events, and the runner
  never stores partial events.
- **A run's events reach the session while it runs.** The runner stores every
  non-partial event before it yields it (ADR-0014). A read during a run
  therefore sees part of that run, and next to a replay of the same run that
  part would show twice.
- **assistant-ui re-attaches through its history adapter.** When
  `ThreadHistoryAdapter.load()` returns `unstable_resume`, the runtime starts
  a run under the history head and consumes `adapter.resume()`:
  - `resume()` yields `ChatModelRunResult` snapshots of one assistant
    message, not AG-UI events;
  - on that path the runtime parses no AG-UI events, so it applies no
    `CUSTOM` or `STATE_*` event either, and its `RunAggregator` is not
    exported;
  - returning `unstable_resume` without a `resume()` starts a real run.
- **Parts that exist.**
  - `sessionguard` fences every acquisition with a token of its own.
  - ADR-0015 §7 stops a Linear turn on any Pod with a marker bound to the
    holder's lease token plus a pub/sub nudge.
  - `internal/eventqueue` is a consumer-group hand-off, not a log that any
    number of readers can replay.
  - `asyncrun`'s watch hub gives a late watcher one snapshot and the frames
    after it, drops slow watchers, and works within one process.
- **A thread has one holder.** Since #388 a session ID belongs to one user per
  app. A run on another user's `threadId`, or on the caller's own thread in
  another workspace, is refused with 403 before the stream opens.

## Decision

1. **Detaching is opt-in, per request.**
   - **Extension**: a client asks for a **Detached Run** with
     `forwardedProps.butterRun = {detach: true}`, a Butter extension next to
     `butterA2UI`. A malformed declaration is a 400, as for `butterA2UI`.
   - **Default**: without it the POST behaves as today, and a disconnect
     cancels the run. A standard AG-UI client has no other way to stop one.
   - **Bound threads only**: Stop and attach are checked against the thread's
     UI Binding (decisions 4 and 5), so a run on a thread without one could
     be neither stopped nor followed. Detaching there is refused before the
     stream opens.
   - **Dashboard**: AG-UI Chat asks for it on every run (#405).
   - **Scope**: the decisions below are about Detached Runs, except the run
     state and reads of decision 6 and the delete of decision 7, which cover
     every run.

2. **The run, not the request, holds the thread lease.**
   - **Before the stream**: the lease is still taken before the stream opens.
     A busy thread is still a 409, and a lease outage still a 503.
   - **One acquisition**: every check that reads the session runs under the
     acquisition that the run then keeps. That covers client tool results
     against the pending calls, the shared-state baseline, the thread checks
     of #388 (403) and form resolution (ADR-0014 §6). No other run can slip
     in between, and every rejection is still an HTTP error, before anything
     is detached.
   - **Held by the run**: the lease is taken on a context detached from the
     request and handed to the run's goroutine. The goroutine keeps it until
     the run's terminal state is recorded and its terminal event is in its
     Run Log. The request only observes (decision 5).
   - **Renewal**: a renewal that errors is retried until the lease would
     really have lapsed, one TTL after the last renewal that succeeded. Only
     that lapse, or a renewal answered "not the holder", cancels the run.

3. **A Detached Run owns its Invocation record.**
   - **Owner**: the AG-UI path writes the record, and the runner records
     nothing for the run (`runner.WithoutInvocationRecording`, as `asyncrun`
     does). The record names the AG-UI path as its owner, so other entry
     points can route a Stop to it (decision 4).
   - **States**: QUEUED → RUNNING → SUCCEEDED / FAILED / CANCELLED. CANCELLED
     means a person ended the run, through a Stop or by deleting the thread.
     Every operational end is FAILED, with a reason.
   - **Idempotency**: `request_id` is the client's `runId`, scoped to the
     thread, because the client chooses it. A POST that repeats a `runId` the
     thread already ran is refused, so the Agent never runs twice for one
     request (ADR-0009).
   - **Limits**: a configurable maximum run duration, 30 minutes by default
     like the async chat's, ends a run FAILED. A graceful shutdown ends the
     runs in flight FAILED with a shutdown reason and releases their leases,
     waiting a bounded time.
   - **Terminal precedence**: `asyncrun`'s (`claimTerminal`). An accepted
     Stop wins. Otherwise a run whose runner returned cleanly is SUCCEEDED,
     even when a shutdown or the deadline raced it. After that come the
     shutdown, the deadline and the run's own error. Across Pods, accepting a
     Stop and claiming the terminal state are each one atomic step on the
     same Redis state: an accepted Stop always ends CANCELLED, and a Stop
     that comes after the claim finds nothing running.
   - **Liveness**: the record carries its owner stamp (#390). It is failed as
     stale only once its owner's liveness key is gone, never because another
     Pod started.
   - **After a reload**: the thread history reports a failed or stopped last
     run, with its input, from this record (`lastRun`). The turn stays
     visible, with Restore input (#403, #408).
   - **No rerun**: a FAILED run is never rerun automatically. The way back is
     to send again, which starts a new run (ADR-0009).

4. **Stop reaches the run on any Pod.**
   - **Endpoint**: `POST /api/agui/:agent_id/threads/:thread_id/stop`, with
     the same auth, workspace and binding checks as the thread reads. It is
     idempotent and answers 204 when nothing is running.
   - **Mechanism**: ADR-0015 §7's. The Stop never takes the lease. In one
     step it sets a short-lived marker that holds the running run's lease
     token, and publishes a nudge on the thread's channel. The run subscribes
     before it checks the marker, so a Stop that races its start is not lost.
     It checks the marker again at every renewal, which covers a nudge that
     never arrives. A marker bound to one lease token never reaches a later
     run on the thread.
   - **Outcome**: the run cancels its turn and ends CANCELLED. A Pi or Cursor
     Agent aborts through its single `AbortSession` path. Observers receive
     `RUN_ERROR` with a stop code, not today's "session lease lost".
   - **Other entry points**: `CancelAgentInvocation` on an AG-UI-owned
     Invocation, and deleting the thread (decision 7), go through the same
     Stop. `runner.CancelInvocation` is not used for these runs, because it
     reaches only its own Pod.
   - **Scope**: Stop addresses Detached Runs. A run without the opt-in still
     ends when its request is aborted.
   - **Dashboard**: Stop calls this endpoint, then cancels locally to end the
     stream. Leaving the page, switching thread or Agent, and New thread only
     abort the request, which now just detaches (#405).

5. **Every observer replays the run's log.**
   - **Run Log**: a Detached Run's AG-UI events go to a bounded Redis Stream,
     keyed like the thread lease by caller, thread and run. The log never
     depends on a thread ID having a single holder, and its key comes from
     the authenticated caller, so no one reaches another caller's run.
   - **Writer**: a buffered writer appends the events, so Redis never stalls
     the model loop. Consecutive `TEXT_MESSAGE_CONTENT` deltas are coalesced.
     The log is capped by entries and by bytes. It expires a few minutes
     after its terminal event, and also when its run dies without one.
   - **Observers**: the sink writes to the log only. Each observer replays
     the log from `RUN_STARTED`, under the original `runId`, and follows it to
     the terminal event, with SSE comment heartbeats in between. The POST
     response is the first observer, and
     `GET /api/agui/:agent_id/threads/:thread_id/run` adds more, under the
     same checks as the thread reads. A disconnect detaches only that
     observer, and a live view and a replay see the same sequence.
   - **Self-contained**: a Detached Run always opens with a `STATE_SNAPSHOT`.
     Its A2UI envelopes move the client on from the Surfaces as they stood
     when the run started, which is what the reads of decision 6 return. The
     thread reads and a replay together rebuild the whole thread.
   - **Fallback**: attaching answers 204 when there is no log: no run in
     flight or recently ended, or a run without the opt-in. An observer that
     cannot follow the run from `RUN_STARTED` to its end finishes with a
     Butter `CUSTOM` marker that tells the client to read the thread instead.
     That happens when the log overflowed or expired, or when the run lost
     its lease without a terminal event.
   - **Connections**: each Pod multiplexes its observers' blocking reads, so
     an observer does not hold a Redis connection of its own.
   - **Not a record**: the Run Log is a replay buffer. The session stays the
     record of the conversation (ADR-0002, ADR-0014 §4), so losing a log
     costs only the live view.

6. **Thread reads never wait for a run.**
   - **No lease**: the Thread History and the UI Snapshot no longer take the
     lease. During a run they answer at once with
     `running: {runId, invocationId}`, and they leave out what the run has
     produced so far.
   - **Run state**: when a run starts, detached or not, it records its run
     state next to the lease: its `runId`, its Invocation ID, and the number
     of events the session held before it. The run state is renewed with the
     lease, so it lapses with the lease when a Pod dies. When a Detached Run
     ends, its run state is marked ended and kept as long as its log, so a
     late attach still finds the run.
   - **The cut**: a read keeps the events before that count, plus the turn
     the run started from: the user's message, Human Input answer or tool
     results. That turn is already on the client's screen, assistant-ui
     hangs a resumed reply under the history head, and an Interrupt the run
     answered must not read as open. The rest of the run comes from the
     replay.
   - **UI Snapshot**: it is cut at the same point. Its cards and open forms
     are those of the session at the cut, derived from the events before it.
     Cards live in session state (ADR-0014 §3), which already holds the run's
     card writes, so cutting the events alone would put a new card in the
     snapshot and again in the replay.
   - **Consistency**: a read holds no lock. It reads the run state before and
     after the session and retries when the two differ, so a run that starts
     or ends mid-read never shows as half a reply.
   - **After the run**: reads include its events, as today.

7. **Deleting a thread stops its run first.**
   - **Order**: deleting an AG-UI thread stops its Detached Run, if one is
     going (decision 4), then takes the thread's lease itself, waiting with a
     bound, and deletes under that lease. No run can start or write in
     between, on any Pod. For AG-UI threads this replaces `DeleteSession`'s
     process-local "deleting" guard.
   - **Timeout**: if the lease is not free within the bound, for instance
     under a run without the opt-in, the delete fails with a retryable error
     rather than deleting under a live run.
   - **Run state and log**: the delete drops both, so a reused `threadId`
     never replays a deleted conversation.
   - **Dashboard**: deleting the open thread only detaches locally. Stopping
     the run is the server's job (#405).

8. **The dashboard re-attaches live, and polls only as a fallback.**
   - **Re-attach**: when history reports `running`, the history adapter's
     `load()` returns `unstable_resume`, and its `resume()` reads
     `GET …/threads/:thread_id/run`. `resume()` folds the AG-UI events into
     `ChatModelRunResult` snapshots of the in-flight assistant message:
     - text, and tool calls with their results;
     - an interrupt outcome as `requires-action`/`interrupt`, with the
       interrupts in `metadata.custom.agui`;
     - pending client tool calls as `requires-action`/`tool-calls`.

     The fold mirrors the aggregator of the assistant-ui release the
     dashboard pins (#391).
   - **Side effects**: the fold itself dispatches `CUSTOM butter.a2ui` to the
     A2UI store and `STATE_SNAPSHOT` / `STATE_DELTA` to the shared state,
     because the runtime applies neither on this path.
   - **Pitfall**: `load()` returns `unstable_resume` only together with a
     working `resume()`. Without one the runtime would start a real run.
   - **Fallback**: when attaching answers 204, or ends with the fallback
     marker, the page polls history until `running` is gone, then hydrates
     the final messages and the UI Snapshot (#406).
   - **Failed and stopped turns**: an inline notice built from `RUN_ERROR`, a
     Stop or `lastRun` shows the outcome and the input, with Restore input
     (#408).
   - **Cutover gate**: AG-UI Chat replaces Chat only once live re-attach
     works (#407). Polling alone is not enough.

9. **The in-process async chat is retired at cutover.**
   - **Route**: AG-UI Chat becomes Chat at `/chat`, and `/agui-chat`
     redirects there (#409).
   - **API**: `SubmitAgentInvocation`, `WatchAgentInvocation` and `asyncrun`
     are removed at cutover, as a breaking change with no deprecation window
     (#410). Only the old Chat uses them. `CancelAgentInvocation`,
     `GetAgentInvocation` by ID and `ListAgentInvocations` stay.
   - **Old sessions**: `web-chat` sessions are neither migrated nor kept
     read-only. A one-off cleanup deletes them with their Invocation records
     and input parts, and only on explicit confirmation (#411).

## Considered Options

- **Detach every run.** Rejected. A standard AG-UI client could no longer
  stop a run: aborting the request would leave it running, and spending,
  until it ended on its own.
- **Run AG-UI turns on `asyncrun`**, the long-term suggestion in
  `docs/research/assistant-ui-integration.md` §6. Rejected. `asyncrun` keeps
  runs, cancellation and observers in one process, and production runs
  several replicas.
- **Fan the events out over pub/sub, or through `asyncrun`'s watch hub.**
  Rejected. Neither keeps anything for a late observer, which needs the run
  from `RUN_STARTED`.
- **Rebuild a late observer's stream from the session.** Rejected. Text
  deltas are never stored, and the sink's mirror and card baseline belong to
  the run.
- **Reuse `internal/eventqueue`.** Rejected. A consumer group hands each entry
  to one consumer, which acknowledges it. A Run Log is read in full by every
  observer.
- **One log per thread.** Rejected. It would mix runs, and keeping callers
  apart would rest on a thread ID having a single holder. A log per caller,
  thread and run holds exactly one run.
- **Keep reads under the lease.** Rejected. Opening a thread mid-run becomes
  the normal case, and a read could wait out the whole maximum run duration.
- **Cut at the run's start, without its input.** Rejected. The user's own
  message would vanish until the run ended, a resumed reply would hang under
  the previous reply, and an Interrupt the run answered would read as open.
- **Keep the run state on the Invocation record.** Rejected. A run without
  the opt-in has no AG-UI-owned record, and a record outlives a dead Pod
  until the sweep, so a dead run would read as running. A key renewed with
  the lease lapses with it.
- **Stop through `runner.CancelInvocation`.** Rejected, as in ADR-0015: it
  reaches only the Pod that receives the Stop.
- **Poll only.** Rejected for cutover (#389): reloading mid-run must resume
  the live stream. Polling stays as the fallback.
- **A deprecation window for the async Invocation API.** Rejected. Only the
  old Chat uses it, and it goes with that chat.

## Consequences

- **Redis per Detached Run**: one Stream append per event, with text
  coalesced, plus the run state, lease renewals, a stop subscription and its
  observers' reads. The log is bounded and short-lived, so Redis memory grows
  with the runs in flight times the cap, not with history.
- **Connections**: blocking reads are multiplexed per Pod. Stop subscriptions
  should share one subscriber connection per Pod as well. Linear holds one
  per turn, but its worker pool bounds its turns; nothing bounds the Detached
  Runs on one Pod.
- **The read cut**: reads hold no lock, but they now depend on the run state.
  A read during a run shows the conversation up to the turn that started it,
  plus `running`. The reply arrives through the replay, or in the first read
  after the run ends. A run whose Pod died reads as running for at most one
  lease TTL.
- **No cap yet**: one run per thread is the only limit. As with the async
  chat today, nothing caps one user's or one workspace's concurrent Detached
  Runs.
- **Deploys**: a rolling deploy fails the Detached Runs on the Pods it
  replaces, FAILED with a shutdown reason. They are not resumed or rerun, and
  the user sends again.
- **Pi and Cursor**: a box turn keeps going after the browser leaves, until
  the Agent's `max_run_seconds` or the AG-UI maximum, whichever comes first.
  An Agent with unlimited box turns is still capped by the AG-UI maximum.
- **Clients that do not opt in**: their POST is unchanged. Reads of their
  threads stop answering 409 during a run. Stop and attach do not reach their
  runs.
- **API tokens**: API-token callers still share the AG-UI user `agui-user`,
  so their threads, and now their Run Logs, are only as separate as their
  thread IDs.
- **Without Redis**: in-process versions of the run lease, the run state, the
  Run Log and Stop serve a single process and the tests. They are held to the
  same contract tests, as the session guard and Linear's coordinator are.
- **Docs**: as the slices land, `docs/api.md` gains
  `forwardedProps.butterRun`, the stop and attach endpoints, `running` and
  `lastRun`, the stop code and the fallback marker. After #410,
  `docs/design-dashboard-chat-async.md` describes a retired design.
- **Left for later**:
  - caps on concurrent Detached Runs per user or workspace;
  - re-attach for standard AG-UI clients through `connectAgent()`, which
    `HttpAgent` does not implement;
  - guidance for third parties building on the Run Log.
