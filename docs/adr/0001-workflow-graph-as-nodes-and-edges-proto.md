# Workflow graphs are configured as explicit nodes + edges in proto

ADK Go v2 introduced a graph-based workflow agent (`agent/workflowagent`). Butter is
config-driven (proto `Agent` stored in the runtime configuration repository), so the
graph must be expressible in configuration. We chose two explicit lists — `nodes` (name, kind, per-node options) and
`edges` (from/to node names plus an optional route label) — mapping 1:1 onto ADK's
`workflow.Edge`/`StringRoute` model, over a compact string DSL, because explicit lists
are trivially validatable, renderable by the dashboard, and need no parser.

Phase-1 node kinds are AGENT (references an independent Agent by `agent_id`), HUMAN_INPUT (butter-owned
node that emits a workflow request-input event and relies on handoff resume — the
engine routes the human's reply to the node's successor), ROUTER (butter-owned node that stamps
`event.Routes` by matching its input text against outgoing edge labels — a trimmed,
case-insensitive exact match that stamps the label as configured, since the engine
compares route tags verbatim; ADK has no built-in way for config-driven graphs to
produce route tags), and JOIN. ADK's
FunctionNode and DynamicNode are deliberately excluded: they require arbitrary Go code and cannot be
expressed from configuration. ToolNode (referencing an MCP server tool) is deferred to
phase 2.

## Amendment: Agent ID references after cutover

After the Agent ID cutover (issue #241), every AGENT node must set
`WorkflowNode.agent_id`. The runtime resolves that ID from the workspace Agent
pool; it never resolves the deprecated `WorkflowNode.agent` name field.

Workflow child Agents are independent records referenced through IDs. Embedded
`sub_agents` remain readable only so historical records can be decoded, but they
are rejected when changed and are never consumed while constructing a runtime
Agent. The migration-era `MigrateAgentsV2` RPC is retired and returns
`Unimplemented`; `VerifyAgentIDCutover` is the read-only diagnostic surface.
