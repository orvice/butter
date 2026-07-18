# Butter

Multi-tenant agent orchestration service: proto/YAML-configured agents built on ADK Go, exposed through chat channels, RPC, and cron.

## Language

### Agent orchestration

**Agent**:
A configured unit of behaviour instantiated as an ADK agent. Typed as LLM, Loop, Sequential, Parallel, or Workflow.

**Sub-agent**:
An agent nested under a parent agent in the config tree; built recursively and available for transfer or as workflow node targets.

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

### Multimodal input

**Input Part**:
One piece of multimodal user input on an agent-invoking RPC (`InputPart` in `agents/v1/content.proto`): either text or Inline Data. A request's `parts` list is ordered and may interleave text and images; when non-empty it is used as the user input and the legacy `message` field is ignored.
_Avoid_: attachment, content part

**Inline Data**:
Raw bytes plus their MIME type carried inside an Input Part. Limited to whitelisted image formats (jpeg/png/gif/webp), 10 MiB per image, 10 images and 20 MiB combined payload per request — enforced by the application layer, not the schema.
_Avoid_: blob, file upload
