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
