# Plane Run control

For webhook identity rejection and Runs stuck before admission, see
[Plane webhook identity and stuck Run recovery](plane-webhook-recovery.md).

Plane admission assigns every accepted event a Redis sequence number.
The signed webhook receiver authenticates the App, installation and Stop
actor before admission. In one Lua operation, admission deduplicates the
event, assigns its order and appends its frozen payload.

Prompts use the existing `butter:plane:events` Stream. Stops use
`butter:plane:controls`, with an independent consumer pool. At Stop admission,
the same operation records the Run's cutoff and marks its current running
holder by token. Cancellation therefore does not wait for a Worker or a free
model slot. The cutoff prevents any older unstarted Prompt from beginning,
including entries a Worker has not read. Newer accepted Prompts are retained.
The Lua operation checks key types and command permissions before changing
the cutoff or holder.

## Holder and completion

One Run control hash is scoped by Butter workspace, App, installation and
Plane Run ID. Each turn gets a fresh token. The Processor polls and renews
this control fact every 250 ms and renews its Mongo claim while the turn
lives. A backend error cancels the turn conservatively; loss of a token
suppresses subsequent result writes. There is no pub/sub dependency: every
poll reads the durable Stop marker.

The holder and Stop arbitrate on this hash:

```text
running
  Stop marks this token -> Finish chooses stopped
  Finish claims result -> later Stop cannot change it
finishing
  Runner has returned; settle terminal progress / reply delivery
  End stores the execution outcome for this token
ended
  a new holder may replace the current hash
```

An accepted Stop that won suppresses even a clean late Runner result. The
Runner receives context cancellation, including on Pi/Cursor turns through
their existing cancellation paths. Control renewal continues until the
Runner returns, so requesting cancellation alone never confirms a Stop.

The control Worker waits for the targeted token's `End` before posting
“Stopped”. Token outcomes remain available for 24 hours even if a newer
turn has begun. A missing holder without a settled outcome produces an
uncertainty message. A delivery error does not rewrite execution's outcome;
the durable reply publisher tracks confirmation delivery separately.

Entries accepted by older versions have no admission sequence or Stop target.
An old Prompt can run unless a newer accepted Stop discarded it. Replayed old
Stops ask the user to send Stop again; they cannot target a newer holder.

## Recovery and limits

The Worker uses 32 model slots and four control confirmation slots per Pod.
Each pool recovers its own pending entries and reclaims abandoned deliveries
after two minutes. An owner-aware heartbeat protects healthy long turns.
Failures before execution remain pending. Beginning execution records a
durable started marker, so reclaiming an event that may have run never invokes
the Agent again.

This control index contains no conversation or pending-message payloads;
sessions and the existing Stream remain their respective sources of truth.
Started markers and Stop cutoffs currently have no automatic cleanup.
Retention must preserve the protection of pending Stream entries.

Redis 7 or newer is required for the admission and terminal permission checks
(`redis.acl_check_cmd`). Redis must satisfy the existing durable queue
requirements and allow the additional control keys and commands.

The Stop control cutover requires all Workers to run this version. Older
Workers neither consume the control Stream nor honor the admission cutoff.
While new pending entries exist, rolling back only the Worker binary can
execute messages that a Stop logically discarded. Retain cutoff handling
until those entries have been drained or explicitly resolved.

The control index now keeps an accepted-order ZSET and a completed-order set
per Run. Admission registers each Prompt before it reaches a Worker; Begin
atomically checks every earlier uncompleted order, and End marks that order
complete and removes it from the pending-order ZSET. Older completed ZSET
entries remain readable: a completed head never hides an unfinished successor.
Stop removes the cutoff range, so old pending Prompts do
not block later accepted work. The memory implementation follows the same
contract for tests. The completed-order set and legacy index entries currently
have no automatic cleanup. Admission preflights the order-index command
permissions before changing the counter, cutoff or Stop marker. End validates
the order key types and permissions before publishing a settled outcome.

This slice does not implement #502's FIFO batching, per-batch invocation
association or actor provenance. A busy Prompt remains pending for recovery;
acquisition order is now FIFO once all events have reached durable admission.
Mapping the stopped control fact to the
Runner-owned Invocation now uses the same terminal arbitration: Stop first
persists CANCELLED with no late output; completion first preserves its result.
The runner calls the Plane finalizer before its terminal save, while the
holder is still renewed. A lost fence suppresses that save. A failed terminal
save leaves the run without an End outcome, so Stop cannot falsely confirm
that persistence settled. Such records can remain RUNNING until process-owner
cleanup; started markers still prevent executing them again.

Plane resolves the configured immutable Agent ID within the App's workspace
before starting the turn and carries that workspace into its Invocation.

Processing-record updates and their Dashboard projection remain separate
work, as does final protocol Gate B
acceptance. Neither #502 nor #503 is closed by this slice.

## Verification

`TestRunControlContract` runs the same fencing and Stop/completion cases against
memory and, with `REDIS_ADDR`, real Redis. The Redis fixture uses database 12.
`TestAdmissionStopCancelsAcrossPodsBeforeAnyControlWorkerReads` covers the
webhook acceptance boundary, cancellation convergence, unread old messages
and a new Prompt after Stop. Worker tests cover a full model pool, heartbeats,
pending recovery and a transient lookup failure. Admission tests reject
invalid Stream types and denied appends without leaving Stop side effects.

`TestRunClaimReturnsItsOwnAtomicFence` runs with `BUTTER_TEST_MONGO_URI` and
checks contention, increasing fences, late completion rejection and the
separation of lease documents from the processing-record list.

`TestRedisFIFOChecksEveryUnfinishedPredecessor` covers three admitted Prompts:
after the first finishes, the third must wait for an unread second, including
with completed index entries left by #519. `TestDeniedFIFOAdmissionCannotChangeControl`
checks denied order mutations leave the acceptance counter and Stop fact
untouched. `TestFIFOEndFailureDoesNotPublishSettlement` checks denied cleanup
and invalid order-key types leave the holder renewing without a settled outcome.
