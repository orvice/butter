# agent_id as the Agent repository's logical primary key

Issue #241 closes the Agent ID rollout (#210, #213): every persisted Agent
carries an immutable, workspace-unique `agent_id`, and every serving path
already resolves agents by it. What remained was the persistence layer, the
lifecycle Saga, the migration-era RPC surface, and the frontend, all of which
still accepted the runtime name as a lookup key.

## Decision

**(workspace_id, agent_id) is the Agent repository's logical primary key.**
`AgentRepository` exposes only ID-keyed operations — `GetAgent`,
`DeleteAgent`, and the mutation methods locate the record via
`agent.agent_id`; `GetAgentByID` and `AgentIDExists` are removed as separate
methods because ID lookup *is* the contract. Both adapters reject an empty
`agent_id` at the seam (`ErrMissingAgentID`), so no caller can silently fall
back to name-shaped semantics.

**MongoDB keeps legacy `_id` values as opaque physical identifiers.** The
historical `_id = workspace_id:name` scheme is not rewritten: all agent CRUD
and CAS filter on the unique `(workspace_id, agent_id)` index and update via
`$set`, so a legacy document's `_id` survives every write, including renames.
New documents get a random ObjectID-hex `_id` (no colon), which can never
collide with a legacy name-shaped value. This avoids downtime, transactions,
shadow collections, and partial migration states. Per-workspace runtime-name
uniqueness — previously an accident of the composite `_id` — moves to an
explicit partial unique index on `(workspace_id, name)`; the memory adapter
enforces the same invariant explicitly because its map is now keyed by
`agent_id`.

**The migration-era surface is retired, replaced by a read-only verifier.**
`AssignAgentID`, `GetMigrationReadiness`, and `MigrateAgentsV2` return
`Unimplemented` (their proto shapes stay so historical clients get a precise
error). `VerifyAgentIDCutover` (global admin) and a non-fatal startup pass
(`application.RunAgentIDCutoverVerifier`, replacing the retired #213
`backfillConsumerAgentIDs`) report exactly which workspace/entity still
violates the contract: missing/invalid/duplicate agent IDs, embedded
`sub_agents`, legacy name-based workflow node refs, unresolved
`child_agent_ids`/workflow refs, `MIGRATION_REQUIRED` lifecycle, consumer
records (channel/cron/automation/forum) without agent IDs, and runtime-name
conflicts under the runner's global-uniqueness constraint.

**Legacy composition paths are gone.** Agent construction consumes only
`child_agent_ids` + the per-workspace pool; embedded `sub_agents` are never
built (the field remains readable so historical records decode). Workflow
AGENT nodes must reference their target via `agent_id`; the legacy `agent`
name field is never resolved.

## Deferred: the internal ADK runtime `Agent.name`

`Agent.name` remains the ADK runtime registration key: the runner registers
runnable agents under their bare name and requires global (cross-workspace)
uniqueness, sessions/invocations key summaries by it, and `display_name` owns
presentation. Making `name` presentation-only — or deriving the runtime key
from `(workspace_id, agent_id)` — is a runner-naming redesign with its own
migration (session continuity, invocation summaries, A2A cards) and is
**explicitly out of scope here**. Until that decision, `name` is
server-controlled: preserved on update, never a selector.

## Consequences

- Repo callers cannot express "find agent by name" at all; the only name-ish
  surfaces left are historical/display snapshots (invocation filters, cron
  execution records, forum post authors) which never route execution.
- The shared conformance suite (`internal/repo/config/repotest.RunAgents`)
  runs against both adapters; Mongo-specific tests pin the legacy-`_id`
  preservation and new-`_id` collision-avoidance behavior.
- A deployment that still has violating records sees them logged at startup
  and via `VerifyAgentIDCutover` instead of being silently patched — the
  backfill era is over; remaining fixes are explicit operator actions.
