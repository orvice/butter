# ADR-0012: Cursor agents reuse the ButterBox model, with a held SendMessage turn

- Status: Accepted
- Date: 2026-09-24
- Issue: #314 (spec), #316 (runtime), #317 (exits), #315 (butter-box CursorService)
- Builds on: ADR-0011 (Pi agents backed by ButterBox)

## Context

We want butter agents backed by Cursor's agent loop (`@cursor/sdk`) running on
the same ButterBox VMs that host pi. butter-box wraps Cursor's standalone
`cursor-sdk-bridge` (Connect/protobuf, `sdk.v1`) behind a small
`butterbox.cursor.v1.CursorService`: `CreateSession(cwd, model, mode)`,
`SendMessage(session_id, message, images)`, `AbortSession`, `ListModels`.
Model inference always goes through Cursor's hosted API with the box's
`CURSOR_API_KEY`.

## Decision

1. **Same shape as Pi, new leaf type.** `AGENT_TYPE_CURSOR` with
   `AgentConfig.cursor` (`butterbox_id`, `working_dir`, `model`, `mode` ∈
   {`agent`, `plan`}, optional `max_run_seconds` — unset 1800, 0 unlimited).
   It is a leaf; instruction, global instruction, MCP servers/refs, skills,
   file mounts, context guard, and remote-agent refs are rejected on write,
   pointing at the box's `.cursor/rules` and `mcp.json`. The ButterBox
   resource, its credential seam, the write-time box check, and the delete
   reference guard are shared with Pi; one box can host both.
2. **Session continuity as in ADR-0011 §3**: one Cursor session per (butter
   session × agent) in ADK session state under `cursorbox:<agent_id>`
   (`cursor_session_id`, `butterbox_id`, `working_dir`). A repointed box or
   directory abandons and recreates; a `NotFound` from the box (restart)
   recreates once; a `NotFound` right after creating is reported as an
   unhealthy box rather than looped on.
3. **A held turn, deliberately.** ADR-0011 §4 made Pi turns asynchronous so a
   box restart yields an honest did-not-finish. CursorService offers no
   cursor/poll pair, and it already guarantees the same honesty: a bridge that
   dies mid-run fails the held call with `Unavailable`, never a stale answer.
   So the cursorbox bridge holds one `SendMessage` for the whole turn. The call
   carries the turn's cancellation but **not** its deadline: a propagated
   Connect timeout would expire on the box first and race butter's
   classification of max-run versus box error. On cancellation or max-run,
   butter also calls `AbortSession` — the single cancellation path.
4. **The Cursor API key stays on the box.** butter never stores it. The box
   tags a missing/rejected key with the stable `google.rpc.ErrorInfo` reason
   `CURSOR_API_KEY_MISSING_OR_INVALID`, which butter turns into "set
   CURSOR_API_KEY on the box", distinct from a rejected box access token.
5. **Box-owned model selection.** Like Pi, a Cursor agent reports no Butter
   model override (`runner.SupportsModelOverride`), so Telegram model
   switching is locked with no Telegram-specific code. The dashboard picks
   from `ButterBoxService.ListButterBoxCursorModels`, with a free-text
   fallback.

## Consequences

- Box resolution (repo lookup, token decryption, bearer) and session-state
  decoding live in `internal/runtime/butterboxconn`, shared by `pibox` and
  `cursorbox`.
- Editing a Cursor agent's `model` or `mode` affects new Cursor sessions only,
  as with Pi's model: the box fixes them when it creates the session.
- A held turn is only as durable as its HTTP connection. A proxy that cuts
  long idle requests ends the turn early (reported as unreachable/did not
  finish); deployments with such proxies must raise their timeouts.
