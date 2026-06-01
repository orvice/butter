# Butter 功能总览

更新时间：2026-06-02

Butter 是基于 Butterfly 框架的 Agent 服务，核心使命是把多种入口（HTTP / Twirp / gRPC / 即时消息 / 定时任务）统一编排为 Google ADK Agent 执行流，并提供配置化、热更新、多执行面、多租户与持久化运行时。

## 0. Workspace 多租户

- **Workspace 实体**：所有 Agent / Channel / MCP Server / Remote Agent / Model Provider / Cron Job / API Token / Invocation / Cron Execution 必须归属一个 workspace。`WorkspaceService` 暴露 workspace 与成员（`owner` / `admin` / `member`）的 CRUD。
- **请求级别 workspace 选择**：客户端通过 `X-Workspace-ID` HTTP 头选择活跃 workspace；`AuthMiddleware` 校验该用户是 workspace 成员（全局 `admin` 角色旁路）。CRUD RPC 缺失 header 时返回 `failed_precondition`。
- **登录返回 workspace 列表**：`AuthService.Login` 在响应中带 `workspaces`，dashboard 登录后弹 workspace 选择器，把选中的 workspace 写入后续请求头。
- **API token 自带 workspace**：`CreateAPIToken` 时记录创建上下文的 workspace；用 API token 鉴权时直接绑定到 token 的 workspace，忽略 header。
- **默认 workspace 自举**：进程启动时若 `workspaces` 集合为空，会自动创建 slug 为 `default` 的 workspace，并把现有所有 user 加为 `owner`。
- **运行时行为**：runner / channel manager / cron scheduler 跨 workspace 拉平所有配置作为全局视图运行（agent 名字目前需全局唯一）；`ContextInfo.workspace_id` 把所属 workspace 沿调用链透传，写 invocation / cron execution 时回填。
- **System agent**：仍为全局注册，其 `list_agents` / `list_cron_jobs` 等读工具跨 workspace，写工具（`create/update/delete_cron_job` 等）要求显式传入 `workspace_id` 参数。

## 1. Agent 编排

- **多类型 Agent 构建**：通过 `agents.v1.Agent` 配置统一生成 ADK Agent，支持四种类型：
  - `AGENT_TYPE_LLM`：LLM Agent，支持 instruction、global instruction、input/output JSON schema、`output_key`、`context_guard`、`include_contents` 等参数。
  - `AGENT_TYPE_LOOP`：Loop workflow，支持 `max_iterations`。
  - `AGENT_TYPE_SEQUENTIAL`：顺序 workflow。
  - `AGENT_TYPE_PARALLEL`：并行 workflow。
- **子 Agent 与委派**：支持嵌套 `sub_agents` 树，结合 `description` 用于 LLM 子 Agent 委派。
- **Labels / Metadata**：每个 Agent 可携带 `labels`、`metadata`，用于路由与索引。
- **内置系统 Agent**：进程启动时注册 built-in system agent，便于诊断和管理类操作。
- **Agent 文件挂载**：Agent 可挂载 workspace 级文件空间，通过内置 `agent_files` 工具集读写（`AgentFileService` + `internal/agentfiletool`）。

## 2. 模型管理

- **模型别名与 Provider 解析**：通过 `model_providers` 配置把别名（如 `flash`）映射到具体模型。
- **运行时 Model Override**：渠道/调用方可在调用时指定 `model_override`；Runner 会 clone 配置、替换模型并缓存 override 后的 Agent 实例。
- **Langfuse Tracing**：运行时初始化 Langfuse plugin 支持模型调用追踪。

## 3. MCP 工具集

- **共享 MCP Server 配置**：通过 `MCPServerService` CRUD 管理 workspace 级 MCP Server。
- **Global MCP Presets**：admin 维护跨 workspace 的 preset 模板；用户通过 REST install 复制到当前 workspace。
- **MCP OAuth2**：支持 OAuth 保护的 MCP server；`StartMCPServerOAuth` 发起浏览器授权，回调走 `GET /api/mcp/oauth/callback`。
- **多种 Transport**：
  - `MCP_SERVER_TRANSPORT_STREAMABLE_HTTP`
  - `MCP_SERVER_TRANSPORT_SSE`
- **工具白名单**：`tool_filter` 控制对外暴露的 MCP 工具。
- **Agent 关联方式**：Agent 可通过 `mcp_servers`（内联）或 `mcp_server_ids`（引用共享配置）挂载 MCP 工具集。

## 4. 远程 Agent

- **A2A 协议**：`REMOTE_AGENT_PROTOCOL_A2A`，通过 `url` 调用远程 ADK Agent。
- **Daemon 协议**：`REMOTE_AGENT_PROTOCOL_DAEMON`，通过 `daemon_capability` 查找在线 daemon 连接，由 `daemon.Bridge` 把 ADK invocation 转成 `DaemonTask`，等 daemon 回传 terminal update 后生成 ADK final event。
- **断连/取消**：连接断开时活跃任务会收到失败更新；context 取消时下发 `CancelTask`。

## 5. Daemon 执行面

- **反向连接架构**：daemon client（`cmd/butter-daemon`）主动连接服务端 gRPC `DaemonConnectorService.Connect`，注册 `DaemonInfo` 与能力。
- **任务下发与执行**：服务端通过 `daemon.GRPCHandler` / `Registry` / `Connection` / `Bridge` 把任务下发到 daemon，daemon 通过本地 shell/opencode executor 执行后回传 `DaemonTaskUpdate`。
- **默认端口**：daemon gRPC 服务默认监听 `:9090`。

## 6. 多入口接入

### HTTP（Gin）

- `GET /ping`：健康检查，免鉴权。
- `GET /status`：配置存储后端与集合计数。
- `GET /a2a/:agent_name/.well-known/agent.json`：A2A agent card（仅 `enable_a2a: true` 的 Agent）。
- `POST /a2a/:agent_name`：A2A JSON-RPC `tasks/send`。
- `POST /api/chat/stream`：SSE 流式聊天（dashboard Chat 页）。
- `POST /api/uploads/*`：头像与静态资源上传（需 S3 配置，见 `docs/storage.md`）。
- `GET /api/global-mcp-servers` + `POST /api/global-mcp-servers/:id/install`：Global MCP preset 列表与安装。
- `POST/PUT/DELETE /api/admin/global-mcp-servers/*`：Global MCP preset 管理（admin）。
- `GET /api/mcp/oauth/callback`：MCP OAuth 浏览器回调（免 Bearer）。

### Twirp RPC（`/api`）

配置类：

- `AgentService`：Agent 配置 CRUD（含分页）+ `InvokeAgent` / `CancelAgentInvocation` / `ReloadAgents` / runtime status / invocations。
- `AgentFileService`：Agent 文件空间 CRUD + 读写/搜索（供 `agent_files` 工具集挂载）。
- `MCPServerService`：MCP Server CRUD + live 探活 + `ListMCPTools` + MCP OAuth 连接管理。
- `ModelProviderService`：模型 Provider CRUD。
- `NotifyGroupService`：Cron 通知组 CRUD（Telegram / Lark / Discord webhook 目标）。
- `RemoteAgentService`：远程 Agent CRUD + `GetRemoteAgentStatus`。
- `ChannelService`：渠道 CRUD + 状态与控制（Restart / Pause / Resume）。
- `CronJobService`：定时任务 CRUD + 执行历史 + `RunCronJobNow`。
- `SessionService`：Session CRUD + `Reply`（含 trace_url）。
- `ForumService`：讨论帖 CRUD + 帖内 Agent 调用。

运维 / 身份类：

- `AuthService`：密码登录 + OAuth 登录 + 用户/Profile 管理。
- `WorkspaceService`：Workspace 与成员 CRUD。
- `DashboardService`：Overview / ActivityFeed / Cron 时序。
- `DaemonService`：Daemon 注册表与 bridge 诊断。
- `APITokenService`：多 token 管理。

除公开路径（`/ping`、Auth 登录/OAuth RPC、MCP OAuth callback）外，HTTP/Twirp 统一经过 `AuthMiddleware`（`internal/authn.Resolver`）：先查 Redis session，再 root token，再 `api_tokens`；通过后解析 `X-Workspace-ID`。

### gRPC

- `DaemonConnectorService.Connect`：daemon 双向流接入。

### 即时消息渠道

- **Telegram**：长轮询 poller，支持文本/图片，按 USER/CHAT scope 派生 session id。
- **Discord**：长轮询 poller。
- **触发与白名单**：Channel 配置中可定义 trigger 与 allowlist。
- **回复能力**：reply、status、debug、clear 多种响应类型。
- **活跃选择存 Redis**：渠道内用户/会话维度的活跃 agent 与 model 选择走 Redis。

### Cron 调度

- 后台 scheduler 按 cron 表达式触发 Agent 执行。
- 支持标准 5 字段表达式和 `@every` / `@daily` / `@hourly` / `@weekly` / `@monthly` 等预定义 schedule。
- 时区可配（默认 UTC）。

## 7. 会话与记忆

- **ADK Session**：MongoDB session service 持久化会话事件，支持按 channel/user/session 查询、列表、删除、回复。
- **ADK Memory**：MongoDB memory service 保存长期记忆。
- **ContextInfo**：runner 调用统一携带 channel、session、user、source、uuid，作为执行上下文。
- **会话维度的 Agent Runner 缓存**：按 `channel:agent:model` 维度缓存 ADK runner 实例。

## 7.1 Forum 讨论

- **Workspace 级讨论帖**：`ForumService` 管理 thread/post；用户发帖或 `@agent` 触发 `InvokeAgentInThread`。
- **持久化**：MongoDB `forum_threads` / `forum_posts`；前端 Forum 列表与 Thread 详情页。

## 7.2 Dashboard Chat

- **SSE 流式对话**：`POST /api/chat/stream` 供 dashboard Chat 页实时展示 agent 输出。
- **Session 复用**：与 ADK session 体系共用 runner，支持 agent/model 选择。

## 8. Cron 自动执行

- **CronJob 配置**：name、schedule、agent_name、input、timezone、enabled、delivery、metadata。
- **结果投递（CronDelivery）**：
  - `CRON_DELIVERY_TYPE_LOG`：写日志。
  - `CRON_DELIVERY_TYPE_WEBHOOK`：HTTP webhook 推送。
  - `CRON_DELIVERY_TYPE_CHANNEL`：转发到指定 `AgentChannel` 的 `chat_id`。
  - `CRON_DELIVERY_TYPE_NOTIFY_GROUP`：按名称加载 `NotifyGroup`，向其中启用的 Telegram、Lark webhook、Discord webhook 目标发送通知。
- **执行记录（CronExecution）**：每次执行写入 MongoDB，包含输入、输出、状态、起止时间，支持分页查询。

## 9. 配置与热更新

- **两层配置**：
  - `AppConfig`：Butterfly 从 `BUTTERFLY_CONFIG_FILE_PATH` 指向的 YAML 加载进程级启动配置（tracing、storage、auth、static 等）；仓库示例为根目录 `config.yaml`。
  - `ConfigStore`：运行时配置仓库，支持 `memory` 或 `mongo` 后端。
- **加载行为（非 seed）**：`ConfigStore.InitFromConfig` 选择后端后，通过 `SyncToConfig` **从持久化仓库读取**各 workspace 的配置并 flatten 到 `AppConfig` 内存视图。YAML 中的 `agents` / `mcp_servers` / `remote_agents` / `channels` / `model_providers` **不再作为 seed 源**；应通过 Twirp RPC（带 `X-Workspace-ID`）写入。`memory` 后端初始为空，重启后丢失。
- **运行时热更新**：
  - Agent / MCP / RemoteAgent / ModelProvider 变更触发 `ReloadRunner`，跨 workspace 拉平配置后重建 agent registry，清空 runner 与 model override 缓存，并 reload channels。
  - Channel 变更触发 `ReloadChannels`。
- **持久化对象**：agents、mcp_servers、remote_agents、channels、model_providers、notify_groups、cron_jobs、cron_executions、global MCP presets。

## 10. 持久化

- **MongoDB**（默认数据库 `butter`，可通过 `mongo_db` 配置）：
  - `adk_sessions` / `adk_events`：ADK session 与事件
  - ADK memories
  - `workspaces`：workspace 元数据，`slug` 唯一索引
  - `workspace_members`：用户—workspace 多对多关系，`(workspace_id, user_id)` 唯一索引 + `user_id` 普通索引
  - `users`：dashboard 用户
  - 运行时配置：`config_agents` / `config_mcpservers` / `config_remoteagents` / `config_channels` / `config_modelproviders` / `config_notifygroups`，`_id` 为 `"{workspace_id}:{name}"` 复合键
  - `cron_jobs` / `cron_executions`，`_id` 为 `"{workspace_id}:{name}"`（job）或随机 uuid（execution，带 `workspace_id` 字段）
  - `invocations`：runner 持久化的每次调用记录（agent / app / user / session / status / input / output / latency / workspace_id）
  - `api_tokens`：DB-stored API tokens（哈希 + prefix + workspace_id）
  - `forum_threads` / `forum_posts`：Forum 讨论帖
  - `mcp_oauth_connections`：MCP OAuth 凭据（workspace + server 维度）
  - `oauth_states`：Dashboard OAuth 与 MCP OAuth 流程 state
  - Agent file 元数据（内容在 S3 或 memory 后端，见 `docs/storage.md`）
- **Redis**（默认 `localhost:6379`）：dashboard auth sessions、渠道内活跃 agent/model 选择。

## 11. 鉴权与多 Token

- **Dashboard user session**：`AuthService.Login` 或 `CompleteOAuthFlow` 换取 opaque session token；用户记录存 MongoDB，session hash 存 Redis。后续请求带 `Authorization: Bearer <token>` + `X-Workspace-ID`。
- **OAuth 登录**：`ListOAuthProviders` / `BeginOAuthFlow` / `CompleteOAuthFlow`（GitHub、Google 等，由 `internal/auth/provider` 注册）；首次登录自动创建用户并绑定 provider `external_id`。
- **Root token**：`config.yaml` 中的 `apiToken`，单值，constant-time 比对，配合 CLI / 应急。Root token 默认接受 `X-Workspace-ID` 头原值。
- **DB-stored tokens**：通过 `APITokenService` 在运行时创建，每个 token 在创建时绑定到当前 `X-Workspace-ID`：
  - 格式 `bt_<48 hex>`；plaintext 仅在 `CreateAPIToken` 返回一次。
  - 存 `sha256(secret)` 哈希，列出时只返回前 12 字符 prefix。
  - 命中后异步更新 `last_used_at`。`RevokeAPIToken` 直接失效。
  - 用 API token 鉴权时自动覆盖请求 workspace，忽略客户端 header。
- **作用范围**：当前任何合法 token 在其 workspace 内拥有全量 API 权限；未实现细粒度 scope。
- Daemon gRPC 也走同一份 root `apiToken`（`metadata` 的 `authorization` 头），不进入 workspace 隔离。

## 12. Invocation 追踪与活动流

- `runner.Service` 接口注入 `InvocationRecorder`：
  - Run 开始时记 `INVOCATION_STATUS_RUNNING` + `started_at`。
  - 命名返回 + defer 在结束时回写 `SUCCEEDED` / `FAILED` + `output` / `error` + `latency_ms`（input/output/error 截到 4096 字符）。
  - 记录失败只 warn 日志，不阻塞 Run。
- `runner.Service.CancelInvocation(id)` 调用注册的 `context.CancelFunc` 取消在飞 invocation；`AgentService.CancelAgentInvocation` 把信号送过去。
- `AgentService.ListAgentInvocations`：按 agent / session 过滤 + 分页。
- `DashboardService.GetActivityFeed`：把最近 invocation 映射成 `ActivityEvent`（kind 派生自 status）。
- `AgentRuntimeStatus`（`GetAgentRuntimeStatus` / `ListAgentRuntimeStatuses`）从最近 100 条 invocation 派生 state / last_run_at / in_flight，驱动前端 Agents 表的 Status 列。

## 13. 在线探活与状态

- `MCPServerService.GetMCPServerStatus`：实跑 MCP handshake（streamable HTTP / SSE）+ `ListTools`，应用 `tool_filter`，返回 `STATE_CONNECTED` + tool_count 或 `STATE_DISCONNECTED` + 错误 detail。
- `MCPServerService.ListMCPTools`：聚合所有 MCP 工具到一个视图，per-tool `allowed` 反映白名单；server 探测失败放进 `errors` map。
- `RemoteAgentService.GetRemoteAgentStatus`：A2A 拉 `/.well-known/agent.json`；DAEMON 协议查注册表，返回 ACTIVE / IDLE / UNREACHABLE + `serving_daemon_id`。
- `ChannelService.GetChannelStatus`：返回 LIVE / PAUSED / ERROR；`channel.Manager.ChannelStatus()` 当前主要看 `enabled` + manager started。
- `DaemonService.ListDaemons` / `GetDaemon`：暴露 daemon 注册表中 version / os / executors / remote_addr / uptime / active task 数。
- `DaemonService.ListDaemonTasks`：每个在飞任务带 `current_step` / `progress` / `elapsed`（progress 从 daemon 的 `DaemonTaskUpdate` 上报）。

## 14. Bridge 诊断

- `internal/runtime/daemon/metrics.go` 维护一个共享 `Metrics` collector：
  - 每次 bridge invocation 调 `RecordLatency(d)` 入 60 条 ring buffer。
  - `Snapshot()` 返回 `MemStats.Sys` 内存、`runtime.NumGoroutine()`、以及自启动以来的平均 CPU%（`runtime/metrics` 的 `/cpu/classes/total:cpu-seconds`）。
- `DaemonService.GetBridgeDiagnostics` 把 snapshot 转 proto，前端 Daemon 监控屏直接消费。

## 15. 可观测性

- **OpenTelemetry Tracing**：`BUTTERFLY_TRACING_PROVIDER`、`BUTTERFLY_TRACING_ENDPOINT` 控制。
- **Langfuse Plugin**：模型调用层 LLM 追踪。当 `cfg.Langfuse.Host` 设置时，`SessionEvent.trace_url = <host>/trace/<invocation_id>` 自动填充。
- **Cron Execution 历史 + Invocation 历史**：作为业务级运行记录，分别支撑 Cron 时序图与 Activity Feed。
- **Bridge Metrics**：见第 14 节，给运维屏提供 CPU / 内存 / latency 折线。

## 16. 前端 Dashboard

- `front/` 是 Vite + React 19 + shadcn/ui 应用，TanStack Query 做数据层。
- Proto TS 绑定通过 `buf.build/bufbuild/es` 输出到 `front/src/gen/`。手写 `front/src/types/api.ts` 维护 snake_case 形状供 `twirpFetch` 消费。
- 主要路由区域：Overview、Integrations、Operations、Agents（含 Agent Files）、MCP Servers、Model Providers、Notify Groups、Remote Agents、Daemons、Channels、Sessions、Chat、Forum、Cron、API Tokens、Users、Workspaces、Profile、Admin Global MCP。
- 全部页面消费上文 RPC/REST；细节见 `docs/api.md`。前端仍缺的能力见 `docs/frontend-required-apis.md`。

## 17. 部署形态

- **后端镜像**：`ghcr.io/<owner>/<repo>`（根 `Dockerfile`，distroless static Go 二进制 + cosign 签名）。
- **前端镜像**：`ghcr.io/<owner>/<repo>-front`（`front/Dockerfile`，node:22-alpine 编译 + nginx:1.27-alpine 运行；带 SPA fallback + `/healthz`）。
- 两个工作流分别为 `.github/workflows/docker-publish.yml` 和 `front-publish.yml`，PR 阶段只 build，main / tag 阶段 push + sign。

## 18. 进程拓扑

```text
cmd/butter (服务端)
  ├── HTTP / Twirp / A2A          (:8080)
  ├── Telegram / Discord poller
  ├── Cron scheduler
  ├── Runner + Invocation recorder
  └── gRPC DaemonConnectorService (:9090)

cmd/butter-daemon (客户端)
  └── 反连服务端 -> 本地 executor (shell / opencode / claude-code)
```

## 19. 当前限制

- `runner.Service.Run` 仍以同步返回最终文本为主，长耗时 daemon 任务会占用调用链。
- MCP toolset 当前支持 streamable HTTP 与 SSE transport。
- A2A remote agent 需要 `url`；daemon remote agent 需要 `daemon_capability` 与在线 daemon。
- `BridgeDiagnostics.memory_limit_bytes` 当前总是 0（未读 cgroup）。
- API tokens 在所属 workspace 内拥有全量 API 权限，没有 `expires_at` / `scopes`。
- Workspace 跨租户隔离尚未到运行时：runner、channel manager、cron scheduler 共享同一个进程视图，**agent 名字需在全部 workspace 内全局唯一**；不同 workspace 用同名 agent 会冲突。
- 内置 system agent 与 daemon gRPC 仍为全局；未做 per-workspace 实例化。
- `pkg/proto/agents/v1` 为生成代码，改动需在 `proto/agents/v1` 中完成后重新生成。
