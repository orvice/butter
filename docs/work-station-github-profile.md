# Work Station: GitHub.com read profile

Profile: `github-rest-2022-11-28`; checked against first-party documentation on
2026-11-03. This connector uses the fixed `https://api.github.com` origin,
`Accept: application/vnd.github+json`, and an explicit
`X-GitHub-Api-Version: 2022-11-28`. It supports GitHub.com, with stable repository
ID plus issue/PR number identity. It does not claim GitHub Enterprise support.

Authenticated remote acceptance remains **unverified**. The workspace's GitHub
access for delivering this PR is not a dedicated, controlled Work Station test
credential and was not reused to probe private repositories. Local fixtures are
synthetic HTTP responses, not captured private-platform evidence. KON-56,
KON-61, and the GitHub real-platform gate in KON-58 remain open until that
evidence exists.

## Official interface baseline

GitHub repository issues support `state=all`, updated sorting, a `since`
timestamp, and pages of up to 100. Issue responses can also contain pull requests,
identified by `pull_request`; those must be excluded from the Issue projection.
Fine-grained private-repository access requires Issues read permission. See
[repository issues](https://docs.github.com/en/rest/issues/issues?apiVersion=2022-11-28#list-repository-issues).

Pull requests use a separate endpoint, support updated sorting and state filters,
and expose draft and merge metadata. The adapter requests explicit descending
updated order rather than relying on a default direction. Fine-grained private
access requires Pull requests read permission. See
[pull requests](https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#list-pull-requests).
Repository discovery and identity probing use
[repositories for the authenticated user](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repositories-for-the-authenticated-user)
and [the authenticated user](https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-the-authenticated-user).

GitHub supports conditional GET using ETag and `If-None-Match`. A correctly
authorized 304 response can save primary request quota. Pagination and query
parameters change the response representation, and updated sorting can shift
items between pages. A 404 can also hide an authorization failure. See
[REST API best practices](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api).

## Implemented reconciliation

Initial and periodic full scans traverse all Issue and PR pages and then apply
the local rolling 90-day terminal retention window. All old open items remain
eligible. Merged PRs use `merged_at`; closed issues and unmerged PRs use
`closed_at`. Missing terminal time produces `PARTIAL` coverage and an unresolved
count; edit time never extends the terminal window. Reopened items clear terminal
status and end time. Unknown states remain visible as `OTHER`.

Incremental reconciliation uses the previous successful scan's **start time**,
with five minutes of overlap. Issues use native `since` filtering. PRs traverse
descending updated pages until the overlap boundary, verifying order before
stopping. Each cycle also traverses complete inventories of all open Issues and
PRs. A previously open item missing from those inventories and the update set is
re-read individually. A missing or transferred item drops its old observation;
401/403 abort the scan. Missing item 404/410/redirect handling is confined to an
individual suspected item; repository 404 still invalidates the whole scope.
Redirects are never followed, including item transfer locations.

More than 100 suspected missing open items triggers a full scan. The runtime
also uses full scans after a day, without a valid checkpoint, or after incomplete
coverage. Daily full reconciliation removes deleted historical terminal items
that an update feed cannot report. Every candidate is staged only after all
required requests succeed and repository identity/access is rechecked. Publication
and checkpoint fencing belong to the shared runtime.

The upstream API does not provide snapshot isolation. Concurrent edits or page
reordering can require a subsequent reconciliation. The adapter does not represent
an overlap cutoff or ETag as proof of remote transactional consistency.

## Conditional-response cache

Cache keys include a SHA-256 credential fingerprint, HTTP method, complete URL
with all query parameters, media type, and API version. Raw credentials and
Authorization headers are never retained in keys or entries. The in-process cache
holds at most 128 entries and 32 MiB of bodies; each body is at most 2 MiB, with
30-minute inactivity expiry. It retains only ETag, Link, and Content-Type metadata.
A changed or rejected response evicts the previous representation for that request.
An unexpected 304 without its matching cached body is a schema failure.

A 304 replays only that page's body and pagination metadata. The next linked page
is still fetched or independently revalidated, and all other required endpoints
are still traversed. The cache is an optimization: process restarts, cache
eviction, or multiple Pods can result in independent ordinary GETs, with the same
projection behavior.

## Verification record

`go test -race ./internal/workstation/github ./internal/workstation/linear` and
`go vet ./internal/workstation/github ./internal/workstation/linear` pass.
The GitHub fixtures cover:

| Case | Local fixture result | Remote acceptance |
| --- | --- | --- |
| Read-only, fixed origin, large numeric stable IDs, separate Issue/PR kinds | Pass | Unverified |
| Complete multi-page traversal, foreign/repeated Link rejection, late failure | Pass | Unverified |
| Merged/draft state, reopen, unknown state, actual terminal timestamp retention | Pass | Unverified |
| Conditional first-page 304 with changed second page | Pass | Unverified |
| Credential, URL/page/query, media type, and API-version cache isolation | Pass | Unverified |
| Cache body/entry bounds, authorization eviction, unexpected 304 | Pass | Unverified |
| Issue `since` overlap, PR overlap traversal and full all-open inventories | Pass | Unverified |
| Missing-open recheck, merge, deletion/transfer, preserved unchanged terminal item | Pass | Unverified |
| Late item permission/server error prevents incremental publication | Pass | Unverified |
| Incomplete coverage forces full reconciliation | Pass | Unverified |

Required controlled-run records: repository IDs and visibility, API version
accepted by the service, exact fine-grained grants (Metadata, Issues read, Pull
requests read), classic-token profile if supported, key owner, pagination dataset,
sanitized HTTP fixtures, request-method audit, rate-limit headers, and outcomes
for create/close/reopen/merge/draft/transfer, key revocation, repository permission
removal and restoration, and a genuine 304 traversal. The existing
`has_issues=false` profile fails the scope with `FEATURE_UNAVAILABLE`; this does
not offer a PR-only disabled-Issues profile. That case needs a separate capability
decision and real repository evidence before claiming support.
