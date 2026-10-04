# ADR-0015: Linear Agent Sessions as a Butter entry point

- Status: Accepted
- Date: 2026-09-30
- Issue: #357 (PRD), #358 (spec), #360–#370 (implementation)
- Builds on: ADR-0002 (Interrupt state derived from session events), ADR-0008
  (Telegram Channels and Destinations), ADR-0009 (the retry boundary), ADR-0011
  and ADR-0012 (Pi and Cursor agents backed by a ButterBox)

## Context

Linear lets an OAuth application act as an agent. The app is installed with
`actor=app` and the `app:assignable` / `app:mentionable` scopes, which gives it
an app user of its own in each Linear organization it is installed in (Linear
calls an organization a workspace; its `viewer.id` differs per installation).
Delegating an issue to that app user or
mentioning it opens an **Agent Session** in Linear. Linear then sends
`AgentSessionEvent` webhooks: `created` when the session opens, carrying a
formatted `promptContext`, and `prompted` for each later message, with the
message in `agentActivity`. The agent answers by posting activities to the
session:

- `thought`, `action` (`action`, `parameter`, optional `result`), `elicitation`,
  `response` and `error`. Only `thought` and `action` may be ephemeral.
- The webhook must be answered within 5 seconds.
- After `created`, an activity or an external URL update must follow within
  10 seconds, or Linear marks the session unresponsive.
- A prompt carrying `signal: "stop"` must halt the agent's work, which then
  confirms with a `response` or `error`.
- An `elicitation` carrying `signal: "select"` offers `signalMetadata.options`
  (`label`, `value`). The user may answer with free text instead of picking an
  option.
- Linear derives the session state it shows (`pending`, `active`,
  `awaitingInput`, `error`, `complete`, `stale`) from these activities.
- Each webhook is signed with HMAC-SHA256 of the raw body under the app's
  signing secret (`Linear-Signature`) and carries a `webhookTimestamp`.

butter-box#20 already runs pi as a Linear agent inside one ButterBox. It keeps
tokens and a Linear-to-pi session map in local files, queues and cancels runs
in process memory, and can only drive pi on that box. That fits a standalone
box, but Butter is multi-tenant and runs on several Pods, and it already solved
the same shape of problem for Telegram:

- a public webhook that means "durably accepted" once it answers 200,
- a Redis Stream consumer group that any Pod can drain,
- per-session Redis leases,
- a processing record whose state says whether the Agent may have run,
- a credential seam backed by the database master key.

Several existing seams do **not** carry over as they are:

- `runner.CancelInvocation` cancels only runs in its own process, but a Linear
  stop can arrive on any Pod.
- A Telegram event whose session is busy is left unacknowledged. It is
  reclaimed only after `reclaimIdle` (2 minutes), too slow for a follow-up in a
  live Agent Session.
- A Telegram worker waits for its whole claimed batch to finish before it reads
  again. A long turn would push another session's `created` past the 10-second
  deadline.
- MCP OAuth keeps its flow state in memory, so its callback only works on the
  Pod that started the flow.
- pibox and cursorbox yield only the final assistant text, so a Pi or Cursor
  run has no tool events to show as progress.

## Decision

1. **Butter serves Linear itself, for any Agent.** Butter terminates Linear's
   webhooks and runs the routed Agent through the normal runner: LLM,
   Workflow, Pi and Cursor Agents alike. Pi and Cursor still reach the box
   through pibox and cursorbox. The box-local integration (butter-box#20)
   stays for standalone boxes. One Linear app points at exactly one of the two,
   because both would answer every session.

2. **A Linear App and its Linear Installations are workspace resources.**
   - **Linear App**: one Linear OAuth application registered by one workspace.
     It has an immutable ID, a display name, the routed `agent_id`, an inbound
     enabled flag, an allowlist of Linear user IDs, an optional
     `max_run_seconds` (default 1800, 0 means unlimited, as in pibox) and an
     optimistic `revision`. The client secret and the webhook signing secret
     are not proto fields. They are encrypted through a credential seam under
     the database master key, as the Telegram Bot Token is (ADR-0005,
     ADR-0008), and the resource reports only a `credential_state`.
   - **Linear Installation**: that app installed in one Linear organization.
     It holds `organization_id`, the organization name, `app_user_id` (the
     installation's `viewer.id`), the granted scopes and a `credential_state`.
     Its access and refresh tokens sit behind the same credential seam.
     `(app_id, organization_id)` is unique.
   - **Routing**: v1 routes one Linear App to one Agent. The app user is the
     Agent's identity in Linear, the name and avatar people mention, so a
     second identity means a second app.
   - **Reference guard**: deleting an Agent a Linear App routes to is refused,
     as it is for Telegram Destinations.

3. **Three routes, each with its own authentication.**
   - **Install**: an authenticated `BeginLinearInstall` RPC (workspace
     owner/admin) returns Linear's authorize URL with `actor=app` and the
     scopes `read,write,app:assignable,app:mentionable`. No install secret
     exists.
   - **OAuth callback**: `GET /api/linear/oauth/callback` is public and shared
     by all apps. It is authenticated by a single-use state bound to
     `(workspace, app, initiating user)` and kept in Mongo with a TTL, like the
     login `oauthstate`, so any Pod can finish the flow.
   - **Webhook**: `POST /api/linear/webhook/:app_id` is public and bypasses
     workspace auth. The app ID in the path is immutable, and the workspace is
     read off the App, never off the payload. The handler checks the signature
     in constant time over the raw body before parsing. It then rejects a
     `webhookTimestamp` more than 60 seconds off and ignores events other than
     `AgentSessionEvent`, as well as events for an organization with no
     installation. The body is capped at 1 MiB, because the signature covers
     the body, so it has to be read before anything can be verified.
   - **Organization**: an event is matched to an installation by its top-level
     `organizationId`. An event without one resolves to the App's only
     installation. If the App has several installations, the event is ignored
     with a warning.
   - **URLs**: both public URLs are derived from the immutable IDs and a
     global-admin Linear public base URL setting, which mirrors the Telegram
     admin setting. They are never stored. Without the setting,
     `BeginLinearInstall` fails with a failed precondition.

4. **Ingestion reuses the Telegram durable-queue contract on its own
   Stream.**
   - **Accept**: HTTP 200 means that one Lua script deduplicated the delivery
     and appended it to `butter:linear:events`. The dedupe key is the
     `Linear-Delivery` header, or the SHA-256 of the raw body when the header
     is absent. A queue failure answers 503, so Linear redelivers rather than
     losing the event.
   - **Separate Stream**: Linear gets its own Stream and consumer group, so a
     Telegram backlog cannot push a `created` past its 10-second deadline.
   - **Shared plumbing**: accept, claim, touch and ack are extracted from
     `telegramqueue` into a protocol-neutral package parameterized by Stream
     key. Telegram keeps its event type and behaves as before.
   - **Worker**: the Linear worker never waits for an Agent turn before reading
     again. Each claimed event runs in its own heartbeated goroutine, with a
     per-Pod concurrency bound. The first thing a handler does for `created` is
     post the acknowledging `thought`, before it takes a lease or resolves the
     Agent.

5. **One Linear Agent Session is one Butter session.**
   - **Session ID**:
     `linear:{app_id}:{agent_session_id}:{agent_id}`, under app name `linear`.
     Re-pointing the App to another Agent therefore starts a fresh history
     rather than inheriting one.
   - **Pi and Cursor**: pibox and cursorbox already bind one box session per
     (Butter session × Agent) in session state. That replaces the Linear-to-pi
     session map butter-box#20 keeps in a file.
   - **User ID**: the ADK user ID is `linear:{app_id}:{organization_id}`. The
     conversation belongs to the Agent Session, not to one person in it.
   - **Principal**: the prompting Linear user is recorded as `linear:<user_id>`
     for Memory Capture provenance. It is not a verified Butter user.
   - **Context**: the first turn carries Linear's `promptContext`. A turn that
     finds no history for the session (it was cleared, or the Agent changed)
     gets the context again.

6. **Turns in one Agent Session are serialized, and a follow-up is queued
   instead of deferred.**
   - **Lease**: a Redis lease per session serializes turns.
   - **Follow-up**: a `prompted` event that finds the lease held is appended to
     a per-session Redis follow-up list, answered with a "queued" `thought`,
     and acknowledged. Its processing record starts in `QUEUED`.
   - **Drain**: before it releases, the lease holder drains the list into its
     next turn, joining the queued messages in order. Their records advance
     together under that turn's invocation ID.
   - **Atomicity**: enqueue-or-acquire and release-or-drain are single Lua
     scripts. A follow-up accepted during a turn is therefore never lost and
     never runs concurrently with that turn.
   - **Crash**: if the holder dies, the reclaim path first settles the
     interrupted turn under ADR-0009, then drains the follow-ups. They never
     started, so running them is safe.

7. **Stop bypasses the lease and cancels across Pods.**
   - **Stop path**: a `stop` prompt never waits behind the turn it is meant to
     stop. In one atomic step its handler clears the follow-up list and, while
     a turn holds the session, sets a short-lived stop marker holding that
     holder's lease token and publishes a nudge on a Redis channel for the
     session. Binding the marker to the token means a stop can never reach a
     later holder.
   - **Holder**: the lease holder subscribes for the length of its turn and
     checks the marker when the turn starts, so a nudge that races the start is
     not lost. It cancels the turn context. A Pi or Cursor Agent aborts through
     its single `AbortSession` path.
   - **Confirmation**: the holder posts the stop confirmation (`response`) once
     the cancelled turn can post nothing more. If no lease is held, the stop
     handler itself answers that nothing is running.
   - **Record**: a stopped turn ends `CANCELLED`. It is not dead-lettered,
     since the user chose it. This cross-Pod cancel is scoped to Linear. Dashboard
     async cancel stays in-process.
   - **Generalized by ADR-0016**: AG-UI runs that outlive their request are
     stopped by this same mechanism, a marker bound to the run's lease token
     plus a nudge, with the holder subscribing before it checks the marker.
     ADR-0016 also retires the in-process dashboard async cancel, together
     with `asyncrun`, when AG-UI Chat becomes the dashboard's only chat
     (#389, #410).

8. **The retry boundary is ADR-0009's.**
   - **Record**: each accepted delivery gets one Linear processing record
     (unique on delivery ID, 30-day TTL) whose state answers "may the Agent
     have run?".
   - **Retries**: everything before that boundary retries automatically.
     `FAILED_UNCERTAIN` after it is dead-lettered and never rerun. A reclaimed
     uncertain turn posts an `error` saying the run was cut short and that a
     new message continues the same session.
   - **Resend**: the complete response is persisted before it is posted, so a
     failed post is resent without invoking the Agent.
   - **Separate type**: the record type is separate from Telegram's because
     the address and delivery shapes differ: one activity rather than message
     segments. The state machine is the same.

9. **Activities come from a progress mapper fed by the runner's turn event
   callback.** It is not a `streamorch.Sink`: `streamorch.Run` returns only the
   final text, and the final activity needs the full turn result, including
   pending Interrupts.
   - **Created**: `created` gets the acknowledging `thought`, and an
     `agentSessionUpdate` sets an external URL to the Butter session in the
     dashboard.
   - **Tool calls**: each function call becomes an `action`: a verb or the tool
     name, plus a short, redacted parameter. At most one is posted every
     3 seconds, and consecutive duplicates are dropped.
   - **Model reasoning**: thought parts are not forwarded. Only operational
     notices go out, as ephemeral `thought`s. In practice that is context
     compaction: the runner exposes no model-retry signal, so there is no
     retry notice to forward.
   - **Final**: the final text becomes the `response`. Failures and the
     `max_run_seconds` deadline become `error`. A response over 8000 runes is
     truncated, with a note and the external URL. The limit matches
     butter-box#20 and stays until Linear's real limit is confirmed.
   - **Human Input**: a Workflow that pauses on a Human Input node posts an
     `elicitation` with the Interrupt's question. A form whose only field is
     `SINGLE_CHOICE` uses `signal: "select"` with its options. Other forms use
     the field instructions every non-A2UI entry point appends (ADR-0014). The
     next `prompted` resumes through the runner's FIFO implicit resume
     (ADR-0002), so there is no new pending store. Linear shows the session as
     `awaitingInput`.
   - **Redaction**: all posted text goes through one shared regex redaction,
     `mem0memory.Redact` moved to a neutral package. That redaction is best
     effort.

10. **Admission is explicit, and the risk is stated.**
    - **Allowlist**: an empty allowlist admits every member of the installed
      organizations. Otherwise only listed Linear user IDs may open or prompt a
      session, and anyone else gets an `error` activity. When an event names
      no prompting user, a non-empty allowlist denies it.
    - **Disabled after acceptance**: an event whose App has since been
      disabled, or whose Agent is not loaded, gets an `error` activity and runs
      nothing. Unlike Telegram, a silent reply is not an option here, because
      Linear would mark the session unresponsive.
    - **Risk**: for a Pi or Cursor Agent, anyone admitted can make it run
      commands on the ButterBox. The dashboard and docs say so where the App is
      configured.

11. **Tokens are resolved per call and refreshed under a lease.**
    - **Per call**: a Linear client is built from the decrypted installation
      token on each use, with no cache, like `telegramapi.Factory`.
    - **Refresh**: refresh tokens rotate, so two Pods refreshing at once would
      invalidate the installation. A refresh therefore runs under a Redis lease
      per installation and is written with a revision compare-and-swap.
    - **Failure**: an `invalid_grant` marks the installation's
      `credential_state` as needing reinstall. It surfaces in the dashboard
      rather than failing silently on every session.

## Considered Options

- **Forward to the box-local integration.** Rejected. It drives only pi, on
  one box, for one tenant, with file state. The workspace would lose every
  other Agent type, Memory, invocation records and Butter's multi-Pod recovery.
- **One platform-wide Linear app, with Linear organizations mapped to
  workspaces.** Onboarding is easier: install, no OAuth app registration.
  Rejected for v1:
  - every tenant's Agent would share one identity in Linear,
  - routing a webhook to a workspace would rest on a claimed organization
    mapping rather than on a per-App secret,
  - operators would hold Linear credentials on behalf of all tenants.
  It remains possible as a later addition.
- **Share Telegram's Stream.** Rejected because of the deadlines. Linear's
  10-second rule must not wait behind Telegram traffic. The mechanics are
  shared instead.
- **Defer busy sessions by leaving the event unacknowledged, as Telegram
  does.** Rejected. A follow-up would wait at least `reclaimIdle`.
- **Handle stop through `runner.CancelInvocation`.** Rejected. It reaches
  only the Pod that happens to receive the stop.

## Consequences

- **Setup**: a workspace registers one Linear OAuth app per Agent identity.
  This is a manual step in Linear's settings that needs Linear admin rights.
  Butter shows the callback and webhook URLs to paste.
- **Redis requirement**: enabling a Linear App requires the same persistent,
  no-eviction Redis that Webhook Telegram Channels require.
- **Refactor**: extracting the Stream plumbing and the redaction helper
  touches Telegram and Memory code, with no change in behavior.
- **Progress for Pi and Cursor Agents** needs pibox and cursorbox to yield
  tool entries as partial function-call events. Until then those Agents show
  the acknowledgement and the final response only. The same events would also
  benefit the dashboard and AG-UI.
- **Left for later**:
  - routing by team, project or label,
  - Linear as an outbound target (Cron delivery, Notify Groups),
  - session plans (`agentSessionUpdate.plan`),
  - the `auth` signal and account linking,
  - repository suggestions,
  - image attachments.
  An Agent that needs to read or write issues uses Linear's MCP server as an
  ordinary MCP Server, not the App's token.

## Questions resolved for v1

The first draft left five questions open. The spec (#358) settles them for v1:

1. **App ownership**: one Linear App per registered OAuth app, owned by one
   workspace (decision 2). A platform-wide app remains a possible later
   addition.
2. **Routing**: one Agent per App. Routing by team, project or label is left
   for later.
3. **Payload assumptions**: decisions 3, 4 and 10 give fallbacks for a missing
   `Linear-Delivery` header, a missing top-level `organizationId` and a missing
   prompting user. These assumptions come from butter-box#20's unchecked test
   items. The first real deliveries are logged at debug level so the
   assumptions can be confirmed and the fallbacks tightened.
4. **Body limit**: responses are truncated at 8000 runes (decision 9) until
   Linear's real limit is confirmed.
5. **Processing record**: a separate type that follows ADR-0009's state
   machine (decision 8). The Telegram record is not generalized.

## Implementation notes

What the implementation (#360–#370) settled beyond the decisions above:

- **Session identity**: the runner uses `ContextInfo.channel_name` as the
  ADK app name, so the channel name is the fixed `linear` rather than the
  App's display name. A rename must never move history. The display name
  travels in metadata instead. The session user is the organization, so
  Memory Capture gets the prompting Linear user through the new
  `ContextInfo.metadata.principal` override (`memoryhook`).
- **Session coordination**: `SessionCoordinator` has an in-memory and a
  Redis implementation, held to one contract suite. A holder that crashed
  leaves `PROCESSING` records behind, and the next holder of the session
  sweeps them (`FAILED_UNCERTAIN`, one "cut short" notice) before running
  the backlog. This also covers follow-ups the dead holder had already
  drained.
- **Worker**: the per-Pod concurrency bound is 32. A Pod with no free slots
  claims nothing and leaves the Stream to other Pods.
- **Payload assumptions are not yet confirmed.** No real Linear delivery
  has been observed. The dedupe header (`Linear-Delivery`), the top-level
  `organizationId` and the prompting-user fields (the activity's `userId`
  or `user.id`, then the session's `creatorId` or `creator.id`, then the
  comment's user) are handled as decisions 3, 4 and 10 describe. The first
  deliveries are logged at debug level so these can be checked, and the
  fallbacks tightened once they are.
