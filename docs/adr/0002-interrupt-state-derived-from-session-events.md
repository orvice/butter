# Pending workflow interrupts are derived from session events, not stored separately

A paused workflow (Human Input node) must turn the session's next user message into a
`FunctionResponse` targeting the pending Interrupt ID. Rather than persisting a
butter-owned pending-interrupt record in Mongo, the runner derives the pending state by
scanning the ADK session's recent events for an `adk_request_input` FunctionCall without
a matching FunctionResponse. ADK v2 already persists workflow run state in
`session.State`; a second butter-owned copy would be a divergent source of truth with
reconciliation bugs, while derivation is restart-safe by construction.

Consequences: resume is implicit — when a session has an unanswered interrupt, the next
inbound message on that session (from any entry point: channel, ReplySession RPC,
StreamAgent) is the answer. When multiple interrupts are pending (parallel branches each
paused), the reply targets the oldest unanswered interrupt (FIFO); precise addressing
can later be added as an optional interrupt ID on ReplySession without breaking this.
Abandoning a paused workflow is done by clearing the session. Phase 1 treats replies as
plain text (no ResponseSchema support).
