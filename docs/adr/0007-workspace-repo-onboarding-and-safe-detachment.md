# Workspace onboarding/offboarding for repository-owned Agent Content

Issue #219 (PRD Phase 4) completes the Git-backed Agent Content lifecycle from
issue #210. The binding, sync, cache, publication, editing, and Saga seams
(#214–#218) already let a bound workspace treat Git as the source of truth. What
remained was the *transition in and out* of that state: how a workspace that has
only ever stored Agent Content in the database first adopts Git, and how it
safely stops.

## Decision

Model onboarding and offboarding as **two explicit, one-shot operations on the
existing publication and materialization seams** rather than new subsystems.

### Onboarding — `OnboardWorkspaceRepository(mode)`

A single owner/admin RPC with a mode selector reconciles a freshly bound
workspace's existing Agent Content with the repository:

- **`EXPORT_CURRENT`** reads every non-deleted agent's database-managed content
  (`description` → `description.md`, `config.instruction` → `prompt.md`,
  `config.global_instruction` → `global-prompt.md`), builds a single changeset,
  and reuses the shared `commitContent` path (commit → sync → validate →
  publish). Empty fields are omitted rather than written as empty files; the
  publication pipeline treats a missing optional file as a cleared value, so the
  round trip is lossless. Export **always commits directly** (DIRECT_COMMIT)
  even when the binding's write mode is CHANGE_REQUEST: onboarding must publish
  an Active Revision to switch the workspace to Git-owned content, and a PR/MR
  would leave content unpublished behind an open review. This mirrors the
  lifecycle Saga, which forces DIRECT_COMMIT for the same reason ([[0006]]).

- **`IMPORT_REPOSITORY`** runs the existing sync + publish. The publication
  pipeline parses content only for directories whose name matches an existing
  workspace Agent ID, so unknown directories stay **unclaimed** and no
  bidirectional merge happens.

In both modes the workspace only switches to Git-owned content — i.e.
`active_commit_sha` advances — after the commit/import, read-back, and content
validation succeed. The publication pipeline (#216) advances the Active Revision
only when `agentcontent.Validate` passes, so a failed validation leaves the
workspace on its previous database-owned state with per-agent errors reported;
there is no partial switch to unwind. The final Effective Agent build happens in
the subsequent `ReloadRunner`, which is **best-effort by #216's design**: a
reload failure keeps the last-known-good runner serving rather than reverting
the Active Revision, mirroring the DEGRADED-but-serving posture. Onboarding
inherits that contract unchanged; it does not add its own runner-build gate.

### Offboarding — safe detachment on `DeleteWorkspaceRepoBinding`

Unbinding is the inverse of publication. Before removing anything, delete
**materializes** the Active Revision snapshot back into database-managed Agent
fields (the inverse of the `ApplyToProto` overlay) and reloads the runtime, so
runtime behavior is preserved once Git is no longer the source of truth. Only
then are the binding, its encrypted PAT (removed atomically by the binding
repository's `Delete`), the repository cache, and the content snapshot removed.
Detachment **never modifies remote Git content** — it is a pure control-plane
operation.

When there is no valid Active Revision snapshot to materialize (never published,
snapshot missing, or snapshot SHA ≠ `active_commit_sha`), detach **refuses** by
default (`FailedPrecondition`) so a missing snapshot can never silently drop an
Agent's live content. The caller may override with an explicit
`RepoBindingDetachRecovery`:

- `KEEP_DATABASE` — keep the current database content as-is and remove the
  binding without materializing from Git.

Git binding rollout stays **independent per workspace**: onboarding and
detachment are per-workspace RPCs keyed off the `X-Workspace-ID` header and
touch no other tenant, so workspaces migrate on their own schedule.

## Consequences

- **No new persistence.** Onboarding/offboarding compose `commitContent`,
  `publishActiveRevision`, the content-snapshot repository, and the agent
  repository. The only new proto surface is one RPC, two enums, and additive
  request/response fields.
- **Export bypasses CHANGE_REQUEST review by design.** Establishing the initial
  Git content is a provisioning action, not an edit; subsequent edits go through
  `CommitAgentContent` and honor the binding's write mode.
- **Detach is honest about partial state.** Materialization is best-effort
  per-agent (a snapshot entry whose agent no longer exists is skipped), and the
  response reports `agents_materialized` so operators can reconcile. A snapshot
  read/agent write failure aborts the detach *before* the binding is removed, so
  a retry is safe.
- **Staged rollout guidance is a prerequisite, not code.** Agent slug/ID
  assignment begins after the compatibility release from the first child ticket
  (#220/#221) and before the Agent V2 cutover; only then can a workspace export
  or import content addressed by `agents/{agent-id}/`. See
  `docs/repo-binding-migration.md`.
