# A cron-triggered workflow that pauses notifies and waits; resume goes through ReplySession

Cron executions have no human on the other end, but workflow agents may pause on a Human
Input node. Instead of forbidding such combinations (which would kill the
approval-style scheduled job — the most valuable HITL scenario), the scheduler detects a
run that ended in a pending interrupt, delivers the node's question through the job's
existing delivery/notify configuration, and records the execution as `WAITING_INPUT`
with the session coordinates. A human resumes by sending a message to that session via
the existing `ReplySession` RPC (or dashboard UI on top of it) — no new resume plumbing,
because interrupt detection lives in the runner (ADR 0002) and applies to every entry
point.

Cron sessions are per-execution (`cron:<job>:<exec-id>`), not per-job: the schedule
keeps firing while an execution waits, and a rerun sharing the waiting execution's
session would have its input consumed as the human's answer (ADR 0002 answers FIFO) and
would close the waiting record with unrelated output. The recorded session coordinates
carry the full ID, so `ReplySession` resume is unaffected; the trade-off is that cron
runs of one job no longer share conversation history.

Direct reply-from-Telegram resume (mapping channel identities onto cron sessions) was
considered and deferred: it is an independent, large piece of session-identity plumbing.
No interrupt timeout in phase 1; a paused execution waits until answered or its session
is cleared.

The "session is cleared" arm is reconciled through the `SessionService.DeleteSession`
RPC only: a session-delete listener (synchronous, mirroring the runner's turn
listeners) cancels the deleted session's `WAITING_INPUT` executions, recording the
reason and delivering per `notify_on`, where cancellations count as failures. Sessions
removed out of band (direct database cleanup) are not reconciled — a lazy sweep was
considered and deferred because no reachable code path deletes cron sessions outside
the RPC today.
