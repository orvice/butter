# Butter 功能总览

更新时间：2026-06-13

Butter 是基于 Butterfly 框架的 Agent 服务，核心使命是把多种入口（HTTP / ConnectRPC / gRPC / 即时消息 / 定时任务）统一编排为 Google ADK Agent 执行流，并提供配置化、热更新、多执行面、多租户与持久化运行时。

## 0. Workspace 多租户

- **Workspace 实体**：所有 Agent / Channel / MCP Server / Remote Agent / Model Provider / Notify Group / Agent File / Forum / Cron Job / Automation / API Token / Invocation / Cron Execution / Automation Run 必须归属一个 workspace。`WorkspaceService` 暴露 workspace 与成员（`owner` / `admin` / `member`）的 CRUD。
- **请求级别 workspace 选择**：客户端通过 `X-Workspace-ID` HTTP 头选择活跃 workspace；`AuthMiddleware` 校验该用户是 workspace 成员（全局 `admin` 角色旁路）。workspace-scoped RPC 缺失 header 时返回 `failed_precondition`；`AuthService` / `WorkspaceService` / `DashboardService` 不需要该 header，`DaemonService` 的配置、credential 签发、在线 daemon / task 查询需要 workspace，daemon 原生 gRPC `Connect` 由 daemon credential 自带 workspace。`SessionService.ReplySession` 建议带 header 以便 runner 在正确 workspace 解析 agent。
- **登录返回 workspace 列表**：`AuthService.Login` 在响应中带 `workspaces`，dashboard 登录后弹 workspace 选择器，把选中的 workspace 写入后续请求头。
- **API token 自带 workspace**：`CreateAPIToken` 时记录创建上下文的 workspace；用 API token 鉴权时直接绑定到 token 的 workspace，忽略 header。
- **DaemonRuntime 自带 workspace**：daemon 执行面是配置仓库中的 `DaemonRuntime` 资源，runtime token 绑定 workspace + daemon_runtime_id；daemon-backed RemoteAgent 在 runtime 上选择 `opencode` / `codex` 这类 ACP runtime。
- **默认 workspace 自举**：进程启动时若 `workspaces` 集合为空，会自动创建 slug 为 `default` 的 workspace，并把现有所有 user 加为 `owner`。
- **运行时行为**：runner / channel manager / cron scheduler / automation scheduler 跨 workspace 拉平所有配置作为全局视图运行（agent 名字目前需全局唯一）；`ContextInfo.workspace_id` 把所属 workspace 沿调用链透传，写 invocation / cron execution / automation run 时回填。
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

## 2. 模型管理

- **模型别名与 Provider 解析**：通过 `model_providers` 配置把别名（如 `flash`）映射到具体模型。
- **运行时 Model Override**：渠道/调用方可在调用时指定 `model_override`；Runner 会 clone 配置、替换模型并缓存 override 后的 Agent 实例。
- **Langfuse Tracing**：运行时初始化 Langfuse plugin 支持模型调用追踪。

## 3. MCP 工具集

- **共享 MCP Server 配置**：通过 `MCPServerService` CRUD 管理共享 MCP Server。
- **多种 Transport**：
  - `MCP_SERVER_TRANSPORT_STREAMABLE_HTTP`
  - `MCP_SERVER_TRANSPORT_SSE`
- **工具白名单**：`tool_filter` 控制对外暴露的 MCP 工具。
- **Agent 关联方式**：Agent 可通过 `mcp_servers`（内联）或 `mcp_server_ids`（引用共享配置）挂载 MCP 工具集。

## 4. 远程 Agent

- **A2A 协议**：`REMOTE_AGENT_PROTOCOL_A2A`，通过 `url` 调用远程 ADK Agent。
- **Daemon 协议**：`REMOTE_AGENT_PROTOCOL_DAEMON`，在当前 workspace 内通过 `daemon_runtime_id` 找到在线 runtime，并用 `acp_runtime` 选择 `opencode` / `codex` adapter；`daemon.Bridge` 把 ADK invocation 转成带 `workspace_id` 的 `DaemonTask`，等 daemon 回传 terminal update 后生成 ADK final event。
- **断连/取消**：连接断开时活跃任务会收到失败更新；context 取消时下发 `CancelTask`。

## 5. Daemon 执行面

- **反向连接架构**：daemon client（`cmd/butter-daemon --url http://localhost:8081/api --token <runtime-token>`）主动连接服务端 `DaemonConnectorService.Connect`，用 workspace-scoped runtime token 鉴权，注册 `DaemonInfo` 与支持的 `acp_runtimes`。
- **DB 配置优先**：daemon 必须先通过 `DaemonService.CreateDaemonRuntime` 写入 workspace runtime，再通过 `CreateDaemonRuntimeToken` 生成 worker token。服务端用 token 决定 authoritative workspace + daemon_runtime_id；同一 workspace/runtime 同时只允许一个 daemon 连接。
- **任务下发与执行**：服务端通过 `daemon.GRPCHandler` / `Registry` / `Connection` / `Bridge` 把任务下发到 daemon，daemon 根据 `DaemonTask.acp_runtime` 选择本地 ACP executor 后回传 `DaemonTaskUpdate`。opencode/codex 分别通过 `opencode acp` / `codex-acp` 这类 ACP stdio 入口接入。
- **h2c 入口**：daemon worker 连接使用独立 `:8081` h2c 端口，路径为 `/api/agents.v1.DaemonConnectorService/Connect`。

## 6. 多入口接入

### HTTP（Gin）

- `GET /ping`：健康检查，免鉴权。
- `GET /a2a/:agent_name/.well-known/agent.json`：A2A agent card（仅 `enable_a2a: true` 的 Agent）。
- `POST /a2a/:agent_name`：A2A JSON-RPC `tasks/send`。
- `POST /api/uploads/*`：头像/静态资源 multipart 上传（REST，非 Connect）；见 `docs/storage.md`。

### RPC（`/api`，ConnectRPC，同时支持 Connect / gRPC-Web / gRPC）

给外部 App / Dashboard 开发者的可复制接入说明集中在 `docs/api.md`，包括 URL 形状、`Authorization` / `X-Workspace-ID` header、JSON snake_case 字段、TypeScript Connect-Web 示例、chat streaming 和错误处理。

配置类：

- `AgentService`：Agent 配置 CRUD（含 `page_size`/`page_token` 分页）+ `InvokeAgent` / `StreamAgent`（dashboard chat server-stream）/ `CancelAgentInvocation` / `ReloadAgents` / `GetAgentRuntimeStatus` / `ListAgentRuntimeStatuses` / `ListAgentInvocations`。
- `MCPServerService`：共享 MCP Server CRUD + `GetMCPServerStatus`（live 探活）+ `ListMCPTools`（聚合工具列表）+ `StartMCPServerOAuth` / `CompleteMCPServerOAuth` / `GetMCPServerOAuthStatus` / `DisconnectMCPServerOAuth`（MCP OAuth2 授权流程）。
- `RemoteAgentService`：远程 Agent CRUD + `GetRemoteAgentStatus`（A2A `/.well-known/agent.json` 探测 / Daemon 注册表查找）。
- `ChannelService`：渠道配置 CRUD + `GetChannelStatus` + `RestartChannel` / `PauseChannel` / `ResumeChannel`。
- `AutomationService`：自动化工作流 CRUD + `RunAutomationNow` + `ListAutomationRuns` / `GetAutomationRun` / `ListAutomationStepRuns`。
- `CronJobService`：定时任务 CRUD + `ListCronExecutions`（分页）+ `RunCronJobNow`。
- `SessionService`：`Create` / `Get`（含 duration / event trace_url）/ `List`（channel/user/date 过滤 + 分页 + total）/ `Delete` / `Reply`。
- `ModelProviderService`：LLM Provider 配置 CRUD（workspace 范围）。
- `NotifyGroupService`：通知组 CRUD，供 cron 投递。
- `AgentFileService`：workspace 范围的文件空间与文件 CRUD（含 `SearchAgentFiles`）。
- `ForumService`：论坛 thread / post 管理。

运维类：

- `DashboardService`：`GetOverview`（counts + health + 最新 daemon 握手）/ `GetActivityFeed`（最近 invocation）/ `GetCronExecutionTimeseries`（1D/7D/30D 桶）。
- `DaemonService`：`ListDaemonRuntimes` / `GetDaemonRuntime` / `CreateDaemonRuntime` / `UpdateDaemonRuntime` / `DeleteDaemonRuntime` / `CreateDaemonRuntimeToken` / `ListDaemons` / `GetDaemon` / `ListDaemonTasks`（含 step+progress+elapsed）/ `CancelDaemonTask` / `GetBridgeDiagnostics`（CPU / 内存 / 延迟样本）。
- `APITokenService`：`ListAPITokens` / `CreateAPIToken` / `RevokeAPIToken`。
- `GlobalMCPServerService`：workspace-agnostic MCP server 预设 CRUD（创建/更新/删除限 admin）+ `InstallGlobalMCPServer`（克隆到 workspace；admin 可跨 workspace 安装）。
- `WorkspaceService`：workspace 与成员 CRUD。
- `AuthService`：登录、OAuth 登录、当前用户、用户管理、资料更新与 `ChangePassword`。

除 `/ping`、`OPTIONS` 预检、MCP OAuth 回调以及 `AuthService.Login` / `ListOAuthProviders` / `BeginOAuthFlow` / `CompleteOAuthFlow` 外，所有 HTTP/RPC 请求统一经过 `AuthMiddleware`：先查 Redis auth session（key `butter:auth:session:<sha256(token)>`），再 root token（`cfg.apiToken`，constant-time），再 `api_tokens` 集合中的 sha256 哈希；命中均会异步更新 `last_used_at`。鉴权通过后再解析 `X-Workspace-ID`：user session 走成员校验，API token 用其绑定的 workspace，root token 接受 header 原值。

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
- 支持 timeout、retry/backoff、并发策略（skip / queue / replace / allow）、通知策略（always / failure / success）和输出截断。

## 7. 会话与记忆

- **ADK Session**：MongoDB session service 持久化会话事件，支持按 channel/user/session 查询、列表、删除、回复。
- **ADK Memory**：MongoDB memory service 保存长期记忆。
- **ContextInfo**：runner 调用统一携带 channel、session、user、source、uuid，作为执行上下文。
- **会话维度的 Agent Runner 缓存**：按 `channel:agent:model` 维度缓存 ADK runner 实例。

## 8. 自动化工作流

- **Automation 配置**：workspace-scoped，包含 name、enabled、trigger、conditions、ordered steps、policy、metadata。v1 运行 manual 和 schedule trigger，同时在模型中保留 webhook / forum / channel / daemon event trigger 形状，便于后续事件触发接入。
- **条件模型**：每个 condition 用 selector + operator + value 表达，支持 equals / not equals / contains / regex match / exists / not exists。所有条件都通过后才执行 steps；任一失败则 run 标记为 skipped 且不创建 step-run。
- **线性 steps**：按定义顺序串行执行，默认首个失败 step 会停止整个 run。支持 `INVOKE_AGENT`、`CALL_WEBHOOK`、`SEND_NOTIFY_GROUP`、`CREATE_FORUM_POST`。
- **执行策略**：Automation 级 policy 支持 timeout、retry/backoff、concurrency、max_output_bytes、notify_on；step 可覆盖自身 policy。
- **运行记录**：每次执行写 `automation_runs`，包含 trigger type、status、trigger payload preview、error、起止时间和 duration；每个实际执行的 step 写 `automation_step_runs`，包含 step type、attempt_count、input/output preview、invocation_id、truncated、duration 等。
- **调度**：启动时 automation scheduler 注册所有 enabled + schedule trigger 的 automation；create/update/delete 自动 reschedule 或 unregister。disabled automation 只持久化，不进入调度器。

## 9. Cron 自动执行

- **CronJob 配置**：name、schedule、agent_name、input、timezone、enabled、delivery、timeout、retry、concurrency_policy、notify_on、max_output_bytes、metadata。
- **结果投递（CronDelivery）**：
  - `CRON_DELIVERY_TYPE_LOG`：写日志。
  - `CRON_DELIVERY_TYPE_WEBHOOK`：HTTP webhook 推送。
  - `CRON_DELIVERY_TYPE_CHANNEL`：转发到指定 `AgentChannel` 的 `chat_id`。
  - `CRON_DELIVERY_TYPE_NOTIFY_GROUP`：按名称加载 `NotifyGroup`，向其中启用的 Telegram、Lark webhook、Discord webhook 目标发送通知。
- **可靠性策略**：每次执行可设置超时、失败重试、重试 backoff、并发处理（默认 skip 保持旧行为）、投递通知时机和输出预览上限。手动 `RunCronJobNow` 与定时触发使用同一执行路径。
- **执行记录（CronExecution）**：每次开始或被并发策略跳过的执行写入 MongoDB，包含 input/output preview、status（success/error/skipped/cancelled）、error、attempt_count、trigger_type、skipped_reason、truncated、起止时间和 duration，支持分页查询。

## 10. 配置与热更新

- **两层配置**：
  - `AppConfig`：Butterfly 从 `BUTTERFLY_CONFIG_FILE_PATH` 指向的 YAML 文件加载启动配置；仓库中的示例文件是根目录 `config.yaml`，部署时可复制为 `config/butter.yaml` 等路径。
  - `ConfigStore`：运行时配置仓库，支持 `memory` 或 `mongo` 后端。
- **Seed 行为**：mongo 后端启动时，`ConfigStore.InitFromConfig`（即 `SyncToConfig`）从 MongoDB 中读取配置并同步到 `AppConfig`；**不会**把 YAML 中的 `agents` / `mcp_servers` 等字段写入 Mongo。如需初始化，请通过 RPC（带 `X-Workspace-ID`）写入。
- **运行时热更新**：
  - Agent / MCP / RemoteAgent 变更触发 `ReloadRunner`，重建 agent registry，清空 runner 与 model override 缓存，并 reload channels。
  - Channel 变更触发 `ReloadChannels`。
- **持久化对象**：agents、mcp_servers、remote_agents、channels、cron_jobs、cron_executions、automations、automation_runs、automation_step_runs。

## 11. 持久化

- **MongoDB**（默认数据库 `butter`，可通过 `mongo_db` 配置）：
  - `adk_sessions` / `adk_events`：ADK session 与事件
  - ADK memories
  - `workspaces`：workspace 元数据，`slug` 唯一索引
  - `workspace_members`：用户—workspace 多对多关系，`(workspace_id, user_id)` 唯一索引 + `user_id` 普通索引
  - `users`：dashboard 用户
  - 运行时配置：`config_agents` / `config_mcpservers` / `config_remoteagents` / `config_daemons` / `config_channels` / `config_modelproviders` / `config_notifygroups`，`_id` 为 `"{workspace_id}:{name}"` 或 `"{workspace_id}:{id}"` 复合键
  - Agent Files：`agent_file_spaces` / `agent_files` / `agent_file_versions`
  - Forum：`forum_threads` / `forum_posts`
  - `cron_jobs` / `cron_executions`，`_id` 为 `"{workspace_id}:{name}"`（job）或随机 uuid（execution，带 `workspace_id` 字段）
  - `automations` / `automation_runs` / `automation_step_runs`，分别保存 workflow 定义、run 历史和 step-run 历史；definition `name` 在 workspace 内唯一，run/step-run 按 workspace + automation/run 建索引。
  - `invocations`：runner 持久化的每次调用记录（agent / app / user / session / status / input / output / latency / workspace_id）
  - `api_tokens`：DB-stored API tokens（哈希 + prefix + workspace_id + kind + scopes + optional expires_at / daemon_runtime_id）
- **Redis**（默认 `localhost:6379`）：dashboard auth sessions（key `butter:auth:session:<hash>`）、渠道内活跃 agent/model 选择。

## 12. 鉴权与多 Token

- **Dashboard user session**：`AuthService.Login` 用用户名+密码换取一个 opaque session token；token hash 存储在 **Redis**（key `butter:auth:session:<sha256(token)>`），登录响应同时返回该用户可访问的 workspace。后续请求带 `Authorization: Bearer <token>` + `X-Workspace-ID`。
- **Root token**：`config.yaml` 中的 `apiToken`，单值，constant-time 比对，配合 CLI / 应急。Root token 默认接受 `X-Workspace-ID` 头原值。
- **DB-stored tokens**：通过 `APITokenService` 在运行时创建，每个 token 在创建时绑定到当前 `X-Workspace-ID`：
  - 格式 `bt_<48 hex>`；plaintext 仅在 `CreateAPIToken` 返回一次。
  - 存 `sha256(secret)` 哈希，列出时只返回前 12 字符 prefix。
  - 普通 API token 为 `API_TOKEN_KIND_USER` + `api:*` scope。
  - Daemon runtime token 为 `API_TOKEN_KIND_DAEMON` + `daemon:connect` scope，只能由 `DaemonService.CreateDaemonRuntimeToken` 为已有 runtime 签发。
  - 命中后异步更新 `last_used_at`。`RevokeAPIToken` 直接失效。
  - 用 API token 鉴权时自动覆盖请求 workspace，忽略客户端 header。
- **作用范围**：普通 API token 只能进入 HTTP/RPC API；daemon runtime token 只能进入 `DaemonConnectorService.Connect`，且必须匹配 daemon_runtime_id 与 workspace。

## 13. Invocation 追踪与活动流

- `runner.Service` 接口注入 `InvocationRecorder`：
  - Run 开始时记 `INVOCATION_STATUS_RUNNING` + `started_at`。
  - 命名返回 + defer 在结束时回写 `SUCCEEDED` / `FAILED` + `output` / `error` + `latency_ms`（input/output/error 截到 4096 字符）。
  - 记录失败只 warn 日志，不阻塞 Run。
- `runner.Service.CancelInvocation(id)` 调用注册的 `context.CancelFunc` 取消在飞 invocation；`AgentService.CancelAgentInvocation` 把信号送过去。
- `AgentService.ListAgentInvocations`：按 agent / session 过滤 + 分页。
- `DashboardService.GetActivityFeed`：把最近 invocation 映射成 `ActivityEvent`（kind 派生自 status）。
- `AgentRuntimeStatus`（`GetAgentRuntimeStatus` / `ListAgentRuntimeStatuses`）从最近 100 条 invocation 派生 state / last_run_at / in_flight，驱动前端 Agents 表的 Status 列。

## 14. 在线探活与状态

- `MCPServerService.GetMCPServerStatus`：实跑 MCP handshake（streamable HTTP / SSE）+ `ListTools`，应用 `tool_filter`，返回 `STATE_CONNECTED` + tool_count 或 `STATE_DISCONNECTED` + 错误 detail。
- `MCPServerService.ListMCPTools`：聚合所有 MCP 工具到一个视图，per-tool `allowed` 反映白名单；server 探测失败放进 `errors` map。
- `RemoteAgentService.GetRemoteAgentStatus`：A2A 拉 `/.well-known/agent.json`；DAEMON 协议在当前 workspace 查注册表，返回 ACTIVE / IDLE / UNREACHABLE + `serving_daemon_runtime_id`。
- `ChannelService.GetChannelStatus`：返回 LIVE / PAUSED / ERROR；`channel.Manager.ChannelStatus()` 当前主要看 `enabled` + manager started。
- `DaemonService.ListDaemonRuntimes` / `CreateDaemonRuntimeToken`：管理 workspace daemon runtime 与 worker token。
- `DaemonService.ListDaemons` / `GetDaemon`：暴露当前 workspace daemon 注册表中 version / os / executors / remote_addr / uptime / active task 数。
- `DaemonService.ListDaemonTasks`：每个在飞任务带 `current_step` / `progress` / `elapsed`（progress 从 daemon 的 `DaemonTaskUpdate` 上报）。

## 15. Bridge 诊断

- `internal/runtime/daemon/metrics.go` 维护一个共享 `Metrics` collector：
  - 每次 bridge invocation 调 `RecordLatency(d)` 入 60 条 ring buffer。
  - `Snapshot()` 返回 `MemStats.Sys` 内存、`runtime.NumGoroutine()`、以及自启动以来的平均 CPU%（`runtime/metrics` 的 `/cpu/classes/total:cpu-seconds`）。
- `DaemonService.GetBridgeDiagnostics` 把 snapshot 转 proto，前端 Daemon 监控屏直接消费。

## 16. 可观测性

- **OpenTelemetry Tracing**：`BUTTERFLY_TRACING_PROVIDER`、`BUTTERFLY_TRACING_ENDPOINT` 控制。
- **Langfuse Plugin**：模型调用层 LLM 追踪。当 `cfg.Langfuse.Host` 设置时，`SessionEvent.trace_url = <host>/trace/<invocation_id>` 自动填充。
- **Cron Execution / Automation Run / Invocation 历史**：作为业务级运行记录，分别支撑 Cron 时序图、Automation 详情页与 Activity Feed。
- **Bridge Metrics**：见第 15 节，给运维屏提供 CPU / 内存 / latency 折线。

## 17. 前端 Dashboard

- `front/` 是 Vite + React 19 + shadcn/ui 应用，TanStack Query 做数据层。
- Proto TS 绑定通过 `buf.build/bufbuild/es`（`include_imports: true`）输出到 `front/src/gen/`，service 定义和 message 类型都包含在内（connect-es v2 直接消费 `GenService`）。每个 service 一个 `front/src/api/*.ts`，用 `makeClient(XxxService)` 拿到类型化 client；共享 `transport.ts` 注入 `Authorization` / `X-Workspace-ID`，默认 **binary protobuf**（`useBinaryFormat: true`），并处理 401 跳登录。手写 `front/src/types/api.ts` 仍保留 snake_case 形状作为 page 层 boundary。Chat 流式走 `AgentService.StreamAgent`（`chat.ts`）；头像上传走 REST multipart（`uploads.ts`），上传后再调 `AuthService.UpdateProfile` 写 `avatar_url`。
- 一级页（`front/src/pages/`）：Login / Chat / Forum / Dashboard(Overview) / Agents / MCP Servers / Remote Agents / Daemons / Channels / Sessions / Cron / Automations / API Tokens / Model Providers / Notify Groups / Agent Files / Workspaces / Users / Profile / Integrations / Operations / Admin。
- 全部页面消费上面 12-16 节描述的 RPC；细节见 `docs/api.md`。

## 18. 部署形态

- **后端镜像**：`ghcr.io/<owner>/<repo>`（根 `Dockerfile`，distroless static Go 二进制 + cosign 签名）。
- **前端镜像**：`ghcr.io/<owner>/<repo>-front`（`front/Dockerfile`，node:22-alpine 编译 + nginx:1.27-alpine 运行；带 SPA fallback + `/healthz`）。
- 两个工作流分别为 `.github/workflows/docker-publish.yml` 和 `front-publish.yml`，PR 阶段只 build，main / tag 阶段 push + sign。

## 19. 进程拓扑

```text
cmd/butter (服务端)
  ├── HTTP / ConnectRPC / A2A     (:8080)
  ├── h2c ConnectRPC              (:8081 /api)
  ├── DaemonConnectorService      (:8081 /api)
  ├── Telegram / Discord poller
  ├── Cron scheduler
  ├── Automation scheduler
  └── Runner + Invocation recorder

cmd/butter-daemon (客户端)
  └── 反连服务端 -> 本地 executor (ACP: opencode / claude-code / gemini, shell)
```

## 20. 当前限制

- `runner.Service.Run` 仍以同步返回最终文本为主，长耗时 daemon 任务会占用调用链。
- MCP toolset 当前支持 streamable HTTP 与 SSE transport。
- A2A remote agent 需要 `url`；daemon remote agent 需要 `daemon_runtime_id`、`acp_runtime` 与在线 daemon runtime。
- `BridgeDiagnostics.memory_limit_bytes` 当前总是 0（未读 cgroup）。
- 当前普通 API token 在所属 workspace 内拥有 `api:*` 权限；daemon token 只有 `daemon:connect`，并可设置 `expires_at`。
- Workspace 跨租户隔离尚未到运行时：runner、channel manager、cron scheduler 共享同一个进程视图，**agent 名字需在全部 workspace 内全局唯一**；不同 workspace 用同名 agent 会冲突。
- Automation v1 是线性有序 step，不支持 DAG、人工审批 gate 或持久化 worker queue；webhook/forum/channel/daemon event trigger 字段已建模但尚未接入事件路由。
- 内置 system agent 仍为全局注册；daemon connector 是 `/api` 下的长连接入口，但连接、registry、任务路由和配置均按 workspace 隔离。
- `pkg/proto/agents/v1` 为生成代码，改动需在 `proto/agents/v1` 中完成后重新生成。
