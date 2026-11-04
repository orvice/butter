# Work Station: Linear read profile

Profile: `linear-graphql-2026-11-03`. Implementation: `internal/workstation/linear`.
This profile has fixture verification; authenticated Linear acceptance remains
**unverified** because no dedicated Linear API key or controlled test organization
was supplied. These results do not complete KON-56, KON-60, or the real-platform
release gate in KON-58.

## Official interface baseline

The fixed endpoint is `https://api.linear.app/graphql`. Personal API keys use the
raw `Authorization` value. This connector accepts independent personal read keys;
it does not reuse a Linear Installation grant or initiate OAuth. Linear permits
restricting personal keys to Read permission and selected Teams. The key owner's
own access still determines visibility. See [GraphQL authentication](https://linear.app/developers/graphql)
and [API key permissions](https://linear.app/docs/api-and-webhooks).

The query field baseline was checked against Linear's
[first-party SDK schema at commit 253713fef2916ab744defdd71c62acccc8e929a2](https://github.com/linear/linear/blob/253713fef2916ab744defdd71c62acccc8e929a2/packages/sdk/src/schema.graphql).
Downloaded schema SHA-256:
`73fac69ae08c0cf8329bbc85a3728a9e24e6e66116ac2f5d6aebbef922b30daf`.
The relevant fields are `viewer.id`, `organization.id`, `Team.id/key/name`,
`Team.issues`, `Issue.team/state/project/creator/assignee/labels`,
`Issue.completedAt/canceledAt/archivedAt`, and `WorkflowState.type`.
Team issues support `includeArchived`, `includeSubTeams`, a filter, and cursor
pagination. State types include triage, backlog, unstarted, started, completed,
canceled, and duplicate. The remote service is continuously deployed; this schema
baseline is not evidence that a particular credential exercised every field.

Queries explicitly request archived Teams, issues, and labels: archived resources
are otherwise excluded. Cursor traversal uses `nodes`, `pageInfo.hasNextPage`,
and `pageInfo.endCursor`, with issues ordered by `updatedAt`. See
[archived resources](https://linear.app/developers/graphql) and
[pagination](https://linear.app/developers/pagination).

Linear supports relationship, date, logical, and null comparators; the adapter
applies the retention predicate inside the selected Team query. See
[filtering](https://linear.app/developers/filtering).
HTTP 200 may contain GraphQL errors alongside data. Every such response fails
the traversal before publishing. See [error handling](https://linear.app/developers/graphql).

Rate limits include requests and complexity, shared by API keys of the same
authenticated user. The adapter recognizes HTTP 429 and GraphQL `RATELIMITED`,
respects `Retry-After`, and reads reset headers in epoch milliseconds. It does not
retry inside the scan. The scheduler owns durable backoff and cross-connection
budgets; adding keys must not bypass a user's limit. See
[rate limiting](https://linear.app/developers/rate-limiting).

## Implemented read behavior

Only five static, named GraphQL queries are available: identity, Team discovery,
one Team, Team issues, and additional issue-label pages. The request seam refuses
mutations. Requests use the approved HTTPS API origin through the shared outbound
client, with redirect and proxy restrictions. Errors expose fixed categories;
upstream text, request bodies, headers, and keys are never returned.

Identity combines the official API origin, organization ID, and viewer principal.
A scope is a stable Team ID, displayed with its key and name. A Linear Project is
item metadata; choosing a Project never imports issues from another Team.

Each scan traverses only its selected Team, with sub-teams explicitly disabled.
The server predicate includes every nonterminal or unknown state, terminal items
within 90 days, and terminal items whose expected terminal timestamp is null.
The local retention check independently enforces the inclusive 90-day boundary.
Archive time and update time never substitute for an actual terminal timestamp.
Completed uses `completedAt`; canceled/duplicate uses `canceledAt`. A terminal
item without that timestamp is omitted and increments unresolved coverage,
producing `PARTIAL`. Reopened items clear terminal status and end time even if an
upstream historical completed timestamp remains populated. Unknown state types
remain visible as `OTHER`, preserving the raw state name.

Items returned with a different actual Team ID are excluded from the old scope.
Successful traversal rechecks Team access before publication. Every page and
label-page continuation must succeed, with present pagination metadata and a
nonempty, advancing cursor. At most 1,000 issue pages are read; hitting the bound
returns `UNSUPPORTED` rather than publishing a truncated snapshot. API bodies are
limited to 16 MiB. Label pages are hydrated when the first 20 labels do not cover
the item; hydration aborts if its Team or update timestamp changes.

Each cycle performs a bounded full reconciliation of the selected retention
window. It does not retain incremental cursors or claim that cursor pagination
provides a remote transaction: a concurrent Team move can settle on the next
successful reconciliation. There is no organization-wide issue polling.

## Verification record

`go test ./internal/workstation/linear` passes with in-memory HTTP responses based
on the documented schema. These fixtures are synthetic, not remote captures.

| Case | Local fixture result | Remote acceptance |
| --- | --- | --- |
| Viewer + organization identity, raw API-key authentication | Pass | Unverified |
| Team discovery, archived flag, cursor pages, stable scope resolution | Pass | Unverified |
| Team-only issue pages, Projects, assignees, labels | Pass | Unverified |
| Archived open issues, state mapping, inclusive 90-day boundary | Pass | Unverified |
| Reopen clears terminal fields; moved-Team response is excluded | Pass | Unverified |
| Missing terminal time produces `PARTIAL` and unresolved count | Pass | Unverified |
| Multi-page labels | Pass | Unverified |
| HTTP 200 partial GraphQL failure, repeated cursors, empty-next pages | Pass | Unverified |
| HTTP 401/403/429, GraphQL authentication/permission/rate errors | Pass | Unverified |
| Access lost at final Team recheck, error redaction, unsafe source URL | Pass | Unverified |

The controlled remote run must record the organization and Team IDs, schema/date
baseline, key's Read and Team restrictions, sanitized response fixtures, request
operation audit, and observed rate-limit metadata. It needs two Teams and more
than one page of issues, with/without Projects, archived records, completed and
canceled records, reopen, a Team move, restricted-Team discovery, key revocation,
and permission restoration. No remote fixtures, tenant IDs, actual grant scopes,
or write-audit evidence were available during this implementation. Those entries
must be completed before claiming the real Linear gate passed.
