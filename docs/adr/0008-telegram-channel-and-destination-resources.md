# ADR-0008: Telegram Channels are transports; Destinations are addresses

- Status: Accepted
- Date: 2026-08-13
- Issue: #264 (PRD), #265 (this decision)
- Supersedes for Telegram: the `AgentChannel` + `TelegramChannelConfig` model in
  `agentchannel.proto`, the raw `TelegramNotifyTarget`, and Cron's
  `channel_name + chat_id` delivery.

## Context

Telegram configuration was spread across three unrelated models. An interactive
`AgentChannel` held a Bot Token and a chat allowlist; a `NotifyGroup` target held
its own Bot Token, chat ID, and thread ID; a Cron delivery held a channel name
and a chat ID. The same Bot Token was therefore stored in as many places as it
was used, in plaintext, with no way to discover, validate, or rotate it
consistently.

The same conflation broke Forum Topics. Because a "channel" was both the
credential and the address, every topic in a group collapsed into one
chat-scoped session, and one Bot could not route different topics to different
Agents without duplicating its credential.

## Decision

Split the concern in two, and make the split the unit of persistence.

A **TelegramChannel** is one Bot transport. It owns the immutable Telegram Bot
identity, the encrypted write-only credential, the receive mode, and the
inbound/outbound lifecycle. It is not an address and carries no chat ID.

A **TelegramDestination** is one exact address under a Channel: a `chat_id` plus
an optional `message_thread_id`. It owns inbound Agent routing and interaction
policy, and it is the only thing Cron jobs and Notify Groups reference.

Four consequences follow, and each is enforced rather than documented:

1. **The Bot identity is pinned at creation.** `CreateTelegramChannel` resolves
   the Bot with `getMe` before anything is committed, stores the resulting Bot
   ID, and refuses any later change to it. Credential rotation validates the new
   token first and accepts it only when it resolves to the same Bot. Without
   this, replacing a token would silently repoint every Destination under the
   Channel at a different account, with nothing in the configuration changing.

2. **Bot IDs are globally unique; addresses are unique per Channel.** Both are
   unique indexes, not read-then-write checks, because these are exactly the
   invariants concurrent callers break: two Pods registering the same Bot, or two
   operators registering the same Topic. A read-then-write check passes in both
   callers and leaves one inbound update matching two Destinations.

3. **Addresses are immutable.** Cron jobs and Notify Groups persist a Destination
   ID. If the address behind that ID were editable, a metadata edit would
   silently redirect every downstream reference. Pointing somewhere else requires
   creating a new Destination, which makes the redirect visible in the
   references that have to be updated.

4. **An absent `message_thread_id` is an address, not a wildcard.** The group's
   non-topic conversation and each Forum Topic are separate Destinations with
   separate sessions and separate policy. Coercing "absent" to `0` would make the
   group general chat collide with a real thread ID.

Telegram-native identifiers are persisted as canonical decimal strings.
Supergroup chat IDs exceed 2^53, so a JSON number would lose precision in every
browser client; canonicalizing on write (`"007"` → `"7"`) is what makes the
uniqueness index trustworthy.

## Credential handling

The Bot Token is deliberately not a field of `TelegramChannel`, following
ADR-0005: callers pass ciphertext into `internal/repo/telegram` and get
ciphertext out, so no read path can surface a token into an API response or a
log line. `credential_state` reports only whether a usable credential exists.

Unlike repository binding PATs, Telegram credentials are encrypted under a
**database-backed master key** (`internal/repo/cryptokey` +
`secretbox.Keyring`) rather than a YAML secret, so credential setup needs no
deployment change in this release. Initialization is atomic across Pods: the
pointer document is inserted, and a duplicate-key error is read as "another Pod
won" rather than overwritten. Every ciphertext records the `key_id` that sealed
it.

This key lives in the same database as the ciphertext it protects, so it does
not defend against complete database compromise. That is an accepted,
transitional trade-off until Secret Manager/KMS support replaces it; the stored
`key_id` is what makes that migration possible.

## Alternatives considered

**Keep one resource with an optional thread ID.** Rejected: it reproduces the
current conflation. The credential's lifecycle (rotation, validation, receive
mode) and the address's lifecycle (which Agent, which users, which trigger)
change for different reasons and at different rates.

**A cross-platform Destination abstraction covering Discord.** Rejected as
premature. Telegram's Forum Topic semantics — an optional thread ID that is part
of the address on both receive and send — have no Discord equivalent, and
inventing the union before the second platform needs it would fix the wrong
shape.

**Mutable addresses with cascade updates to Cron and Notify Groups.** Rejected:
a cascade turns one edit into an invisible change of behavior in unrelated
resources. Blocking the edit and naming the referencing resources keeps the
operator in the loop.
