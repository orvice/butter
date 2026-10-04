# Butter

Multi-tenant agent orchestration service: proto/YAML-configured agents built on ADK Go, exposed through chat channels, RPC, and cron.

## Language

### Agent orchestration

**Agent**:
A workspace-scoped, independently managed unit of behaviour instantiated as an ADK agent. Typed as LLM, Loop, Sequential, Parallel, or Workflow, and identified by an immutable Agent ID.

**Agent ID**:
An immutable, workspace-unique slug that identifies an Agent in APIs, relationships, and repository paths. The display name may change without changing the Agent ID.
_Avoid_: agent name, agent UUID

**Sub-agent**:
An independent Agent referenced by another Agent as a child. It is not embedded in the parent configuration, and an Agent may have at most one parent relationship.
_Avoid_: nested agent, embedded agent

**Agent Content**:
The human-authored description and prompt text of an Agent. A repository-bound workspace treats the bound repository as the source of truth for this content.
_Avoid_: agent config

**Effective Agent**:
The runnable Agent produced by combining its operational configuration with the currently active Agent Content revision.

**Model Context Capacity**:
Optional operator-supplied metadata describing a provider Model's input context window in tokens. Zero or absent means the embedded model registry remains authoritative. It does not configure maximum output tokens or change the provider's actual limit.
_Avoid_: model max tokens, output limit

**Agent Context Override**:
The threshold ContextGuard value stored in `context_guard.max_tokens`. It overrides Model Context Capacity for that Agent's input-context calculations and is retained under the existing wire name for compatibility.
_Avoid_: maximum output tokens

**Effective Context Window**:
The input-context value ContextGuard uses after resolving the Agent Context Override, configured Model Context Capacity, embedded model metadata, and the 128,000-token unknown-model fallback in that order.
_Avoid_: provider hard limit

**Git Host**:
A platform-admin-configured Git service endpoint (GitHub, GitHub Enterprise, GitLab, or self-hosted GitLab) with a fixed API base URL. Workspaces bind repositories only on configured hosts; workspace input can never introduce an arbitrary URL.
_Avoid_: git provider config, git server

**Workspace Repository Binding**:
The association between a Workspace and one repository location, consisting of a Git host, repository, branch, and root path. The binding determines where that workspace reads and writes Agent Content. Each binding owns an independently encrypted PAT that is write-only through the API (ADR-0005).

**Observed Revision**:
The most recent repository revision known to Butter, whether or not its Agent Content passed validation.

**Active Revision**:
The last validated repository revision used to build Effective Agents. It remains active when a newer observed revision is invalid or temporarily unavailable.

**Remote Agent**:
An externally hosted agent (A2A, OpenCode HTTP, or Daemon protocol) referenced by ID from a shared registry and attached as a sub-agent.
_Avoid_: external agent

### Workflow graphs

**Workflow Agent**:
An agent whose behaviour is a directed graph of nodes and edges, executed by the ADK v2 workflow engine. Distinct from the legacy Loop/Sequential/Parallel agents.
_Avoid_: graph agent, DAG agent

**Node**:
A single step in a workflow agent's graph. Phase-1 kinds: Agent, Human Input, Router, Join. Tool nodes are planned for phase 2.

**Edge**:
A directed connection between two nodes, optionally guarded by a Route.

**Route**:
A string label on an edge; the edge is taken only when the emitting node's output carries a matching route value. Enables branching.
_Avoid_: condition, guard

**Human Input Node**:
A node that pauses the workflow, asks a human a question, and resumes the graph when the reply arrives.
_Avoid_: HITL node, approval node

**Router Node**:
A node that matches its input text against the route labels of its outgoing edges (trimmed, case-insensitive exact match) and stamps the winning label on the event, steering the branch taken.
_Avoid_: switch node, decision node

**Interrupt**:
The paused state of a workflow awaiting a human reply, identified by an Interrupt ID. Survives process restarts via session state.
_Avoid_: pause, suspension

**Parallel Worker**:
A node option that runs the node once per item of a list-typed input, concurrently, then aggregates outputs.

### Generative UI (A2UI)

**Surface**:
One A2UI-rendered area in AG-UI Chat, identified by a server-assigned surface ID: either a Result Card or a Human Input Form. It reaches a client as `butter.a2ui` CUSTOM events and from the UI Snapshot (ADR-0014).
_Avoid_: widget, generative component

**Result Card**:
A read-only Surface a model creates, updates, or removes with the `render_ui` tool during an A2UI-capable AG-UI run. Persisted in the session's hidden UI state namespace; it cannot collect input or trigger actions.
_Avoid_: rich message, UI message

**Human Input Form**:
The optional form presentation of a Human Input Node's question: ordered text and single-choice fields. Its binding to one Interrupt is frozen into the request-input event that opens it; submitting it answers exactly that Interrupt with a JSON object text, while other channels still answer in text.
_Avoid_: HITL form, typed resume

**UI Binding**:
The authenticated principal, Workspace, Agent ID, and thread an AG-UI session's Surfaces belong to, recorded when the session is created. UI is exposed and accepted only for a request whose context matches it.
_Avoid_: session owner

**UI Snapshot**:
A thread's current Result Cards and unanswered Human Input Forms, rebuilt from the persisted session without running the agent.
_Avoid_: UI cache, UI history

**Thread History**:
A thread's conversation rebuilt from the persisted session without running the agent. It holds the user turns and one reply per run, with that run's tool calls and results and the Human Input questions it asked. Each Surface of the UI Snapshot sits in the reply that produced it.
_Avoid_: transcript, chat log, message history

### Runs

**Detached Run**:
An AG-UI run whose client asked for it to outlive the request that started it. Disconnecting, reloading or leaving the page stops someone watching the run, not the run (ADR-0016).
_Avoid_: background run, async run, async invocation

**Run Log**:
The bounded, short-lived sequence of one Detached Run's AG-UI events, which every observer of the run replays from its start. It is a replay buffer; the session stays the record of the conversation.
_Avoid_: event buffer, run stream, watch hub

**Run State**:
What an AG-UI run in progress, detached or not, records about itself next to its thread lease: its runId, its Invocation ID, and how many events the thread's session held before it. It lapses with the lease, so a run whose process died stops reading as running within one lease TTL. Thread reads use it to tell that a run is going and where it started.
_Avoid_: run status, active run marker

**Stop**:
A person's explicit request to end a run in progress, such as the dashboard's Stop button or a Linear `stop` signal. It reaches the run wherever it executes, and the run is recorded as stopped, not failed; a dropped connection is never a Stop.
_Avoid_: cancel, abort

### Skills

**Skill**:
A workspace-level shared bundle of instructions and resources (SKILL.md plus optional references/assets/scripts), following the agentskills.io spec. Agents opt in by listing Skill names in their config; an empty list means no skill toolset is attached.
_Avoid_: plugin, capability

**Skill Name**:
The sole identifier of a Skill, unique per workspace and validated against the agentskills.io spec (1–64 chars, lowercase alphanumeric and hyphens). There is no separate generated ID; renaming a Skill is delete-and-recreate.
_Avoid_: skill ID, skill slug

**Skill Resource**:
A file attached to a Skill under one of the spec directories (`references/`, `assets/`, `scripts/`), addressed by its skill-root-relative path. Path metadata is indexed in Mongo; content lives in the ContentStore. Read at runtime via `load_skill_resource`. Limits: 10 MiB per resource (fixed, aligned with ADK's read cap), 100 per skill (configurable).
_Avoid_: skill file, attachment

### Multimodal input

**Input Part**:
One piece of multimodal user input on an agent-invoking RPC (`InputPart` in `agents/v1/content.proto`): either text or Inline Data. A request's `parts` list is ordered and may interleave text and images; when non-empty it is used as the user input and the legacy `message` field is ignored.
_Avoid_: attachment, content part

**Inline Data**:
Raw bytes plus their MIME type carried inside an Input Part. Limited to whitelisted image formats (jpeg/png/gif/webp), 10 MiB per image, 10 images and 20 MiB combined payload per request — enforced by the application layer, not the schema.
_Avoid_: blob, file upload

### Memory

**Workspace Memory**:
The memory pool shared by every member, entry point, and memory-enabled Agent of a Workspace, held by the workspace's mem0 OSS server. In v1 it has no per-person partition; it is the target of every Memory Capture (ADR-0013).
_Avoid_: user memory, user scope

**Agent Memory**:
Memory private to one Agent within a Workspace, written only explicitly through the Agent's memory tool and recalled alongside Workspace Memory.
_Avoid_: agent knowledge

**Workspace Memory Config**:
The zero-or-one per-Workspace connection to a mem0 OSS server: base URL, enabled flag, and a write-only encrypted API key. Without an enabled one, memory-enabled Agents run without memory.
_Avoid_: memory provider, mem0 resource

**Memory Recall**:
Retrieving Workspace and Agent Memory relevant to the current turn and presenting it to the model for that turn only; recalled memories are never written into session history.
_Avoid_: memory preload, memory injection

**Memory Capture**:
Best-effort submission of one turn's user and assistant text to Workspace Memory for extraction after the turn completes.
_Avoid_: memory sync, memory flush

### Linear

**Linear App**:
A Workspace's registration of one Linear OAuth application, routed to exactly one Agent. Its app user is that Agent's identity in Linear (ADR-0015).
_Avoid_: Linear integration, Linear bot, Linear channel

**Linear Installation**:
A Linear App installed into one Linear organization, holding that organization's app user and access grant.
_Avoid_: Linear workspace (collides with Workspace), Linear connection

**Linear Agent Session**:
Linear's conversation on an issue between people and a Linear App's app user, opened by delegation or mention. In Butter it is exactly one session of the routed Agent.
_Avoid_: Linear thread, Linear conversation
