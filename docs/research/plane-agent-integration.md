# Plane Agent Runs integration

Research date: 2026-11-03 — continuing the investigation recorded in
[butter#487]. This note answers the protocol questions that issue listed as
"待确认" and re-checks the ones it treated as settled. It contains **no**
implementation; the goal is to fix the contract before an Agent is pointed at
it.

Everything below was checked against four sources, each named at the claim it
supports:

1. the current developer documentation at `developers.plane.so` (fetched
   2026-11-03),
2. the official Node SDK `@makeplane/plane-node-sdk@0.3.1` (its TypeScript
   types are the most precise published statement of the wire schema),
3. the AGPL source of Plane at `makeplane/plane@618d0eb7` (2026-11-03), which
   is what a self-hosted Community instance runs,
4. live unauthenticated probes of `api.plane.so`, which tell us which
   endpoints exist at all.

## Conclusion

The beta protocol is **implementable against the shape ADR-0015 already
established** — durable queue, one Session per Run, serialized turns, a
processing record as the retry boundary — but four things from the issue are
now wrong or unanswerable as written:

1. **Plane Agent Runs are not in the AGPL codebase.** A self-hosted Community
   instance cannot offer them at any version. Only Plane Cloud (and, by
   inference, the closed Commercial/Airgapped builds derived from it) has the
   endpoints; `api.plane.so` serves them today. Any acceptance criterion that
   says "install in the target Plane instance" therefore first has to name an
   instance that is not Community Edition.
2. **`external_link` cannot be backfilled.** The run is created by Plane, and
   the SDK exposes create/retrieve only — there is no run update. The issue's
   "自动创建 Run 如何补充 external_link" has no answer: it is not possible.
   A Butter deep link has to ride in the agent's own output instead.
3. **The first delivery is not a prompt.** `agent_run_create` carries no
   message body at all, and neither payload carries the work item or the
   comment text, so the "promptContext carries context" assumption from
   ADR-0015 has no Plane equivalent. Butter must read the work item and
   comment thread through the REST API on first touch, exactly as the issue
   suspected.
4. **`error` activity persistence is contradictory in the docs themselves**,
   and the SDK shows why the question cannot be settled client-side: the
   create request has no `ephemeral` field, so ephemerality is a server rule
   keyed on `type`. Assume the body may never reach the user and never rely on
   it.

Against that, the parts that looked hardest are ordinary: the webhook is an
HMAC over the raw body with the same headers as any Plane webhook, the four
agent scopes are enough for the whole chain, and the bot token is a
`client_credentials` mint that can be redone at will rather than a rotating
refresh token.

## 1. Availability: which Plane can run this

The AGPL source has no trace of the feature. Greps over the whole tree at
`618d0eb7` return nothing for `agent_run`, `AgentRun`, `app_installation`,
`AppInstallation`, `client_credentials`, `authorize-app`, or `mentionable`,
and the only webhook model is the workspace one
(`apps/api/plane/db/models/webhook.py`). `apps/api/plane/urls.py` mounts
`plane.app`, `plane.space`, `plane.license`, `plane.api` and `plane.web` — no
app-installation or run router.

The Cloud host serves the same endpoints the docs describe, so they exist
somewhere closed:

```
GET  https://api.plane.so/api/v1/workspaces/example/runs/
     → 401 {"detail":"Authentication credentials were not provided."}
GET  https://api.plane.so/auth/o/authorize-app/        → 302
GET  https://api.plane.so/auth/o/app-installation/?id=x → 401
POST https://api.plane.so/auth/o/token/  (bad client)   → 401 {"error":"invalid_client"}
```

A 401 rather than 404 means the routes are mounted, i.e. the feature is served
by the codebase behind `api.plane.so`, not by the AGPL one.

Plane's own edition docs state that the Community Edition has "no code
dependencies or restrictions on and from the Commercial Edition" and is "at
par with the Free tier of the Cloud edition", while Commercial and Airgapped
are closed-source builds of the Cloud codebase ([editions and versions]). So:

- **Cloud**: available now, any plan that can open Settings → Integrations →
  "Build your own". Whether a Free workspace exposes it is unverified; the docs
  never mention a plan gate.
- **Community self-hosted**: not available, and no version will make it
  available.
- **Commercial / Airgapped self-hosted**: possible in principle because they
  track the Cloud codebase, but which release first shipped it is not
  published; an operator can only check by opening the integrations page and
  looking for "Enable App Mentions".

Consequence for the issue: the task "在目标 Plane 实例完成真实安装" is only
meaningful with a named non-Community instance, and Butter's documentation
should say so up front rather than during onboarding.

## 2. Install and credentials

### The flow

Registration (Settings → Integrations → "Build your own") collects Setup URL,
Redirect URI and Webhook URL, and two switches that matter: **Enable App
Mentions** (what makes the app mentionable and turns it into a bot user) and
the **agent scopes**. Installation is a normal OAuth consent redirect:

```
GET https://api.plane.so/auth/o/authorize-app/
      ?client_id=...&response_type=code&redirect_uri=...&scope=...
    →
GET {redirect_uri}?app_installation_id=...&code=...
```

There is **no state parameter in the bot flow** — the docs add `state` only to
the user-token flow, and Plane's own complete example (`dev-tools/build-plane-app/examples`)
keys the callback on nothing but `app_installation_id`. This settles the
issue's question: there is no install `state` to associate, and Butter must not
build a flow that depends on one.

### Bot token

The authorization code in the redirect is unused. The token is minted from the
installation ID:

```
POST https://api.plane.so/auth/o/token/
Authorization: Basic base64(client_id:client_secret)
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&app_installation_id=...&scope=...
```

and answers `{"access_token":"pln_bot_…","token_type":"Bearer","expires_in":86400,"scope":"…"}`.
The same call re-mints it when it expires — there is no refresh token, no
rotation, and no `invalid_grant` re-install state to model. Butter's
`linearconn.TokenSource` (refresh under a per-installation lease with a
token-revision CAS, `NEEDS_REINSTALL` on `invalid_grant`) therefore has a
simpler Plane counterpart: mint on demand, cache until ~5 minutes before
`expires_in`, and treat a 401 as "mint again once" before failing the turn.

The installation record itself comes from:

```
GET https://api.plane.so/auth/o/app-installation/?id=APP_INSTALLATION_ID
Authorization: Bearer BOT_TOKEN
→ [{"id":..., "workspace":"uuid", "workspace_detail":{"name":...,"slug":...},
     "app_bot":"bot-user-uuid", "status":"installed"}]
```

`workspace_detail.slug` is the path segment of every REST call, and `app_bot`
is the agent's identity in that workspace — the Plane analogue of Linear's
`app_user_id`.

### The webhook secret is undocumented

Nothing in the registration docs, the webhook docs, or the example app says
where an **OAuth app's** webhook signing secret comes from. The example reads
`PLANE_WEBHOOK_SECRET` from the environment and never shows it being issued,
which is an odd gap given that workspace webhooks get a CSV download whose key
is prefixed `plane_wh_`. Two candidate answers, both cheap to try at install
time: the app's Client Secret, or a per-app secret shown in the app's edit
form next to the Webhook URL. Butter should store a configurable webhook
secret instead of inferring one, and the very first real delivery should be
verified with both candidates and the result recorded.

### Routing (design decision, not a protocol fact)

Because the bot flow has no state and the webhook URL is configured per app,
the Linear model carries over unchanged: a `PlaneApp` is a **workspace
resource** holding one Plane `client_id`, and the installation, the run and the
session are all resolved from the App. A delivery cannot name a Butter
workspace, so the App must.

## 3. Webhook contract

### Headers and signature

Same headers as any Plane webhook ([handling webhooks]):

| Header | Meaning |
| --- | --- |
| `X-Plane-Delivery` | delivery ID |
| `X-Plane-Event` | event name |
| `X-Plane-Signature` | HMAC-SHA256 hex digest of the raw body |

The signature is over the **raw request bytes**, hex, no prefix, verified in
constant time. The AGPL sender confirms the construction and shows what the
digest is computed over (`apps/api/plane/bgtasks/webhook_task.py`):

```python
hmac_signature = hmac.new(webhook.secret_key.encode("utf-8"),
                          json.dumps(payload).encode("utf-8"), hashlib.sha256)
headers["X-Plane-Signature"] = hmac_signature.hexdigest()
```

so a body that a proxy re-serializes fails verification — the Linear receiver's
"read the raw body, cap it, verify before parsing" shape is exactly right.

### Payloads

Two events, both documented with samples ([building an agent]):

`agent_run_create` — `type: "agent_run"`, `action: "created"`, the run object
(`id`, `agent_user`, `issue`, `project`, `workspace`, `status: "created"`,
`type: "comment_thread"`, `started_at`) plus `agent_user_id`, `app_client_id`,
`issue_id`, `project_id`, `workspace_id`, `comment_id`. **No message body, no
issue title, no comment text.**

`agent_run_user_prompt` — `type: "agent_run_activity"`, `action: "prompted"`,
the prompt activity (`id`, `agent_run`, `type: "prompt"`, `content.body`,
`signal`, `actor`, `workspace`), the run object (now `status: "in_progress"`),
and the same id block.

Three consequences:

- **Route on `app_client_id`.** Plane puts the app's client ID in every
  payload, so a single shared endpoint can resolve the App without a path
  segment. Keep the App ID in the URL too (as Linear does) and require the two
  to agree; a mismatch is a configuration error, not a forged request.
- **Dedupe on the activity ID, not the delivery ID.** The AGPL sender builds
  `X-Plane-Delivery` with `str(uuid.uuid4())` *inside the sending task*, so
  every retry of the same logical event carries a different value. Whether the
  app-webhook path does the same is unknown — which is reason enough not to
  depend on it. Business identity is `(installation, agent_run.id, activity.id,
  action)`.
- **`X-Plane-Event` and `type` disagree.** The docs label the events
  `agent_run_create` / `agent_run_user_prompt` while the sample payloads carry
  `type: "agent_run"` / `agent_run_activity` (and their own handler filters on
  `type !== "agent_run_activity"`). Do not guess: log the header on the first
  live delivery, dispatch on `type` + `action` with the header as a hint, and
  treat an unknown pair as retryable rather than dropped.

### Delivery guarantees

The documented policy for workspace webhooks ([intro webhooks]): asynchronous
POST, any 2xx is success, failures retry with exponential backoff at a
**~10-minute base with jitter** up to **5 attempts**, after which the webhook is
**disabled** and the creator emailed; **4xx is never retried**. The AGPL sender
matches it (`autoretry_for=(requests.RequestException,)`, `retry_backoff=600`,
`max_retries=5`, deactivation at exhaustion) and uses a 30s request timeout.

For app webhooks none of this is documented. Design for the pessimistic case:

- never answer a non-2xx to a delivery that might still be processed — a 4xx
  means "never retry", so a validation bug would silently lose a user turn;
- a 5xx costs one of five attempts and, at the end, the app's webhook may be
  switched off, which is a silent, install-level outage;
- acknowledgement must be fast. Plane's guidance is "return 200 within
  seconds and process asynchronously", which is what the durable-queue
  receiver already does.

### Ordering

Undocumented, and not to be assumed. `agent_run_create` and the first
`agent_run_user_prompt` describe the same trigger and can be delivered in
either order or with one repeated. Butter must therefore be correct when a
prompt arrives before its run is known: the prompt is self-sufficient (it names
`agent_run.id`, `comment_id`, `issue_id` and `project_id`), so it can be
queued and executed, with the create event only enriching what is stored.

## 4. Activity contract

Create path: `POST /api/v1/workspaces/{slug}/runs/{run_id}/activities/`
(list/retrieve are also available and are what a reconciliation pass should
use). Request body, per the SDK types:

```ts
interface CreateAgentRunActivityRequest {
  content: { type: "thought"|"action"|"response"|"elicitation"|"error"; ... };
  content_metadata?: Record<string, any>;
  signal?: "auth_request" | "continue" | "select" | "stop";
  signal_metadata?: Record<string, any>;
  type: Exclude<AgentRunActivityType, "prompt">;
  project?: string;
}
```

Response activities carry what the request cannot set: `ephemeral`,
`comment`, `actor`, and the timestamps.

### Ephemerality is a server rule

`ephemeral` is **absent from the create request** and present on the response.
The documented rule is a type list: `thought`, `action`, **`error`** are
"automatically marked as ephemeral" ([building an agent], [signals and content
payload]). The same two pages also say an `error` "creates a comment and sets
the run to failed", which contradicts the list.

This cannot be resolved by reading harder; the SDK proves the client has no
say, so the server decides, and both pages are describing one rule. A
TypeScript spike confirms the shape (see appendix):

- a create with `ephemeral: true` is rejected by the type checker ("Object
  literal may only specify known properties");
- `type: "prompt"` is rejected outright — only Plane creates prompts;
- `parameters` values must be `string`; structured data belongs in
  `content_metadata`;
- `signal_metadata` is `Record<string, any>`, so nothing type-checks the option
  shape.

**Implication for Butter:** an `error` activity is a status change, not a
message. A failure the user must read is a `response` (or `elicitation`)
followed by the `error`, exactly as ADR-0015's "never lose the reason" rule
already wants. And `failed` must be inferred from the activity, not from
whether a comment appeared.

### Content by type

| type | content | ephemeral | effect |
| --- | --- | --- | --- |
| `thought` | `{type, body}` | yes | progress only |
| `action` | `{type, action, parameters: {string: string}}` | yes | tool progress; result goes in `parameters.result` or `content_metadata` |
| `response` | `{type, body}` (Markdown) | no | comment reply |
| `elicitation` | `{type, body}` (Markdown) | no | comment reply, run → `awaiting` |
| `error` | `{type, body}` | **yes** (per the automatic list; the same pages also say "comment") | run → `failed` |

`signal: "stop"` is never created by the agent — it is what Plane **sends** in
a prompt when a user stops the run. `signal: "auth_request"` requires
`signal_metadata.url` starting with `https://`. `signal: "select"` carries
options as

```json
{"options":[{"id":"frontend","label":"Frontend App","description":"…"}],"allow_multiple":false}
```

Note the shape: **`id` + `label`**, not Linear's `label` + `value`. The
Linear→Plane activity mapping therefore is not mechanical:

- Linear `action` uses `parameter` (one string); Plane wants
  `parameters` (`map<string,string>`);
- Linear `elicitation` `signalMetadata.options` is `{label, value}`; Plane is
  `{id, label, description?}`;
- Linear has an external-URL update on the session; Plane has nothing of the
  kind (see below).

### Idempotency, limits, throttling

Nothing documents an idempotency key on activity creation, and the SDK sends a
plain POST. Assume **at-least-once is not even offered**: a retried `response`
can duplicate a comment. The retry boundary must stay ADR-0009's — persist the
reply, mark the record, post once, and make a resend an explicit operator
action that first lists activities and compares bodies. Butter's Linear
progress throttle (one `action` per 3s, consecutive repeats dropped,
`internal/runtime/linear/progress.go`) should be reused as-is; Plane's only
guidance is "send thoughts for milestones, not for every internal call" and
its own heartbeat advice of one update per minute.

## 5. Run lifecycle, reuse and races

Statuses ([signals and content payload], [building an agent]):
`created`, `in_progress`, `awaiting`, `completed`, `stopping`, `stopped`,
`failed`, `stale`; `ended_at` / `stopped_at` / `stopped_by` / `error_metadata`
are set by Plane.

- **Stale is a real deadline.** A run "not updated in 5 minutes" is marked
  `stale`. Updates are agent activities, so any turn that can exceed five
  minutes must heartbeat a `thought` about once a minute. Butter's default
  `max_run_seconds` is 1800 — a heartbeat is mandatory, not cosmetic, and it
  must be independent of whether the Agent produced visible progress.
- **Follow-ups reuse the run while it is active.** "If an active Agent Run
  exists for that thread, the reply is added as a new activity." Whether a
  reply after `completed` / `failed` / `stale` opens a **new** run (and a new
  `agent_run_create`) is not stated, but it follows from "active" and from the
  run's terminal fields; the first live test should confirm it, because it is
  what decides whether a second question starts a fresh Butter session.
- **Stop is a prompt, not an endpoint.** Plane sends a prompt activity with
  `signal: "stop"` and expects a confirmation response, which "transitions the
  run to `stopped`". Two races follow directly:
  - a stop that arrives while a turn is running must not wait behind it —
    ADR-0015's out-of-band marker + nudge is the right shape, and the run's
    status (`stopping`) is Plane's own cue that someone is waiting;
  - a stop that arrives after the turn already posted its `response` must not
    produce a second confirmation. The holder knows whether it posted;
    the record's terminal state decides, not the delivery order.
- **No `external_link`.** `external_link` exists on the run and on
  `CreateAgentRunRequest`, but Plane creates the run and the SDK exposes no
  update. The dashboard deep link can only appear in the agent's first
  `response` body (or `content_metadata` on an activity). Keep
  `externalURL` derivation for the dashboard's own processing records.

## 6. What is now answered, and what is not

| Question from the issue | Answer |
| --- | --- |
| Target edition / minimum version | Cloud only today; not in AGPL at any version; Commercial/Airgapped unverified, version unspecified. Verify per instance by the presence of "Enable App Mentions". |
| OAuth app webhook secret | Undocumented. Store it as configuration; test Client Secret vs. an app-form secret on the first delivery. |
| Install callback `state` | Does not exist in the bot flow; the callback carries `app_installation_id` (and an unused `code`). |
| Real headers / order / retries | Headers as documented; order unguaranteed; retries 5×, ~10 min base, 4xx never, webhook disabled on exhaustion (documented for workspace webhooks, unverified for app webhooks). |
| `X-Plane-Event` vs payload `type` | They disagree in the docs; record the real header pair before coding dispatch. |
| Business dedupe key | Installation + `agent_run.id` + prompt `activity.id`; `X-Plane-Delivery` is regenerated per retry in the AGPL sender. |
| Stop/complete race, run reuse | Stop is a prompt with `signal:"stop"`; confirm exactly once; follow-ups stay in the run while active; terminal runs presumably open a new one (verify). |
| Activity POST idempotency | None documented; treat as non-idempotent, reconcile by listing. |
| `error` persistence | Contradictory docs; `ephemeral` is server-side by type, so the body may never be a comment. Never depend on it. |
| `external_link` backfill | Impossible — no run update endpoint; use the first `response`. |
| Select result format | Options are `{id, label, description?}` + `allow_multiple`; the user's answer arrives as a normal prompt body, i.e. free text or an option label. |

### Still needs a live instance

1. Webhook secret source, and whether `X-Plane-Event` is `agent_run_*` or
   `agent_run`.
2. App-webhook retry policy and request timeout.
3. `error` activity: `ephemeral` flag and whether `comment` is set.
4. Whether a reply after a terminal run opens a new run (and fires
   `agent_run_create` again).
5. Whether `awaiting` runs also go `stale`, and whether an agent heartbeat
   resets `stale` (docs say yes; unverified).
6. Activity rate limits / max body length.
7. Cloud plan availability (Free vs paid).

## 7. What carries over from ADR-0015, and what changes

Reusable as-is, because they are already protocol-neutral or trivially
re-parameterized:

- `internal/eventqueue` — dedupe-and-append, consumer group, reclaim, touch,
  ack: nothing Linear-specific; a second Stream key is the whole change.
- The session coordinator's contract and Lua scripts
  (`internal/runtime/linear/coordination.go`): the enqueue-or-acquire /
  release-or-drain / stop-marker-by-lease-token design is exactly the Plane
  stop race. Only the four Redis key prefixes (`butter:linear:lease:session:`,
  `:followups:`, `:stop:`, `:stop-nudge:`) are Linear-shaped; parameterize them
  (or hoist the package) and run its contract test suite twice.
- The processing record and its states (`RECEIVED` → … → `READY_TO_DELIVER` →
  `SUCCEEDED` / `FAILED` / `FAILED_UNCERTAIN`, resend that never re-invokes the
  Agent): mapping one-to-one onto Plane, with the added "list activities before
  resend" step.
- `worker` shape: one goroutine per claimed event, per-Pod bound, heartbeated
  lease, never block reading on a turn.
- `progress`: throttled `action` activities from tool calls, redacted via
  `internal/redact`.

New work specific to Plane:

- `internal/planeapi` — OAuth (`authorize-app` URL builder, `client_credentials`
  mint, `app-installation` lookup) and REST (`GET runs/{id}`,
  `GET/POST runs/{id}/activities`), typed errors, plus an in-process fake as
  `lineartest` does.
- `internal/runtime/planeconn` — mint-and-cache bot token, mint again on 401,
  no refresh/rotation state.
- `internal/runtime/plane` — receiver (raw-body HMAC, `app_client_id` check,
  business dedupe), worker, orchestrator, activity mapping, and the
  first-touch context fetch (work item + comment thread over REST, length
  capped, because no `promptContext` exists).
- `proto/agents/v1/plane.proto` — `PlaneApp` (client_id, agent_id,
  inbound_enabled, allowed_user_ids, max_run_seconds, revision,
  credential_state, `webhook_url`, install callback URL), `PlaneInstallation`
  (workspace UUID + slug, `app_installation_id`, `app_bot` user, scopes,
  status, credential_state), `PlaneProcessingRecord` (mirroring Linear's),
  plus the three services the dashboard needs.
- **Activity mapping deltas** — `parameters` instead of `parameter`; select
  options `{id, label}` instead of `{label, value}`; `response` before `error`;
  never an `error` as the only user-visible message.
- **Session identity** — `plane:{app_id}:{agent_run_id}:{agent_id}`, ADK app
  name `plane`, user `plane:{app_id}:{plane_workspace_id}`, prompting actor as
  the Memory Capture principal. Repointing the App to another Agent starts a
  new history, as in Linear.
- **Context fetch** — scopes `projects.work_items:read` and
  `projects.work_items.comments:read` (plus the four `agents.*` scopes) are
  what the first turn needs; a bot without them gets an empty prompt, so
  installation should record and surface the granted scope string.

## 8. Minimal chain verification runbook

To be run once against a named Cloud workspace (a tunnel to a local receiver is
fine). Each step has a pass condition; steps 1–5 are the protocol confirmation
that unblocks `ready-for-agent`, 6–8 are the reliability edges.

1. **Register** the app with "Enable App Mentions", scopes
   `agents.runs:read agents.runs:write agents.run_activities:read
   agents.run_activities:write projects.work_items:read
   projects.work_items.comments:read`, Webhook URL pointing at the receiver.
   *Pass: the app appears in the mention picker.*
2. **Install** and mint a token from the `app_installation_id` in the redirect.
   *Pass: `app-installation` returns the slug and `app_bot`; a `GET runs/…`
   with the token is 404 (unknown run) rather than 401.*
3. **Mention** the agent in a work item comment with the receiver recording raw
   bodies and headers. *Pass: both deliveries captured; record the
   `X-Plane-Event` values, the signature header, and which secret verifies.*
4. **Answer** with a `thought` then a `response`. *Pass: the response appears as
   a comment reply and the run leaves `created`.*
5. **Ask a second question** in the same thread. *Pass: the prompt names the
   same `agent_run.id`; the turn continues the Butter session.*
6. **Stop** mid-turn and after completion. *Pass: exactly one confirmation
   response per stop; the run ends `stopped`.*
7. **Fail** a turn: `response` followed by `error`. *Pass: the reason is
   visible; record whether the `error` has `ephemeral: true` and a null
   `comment`.*
8. **Replay** a captured delivery verbatim, then send two prompts back to back.
   *Pass: one turn for the replay (business dedupe), two serialized turns for
   the pair, no duplicate comments.*

## 9. Recommendation

The protocol is now specific enough to design against, and the design is
ADR-0015 with a second protocol adapter — not a second architecture. Two
caveats should stay in the issue:

- the acceptance item "在目标 Plane 实例完成真实安装" must name a Cloud (or
  Commercial) instance, because Community Edition can never satisfy it;
- `error` and `external_link` behave differently from the issue's assumptions,
  so any spec written from those sentences needs the corrections in §4 and §5.

Do not mark the issue `ready-for-agent` until runbook steps 3–5 have run once:
those are the three places where the docs contradict themselves or each other,
and they decide the shape of the receiver, the mapping and the session key.

## Appendix: how the claims were checked

Docs, fetched 2026-11-03 (`curl` + HTML-to-text):

- `dev-tools/agents/overview`, `building-an-agent`, `best-practices`,
  `signals-content-payload`
- `dev-tools/build-plane-app/overview`, `create-oauth-application`,
  `choose-token-flow`, `oauth-scopes`, `webhooks`, `examples`
- `dev-tools/intro-webhooks`, `dev-tools/openapi-specification`,
  `self-hosting/editions-and-versions`

The public API reference (`/api-reference/*`, 329 pages) has **no** agent-run or
activity endpoint — the SDK is the only published schema.

SDK: `npm i @makeplane/plane-node-sdk@0.3.1 typescript@5`, then `tsc --strict`
over a file that builds each `CreateAgentRunActivityRequest`. Positive cases
(`thought`, `action`, `response`, `elicitation` + `select`, `elicitation` +
`auth_request`, `error`, `CreateAgentRunRequest` with `external_link`) compile;
negative cases fail exactly as reported:

```
neg/a.ts: 'ephemeral' does not exist in type 'CreateAgentRunActivityRequest'
neg/b.ts: Type '"prompt"' is not assignable to type '"action" | "elicitation" | "error" | "response" | "thought"'
neg/c.ts: Type 'number' is not assignable to type 'string'   (action parameters)
neg/d.ts: compiles — signal_metadata is Record<string, any>, options unverified
```

Plane source: `git clone --depth 1 https://github.com/makeplane/plane` at
`618d0eb7`; `grep -rn` for the feature's identifiers returns nothing outside
i18n, and `apps/api/plane/bgtasks/webhook_task.py` supplies the header,
signature and retry facts quoted above.

Live probes: unauthenticated `GET`/`POST` against `api.plane.so` paths, status
codes recorded in §1.

[butter#487]: https://github.com/orvice/butter/issues/487
[editions and versions]: https://developers.plane.so/self-hosting/editions-and-versions
[building an agent]: https://developers.plane.so/dev-tools/agents/building-an-agent
[signals and content payload]: https://developers.plane.so/dev-tools/agents/signals-content-payload
[handling webhooks]: https://developers.plane.so/dev-tools/build-plane-app/webhooks
[intro webhooks]: https://developers.plane.so/dev-tools/intro-webhooks
