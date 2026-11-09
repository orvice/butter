# Plane webhook envelope: follow-up to #524, #526 and #527

Reviewed 2026-11-09 UTC. This note records the protocol evidence used by the
follow-up parser fix, including what the earlier fixes missed. It is not a
claim of successful Plane Cloud end-to-end acceptance.

## Evidence and its limits

| Source | What it establishes |
| --- | --- |
| [Official PRD Agent types, revision `1cda0ff35c3afd892b8f3b5d49d89de45980633e`](https://github.com/makeplane/prd-agent/blob/1cda0ff35c3afd892b8f3b5d49d89de45980633e/src/lib/types.ts) | `WebhookEvent` has outer `event`, `action`, `webhook_id`, `workspace_id`, `data`, and optional `activity: Record<string, unknown>`. Its `data` is the separate `AgentRunUserPromptWebhook` type. Business Activity actor and Run references are UUID strings. |
| [Official PRD Agent entry point, same revision](https://github.com/makeplane/prd-agent/blob/1cda0ff35c3afd892b8f3b5d49d89de45980633e/src/index.ts) | The webhook handler selects `webhook.data`, dispatches on that object's `type`, and passes that object to the Agent handler. It does not combine the audit activity with the business activity. |
| [App webhook documentation](https://developers.plane.so/dev-tools/build-plane-app/webhooks) | The generic `event/data/activity` envelope shows an expanded object at outer `activity.actor`. This is separate from the Agent-specific payload. |
| [Building an agent](https://developers.plane.so/dev-tools/agents/building-an-agent#agentrun-webhooks) | Shows the flat inner Agent payload, including `agent_run_activity.actor`, `.agent_run`, `.content.body`, and `agent_run`. Create does not contain a Prompt. |
| [OAuth scopes](https://developers.plane.so/dev-tools/build-plane-app/oauth-scopes) | The Activity scope names are `agents.run_activities:read` and `agents.run_activities:write`, not `agents.activities:*`. |
| Production Loki, 2026-11-09 04:51:17–04:51:30 UTC, after #527 | Six authenticated deliveries failed before admission: four reported `activity.actor` string/object mismatch, two `activity.agent_run` string/object mismatch. No queued/processing/Invocation-start log was observed in that query window. |

The production evidence consists of safe diagnostic metadata, **not a captured
body**. It proves the two observed JSON kinds and their paths, not the complete
live payload. The generic envelope documented by the first-party sources
explains why expanded objects can exist at these outer paths without changing
the UUID-string contract of the inner business fields.

The earlier local `plane-gate-a-doc-review.md` (2026-11-07, §5) had already
identified the distinction between the flat Agent guide and the official
example's `data` envelope. That distinction was not carried into #526's
normalizer or its fixture. The source code was instead treating outer audit
`activity` as a legacy business alias. #527 supplied the missing safe decoder
evidence; it did not resolve this parser-role error.

## Three defects fixed together

1. **Metadata interpreted as a business activity.** Decoding outer
   `activity.actor` / `.agent_run` as business UUID strings rejects the whole
   request before reaching its valid `data` payload. Extracting `id` from
   arbitrary audit objects would be the wrong fix: those values must not
   authorize a person, select a Run, supply a Prompt/Stop or choose a dedupe key.
2. **Transport and business actions merged.** Even with audit metadata removed,
   a synthetic envelope whose transport action is `created` and inner business
   action is `prompted` fails the old alias-conflict check. These actions belong
   to different objects. Their independent meanings follow from the separate
   first-party types; those exact values are a test probe, not a production
   capture.
3. **Incorrect scope names hidden by circular test setup.** The receiver required
   undocumented `agents.activities:*`. Its old installation fixture populated
   grants from that same implementation constant, so tests could not catch
   the error. An independent fixture using the documented names reproduced a
   subsequent `scope_incomplete` failure after parsing was fixed. The checks and
   fixtures now use the documented spelling; missing permissions still fail
   closed, and the incorrect spelling is not accepted as an alias.

## Bounded normalization and identity rules

- A nonempty outer `event` plus an object `data` selects the documented envelope.
  The inner object alone supplies the required business `type`, `action`,
  App client, workspace, bot, Run, Activity, actor, Prompt and Stop signal.
- Outer `event`, `action`, `webhook_id` and `activity` are transport metadata.
  `activity` stays opaque; its shape and any IDs, author, body or `stop` signal
  never enter business normalization. The inner `type/action` pair controls
  dispatch, as in the official example; an event label is not authorization.
- Shared outer identity claims constrain the inner identity. Workspace, client,
  bot and explicitly named Run/Agent-activity references cannot contradict it.
  They cannot manufacture missing inner identities. An explicit installation
  ID at either level must match the verified stored installation; two different
  installation claims are rejected.
- The flat Agent guide format remains supported. Legacy business aliases
  (`run`, `agent_activity`, `activity`, `user_id`, `workspace_id`) retain their
  exact scalar typing and conflict checks. A `data` alias without the explicit
  transport marker keeps the previous strict merge behavior.
- No recursive search for arbitrary `id`, actor or Run, no object-to-string
  coercion, and no fallback to a Run creator or audit author is introduced.
  The raw-body signature check, body cap, user policy, installation resolution,
  queue ownership and safe diagnostics stay in place.

A stored installation carrying Plane's actual `agents.run_activities:*` grants
needs no data migration. Reinstallation is needed only if a genuinely required
permission/binding is missing, not to compensate for the old code's spelling.

## Fixture provenance and regression coverage

`internal/runtime/plane/testdata/agent_run_user_prompt_envelope.json` is
**synthetic_shape_only**. It combines the first-party outer/inner types with
expanded audit actor/Run objects matching the observed production kinds. Its
synthetic audit IDs/author/body/Stop signal deliberately differ from the inner
business data, to detect accidental trust in the wrong layer. It is not
`live_redacted` and contains no production identities, text or credentials.

Coverage includes:

- Expanded audit objects, arbitrary opaque metadata, no metadata, and distinct
  transport/business actions; correct business identity, Prompt and signal.
- Create ignored, Prompt and follow-up admitted, authorized Stop still allowed
  with inbound disabled, and an audit Stop unable to stop a normal Prompt.
- Missing inner identities, missing Prompt, conflicting claims/aliases, bot or
  unauthorized actor, stale explicit installation and malformed *business*
  references still rejected.
- Documented grants accepted; each missing scope and the old undocumented
  spelling rejected, with fixtures independent of the requirement constant.
- Public HTTP 200/400/401 behavior and safe logs for flat and enveloped inputs.
- Flat and enveloped signed admission → processing projection → real Runner /
  Invocation recorder → HTTP activity adapter, with deterministic test doubles.
- `TestRedisEnvelopeMentionReachesReplyAndDeduplicates` uses real Redis admission,
  actual worker loops and Redis control/batching, the real Runner/recorder, and
  the real activity HTTP adapter. It verifies two actors' separate turns and
  Invocation IDs, final replies, and a retry whose changed audit ID cannot
  bypass business deduplication. Storage, the agent leaf and Plane HTTP remain
  test doubles. The existing Redis suites cover Stop, contention and recovery.

This follow-up was exercised with an isolated local Redis 7.4.2, not production
Redis. Mongo-backed cases still require the normal integration environment.
After deployment, a new real mention/follow-up must still demonstrate queued →
processing → Invocation → confirmed Plane response. If a later stage fails,
use its actual diagnostics and frozen processing/delivery record; never replay
an uncertain execution merely to obtain a successful-looking Run.
