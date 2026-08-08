# Workflow graphs are configured as explicit nodes + edges in proto

ADK Go v2 introduced a graph-based workflow agent (`agent/workflowagent`). Butter is
config-driven (proto `Agent` stored in YAML/Mongo), so the graph must be expressible in
configuration. We chose two explicit lists — `nodes` (name, kind, per-node options) and
`edges` (from/to node names plus an optional route label) — mapping 1:1 onto ADK's
`workflow.Edge`/`StringRoute` model, over a compact string DSL, because explicit lists
are trivially validatable, renderable by the dashboard, and need no parser.

Phase-1 node kinds are AGENT (references a sub-agent by name), HUMAN_INPUT (butter-owned
node that emits a workflow request-input event and relies on handoff resume — the
engine routes the human's reply to the node's successor), ROUTER (butter-owned node that stamps
`event.Routes` by matching its input text against outgoing edge labels — a trimmed,
case-insensitive exact match that stamps the label as configured, since the engine
compares route tags verbatim; ADK has no built-in way for config-driven graphs to
produce route tags), and JOIN. ADK's
FunctionNode and DynamicNode are deliberately excluded: they require arbitrary Go code and cannot be
expressed from configuration. ToolNode (referencing an MCP server tool) is deferred to
phase 2.

## Amendment: Agent V2 — ID-based AGENT node references

As of the Agent V2 migration (issue #212), AGENT nodes in workflow graphs may
reference an independent agent by its immutable `agent_id` via the new
`WorkflowNode.agent_id` field. This replaces the legacy `WorkflowNode.agent`
name-based reference for V2 agents.

During the migration observation period both fields are supported:

- **V2 agents** set `agent_id` on AGENT nodes. The runtime resolves the
  referenced agent from the workspace-wide agent pool.
- **Legacy agents** continue to use the `agent` (name) field referencing
  embedded sub-agents.

The `MigrateAgentsV2` RPC (APPLY mode) automatically converts name-based
references to `agent_id` references when expanding embedded sub-agents into
independent records. After the migration period, the name-based `agent` field
will be deprecated and removed in a future cleanup.
