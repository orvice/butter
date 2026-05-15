# Butter 功能总览

更新时间：2026-05-15

Butter 是基于 Butterfly 框架的 Agent 服务，核心使命是把多种入口（HTTP / Twirp / gRPC / 即时消息 / 定时任务）统一编排为 Google ADK Agent 执行流，并提供配置化、热更新、多执行面与持久化运行时。

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
  - `MCP_SERVER_TRANSPORT_STDIO`（命令 + 参数 + 环境变量）
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
- `GET /a2a/:agent_name/.well-known/agent.json`：A2A agent card（仅 `enable_a2a: true` 的 Agent）。
- `POST /a2a/:agent_name`：A2A JSON-RPC `tasks/send`。

### Twirp RPC（`/api`）

- `AgentService`：Agent 配置 CRUD。
- `MCPServerService`：共享 MCP Server CRUD。
- `RemoteAgentService`：远程 Agent CRUD。
- `ChannelService`：渠道配置 CRUD。
- `SessionService`：会话查询、回复（`ReplySession`）、清理。
- `CronJobService`：定时任务 CRUD + 执行记录查询（分页）。

除 `/ping` 外，HTTP/Twirp 请求统一经过 `APITokenAuthMiddleware`，使用 `Authorization: Bearer <apiToken>` 鉴权。

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

## 8. Cron 自动执行

- **CronJob 配置**：name、schedule、agent_name、input、timezone、enabled、delivery、metadata。
- **结果投递（CronDelivery）**：
  - `CRON_DELIVERY_TYPE_LOG`：写日志。
  - `CRON_DELIVERY_TYPE_WEBHOOK`：HTTP webhook 推送。
  - `CRON_DELIVERY_TYPE_CHANNEL`：转发到指定 `AgentChannel` 的 `chat_id`。
- **执行记录（CronExecution）**：每次执行写入 MongoDB，包含输入、输出、状态、起止时间，支持分页查询。

## 9. 配置与热更新

- **两层配置**：
  - `AppConfig`：Butterfly 从 `config/butter.yaml` 加载的启动配置。
  - `ConfigStore`：运行时配置仓库，支持 `memory` 或 `mongo` 后端。
- **Seed 行为**：mongo 后端首次启动时若集合为空则从启动配置 seed，否则以 Mongo 中的数据为准。
- **运行时热更新**：
  - Agent / MCP / RemoteAgent 变更触发 `ReloadRunner`，重建 agent registry，清空 runner 与 model override 缓存，并 reload channels。
  - Channel 变更触发 `ReloadChannels`。
- **持久化对象**：agents、mcp_servers、remote_agents、channels、cron_jobs、cron_executions。

## 10. 持久化

- **MongoDB**（默认数据库 `butter`，可通过 `mongo_db` 配置）：
  - ADK sessions / events
  - ADK memories
  - 运行时配置（agents、mcp、remote agents、channels）
  - Cron jobs 与 executions
- **Redis**（默认 `localhost:6379`）：渠道内活跃 agent/model 选择；Redis 不可用不会阻塞服务启动。

## 11. 可观测性

- **OpenTelemetry Tracing**：`BUTTERFLY_TRACING_PROVIDER`、`BUTTERFLY_TRACING_ENDPOINT` 控制。
- **Langfuse Plugin**：模型调用层 LLM 追踪。
- **Cron Execution 历史**：作为业务级运行记录。

## 12. 进程拓扑

```text
cmd/butter (服务端)
  ├── HTTP / Twirp / A2A
  ├── Telegram / Discord poller
  ├── Cron scheduler
  └── gRPC DaemonConnectorService

cmd/butter-daemon (客户端)
  └── 反连服务端 -> 本地 executor (shell / opencode / claude-code)
```

## 13. 当前限制

- `runner.Service.Run` 仍以同步返回最终文本为主，长耗时 daemon 任务会占用调用链。
- MCP toolset 当前支持 streamable HTTP 与 SSE transport（stdio 通过 MCP server config 支持，转发由 ADK 处理）。
- A2A remote agent 需要 `url`；daemon remote agent 需要 `daemon_capability` 与在线 daemon。
- `pkg/proto/agents/v1` 为生成代码，改动需在 `proto/agents/v1` 中完成后重新生成。
