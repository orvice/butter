# Work Station qualification ledger

Updated 2026-11-04 for KON-55 / PR #490. Release verdict: **not qualified for
the complete Work Station Spec**. Local implementation and fixture tests are
reviewable; real source-platform and deployed two-Pod gates remain open. No
child issue becomes accepted merely because the corresponding adapter compiles.

## Evidence levels and local commands

The shared repository contract runs against memory and, only when
`BUTTER_TEST_MONGO_URI` is set, a real Mongo server. A skipped Mongo test is not
Mongo evidence. `mongostore/multiprocess_test.go` starts independent copies of
the Go test binary, each with its own Mongo client and service, using one shared
test database. It proves competing claim admission, takeover, late success/error
fencing, durable Refresh/cooldown, process termination/restart, and shared
instance/account budget admission/backoff. A deterministic database lease-expiry
injection avoids waiting the production two-minute TTL. The expired publication
predicate additionally exercises real server `$$NOW`.

Those processes are **not deployed Kubernetes Pods**. The test does not prove
deployment networking, secret distribution, scheduler/termination grace periods,
real provider traffic or infrastructure failover. Redis is not used by this
Work Station coordinator; claims and provider budgets are persisted in Mongo.

```sh
go test ./...
go vet ./...
BUTTER_TEST_MONGO_URI=mongodb://127.0.0.1:27028 go test -race ./internal/workstation/...
cd front
npm run build
npm run typecheck:e2e
PLAYWRIGHT_PORT=5285 npx playwright test e2e/work-station.spec.ts
```

Use an isolated disposable Mongo database/server: test databases have unique
names and are dropped after execution. Keep infrastructure credentials out of
committed manifests and shell output.

## B01–B16 evidence map

This maps test entry points to the required behavior; it is not a claim that
every clause has passed on real platforms. The final PR validation record must
name the executed commands and distinguish skipped live tests.

| Gate | Automated evidence in this PR | Remaining qualification |
| --- | --- | --- |
| B01 workspace boundary | `application/workstation_service_test.go`: membership/workspace rechecks; shared `repotest/contract.go`: staging/workspace isolation and query-token binding | Live role/membership matrix and all newly added RPCs must remain in the same authorization boundary |
| B02 role/Host policy | `TestWorkStationPermissionsAndWriteOnlyCredentials`, `TestPlaneHostAuthorityAndRevision`; owner-only connection changes, global-admin Host changes | Real member/admin/owner/global-admin UI journeys |
| B03 credential/revision | shared `config_CAS_and_secret_seam`, application credential tests, adapter redaction tests, `report_test.go` | Deployed logs/traces inspection; actual key rotation and permissions |
| B04 selected scopes only | shared staging/discovery isolation, unauthorized scope handling; provider-scoped resolution | Real private scope discovery/revocation and facet candidates |
| B05 provider traversal | `github`, `linear`, `gitlab`, `plane` adapter fixtures cover pages, failure/schema/error classes; Plane 10,001-row cursor traversal; Linear partial GraphQL errors | Target-sourced sanitized fixtures and exact deployed schema/version for six target profiles |
| B06 issue/PR/MR rules | GitHub separation/merged/draft/unknown-state tests; GitLab normalization fixtures | Controlled private repository/project source-UI comparison |
| B07 retention and end time | `TestRetentionExactBoundary`, shared window contract, adapter old-open/reopen/end-time fixtures | Real reopen/reclose lifecycle timestamps; Plane cancellation remains PARTIAL pending verified activity profile |
| B08 archive/draft/move | Plane full archive query/local draft exclusion; Linear archived/team-move tests; failed traversals never publish | Target archive-without-updated change, movement/deletion and suspected-missing lifecycle matrix |
| B09 identity/dedup | `identity_test.go`, shared whole-observation choice/source fallback; exact opaque/numeric provider IDs | Plane immutable external workspace ID/slug rename qualification remains open |
| B10 source visibility | shared `source_choice_filters_and_immediate_fallback`; Mongo pipeline joins current snapshot eligibility before content choice/filtering | Real revocation with List/Get/Search/FilterOptions and open details in other sessions |
| B11 failure handling | shared failure-preserves-snapshot/revocation, reconciler isolation tests, provider rate-limit/error fixtures | Real 401/403/404/429 and inaccessible-resource behavior across target versions |
| B12 fencing/recovery | shared late publish/error/epoch/key/scope/delete tests; real-Mongo independent-process takeover and late writes | Actual two deployed Pods, real lease duration, shutdown/restart/network failure observations |
| B13 durable intent/budgets | controlled-clock reconciler and durable-batch resume tests; independent processes share Mongo Refresh/cooldown/budget state across restart | Actual two-Pod scheduling/cross-workspace account budgets and provider Retry-After |
| B14 reclamation | memory+Mongo shared reclamation contracts: physical rows, delayed cleanup, >1,000 orphan batch drain, current/hidden/other-source preservation | Long-running deployment footprint and crash-during-staging observation |
| B15 outbound/UI safety | confined provider client, fixture unsafe URL/path/redirect checks, no-write audit; Dashboard renders descriptions as text | Deployment DNS/egress policy; real source links and no external image/private attachment fetch inspection |
| B16 query pagination | shared source choice/window/keyset/filter-bound tokens; Mongo projection matches memory behavior | Static real dataset pagination comparison; dynamic result sets do not promise an unimplemented read snapshot |

## Live target read runner

`internal/workstation/acceptance/profile_test.go` is an opt-in real adapter read
runner for all four providers. It probes identity, resolves a controlled scope,
scans it, checks configured item/state/draft/description expectations and counts
GET or named GraphQL query requests. Other HTTP verbs and GraphQL mutations are
rejected by the audit transport. Actual DNS/TLS/redirect confinement is retained.
No token, upstream body, scope name, private item title or description is printed.

Provide a private UTF-8 JSON manifest through
`BUTTER_WORK_STATION_ACCEPTANCE_MANIFEST`. Token values are referenced through
environment variable names, never stored in the manifest. For example:

```json
[
  {
    "name": "gitlab-selfhost-controlled",
    "provider": "gitlab",
    "api_base": "https://gitlab.example/api/v4",
    "web_base": "https://gitlab.example",
    "scope_id": "123",
    "credential_env": "WORK_STATION_ACCEPTANCE_GITLAB_TOKEN",
    "version": "record the actual target version before running",
    "profile": "gitlab-v4-offset",
    "minimum_items": 1,
    "expected_coverage": "COMPLETE",
    "expected": [
      {"remote_id": "1", "kind": "ISSUE", "status": "NOT_STARTED", "description_available": true},
      {"remote_id": "2", "kind": "MERGE_REQUEST", "status": "MERGED", "draft": false}
    ]
  }
]
```

```sh
go test -v ./internal/workstation/acceptance -run TestRealProviderProfiles -count=1
```

Use `provider=linear` or `github` for their fixed official origins, and
`provider=plane` plus `workspace_slug` for approved Cloud/self-hosted API/web
bases. Optional `tenant_id` pins an expected external tenant. Use exact remote
IDs emitted by the adapter, not display identifiers. `absent=true` can assert
that an old ended item or draft is excluded. Set `minimum_items=0` explicitly
only for an intentionally empty controlled dataset. Passing a manifest name or
claimed version alone does not establish the version: attach independent
nonsecret version evidence to the record below.

The read runner does not mutate source systems. Controlled state changes,
archive/reopen/revocation and credentials must be prepared through the source
platform's authorized UI/operations and recorded separately. The adapter's
GET/query audit is evidence of what this runner sent, not proof of the token's
underlying grants. One successful read is not the whole acceptance matrix.

## Required real-platform records

| Target | Target/version/profile evidence | Credentials and dataset | Current result |
| --- | --- | --- | --- |
| Linear official | Official schema revision/date and organization/team IDs | Dedicated read/team-limited key; archive/move/reopen and GraphQL error datasets | Not run: credentials and controlled target absent |
| Plane Cloud | Exact accepted v2 shape and endpoint evidence | Dedicated account key (inherits account privileges), workspace/project lifecycle dataset | Not run: credentials and controlled target absent; cancellation profile unresolved |
| Plane self-hosted | Named instance, installed version/build, matching v2 endpoint/schema; independently qualify v1 if required | Approved API/web bases and dedicated account/project | Not run: instance/version/credentials absent |
| GitHub.com | REST version and endpoint/profile record | Private repository, least-privilege issue/PR reads, real 401/403/404/429/revoke | Not run: controlled private repository/credentials absent |
| GitLab.com | Actual `/version` evidence and v4 pagination capability | `read_api` personal/group/project token; private Issues/MRs | Not run: credentials and controlled target absent |
| GitLab self-hosted | Named instance/version and offset/deep-page limits | Approved API/web bases, `read_api` token; no job/deploy token substitution | Not run: instance/version/credentials absent |

For every target, record: timestamp/operator, approved origin and profile,
nonsecret credential grant/role summary, controlled dataset action timeline,
expected item identifiers/state/end timestamps, sanitized response fixture
provenance, request-method audit, actual outcomes and remaining incompatibilities.
Never put live keys or private descriptions in the PR evidence. The Plane
profile's immutable workspace identity and reliable cancellation-time limits
are described in `work-station-plane-profile.md`.

## Actual deployed two-Pod gate

Record one deployment with two distinct Pod identities using the same Mongo
database and keyring. Include the deployment image/commit, Mongo version,
nonsecret configuration identity and timestamped operator observations. Run:

1. Submit simultaneous Refresh requests through both Pods; observe one durable
   intent and cross-Pod cooldown without falsely declaring source refresh done.
2. Start a controlled multi-page scan on Pod A, terminate A during staging,
   expire its real lease, and observe Pod B take over and publish a complete
   result. Retain evidence that intermediate staging was never visible.
3. Delay A's old success and error paths, rotate a credential/change selection
   or publish B's successor, and show neither stale success nor stale failure
   can overwrite/hide B's state. Repeat across process and Mongo restart.
4. Exercise a shared account through connections in two workspaces; show
   instance/account admission and upstream Retry-After survive takeover and
   restart. Show another independent account/instance remains isolated.
5. Revoke one source of a duplicate item; verify immediate server-side fallback
   for list/detail/search/facets and subsequent client detail/cache cleanup.
6. Inspect physical observations after repeated scans and abandoned staging;
   current snapshots and other connections remain intact, and expired orphan
   batches drain.

Current deployed two-Pod result: **not run**. There is no supplied deployment,
controlled provider target or deployment credential in this workspace. Local
Mongo processes are useful atomicity/restart evidence and must be reported with
that narrower scope. These external requirements keep KON-58 and related
source-platform acceptance gates open.
