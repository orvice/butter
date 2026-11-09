# Plane webhook acknowledgements and asynchronous work

The callback is `POST /api/plane/webhook/{app_id}`. Plane can disable a callback
after repeated HTTP delivery failures. A transport acknowledgement must therefore
be distinguished from execution success, without turning a lost message into a
false success.

## Response contract

| Result | HTTP / body | Durable effect before responding |
| --- | --- | --- |
| Valid Agent input | 200, `decision=accepted` | Redis has appended/deduplicated the authorized event and assigned its order/Stop marker. |
| Previously admitted input | 200, `decision=duplicate` | Existing business deduplication owns the event; no second execution. |
| Explicitly supported ignored notification/Create/disabled normal inbound | 200, `decision=ignored` | No work is promised or scheduled. Signature and applicable validation still run. |
| Authenticated permanent protocol/identity rejection | 200, `decision=rejected`, `receipt_id` | A scoped permanent diagnostic receipt has been written. No Agent execution or Stop is scheduled. |
| Wrong/missing signature | 401 | No receipt, queue entry or execution. |
| Unknown App | 404 | No work. |
| Body over 1 MiB | 413 | Not authenticated/accepted as a complete body. |
| App/credential lookup, Redis admission or rejection-receipt persistence unavailable | 503 | No false success. The provider may retry. |

**Compatibility change:** signed malformed JSON, unsupported schema, missing or
conflicting Agent identities, insufficient grants, and actor-policy rejections
used to return 400/401. They now receive a 200 **rejected** acknowledgement only
once the diagnostic is persisted. `Receiver.Deliver` itself still returns the
same strict validation errors; the HTTP adapter owns the receipt/ACK policy.
Unauthenticated input does not become successful or get a diagnostic write.

A 200 is never a statement that an Invocation succeeded or the Plane Run is
completed. Consumers must distinguish `accepted`, `ignored` and `rejected` and
follow processing/Invocation/delivery outcomes for admitted work. A rejected attempt is not made authorized by merely retrying it without fixing
its cause; later software/permission changes do not rewrite that past receipt.

Unknown signed events are not blanket-ignored or executed. Unless a known
notification route can safely ignore them, they follow strict validation and
produce a permanent rejection receipt. This avoids guessing whether a new
event is an ordinary notification, Prompt, Stop or another control variant.

## What runs asynchronously

The model already runs in the existing Redis-backed worker, not the HTTP
request. The receiver still authenticates and authorizes before publishing an
execution event or Stop. It does not launch an in-memory fire-and-forget task
and then claim success.

After Redis admission, the HTTP path now returns without the best-effort Mongo
`RECEIVED` projection. The worker writes that projection after resolving the
scoped App and before processing. A projection failure leaves the Stream entry
pending and cannot masquerade as a terminal input error. Recovery retries the
projection before invoking the model, preserving the existing may-have-run
boundary, deduplication, ordering and Stop arbitration. Management reads can
briefly lag the transport ACK; they must not interpret a missing projection as
proof that Redis did not accept the event.

A caller disconnecting after ACK does not cancel the worker. Redis/diagnostic
failures still return 503: eliminating every possible HTTP failure would need
a separate reliable store, not an unconditional 200 or request goroutine.

## Rejection receipts are not an inbox

- Receipt keys fingerprint the trusted App/workspace and exact body, allowing
  correlation/merging of repeated rejections without storing the body. They
  are not Invocation IDs, credentials, authorization tokens or replay handles.
- Records live in `plane_diagnostics`, scoped by the server-resolved workspace
  and App, with `class=permanent` and `recoverable=false`. No unverified payload
  identity is used as the record's tenant or foreign-key authority.
- Only a bounded safe error summary and event-name fingerprint are stored.
  Unknown identifiers, prompts, bodies, tokens, secrets and signatures are not
  written to the diagnostic or logs. `event_hash` can be compared to public
  event-name candidates; it is not a routing signal.
- The HTTP adapter bounds the receipt write to two seconds and acknowledges it
  only on success. Cancellation, absent persistence and failed writes remain
  retryable 503s. Logs distinguish rejection recorded from transport ACK.
- New receipt records have a 30-day `expires_at` TTL (Mongo's periodic expiry).
  Existing execution diagnostics are not assigned a new expiry or deleted.
  Recent diagnostic lists return at most 200 entries, newest first.
- The Mongo adapter now writes explicit BSON names matching its queries/indexes.
  Legacy lowercase field names remain readable; legacy values cannot override
  a present canonical workspace/App field. No destructive migration is run.

The receipt is a **terminal rejection decision**, not durable storage of an
original message awaiting future parsing. There is no automatic replay and no
new public receipt-fetch RPC. Operators can correlate the opaque key with scoped
diagnostics/logs; detailed management UI integration remains tracked separately.
Do not label a rejected receipt as queued or silently replay an uncertain run.

## Restoring a callback disabled by Plane

1. Merge the change and wait for the main Docker build/push/sign to succeed.
2. Deploy and verify the actual startup commit, not merely Pod Ready status.
3. Re-enable the disabled webhook in Plane using an authorized operator. Butter
   does not do this automatically or request broader OAuth scopes.
4. Send a fresh mention. Check for an `accepted` ACK, asynchronous Invocation,
   terminal outcome and confirmed reply. Normal ignored events and recorded
   rejections should no longer accumulate HTTP 400/401 delivery failures.
5. Investigate `rejected` business diagnostics rather than treating HTTP 200 as
   proof every notification type is supported. Real signature/configuration or
   persistence outages can still fail HTTP delivery and must not be hidden.

Do not bulk-replay old failed callbacks as part of re-enabling the transport.
Some may have been admitted/executed already; use business identities and the
existing processing/delivery records before deciding any recovery action.
