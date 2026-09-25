# ADR-0013: Workspace and Agent Memory backed by a mem0 OSS server

- Status: Proposed
- Date: 2026-09-25
- Issue: #332 (PRD), #333–#338 (slices), #339 (follow-up)
- Builds on: ADR-0005 / ADR-0008 (credentials outside the public model), ADR-0010 (Agent ID as logical key)

## Context

We want agent sessions to carry memory across conversations, following the
hook model of mem0's official Claude Code / Codex / pi plugins: relevant memories
are recalled and injected before the model runs, and each finished exchange
is captured for extraction, without the model having to ask. We also want an
optional tool so the model can search or add memories itself.

The existing `internal/runtime/memory/mongo` implements ADK `memory.Service`
and is wired into every runner, but in practice it does nothing: nothing
calls `AddSessionToMemory` and no agent mounts a memory tool.

Four facts shape the design. All were verified against mem0 `main` on 2026-09-25:

- **The official plugins only work with the hosted platform.** They call the
  platform's `/v3/...` routes and poll its asynchronous events. The OSS REST
  server (`server/main.py`) has none of those routes, so we can reuse the
  plugins' hook behavior but not their protocol.
- **OSS `POST /memories` is synchronous.** With `infer=true` it returns only
  after one LLM extraction call plus embedding. The server has no queue or
  async mode.
- **OSS has only the identity fields `user_id` / `agent_id` / `run_id`,** plus
  flat `metadata`. It has no `app_id` field. Deduplication on add only
  compares against memories that match **all** of the identity fields given.
  v3 extraction is ADD-only, so memories accumulate.
- **Passing `agent_id` without `user_id` changes the extraction prompt** to an
  agent-learned framing. Passing both uses the plain user framing.

## Decision

### 1. mem0 OSS only, through a minimal hand-written client

butter speaks only the mem0 OSS REST API (`X-API-Key` auth, unversioned
paths), and only the `add` and `search` endpoints. It supports neither the
hosted platform nor the official plugins' protocol, and it never calls the
server's `/configure` or `/reset`. The mem0 operator chooses the extraction
LLM and the embedder.

### 2. One Workspace Memory Config per workspace; the API key never enters the proto

`WorkspaceMemoryConfig` holds the base URL, an `enabled` flag and a
`credential_state`. A workspace has zero or one, enforced by upsert keyed by
workspace, following the `WorkspaceRepoBinding` precedent. Members can read
it; owners, admins and global admins can manage it.

The API key is write-only. It is encrypted through a credential seam
(secretbox), following ADR-0005/0008, and is resolved and decrypted on every
call with no cache, so a rotation takes effect on the next turn.

On save, the service runs a probe search. 401 or 403 rejects the save; an
unreachable server only produces a warning. A separate test-connection RPC
repeats the probe.

Deleting or disabling the config does not check which agents reference it.
Memory-enabled agents keep running, just without memory.

### 3. Two scopes, and v1 shares Workspace Memory across the whole workspace

Workspace Memory has no per-person partition in v1. Every member, every
entry point (dashboard, AG-UI, API, OpenAI API, Telegram, A2A, cron,
automation, forum) and every memory-enabled agent reads from and writes to
one pool. Partitioning per person is deferred. Its prerequisite, a stable
identity for one human across channels, does not exist yet.

The workspace is encoded in the identity fields themselves, not only in
`metadata`. That way a forgotten filter cannot cross tenants.

| Scope | Written with | Recalled by filter |
|---|---|---|
| Workspace Memory | `user_id = "ws:<workspace_id>"` only | `{user_id: "ws:<workspace_id>"}` |
| Agent Memory | `agent_id = "ws:<workspace_id>:agent:<agent_id>"` only | `{agent_id: "ws:<workspace_id>:agent:<agent_id>"}` |

`run_id` is never set, and automatic capture never sets `agent_id`. Either
field would narrow mem0's add-time dedup, so the same fact would be stored
once per session or once per agent.

Provenance lives in `metadata`: `agent_id`, `session_id`, `invocation_id`,
`channel` and `principal`. A later per-person scope can then filter on it
without migrating data.

Agent Memory is written only explicitly, through the memory tool, and only
when the agent allows it. Automatic capture writes Workspace Memory only.

### 4. mem0 replaces the Mongo implementation of ADK `memory.Service`

A mem0-backed `memory.Service` replaces `internal/runtime/memory/mongo`.
There is still one global instance per runner service. ADK's
`SearchRequest` carries only `Query/UserID/AppName`, so the service derives
the workspace, agent and invocation from the context the runner already
populates. When the workspace has no enabled config, it does nothing.

Operations the ADK interface cannot express are extension methods in the
same package, not a second service. Examples are choosing a scope and
writing Agent Memory.

The Mongo code is removed. The `adk_memories` collection is left in place
and never dropped by butter. Nothing was ever written to it.

### 5. Memory Recall runs once per turn and is injected ephemerally

The invocation's **root agent** decides whether memory applies, through
`AgentConfig.memory`. At the start of a turn, a runner plugin searches both
scopes once, merges the results, drops duplicate memory IDs, keeps the top
`top_k` by score, and caches the result for the invocation.

The cached block is appended to the system instruction of **every LLM call
in the invocation**, including the LLM sub-agents of a composite or
Workflow root. It is never written into session history, so history stays
clean and ContextGuard summaries never absorb memories.

Because the injection is not persisted, recall must repeat on every turn.
This follows the pi plugin's cadence, not the Claude/Codex first-prompt-only
cadence.

- **Query:** the current user text. When it is shorter than 20 characters,
  the previous user message is appended, so replies like "ok, go on" still
  recall context. A turn with no text skips recall.
- **Injection:** a `<memories>` block with separate Workspace and Agent
  sections. Every entry carries its date. The block states that memories are
  historical, may be outdated, and yield to the current conversation.
  Because v3 is ADD-only, superseded facts remain in the pool.
- **Defaults:** `top_k` 5, `threshold` 0.3, a 2 s timeout, a 4000-character
  cap. `top_k` and `threshold` are overridable per agent.

ADK's `preloadmemorytool` is not used. It searches on every model call, and
it fails the model call when the search fails.

### 6. Memory Capture is best-effort, per turn, and scoped to the turn's invocation

After each turn, the runner's turn-completion seam starts a background
submission to mem0 with a 60 s timeout, detached from the request so that
extraction latency never adds to the reply. The submission contains only
the events of **that turn's invocation**, identified by the invocation ID
on the events. No offset is stored, and a Workflow resume is naturally its
own invocation.

- **What is sent:** user text and final assistant text only. Tool calls,
  tool results and thoughts are excluded; images become a placeholder.
- **Redaction:** common token and key patterns are removed by regex before
  sending.
- **Extraction:** `infer=true`.

A Pod crash loses that turn's capture, and we accept this. A durable Redis
queue was rejected as disproportionate for a best-effort enhancement.

### 7. Memory tools carry server-injected identity

`search_memory` and `add_memory` take a `scope` of `workspace` or `agent`.
`add_memory` uses `infer=true`, so explicit writes also benefit from mem0's
dedup. Identity is injected from the run context, and the model cannot name
a `user_id` or `agent_id`.

The tools are mounted only when the root agent is itself an LLM agent.
Composite roots do not have tools pushed into their sub-agents. Writing
Agent Memory also requires `allow_agent_scope_write`.

There are no update or delete tools. Deleting memories is for people, not
models.

### 8. Configuration and failure semantics

- **Agent config:** `AgentConfig.memory` (`MemoryConfig`, field 8) holds
  `enabled`, `disable_auto_recall`, `disable_auto_capture`, `enable_tools`,
  `allow_agent_scope_write`, and optional `top_k` / `threshold`. It is
  DB-only, not Agent Content.
- **PI and CURSOR agents** reject the field on write, as a box-owned
  behavior surface (ADR-0011/0012). As leaves inside an ADK tree they are
  simply unaffected.
- **Failures:** every mem0 failure degrades. Recall injects nothing, capture
  is dropped, and the turn proceeds.
- **Observability:** recall hits, latency and degradations go to logs and
  Langfuse spans, not to the invocation record.

## Considered options

- **Mounting mem0's MCP server.** Identity would be chosen by the model or
  by static config, so workspace isolation could not be enforced per session.
- **The hosted mem0 platform and the official plugins' protocol.** Out of
  scope. We run mem0 OSS.
- **Persisting injected memories into the user message or session events.**
  This pollutes history, gets memories compacted into ContextGuard
  summaries, and duplicates them turn after turn.
- **Per-person user scopes now,** such as `web:<uid>` and `tg:<tg_uid>`.
  butter has no cross-channel human identity, and several entry points use
  fixed placeholder user IDs. v1 shares at the workspace level, and
  provenance metadata keeps per-person scoping open.
- **Storing a capture offset in session state.** Every turn would append a
  state-delta event. Filtering by the turn's invocation ID is stateless.
- **Automatic capture into Agent Memory,** or tagging captures with both
  IDs. Tagging with both narrows dedup to one agent, and automatic capture
  into Agent Memory adds little over Workspace Memory in v1.

## Consequences

- **Memory is shared inside the workspace by design.** What one member tells
  one agent can surface to another member using another agent. This is
  documented, including the poisoning risk of a Telegram bot whose
  `allowed_user_ids` is empty.
- **Cron and automation runs capture too,** so their repeated prompts feed
  the pool.
- **The pool accumulates near-duplicates and superseded facts** (v3 is
  ADD-only). v1 has no management UI. Viewing and deleting memories is
  tracked as follow-up work.
- **The runner's turn-completion seam must expose the session and
  invocation identity** so that capture can select the turn's events.
