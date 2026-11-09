# Plane webhook identity and stuck Run recovery

## Contract and fix (#524)

Primary sources:

- [Building an agent: AgentRun webhooks](https://developers.plane.so/dev-tools/agents/building-an-agent#agentrun-webhooks)
- [Handling webhooks: signature verification](https://developers.plane.so/dev-tools/build-plane-app/webhooks)
- [Signals and content payload](https://developers.plane.so/dev-tools/agents/signals-content-payload)

The documented Agent payload is the event object itself. Butter also accepts
that object inside `data` for envelope compatibility. The identity fields are:

| Field | Meaning |
| --- | --- |
| `app_client_id` | OAuth App client ID; must match the App selected by the webhook URL. |
| `workspace_id` | Plane workspace UUID, **not** a Butter workspace ID or Plane workspace slug. |
| `agent_user_id` | The App's bot user; must match the stored installation's `app_bot`. |
| `agent_run.id` | Plane Run ID. |
| `agent_run.workspace` | Run's Plane workspace UUID; if supplied, must match `workspace_id`. |
| `agent_run.agent_user` | Run's bot user; if supplied, must match `agent_user_id`. |
| `agent_run_activity.id` | Prompt/Stop activity ID, used for logical deduplication. |
| `agent_run_activity.actor` | The **current** activity's author; never substitute the Run creator. |
| `agent_run_activity.workspace` / `.agent_run` | If supplied, must match the workspace and Run. |
| `agent_run_activity.content.body` | Prompt text. |

The official examples **do not contain `app_installation_id`**. After raw-body
HMAC verification and client-ID validation, Butter resolves exactly one stored
installation within the selected Butter workspace/App using the Plane
workspace UUID and bot user. It must be installed and have all four scopes:
`agents.runs:read`, `agents.runs:write`, `agents.activities:read`, and
`agents.activities:write`. An explicit `app_installation_id`, when present,
remains mandatory to match; it never falls back to a different installation.
Missing required identities, ambiguous bindings and conflicting aliases are
rejected. The optional resource-level fields do not replace required top-level
(or `data`-level) identities.

Previous `run`, `activity`, `agent_activity`, `workspace_id` and activity
`user_id` aliases remain accepted, but contradictory aliases are rejected.
`agent_run/created` is authenticated and ignored: only a Prompt starts an
Invocation. A Stop uses the activity actor and is admitted even with inbound
disabled, subject to the normal user allowlist.

The installation callback already verifies the installation using its bot
token and upserts by Butter workspace/App and Plane workspace. A reinstall
replaces the external installation ID, bot identity, grant and token while
retaining the local installation record ID. No schema migration is needed.

## Safe diagnostics

The HTTP response remains generic (`401 unauthorized` for identity rejection).
The `rejected plane webhook` log now includes `reason` and `field` in its error:

- `missing`: required client, workspace, bot, actor or stored installation ID.
- `app_client_mismatch`: webhook URL and payload select different Apps.
- `installation_not_found`: no installation, or an explicit installation ID is stale.
- `workspace_mismatch` / `bot_user_mismatch`: stored installation binding differs.
- `installation_ambiguous` / `installation_inactive`: cannot safely select a live installation.
- `scope_incomplete`: `field` names the missing required scope.
- `conflict`: resource/envelope/legacy aliases disagree.
- `actor_is_bot` / `actor_not_allowed`: actor policy rejected the Prompt/Stop.

Identity values appear only as presence flags and 12-hex SHA-256 fingerprints
(first six bytes). `app_hash`, `installation_hash`, `run_hash`, `activity_hash`
and `invocation_hash` correlate the new Plane lifecycle logs. These diagnostics
never include credentials, signatures, raw bodies, prompts or raw IDs. Existing
shared Runner logs retain their separate logging contract.

To compare an ID locally without publishing it, compute SHA-256 of its exact
UTF-8 bytes (no trailing newline) and take the first 12 hex characters. Do not
paste credentials or complete webhook payloads into logs, Issues or PRs.

## After deployment

1. Verify the App's webhook endpoint and secret. If the diagnosis indicates an
   actual stale binding or missing scopes, reinstall through Butter's Plane App
   page with the four scopes above. Do not edit database identity fields or
   disable checks to make a payload pass.
2. Start a new Run or send a new Prompt. Look for `plane webhook received`
   (`decision=accepted`), `plane webhook queued`, `plane event processing`,
   `plane invocation starting`, `plane invocation settled`, and
   `plane reply delivered`, correlated by hashes.
3. Confirm a Plane processing record and Invocation exist. On success the
   processing record is `SUCCEEDED`, its reply delivery is `delivered`, and
   the response activity is visible in Plane. A create-only notification is
   intentionally `ignored`; a duplicate is not a second execution.
4. If execution or delivery fails after admission, inspect that processing
   record and its delivery status instead of replaying the model. Use the
   existing delivery recovery actions where applicable.

## Runs rejected before admission

A rejected webhook has no durable queue entry, processing record or Invocation
from that delivery. Updating Butter does **not** replay it, and an HTTP 401 is
not an acknowledgement. App-webhook retry guarantees are undocumented; do not
wait for a retry as the recovery plan.

- Once the fix/binding is verified, send a **new Prompt** in the thread or start
  a new Run. This creates a new activity and passes through normal admission.
  Tell the user it starts a new execution; any earlier rejected prompt is not
  automatically included.
- Before retiring an old Run, check for accepted deliveries, processing records
  and live Invocations. If work was admitted, use normal Stop/delivery recovery;
  do not label an active Run failed or replay an uncertain execution.
- For a Run confirmed never admitted and still showing working, an authorized
  operator can use that installation's bot token to create an error activity:
  `POST /api/v1/workspaces/{workspace_slug}/runs/{run_id}/activities/` with
  `Authorization: Bearer <bot token>` and this body:

  ```json
  {
    "type": "error",
    "content": {
      "type": "error",
      "body": "This request was not accepted because of a webhook configuration error. Please send a new prompt."
    }
  }
  ```

  Plane documents an error activity as setting the Run to `failed`. Resolve
  the workspace **slug** and the Run ID from the verified installation/Run;
  never use the workspace UUID in that route. Keep the token in the operator's
  credential tooling, not shell history, screenshots or diagnostic logs. Verify
  the resulting Run state before proceeding. Butter deliberately does not
  respond through an installation whose webhook identity failed verification.

## Regression coverage

`TestReceiverOfficialPlanePayload` reproduces the pre-fix identity rejection
with a synthetic-ID fixture matching the official example, directly and inside
`data`. Admission tests also cover create/Stop, legacy aliases, missing and
contradictory identities, actor policy, reinstall, scopes and ambiguity.
`TestOfficialPlanePromptReachesInvocationAndActivity` drives signed admission,
worker dispatch, processing projection, the real Runner/Invocation recorder and
the activity HTTP adapter; the queue, stores, agent leaf and Plane HTTP transport
are deterministic test doubles. HTTP tests verify 200/401 behavior and safe
logs. This is not a claim of production end-to-end acceptance: the new Run and
old-Run retirement steps above still need to be verified after deployment.
