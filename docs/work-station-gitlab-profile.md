# Work Station GitLab REST v4 profile

This adapter supplies the KON-62 implementation for the shared Work Station projection. It supports GitLab.com and administrator-registered HTTPS GitLab instances, including an installation under a URL prefix. **Neither GitLab.com nor a self-managed deployment has been qualified in this runtime.** The fixtures below are synthetic contract tests, not captured platform evidence.

## Endpoint and credential contract

The approved API root must be exactly `<approved web root>/api/v4`; redirects and API-version fallbacks are refused. The shared outbound client enforces the approved origin and checks public addresses when dialing. The adapter exposes GET requests only, sends the credential in `PRIVATE-TOKEN`, and does not put it in URLs, errors, or returned items.

Personal, group, and project access tokens with `read_api` are the intended credentials. Their effective reach still depends on the principal's project permissions. Deploy tokens and CI job tokens are outside this profile; they are not treated as interchangeable access tokens. A successful `/user` probe and `/version` response establish a principal and an observed instance version, while selected-project reads check the actual data permission. Missing scope (`403`) is `ACCESS_DENIED`, invalid or revoked credentials (`401`) are `AUTH_INVALID`, and hidden or inaccessible projects (`404`) remain `ACCESS_DENIED`. These choices follow GitLab's [REST authentication contract](https://docs.gitlab.com/api/rest/authentication/) and [access-token scope definitions](https://docs.gitlab.com/security/tokens/access_token_scopes/).

The read profile is fixed to REST v4. It does not claim that an arbitrary reported server version is qualified: every requested response must match the fields used here, and incompatible response shapes fail `SCHEMA_ERROR`. An unavailable filter, offset limit, or profile operation fails `UNSUPPORTED`. `InspectProfile` reads version/revision from `/version`, a documented [instance metadata endpoint](https://docs.gitlab.com/api/metadata/). The deployment acceptance record must preserve that observation alongside the actual fixtures and permission grants.

## Scope and item semantics

Discovery reads `/projects` with stable ID ordering; selected scopes use `/projects/{numeric project ID}`. Each scan reads both `/projects/{id}/issues` and `/projects/{id}/merge_requests` with explicit `scope=all`, `state=all`, and `order_by=updated_at&sort=desc`. It performs bounded complete traversals, so open work with old timestamps, reopened work, and archived-project contents are not dropped through an incremental timestamp cutoff. The [Issues API](https://docs.gitlab.com/api/issues/) distinguishes project access from global user-centered lists. The [Merge Requests API](https://docs.gitlab.com/api/merge_requests/) supplies distinct opened, closed, and merged states and the draft field.

Project ID, principal ID, assignee ID, and item IID stay exact decimal strings, including numbers beyond JavaScript's safe integer range. Project ID and kind qualify every IID. Issues map opened to `NOT_STARTED`, closed to `DONE`; merge requests map opened/locked to `OPEN`, closed to `CLOSED`, and merged to `MERGED`. Unknown states stay `OTHER`. MR draft is required by this profile rather than silently inferred from titles or older alternative fields. Terminal time comes exclusively from issue/MR `closed_at` or MR `merged_at`. A missing terminal time increments the unresolved count and makes coverage `PARTIAL`; `updated_at` is never substituted. Completed work is filtered through the common rolling 90-day retention rule after deduplication; open work has no age cutoff.

Descriptions are source text for the local detail view. Source links are constructed from the approved web root and a verified project path, ignoring upstream `web_url`. Disabled Issues or Merge Requests fail `FEATURE_UNAVAILABLE` for the combined scope instead of presenting a successful empty collection. The project is checked again after both traversals; revocation or movement prevents publication.

## Pagination and resource limits

The [GitLab REST pagination contract](https://docs.gitlab.com/api/rest/#pagination) distinguishes offset and endpoint-specific keyset support. This profile uses offset pagination with 100 objects/page and a conservative maximum of 500 pages (50,000 records per endpoint). It does not assume that Issues and MR keyset parameters are supported on every deployment. Exceeding the cap or a server-reported pagination limit returns `UNSUPPORTED` and publishes no new snapshot.

`X-Next-Page` and next links must agree when both exist. Every next link must retain the approved origin, exact endpoint, and full query, changing only the page to its immediate successor. Repeated page numbers, repeated item pages, no-progress pages, contradictory metadata, malformed lists, and failed later pages reject the entire scan. If a full page lacks next-page metadata, the adapter requests the next page rather than assuming completion. A response body is limited to 16 MiB. HTTP 429 preserves `Retry-After` as a duration for shared runtime scheduling; response bodies never become error text.

## Reproducible local evidence

Run `go test ./internal/workstation/gitlab` from the repository root. The fixtures exercise:

- GET-only access and header-based credentials, project-scoped all-state requests, two-page traversal, an old open issue, exact large IDs, subgroup/source links, description, merged time, and draft state.
- Failure after an earlier successful page, repeated page/cursor, foreign origin, changed query, null schema, wrong project ID, post-scan revocation/movement, disabled features, missing MR draft, and redirects.
- Terminal retention, missing close/merge timestamps with unresolved coverage, identity/discovery, and instance version/revision capture.
- 401/403/404/429 classification, Retry-After, malformed JSON, transport-error redaction, incompatible profile arguments, pagination limits, absent headers, and unsafe scope IDs.

## Required real acceptance record

KON-56, KON-62, and KON-58 remain unaccepted until both a GitLab.com project and the selected self-managed deployment supply the following evidence. Do not infer a pass from the fixture suite.

| Field | GitLab.com | Target self-managed instance |
| --- | --- | --- |
| Approved web/API roots | Not supplied | Not supplied |
| Observed `/version` version/revision | Not observed | Not observed |
| Token type, granted scopes, principal and selected project roles | Not supplied | Not supplied |
| Sanitized private-project Issue/MR, draft, merge/close/reopen and archived-project fixtures | Not captured | Not captured |
| Offset headers, deep-page limit and incompatible-profile response | Not observed | Not observed |
| 401 revocation, insufficient-scope 403, hidden-project 404, 429 Retry-After and post-scan permission loss | Not exercised | Not exercised |
| Recorded outbound request audit showing only approved GET endpoints | Not captured | Not captured |

Required operator inputs are access to controlled projects on both deployments, permission to create the acceptance dataset outside this read-only adapter, suitable access tokens through the application's credential flow, and the self-managed instance's approved origin. No live token, source body, or private-instance identity belongs in this public evidence template.
