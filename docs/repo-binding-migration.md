# Repository binding migration & operations guide

This guide covers moving a workspace's Agent Content between database-managed
storage and a bound Git repository: the staged rollout, the onboarding
(export/import) and offboarding (safe detach) operations, how to verify a
migration, and cross-provider acceptance. It is the operational companion to
[ADR-0007](adr/0007-workspace-repo-onboarding-and-safe-detachment.md) and the
epic in issue #210.

## Rollout order (release guidance)

Git-backed Agent Content depends on every Agent having a stable, immutable
**Agent ID** and on the Agent V2 model (independent entities, ID-based
composition). Roll out in this order:

1. **Compatibility release (#220/#221).** Ship Agent ID assignment and the
   migration-readiness UI **first**. Agent slug/ID assignment begins *after*
   this compatibility release and *before* the Agent V2 cutover — not before.
   Runtime behavior is unchanged in this phase; administrators assign IDs and
   inspect readiness.
2. **Agent V2 cutover (#222).** Migrate ready agents and references to the
   ID-based model.
3. **Repository binding & content (#214–#219).** Only now may a workspace bind a
   repository and export or import Agent Content addressed by
   `agents/{agent-id}/`. Because content paths are keyed by Agent ID, a
   workspace cannot onboard Git content until its agents have IDs.

Rollout is **independent per workspace**. Binding one workspace never requires
any other workspace to migrate; each owner/admin chooses their own schedule.

## Onboarding a workspace

Prerequisites: a bound repository (`PutWorkspaceRepoBinding`), a stored PAT
(`SetWorkspaceRepoBindingCredential`), and a passing validation
(`ValidateWorkspaceRepoBinding`, state `OK`).

Then call `OnboardWorkspaceRepository` with one mode:

### `EXPORT_CURRENT`

Writes the workspace's existing database-managed Agent Content to the
repository as a single validated commit, reads it back, validates it, builds
the Effective Agents, and publishes the Active Revision.

- Field mapping: `description` → `agents/{id}/description.md`,
  `config.instruction` → `agents/{id}/prompt.md`, `config.global_instruction` →
  `agents/{id}/global-prompt.md`.
- Full snapshot, not a merge: non-empty fields are written (PUT); an empty field
  whose managed file already exists remotely is deleted (DELETE) so Git ends up
  reflecting the database exactly. When there is nothing to export and nothing
  stale to clear, the repository is **not** silently adopted (that is IMPORT);
  the workspace stays database-owned.
- Always a **direct commit**, even for CHANGE_REQUEST bindings, so onboarding
  can publish immediately.
- On a validation failure (e.g. an LLM agent with a description but no prompt),
  **no commit is created**, the Active Revision is unchanged, and
  `validation_errors` explains why. The workspace stays database-owned; fix the
  content and retry.

Response: `commit_sha`, `published`, `agents_exported`, `validation_errors`.

### `IMPORT_REPOSITORY`

Adopts the repository's existing Agent Content. Only directories matching an
existing workspace Agent ID are imported; **unknown directories remain
unclaimed** and no bidirectional merge is performed. Equivalent to a sync +
publish.

Response: `published`, `agents_imported`, `validation_errors`.

A workspace switches to Git-owned content (its `active_commit_sha` advances)
**only after** the commit/import, read-back, validation, Effective Agent
construction, and Active Revision publication all succeed.

## Offboarding (safe detachment)

`DeleteWorkspaceRepoBinding` safely detaches a workspace:

1. Materializes the Active Revision snapshot back into database-managed Agent
   fields (`description` / `instruction` / `global_instruction`).
2. Reloads the runtime so it runs from the now-authoritative database content.
3. Removes the binding, its encrypted PAT, the repository cache, and the content
   snapshot.

Detachment **never modifies remote Git content**. When there is no valid Active
Revision snapshot to materialize, detach **refuses** unless a recovery path is
selected:

- `recovery = KEEP_DATABASE` — keep current database content as-is and remove
  the binding.

Response: `agents_materialized` (how many agents had content written back).

## Migration verification checklist

After **export**:

- `OnboardWorkspaceRepository` returned `published = true`, empty
  `validation_errors`, and `agents_exported` equal to the number of agents with
  content.
- `GetWorkspaceRepoBinding` shows `active_commit_sha == observed_commit_sha` and
  `last_publication_error` empty.
- `GetRepositoryFile` for each `agents/{id}/prompt.md` returns the expected
  content; `ListRepositoryEntries` shows the agent directories as `claimed`.
- A test invocation of each agent produces the same behavior as before.

After **import**:

- `agents_imported` matches the number of matching agent directories; unknown
  directories are listed but `claimed = false`.
- The published snapshot contains only matching Agent IDs.

After **detach**:

- `agents_materialized` matches the previously active agents.
- Each agent's database `instruction`/`description`/`global_instruction` equals
  the last active Git content.
- `GetWorkspaceRepoBinding` returns no binding; the credential is gone.
- Remote Git history is unchanged (no new commits from the detach).
- A test invocation of each agent still produces the same behavior.

## Failure recovery

- **Export validation failed.** No commit was made; fix the offending content
  (the `validation_errors` name the Agent ID and path) and retry. The workspace
  remains database-owned.
- **Provider/PAT outage during sync or publish.** The binding enters `DEGRADED`;
  the last-known-good Active Revision keeps serving. Restore access and retry
  the onboarding call.
- **Detach aborted mid-materialization.** The binding is only removed after
  materialization and reload succeed, so a failed detach leaves the binding
  intact and the operation is safe to retry.
- **No snapshot to materialize.** Publish an Active Revision first, or detach
  with `recovery = KEEP_DATABASE`.
- **Upgrade from a release that kept Active Agent Content in process memory.**
  The startup reconciler immediately checks existing bindings. When the raw
  repository cache still matches `active_commit_sha`, publication rebuilds the
  missing Mongo snapshot from that cache even when the Git HEAD SHA is
  unchanged. `GetAgent` and runtime reload return an explicit error until the
  repair completes instead of silently falling back to stale database content.

## Cross-provider acceptance

The onboarding, offboarding, sync, publication, and commit paths all run through
the provider-neutral `internal/gitprovider` contract, exercised against both the
GitHub and GitLab REST adapters by `internal/gitprovider/contract_test.go`.
Before declaring a provider generally available, run the full onboarding →
verify → edit → rollback → detach cycle against a real repository on that
provider in both DIRECT_COMMIT and CHANGE_REQUEST write modes.
