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

Direct reply-from-Telegram resume (mapping channel identities onto cron sessions) was
considered and deferred: it is an independent, large piece of session-identity plumbing.
No interrupt timeout in phase 1; a paused execution waits until answered or its session
is cleared.
