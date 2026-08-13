# ADR-0009: The Telegram retry boundary is "may the Agent have run?"

- Status: Accepted
- Date: 2026-08-13
- Issue: #264 (PRD), #271 (this decision)
- Builds on: ADR-0008 (Telegram Channels and Destinations)

## Context

Telegram delivers an update at least once, and Butter consumes it at least
once more: the webhook handler enqueues to a Redis Stream, and a worker in any
Pod claims it from a consumer group. A worker can die at any point — before
touching the Agent, mid-turn, after the Agent produced output, or partway
through delivering a multi-segment reply.

"Retry everything" and "retry nothing" are both wrong. Retrying everything can
repeat an Agent's external side effects: a tool that filed a ticket, sent an
email, or moved money runs twice. Retrying nothing loses work that never
started and turns a transient Redis blip into a message the user never gets an
answer to.

## Decision

Persist one processing record per accepted update, and let its state answer a
single question: **may the Agent have run?**

Everything before that point is retried automatically. Everything after it is
not.

The states exist to mark that boundary, not to describe progress for its own
sake:

- `RECEIVED` — claimed, nothing started. Safe to retry.
- `PROCESSING` — the Agent may be running tools. **Not** safe to retry.
- `READY_TO_DELIVER` — output is persisted; only sending remains. Safe to
  retry *as a send*.
- `SUCCEEDED` — delivered.
- `FAILED` — failed on the safe side of the boundary, or failed to deliver
  output that still exists.
- `FAILED_UNCERTAIN` — failed after the Agent may have acted. Dead-lettered
  for a person, never re-run automatically.

Four consequences follow:

1. **The complete Agent response is persisted before delivery begins.** That
   single ordering is what turns a delivery failure from "re-run the Agent to
   reproduce the text" into "send text we already have". It is also what makes
   the Dashboard's resend action safe.

2. **Resend continues from unsent and failed segments.** An operator
   recovering a partial delivery must not double-post the half that worked, so
   per-segment state is persisted alongside the output.

3. **There is no rerun action.** Not in the API, not in the Dashboard. Once
   Agent work may have produced side effects, deciding to repeat it needs
   context the system does not have — whether the tool was idempotent, whether
   the ticket was already filed. `FAILED_UNCERTAIN` says "a person must look",
   which is honest, rather than guessing on their behalf.

4. **A completed duplicate costs nothing.** `(channel_id, update_id)` is
   unique on the record, so a re-delivered Telegram update finds the finished
   record and is acknowledged without a second Agent run or a second message.

Turns within one derived session are serialized by a Redis lease. Two updates
for the same conversation running concurrently would interleave their history
writes, producing a session that reads as two people talking over each other.
The lease is per session, not global: unrelated conversations stay parallel,
which is the whole point of a multi-Pod worker fleet.

Records, persisted output, and delivery segments expire after 30 days via a
Mongo TTL index — long enough to investigate an incident, short enough that
Telegram message content is not retained indefinitely.

## Alternatives considered

**Retry until success, with idempotency keys on tools.** Rejected as out of
reach: Butter does not control what an Agent's tools do, and MCP servers offer
no general idempotency contract. Claiming exactly-once across arbitrary
external side effects would be a promise the system cannot keep.

**Treat a mid-turn crash as failure and retry.** Rejected: it is the case most
likely to repeat a side effect, and it is indistinguishable from success
followed by a lost acknowledgement.

**Automatically resend after a delivery failure without operator action.**
Partly adopted: delivery *is* retried automatically while the worker still
holds the event, because sending is idempotent per segment. The Dashboard
action exists for the case where the worker gave up entirely.

**Acknowledge on claim rather than on completion.** Rejected: it makes every
crash silently lose a message, which is exactly the failure operators cannot
detect.
