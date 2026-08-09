# Agent lifecycle as a durable Saga over the existing mutation + content seams

Issue #218 (PRD Phase 3) coordinates database-owned Agent configuration with
Git-owned Agent Content across bound-Agent creation, composite save, soft
delete, restore, and explicit content purge. Mongo and Git cannot share a
transaction, so the control plane must expose durable operation IDs and honest
partial-failure reporting **without claiming a distributed transaction**.

## Decision

Model each cross-system change as a **durable Saga** recorded in a new
`AgentOperation` document (`internal/repo/agentop`, memory + Mongo), with an
explicit status (PENDING/RUNNING/SUCCEEDED/FAILED) and embedded per-step
progress. The operation ID doubles as an **idempotency key**: re-invoking
a lifecycle RPC (or `RetryAgentOperation`) with the same ID resumes from the first
unfinished step, skipping steps already recorded SUCCEEDED. Operation-state writes
use `context.WithoutCancel`, mirroring the automation engine, so a client
disconnect cannot corrupt the record.

The coordinator (`internal/application/agent_saga.go`) **composes** the existing
single-step `mutateWithRuntime`/`deleteWithRuntime` seam (issue #183) for every
DB+runtime step rather than reimplementing write/reload/rollback. It reaches
Git-owned content through a narrow in-package interface
(`agentContentCoordinator`) satisfied by `RepoBindingServiceServer`: the body of
`CommitAgentContent` was extracted into a plain-args `commitContent` so the Saga
and the RPC handler share one commit → validate → publish path. Both servers live
in the same `internal/application` package, so this is a decoupling seam, not a
cycle break — the domain path never imports transport types.

Agent lifecycle gains four states — PROVISIONING, ERROR, DELETING, DELETED —
alongside ACTIVE/MIGRATION_REQUIRED. **Only ACTIVE (and legacy UNSPECIFIED) agents
enter the runner registry**; the single filter lives at the runner build/reload
seam (`internal/runtime/runner`, `isRunnableAgent`). `DeleteAgent` becomes a
**tombstone**: it flips the agent to DELETED and sets `deleted_at`, retaining the
Agent ID (so `AgentIDExists` still blocks reuse and `CreateAgent` steers callers to
`RestoreAgent`) and never touching Git content. `PurgeAgentContent` is a separate
owner/admin action on the repo-binding service that refuses while any non-deleted
agent — in this workspace or an overlapping-binding workspace — still claims the
effective path. Purge is a **direct action, not a durable operation**: it is a
synchronous refuse-or-single-commit with no DB state machine to track, so it
returns its result directly rather than an `AgentOperation` record.

On a create-Saga failure the agent is left in **ERROR** (not PROVISIONING) so the
documented "Saga left it partial, not runnable" contract holds and
`RetryAgentOperation` can resume it to ACTIVE.

Composite save (`UpdateAgentConfiguration`) carries `expected_agent_version`;
optimistic concurrency is enforced by a repository-level compare-and-swap
(`UpdateAgentCAS`, promoted `version` field) so a race between read and write is
caught atomically, returning `Aborted` on conflict.

## Consequences

- **Compensation is best-effort, not atomic.** Local DB+runtime pairs roll back via
  `mutateWithRuntime`; Git commits are never reverted by rewriting history. A
  failure after a successful commit leaves content in place and the agent in
  PROVISIONING/ERROR, recoverable by `RetryAgentOperation`. Composite save commits
  content before patching config, so a config-reload failure leaves a benign
  "content ahead of config" partial that retry reconciles — deliberately not
  presented as a rollback.
- **Lifecycle Saga content commits force DIRECT_COMMIT** regardless of the binding's
  write mode, since agent provisioning cannot be gated on a review PR; ongoing
  content edits (`CommitAgentContent`) still honor CHANGE_REQUEST mode.
- Unbound workspaces (no repository binding) keep the pre-existing single-step
  create path and activate immediately; the Saga path engages only when a binding
  and the operation store are both wired.
- Publication remains serialized per binding by the existing in-process mutex; a
  shared lease for multi-instance deployments is still future work, as is a
  background reconciler over `ListResumable` to auto-heal stuck operations
  (`RetryAgentOperation` covers the manual case today).
