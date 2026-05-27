# Agent Client Protocol Integration Analysis

Date: 2026-05-27

## Summary

接入 ACP 可行，但第一阶段更适合把 Butter 暴露成 **ACP agent server**，而不是先把 ACP agent 作为 Butter 的 `RemoteAgent` 消费。

原因：

- ACP 的主流稳定传输是 client 启动 agent 子进程，通过 stdin/stdout 交换 newline-delimited JSON-RPC。Butter 当前是常驻 HTTP/Twirp/gRPC 服务，天然不符合编辑器直接启动的进程模型，但可以通过一个很薄的 `cmd/butter-acp` stdio 适配器桥接到现有 Butter API。
- Butter 已有 session、runner、SSE 流、取消、workspace、MCP HTTP/SSE、invocation 记录等能力。ACP agent server 的核心工作主要是协议适配，不需要重写 agent runtime。
- 反向方向，即 Butter 作为 ACP client 去消费外部 ACP agent，需要在服务端长期托管子进程、实现 client-side filesystem/terminal/permission RPC，并解决多租户隔离和执行安全，风险明显更高。

建议路线：

1. Phase 1：新增 `cmd/butter-acp`，作为 stdio ACP agent wrapper，连接一个已有 Butter server。
2. Phase 2：补齐会话恢复、取消、MCP session 参数映射和更完整的 `session/update` 事件。
3. Phase 3：如确有编排需求，再把 `RemoteAgentProtocol` 扩展出 `ACP`，让 Butter 能作为 ACP client 消费外部 agent。

## ACP Facts Relevant To Butter

ACP 定义的是 agent 与 client 之间的 JSON-RPC 协议。典型 client 是代码编辑器，agent 通常作为 client 启动的子进程运行。

稳定基线：

- 初始化：client 调 `initialize`，双方协商 protocol version、capabilities、auth methods。
- 会话：client 调 `session/new` 创建会话，也可在 agent 支持时调 `session/load`、`session/resume`、`session/close`。
- 对话：client 调 `session/prompt`，agent 用 `session/update` notifications 流式上报文本、计划、tool call 等状态，最终返回 `stopReason`。
- 取消：client 发送 `session/cancel` notification，agent 应停止当前 turn 并让 pending prompt 以 `cancelled` stop reason 结束。
- 传输：当前稳定传输是 stdio。JSON-RPC 消息必须 UTF-8、单行 newline-delimited；agent stdout 只能输出 ACP 消息，日志走 stderr。
- 客户端能力：client 可暴露 `fs/read_text_file`、`fs/write_text_file`、`terminal/*`、`session/request_permission` 等反向 RPC 给 agent。
- MCP：`session/new` / `session/load` / `session/resume` 可携带 MCP server 配置。ACP 要求 agent 支持 MCP stdio，HTTP 和 SSE 通过 `mcpCapabilities` 宣告。HTTP 推荐，SSE 已被 MCP 侧废弃。

参考文档：

- https://agentclientprotocol.com/get-started/architecture.md
- https://agentclientprotocol.com/protocol/overview.md
- https://agentclientprotocol.com/protocol/transports.md
- https://agentclientprotocol.com/protocol/initialization.md
- https://agentclientprotocol.com/protocol/session-setup.md
- https://agentclientprotocol.com/protocol/prompt-turn.md
- https://agentclientprotocol.com/libraries/typescript.md

## Current Butter Fit

Butter 现有能力与 ACP 的映射：

| ACP concept | Butter today | Fit |
| --- | --- | --- |
| `initialize` | service metadata + runtime registry | Easy: static wrapper response plus optional agent list metadata |
| `session/new` | ADK session service, `runner.Service.Run` creates session when needed | Easy for create-on-first-prompt; explicit create is also possible through `SessionService.CreateSession` |
| `session/prompt` | `POST /api/chat/stream` and `runner.Service.RunSSE` | Good fit: SSE events can map to ACP `session/update` |
| `session/update` text chunks | `chatStream` `text_delta`, `agent_event`, `final` | Good first version for text; tool-specific richness needs more mapping |
| `session/cancel` | HTTP context cancellation and `CancelAgentInvocation` | Feasible; wrapper must keep session -> invocation/request cancel state |
| `session/load` / `resume` / `close` | `SessionService.GetSession`, persistent ADK sessions | Feasible but needs replay/active session tracking semantics |
| MCP HTTP/SSE | `MCPServerTransport_STREAMABLE_HTTP`, `SSE`, OAuth/static headers | Good fit |
| MCP stdio | intentionally removed/reserved in `agent.proto`; security review notes stdio MCP was removed to avoid arbitrary process execution | Deliberate gap; do not re-enable inside server by default |
| client fs/terminal callbacks | no ACP client-side RPC model in Butter server | Only needed when consuming external ACP agents or when exposing local editor capabilities to Butter tools |
| process model | Butter server is long-running HTTP/Twirp/gRPC; ACP expects stdio subprocess | Main impedance mismatch; solve with stdio wrapper |

Existing files most relevant to implementation:

- `internal/handler/http/chat_stream.go`: already turns runner events into SSE.
- `internal/runtime/runner/runner.go`: central run, streaming, cancellation, session creation, workspace scoping.
- `internal/application/session_service.go`: create/get/list/delete/reply for ADK sessions.
- `internal/application/agent_service.go`: `InvokeAgent` and `CancelAgentInvocation`.
- `proto/agents/v1/agent.proto`: remote agent and MCP transport config surfaces.
- `cmd/butter-daemon`: useful pattern for a secondary binary and config-driven server connection.

## Recommended Integration: Butter As ACP Agent

Add a small stdio executable:

```text
ACP Client (Zed / other editor)
  <stdio JSON-RPC>
cmd/butter-acp
  <HTTP/SSE + Bearer + X-Workspace-ID>
Butter server
  -> runner.Service
  -> ADK agents / MCP / daemon / A2A
```

This preserves Butter's server architecture and lets ACP clients run a local agent command without embedding the full Butter server into the editor process.

### Wrapper Configuration

`cmd/butter-acp` should accept flags/env:

- `--server-url` / `BUTTER_ACP_SERVER_URL`
- `--token` / `BUTTER_ACP_TOKEN`
- `--workspace-id` / `BUTTER_ACP_WORKSPACE_ID`
- `--agent` / `BUTTER_ACP_AGENT`
- `--user-id`, default `acp`
- optional `--app-name`, default `acp`
- optional `--model`
- optional `--title` for client-visible agent info

The wrapper should write logs only to stderr.

### Protocol Mapping

`initialize`:

- Return protocol version `1`.
- Advertise:
  - `promptCapabilities.image`: only if the wrapper can translate image content to Butter request parts. Initial version can set false.
  - `promptCapabilities.embeddedContext`: initially true for text resources by flattening into prompt text, or false if keeping MVP strictly text-only.
  - `mcpCapabilities.http`: true.
  - `mcpCapabilities.sse`: true only because Butter supports SSE today, while noting it is deprecated upstream.
  - `loadSession`: phase 2.
  - `sessionCapabilities.resume` / `close`: phase 2.
- `authMethods`: empty if the wrapper authenticates to Butter via config. ACP-side interactive auth can be postponed.

`session/new`:

- Generate an ACP session ID or use Butter session ID directly.
- Store session state in the wrapper:
  - ACP session ID
  - Butter session ID
  - cwd
  - selected agent
  - optional model override
  - MCP server config received from the ACP client
- Initial implementation can defer Butter session creation until first prompt because `runner.Service.RunSSE` already ensures a session exists. Phase 2 can call `SessionService.CreateSession` explicitly.

`session/prompt`:

- Convert ACP text content blocks into Butter `message`.
- For resource blocks, either include text inline with file URI headings or return unsupported until embedded context support is implemented.
- Call `/api/chat/stream` with:
  - `agent_name`
  - `message`
  - `app_name = acp`
  - `user_id`
  - mapped `session_id`
  - `model_override`
  - `Authorization: Bearer <token>`
  - `X-Workspace-ID: <workspace-id>`
- Translate SSE events:
  - `text_delta` -> `session/update` with `agent_message_chunk`
  - `agent_event` with non-text ADK content -> best-effort `tool_call` / `tool_call_update` once event mapping is defined
  - `error` -> JSON-RPC error or stop reason `refusal` depending on timing
  - `final` -> `session/prompt` response `{ "stopReason": "end_turn" }`

`session/cancel`:

- Keep a cancel function for each active prompt request.
- Cancel the in-flight HTTP/SSE request immediately.
- Optionally also call `AgentService.CancelAgentInvocation` using the invocation ID from `invocation_started`.
- Respond to the pending `session/prompt` with `{ "stopReason": "cancelled" }`, not a generic JSON-RPC error.

`session/load` / `session/resume`:

- Phase 2.
- Butter can map this to existing ADK sessions because sessions are already persisted by app/user/session.
- `session/load` requires replaying prior conversation as ACP `session/update`; this needs robust conversion from ADK session events to ACP content blocks.
- `session/resume` is easier: verify session exists and do not replay history.

`session/close`:

- Phase 2.
- Cancel active prompt for that session and release wrapper-local state. Deleting the persisted Butter session should be a separate explicit policy, not the default, because ACP close means free active resources rather than necessarily delete history.

### MCP Handling

Do not reintroduce server-side stdio MCP in Butter as part of ACP Phase 1. The current repo reserved/removed stdio MCP fields, and `docs/security-review.md` records that this avoided arbitrary process execution risk.

Recommended handling:

- Accept ACP HTTP/SSE MCP entries and map them to temporary per-session MCP config only after Butter has a runtime path for session-scoped MCP servers.
- Until then, document that Butter ACP wrapper uses MCP servers already configured in Butter's workspace/agent config.
- For ACP stdio MCP entries, either:
  - reject with a clear unsupported error, or
  - later implement a local wrapper-side proxy that converts stdio MCP to HTTP/SSE before forwarding to Butter. This keeps arbitrary subprocess execution in the local editor-owned wrapper process, not the multi-tenant Butter server.

## Alternative: Butter As ACP Client

This means adding `REMOTE_AGENT_PROTOCOL_ACP` and letting a Butter agent delegate to an external ACP agent.

It is feasible, but should not be Phase 1.

Required work:

- Add `REMOTE_AGENT_PROTOCOL_ACP` to `proto/agents/v1/agent.proto`, plus fields for command, args, env, cwd, startup timeout, and maybe registry reference.
- Implement an ACP client runtime that:
  - starts/manages subprocesses
  - speaks JSON-RPC bidirectionally over stdio
  - calls `initialize`, `session/new`, `session/prompt`
  - maps ACP stream updates into ADK events
  - implements cancellation and process cleanup
- Implement client-side ACP methods:
  - `session/request_permission`
  - optionally `fs/read_text_file`, `fs/write_text_file`
  - optionally `terminal/*`
- Define tenant isolation:
  - where subprocesses run
  - filesystem root constraints
  - env secret handling
  - per-workspace process limits
  - stdout/stderr log redaction
- Extend dashboard config and remote-agent status probing.

Security concerns are materially higher because the Butter server would run arbitrary agent subprocesses and possibly expose filesystem/terminal methods. If this is needed, align it with the existing `cmd/butter-daemon` execution-plane design rather than running arbitrary ACP agents directly inside the main server.

## Implementation Plan

### Phase 1: Minimal ACP Wrapper

Deliverables:

- Add `cmd/butter-acp`.
- Implement a small JSON-RPC stdio loop.
- Implement methods:
  - `initialize`
  - `session/new`
  - `session/prompt`
  - `session/cancel`
- Bridge prompt turns to `/api/chat/stream`.
- Support text-only prompts and text streaming.
- Add a README section with an ACP client command example.
- Add unit tests for JSON-RPC parsing, request dispatch, and SSE-to-ACP event mapping.

Acceptance criteria:

- An ACP client can launch `butter-acp` as a subprocess.
- A text prompt reaches an existing Butter agent in a workspace.
- Streaming text appears as ACP `session/update` chunks.
- Prompt completion returns `stopReason: end_turn`.
- Cancel returns `stopReason: cancelled`.
- Wrapper logs never pollute stdout.

### Phase 2: Session And Event Completeness

Deliverables:

- `session/load`, `session/resume`, `session/close`.
- ADK event -> ACP history replay.
- Better tool call mapping from ADK function calls / responses to ACP `tool_call` updates.
- Optional support for embedded text resources.
- Optional session-scoped MCP HTTP/SSE injection if the runner can accept per-run MCP config without mutating workspace agent config.

### Phase 3: Remote ACP Agents

Deliverables:

- `REMOTE_AGENT_PROTOCOL_ACP`.
- ACP client runtime package.
- Subprocess lifecycle manager.
- Security policy and workspace isolation.
- Dashboard fields/status probe.

This phase should probably reuse daemon infrastructure or require ACP agents to run behind a daemon, keeping execution close to the user's machine instead of the Butter control plane.

## Open Questions

- Which ACP clients are the target first: Zed only, or broader compatibility?
- Should `cmd/butter-acp` select a single configured Butter agent, or expose a small ACP slash command / mode selector for multiple Butter agents?
- Should ACP `cwd` be trusted and forwarded into Butter context, or only recorded as metadata?
- Should Butter support session-scoped MCP config, or require ACP users to configure MCP servers in Butter first?
- Is interactive ACP `authenticate` needed, or is a locally configured Butter token acceptable for the initial release?

## Recommendation

Proceed with `cmd/butter-acp` as the MVP. It provides real user value for ACP clients with the least architectural disruption, keeps risky local execution out of the multi-tenant server, and reuses the existing Butter runner/session stack. Avoid adding `REMOTE_AGENT_PROTOCOL_ACP` until there is a concrete need for Butter to orchestrate external ACP agents.
