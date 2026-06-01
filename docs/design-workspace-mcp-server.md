# Workspace MCP Server 设计方案

## 背景

Butter 已经支持把外部 MCP server 配置到 agent 上：`MCPServer` 配置按 workspace 隔离，存储在 `ConfigStore`/Mongo 中，并由 `internal/agent` 在构建 ADK agent 时解析为 MCP client。当前缺少的是反方向能力：让每个 Butter workspace 自身暴露为一个 MCP server，使外部 MCP client 可以通过标准 MCP 协议访问该 workspace 内的工具、资源和数据。

后端建议直接基于 `github.com/modelcontextprotocol/go-sdk/mcp`。仓库当前已依赖 `github.com/modelcontextprotocol/go-sdk v1.4.1`，该版本提供 `mcp.NewServer`、`mcp.AddTool`、`mcp.NewStreamableHTTPHandler` 和 `mcp.NewSSEHandler`，足够实现第一版 workspace MCP endpoint。

## 目标

- 为每个 workspace 暴露标准 MCP endpoint。
- 复用现有 `Authorization: Bearer ...` 和 workspace 鉴权模型。
- 复用现有 repo/service 层，不绕过业务校验和 workspace 边界。
- 首版优先支持 Streamable HTTP，保留 SSE 兼容入口。
- 工具以明确 allowlist 暴露，避免把管理能力一次性全量开放给 MCP client。
- 让 agent 也能通过现有 `MCPServer` 配置接入本 workspace 的 MCP endpoint，形成统一工具接入模型。

## 非目标

- 不在第一版实现完整 MCP OAuth authorization server。现有 API token/session bearer token 先作为 MCP endpoint 的认证方式。
- 不把 Twirp API 自动反射为 MCP tools。所有工具需要手写适配器和权限标注。
- 不在第一版实现跨 workspace 聚合查询。每个 MCP session 只绑定一个 workspace。
- 不改变现有外部 MCP client 配置、OAuth、global preset 的行为。

## Endpoint 设计

建议新增独立 HTTP handler，而不是 Twirp service：

| 路径 | 协议 | 说明 |
|------|------|------|
| `POST /api/workspaces/:workspace_id/mcp` | MCP Streamable HTTP | 首选 endpoint |
| `GET /api/workspaces/:workspace_id/mcp` | MCP Streamable HTTP/SSE stream | 由 SDK handler 处理 |
| `DELETE /api/workspaces/:workspace_id/mcp` | MCP session close | 由 SDK handler 处理 |
| `/api/workspaces/:workspace_id/mcp/sse` | MCP SSE legacy | 可选兼容入口 |

路由继续挂在 Gin 上，但具体协议处理交给 SDK：

```go
handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
    return workspaceMCP.ServerForRequest(r)
}, nil)
r.Any("/api/workspaces/:workspace_id/mcp", gin.WrapH(handler))
```

`ServerForRequest` 不应返回全局共享可变 server。建议按 workspace 和权限 profile 缓存 `*mcp.Server`，工具 handler 内再从 `context.Context` 读取当前调用者、workspace、request id 等信息。若后续需要按用户过滤 tool list，可把 cache key 扩展为 `workspace_id + principal_role + toolset_version`。

## 认证和 Workspace 解析

现有 `AuthMiddleware` 已支持三类 token：

- 登录 session token，配合 `X-Workspace-ID` 做成员校验。
- root API token，按 admin 处理。
- workspace-scoped API token，直接把 workspace id 写入 context。

MCP endpoint 建议支持两种 workspace 选择方式：

1. 路径参数 `:workspace_id` 作为显式目标 workspace。
2. API token 自带 workspace 时，必须和路径 workspace 一致；不一致返回 403。

实现上新增 `internal/handler/http/workspace_mcp.go`，在 Gin handler 中做路径 workspace 校验，然后把 request context 改写为 `workspace.WithID(ctx, workspaceID)` 再交给 SDK handler。不要只依赖 `X-Workspace-ID`，否则 MCP client 配置会更脆弱，也更难区分多个 workspace endpoint。

权限建议第一版采用 workspace member 可读、owner/admin 可写：

| Tool 类型 | member | owner | admin |
|----------|--------|-------|-------|
| list/get/read/search | 允许 | 允许 | 允许 |
| create/update/delete/run side-effect | 默认拒绝 | 允许 | 允许 |
| cross-workspace/admin/global preset | 拒绝 | 拒绝 | 允许 |

API token 目前没有细粒度 scope。若要对外开放写工具，应先给 `APIToken` 增加 scopes，例如 `mcp:read`、`mcp:write`、`agents:run`、`files:write`，否则首版建议只暴露只读工具和一个明确 opt-in 的 agent run 工具。

## 模块拆分

建议新增以下包：

```text
internal/mcpserver/
  service.go          # Workspace MCP server factory/cache
  auth.go             # principal extraction and permission helpers
  tools.go            # tool registry and registration
  tools_agent.go      # agent/session related tools
  tools_files.go      # agent file tools
  tools_config.go     # read-only config tools
  resources.go        # MCP resources/templates
  prompts.go          # optional workspace prompts

internal/handler/http/
  workspace_mcp.go    # Gin route adapter around mcp HTTP handlers
```

`internal/app/routes.go` 负责实例化 `mcpserver.Service` 并注入依赖：

- `configStore`，用于 workspace-scoped agent、MCP server、remote agent、channel、model provider 配置读取。
- `AgentFileRepo`，用于文件空间和文件读取/写入。
- `RunnerSvc`，用于可选的 agent run tool。
- `WorkspaceRepo`/auth repo，若需要在 handler 里做成员和角色校验。

注意 `SetupRoutes` 早于 bootstrap 完成，依赖注入要遵循现有 `Handlers.Wire` 模式：先创建 service，bootstrap 后再设置 repo/runner。

## 首版 Tool 清单

优先选择对模型有价值、风险可控、能稳定映射现有服务的工具。

### 只读工具

| Tool | 输入 | 输出 | 依赖 |
|------|------|------|------|
| `workspace_info` | `{}` | workspace id/name/slug | workspace repo |
| `list_agents` | `{label_filter?, limit?}` | agents summary | config repo |
| `get_agent` | `{name}` | agent config redacted | config repo |
| `list_mcp_servers` | `{}` | MCP configs summary redacted | config repo |
| `list_file_spaces` | `{}` | file spaces | agentfile repo |
| `list_files` | `{space_id, path_prefix?}` | file metadata | agentfile repo |
| `read_file` | `{space_id, path, version?}` | content + metadata | agentfile repo |
| `search_files` | `{space_id, query, limit?}` | matches | agentfile repo |

### 可选写入/执行工具

| Tool | 输入 | 权限 | 说明 |
|------|------|------|------|
| `write_file` | `{space_id, path, content, content_type?, metadata?}` | owner/admin or `files:write` | 复用 `AgentFileService` 限制，必须受 size limit 约束 |
| `run_agent` | `{agent_name, prompt, session_id?, model?}` | owner/admin or `agents:run` | 调用 `runner.Service.Run`，设置 timeout 和 max output |
| `create_session_reply` | `{agent_name, session_id, message}` | owner/admin or `agents:run` | 若要保留聊天语义，可封装现有 SessionService |

Tool 命名建议使用稳定 snake_case。不要把 workspace id 放进 tool input，workspace 必须来自 MCP endpoint context。

## Resources 和 Prompts

Resources 可以作为只读补充，便于 MCP client 浏览上下文：

| URI | 内容 |
|-----|------|
| `butter://workspace` | 当前 workspace 摘要 |
| `butter://agents` | agent 列表 |
| `butter://agents/{name}` | 单个 agent 配置摘要 |
| `butter://files/{space_id}/{path}` | agent file 内容 |

Prompts 首版可选。较有价值的是：

- `run_workspace_agent`：生成调用 `run_agent` 的标准提示。
- `summarize_workspace`：汇总当前 workspace 的 agents、channels、files。

## 数据模型

第一版可以不新增独立 workspace MCP server 表，直接通过 workspace、API token、route 暴露能力。

若需要前端可配置开关和 tool allowlist，建议新增 `WorkspaceMCPSettings`：

```protobuf
message WorkspaceMCPSettings {
  string workspace_id = 1;
  bool enabled = 2;
  repeated string enabled_tools = 3;
  bool enable_sse_legacy = 4;
  int32 max_request_seconds = 5;
  int32 max_response_bytes = 6;
  google.protobuf.Timestamp updated_at = 20;
}
```

存储可以放在 config Mongo collection，例如 `config_workspace_mcp_settings`，key 为 `workspace_id`。但如果产品暂时默认启用只读 MCP，则该表可以推迟到第二阶段。

## 错误、超时和限流

- 所有 tool handler 必须设置 request timeout。默认 30s；`run_agent` 可单独配置 120s 或走异步任务。
- MCP tool error 应返回结构化错误文本，不泄露 secret、token、内部 DSN。
- 输出需要统一截断，默认建议 64 KiB；文件读取可支持 `offset`/`limit` 后续分页。
- 对 `run_agent`、`write_file` 等高成本工具加 workspace+principal 维度限流。
- server 日志记录 `workspace_id`、principal、tool、duration、status，不记录 tool input 中的完整文件内容或 prompt。

## 与现有 MCP Client 配置的关系

Butter 当前的 `MCPServer` proto 表示“agent 连接外部 MCP server”。新增 workspace MCP endpoint 后，可以自动生成一个可安装的内部 preset：

```yaml
name: butter-workspace
transport: MCP_SERVER_TRANSPORT_STREAMABLE_HTTP
url: https://<butter-host>/api/workspaces/<workspace_id>/mcp
auth:
  type: MCP_SERVER_AUTH_TYPE_STATIC_HEADERS
headers:
  Authorization: Bearer <workspace-api-token>
```

这能让 Butter agent 复用现有 `mcp_server_ids` 接入本 workspace 工具。注意不要把当前用户 session token 写入持久化配置；用于 agent 的 internal MCP server 应创建 workspace-scoped API token 或服务 token，并按最小 scope 授权。

## 实施阶段

### Phase 1: 只读 Workspace MCP Endpoint

1. 新增 `internal/mcpserver.Service`，封装 SDK server 创建、tool 注册、依赖注入。
2. 新增 Gin route `/api/workspaces/:workspace_id/mcp`，复用 `AuthMiddleware` 并强制路径 workspace 和 context workspace 一致。
3. 注册只读 tools：workspace info、list/get agents、list MCP servers、list/read/search files。
4. 增加 redaction：MCP server headers、OAuth client secret、agent metadata 中未来可能的 secret key。
5. 增加 handler/service 单元测试和 streamable HTTP integration test。
6. 更新 `docs/api.md`，给出 MCP client 配置示例。

### Phase 2: Scoped Tokens 和写工具

1. 扩展 `APIToken` proto/repo/service，加入 scopes。
2. MCP handler 把 token scopes 写入 context principal。
3. 开启 `write_file` 和 `run_agent`，并增加 timeout、输出截断、限流。
4. 前端增加 workspace MCP token 创建和复制 endpoint 配置。

### Phase 3: 配置化和高级 MCP 能力

1. 增加 `WorkspaceMCPSettings`，支持 enabled_tools、legacy SSE 开关、limits。
2. 增加 resources/resource templates。
3. 增加 prompts。
4. 支持 tool list change notification：settings 或 config 变更后通知活跃 MCP session。
5. 评估 MCP OAuth provider 支持，让第三方 MCP client 走标准授权流程。

## 测试计划

- `internal/mcpserver` tool handler 单元测试：
  - workspace id 缺失返回错误。
  - 只返回当前 workspace 数据。
  - secret 字段被 redacted。
  - 文件路径和 size limit 生效。
- HTTP integration test：
  - 无 token 访问 MCP endpoint 返回 401。
  - session token + 非成员 workspace 返回 403 或 tool 不可用。
  - API token workspace 与路径 workspace 不一致返回 403。
  - `initialize`、`tools/list`、`tools/call` 走通。
- Regression test：
  - 现有 `MCPServerService` CRUD、OAuth callback、global MCP preset install 不受影响。
  - 现有 Twirp route prefix `/api/agents.v1.*` 不被 MCP route 捕获。

## 风险和开放问题

- `mcp.NewStreamableHTTPHandler` 会管理 MCP session，需确认生产环境下长连接和 abandoned session 的资源回收策略；必要时配置 SDK keepalive，并在反向代理层设置合理 idle timeout。
- 现有 API token 没有 scopes，直接开放写工具风险过高；建议 Phase 1 只读。
- `run_agent` 是长耗时工具，MCP request 同步返回可能不适合所有 client；若实际任务超过 2 分钟，应改成 `start_agent_run` + `get_agent_run` 的异步工具形态。
- 如果未来每个用户看到的 tool list 不同，server cache key 不能只按 workspace 缓存。
