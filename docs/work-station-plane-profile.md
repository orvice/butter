# Plane Work Station read profile

Research date: 2026-11-03. Implementation: `internal/workstation/plane`.
Profile identifier: `plane-v2-cursor`. This is documentation-backed fixture
qualification; neither Cloud nor a self-hosted target has been accepted with
real credentials in this workspace.

## Fixed requests and permissions

The adapter uses only GET requests under the approved API base's `/api/v2/`.
It authenticates with `X-Api-Key`; the credential probe is `users/me/` and
requires an `api_key` principal. API keys inherit their creator's privileges,
so this adapter's read-only requests do not make the key itself read-only. Use
a dedicated account with the minimum necessary workspace/project access.
[Authentication](https://developers.plane.so/api-reference/v2/authentication),
[current-user response](https://developers.plane.so/api-reference/v2/users/get-current-user).

Project discovery calls `workspaces/{slug}/projects/` with
`include_archived=true`; selected project validation calls the corresponding
project detail. Both request `id,name`. Project IDs are opaque stable scopes;
names are display metadata, refreshed before and after a scan. A missing or
inaccessible project fails the attempt.
[Project listing](https://developers.plane.so/api-reference/v2/projects/list-projects),
[project detail](https://developers.plane.so/api-reference/v2/projects/get-project).

Work-item traversal requests `archived=true`, `is_draft=false`, `order_by=id`,
explicit `description_html` and relevant identity/time fields, plus
`expand=state,assignees,labels`. The local decoder also excludes drafts. No
updated-time cutoff is used: each completed traversal includes archived items
and old open items, since archive changes do not advance `updated_at`.
[Work-item listing](https://developers.plane.so/api-reference/v2/work-items/list-work-items).

## Pagination and normalization

Every collection request explicitly sets `paginate=cursor&per_page=100`.
The decoder requires `data`, `pagination.style=cursor`, `has_more`, and a
nonempty advancing `next_cursor` when more pages exist. Tokens are parameters,
never executable URLs. Offset/v1 envelopes, repeated cursors/rows, empty
continuation pages and traversal beyond 1,000 pages fail the complete attempt.
This bounded failure is reported rather than publishing a truncated snapshot.
[Pagination contract](https://developers.plane.so/api-reference/v2/pagination).

Expanded relation IDs must match their explicit ID arrays. Only user IDs and
display names, label names, and state ID/name/group are retained; upstream
email addresses/avatar URLs are discarded. Missing expansions fail schema
validation. Missing/null descriptions remain explicitly unavailable; supplied
HTML is stored as text and is not executed by this adapter.
[Expansion schemas](https://developers.plane.so/api-reference/v2/expanding-relations).

Workflow groups map to NOT_STARTED (backlog/unstarted/triage), IN_PROGRESS
(started), DONE (completed), CANCELED (cancelled), or OTHER. Names alone never
determine terminal status. DONE uses explicit `completed_at`; unknown terminal
times are omitted, increment unresolved count, and yield PARTIAL coverage.
[State groups](https://developers.plane.so/api-reference/v2/states/list-states).

The activities API documents generic field/value/identifier records, but does
not establish a verified cancellation/reopen event encoding for either target.
The adapter deliberately does not infer cancellation time from `updated_at`,
`archived_at`, or `completed_at`. Cancelled items currently require PARTIAL
coverage until a target-specific activity profile is verified and implemented.
[Activities schema](https://developers.plane.so/api-reference/v2/work-item-activities/list-activities).

## Identity and genuine acceptance blockers

The documented v2 current-user and project shapes do not expose an immutable
workspace ID. The implementation therefore explicitly namespaces a connection
by approved canonical API origin and `workspace-slug:<slug>`, then project ID
and item ID. This is a known qualification limit: a workspace slug rename needs
a new connection and is not represented as an immutable tenant migration.
No undocumented workspace lookup endpoint or v1 fallback is invented. Review
against actual target schema is required before changing that identity rule.

| Required target evidence | Current result |
| --- | --- |
| Cloud version/profile, accepted sanitized response fixtures | Not run: target credentials and controlled dataset absent |
| Named self-hosted instance/version with matching v2 schema | Not run: instance/version and credentials absent |
| Revoke/read-only audit and account permission evidence | GET-only fixture audit passes; real target audit not run |
| Deep traversal, archived/draft, descriptions and relation metadata | Local injected-transport regression suite passes |
| Reliable cancellation and reopen transition time | Blocked: no verified target activity encoding |
| Immutable external workspace identity and rename behavior | Blocked: documented v2 shape has no immutable workspace ID |

There is no claim that these fixture results satisfy KON-56, KON-63 or the
real-platform release gate in KON-58. For real acceptance, record target version,
approved API/web origins, workspace slug, nonsecret grant summary, controlled
dataset actions/times, sanitized fixtures and GET-only request audit. Compare
old open, recently completed, cancelled, reopened, archived, draft and revoked
project cases with the source UI. Incompatible self-hosted v1 requires its own
profile and equivalent tests before support can be claimed.

An optional GET-only target probe is available with
`go test -v ./internal/workstation/plane -run TestRealV2ReadProfile`.
It skips unless `BUTTER_TEST_PLANE_ACCEPTANCE=1`. Supply approved
`BUTTER_TEST_PLANE_API_BASE`, `BUTTER_TEST_PLANE_WEB_BASE`,
`BUTTER_TEST_PLANE_WORKSPACE`, `BUTTER_TEST_PLANE_PROJECT_ID`,
`BUTTER_TEST_PLANE_VERSION`, and securely provision
`BUTTER_TEST_PLANE_API_KEY` in the process environment. The command never prints
the key or upstream bodies. A successful probe is only profile-read evidence;
the controlled lifecycle and permission matrix still requires separate records.
