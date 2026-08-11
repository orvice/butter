# Dashboard Chat: New-Chat Home and Asynchronous Invocations

Status: Proposed

## Context

The dashboard currently opens `/chat` on the user's most recent `web-chat`
session. A separate new-chat state already exists behind `?new=1`, and the
global sidebar already lists recent conversations.

The target experience follows Open WebUI's interaction model:

- `/chat` opens a new-chat draft rather than an existing conversation.
- The main area presents a composer with an Agent selector.
- Recent conversations remain in the sidebar.
- Sending a message starts work that is independent of the browser request.
  Navigating away, closing the tab, or losing the connection must not cancel
  the Agent invocation.

This design is limited to dashboard chat. Existing synchronous Agent, session,
channel, cron, automation, and external API behavior remains compatible.

## Goals

- Make new chat the canonical `/chat` entry state.
- Avoid persisting empty sessions when a user only selects an Agent.
- Bind every dashboard chat session to one workspace and one Agent.
- Let Agent generation continue after the observing frontend disconnects.
- Preserve live output while the user is watching and allow observation to be
  resumed after navigation.
- Surface running, failed, waiting-input, and unread states in chat history.
- Make async submission idempotent and preserve accepted multimodal input.

## Non-Goals

- Replacing the synchronous semantics of `InvokeAgent`, `StreamAgent`,
  `ReplySession`, channels, cron, or automation.
- Supporting multiple Butter service instances in the first release.
- Resuming an interrupted invocation after a Butter process restart.
- Supporting API Token callers in the first release.
- Adding per-user or per-workspace concurrency limits in the first release.
- Replaying every transient token delta emitted while no observer was attached.
- Migrating legacy sessions that do not have a workspace identity.

## Chat Navigation and Layout

The canonical routes are:

| Route | Meaning |
| --- | --- |
| `/chat` | New, unpersisted chat draft |
| `/chat?agent=<agent_id>` | New draft with an Agent preselected |
| `/chat?session=<session_id>` | Existing persisted conversation |

The old `/chat?new=1` route remains accepted for compatibility and is
normalized to `/chat`.

The new-chat page uses an Open WebUI-style composer with a searchable Agent
selector. It does not use the current full-page Agent card gallery. The
selector only includes runnable Agents: lifecycle `ACTIVE` and legacy
`UNSPECIFIED`. The selected Agent can change until the first message is sent.
After the session is created, its Agent is fixed.

The client remembers the last selected runnable Agent separately for each
workspace. If that Agent no longer exists or is not runnable, the selection is
cleared. Message text and attachments are not persisted as a local draft
across refreshes or workspace switches.

The sidebar continues to show the newest 100 conversations in the current
workspace, grouped by Today, Previous 7 days, and Older. Search applies to
those 100 records. The existing `/sessions` area remains the full operational
session explorer.

Switching workspace always returns chat navigation to `/chat` and loads the
new workspace's history and remembered Agent. A session URL that does not
exist in the current workspace renders an explicit "Chat not found in this
workspace" state rather than silently falling back to a new chat.

If a session's Agent later becomes unavailable, its history remains readable,
but the composer is disabled and the UI offers a new-chat action. The session
must never be silently rebound to another Agent.

## Workspace and User Ownership

Workspace ownership is a first-class session field, not a client-editable
entry in `state`.

- Session creation derives `workspace_id` from the authenticated request
  context.
- Dashboard list, get, update, read-state, and delete paths validate the
  session workspace on the server.
- New chat history ignores legacy sessions without `workspace_id`; no legacy
  inference or data migration is required.
- Dashboard members and workspace owners may only submit, watch, get, cancel,
  mark read, or delete their own chat sessions.
- A global admin may access another user's session for operational support.
- The existing workspace-level invocation operations view remains separate
  and does not grant access to the live private-chat observation stream.

## Async RPC Contract

Add explicit RPCs instead of changing `StreamAgent` semantics:

### `SubmitAgentInvocation`

A unary command that validates and durably accepts one dashboard chat turn,
then returns without waiting for Agent execution.

The request includes:

- `request_id`: client-generated idempotency identifier.
- `agent_id`: required for a new session and required to match the bound Agent
  for an existing session.
- `session_id`: empty for a new chat, otherwise an existing session.
- Text or multimodal `InputPart` values.
- Optional model override if dashboard chat exposes that control later.

The response includes:

- `session_id`
- `invocation_id`
- Current invocation status
- Whether a new session was created

When `session_id` is empty, this command creates the workspace-owned,
Agent-bound session and the invocation as one coordinated operation. The
frontend must not call `CreateSession` first.

Submission is idempotent by `(workspace_id, session_id-or-new-draft,
request_id)`. A repeated request returns the original session and invocation
instead of running the Agent twice.

The response means the complete validated input has been persisted. Since a
request may contain up to 20 MiB and a MongoDB document has a 16 MiB limit,
multimodal parts must be stored as separate invocation-input records rather
than embedded in the invocation document. Temporary input copies may be
removed after the corresponding session input event is durable. Input for a
run that failed before that point is retained for a bounded recovery period.

### `WatchAgentInvocation`

A read-only server stream keyed by `invocation_id`.

- Attaching or detaching an observer never starts, stops, or owns execution.
- Multiple observers may watch the same invocation.
- The stream reports the current state, subsequent run events and text deltas,
  and one terminal state.
- A slow or disconnected observer must not apply backpressure to the Agent
  run.
- Re-entering a chat first loads persisted session events and invocation
  status, then attaches to the live stream if the invocation is still active.
- The system does not persist and replay every token delta emitted without an
  observer. Persisted session events and the final result remain complete.

### `GetAgentInvocation`

Returns the authoritative state of one invocation, scoped by workspace and
session ownership. This lets a reloaded client distinguish an active,
terminal, or stale invocation before it starts watching.

### `CancelAgentInvocation`

The existing cancellation RPC remains the cancellation command. It is updated
to understand queued and async dashboard invocations and to enforce chat
ownership. Cancellation is explicit: navigation and observer disconnects do
not invoke it.

## Invocation Lifecycle

The status model becomes:

```text
QUEUED -> RUNNING -> SUCCEEDED
                  -> FAILED
                  -> CANCELLED
```

`CANCELLED` represents an explicit user stop or coordinated chat deletion and
must not be counted as a failed invocation. A process shutdown or maximum-run
timeout records `FAILED` with an actionable reason.

The first release has no general user or workspace concurrency cap. It does
enforce one active invocation per session. A second submission to the same
session returns `failed_precondition` with the active `invocation_id`; it is
not queued behind the current turn. Different sessions may run concurrently.

The async manager detaches execution from the request context while preserving
the authenticated workspace and user identity captured at acceptance time. A
configurable `chat_async.max_run_duration` bounds a run, with a default of 30
minutes. Workflow Human Input is not an indefinitely running invocation: the
turn ends successfully and the session becomes `WAITING_INPUT`.

The first release is process-local. On startup, stale `QUEUED` or `RUNNING`
dashboard invocations left by a previous process are marked `FAILED`; they are
not retried automatically because Agent tools may have external side effects.

## Frontend Runtime Behavior

Sending the first message calls `SubmitAgentInvocation` directly. The client
navigates to `/chat?session=<id>`, renders the accepted user turn, and attaches
to `WatchAgentInvocation`. Existing sessions use the same command without
creating another session.

Navigation is independent of execution:

- Users may open another session, start another chat, leave the chat page, or
  close the tab while an invocation runs.
- The background invocation continues.
- Returning to the session reloads persisted events and resumes observation.
- Stop is the only ordinary action that cancels a run.

The recent-chat row exposes state without requiring the conversation to be
open:

- Spinner while an invocation is queued or running.
- "Needs input" when a workflow session has pending human input.
- Failure indicator for a failed invocation.
- Unread completion indicator when new terminal content has not been viewed.

While the dashboard remains open, a background completion or failure also
produces an in-app toast. The first release does not request browser or
operating-system notification permission.

Invocation failure is rendered inline beneath the associated user turn. The
error remains invocation metadata and is not inserted as an Agent-authored
session event. Retry is never automatic. The UI restores the submitted input
for review or editing, and resubmission creates a new invocation that may
repeat external tool side effects.

## Read State

Session read state is server-backed so it works after a tab is closed and
across browsers or devices. A session records `last_read_at` for its owning
user.

The client marks a session read only when:

- That session is the active view.
- The document is visible.
- The client has loaded through the latest known event or terminal invocation
  state.

A background observer, hidden tab, or sidebar status update does not mark the
conversation read.

## Workflow Human Input

When a Workflow Agent requests human input:

- The current invocation reaches `SUCCEEDED`.
- Pending Interrupt state continues to be derived from session events through
  `internal/runtime/interrupt`.
- The chat session summary exposes `WAITING_INPUT` / "Needs input".
- The composer is enabled because no invocation is active.
- The user's next `SubmitAgentInvocation` on the same session follows the
  existing implicit FIFO resume path.

## Titles

Title generation moves from the frontend to a server-side best-effort action
after the first successful turn. Invocation success does not wait for title
generation. Manual and concurrent titles continue to win, and failure falls
back to Agent display name or shortened session ID without reordering history.

## Delete Semantics

Deleting a session with an active invocation is one coordinated server
operation:

1. Reject further submissions for the session.
2. Cancel the active invocation.
3. Wait until it can no longer append session events.
4. Delete session events and stored invocation input parts.
5. Remove user content from associated invocation records.

Operational invocation metadata such as Agent identity, terminal status,
timestamps, and latency may remain for metrics and audit. The `input`,
`output`, and content-bearing `error` fields are redacted.

## Single-Instance Boundary

The initial async manager, cancellation registry, and observer fan-out are
in-process. This is consistent with other repository areas that currently
defer distributed coordination. The deployment documentation must state that
dashboard async chat is single-instance only.

## Required Follow-Ups

These items are explicitly deferred, not rejected:

1. Multi-instance task ownership using a shared queue or lease.
2. Cross-instance event observation through Redis Pub/Sub, Streams, or another
   shared event transport.
3. Distributed cancellation routing and session-level exclusion.
4. Durable recovery or operator-controlled replay after process restart.
5. API Token support with an explicit service-account ownership model.
6. Configurable per-user and per-workspace concurrency and backpressure.
7. Retention and garbage-collection policy for failed pre-run input records.

## Acceptance Criteria

- Opening `/chat` never automatically opens an existing session.
- Selecting an Agent without sending does not create a session.
- Sending the first message returns an invocation promptly and creates exactly
  one workspace-owned session, including after an idempotent retry.
- Closing or navigating away from the page does not cancel the invocation.
- Returning to a running chat shows persisted progress and resumes live
  observation.
- Two simultaneous submissions to one session cannot run concurrently.
- Sidebar state distinguishes running, failed, waiting-input, and unread chat.
- A stopped invocation is `CANCELLED`, not `FAILED`.
- A stale invocation after service restart becomes `FAILED` and is not rerun.
- An unavailable Agent leaves its historical session readable but not writable.
- Deleting an active chat cannot leave a task writing new events afterward.
- Ordinary users cannot watch or cancel another user's dashboard chat.
