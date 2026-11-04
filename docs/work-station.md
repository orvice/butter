# Work Station: GitHub.com preview

This PR delivers the first end-to-end slice of [Spec #489](https://github.com/orvice/butter/issues/489), tracked by KON-55. It does **not** complete the four-provider release gate. The dashboard labels the feature as a preview. Linear, Plane, GitLab and production compatibility acceptance remain open.

## Using the preview

A global administrator first registers a Git Host with kind GitHub, API URL `https://api.github.com` and web URL `https://github.com`. Reuse an existing matching Git Host; no additional endpoint registry is introduced.

Open **Work Station → Manage connections** as a workspace owner or global administrator. Create a named connection with a fine-grained PAT that can read repository metadata, Issues and Pull requests. Discovery lists candidates but imports nothing. Explicitly select repositories and save. Their titles and descriptions are shared with **every member of the Butter workspace**, including members without source-platform access.

A member or workspace admin can browse the local projection, view details, filter and request refresh. Workspace admins cannot manage connections. All management RPCs independently enforce owner/global-admin permissions. Membership is checked again on every read. Root/global-admin bypass follows the existing application convention and is audited.

Keys are write-only. An omitted update keeps the key, a nonempty update replaces it, and `""` clears it. Rotation validates the same GitHub principal; another principal requires a new connection. Clearing/rotating a key hides its old snapshots immediately. Test connection is the owner recovery action: it rechecks the key and queues new snapshots, without restoring stale snapshots from a previous epoch.

Default browsing shows open Issues or Pull requests. All retained items includes terminal items with a verified end time in the last 90 days. The window uses `closed_at`/`merged_at`, never a later edit time. Descriptions are displayed as plain text, so HTML, images and Markdown links cannot execute or load private resources. Comments/reviews/diffs remain on GitHub.

## Protocol and storage

`agents.v1.WorkStationService` is mounted at `/api/agents.v1.WorkStationService/<Method>` using the existing ConnectRPC snake_case codec. Generated Go/Connect/TypeScript bindings are checked in. Every response, including errors, has `Cache-Control: no-store`.

| RPC | Permission | Behavior |
| --- | --- | --- |
| ListConnections | member | Only saved scopes and sanitized health; no candidate discovery or credentials |
| GetConnection | member | Reads one saved connection with the same sanitized fields and live Host/profile health |
| CreateConnection | owner/global admin | Probe, encrypt, create with empty selection |
| UpdateConnection | owner/global admin | `expected_revision` CAS; rename or optional key keep/replace/clear |
| DeleteConnection | owner/global admin | Revision CAS; atomic removal of visibility and ciphertext; idempotent retry |
| TestConnection | owner/global admin | Probe saved key, pin principal, reset visibility epoch and requeue |
| DiscoverScopes | owner/global admin | Paginated repository discovery, without importing |
| SetSelectedScopes | owner/global admin | Revision CAS; validate every stable repository ID; full replacement; empty means none |
| Refresh | member | Durable intent; QUEUED/ALREADY_PENDING/RUNNING/DEFERRED/PAUSED/NO_SCOPES |
| ListItems | member | Local source restriction → source choice → content/window filtering → keyset pagination |
| ListFilterOptions | member | Facets from visible local projections only; no remote discovery |
| GetItem | member | Same live-source/window authorization as ListItems |

The scope health projection now exposes `sync_phase` (`QUEUED`, `SCANNING`, `READY`, `ERROR`, or `WAITING_RETRY`) and `scanned_item_count`. The count is the number returned by the completed provider traversal for the latest committed attempt; it is an observability signal, not a remote total or a completeness claim. A failed attempt keeps the previous committed snapshot and reports the phase/error separately.

`ListFilterOptions` currently derives status, assignee and label facets from up to the first 1,000 visible local items for the requested kind/scope. It never calls a provider or exposes unselected resources; complete large-scale facets and progress counts remain part of KON-64/KON-59.

`internal/workstation` owns provider-independent rules. A read-only `Provider` has Probe, DiscoverScopes, ResolveScope and Scan. It cannot write the business repository or source platform. The GitHub adapter is constructed with a transport, never a workspace-supplied endpoint. It uses only GET, rejects redirects and unapproved next links, and checks public DNS destinations when opening production connections. Upstream response bodies and transport errors never become RPC errors or logs.

Mongo collections:

- `work_station_connections`: unique `(workspace_id, id)`, internal optimistic version, public configuration revision, visibility epoch, encrypted credential/key ID, selected scopes/generations, due times, claims and committed snapshot pointers.
- `work_station_observations`: one immutable source-qualified item per `(workspace_id, connection_id, scope_id, snapshot, item_id)`. Bodies are separate documents, rather than an unbounded array in the connection document.

Keeping bounded scope metadata and pointers in the connection document makes configuration revocation and scope publication share one atomic boundary. Every mutation compares the internal version; publication/error writes also compare the claim token and require the lease expiry to exceed Mongo `$$NOW` **at the write**. Takeover tokens and selection generations are never reused. A configuration race retries against the new record and rechecks the captured epoch/generation/token. The memory implementation runs the same contract.

Scans stage all normalized observations under a fresh snapshot token. Only a completed traversal can publish its pointer. Errors retain the previous pointer; invalid credentials pause/hide the connection, and an inaccessible repository hides that scope. Sources are never merged field by field: after restricting connection/repository filters, the latest valid observation supplies the whole item. Revoking the latest source falls back to the remaining source's body. Canonical identity includes workspace, persisted provider/API origin/external tenant, stable repository ID, item kind and repository number. Identity comes from the provider probe, with canonical origin normalization; it does not depend on a Host record ID, display name or token principal. Existing source-less GitHub preview records retain their original namespace and item IDs. Future providers/instances/tenants get distinct keys without changing the format.

Mongo List/Get first restricts observations to exact current `(workspace, connection, scope, snapshot)` pointers using the compound index, then joins the **current** connection again to reject intervening revocation/publication races. Retired/orphan history never enters the workspace-wide source grouping. It sorts source versions, groups by canonical identity, applies the rolling window/content filters, and then keyset pagination. The cursor binds workspace, filters and page size. Sources are reauthorized on each page. Queries do not promise a global historical snapshot while refreshes publish.

The reconciler scans durable due state every 10 seconds with up to four scans per process. Healthy scopes become due after five minutes plus deterministic jitter. Claims last two minutes; attempts, including staging, are bounded to 100 seconds. A crashed process leaves a durable due intent that another process can reclaim after expiry. Manual requests have a 60-second connection cooldown and cannot override saved Retry-After. After successful fenced publication, cleanup immediately deletes exactly the replaced snapshot token for that workspace/connection/scope. It never deletes “all except my new token”: a newer publisher or staging attempt may already exist. A cleanup error is reported as CLEANUP_PENDING without changing the committed pointer or marking the scan failed. The lifecycle cleanup pass retries orphan collection up to three times for transient storage errors; a persistent failure remains reportable and is retried by the next scheduled pass.

The fallback GC checks after reconciliation, at most once per minute, with a 30-second context. It drains 1,000-row batches until no expired orphan remains or its context ends; 1,000 is a batch size, not an hourly deletion budget. The four-minute orphan grace exceeds the two-minute lease and 100-second attempt, preventing a GC candidate from becoming a new valid publication. Every referenced snapshot is protected even if its connection is paused or its scope hidden. This handles crashed/failed cleanup, configuration removal and abandoned staging. A timeout retries on a later pass; sustained storage outages can still accumulate physical rows, while indexed current-pointer reads remain isolated from that backlog.

## Supported behavior and remaining limits

The adapter independently traverses Issues and Pull requests; Issues responses containing `pull_request` are excluded from the Issues projection. Merged wins over closed, Draft is separate, and unresolved terminal times produce PARTIAL coverage rather than fabricated history. Unknown/missing required schema and failed pages do not publish a truncated snapshot. GitHub's issue/PR distinction and list parameters follow the official [Issues](https://docs.github.com/en/rest/issues/issues) and [Pull requests](https://docs.github.com/en/rest/pulls/pulls) contracts.

A Linear GraphQL read-only adapter basis now exists in `internal/workstation/linear`: it fixes the approved endpoint, uses injected transport for fixtures, supports viewer probe, non-archived team discovery, team resolution and bounded issue normalization, and classifies HTTP/GraphQL errors without returning upstream bodies. Its GraphQL operations are query-only; pagination with `hasNextPage` is marked PARTIAL until a restartable cursor is wired. It is not registered as a production profile yet, and no real Linear instance or credential acceptance is claimed. A checkpoint binds an opaque cursor and ETag to a scan generation, uses a monotonic sequence within that generation, and is advanced with revision CAS. Stale writers and sequence regressions are rejected; changing generation is the explicit reset boundary. Its cleanup path is also revision-fenced: deleting a checkpoint with an old revision returns conflict, while deleting an already absent key is idempotent. This prevents a delayed cleanup from deleting a newer generation.

The local cleanup retry seam classifies successful/idempotent deletes and revision conflicts as terminal, and schedules other storage errors with capped exponential backoff until a deadline. It produces a plan only; a future Mongo/Redis scheduler must persist the task, honor `next_at`, and coordinate ownership across Pods.

A GitLab.com REST read-only adapter basis now exists in `internal/workstation/gitlab`: it fixes the `/api/v4` origin, uses injected transport for fixtures, supports identity probe, project discovery, project resolution and issue normalization, and classifies HTTP failures without returning upstream bodies. It uses `GET` requests only; `X-Next-Page` is exposed as a page cursor and scans with another page are marked PARTIAL until checkpoint continuation is wired. It is not registered as a production profile yet, and no real GitLab Cloud/self-hosted acceptance is claimed.

The reconciler has an optional checkpoint store hook that advances a completion boundary after a snapshot is committed. It records generation/sequence/time only; it does not consume or produce provider cursors. If this write fails, the committed snapshot remains visible, the scope exposes `checkpoint_lag=true` with a sanitized `checkpoint_error` category, and the run reports lag for retry. A later successful scan clears the lag fields. This is the only current wiring path and is deliberately insufficient to claim restartable incremental scans. Neither adapter is wired into the GitHub full scan yet: durable crash recovery, cleanup policy, and provider-specific cursor semantics still require integration design and validation.

This implementation uses bounded **full scans**, not the Spec's optimized incremental protocol. The provider has an optional process-local ETag cache seam keyed by the full request and a token hash; when enabled, a matching 304 decodes the exact cached body, while an uncached 304 remains fail-closed as a protocol/schema error. The cache has no TTL or cross-Pod coordination and is not a production release claim. There is still no incremental checkpoint to advance incorrectly. During each endpoint traversal it rejects repeated page signatures, repeated remote IDs across pages, cursor regressions and unapproved next links; a detected collection movement fails the attempt so the previous snapshot remains visible. Each endpoint is capped at 1,000 pages and each response at 16 MiB; exceeding a cap or the attempt deadline fails the scan and retains a safe existing snapshot. Large repositories need incremental/restartable staging before production use. Offset pagination under concurrent source edits still needs controlled platform tests and missing-item reconciliation; a completed HTTP traversal is not proof of a stable remote point-in-time snapshot.

Retry-After is persisted per connection. The reconciler also accepts an optional in-process shared budget keyed by provider/API origin/external tenant, so multiple connections in one process cannot consume more than the configured local window. A budget rejection uses the same `RATE_LIMITED`/`WAITING_RETRY` path and keeps the committed snapshot. This is only a local seam and does not provide cross-Pod coordination; production must replace it with Redis or another durable shared implementation before claiming the Spec's account/instance budget guarantee. Host reference guards reject normal endpoint changes/deletion while connections exist, and every read/attempt revalidates the fixed official profile. A cross-document Host mutation concurrent with connection creation is not transactionally serialized; an invalid/missing Host fails closed and requires administrator repair.

GitHub repositories with Issues disabled currently return FEATURE_UNAVAILABLE for the scope, including its Pull requests. Module-specific coverage, live progress counts, automatic background scope recovery and complete metadata facets need follow-up work. Saving the selected repository set revalidates access and requeues hidden scopes under new generations, without clearing healthy siblings. Old hidden snapshots are not restored, and old publishers remain fenced. Repositories no longer accessible are hidden on the next explicit failure; source revocation cannot be detected instantaneously. No source-platform writes, OAuth, Enterprise endpoints or agent execution actions are included.

## Verification and acceptance evidence

Automated evidence is local, using synthetic provider responses. It is **not real-platform acceptance**.

- Shared memory/Mongo contract: staging invisibility, dedup/source fallback, immediate revocation, key rotation, config CAS, stale publishers/errors, lease takeover, retention, page-size-bound cursors, scope-local recovery, bounded physical rows across repeated 1,001-item publications, delayed pruning/cleanup failures, and a 2,501-orphan multi-batch GC with current/staging protection.
- HTTP adapter tests: complete and failed pagination, malformed/foreign cursors, redirects, Issues/PR separation, Draft/Merged, large numeric IDs, missing terminal time, permission/rate-limit/schema/network errors and secret-free failures.
- Service/router tests: owner-only management, member/admin read/refresh, membership removal, workspace boundaries, no secret fields, official Host validation/reference guards and no-store route mounting. A real Connect client/handler HTTP smoke test exercises create → select → refresh → reconcile → list/detail through memory storage; it runs in the existing Go CI suite.
- Local browser fixtures (not a CI or real-platform acceptance claim): connection creation/selection/key clearing/deletion, no secret persistence, source HTML escaping, tabs/filter/pagination, refresh intent, non-owner management exclusion, failed details and late old-workspace responses, and connection-filtered refresh that skips empty/paused connections. Refresh queues only the selected connection, or all active selected connections when the filter is empty, with at most four concurrent intent RPCs.

Run:

```sh
go test ./...
BUTTER_TEST_MONGO_URI=mongodb://127.0.0.1:27017 go test ./internal/workstation/... -count=1
go test -race ./internal/workstation/...
buf lint proto/agents/v1/workstation.proto
cd front
npm ci
npm run build
npm run typecheck:e2e
npx eslint src/features/work-station src/api/work-station.ts src/routes/_authenticated/work-station.tsx
PLAYWRIGHT_PORT=5285 npx playwright test e2e/work-station.spec.ts
```

Real acceptance is still blocked by the absence of designated instances, least-privilege test credentials and a controlled dataset. No private credential was borrowed from the development CLI. Complete the existing tickets in dependency order:

| Ticket | Remaining evidence/work |
| --- | --- |
| KON-56 | Real Linear/Plane Cloud + self-hosted/GitHub/GitLab Cloud + self-hosted versions, permissions, schema, paging and sanitized fixtures |
| KON-57/59 | Review the partial RPC/storage contract; complete facets, shared budgets, incremental/restartable staging and scale/explain evidence |
| KON-61 | Real GitHub private-repo matrix, token permissions, access revocation, close/reopen, merge/Draft and read-only request audit; incremental/ETag strategy |
| KON-60/62/63 | Linear, GitLab and Plane adapters, approved Hosts/profiles, fixtures and real acceptance |
| KON-64 | Complete provider/facet UX, progress/health refinements and all full-Spec empty-state/E2E scenarios |
| KON-58 | Two actual Pods, restart/takeover/late-write and rate-budget exercises, all provider controlled datasets and final release decision |

For each real run record: date, operator, provider/profile/version, approved instance, masked principal/permission grant, selected stable scope IDs, controlled fixture counts, exact read-only steps, expected/actual result, sanitized request methods and failure categories. Explicitly mark PASS / PARTIAL / BLOCKED. Never store tokens or raw private bodies in the evidence. The four-provider release remains BLOCKED until all target profiles and multi-Pod gates pass.

## PR review follow-up

The P1 retention/query-cost issue is fixed by immediate predecessor pruning, draining orphan GC and indexed current-pointer reads. The physical-count regression fails on the original implementation (second 1,001-item scan produces 2,005 rows instead of 1,004) and runs against both memory and real Mongo.

Both P2 findings are fixed here: explicit selection save recovers a revalidated hidden scope in isolation, and persisted provider/origin/tenant identity preserves the original GitHub namespace while isolating future providers. No child ticket is closed by these corrections.

P3 decisions: NO_SCOPES is explicit; page size is bound into cursors; runtime logs include the failed operation, a safe category and underlying cause type rather than upstream bodies/credentials; a real HTTP Connect smoke test is covered by the existing Go CI suite; local browser fixture evidence is labeled separately. Refresh honors the current connection filter, skips paused/empty connections and bounds concurrent intent RPCs. Adding Playwright to CI, automatic background recovery, shared upstream account budgets and complete scale/explain acceptance remain in KON-59/KON-64/KON-58; the actual platform/two-Pod gates remain blocked.
