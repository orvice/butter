# Plane Prompt batches and recovery

PR #520 completes the follow-up execution path in #502 and the related Stop
recovery in #503. This is a local execution contract; it does not certify the
target Plane deployment's Gate A/B protocol acceptance.

## Execution contract

The existing `butter:plane:events` Redis Stream remains the only pending
payload store. Admission freezes the Agent ID, Activity actor and accepted
order. A per-Run sorted order index and an order-to-Stream-ID hash contain
references, not another copy of pending messages.

An eligible holder selects up to 32 consecutive prompts, with a 1 MiB
combined text budget. Each text part also passes `internal/userinput.Turn`.
A batch that would exceed its limit splits at a Prompt boundary. An
individually invalid input is explicitly failed and retired, so it cannot
block subsequent valid messages. User text is never sliced to fit a batch.

```text
durable admission
  -> reserved batch + monotonic Mongo claim fence
  -> processing records for every member
  -> atomic StartBatch markers
  -> one Runner invocation
  -> Stop/Finish arbitration
  -> stage complete output in the durable delivery store
  -> project execution outcome
  -> send one final response
  -> atomically settle every member and release the Run
```

Every Prompt keeps an independent processing record. Batch members share
one Invocation ID and ordered Activity ID list. A queue acknowledgment is
a `thought` with its own durable delivery key, not an Agent `response`.
Execution state, queue confirmation delivery and final response delivery
are independent fields. A delivery error never reruns a successful Agent.
Outgoing activity text passes the shared best-effort credential redactor.

The ADK app is `plane`. Session identity includes Butter workspace, App,
local Installation, external Run and frozen Agent. The ADK user is an
installation-scoped service identity. Create is ignored as model input;
late or repeated Create never clears the conversation or pending work.
An installation's current trusted workspace slug and external installation
ID are resolved for outbound activity routing.

The current App binding, inbound state and actor allowlist are checked
again before execution. Missing actors and Bot actors cannot execute.
Conflicting Activity actor aliases are rejected; the Run creator is never
an actor fallback. Every member's provenance is retained in records and
`ContextInfo.metadata.plane_prompts`. A single actor contributes
`principal=plane:<actor>`. Mixed or unknown provenance suppresses automatic
Memory Capture with a recorded reason.

Workflow-bearing Agent trees consume one exact Prompt per turn. They do
not join multiple answers or prepend provenance text to an Interrupt
answer. This preserves the existing oldest-Interrupt reply contract.

## Recovery boundaries

| Failure window | Recovery |
| --- | --- |
| Before `StartBatch` | No started markers exist. The expired reservation can be acquired again; a new claim fence prevents the former holder's projection from overwriting it. |
| After `StartBatch`, while the holder is live | Other Workers defer every selected member. They do not acknowledge another holder's running batch. |
| Started holder disappears without settlement | Members without a known terminal outcome recover the original Invocation ID as `FAILED_UNCERTAIN`; the scoped active Invocation is failed without changing terminal Invocations. The Agent is never called again for those messages. |
| Terminal processing projection fails | Persistent per-event execution outcomes repair the records on pending replay. The batch association, actor provenance and memory decision survive a singleton replay. |
| Complete reply staged, but not sent | The durable publisher recovers delivery from the saved output, without the Runner. |
| POST may have reached Plane | The delivery remains `unknown` and requires the existing owner-confirmed retry flow. It is not automatically posted again. |
| Execution settled but no reply was durably recorded | Recovery records reply delivery as `not_recorded`. There is no automatic reconstruction or model rerun. |
| PEL entry moves to another consumer | Atomic owner-checked acknowledgment prevents the former consumer from deleting the new consumer's work. |

Workers page through both their own pending entries and reclaimable
entries. An active head does not permanently hide deferred work behind it.
Own pending work is retried promptly; another Pod's work is reclaimed after
the existing two-minute idle bound. The execution lease remains 30 seconds.
Stop has its existing independent control pool and admission cutoff, and
also cancels a reserved batch before its execution boundary.

Processing projections use Mongo compare-and-swap and a monotonic claim
fence. Late admission/queued projections cannot replace processing or
terminal outcomes. Terminal execution is immutable; confirmed delivery
cannot regress to an unconfirmed delivery. Legacy processing keys/status
overlays remain readable and are upgraded when projected.

## Deployment and verification

Upgrade all Plane Workers together. Old Workers do not understand batch
reservations, the new session identity or the new projection fence. Do not
mix old and new execution Workers while accepting Plane traffic. Drain
existing turns before switching; do not delete the Streams or started
markers. Unstarted legacy entries without a frozen Agent or valid actor
fail closed rather than being silently handed to the current Agent.

The started markers and per-event outcomes have no automatic expiry, as
with the pre-existing started markers. This avoids re-execution after a
delayed pending replay; retention/compaction needs a separate lifecycle
policy before removing those facts.

After deploying:

1. Hold a long Agent turn open. Send two more messages in the same Run,
   including a second admitted actor. Verify `Queued` records and separate
   system thoughts, then one following Invocation and one final response
   for the two queued messages.
2. In Dashboard's Plane processing panel, verify each Activity's actor,
   frozen Agent, shared Invocation, batch member IDs, execution outcome
   and independent delivery statuses. The mixed-actor batch must show a
   Memory Capture suppression reason.
3. Repeat with enough text to exceed 1 MiB combined. Verify successive
   Invocations preserve each complete Prompt in accepted order.
4. While that Run is busy, prompt an independent Run. Verify it executes
   concurrently.
5. Stop a running or reserved turn, then send a new message. Verify old
   queued messages are cancelled, the old turn does not publish a late
   result after Stop wins, and the new message can continue.
6. Restart the holder Pod during execution. After lease expiry/reclaim,
   verify uncertain started messages retain their Invocation ID and are
   never invoked again. New messages must eventually run.
7. Verify same-Run history survives completion/failure/Stop where the
   target Plane deployment permits it. A different Installation, Run or
   Agent must use a different history. Real Plane lifecycle acceptance
   remains part of the protocol verification.

Automated coverage uses real Redis and MongoDB for batches, two Workers,
input splitting, actor policy, crash boundaries, projection repair,
durable reply recovery, Stop reservations, Lua permission failure
atomicity, cursor recovery and stale-owner acknowledgment. The Invocation
recovery contract runs against both memory and Mongo stores.
