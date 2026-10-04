# Butter 系统架构

更新时间：2026-07-14

## 概览

Butter 是基于 Butterfly 框架的 Agent 服务。系统把 HTTP/RPC/channel 输入统一转为 ADK agent 执行，并通过 MongoDB、Redis、运行时配置仓库和 daemon 长连接支撑会话、记忆、渠道状态、定时任务、远程执行、invocation 历史与运维面板。

核心能力：

- 多租户 Workspace：所有 Agent / Channel / MCP / Remote / Model Provider / Notify Group / Agent File / Forum / Cron / Automation / API Token / Invocation / Cron Execution / Automation Run 都归属于一个 workspace；客户端通过 `X-Workspace-ID` 选择工作区。
- Agent 配置化：从配置仓库（MongoDB）加载 `agents.v1.Agent`，构建 LLM、Loop、Sequential、Parallel、Workflow agent。YAML 不是 Agent 的配置入口。
- 多入口接入：Gin HTTP、ConnectRPC（同 endpoint 同时支持 Connect/gRPC-Web/gRPC）、Telegram、Cron、Automation schedule、A2A、daemon connector 与 opencode HTTP server 直连。
- 运行时热更新：Agent、MCP Server、Remote Agent、Channel 配置可通过 RPC 修改并触发 runner/channel reload；`AgentService.ReloadAgents` 公开触发。
- 多执行面：本地 ADK agent、A2A 远程 agent、daemon 反向连接 agent、opencode HTTP server 直连 agent；A2A、opencode HTTP、MCP 都提供 live probing。
- 持久化运行时：MongoDB 保存 ADK session/memory、配置、cron 执行记录、automation run/step-run、invocation 历史、API tokens、workspaces 与 workspace_members；Redis 保存渠道内活跃 agent/model 选择。
- 运维面板：`WorkspaceService` / `DashboardService` / `DaemonService` / `APITokenService` 暴露 workspace + 成员管理 / counts / health / activity feed / 桥诊断 / daemon 任务 / 多 token 管理。

## Workspace 多租户模型

所有配置实体（`Agent` / `TelegramChannel` / `TelegramDestination` / legacy `AgentChannel` / `MCPServer` / `RemoteAgent` / `ModelProvider` / `NotifyGroup` / `CronJob` / `Automation` / `APIToken`）、内容实体（`AgentFileSpace` / `AgentFile` / `ForumThread` / `ForumPost`）以及运行时记录（`Invocation` / `CronExecution` / `AutomationRun` / `AutomationStepRun`）都归属于一个 workspace。`Workspace` 自身和 `WorkspaceMember` 由 `WorkspaceService` 管理，持久化为 MongoDB 的 `workspaces` 和 `workspace_members` 集合（`(workspace_id, user_id)` 唯一索引）。

请求流：

1. 客户端登录后，`AuthService.Login` 返回该用户可访问的 workspace 列表（全局 `admin` 角色得到所有 workspace）。
2. 客户端选择一个 workspace，后续请求带 `X-Workspace-ID` 头。
3. `AuthMiddleware` 校验调用方是该 workspace 的成员（admin 旁路），把 workspace id 注入 `context.Context`。
4. RPC 服务通过 `internal/workspace.FromContext(ctx)` 取出 workspace id，下传到 repo CRUD 调用；写入实体时自动写回 `workspace_id` 字段。
5. API token 自身绑定到一个 workspace，认证成功后直接覆盖请求 workspace（忽略 header）。

启动时 `application.BootstrapDefaultWorkspace` 检查 `workspaces` 集合是否为空，若为空则自动创建 slug 为 `default` 的 workspace，并把现有所有用户加为 `owner`。

运行时（runner / channel manager / cron scheduler / automation scheduler）通过新增的 `*AcrossWorkspaces` repo 接口拉平所有 workspace 的配置，构建一个全局视图。**在当前阶段 agent 运行时名字仍要求跨 workspace 全局唯一**，但所有 agent 消费方（interactive RPC、channel、cron、automation、forum、A2A、OpenAI 兼容 API）已迁移到不可变的、workspace 内唯一的 **Agent ID**（`agent_id`）作为**唯一引用**（agent_id-only）：调用/绑定 agent 时 `agent_id` 必填，legacy `agent_name` 不再作为输入，仅保留为服务端回写的显示名与历史记录字段。channel、cron job 与 automation step 通过 `ContextInfo.workspace_id` 把所属 workspace 透传到执行链。

## 进程入口

```text
cmd/butter/main.go
  -> core.New(...)
  -> SetupRoutes(cfg)
  -> SeedConfig(ctx, cfg)
  -> StartChannels(...)
  -> Wire(bootstrap result)
```

`cmd/butter` 是主服务进程，负责启动 Butterfly HTTP 服务；dashboard API 和 daemon connector 都挂载在同一个 `/api` ConnectRPC 入口下。

`cmd/butter-daemon` 是 daemon client 进程，携带 workspace-scoped daemon credential 主动连接服务端 `DaemonConnectorService.Connect`，注册自身能力，接收任务并通过本地 executor 执行。coding agent 类工具优先通过通用 ACP executor 接入，例如 `opencode acp`；shell executor 仍作为独立能力保留。

## 分层结构

```text
Access Layer
├── Gin HTTP handlers: /ping, /a2a/:agent_ref/..., /api/v1/* (OpenAI 兼容)
├── ConnectRPC: /api/agents.v1.*Service/*    # dashboard API + daemon connector
├── Telegram poller
├── Cron scheduler
└── Automation scheduler

Application / Transport Services
├── internal/application/*ServiceServer
├── internal/handler/http
└── internal/app routes/grpc/bootstrap wiring

Runtime Layer
├── runner.Service (+ InvocationRecorder, CancelInvocation, workflow_resume)
├── interrupt (Pending / Resume — single pending-Interrupt derivation seam)
├── cron.Scheduler (+ RunJobNow, ListByTimeRange, WAITING_INPUT)
├── automation.Engine + automation.Scheduler (+ run/step-run repositories)
├── daemon.Registry / Connection / Bridge / GRPCHandler / Metrics
├── session/mongo (+ CountSessions)
└── memory/mongo

Agent Layer
├── internal/agent.NewFromProto()
├── internal/agent.ProbeMCPServer()        # live MCP handshake
├── internal/agent/workflow.go             # graph validation + workflowagent construction
├── internal/agent/workflow_router.go      # Router node (route label matching)
├── internal/agent/workflow_human_input.go # Human Input node (Interrupt)
├── model provider resolution
├── MCP toolset construction
├── A2A remote agent resolution
├── daemon remote agent bridge
└── built-in system agent

Config Layer
├── AppConfig loaded by Butterfly
├── ConfigStore runtime backend wrapper
├── repo/config interfaces                # workspace-scoped CRUD + AcrossWorkspaces listings
├── repo/config/{memory,mongo}
├── repo/apitoken/{memory,mongo}          # api_tokens collection (workspace-scoped)
├── repo/invocation/{memory,mongo}        # invocations collection
├── repo/skill/{memory,mongo}             # skills + skill_resources metadata; ContentStore (memory/S3) for bodies
└── repo/workspace/{memory,mongo}         # workspaces + workspace_members

Workspace Layer
├── internal/workspace                    # ctx propagation (FromContext / WithID / HeaderName)
├── handler/http.AuthMiddleware           # resolves X-Workspace-ID, validates membership
└── application.WorkspaceServiceServer    # CRUD + memberships

Persistence
├── MongoDB: session, memory, config, cron, automation, invocations, api_tokens, workspaces, workspace_members, agent files, forum
└── Redis: channel active agent/model selection
```

## 启动装配

`internal/app` 是服务装配层：

- `routes.go` 创建 Gin handler、ConnectRPC handler 和 `Handlers` 容器。每个 `internal/application/*ServiceServer` 直接实现 `agentsv1connect.XxxServiceHandler`，由 `agentsv1connect.NewXxxServiceHandler(svc, connectOpts...)` 挂载在 `/api/agents.v1.XxxService/*`（`http.StripPrefix("/api", handler)`）。共享 codec/transport 见 `internal/transport/connectx`（含 snake_case JSON 兜底；dashboard 浏览器走 binary protobuf）。
- `config_store.go` 根据 `storage_backend` 选择 memory 或 mongo 配置后端，并把配置同步回 `AppConfig`。
- `config_runtime.go` 在配置变更后同步 `AppConfig`，并触发 runner/channel reload。
- `runtime.go` 初始化 MongoDB、Redis 和 Langfuse plugin。
- `channels.go` 创建 ADK session/memory、runner、cron scheduler、automation engine/scheduler、system agent 和 channel manager。
- `cron.go` 创建 cron repository 和 scheduler。
- `invocations.go` 发布本进程的存活键，并启动 Invocation 遗留清理（见下文"Invocation 的 owner 与遗留清理"）。
- `automation` runtime 创建 MongoDB-backed definition/run/step-run repositories，`Engine` 负责手动/调度执行与 step lifecycle（step 输入支持 `{{ selector }}` 模板插值，见 `template.go`），`Scheduler` 负责注册 enabled schedule-triggered automations。多 Pod 语义由 `internal/redislease` 承载：scheduler leader lease（`butter:automation:lease:scheduler`）保证一个 schedule 只由一个 Pod 触发；每个 automation 的 run lease（`butter:automation:lease:run:*`，`redislease.Guard`，续租 TTL/3、丢锁即取消 run context）把 SKIP/QUEUE 并发策略扩展到跨实例（REPLACE 跨实例退化为 QUEUE）。`RunAutomationNow` 异步执行：同步落 RUNNING 记录后在 engine base context 上后台执行。启动时 `ReconcileStaleRuns` 把超过 `StaleRunAge`（24h）仍 RUNNING 的 run 标记为 FAILED；完成的 run/step-run 由 `finished_at` TTL 索引保留 30 天。
- `system_agent.go` 注册内置系统 agent。

启动时先创建 HTTP/ConnectRPC handler，再初始化配置仓库。配置仓库 seed 完成后，`StartChannels` 用当前配置构建 runner、cron 和渠道管理器。最后 `Handlers.Wire` 把 runner、session、cron、config runtime 等运行时依赖注入到已创建的 RPC/HTTP handler。

**Invocation 的 owner 与遗留清理（#390，ADR-0016 决策 3）**：每个进程启动时生成一个 instance ID，在 Redis 写入存活键 `butter:instance:{id}`（TTL 30s，`liveness.Keep` 每 10s 续期，直到进程退出），并且先于任何 Invocation 记录的写入完成发布。invocation 仓库（`WithOwner`）在**创建**每条记录时把这个 ID 记为 owner；之后任何进程的保存（终态、redact）都保留它。owner 只存在于 Mongo 文档的顶层 `owner` 字段，不进 `Invocation` proto，不会出现在 API 响应里。`invocation.StaleSweeper` 在启动时运行一次，此后每分钟运行一次：只有当 owner 的存活键已消失时，才把它的 `QUEUED`/`RUNNING` 记录标为 `FAILED`，原因里写明丢失的 instance；另一个 Pod 启动本身不会让任何记录失败，清理进程也从不把自己的记录判为遗留。本变更之前写入、没有 owner 的记录，只有超过 `LegacyStaleAge`（24h；`chat_async.max_run_duration` 更长时取后者）才会失败。读不到存活状态时这一轮什么都不改；失败操作只改与读取时完全一致的记录，读取之后又被保存过的记录留给下一轮判断。清理只改记录，从不重放 Agent。没有 Redis 时只有一个进程，内存版 registry 只含本进程：启动清理把所有不属于本进程的 `QUEUED`/`RUNNING` 记录标为 `FAILED`，与之前一致，之后不再周期清理。

启动时还会执行只读的 Agent-ID cutover 校验（`application.RunAgentIDCutoverVerifier`，替代已退役的一次性回填 `backfillConsumerAgentIDs`，ADR-0010）：逐条检查所有 workspace 的 agent（agent_id 缺失/非法/重复、内联 `sub_agents`、legacy workflow 名字引用、`child_agent_ids`/workflow 引用不可解析、`MIGRATION_REQUIRED`、运行时名字冲突）与 consumer 记录（channel / cron / automation / forum 是否携带 `agent_id`），违规仅记 warning 日志、不自动修补、不阻断启动；`VerifyAgentIDCutover` RPC（全局管理员）提供同一诊断的按需版本。

## Agent 构建模型

Agent 源配置来自 `agents.v1.Agent`：

- `AGENT_TYPE_LLM` 或未指定：构建 ADK `llmagent`。
- `AGENT_TYPE_LOOP`：构建 ADK loop workflow agent。
- `AGENT_TYPE_SEQUENTIAL`：构建 ADK sequential workflow agent。
- `AGENT_TYPE_PARALLEL`：构建 ADK parallel workflow agent。
- `AGENT_TYPE_WORKFLOW`：构建 ADK 图 workflow agent（`workflowagent.New`），含 butter 自有的 Router、Human Input 和 Parallel Worker 节点。

构建流程：

```text
Agent proto
  -> resolve MCP server ids
  -> recursively build sub_agents
  -> resolve remote_agent_ids
      -> A2A remoteagent.NewA2A(...)
      -> DAEMON daemon.Bridge.BuildAgent(...)
      -> OPENCODE_HTTP opencode.Bridge.BuildAgent(...)
  -> resolve model alias/provider
  -> build MCP toolsets
  -> switch on type:
      -> LLM/UNSPECIFIED: llmagent.New(...)
      -> LOOP/SEQUENTIAL/PARALLEL: loopagent/sequentialagent/parallelagent.New(...)
      -> WORKFLOW: newWorkflowAgent(proto, subAgents)
          -> validate graph (nodes, edges, references)
          -> build ADK nodes: AgentNode, routerNode, humanInputNode, parallelWorkerNode, JoinNode
          -> map edges with Route / Default / unconditional
          -> workflowagent.New(config)
```

模型通过 `model_providers` 解析。alias-first 解析在 Agent 构建阶段完成，OpenAI-compatible 和 Gemini LLM 的 `Name()` 都返回实际 provider Model ID；ADK 在每个 model callback 前把该值写入 `LLMRequest.Model`。Runner 支持运行时 model override：如果渠道选择了不同模型，`runner.Service` 会先把 alias 解析为实际 ID，再 clone proto 配置、替换 model，并缓存 override 后的 agent。

Effective Context Window 仍由 ContextGuard 执行全部 compaction、buffer、summary、retry、session state 和通知语义，Butter 不复制其策略。Runner 从配置构建一个 source-aware `configuredModelRegistry`：配置容量覆盖 Crush/catwalk；内置 catalog 的 last-ID-wins 索引只用于准确区分 `embedded` 与数值同为 128,000 的 unknown `fallback`，output-token metadata 继续直接委托 Crush。Agent Threshold override 仍通过 `contextguard.WithMaxTokens` 传入，因此保持最高优先级；没有 Agent override 时，ContextGuard 在 callback 时按 `LLMRequest.Model` 查询实际选中 Model 的容量，model override 不会误用默认 Model 容量。

受管 LLM callback 的 plugin 顺序固定为：

```text
base plugins
  -> effective_context_window_logger (只读取 Agent name、strategy、LLMRequest.Model 与容量 metadata)
  -> ContextGuard (可能改写 request contents，并维护既有 state keys)
  -> compaction_notifier (观察 ContextGuard 写入的既有 state，不改变 callback contract)
```

logger 与 ContextGuard 都按现有 ADK runtime Agent name 查 policy；default 与 model-overridden Agent 保持同名，因此共享既有 ContextGuard state namespace。logger 不读取或附加 contents、summary、state value、tool/payload。ContextGuard 注册时捕获的默认 LLM 继续承担 summary call；#324 只让 callback-time selected Model 决定 Effective Context Window，不改变 summarizer、session state 或 notifier 语义。

同一 `(channel/app, user, session)` 的 turn 在 Runner 层串行执行，避免 Telegram、RPC 或 cron 同时写入一个 ADK session；不同 session 仍可并行。每个 turn 返回结构化诊断（event count、finish reason、error code），无可见文本时优先渲染 workflow `Event.Output`。Mongo session store 使用完整 `event_json` 保存 ADK event，并兼容读取旧的 `content_json` 文档，确保 workflow Output、Routes、Human Input 和终止元数据在重启后保留。ADK v2.2 起 event 的 JSON 键改为 camelCase（如 `invocationId`），此前写入的文档仍是 Go 字段名（如 `InvocationID`）；`encoding/json` 解码不区分大小写，两种文档都能读回，`event_roundtrip_test.go` 用 v2.1.0 的真实编码覆盖了这一点。

ADK 依赖已升级到 v2.5.0。OpenAI provider 仍不切换到 ADK 原生的 `openaimodel`（实验性，v2.5.0 起同时支持 Responses 与 Chat Completions）：它仍未把 `genai.InlineData` / `FileData` 转换为图片或文件输入。生产路径继续使用现有 Chat Completions adapter。v2.1.0 时的评估见 `docs/research/adk-go-v2.1-openai.md`。

**V2 create 契约（Agent ID）：** `AgentService.CreateAgent` 现在**要求** `agent_id`——不可变、workspace 内唯一的 slug（`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$`，保留字 `user`/`system`/`admin`/`start`/`default`/`api`/`new`，校验见 `internal/agent/agentid.go`），并**拒绝**内联 `sub_agents`；新 agent 以 `lifecycle_status: ACTIVE` 创建，子 agent 通过 `child_agent_ids` 引用独立记录（`ValidateAgentRelationships` 对 workspace 内 agent 池校验）。`UpdateAgent` 以 `agent.agent_id` 定位记录（必填；`name` 由服务端保留、从不选取记录），并拒绝修改内联 `sub_agents`（未变更的历史记录可原样往返，但构建时从不消费）。Workflow AGENT 节点必须用 `agent_id` 引用（legacy `agent` 名字字段已弃用、从不解析）。持久层以 `(workspace_id, agent_id)` 为逻辑主键，历史 Mongo `_id` 原样保留为不透明物理标识（ADR-0010）。迁移期 RPC `AssignAgentID` / `GetMigrationReadiness` / `MigrateAgentsV2` 已退役；`VerifyAgentIDCutover` 提供只读 cutover 校验。

## Runner 执行流

所有入口最终调用 `runner.Service.Run(...)`。入口在调用前先把请求携带的 agent 引用解析成运行时名字：

```text
agent 引用解析（internal/runtime/runner/resolve.go: ResolveAgentRef(workspaceID, agentID)）
  仅按 agent_id 解析，没有 legacy name 解析路径：
    空 agentID -> 直接 miss
    按 workspace-scoped agent_id 索引查找已注册 agent
      命中 -> 返回运行时名字
      set-but-unknown -> 直接 NotFound（不会误命中同名 agent）
  内置 agent（无 proto 配置，如 system agent）name 即 agent_id，
    且无 workspace 绑定，仅在 system context（空 workspaceID）下解析
  空 workspaceID 为 admin/system 路径，匹配任意 workspace
```

`resolveAgentRunnerRef`（application 层）把这一 seam 暴露给 interactive RPC（`InvokeAgent` / `StreamAgent` / `ReplySession`）：`agent_id` 缺失 -> `InvalidArgument`，未知 -> `NotFound`。channel、cron scheduler（`internal/runtime/cron/scheduler.go`）与 automation engine（`internal/runtime/automation/engine.go`）在执行时同样只按 `agent_id` 调用 `ResolveAgentRef`，并在写回配置/执行记录时把解析出的运行时名字 stamp 进 `agent_name`（仅显示用）、记录 `agent_display_name` 快照。跨 workspace 隔离仍在 ConnectRPC 边界与 runner 内共同强制。

解析出运行时名字后：

```text
input parts + ContextInfo
  -> lookup agent
  -> optional model override
  -> get/create ADK runner by channel:agent:model
  -> ensure session exists
  -> implicit workflow resume check (ADR 0002), via interrupt.Resume:
      if agent tree contains WORKFLOW type
      AND session has pending Interrupts (interrupt.Pending)
      AND input is plain text:
        rewrap as FunctionResponse targeting oldest pending Interrupt
  -> Memory Recall (memoryhook.Begin, ADR-0013), when the root agent's
     config.memory is enabled: search both scopes once, carry the
     <memories> block + memory scope in ctx
  -> run ADK runner
      (memory injection plugin appends the block to every model call's
       system instruction; memory tools read the scope from ctx)
  -> collect final response text; record the ADK invocation ID
  -> detect Human Input questions in session events
      -> append question text to output
  -> re-scan session for remaining pending Interrupts (interrupt.Pending)
  -> Memory Capture (memoryhook.Finish): background goroutine sends the
     turn's user input + final replies to Workspace Memory
  -> notify TurnListeners (cron uses this for WAITING_INPUT)
  -> stream non-final events to callback
```

`ContextInfo` 提供 channel、session、user、source 和 uuid。Runner 使用 MongoDB session service 保持 ADK 上下文，使用 mem0 支持的 memory service（`internal/runtime/mem0memory`，ADR-0013）读写各 workspace 的 Workspace/Agent Memory，并按 channel/agent/model 维度缓存 ADK runner。

**长期记忆**（ADR-0013）：记忆由各 workspace 自托管的 mem0 OSS 服务端保存。butter 一侧分四层：

- `internal/mem0`：最小的 REST 客户端，只实现 `/memories` 和 `/search`，认证用 `X-API-Key`。
- `internal/runtime/memoryconn`：每次调用都重新读取 `WorkspaceMemoryConfig` 并解密 key，不做缓存。
- `internal/runtime/mem0memory`：实现 ADK `memory.Service`，替换了原来的 Mongo 实现。在此之上提供 `Search` / `Recall` / `Add` / `CaptureTurn` 扩展方法，并负责 scope 和 ID 编码。
- `internal/runtime/memoryhook`：在每轮 runner 前后接入召回和写入。

各部分的具体做法：

- **根 agent 决定是否启用。** runtime 里 agent 以名字注册，runner 按名字找到根 agent 的 proto，从中取出 Agent ID 和 `config.memory`。
- **召回在 runner 里做。** 召回放在 ADK run 之前，不放在插件的 `BeforeRun` 回调里，因为 LLM 根 agent 会走 ADK 的另一条节点运行时（`runNode`）。召回结果经 ctx 传到 BeforeModel 插件，这条路径在两种运行时下都能到达。
- **插件排在最前面。** 注入插件注册在 Langfuse 和 ContextGuard 之前，这样 Langfuse 记录、ContextGuard 计数看到的都是注入后的请求。
- **写入的内容范围。** workflow 恢复时 ADK 会沿用暂停前那一轮的 invocation ID，恢复时用户的回复又被重包成 FunctionResponse。所以写入的内容是"运行前发出的用户输入"，加上"运行前 session 事件数之后、属于本轮 invocation 的最终回复"。
- **记忆工具集。** `internal/memorytool` 的工具集在构建 agent 时挂到开了 `enable_tools` 的 LLM agent 上，但每次模型调用时由 `Tools(ctx)` 判断是否提供：只有这个 agent 是本次调用的根 agent、并且 workspace 配置已启用时才提供。
- **失败处理。** mem0 出任何问题都降级为"没有记忆"。召回和写入的结果写日志，同时生成 OTel span（`memory.recall` / `memory.capture`）。

**Workflow 暂停/恢复**（`internal/runtime/interrupt`，单一派生 seam）：pending Interrupt 从 session events 派生（`interrupt.Pending` 扫描 `adk_request_input` FunctionCall/FunctionResponse 对，FIFO 最老优先），不额外存储。当 session 有未回答的 Interrupt 且新消息为纯文本时，`interrupt.Resume` 隐式将文本重包为最老 Interrupt 的 FunctionResponse，workflow engine 在该 session 上恢复；`runner/workflow_resume.go` 只负责把隐式恢复限定在含 Workflow 的 agent 上。已携带 FunctionResponse 的精确地址回复直接透传。cron 的 WAITING_INPUT 判定通过 `TurnResult.Pending`（同一 seam 产出）消费，不自行扫描 events。删除 session（`ClearSession`）可放弃暂停中的 workflow。

当 Agent、Model Provider、MCP server 或 remote agent 配置发生变更时，`ConfigRuntime.ReloadRunner` 会同步扁平配置，重新构建 proto agent registry、source-aware ModelRegistry 与 plugin chain，然后在一次锁内推进单调 runtime generation、替换 registry 并清空 runner / model override cache。overridden Agent 与 ADK runner 在锁外构建时会携带 generation snapshot，发布前再次校验；若 reload 已推进 generation，旧 build 被丢弃并从新 snapshot 重试，避免旧 Agent 搭配新 plugin 回填 cache。`sessionSvc` 不参与 swap，因此 reload 只影响后续 model call，不删除或迁移 history，也不改变 ContextGuard state key。

## 异步 Invocation 与只读观察流（issue #243 系列）

Dashboard chat 的文本轮次通过 `AgentService.SubmitAgentInvocation` 异步执行：请求在输入持久化（`QUEUED` Invocation 落库）后立即返回，`internal/runtime/asyncrun.Coordinator` 在后台 goroutine 中驱动执行（`QUEUED → RUNNING → SUCCEEDED/FAILED/CANCELLED`），生命周期与浏览器连接完全解耦。执行复用共享编排 `streamorch.Run`（与 `StreamAgent` 同一条路径），不引入第二套执行实现。提交要求客户端幂等 `request_id`；同一 session 同时只允许一个活跃 Invocation。首个版本为单实例：执行、取消和观察者都在进程内。遗留的 `QUEUED/RUNNING` 记录由 invocation 层的遗留清理（`invocation.StaleSweeper`）在其 owner 进程退出后标记为 `FAILED`，不自动重放。

**观察与执行分离**（`WatchAgentInvocation`，只读 server stream）：

```text
Coordinator.execute
  -> streamorch.Run(sink = hubSink)
       TextDelta / RunEvent -> watchHub.publish（非阻塞 fan-out）
  -> RUNNING / 终态落库后 publishState（权威 Invocation 快照）
  -> 终态帧发布后 closeAll（观察者看到 terminal-then-close）

WatchAgentInvocation handler（internal/application/agent_watch.go）
  -> 先 Subscribe 再读快照（终态帧在落库后才发布，保证不丢终态）
  -> 首帧恒为权威 state 快照；已终态则单帧后干净关闭
  -> 转发 run_event / text_delta / state 帧，一个终态帧后结束
```

任意数量的授权观察者可同时 attach；断开任何/全部观察者不影响执行。每个观察者持有有界 channel（256 帧），发布侧永不阻塞——落后的观察者被摘除并以 `resource_exhausted` 断开，客户端回读持久化 session events 与权威 Invocation 状态后重新 attach（token 级增量不做持久重放，ADR 精神同 #243 PRD）。鉴权与 `GetAgentInvocation` 一致：workspace 隔离 + `dashboard-async` 来源的私有会话仅提交者本人可见（含 watch 帧），全局 admin 保留支持通道；`GetAgentInvocation` 亦支持按 `session_id` 查活跃 Invocation（重连路径），`latest` 参数返回会话最近一次 Invocation（reload 后内联渲染失败/停止用），`include_input_parts` 返回失败/取消 Invocation 保留的 Input Parts（显式重试恢复输入用）。前端 `chat-window.tsx` 以 submit + watch 渲染实时输出，不再依赖高频全量 session 轮询；显式 Stop 走 `CancelAgentInvocation`（终态 `CANCELLED`），导航/关页仅断开观察者。

**失败语义（诚实终态，首版单实例）**：async 执行为单实例进程内模型，部署不得多副本依赖 dashboard async chat。三类运维性失败均记录可行动的 `Invocation.error`，且**绝不**写入 Agent 署名的 session events：(1) **超时** —— 超过 `chat_async.max_run_duration`（默认 30 分钟）的运行被取消并记 `FAILED`（原因含配置时长）；(2) **优雅停机** —— `Coordinator.Shutdown`（`cmd/butter/main.go` TeardownFunc 接线，15 s 上限）取消进程内运行并等待各自持久化 `FAILED` 停机原因；(3) **进程退出** —— 运行它的进程退出后（存活键过期），遗留清理把它的 `QUEUED`/`RUNNING` 记录标为 `FAILED`，原因里写明丢失的 instance；只标记、绝不自动重放 Agent 或重复工具副作用。用户显式 Stop 恒为 `CANCELLED`（前端呈现为"已停止"而非失败），即使与停机竞争。重试始终显式：前端恢复原始输入（文本 + Input Parts，失败/取消时保留）供用户审阅编辑，重新发送使用全新 `request_id` 创建全新 Invocation，UI 明示可能重复外部工具副作用。

## Session 标题生成（LLM）

标题生成与鉴权分开：`SessionServiceServer.TitleSession` 只负责生成并存储标题，不做鉴权。三个入口都调用它：

- **Web Chat**：异步 Invocation 成功后，`asyncrun.Coordinator` 以 `context.Background()` 调用 `SessionServiceServer.AsyncTurnComplete`，后者直接调用 `TitleSession`（Invocation 提交时已鉴权）；标题生成不阻塞 Invocation 终态持久化。
- **AG-UI**：run 成功、且 thread 的 session 在 run 开始时没有标题，handler 就在后台调用 `TitleSession`（`internal/handler/http/agui_title.go`）。它用脱离请求的 context（`context.WithoutCancel`，30 秒超时），不持有 thread lease，响应和该 thread 的下一次 run 都不等它。失败的 run 不生成标题。
- **`GenerateSessionTitle` RPC**：先做 self-only 鉴权（非 admin 只能为自己的 session 生成），再调用 `TitleSession`。

实现位于 `internal/application/session_async_title.go`、`session_service.go` 与 `session_title_llm.go`。

```text
TitleSession
  -> load session + all events
  -> if effective title exists (first-class / legacy state["title"]): return generated=false
  -> titleGenerator.generate (when resolver + provider lister wired):
      derive agent name from events (first non-"user" author, skip tool-only)
      resolve workspace_id + agent type + model via runner.Service.GetAgentMeta (TitleModelResolver)
      ListModelProviders(workspace_id only) via config repo
      model_ref = chat_title_model || agent.config.model; each candidate must
        resolve (alias/name) inside the workspace provider list — otherwise it
        is skipped, so no call can use credentials outside the workspace
      skip LLM when: no agent, no workspace, non-LLM agent, no candidate resolves
      prompt input = first user message + first final assistant response
        (Event.IsFinalResponse; partials and tool-call text skipped)
      direct non-streaming LLM call (fixed prompt, 10s timeout, max 64 output tokens)
      normalize to single line, max 30 Unicode code points
  -> on LLM skip/failure: deriveAutoTitle (deterministic truncation / "Image chat")
  -> SetSessionTitleIfEmpty (Mongo CAS on adk_sessions.title)
```

**与 Runner 的边界：** 标题生成不走 `runner.Service.Run`，不执行 ADK agent、工具或 workflow；是一次独立的 `internalagent.ResolveModel` + `GenerateContent` 调用。

**装配：** `internal/app/routes.go` 从 YAML `chat_title_model` 调用 `SetChatTitleModel`，并把 `SessionServiceServer` 作为 `AGUISessionTitler` 注入 AG-UI handler；`channels.go` 把 `runner.Service` 作为 `TitleModelResolver`、config repo 作为 `WorkspaceModelProviderLister` 注入 `SessionServiceServer`。

**副作用约束：** 不追加 session events、不写 `invocations`、不碰 ADK memory、不更新 `last_update_time`。手动 `UpdateSessionTitle` 与并发 CAS 保证 manual-title-wins。

**日志：** `session_id`、`workspace_id`、`model_ref`、`outcome` / `fallback_reason`、`elapsed_ms`。

## Skills 存储与运行时

Skills 是 workspace 级共享的 agentskills.io 能力包，服务于 ADK `skilltoolset`（ADR 0004）。

**存储切分**：ADK 的 `SkillToolset.ProcessRequest` 在**每次 LLM 请求**都会调用 `ListFrontmatters` 把技能目录注入系统指令，因此纯 S3 方案会把 `ListObjectsV2` 加多次读放到热路径上。改为：上传时解析 frontmatter 与资源路径元数据、持久化到 MongoDB（每轮一次带索引的查询,无缓存、无跨实例失效问题），而 `SKILL.md` 正文与资源文件走 `skill.ContentStore`（S3,与 Agent Files 同模式;未配置 bucket 时回退内存实现）。

- `skills` 集合：元数据 + 内容 key,`(workspace_id, name)` 唯一索引,既保证每 workspace 内名字唯一,又服务 List 热路径。
- `skill_resources` 集合：资源的路径索引（path、size、content_type + 内容 key）,建 `(workspace_id, skill_name, path)` 索引;`ListSkillResources` 与 Source 的 `ListResources` 都只读这里,**不触发 S3 List**。
- 内容 key：`<key_prefix>/<workspace_id>/<skill_name>/SKILL.md`（正文）与 `<key_prefix>/<workspace_id>/<skill_name>/<resource_path>`（资源）。
- 路径安全：`skill.CleanResourcePath` 是自定义 Source 唯一的穿越防护（ADK `FileSystemSource` 的保护不适用于自定义 Source）,`path.Clean` 后必须落在 `references/`|`assets/`|`scripts/` 内,任何 `..` 段、反斜杠、绝对路径一律拒绝。
- 级联删除：`DeleteSkill` 先从路径索引收集资源内容 key(取 key 失败则中止,不产生半删),删元数据后先删内容再删索引(索引作为内容 key 的恢复记录),内容批量删除对齐 S3 `DeleteObjects` 的 1000-key 上限。

**Source adapter**（`internal/skilltool`）：`Source` 实现 ADK `skill.Source`,构造时绑定到一个 workspace 与一个 agent 的技能名白名单,每次调用实时查库——技能改动无需重建 agent,下一轮 LLM 即生效。白名单外或不存在的技能都返回 ADK 的 `ErrSkillNotFound` 哨兵(两者对调用方不可区分);资源方法用 `CleanResourcePath` 校验后映射到 ADK 的 `ErrInvalidResourcePath` / `ErrResourceNotFound` 哨兵。

**ToolsetFactory 装配**（`internal/runtime/runner`）：`newToolsetFactory` 在 agent 具备 workspace 绑定且 `config.skills` 非空时,为该 agent 构建一个绑定其白名单的 `Source` 并挂上 ADK 技能工具集,向模型暴露 `list_skills` / `load_skill` / `load_skill_resource`。无技能仓库或空技能列表时不挂载。

## Automation 执行流

`internal/runtime/automation.Engine` 是 workflow 层，位于触发入口和具体动作之间。`AutomationService.RunAutomationNow` 与 `automation.Scheduler` 都调用同一个 engine：

```text
Automation trigger
  -> create AutomationRun(RUNNING)
  -> evaluate conditions over trigger payload/context
  -> execute ordered steps
      -> INVOKE_AGENT: runner.RunTurnSSE(...)  // 返回 TurnResult，携带 Pending Interrupt
      -> CALL_WEBHOOK: HTTP request
      -> SEND_NOTIFY_GROUP: NotifyGroup sender
      -> CREATE_FORUM_POST: Forum repository/service path
  -> persist AutomationStepRun per executed step
  -> 若 INVOKE_AGENT 暂停于 Human Input Node：AutomationRun(WAITING_INPUT) + session 坐标
  -> 否则 finish AutomationRun(SUCCEEDED / FAILED / SKIPPED / CANCELLED)
```

Engine 使用 automation definition/run/step-run repositories 按 workspace 写入 MongoDB。Automation 级 policy 提供 timeout、retry、concurrency 与 output truncation 行为；step policy 可覆盖 automation policy。v1 是线性顺序执行，默认首个失败 step 会停止后续 steps。

### Human Input 暂停与恢复（issue #176，镜像 cron ADR-0003）

当某个 `INVOKE_AGENT` step 调用的 Workflow Agent 暂停在 Human Input Node 时，`RunTurnSSE` 返回的 `TurnResult.Pending` 非空。Engine 不把该 run 记为成功，而是写入 `AUTOMATION_RUN_STATUS_WAITING_INPUT`，并记录 per-run session 坐标（`automation:<name>` / `automation:<workspace>` / `automation:<run-id>`）与暂停 agent 的 `agent_id`（回复时凭该 `agent_id` 定位）。按 Option A，暂停即为该 pipeline 的终点——后续 steps 不再执行。

恢复走 runner 的 TurnListener（与 cron 相同的机制）：`automation.Engine.HandleTurn` 注册为 turn listener，人通过 `ReplySession`（或任意入口）向该 session 发消息，runner 的隐式 resume 完成 workflow；当某个 automation session 上的 turn 结束且无 pending Interrupt 时，Engine 通过 `RunRepo.ListWaitingBySession` 找到等待中的 run，将其 finalize 为 `SUCCEEDED`，把恢复后的输出写回暂停 step，并清理该 session。删除暂停 session（ADR-0002 的放弃语义）由 `HandleSessionDeleted` 监听，将 run 置为 `CANCELLED`。Automation 没有顶层 delivery/notify 配置，因此 node 的问题仅记录在暂停 step 的输出中。

## Telegram 执行流（issue #264）

Telegram 由两个资源描述：**Telegram Channel** 是一个 Bot 传输通道，**Telegram
Destination** 是它下面的一个精确地址（`chat_id` 加可选的
`message_thread_id`）。Channel 只持有凭据与生命周期，不持有地址；Destination
持有地址与入站策略，并且是 Cron / Notify Group 唯一引用的对象（ADR-0008）。

### 接收：两种传输，一条管道

Webhook 与 Long Polling 是两种**传输**，不是两条管道。两者都进入同一个
`Router`：同样的地址匹配、同样的策略快照、同样的 Redis Stream、同样的 worker。

```text
webhook callback  ─┐
                   ├─> Router ─> 策略快照 ─> Lua(去重 + XADD) ─> Redis Stream
getUpdates batch  ─┘                                                  │
                                                                      v
                                              consumer group ─> worker ─> Agent ─> 统一发送器
```

**Webhook**：`POST /api/telegram/webhook/{channel_id}` 在每个 Pod 上可达，绕过
workspace 鉴权，先用常量时间比较校验 per-Channel secret，再解析 body。HTTP
`200` 的含义被严格定义为「事件已持久化进 Redis Stream」——在此之前返回 200
会让 Telegram 认为已处理，而事件只存在于某个 Pod 的内存里。去重与入队由**一段
Lua 脚本**原子完成：分成两条命令会让两个 Pod 同时通过 SETNX 检查，或者留下一个
永远不会被消费的去重标记。

**Long Polling**：由 Redis 租约选出每个 Channel 唯一的消费者。offset 是对
Telegram 的**承诺**而不是游标——带 offset N 的请求即确认 N 以下全部已处理。因此
offset 只越过「已持久化」或「已明确忽略」的 update；入队失败则停在原地，Telegram
会重发，`(channel_id, update_id)` 幂等抑制重复。offset 提交由 Lua 脚本按租约
持有者围栏（fencing）：暂停后苏醒的旧 leader 不能确认新 leader 仍在处理的 update。

Webhook 与 Long Polling 互斥。reconciler 会为 `LONG_POLLING` Channel 删除残留的
webhook 注册——Telegram 在已注册 webhook 时对 `getUpdates` 返回 409。

### Redis 是基础设施，不是缓存

启用入站要求 Redis 已配置**持久化、足够存储、无驱逐**。Stream 承载的是尚未处理
的用户消息，被驱逐即等于丢失。这一点在启用前检查（preflight blocker），而不是在
第一条消息到达时才发现。

### Forum Topic 精确路由

匹配以 `(channel_id, chat_id, message_thread_id)` 精确进行；未配置的地址被忽略
（`/where` 除外）。**缺省的 thread id 是一个地址，不是通配符**：群的普通会话与
每个 Topic 是不同的 Destination，拥有各自的 session 与策略。所有出站消息——回复、
处理中占位、命令回执、debug、分段的后续片段——都带上该 thread id，因此响应不会
漏到群的 general。

### 会话隔离

session key 为 `tg:{channel}:{destination}:{subject}:{agent}`。四段都有理由：两个
Topic 不能共享历史；切换 Agent 不能继承另一个 Agent 的上下文（切回则恢复）；
`USER` 策略下每个用户各自独立。切换 Model 刻意**不**移动 session。

一个派生 session 内的处理由 Redis 租约串行化：同一会话的两条消息并发执行会交错
写入历史。租约是**按会话**的，不相关的会话保持并行。

### 处理记录与重试边界（ADR-0009）

每条被接受的 update 有一条 Mongo 处理记录，其状态回答唯一一个问题：**Agent 可能
已经跑过了吗？** 边界之前自动重试；边界之后（`FAILED_UNCERTAIN`）进入死信，等待
人工判断，绝不自动重跑——Butter 无法保证 Agent 工具的幂等性。

完整回复在投递**之前**持久化，因此投递失败的重试是「发送已有文本」，而不是「重新
运行 Agent」。记录、输出与分段状态 30 天 TTL 过期。

### 统一出站

所有出站都经过 `internal/telegramsend`，入参是 Destination ID：Cron 投递、Notify
Group、Dashboard 测试消息、Agent 回复。它统一做 Markdown → MarkdownV2 转换与纯文
本回退、遵守 Telegram 的 `retry_after`、按 rune 边界分段（优先段落/换行/词边界）、
把「处理中」占位编辑成第一段。唯一被允许的原始地址出口是
`Sender.SendRaw`，只服务于传输层的 `/where` 命令。

### 遗留 Channel

`channel.Manager` 不再启动任何传输。它只在启动与 reload 时报告数据库中残留的
遗留 `AgentChannel` 记录，让运维看见「哪些不会运行」。遗留记录**不会被自动迁移**：
通用 Channel 只有 chat 白名单而没有精确地址，推断不出应该变成哪个 Destination，
猜测会把回复发到没人选择的地方。

## Linear 执行流（ADR-0015）

Linear Agent Session 是一个入口，复用 Telegram 的持久接收契约（ADR-0008 /
ADR-0009），但为 Linear 的时限单独设计了 worker、会话协调与 stop。

```text
POST /api/linear/webhook/{app_id} ─> 验签 + 时间窗 ─> Lua(去重 + XADD) ─> butter:linear:events
                                                                              │
                          consumer group ─> worker(不因 turn 阻塞读取) ─> Orchestrator.Handle
                                                                              │
                  处理记录 claim ─> 准入 ─> 确认 thought ─> enqueue-or-acquire ─> runner ─> activity
```

### 接收

`POST /api/linear/webhook/{app_id}` 在每个 Pod 上可达、绕过 workspace 鉴权；
workspace 只从 App 读取。签名覆盖整个 body，所以先按 1 MiB 上限读 body，再常量
时间校验 `Linear-Signature`（HMAC-SHA256），然后才解析并检查 `webhookTimestamp`
的 ±60 秒窗口。`200` 仍被严格定义为「已持久化进 Redis Stream（或被明确忽略）」；
队列失败返回 `503` 让 Linear 重投。去重键是 `Linear-Delivery`，缺失时用 body 的
SHA-256。Stream 与 Telegram 分开（`butter:linear:events`），这样 Telegram 积压不会
把一个新 session 推过 Linear 的 10 秒时限；Stream 的通用机制（去重追加、claim、
心跳、ack、持久性检查）抽到 `internal/eventqueue` 共用。

### Worker

Telegram 的 worker 会等一批 claim 全部处理完再读下一批；Linear 不能这样——一个 turn
可能跑半小时，而排在后面的新 session 必须在 10 秒内得到确认。Linear worker 每个事件
一个 goroutine，按空闲槽位（默认 32）读取，从不因 turn 阻塞读取；忙的 Pod 把剩余
工作留给其他 Pod。

### 会话、追加与 stop

一个 Linear Agent Session 对应一个 Butter 会话：session
`linear:{app_id}:{agent_session_id}:{agent_id}`，ADK app 名固定为 `linear`（不能用
可改名的显示名），user 为 `linear:{app_id}:{organization_id}`——session 属于 issue 上
所有人，而不是其中某个人。发消息的 Linear 用户通过 `ContextInfo.metadata.principal`
进入 Memory Capture 的 provenance。

turn 由 `SessionCoordinator` 串行化，它有内存与 Redis 两个实现：
- **enqueue-or-acquire**：会话空闲就拿到租约（并带走上一个持有者留下的 backlog），
  否则把消息追加进该会话的 follow-up 列表，事件立即 ack 并回「Queued」。
- **release-or-drain**：没有排队消息就释放租约；有就一次取走全部并保留租约，合并成
  下一轮。两者各是一段 Lua，因此追加的消息既不会丢，也不会与当前轮并发。
- **stop**：`RequestStop` 不拿租约——它原子地清空 follow-up 列表，并在有持有者时写入
  绑定该持有者租约 token 的 stop 标记（10 分钟 TTL）、发布一次 pub/sub nudge。持有者
  先订阅再读标记，所以与 turn 启动赛跑的 stop 不会丢，也不会误伤下一个持有者。持有者
  取消 turn（Pi / Cursor 经各自 bridge 的 abort），turn 真正结束后才发确认。

持有者崩溃后，下一个拿到会话的一方会先扫描该会话残留的 `PROCESSING` 记录（标记
`FAILED_UNCERTAIN` 并在 session 里报「cut short」），再运行 backlog。

### 处理记录与回复

每次投递一条 `LinearProcessingRecord`（`(app_id, delivery_id)` 唯一、30 天 TTL），
语义同 ADR-0009：进入 `PROCESSING` 之前的失败自动重试（最多 3 次）；之后的失败
死信、绝不重跑；回复（response、elicitation 或无需 Agent 的 error）先持久化为
`READY_TO_DELIVER` 再发送，发送失败可重投或经 `ResendLinearReply` 补发而不重跑
Agent。记录写入由可续约的 claim 租约围栏。

工具进度由 runner 的事件回调驱动：函数调用映射成 action，最新优先、每 3 秒最多一条；
在最终 activity 之前关闭，因此之后不会再有进度。Workflow 停在 Human Input 节点时，
最老的 Interrupt 以 elicitation 发出；表单只有一个单选字段时带 `select` 信号与选项
（从 request-input 事件冻结的表单读取）。下一条消息经 runner 的 FIFO 隐式 resume
（ADR-0002）继续，不新增任何 pending 存储。

### 凭据

App 的 client secret / webhook secret 与 Installation 的 access / refresh token
都在凭据接缝后、由数据库主密钥加密，从不作为 proto 字段。token 每次调用解密，不缓存；
快过期（5 分钟内）时刷新。Linear 会轮换 refresh token，所以刷新在按 installation 的
Redis 租约下进行，并按 token revision 做 compare-and-swap；`invalid_grant` 把
installation 标为 `NEEDS_REINSTALL`，之后快速失败、不再联系 Linear。

## HTTP 与 RPC

HTTP handler 位于 `internal/handler/http`：

- `GET /ping`：健康检查，不需要 Bearer token。
- `GET /status`：运行时状态，返回当前配置存储 backend 和配置集合数量。
- `GET /a2a/:agent_ref/.well-known/agent.json`：A2A agent card；`:agent_ref` **即** agent_id（legacy name 匹配已移除），仅对有 agent_id 且 `enable_a2a: true` 的 agent 开放，agent card 广告 agent_id URL。
- `POST /a2a/:agent_ref`：A2A JSON-RPC task send。
- `GET /api/v1/models` / `POST /api/v1/chat/completions`：OpenAI 兼容 API，`model` 字段**即** agent_id（legacy name 查找已移除），仅对 `enable_openai_api: true` 的 agent 开放；`/v1/models` 只列出有 agent_id 的 agent。
- `POST /api/uploads/*`：头像与静态资源 multipart 上传（见 `docs/storage.md`）；不走 ConnectRPC。
- `ANY /api/workspaces/:workspace_id/mcp`：工作区范围的 MCP HTTP 端点，转发给工作区 MCP service。
- `GET /api/mcp/oauth/callback`：MCP OAuth2 授权码回调，由 `MCPServerService.CompleteMCPServerOAuthCallback` 处理后重定向。
- `POST /api/agui/:agent_id`：AG-UI 入口（登录用户可访问任何 runner 能运行的 Agent；API token 与 root token 需要 `enable_agui`），SSE 流式事件；`GET /api/agui/:agent_id/threads/:thread_id/ui`：A2UI UI 快照；`GET /api/agui/:agent_id/threads/:thread_id/messages`：thread 历史（见下文 “AG-UI 上的 A2UI”）；`POST …/threads/:thread_id/stop`：Stop；`GET …/threads/:thread_id/run`：attach，重放 Detached Run 的 Run Log（见下文 “AG-UI 的 Detached Run”）。

### AG-UI 的 Detached Run（issue #401，ADR-0016）

- **租约与检查**：每次运行都在打开 SSE 之前拿 thread 的 session lease（有 Redis 时跨 Pod，否则进程内），所有读 session 的检查（thread 归属 403、工具结果、shared state 基线、表单提交）都在这一次租约内完成。续租出错会重试，直到距上次成功续租满一个 TTL 才算丢失租约；只有这种失效或“已不是持有者”才取消运行（`sessionguard.ErrLeaseLost`）。
- **Run state**：每次运行（无论是否 detached）都在租约旁记录 run state（`internal/runtime/runstate`：`runId`、Invocation ID、运行前 session 的事件数、是否 detached）。`runstate.Keep` 只在运行持有租约期间按租约节奏续期，续期与删除都以 Invocation ID 隔离，且只延长仍属于本次运行的 state，绝不重建已过期或已结束的 state；运行丢失租约或 Pod 挂掉时，它随租约一起过期。thread 读取用它在运行起点处截断（#403）。Detached Run 结束时，其 run state 标记为 ended（`End` 带保留时长），与其 Run Log 保留同样久，之后的 attach 仍能找到这次运行；读取、续期与 Stop 都把它当作不在运行，同一 thread 的下一次运行的 `Begin` 会替换它。
- **Detached Run**：客户端带 `forwardedProps.butterRun = {"detach": true}` 时，租约在脱离请求的 context 上获取（上限 `agui.max_run_duration`，默认 30 分钟），由运行 goroutine 持有到终态落库、且终止事件写进其 Run Log 为止；没有 UI 绑定的 thread 在打开流之前以 400 拒绝。事件只写入 Run Log（见下一条），POST 响应只是第一个观察者，带 SSE 注释心跳；断开只摘掉这个观察者。
- **Run Log 与 attach（issue #404，决定 5、6、7）**：每个 Detached Run 有一条 Run Log（`internal/runtime/runlog`），按 thread 租约的方式以调用者与 thread 为键、再加上本次运行的 Invocation ID（`butter:agui:log:session:{user}:{session}:{invocation}`）：有 Redis 时是一条 Redis Stream，否则是进程内实现，两者共用一套契约测试。写入方 `aguiLogWriter`（`agui_runlog.go`）就是运行的 `aguiEmitter`：只编码并入队，由自己的 goroutine 在后台批量追加，所以存储永远拖不慢模型循环；同一条消息相邻的 `TEXT_MESSAGE_CONTENT` 增量会合并（每段最多压 50 ms）。每次追加都带上写入方已知的条目数，丢失应答后的重试不会重复写入，也绝不会让已过期或已删除的日志复活。日志上限为 50,000 条、8 MiB；写入方积压超过 8 MiB（存储跟不上）同样截断：日志以截断标记结束，运行照常进行。运行存活期间由写入方续期（Pod 挂掉后随 TTL 过期，与 run state 一致），结束后保留 5 分钟；运行最多等 10 秒让最后的事件写进日志，然后才释放 thread 租约。Detached Run 总以 `STATE_SNAPSHOT` 开头，重放可以自成一体。每个观察者都从 `RUN_STARTED` 重放日志并跟随到运行结束：POST 响应，以及 `GET /api/agui/:agent_id/threads/:thread_id/run`（`AttachRun`，与 thread 读取相同的检查；没有日志时答 204：thread 空闲、结束已超过保留期、运行未 opt-in 或 thread 已删除）。读取永不阻塞，等待在每个进程内经一条 `XREAD` 复用（`tailer`），观察者不各自占用 Redis 连接。无法从头跟到尾的观察者（日志被截断或已不在、run state 已结束或消失且宽限期内没等到结束事件）以 `CUSTOM butter.fallback` 标记结束（`reason`：`truncated` / `expired` / `lost`），客户端改为读取 thread。删除 thread 时连同 run state 一起删掉这次运行的日志，复用的 `threadId` 不会重放已删除的对话。
- **Invocation 记录**：Detached Run 自己写记录（`source = agui-detached`，`request_id` 为按 thread 限定的 `runId`，重复则 409 `run_exists`），runner 不再记录（`runner.WithoutInvocationRecording`）。状态 QUEUED → RUNNING → SUCCEEDED / FAILED / CANCELLED，终态优先级与 asyncrun 的 `claimTerminal` 相同；超时、优雅关闭（`AGUIHandler.Shutdown`；Butterfly 不会调用 `TeardownFunc`，所以 `cmd/butter` 在收到 SIGTERM/SIGINT 时自己执行 teardown 再退出）、丢失租约都记为 FAILED 并写明原因。sink 先扣住 `RUN_FINISHED`，等记录和 run state 落定后再发出。
- **默认运行**：不带 opt-in 的运行仍活在请求里，断开即取消，记录仍由 runner 写入。
- **Stop（issue #402，决定 3、4、7）**：`POST /api/agui/:agent_id/threads/:thread_id/stop`（`agui_stop.go`）做与 thread 读取相同的鉴权、工作区与绑定检查，从不拿租约：`runstate.Store.Stop` 一步之内在仍未认领终态的 Detached Run 的 run state 上写入绑定其租约 token 的标记，并在该 thread 的频道上发布 nudge（Redis 为 Lua 脚本 + `PUBLISH`，每个进程用一条 pattern 订阅分发给本进程的运行；进程内实现直接唤醒）。运行在 run state 存在之前就订阅 nudge，并在每次续期时检查标记，兜底丢失的 nudge；`sessionguard.Token` 暴露租约获取 token，所以标记永远到不了同一 thread 的后续运行。运行收到后取消本轮（Pi/Cursor 走 `AbortSession`），在同一 run state 上一步认领终态：认领前已接受的 Stop 一律记为 CANCELLED，`RUN_ERROR` 带 `code: "stopped"`；认领后的 Stop 什么也找不到（204）。`CancelAgentInvocation` 把 `agui-detached` 记录交给同一个 Stop；`DeleteSession` 删除 AG-UI thread 时先 Stop，再自己限时（10 秒）拿 thread 租约并在租约下删除、清掉 run state 和这次运行的 Run Log，拿不到（例如未 opt-in 的运行占着 thread）就以 `unavailable` 失败且什么都不删；这取代了进程内的 deleting 标记。
- **运行中的读取（#403）**：thread 历史与 UI 快照不拿 lease，也不等运行结束（`agui_read.go`）。运行中立即返回 `running: {runId, invocationId}`，并按 run state 记下的事件数截断：保留运行之前的事件和启动这次运行的那一轮（用户消息、Human Input 回答或工具结果），运行已写入的其余事件不返回；卡片按截断点之前事件的 state delta 还原，因为 session state 已含本次运行写入的卡片。读取不加锁：读 session 前后各读一次 run state，不一致就重读；没有运行时再比较该 session 最新的 Invocation 记录（运行在写第一个事件之前写记录：Detached Run 总会，runner 除非写入失败），以发现在读取期间开始又结束的运行。运行刚开始、还没写入第一轮时，读取稍等片刻（合计不到一秒）。没有运行时，若该 thread 最近一次运行为 FAILED 或 CANCELLED，历史附带取自 Invocation 记录的 `lastRun: {status, error, input}`；只认本 app、本调用者的记录。
- **Dashboard（#405，决定 1、4、7）**：AG-UI Chat 的每次运行都带 `butterRun: {"detach": true}`（`ButterAGUIAgent`）。离开页面、刷新、切换 thread、New thread 都只中止请求，也就是只摘掉这个观察者。Stop 先调用 stop 端点，得到 202 或 204 之后才在本地取消，以立即结束这条流；端点出错时不取消，提示后可重试。删除当前 thread 只在本地经 runtime 断开，停止交给 `DeleteSession`。带 `code: "stopped"` 的 `RUN_ERROR` 不作为失败提示。
- **Dashboard 打开运行中的 thread（#406、#407，决定 8）**：thread 历史或 UI 快照任一报告 `running` 时，history adapter 的 `load()` 返回截断后的对话并带 `unstable_resume`，runtime 于是在启动这次运行的那一轮下面放一条运行中的回复，并调用 adapter 的 `resume()`，即 `RunFollower.follow`（`front/src/features/agui-chat/run-follower.ts`）。
  - **实时重新接入（#407）**：只对 thread 历史报告的那次运行（`RunFollower.found`，历史以它的那一轮结尾，回复就挂在这一轮下面），`follow` 经 `attachAGUIRun` 读 Run Log（`GET …/threads/:thread_id/run`），把重放、随后实时到达的 AG-UI 事件折叠成这条回复的 `ChatModelRunResult` 快照（`run-fold.ts` 的 `RunFold`）；重放的 `RUN_STARTED` 若不是这次运行的 `runId`（这次运行已结束、另一次运行接着占用了 thread），就不流式显示。这条路径上 runtime 只接收快照、不解析事件，它的 `RunAggregator` 也不导出，所以 `RunFold` 按 dashboard 锁定的 `@assistant-ui/react-ag-ui` 0.0.63 的聚合器折叠：文字；工具调用及其结果；interrupt 结局为 `requires-action`/`interrupt`，Interrupt 放在 `metadata.custom.agui`；留下未答客户端工具调用的成功结局为 `requires-action`/`tool-calls`。`run-fold.test.ts` 从包内文件读出该聚合器，对同一串服务端形状的事件逐个快照对照。不做的只有三件：一次 run 只有一个 message ID，所以不拆成多条回复；不带 timing（重放的计时不是 run 的）；不投影工具审批（Butter 的 Interrupt 是 `human_input`，不是 `tool_call` 门）。runtime 只为自己流式的 run 应用的副作用，由折叠经 `RunEffectsBridge` 自己派发：`butter.a2ui` 交给 A2UI store（与 `useOwnedA2UIStore` 喂给它的入口相同，`A2UIStore.apply`），`STATE_SNAPSHOT`/`STATE_DELTA` 经 `useAgUiSetState` 交给共享状态，下一次 run 发回的就是它。run 的结束交给 `RunEnd.ended`：`RUN_ERROR` 经 `useLastRun` 的 `runError` 显示失败或停止提示（#408），非停止时还弹 toast；随后像本页自己的 run 一样收尾（刷新会话列表以显示标题）。提示引用的那一轮取自读取结尾的那一轮（`runningInput`，工具结果没有输入）；读取分不出回答问题与普通消息，所以回答问题的 run 会把回答当作这一轮，而刷新后取自 Invocation 记录的 `lastRun` 没有输入。和本页发起的 run 一样，跟随的 run 一开始就清除之前的失败或停止提示（`useLastRun` 的 `following`）。流式期间 composer 可用；Stop 先调 stop 端点，得到 202/204 后在本地取消、结束这条流，与本页发起的 run 相同。
  - **轮询后备（#406）**：接入返回 204 或失败、流以 `butter.fallback` 标记结束或中途断开、重放的是另一次运行，或者只有 UI 快照报告 `running`（运行在读取历史之后才开始，历史里没有它的那一轮）时，`follow` 调 `pollUntilEnded`：它带退避（1 秒起，翻倍到 5 秒）重读历史，直到不再有 `running`，再读 UI 快照，也没有 `running` 时用这次读取经 `runtime.thread.import` 替换整段对话（运行中的占位回复是 optimistic 消息，连同已流式显示进去的内容随之被移除）并用 `A2UIStore.restore` 替换卡片与表单，与打开 thread 时的读取走同一个 `restoreThread`。结束等待的那次历史读取也经 `useLastRun` 的 `reading`（`RunEnd.reading`）交给失败与停止提示，所以等到的 run 失败或被停止时，不必刷新就能看到提示（#408）。等待期间 runtime 处于运行中（显示 “Running…” 和 Stop），composer 通过 runtime 的 `isDisabled` 禁用；Stop 先调 stop 端点，再立即重读、等到历史显示运行结束，而不在本地取消。502/503/504 与网络错误在下一次读取重试，409 与其他失败结束等待并显示加载失败；打开 thread 时的读取不再因 409 重试。
  - 本页 Stop 在本地取消、删除 thread 或离开页面会中止 `resume()`：不再读取 Run Log 或 thread，也不发 Stop；页面随即经 `RunEnd.detached` 收尾（刷新会话列表），如同中止本页发起的 run 的请求之后。
- **失败与停止的 run（#408，决定 8）**：AG-UI Chat 在对话下方显示与 Chat 共用的 `RunNotice`（`front/src/components/chat/run-notice.tsx`），内容是 run 的结局和触发它的那一轮输入，并提供 Restore input。页面有三个来源：run 的 `RUN_ERROR`（`stopped` code 显示为停止；客户端自己中止请求时的 `abort` code 两者都不算）、本页 Stop 在本地结束 run、刷新后历史的 `lastRun`；之后开始的 run 清除它（`use-last-run.ts`）。实时的输入取自 run 请求里服务端会记录的那一轮（`runMessages` 的最后一条用户消息）；带 resolved `resume` 或工具结果的 run 没有自己的输入。刷新后，历史最后一条用户消息的文字按记录规则（文本片段以空格连接，截断到 4096 字节）等于 `lastRun.input` 时，从这条消息取完整文字和图片，否则只用记录的文字（`last-run.ts`）。

### AG-UI 上的 A2UI（issue #350，ADR-0014）

- **模块**：`internal/a2ui` 是 A2UI v0.9.1 层——`butter-basic-v1` catalog 与校验、结果卡片的批处理与 session state 记录（`butter:a2ui:card:<id>`，删除留 tombstone）、UI 绑定（`butter:a2ui:binding`）、Human Input 表单绑定（冻结在 request-input 事件的 `CustomMetadata`）与提交校验。`internal/a2uitool` 提供 `render_ui`，与 `aguitool` 一样挂在每个 LLM Agent 上，但只有 run context 带 `a2ui.Run` 时才出现。
- **协商与绑定**：handler 读 `forwardedProps.butterA2UI`；在拿到 session lease 之后，若 thread 还没有 session，就以调用者的 `{principal, workspace, agent_id, thread_id}` 绑定创建它。已有 session 绑定的是其他 Workspace 或 Agent 时，运行在打开流之前被拒绝（403）；历史 session 没有绑定时 A2UI 不生效，文字聊天照常。协商成功且绑定匹配才把 `a2ui.Run`（thread/run/message ID）放进 run context。
- **先持久化后发送**：ADK runner 先 `AppendEvent` 再 yield，`render_ui` 通过工具 state delta 写卡片记录；`aguiSink` 只从已存储事件推导 envelope（客户端已有状态 → session 当前状态的转换），并对带表单绑定的 request-input 事件生成表单 envelope，逐条以 `CUSTOM butter.a2ui` 发出。
- **表单提交**：`resume` 中带 `butterForm` 的条目在 lease 内由 `a2ui.Resolve` 校验（绑定、token、revision、Interrupt 仍 pending、字段规则），通过后替换为按配置顺序编码的 JSON 字符串 payload，其余走 ADR-0002 的普通文本回复；任何拒绝都在打开 SSE 前返回。运行结束时 sink 重读 session：已被回答的表单发 `/status = answered`，带 resume 的运行在 `RUN_FINISHED` 中列出所有仍待回答的 Interrupt。
- **快照**：`UISnapshot` 经 `readThread`（`agui_read.go`，不拿 lease）读 session，用 `a2ui.LiveCards` + `a2ui.PendingForms`（`interrupt.Pending` ∩ 表单绑定）重建，不运行 Agent，不新增集合。
- **历史**：`ThreadMessages`（`agui_history.go`）与快照共用 `readThread`（相同的鉴权与绑定，不拿 lease），按事件顺序把 session 还原成 AG-UI 消息：
  - 两次用户输入之间 Agent 产出的内容合成一条 assistant 消息，工具调用沿用 session 的 FunctionCall ID，结果作为 tool 消息跟在后面。
  - 思考、request-input 握手和 `render_ui` 调用与实时流一样隐藏；卡片按其 state delta 第一次出现的位置、待答表单按其 request-input 事件，定位到所属回答（`surfaces`）。
  - 已回答的 Human Input 还原为“问题 + 用户回答”，表单答案用 `a2ui.Form.ReadableAnswer` 格式化；仍待回答的按 `RUN_FINISHED` 的格式放进 `interrupts`。
  - 只保留有结果、或 session 仍在等客户端结果（`interrupt.PendingToolCalls`）的工具调用，避免客户端取消一个服务端不认的调用。
  - Dashboard 通过 assistant-ui 的 history adapter 加载（`front/src/features/agui-chat/history.ts`），把卡片和表单作为数据片段放回回答里，把待答 Interrupt 挂到最后一条回答上。

RPC 服务位于 `internal/application`，挂载在 `/api`，使用 ConnectRPC（同一 URL 兼容 Connect binary/protobuf、Connect JSON、gRPC-Web 和 gRPC）。Dashboard 浏览器默认 `application/proto`；JSON codec 仍输出 snake_case field names（`connectx.HandlerOptions`）。外部 App 接入约定和可复制示例见 `docs/api.md`。

配置 / 执行：

- `AgentService`：Agent 配置 CRUD（分页）+ `InvokeAgent` / `StreamAgent`（同步兼容）/ `SubmitAgentInvocation` / `GetAgentInvocation` / `WatchAgentInvocation` / `CancelAgentInvocation` / `ReloadAgents` / `GetAgentRuntimeStatus` / `ListAgentRuntimeStatuses` / `ListAgentInvocations`，Agent lifecycle 的 `UpdateAgentConfiguration` / `RestoreAgent` / `GetAgentOperation` / `ListAgentOperations` / `RetryAgentOperation`，以及只读 cutover 校验 RPC `VerifyAgentIDCutover`（迁移期 RPC 已退役，恒返回 `Unimplemented`）。dashboard async chat 的短提交事务在单实例内串行化，保证每个 Session 最多一个 QUEUED/RUNNING Invocation；不同 Session 的 runner 并发执行。Get/Cancel 同时校验 Workspace 与 private Session owner，显式 Stop 终态为 CANCELLED，导航和观察者断开只停止本地 observer，不影响服务端执行。interactive RPC 以 `agent_id` 为**唯一引用**（必填，未知直接 NotFound，不回退 name）；runtime-status 查询仅接受 `agent_id`（携带 legacy `names` 过滤会被拒绝）；invocation 查询以 `agent_id` 为主，仅历史记录过滤仍兼容 `agent_name` 快照。
- `MCPServerService`：共享 MCP server CRUD + `GetMCPServerStatus`（live probing）+ `ListMCPTools` + MCP OAuth2 流程（`StartMCPServerOAuth` / `CompleteMCPServerOAuth` / `GetMCPServerOAuthStatus` / `DisconnectMCPServerOAuth`）。
- `RemoteAgentService`：远程 agent CRUD + `GetRemoteAgentStatus`（A2A / Daemon / OpenCode HTTP live probing）。
- `ChannelService`：**已废弃**的 generic `AgentChannel` 兼容 API，仅保留读取、状态查看和删除；创建/更新/重启/暂停/恢复均返回 `Unimplemented`。当前 Telegram 使用 `TelegramChannelService`、`TelegramDestinationService`、`TelegramAdminService` 与 `TelegramProcessingService`。
- `SessionService`：`Create` / `Get`（含 duration + trace_url）/ `List`（filter + page）/ `Delete` / `Reply` / `UpdateSessionTitle` / `GenerateSessionTitle`。Title 存为 Butter-owned 元数据（`adk_sessions.title`），不经由 ADK state；有效标题优先级：first-class title → legacy `state["title"]` → agent name → shortened session ID。`GenerateSessionTitle` 在首轮对话后尝试 LLM 标题（可选 YAML `chat_title_model`，按 agent workspace 解析 provider），失败则确定性截断；见上文 Session 标题生成。重命名与自动生成均不影响 `last_update_time` 和排序。
- `AutomationService`：自动化工作流 CRUD + `RunAutomationNow` + run/step-run history 查询。
- `CronJobService`：定时任务 CRUD + `ListCronExecutions` + `RunCronJobNow`，含 timeout/retry/concurrency/notify/output reliability policy。
- `ModelProviderService`：LLM Provider CRUD。
- `NotifyGroupService`：通知组 CRUD，供 cron 投递使用。
- `AgentFileService`：workspace 范围的文件空间与文件 CRUD（含 `SearchAgentFiles`）。
- `ForumService`：论坛 thread / post 管理（`ListThreads` / `ListThreadLabels` / `GetThread` / `CreateThread` / `UpdateThread` / `DeleteThread` / `CreatePost` / `DeletePost` / `InvokeAgentInThread`）。

运维：

- `DashboardService`：`GetOverview` / `GetActivityFeed` / `GetCronExecutionTimeseries` / `GetActivityMetrics`。
- `DaemonService`：workspace daemon config CRUD、daemon credential 签发、`ListDaemons` / `GetDaemon` / `ListDaemonTasks` / `CancelDaemonTask` / `GetBridgeDiagnostics`。
- `APITokenService`：`ListAPITokens` / `CreateAPIToken` / `RevokeAPIToken`。
- `GlobalMCPServerService`：admin 管理的 workspace-agnostic MCP server 预设；`InstallGlobalMCPServer` 把预设克隆到目标 workspace（admin 可跨 workspace 安装，审计日志记录）。
- `WorkspaceService`：workspace 与成员 CRUD（无需 `X-Workspace-ID`）。
- `AuthService`：登录、OAuth 登录、当前用户、用户管理、资料更新与 `ChangePassword`。
- `GitHostService` / `WorkspaceRepoBindingService`：Git endpoint allowlist、workspace repository binding、Agent Content onboarding/sync/publish/commit/rollback 和安全 detach。

除 `/ping`、OPTIONS 预检、MCP OAuth 回调、Telegram webhook、daemon connector 方法以及 `AuthService.Login` / `ListOAuthProviders` / `BeginOAuthFlow` / `CompleteOAuthFlow` 之外，所有 HTTP/RPC 请求经过 `AuthMiddleware`。Telegram webhook 和 daemon connector 会在各自 handler 内执行专用鉴权：

1. 优先解析 `Authorization: Bearer <token>` 中的 user session（Redis `butter:auth:session:<sha256(token)>`，命中后异步 `TouchSession`）。
2. 不匹配则用 `subtle.ConstantTimeCompare` 比对配置的 root token (`cfg.apiToken`)。
3. 再不匹配则查 `apitoken.Repository.Lookup(sha256(token))`；只有 `API_TOKEN_KIND_USER` + `api:*` scope 的 token 可进入 HTTP/RPC API，命中后异步 `TouchLastUsed`。
4. 通过后解析 `X-Workspace-ID` 头：用户 session 走成员关系校验（admin 旁路）；API token 直接绑定到其存储的 workspace，覆盖 header；root token 与未配置 repo 时接受 header 原值。
5. 全部失败返回 `401 Unauthorized`。

`AuthService.Login` 返回该用户可见的 workspace 列表（admin 看全部），前端在登录后弹出 workspace 选择器，把选中的 workspace id 写入后续请求头。

Daemon gRPC `Connect` 不接受 root token 或普通 API token，只接受 `DaemonService.CreateDaemonRuntimeToken` 为已有 `DaemonRuntime` 签发的 daemon runtime token。

`AuthService` / `WorkspaceService` / `DashboardService` 不要求 `X-Workspace-ID`。`DaemonService` 的配置、credential 签发、在线 daemon / task 查询都按 workspace scoped 处理；daemon 原生 gRPC `Connect` 则由 daemon credential 自带 workspace。`SessionService` 的 session CRUD 按 `app_name` + `user_id` + `session_id` 查询，调用 `ReplySession` 时应带 `X-Workspace-ID`，让 runner 在正确 workspace 解析 agent。

## Daemon 执行面

Daemon agent 用于服务端无法主动访问执行端的场景。连接方向是 daemon client 主动连到 server：

```text
cmd/butter-daemon
  -> gRPC Connect(authorization: Bearer <daemon credential>)
  -> register DaemonInfo
  -> wait for DaemonTask
  -> execute through ACP/shell executor
  -> send DaemonTaskUpdate

cmd/butter server
  -> daemon.GRPCHandler
  -> daemon.Registry
  -> daemon.Connection
  -> daemon.Bridge as ADK agent
```

Daemon 是 workspace runtime 资源：先通过 `DaemonService.CreateDaemonRuntime` 写入 `config_daemon_runtimes`，再通过 `CreateDaemonRuntimeToken` 签发 `API_TOKEN_KIND_DAEMON` + `daemon:connect` token。`GRPCHandler` 在第一条 register 消息后校验 token，token 的 workspace 与 daemon_runtime_id 为 authoritative 值；daemon 自报 workspace 不一致会被拒绝，未声明 `acp_runtimes` 时默认支持 `opencode` 与 `codex`。

`daemon.Registry` 以 `workspace_id -> daemon_runtime_id -> connection` 管理连接，同一 workspace/runtime 同时只允许一个连接。配置中 `RemoteAgent.protocol = REMOTE_AGENT_PROTOCOL_DAEMON` 时，`internal/agent` 会创建带 workspace 的 `daemon.Bridge`，并按 `daemon_runtime_id` 查找在线连接、按 `acp_runtime` 选择 daemon 侧 ACP executor。Bridge 把 ADK invocation 转成带 `workspace_id` / `daemon_runtime_id` / `acp_runtime` 的 `DaemonTask`，等待 daemon 回传 terminal update 后生成 ADK final event。

当前 daemon 执行仍是同步等待 terminal result 的模型。连接断开时，活跃任务会收到失败更新；取消上下文时会向 daemon 发送 `CancelTask`。

## OpenCode HTTP 直连

`RemoteAgent.protocol = REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP` 是 daemon 执行面的一个对称替代：当 opencode 已经在某个可访问的地址上长期运行（`opencode serve --port 4096`）时，butter 通过其 HTTP API 直接调用，不需要拉起 daemon client。

`internal/runtime/opencode.Bridge` 把 ADK invocation 转成 `POST /session` + `POST /session/:id/message` 请求，从返回 message 的 `parts[*].text` 拼出 final event；取消上下文时调用 `POST /session/:id/abort`。鉴权使用 HTTP Basic Auth（`username` 默认填 `opencode`，`password` 取自 `OPENCODE_SERVER_PASSWORD` 一类配置），由 `RemoteAgent.username` / `password` 字段直接携带。`opencode_agent` 与 `opencode_model` 字段透传到 opencode server 用于覆盖默认 agent/model。

`RemoteAgentService.GetRemoteAgentStatus` 对该协议探测 `GET /global/health`：2xx 视为 ACTIVE、401 视为 ERROR、其它视为 UNREACHABLE。

当前实现使用同步 `/message`，不订阅 `/global/event` SSE 流，每次调用建立一个一次性 opencode session。

## 配置与热更新

配置来源分两层：

- `AppConfig`：Butterfly 从 YAML 加载的启动配置。
- `ConfigStore`：运行时配置仓库，可选 memory 或 mongo 后端。

启动时 `ConfigStore.InitFromConfig` 会根据 `storage_backend` 选择后端：

- `mongo` 或空：使用 MongoDB 仓库；直接读取 MongoDB 中的配置。YAML 中的 `agents` / `mcp_servers` / `remote_agents` / `channels` / `model_providers` 已不再作为 seed 源，应通过 RPC（带 `X-Workspace-ID`）写入。
- `memory`：使用内存仓库；初始为空，进程重启后数据丢失。

RPC 修改配置后，service server 从 `ctx` 取 workspace id 后写入对应 workspace；写完成调用 `ConfigRuntime`：

- Agent/MCP/RemoteAgent/ModelProvider 变更触发 `ReloadRunner`，跨 workspace 拉平所有配置后重新构建 agent registry 并 reload channels。
- Channel 变更触发 `ReloadChannels`。

## 持久化

默认数据库名为 `butter`，可通过 `mongo_db` 配置。MongoDB 负责：

- ADK sessions（`adk_sessions` / `adk_events`）。`session/mongo.Service.CountSessions` 给 dashboard overview 用。
- ADK memories：旧的 `adk_memories` 集合已不再使用，只保留不删除。长期记忆存在各 workspace 的 mem0 OSS 服务端上（ADR-0013）。
- `workspace_memory_configs`：每个 workspace 最多一份 mem0 连接配置，`_id` 为 workspace ID；API key 用数据库主密钥加密后存进单独的凭证列。
- 配置仓库：`config_agents` / `config_mcpservers` / `config_remoteagents` / `config_daemon_runtimes` / legacy `config_channels` / `config_modelproviders` / `config_notifygroups`，`_id` 为 `"{workspace_id}:{name}"` 或 `"{workspace_id}:{id}"` 复合键，并对 `(workspace_id, name)` 建索引。
- Agent Files：`agent_file_spaces` / `agent_files` / `agent_file_versions`。
- Skills：`skills`（元数据 + 内容 key，`(workspace_id, name)` 唯一索引）与 `skill_resources`（资源路径索引，`(workspace_id, skill_name, path)` 索引）；`SKILL.md` 正文与资源内容走 `skill.ContentStore`（S3，未配置 bucket 时回退内存）。
- Forum：`forum_threads` / `forum_posts`。
- `workspaces`：workspace 元数据，`slug` 唯一索引。
- `workspace_members`：用户与 workspace 的多对多关系，`(workspace_id, user_id)` 复合唯一索引、`user_id` 普通索引。
- `users`：dashboard 用户、bcrypt password hash 与全局角色。
- Cron jobs / executions（`cron_jobs` / `cron_executions`，`_id = "{workspace_id}:{name}"`；含 `ListByTimeRange` 支撑时序聚合，`workspace_id` 字段可作过滤）。
- Automations / runs / step runs（`automations` / `automation_runs` / `automation_step_runs`，definition 按 workspace + name 唯一；run 列表按 workspace + automation_name + started_at 倒序；step-run 按 workspace + run_id + order 查询）。
- `invocations`：runner 持久化的每次 ADK 调用（runner → `InvocationRecorder.Save`，RUNNING 起记，defer 写终态，附带 `workspace_id`，并记录 `agent_id` 与 `agent_display_name` 快照；历史记录只保留 `agent_name`）。驱动 ActivityFeed + AgentRuntimeStatus + ListAgentInvocations。
- `api_tokens`：DB-stored API tokens（带 `workspace_id` + `secret_hash` + `prefix` + `kind` + `scopes` + optional `expires_at` / `daemon_runtime_id` + `last_used_at` + `revoked`）。
- Telegram：`telegram_channels` / `telegram_destinations` / `telegram_processing_records`、平台设置和数据库主密钥集合；Bot credential 不进入 proto 响应。
- Repository binding：Git host、workspace binding、缓存、Agent Content snapshot 与 Agent lifecycle operation 集合。

后端选择：`storage_backend` 为空或等于 `"mongo"` 时全部走 mongo；显式设置为 `"memory"` 时用内存仓库（`api_tokens` / `invocations` 也支持 memory 实现，方便测试）。

Redis 地址默认 `localhost:6379`。Dashboard session 存在 Redis；Redis 不可用时登录/session 校验会失败。

## 运维面板与可观测性

- `DashboardService.GetOverview` 实时探活 MongoDB / Redis / Runner（带 latency），聚合所有 counts。
- `GetActivityFeed` 从 `invocations` 集合派生最近活动。
- `GetCronExecutionTimeseries` 用 `cron_executions.ListByTimeRange` + bucket 聚合（1D=1h / 7D=1d / 30D=1d）。
- Automation list/detail 页直接读取 `automations`、`automation_runs` 和 `automation_step_runs`，展示 workflow 配置、手动 run 和 step 级历史。
- `DaemonService.GetBridgeDiagnostics` 使用 `internal/runtime/daemon/metrics.go` 的 `Metrics` collector，记录每次 bridge 调用 latency 到 60 条 ring buffer，并按需读取 `runtime/metrics` 的 `/cpu/classes/total:cpu-seconds` 与 `runtime.MemStats.Sys` / `runtime.NumGoroutine()`。
- `SessionEvent.trace_url` 当 `cfg.Langfuse.Host` 设置时拼接 `<host>/trace/<invocation_id>`，前端 Session detail 一键跳 Langfuse。

## 前端 Dashboard 与镜像

- `front/`：Vite + React 19 + shadcn/ui + TanStack Query。TanStack Router 路由位于 `front/src/routes/`，资源实现位于 `front/src/features/`；一级页包含 Login、Chat、Forum、Dashboard、Agents、MCP Servers、Remote Agents、Daemons、Telegram Channels/Destinations、Sessions、Automations、API Tokens、Model Providers、Notify Groups、Agent Files、Workspaces、Users、Profile、Integrations、Admin 等。
- Proto TS 绑定通过 `buf.build/bufbuild/es`（`include_imports: true`）生成到 `front/src/gen/`，运行时类型走 `@bufbuild/protobuf`。
- 后端镜像：`ghcr.io/<owner>/<repo>`（根 `Dockerfile`，distroless static + cosign 签名）。
- 前端镜像：`ghcr.io/<owner>/<repo>-front`（`front/Dockerfile`，node:22-alpine 编译 + nginx:1.27-alpine 运行 + SPA fallback + `/healthz`）。
- CI workflows：`.github/workflows/docker-publish.yml`（后端，cron + push + PR）与 `front-publish.yml`（前端，`paths: front/**` 过滤），均带 cosign keyless 签名。

## 仓库绑定与 Agent Content 生命周期（issue #210 系列）

每个 workspace 可绑定零或一个 Git 仓库（`WorkspaceRepoBindingService`），把 Agent Content（`agents/{agent-id}/description.md`、`prompt.md`、可选 `global-prompt.md`）交给 Git 托管，而运行态仍由 DB 拥有的运营配置驱动。PAT 经 `internal/secretbox` 加密、走独立凭据 seam，绝不进入 proto/响应/日志（ADR-0005）。同步（`SyncWorkspaceRepository`）把仓库树读入 workspace 级 DB 缓存；发布（`publishActiveRevision`）先把校验通过的 Agent Content 按 `workspace_id + commit_sha` 持久化到 Mongo，再推进 binding 的 Active Revision 并 reload runner。读取和 runtime reload 始终用 `active_commit_sha` 精确选择快照，把 Content 叠加到 DB agent proto（`ConfigRuntime.applyActiveContent` → `agentcontent.ApplyToProto`）构建 Effective Agent；进程重启和多副本不会丢失或分叉 Effective Content。

**上线/下线（issue #219，ADR-0007）：**

- **上线** `OnboardWorkspaceRepository(mode)`：`EXPORT_CURRENT` 把 DB 现有 Content 作为单次校验提交写入 Git 再发布；`IMPORT_REPOSITORY` 同步+发布并只采纳匹配 Agent ID 的目录（未知目录保持 unclaimed，不做双向合并）。两种模式都在提交/导入、回读、内容校验（`agentcontent.Validate`）通过后才推进 `active_commit_sha` 切换为 Git 拥有；随后的 `ReloadRunner`（Effective Agent 构建）按 #216 约定为 best-effort，reload 失败时保留 last-known-good runner 而非回退 Active Revision。EXPORT 恒为 direct commit（即使绑定为 CHANGE_REQUEST），否则内容会挂在未合并的 PR/MR 上无法发布。
- **下线** `DeleteWorkspaceRepoBinding`：先把 Active Revision 快照物化回 DB 字段（`ApplyToProto` 的逆操作）并 reload runtime，再移除绑定/PAT/缓存/快照；**绝不改动远端 Git**。无有效快照时默认拒绝（`FailedPrecondition`），除非显式选择 `RepoBindingDetachRecovery`（`KEEP_DATABASE` 保留当前 DB 内容）。

绑定按 workspace 独立推进，互不牵连。分阶段上线顺序、迁移校验清单与跨 provider 验收见 `docs/repo-binding-migration.md`。

## 关键约束

- `pkg/proto/agents/v1` 是生成代码，手动改动应在 `proto/agents/v1` 中完成后重新生成。
- `runner.Service.Run` 当前仍以同步返回最终文本为主，长时间 daemon 任务会占用调用链。
- MCP toolset 当前支持 streamable HTTP 和 SSE transport。
- A2A remote agent 需要 `url`；daemon remote agent 需要 `daemon_runtime_id`、`acp_runtime` 和在线 daemon runtime 连接。
- 跨 workspace 共享同一个 runner / channel manager / cron scheduler / automation scheduler，**agent 运行时名字需在所有 workspace 内全局唯一**；引用 agent 的 channel、cron job 与 automation step 通过 `ResolveAgentRef` **仅按 `agent_id`** 解析（无 legacy name 解析路径）。启动时的 Agent-ID verifier 只读检查违规记录并记录 warning，不自动回填或修复；全局管理员可通过 `VerifyAgentIDCutover` 手动运行同一诊断。
- Automation v1 是线性有序 step，不包含 DAG、人工审批 gate 或持久化 worker queue；webhook/forum/channel/daemon event trigger 已在 proto 中预留，但当前运行时只执行 manual 和 schedule trigger。
- 内置 system agent 仍为全局注册，其管理类工具读跨 workspace、写则要求显式传入 `workspace_id`。
