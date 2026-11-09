# Plane global read/write grants and ordinary comment notifications

Reviewed 2026-11-09 UTC; follow-up to #524 and #529. This corrects the
named-scope-only assumption in the previous receiver checks. It does not
represent completed Plane Cloud end-to-end acceptance.

## Evidence

1. The owner supplied the target Plane App's **All Workspace Resources** UI:
   Read and Write are both selected. This is the global resource/action choice,
   not an instruction to request additional named permissions.
2. The owner supplied Butter's verified installation card: `installed`,
   credentials `valid`, and **Granted scopes: read, write**. The stored grant is
   not empty; named requirements being absent is not equivalent to no access.
   These screenshots are not copied into this repository.
3. The [official Python OAuth client at `4fe2ec815ab5fc595b38ddfd0ef6ee4bcadd0fb9`](https://github.com/makeplane/plane-python-sdk/blob/4fe2ec815ab5fc595b38ddfd0ef6ee4bcadd0fb9/plane/client/oauth_client.py)
   uses `read write` as the default for `OAuthClientCredentialsParams` and
   `get_client_credentials_token`, including installation-specific Bot tokens.
   The response model preserves the returned `scope` string.
4. The [OAuth scope catalog](https://developers.plane.so/dev-tools/build-plane-app/oauth-scopes)
   documents the granular alternative: `agents.runs:read` and
   `agents.run_activities:read/write`. It does not separately explain the
   global UI choice. The [App registration page](https://developers.plane.so/dev-tools/build-plane-app/create-oauth-application)
   points to that same Scopes & Permissions selector.
5. On deployed commit `ce118b6`, Loki at 09:09:25 UTC classified Agent Create and
   Prompt notifications correctly but reported all three named scopes missing.
   Two other failures were explicitly `event_kind=issue_comment`, with no
   Agent business discriminator. There was still no accepted Invocation.

The SDK establishes that these global literals are legitimate Bot-token
values; it is an HTTP client, **not** the provider's authorization resolver.
The global coverage policy below follows the target Cloud UI's resource/action
choice and its actual persisted grant, rather than claiming to have inspected
closed-source Cloud authorization code. No live Activity API permission probe
or successful Cloud response is claimed. Plane remains the authority on actual
API access and can reject a request even after local admission.

## Coverage policy

Requirements stay minimal: read Runs, read Activities, write Activities.
For each requirement, accept its exact granular scope **or** the exact global
verb (`read` for a read, `write` for a write).

- Global `read` does not satisfy any write requirement.
- Global `write` does not supply read permission.
- Mixed global/granular grants can satisfy only their declared operations.
- Other resources' granular grants, case variants, prefixes, wildcard strings,
  unknown or empty grants do not imply the required permission.
- Only the verified installation's stored provider grants participate. A
  webhook's `scope`/`granted_scopes` claims never authorize execution.
- Preserve the provider's original grant list. Do not rewrite `read/write` into
  invented granular grant metadata, request broad access for everyone, or
  change OAuth requests/renewal merely to repair a literal-comparison bug.

New least-privilege installations may use the three named scopes. The owner's
already-approved global installation needs neither broader access nor a data
migration/reinstall for this checker correction. Workspace/App/installation/bot,
Run references, current actor policy and signature checks are unchanged.

The red regression with stored `read, write` reproduced the exact production
`missing_scopes` rejection. Positive tests now cover global, granular and mixed
representations; read-only, write-only and lookalike grants remain rejected.

## Ordinary issue_comment filtering

The [generic App envelope](https://developers.plane.so/dev-tools/build-plane-app/webhooks)
uses `event`, `data` and audit `activity`. Production classification now confirms
that ordinary `issue_comment` deliveries accompany the Agent events. A comment
resource is not an Agent payload and need not carry `data.workspace_id`,
`app_client_id` or an Agent Run.

A bounded filter acknowledges only a signed, valid explicit `issue_comment`
envelope with object data and **no Agent claim** at either business level.
Outer audit activity remains opaque. Canonical/legacy Agent resource claims,
unknown type discriminators, malformed data and missing Agent identities do
not become harmless just because the outer event says `issue_comment`.

The filter uses the existing JSON v1 field matching so case-folded Agent fields
cannot evade its guard. It runs after HMAC verification and never adds a queue
entry, Stop marker, processing record or Invocation. Known ignored comments get
an `event_kind=issue_comment` log and the existing HTTP 200 ignored response;
unsigned requests still get 401. Unknown event families are not guessed away.

## Regression boundary

- Signed receive with global/granular/mixed grants, missing operation rejection,
  unchanged stored grant metadata and rejection of payload-supplied grants.
- Global grants do not bypass any tenant/installation/Run/current-actor check.
- The ordinary comment shape reproduces the old missing-workspace error and is
  now ignored. A comment label cannot hide valid or incomplete Agent payloads,
  including case-folded field spellings and legacy aliases.
- Public HTTP accepts both grant representations and ignores only the known
  non-Agent shape; diagnostics do not leak input or grant values.
- The real Redis/worker/batching/fencing/Runner/Invocation/reply-adapter suite
  runs separately for granular minimum and global read/write installations,
  including duplicate delivery and a second actor's follow-up. Storage, the
  agent leaf and Plane HTTP are still deterministic test doubles.

A post-merge image publication, deployed commit check and new real mention are
still required. A successful local suite or a verified grant representation
alone is not proof that the target Plane API accepted a response.
