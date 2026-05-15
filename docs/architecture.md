# Butter 系统架构

更新时间：2026-05-15

## 概览

Butter 是基于 Butterfly 框架的 Agent 服务。系统把 HTTP/Twirp/gRPC/channel 输入统一转为 ADK agent 执行，并通过 MongoDB、Redis、运行时配置仓库和 daemon 长连接支撑会话、记忆、渠道状态、定时任务和远程执行。

核心能力：

- Agent 配置化：从 YAML 或配置仓库加载 `agents.v1.Agent`，构建 LLM、Loop、Sequential、Parallel agent。
- 多入口接入：Gin HTTP、Twirp RPC、Telegram、Discord、Cron、A2A 和 daemon gRPC。
- 运行时热更新：Agent、MCP Server、Remote Agent、Channel 配置可通过 RPC 修改并触发 runner/channel reload。
- 多执行面：本地 ADK agent、A2A 远程 agent、daemon 反向连接 agent。
- 持久化运行时：MongoDB 保存 ADK session/memory、配置、cron 执行记录；Redis 保存渠道内活跃 agent/model 选择。

## 进程入口

```text
cmd/butter/main.go
  -> core.New(...)
  -> SetupRoutes(cfg)
  -> SeedConfig(ctx, cfg)
  -> StartChannels(...)
  -> Wire(bootstrap result)
  -> SetupGRPCServer(...)
```

`cmd/butter` 是主服务进程，负责启动 Butterfly HTTP 服务和 daemon gRPC 服务。

`cmd/butter-daemon` 是 daemon client 进程，主动连接服务端 gRPC `DaemonConnectorService.Connect`，注册自身能力，接收任务并通过本地 executor 执行。

## 分层结构

```text
Access Layer
├── Gin HTTP handlers: /ping, /a2a/:agent_name/...
├── Twirp RPC: /api/agents.v1.*Service/*
├── gRPC: DaemonConnectorService.Connect
├── Telegram poller
├── Discord poller
└── Cron scheduler

Application / Transport Services
├── internal/application/*ServiceServer
├── internal/handler/http
└── internal/app routes/grpc/bootstrap wiring

Runtime Layer
├── runner.Service
├── cron.Scheduler
├── daemon.Registry / Connection / Bridge / GRPCHandler
├── session/mongo
└── memory/mongo

Agent Layer
├── internal/agent.NewFromProto()
├── model provider resolution
├── MCP toolset construction
├── A2A remote agent resolution
├── daemon remote agent bridge
└── built-in system agent

Config Layer
├── AppConfig loaded by Butterfly
├── ConfigStore runtime backend wrapper
├── repo/config interfaces
├── repo/config/memory
└── repo/config/mongo

Persistence
├── MongoDB: session, memory, config, cron
└── Redis: channel active agent/model selection
```

## 启动装配

`internal/app` 是服务装配层：

- `routes.go` 创建 Gin handler、Twirp server 和 `Handlers` 容器。
- `config_store.go` 根据 `storage_backend` 选择 memory 或 mongo 配置后端，并把配置同步回 `AppConfig`。
- `config_runtime.go` 在配置变更后同步 `AppConfig`，并触发 runner/channel reload。
- `runtime.go` 初始化 MongoDB、Redis 和 Langfuse plugin。
- `channels.go` 创建 ADK session/memory、runner、cron scheduler、system agent 和 channel manager。
- `grpc.go` 启动 daemon gRPC server，默认端口 `9090`。
- `cron.go` 创建 cron repository 和 scheduler。
- `system_agent.go` 注册内置系统 agent。

启动时先创建 HTTP/Twirp handler，再初始化配置仓库。配置仓库 seed 完成后，`StartChannels` 用当前配置构建 runner、cron 和渠道管理器。最后 `Handlers.Wire` 把 runner、session、cron、config runtime 等运行时依赖注入到已创建的 RPC/HTTP handler。

## Agent 构建模型

Agent 源配置来自 `agents.v1.Agent`：

- `AGENT_TYPE_LLM` 或未指定：构建 ADK `llmagent`。
- `AGENT_TYPE_LOOP`：构建 ADK loop workflow agent。
- `AGENT_TYPE_SEQUENTIAL`：构建 ADK sequential workflow agent。
- `AGENT_TYPE_PARALLEL`：构建 ADK parallel workflow agent。

构建流程：

```text
Agent proto
  -> resolve MCP server ids
  -> recursively build sub_agents
  -> resolve remote_agent_ids
      -> A2A remoteagent.NewA2A(...)
      -> DAEMON daemon.Bridge.BuildAgent(...)
  -> resolve model alias/provider
  -> build MCP toolsets
  -> create ADK agent
```

模型通过 `model_providers` 解析。Runner 支持运行时 model override：如果渠道选择了不同模型，`runner.Service` 会 clone proto 配置、替换 model，并缓存 override 后的 agent。

## Runner 执行流

所有入口最终调用 `runner.Service.Run(...)`：

```text
input parts + ContextInfo
  -> lookup agent
  -> optional model override
  -> get/create ADK runner by channel:agent:model
  -> ensure session exists
  -> run ADK runner
  -> collect final response text
  -> stream non-final events to callback
```

`ContextInfo` 提供 channel、session、user、source 和 uuid。Runner 使用 MongoDB session service 保持 ADK 上下文，使用 memory service 保存 ADK memory，并按 channel/agent/model 维度缓存 ADK runner。

当 agent 配置、MCP server 或 remote agent 发生变更时，`ConfigRuntime.ReloadRunner` 会重新构建 proto agent registry，并清空 runner 与 model override 缓存。

## Channel 执行流

Telegram 和 Discord 都由 `channel.Manager` 管理。Channel manager 从 `ChannelRepository` 加载渠道配置，启动对应平台 poller，并在配置变更时 reload。

典型消息流：

```text
platform update
  -> poller checks allowlist/triggers
  -> derive session id by USER or CHAT scope
  -> read active agent/model from Redis or channel defaults
  -> convert text/photo to genai parts
  -> runner.Run(...)
  -> send reply/status/debug/clear response
```

Redis 主要保存用户或会话维度的活跃 agent/model 选择；MongoDB 保存长期 session/memory。

## HTTP 与 RPC

HTTP handler 位于 `internal/handler/http`：

- `GET /ping`：健康检查，不需要 Bearer token。
- `GET /status`：运行时状态，返回当前配置存储 backend 和配置集合数量。
- `GET /a2a/:agent_name/.well-known/agent.json`：A2A agent card。
- `POST /a2a/:agent_name`：A2A JSON-RPC task send。

Twirp server 位于 `internal/application`，挂载在 `/api`：

- `AgentService`：Agent 配置 CRUD。
- `MCPServerService`：共享 MCP server 配置 CRUD。
- `RemoteAgentService`：远程 agent 配置 CRUD。
- `ChannelService`：渠道配置 CRUD。
- `SessionService`：会话查询、回复和清理。
- `CronJobService`：定时任务配置和执行记录。

除 `/ping` 外，HTTP/Twirp 请求经过 `APITokenAuthMiddleware`，使用 `Authorization: Bearer <apiToken>` 鉴权。

## Daemon 执行面

Daemon agent 用于服务端无法主动访问执行端的场景。连接方向是 daemon client 主动连到 server：

```text
cmd/butter-daemon
  -> gRPC Connect(register DaemonInfo)
  -> wait for DaemonTask
  -> execute through shell/opencode executor
  -> send DaemonTaskUpdate

cmd/butter server
  -> daemon.GRPCHandler
  -> daemon.Registry
  -> daemon.Connection
  -> daemon.Bridge as ADK agent
```

配置中 `RemoteAgent.protocol = REMOTE_AGENT_PROTOCOL_DAEMON` 时，`internal/agent` 会创建 `daemon.Bridge`，并按 `daemon_capability` 从 `daemon.Registry` 查找在线连接。Bridge 把 ADK invocation 转成 `DaemonTask`，等待 daemon 回传 terminal update 后生成 ADK final event。

当前 daemon 执行仍是同步等待 terminal result 的模型。连接断开时，活跃任务会收到失败更新；取消上下文时会向 daemon 发送 `CancelTask`。

## 配置与热更新

配置来源分两层：

- `AppConfig`：Butterfly 从 YAML 加载的启动配置。
- `ConfigStore`：运行时配置仓库，可选 memory 或 mongo 后端。

启动时 `ConfigStore.InitFromConfig` 会根据 `storage_backend` 选择后端：

- `memory` 或空：直接用启动配置 seed 内存仓库。
- `mongo`：如果对应 collection 为空，则从启动配置 seed；否则以 MongoDB 中的配置为准。

RPC 修改配置后，service server 会调用 `ConfigRuntime`：

- Agent/MCP/RemoteAgent 变更触发 `ReloadRunner`，重新构建 agent registry 并 reload channels。
- Channel 变更触发 `ReloadChannels`。

## 持久化

默认数据库名为 `butter`，可通过 `mongo_db` 配置。MongoDB 负责：

- ADK sessions/events。
- ADK memories。
- 配置仓库：agents、mcp servers、remote agents、channels。
- Cron jobs 和 executions。

Redis 地址默认 `localhost:6379`。Redis 连接失败不会阻止服务启动，但渠道内 active agent/model 选择可能无法持久化。

## 关键约束

- `pkg/proto/agents/v1` 是生成代码，手动改动应在 `proto/agents/v1` 中完成后重新生成。
- `runner.Service.Run` 当前仍以同步返回最终文本为主，长时间 daemon 任务会占用调用链。
- MCP toolset 当前支持 streamable HTTP 和 SSE transport。
- A2A remote agent 需要 `url`；daemon remote agent 需要 `daemon_capability` 和在线 daemon 连接。
