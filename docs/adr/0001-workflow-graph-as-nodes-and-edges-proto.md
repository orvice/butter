# Workflow graphs are configured as explicit nodes + edges in proto

ADK Go v2 introduced a graph-based workflow agent (`agent/workflowagent`). Butter is
config-driven (proto `Agent` stored in YAML/Mongo), so the graph must be expressible in
configuration. We chose two explicit lists — `nodes` (name, kind, per-node options) and
`edges` (from/to node names plus an optional route label) — mapping 1:1 onto ADK's
`workflow.Edge`/`StringRoute` model, over a compact string DSL, because explicit lists
are trivially validatable, renderable by the dashboard, and need no parser.

Phase-1 node kinds are AGENT (references a sub-agent by name), HUMAN_INPUT (butter-owned
node that calls `workflow.ResumeOrRequestInput`), ROUTER (butter-owned node that stamps
`event.Routes` by exact-matching its input text against outgoing edge labels — ADK has no
built-in way for config-driven graphs to produce route tags), and JOIN. ADK's
FunctionNode and DynamicNode are deliberately excluded: they require arbitrary Go code and cannot be
expressed from configuration. ToolNode (referencing an MCP server tool) is deferred to
phase 2.
