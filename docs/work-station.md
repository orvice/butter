# Work Station read-only preview

PR #490 implements the shared projection for [Spec #489](https://github.com/orvice/butter/issues/489), tracked by KON-55. GitHub Issues/PRs, GitLab Issues/MRs, Linear Issues and Plane v2 work items share the connection, synchronization, query and Dashboard paths. **This is a development preview, not a qualified release.** All six real target profiles and the deployed two-Pod gate remain BLOCKED. Synthetic responses and independent local processes do not satisfy those gates.

## Connections and sharing

Workspace owners and global administrators create connections. Ordinary members and workspace administrators can browse and request refresh; they cannot discover candidate resources, edit credentials or select scopes. The server rechecks workspace membership on every operation. Global-admin access is audited.

| Provider | Approved endpoint source | Smallest selected scope |
| --- | --- | --- |
| GitHub.com | Existing global Git Host: `https://api.github.com`, web `https://github.com` | Stable repository ID |
| GitLab.com / self-managed | Existing global Git Host, HTTPS `<web root>/api/v4` | Stable project ID |
| Linear official | Fixed `https://api.linear.app/graphql` | Team ID; Linear Project remains metadata |
| Plane Cloud / self-managed | Built-in `plane-cloud` or global-admin registered Plane Host, fixed v2 profile | Project ID plus explicit external workspace slug |

Open **Work Station → Manage connections**, choose a platform, approved Host where applicable, connection name and API key. Plane additionally needs its external workspace slug. Creation selects no scopes. Discovery only lists candidates; saving an explicit full scope selection starts backfill. Imported titles and descriptions are shared with **every Butter workspace member**, including people without source-platform access.

Keys are encrypted and write-only. Omitted key updates keep the key, nonempty updates replace it and empty strings clear it. Replacement/Test must retain the pinned provider, instance, external tenant and principal. Another identity needs a new connection. Rotation/clear revokes old snapshots in the connection's atomic document. Test explicitly revalidates a paused connection and queues fresh snapshots. Re-saving a selected hidden scope revalidates only that scope under a new generation, preserving healthy siblings. Old hidden snapshots never reappear.

Plane API keys inherit their creator's permissions; the UI explains that the adapter's read-only behavior does not make the credential itself read-only. Use a dedicated minimum-permission account. Source comments, reviews, diffs, writes and agent execution are outside this feature.

## RPC and Dashboard contract

`agents.v1.WorkStationService` and `WorkStationHostService` are ConnectRPC services under `/api`. Responses, including errors, use `Cache-Control: no-store`. Readable messages contain no credential ciphertext or plaintext.

| RPC | Authority | Behavior |
| --- | --- | --- |
| ListConnections / GetConnection | member | Saved configuration/scopes and sanitized coverage/progress |
| Create / Update / DeleteConnection | owner/global admin | Empty initial selection; revision CAS; optional-key keep/replace/clear; atomic revocation |
| TestConnection / DiscoverScopes / SetSelectedScopes | owner/global admin | Identity revalidation; candidate pagination; authorized full selection replacement |
| Refresh | member | Durable intent; QUEUED, RUNNING, ALREADY_PENDING, DEFERRED, PAUSED or NO_SCOPES |
| ListItems / GetItem | member | Local projection with the same source/window qualification |
| ListFilterOptions | member | Projection-derived provider, scope, project, status and assignee options; literal search and keyset pagination |
| Host List/Get | authenticated | Approved Plane endpoint choices; no workspace key |
| Host Create/Update/Delete | global admin | Fixed v2 profile, HTTPS validation, revision and connection-reference guards; Cloud cannot be edited/deleted |

The three object tabs are Issues, Pull requests and Merge requests. Default results are open work. Filters include provider, connection, scope, project, normalized status, assignee and literal title/identifier keyword. List cursors bind workspace, every filter and page size (50 default, 100 maximum). Facet cursors also bind dimension/search. Candidates never come from upstream discovery. The legacy unpaged filter RPC shape remains readable with its original protobuf field numbers; new paged fields are additive. Previously published scope progress/checkpoint field numbers are preserved.

Dashboard transports are pinned to user/token/workspace. Identity changes remount the feature and cancel/remove its private queries. Requests carry abort signals. Lists, details, health and facets refresh periodically; a failed detail response hides cached body content. Descriptions render as text, without executing HTML or fetching embedded images/private attachments. Source links must remain inside the approved connection web origin/path. Keys are never refilled or persisted in localStorage. Global Plane Host administration lives in the Admin Git Hosts area, in a separate global-admin dialog; it is absent from workspace connection settings.

The previous app-level provider configuration remains a validated default; actual workspace endpoint selection now uses approved Host registries and each connection's provider/tenant. The explicit `providerregistry.NewService` / handler construction seam remains available for tests and fixed-profile integrations. Plane construction through that seam now requires an explicit workspace slug. Legacy global endpoint configuration does not silently introduce a new workspace Host.

## Storage and publication

`internal/workstation` owns visibility, identity, retention, claims and publication. Adapters expose only read capabilities. Production HTTP clients disable proxies/redirects, enforce HTTPS origin/method restrictions, reject private/reserved DNS answers at dial time and dial the checked IP with normal TLS verification. Publicly reachable self-managed instances are supported; private-network approval exceptions are not implemented. Invalid schema, unsafe paging, unsupported profiles and resource caps fail closed, rather than returning empty success. Errors/logs contain sanitized categories, operation and cause type, never upstream bodies or keys.

Mongo collections hold connection documents, immutable observations, approved Plane hosts and shared request budgets. Existing optional checkpoint-store APIs remain compatible; the authoritative runtime checkpoint and pending-page registration live in the same connection document as the visibility epoch, scope generation and published pointer. This avoids using a stale sidecar cursor to authorize data.

Every publication/error/progress write compares version, epoch, generation and claim; Mongo additionally requires the captured lease to exceed server `$$NOW` at the write. A connection deletion atomically removes ciphertext and visibility. Selected-scope removal, key rotation and takeover fence old success, progress and error writes.

Full scans are resumable for all four adapters. Each complete page is staged under a unique token, then its cursor/token is registered with a fenced CAS. A crash before registration leaves only an orphan; a later worker resumes the registered cursor without trusting unregistered rows. At most ten batches occupy one lease, after which health becomes LAGGING and the saved scan is queued for continuation. Final completion is persisted too, so a crash between the last page and publication does not replay the scan. Network/rate-limit failures retain progress; invalid cursor/schema/profile errors abandon it. Progress older than 24 hours restarts. Pending snapshot references are protected by GC, including paused/hidden records.

Only the completed traversal can replace a scope snapshot. Staged rows are deduplicated by the latest complete source observation before end-time/90-day filtering. A newer unresolved terminal marker cannot resurrect an earlier open version. Per-cursor metadata is bounded to 4 MiB and total pending connection JSON to 8 MiB, below Mongo's document ceiling; excess fails without publishing. Adapter limits can be smaller (GitHub/Linear cursor metadata 16 KiB; GitLab 500 pages / 50,000 rows per endpoint; Plane 1,000 pages). Bodies remain separate observation documents.

Canonical identity includes workspace, provider, canonical API instance, external tenant and kind. GitHub/GitLab qualify local number/IID by stable repository/project ID. Linear/Plane globally unique item IDs retain identity across container moves. Existing GitHub/source-less preview IDs stay unchanged. Plane's documented v2 schema does not expose an immutable workspace ID: its explicit `workspace-slug:<slug>` namespace is a known qualification limit; slug changes require a new connection.

List/Get/facets restrict valid source pointers first, choose one whole newest observation, then apply content/window filters. Revoking the newest source falls back to the remaining source's complete body, without mixing fields. Mongo starts with indexed exact snapshot pointers and rechecks live connection/scope eligibility before grouping. Orphan history cannot inflate that grouping input. Keyset queries reauthorize every request and do not promise a historical snapshot during concurrent refreshes.

## Scheduling, budgets and cleanup

The reconciler checks durable due state every ten seconds, with at most four attempts per process. Healthy scopes become due after five minutes plus deterministic jitter. Claims last two minutes and each attempt, including staging, is limited to 100 seconds. Accepted Refresh survives process loss; manual requests have a 60-second connection cooldown and honor connection and shared Retry-After state.

Production request admission is enabled through the approved-provider resolver: Mongo server-time reservations aggregate instance and account budgets across workspaces/connections/keys. The local pacing is conservative (instance 100 ms, identified account 500 ms per request); provider quota headers/GraphQL errors can impose longer shared backoff. Unknown-principal probes use a conservative shared anonymous bucket. This is not evidence that every provider complexity quota is tuned for production. Legacy `MemoryBudget` remains a single-process test seam and is not the runtime coordinator.

GitHub uses Issues `since` overlap, PR updated overlap plus all-open enumeration, missing cached-open checks and daily full reconciliation. Its bounded ETag cache isolates full URL/method/media/API version/credential digest and replays pagination metadata; a first-page 304 never skips later pages. Cache entries are bounded by count/bytes/TTL and scoped to a connection epoch. Linear/Plane/GitLab currently reconcile with resumable complete traversals rather than optimized delta queries. This preserves archive/move coverage but still needs representative throughput/scale qualification and further incremental optimization.

Successful publication prunes exactly the replaced snapshot token. Delayed cleanup cannot delete a newer publisher or another batch. Cleanup failure reports CLEANUP_PENDING without downgrading committed data. Fallback GC drains 1,000-row batches under a 30-second context, at most once per minute after reconciliation. It retains all committed and registered-pending references, even paused/hidden ones; unregistered orphans get a four-minute grace, exceeding the lease. Transient collection failures get up to three attempts with bounded backoff/deadline planning; a later lifecycle pass remains the fallback. Storage outages can still accumulate physical orphans.

A Host mutation racing connection creation is not transactionally serialized across documents. Reads and attempts compare the approved endpoint against the pinned source and fail closed after a change. Administrators must repair or remove affected connections. GitHub repositories with Issues disabled and GitLab projects missing either Issues or MRs currently mark the combined scope FEATURE_UNAVAILABLE. Automatic recovery of hidden scopes remains an explicit owner revalidation operation.

## Validation and ticket status

See [the qualification ledger](work-station-acceptance.md) for B01–B16, opt-in real read probes, the six target evidence records and deployed two-Pod procedure. Provider contract/source references are recorded in the [GitHub](work-station-github-profile.md), [Linear](work-station-linear-profile.md), [GitLab](work-station-gitlab-profile.md) and [Plane](work-station-plane-profile.md) profiles.

Local evidence includes shared memory/actual-Mongo contracts, independent OS-process takeover/restart/budget tests, adapter multi-page/error/schema fixtures, batch restart/GC protection, real Connect HTTP create/select/reconcile/List/Get/facet smoke tests, and 14 browser journeys with intercepted RPCs. The Work Station UI workflow now runs build/types/browser contract journeys in CI; a local pass does not claim the new workflow has run. Required commands:

```sh
go test ./...
go build ./...
BUTTER_TEST_MONGO_URI=mongodb://127.0.0.1:27028 go test -race ./internal/workstation/...
buf lint proto/agents/v1/workstation.proto
cd front
npm run build
npm run typecheck:e2e
npx eslint src/features/work-station src/api/work-station.ts src/routes/_authenticated/work-station.tsx
PLAYWRIGHT_PORT=5285 npx playwright test e2e/work-station.spec.ts --workers=1
```

| Ticket | Included in PR | Still required before acceptance |
| --- | --- | --- |
| KON-56 | Official-source fixed-profile notes, synthetic fixtures and opt-in qualification runner | Six real target versions, schema/permissions, sanitized captured fixtures |
| KON-57 | Connection CRUD/Get/Test/discovery/selection, paged facets, Host CRUD, roles, encrypted-key CAS | Full RPC/membership/Host matrix on deployed targets; Host race limitation review |
| KON-59 | Durable intents/claims, page resume, atomic checkpoint/progress, shared Mongo budgets, source-safe query/facets, reclamation | Broader delta optimization, representative scale/explain and deployed failure qualification |
| KON-60 | Linear organization/Team identity, archived scoped reads, status/end-time/project metadata, GraphQL partial-error handling, page resume | Real Linear Team/project/archive/move/reopen matrix |
| KON-61 | GitHub Issues/PR mapping, overlap/checkpoint, scoped ETag, all-open reads and full reconciliation | Real private-repo/permissions/revocation/read-only audit and scale |
| KON-62 | GitLab cloud/self-managed v4, project Issues/MRs, Draft/Merged/time fields, guarded pagination, page resume | Real cloud/self-managed versions/tokens/deep-page/permission matrix |
| KON-63 | Plane approved Hosts/v2, full cursor traversal, archive/draft/description metadata, page resume | Real Cloud/self-managed v2; reliable cancellation activity and immutable tenant-ID qualification |
| KON-64 | Four-provider connection UI, three object tabs, local facets, global Host admin, cache isolation, progress and browser CI definition | Real role/cache/source-revocation journeys and final product acceptance |
| KON-58 | B01–B16 ledger, local Mongo multi-process and failure tests, all-provider live read runner | Controlled real datasets, actual deployed two Pods/fault recovery and release decision |

No ticket is closed by local implementation or mock evidence. Record actual operator/date/version/profile/grants, expected/actual source state and a sanitized GET/query request audit for every real run. Credentials and private source bodies must not enter committed evidence.

The previous P1 observation-GC and P2 recovery/canonical-identity fixes remain. P3 choices remain explicit: NO_SCOPES for empty selection, filter/page-size-bound cursors, redacted operational logs, bounded connection-filtered refresh, safe text descriptions and separate HTTP/browser/real-platform evidence. Optimized non-GitHub deltas, private-network policy exceptions, representative scale and actual platform/two-Pod acceptance are not claimed complete.
