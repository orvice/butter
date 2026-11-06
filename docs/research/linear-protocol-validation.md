# Linear Agent Session protocol validation

Status: Stage 1 evidence package (Stage 2 updated it for the accepted fixes). The repository checkout was at commit `e91a219` on 2026-11-05. No Linear administrator or deployed test organization was available in this run, so the protocol observations below separate implementation evidence from real-environment evidence. An unknown `AgentSessionEvent` action is now acknowledged and ignored (Stage 2, `e12f691`), not queued.

## Scope and evidence boundary

The requested live actions require a Linear OAuth app, an installed test organization, an administrator who can delegate and mention the app, and a deployed Butter webhook URL. None of those credentials or endpoints were available. No request was sent to Linear, and no token, signature, full `promptContext`, or full Issue body is included here. Existing Go tests are implementation checks; they do not establish the official payload contract.

| Item | Status | Evidence |
| --- | --- | --- |
| Delegation opens an Agent Session and emits `created` | **Unverified live**; implementation-supported | ADR-0015 §Context/§4; `internal/runtime/linear/event.go`; `internal/handler/http/linear_webhook_test.go` |
| @mention opens an Agent Session and carries the mentioning comment | **Unverified live**; fallback implemented | ADR-0015 §Context/§5; `Receiver.Deliver` uses `agentSession.comment.body` when activity text is empty |
| Follow-up prompt is `prompted` and text is in `agentActivity.content.body` | **Unverified live**; implementation-supported | `webhookPayload.AgentActivity`; `TestAcceptedDeliveryIsAFrozenSnapshot` |
| Stop is represented by `agentActivity.signal = "stop"` | **Unverified live**; implementation-supported | `TestStopSignalIsCarried`; `Receiver.Deliver` sets `Event.Stop` |
| Success reply is a `response` activity | **Unverified live**; implementation-supported | `Orchestrator.run`; `linearapi.CreateActivity` |
| Failure reply is an `error` activity | **Unverified live**; implementation-supported | `turn.fail`; ADR-0015 §9 |
| Redelivery uses `Linear-Delivery` | **Unverified live**; implementation-supported | `deliveryID`; `TestRedeliveriesAreDeduplicated` |
| Missing delivery header can be deduplicated by body SHA-256 | **Unverified live**; implementation-supported fallback | `deliveryID` |
| Top-level `organizationId` identifies the installation | **Unverified live**; implementation-supported | `Receiver.installationFor` |
| Missing organization ID falls back to one installation only | **Unverified live**; implementation-supported | ADR-0015 §3; `installationFor` |
| Prompting user identity is activity user, then session creator, then comment user | **Unverified live**; implementation fallback | `webhookPayload.promptingUserID` |
| `webhookTimestamp` is Unix milliseconds and signed body is HMAC-SHA256 hex | **Unverified live**; implementation-supported | `freshTimestamp`, `validSignature`, webhook tests |
| ±60 second replay window and 1 MiB body cap | **Code-confirmed** | `webhookMaxSkew`, `linearMaxWebhookBytes` |
| 200 means accepted, duplicate, or intentionally ignored; queue failure is 503 | **Code-confirmed** | `LinearWebhookHandler`, `Receiver.Deliver` |
| Activity response limit is 8000 runes | **ADR/implementation assumption; live limit unverified** | ADR-0015 §9; `truncateResponse` |

## Field mapping and semantics

| Linear source path | Butter event field | Unit/encoding | Behavior |
| --- | --- | --- | --- |
| `type` | routing discriminator only | UTF-8 string | Only `AgentSessionEvent` is queued; other types are acknowledged and ignored. |
| `action` | `Event.Action` | UTF-8 string; expected `created`/`prompted` | `created` gets an early acknowledgement thought; both actions can run a turn. Any other action on a signed `AgentSessionEvent` is acknowledged (200) and ignored, so an unconfirmed protocol action can never start an Agent. |
| `organizationId` | `Event.OrganizationID` after installation lookup | Linear ID, UTF-8 | Installation is selected by `(app, organizationId)`. Empty is accepted only when exactly one installation exists. |
| `webhookTimestamp` | freshness check | Unix milliseconds | Accepted only within ±60 seconds of server time. It is not persisted as the event timestamp. |
| `agentSession.id` | `Event.AgentSessionID` | Linear ID, UTF-8 | Required. Session key is `linear:{app}:{agentSession}:{agent}`. |
| `agentSession.creatorId` / `creator.id` | prompting-user fallback | Linear user ID, UTF-8 | Used after activity user fields are absent. This is not a verified Butter identity. |
| `agentSession.comment.body` | `Event.PromptText` fallback | UTF-8 text | Used for a mention-created session when no activity body exists. The complete comment is not copied into this evidence package. |
| `agentActivity.userId` / `user.id` | `Event.PromptingUserID` | Linear user ID, UTF-8 | Preferred prompting identity when present. |
| `agentActivity.content.body` | `Event.PromptText` | UTF-8 text, surrounding whitespace trimmed | Follow-up input. `promptContext` is separately retained for the first turn. |
| `agentActivity.signal` | `Event.Stop` | UTF-8 enum-like string | Exact `stop` marks a stop request. Other signals are not treated as stop. |
| `promptContext` | `Event.PromptContext` | UTF-8 formatted context | Passed only to the first turn when the Butter session has no history. It must not be committed to fixtures or logs in full. |
| `Linear-Delivery` | `Event.DeliveryID` | UTF-8 header | Primary dedupe key, scoped by App in the queue. |
| raw request body when delivery header is absent | `Event.DeliveryID` | lowercase hex SHA-256 prefixed `sha256-` | Fallback dedupe key. It is byte-sensitive and therefore requires the exact raw body. |
| `Linear-Signature` | authentication input | lowercase/uppercase-insensitive hex HMAC-SHA256 | HMAC covers the exact raw body; comparison is constant-time. |

The app identity is the installation's `viewer.id` obtained during OAuth. Butter stores it as `LinearInstallation.app_user_id`; the routed Butter Agent is represented by that app user in Linear. The prompting user is a separate Linear user and is stored only as a principal (`linear:{user_id}`) for Memory Capture. The ADK session user is the organization-scoped value `linear:{app}:{organization}` so a follow-up from another participant continues the same Agent Session.

A non-empty allowlist is fail-closed when the prompting user cannot be resolved: `admitted` returns false for an empty user ID. An empty allowlist admits all users, as configured by the App. This is an authorization choice, not proof that the payload's user field is authoritative; the live field source still needs confirmation.

## Delivery and retry rules

The receive path verifies the signature and timestamp before JSON interpretation, then resolves the installation and app state before appending a frozen event to `butter:linear:events`. The queue deduplicates by App plus `Linear-Delivery`, or App plus body hash when the header is missing. A duplicate returns HTTP 200 with a duplicate decision. Infrastructure failure returns HTTP 503 so Linear can retry. Invalid signature or stale timestamp returns HTTP 401; malformed JSON returns HTTP 400.

The processing record follows the ADR-0009 boundary. Failures before the Agent may have run are retryable. Once the Agent may have run, a runner failure is `FAILED_UNCERTAIN` and is never automatically rerun. A completed reply is persisted as `READY_TO_DELIVER` before the activity is posted; resend posts that reply and does not invoke the Agent. A Stop cancels the active turn and discards queued follow-ups; it is confirmed after the holder can post no more turn activity.

Activities map as follows: `created` acknowledgement and operational progress use ephemeral `thought`/`action` where allowed; tool calls become redacted, rate-limited `action`; workflow pauses become `elicitation` and may use `signal=select` plus options; successful output is `response`; failures and timeouts are `error`. The response truncation policy is currently an implementation/ADR assumption until a real long-response test confirms Linear's actual limit.

## Required live validation run

An administrator should run the fixtures in this order against a disposable organization and record the UTC timestamp, Butter deployment commit, App ID, organization ID, session ID, delivery ID, HTTP status, observed webhook field paths, and resulting activity types. Record only IDs, lengths, hashes, and redacted snippets.

1. Delegate a disposable issue to the app. Capture the `created` webhook shape, installation organization field, creator fields, `promptContext` presence, and the first `thought` followed by a short `response`.
2. Mention the app in a disposable issue comment. Capture whether the same or a new session is opened, where the comment text appears, and which user field identifies the prompting user.
3. Send a follow-up in the session. Capture `prompted`, activity identity, body path, and whether the Butter session remains the same.
4. Start a long-running disposable turn, send Stop, and capture the stop webhook, cancellation timing, queued-message behavior, and final confirmation activity.
5. Force a delivery retry by replaying the same signed body and delivery ID. Confirm one processing record and one Agent run. Repeat without `Linear-Delivery` only if Linear can produce that case safely, and record whether the raw body is byte-identical.
6. Produce a successful long response and a controlled Agent failure. Record the activity types, body lengths, truncation marker, and whether Linear accepts the full response.

The live results should replace the “Unverified live” rows above. Any mismatch in identity precedence, organization routing, stop signal, activity limits, or dedupe behavior is a Stage 2 security/protocol fix input.

Stage 2 result: no protocol defect was proven from this package, so the fixes were limited to ignoring unsupported session actions and canonicalizing allowlist IDs (case and surrounding whitespace). The fixture remains `synthetic_shape_only`; the live rows above are still unverified. The regression test `TestStageOneFixturesAreAcceptedAtTheWebhook` replays the sanitized request shapes through the public webhook route to keep the documented fields, dedupe and organization routing consistent with the fixture.

## Stage 2 repair inputs

- Confirm the official source and precedence for prompting user identity. If a non-empty allowlist is configured and the source is absent or ambiguous, reject the event and emit an error without running the Agent.
- Confirm whether `organizationId` is always present and whether its value is the organization/workspace ID used by OAuth installation lookup. Keep the single-installation fallback only if live behavior supports it.
- Confirm whether `Linear-Delivery` is stable across retries and whether its scope is App-local. If it is absent or mutable, strengthen the dedupe key with a verified event identifier.
- Confirm Stop's exact action/signal shape and whether a stop can arrive without `agentActivity.content` or a prompting user.
- Confirm activity body and long-response limits, including Unicode rune versus byte behavior, then update `truncateResponse` and the resend fixture if needed.
- Confirm timestamp unit, accepted skew, and signature encoding against a captured request. Keep raw-body HMAC verification and constant-time comparison regardless of the observed formatting.
- Confirm whether `created`, `prompted`, and a follow-up after an elicitation have distinct action semantics; ensure resume remains FIFO and does not create a second pending-input store.
