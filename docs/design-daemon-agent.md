# Daemon Agent 设计方案

> **状态：历史设计文档（已落地，正文保留旧方案记录）。** 本文是 daemon agent 设计阶段的方案稿，
> 描述当时的 Twirp + Gin 架构与拟新增的 daemon 服务。daemon 已全部落地；
> 自 2026-06-02 起 RPC 层已迁移到 ConnectRPC，daemon 自身仍走独立 gRPC
> 端口 `:9090`。RPC 部分提到的 Twirp 文件 / Twirp
> 路径请参照 `docs/architecture.md` 与 `docs/api.md` 阅读现状。自
> 2026-06-05 起，daemon coding agent 执行器已从 opencode 专用 CLI 调用迁移为
> 通用 ACP executor（`github.com/coder/acp-go-sdk`），opencode 通过 `opencode acp`
> 接入；旧 `executors.opencode` 配置仅作为兼容入口保留。自 2026-06-06 起，
> 当前实现已使用 workspace-scoped `DaemonRuntime` 替代旧 `DaemonConfig` /
> `daemon_capability` 模型；正文中仍出现的旧名均为历史方案语境。

## 背景

Butter 当前已支持通过 ADK Go 封装的本地执行 agent（LLM、Loop、Sequential、Parallel），以及通过 A2A 协议调用的远程 agent。本文档描述支持第三种执行模式——**Daemon Agent**：Client 作为长驻进程与 Server 建立持久连接，Server 将任务下发给 Client，Client 通过 ACP executor 调用 opencode / codex 等本地 coding agent 并回传结果。

## DaemonRuntime 重构目标（2026-06-06，已落地）

这一节记录已经落地的产品与架构目标。它不是在旧 `DaemonConfig` 上小修小补，而是对 daemon 资源模型、RPC、仓储、前端页面和 daemon 启动方式做完整重构。

### 已确认需求

- `DaemonRuntime` 完全替代旧 `DaemonConfig`。旧的 `allowed_capabilities` / `daemon_capability` 模型已经废弃，不要求保持内部结构兼容。
- Workspace 可以创建 `DaemonRuntime`。创建后服务端签发 daemon token，并展示一条最小启动命令：

```bash
butter-daemon --url <daemon-grpc-url> --token <daemon-runtime-token>
```

- daemon token 绑定 `workspace_id + daemon_runtime_id`。`cmd/butter-daemon` 启动时只需要 `url + token` 即可连接；服务端从 token 推导 authoritative workspace/runtime，不信任客户端自报。token 生命周期暂时与普通 API token 一致：默认可长期使用，后续如需 rotate/revoke/expiry 再扩展。
- `cmd/butter-daemon` 连接成功后代表一个 runtime 执行面，而不是一个 agent。它负责接收任务、启动对应 ACP runtime、回传 `DaemonTaskUpdate`。
- 当 workspace 有 `DaemonRuntime` 后，可以基于该 runtime 创建 daemon-backed `RemoteAgent`。重构后仍保留 `RemoteAgent` 作为 agent 配置入口，但 DAEMON 协议不再通过 `daemon_capability` 路由，而是引用 `daemon_runtime_id` 并选择 `acp_runtime`。
- `acp_runtime` 是创建 daemon agent 时选择的执行器类型。第一版写死支持 `opencode` 和 `codex`，后续再扩展为可配置 runtime catalog。同一个 `DaemonRuntime` 可以支持多个 ACP runtime，并由它派生多个 daemon agent。
- `work_dir` 不属于 daemon 启动参数，也不应该固定在 runtime 连接配置里。临时方案是运行 agent 时按 session id 在 `/tmp` 下自动创建隔离工作目录；后续再升级为由 request/context/channel 配置显式传入。
- 同一个 `DaemonRuntime` 不允许多个 daemon 同时在线连接。Registry 发现 `workspace_id + daemon_runtime_id` 已在线时，应拒绝新的连接；断开后可用同一个 token 重连。
- 第一阶段 ADK 行为保持当前模式：daemon bridge 等待 terminal update，最终 yield 一个 `session.Event`。`RUNNING` 进度只记录为 daemon task 状态，不强制映射成 ADK streaming event。

### 目标资源模型

重构后的核心关系：

```
Workspace
  └── DaemonRuntime
        ├── token: daemon:connect, binds workspace_id + runtime_id
        ├── single online connection: butter-daemon --url ... --token ...
        └── supported ACP runtimes: opencode, codex (hardcoded in v1)

RemoteAgent(protocol=DAEMON)
  ├── daemon_runtime_id
  ├── acp_runtime: opencode | codex | ...
  └── normal Agent config references it via remote_agent_ids

Agent invocation
  ├── user input
  ├── work_dir generated from session id under /tmp
  └── Bridge dispatches DaemonTask(runtime_id, acp_runtime, work_dir, input)
```

Suggested proto shape:

```protobuf
message DaemonRuntime {
  string id = 1;
  string name = 2;
  string description = 3;
  map<string, string> labels = 4;
  string workspace_id = 100;
}

message RemoteAgent {
  string id = 1;
  string name = 2;
  string url = 3; // A2A only
  RemoteAgentProtocol protocol = 4;

  reserved 5;
  reserved "daemon_capability";

  string daemon_runtime_id = 6; // DAEMON only
  string acp_runtime = 7;       // DAEMON only, e.g. "opencode" or "codex"
  string workspace_id = 100;
}

message DaemonTask {
  string task_id = 1;
  string agent_name = 2;
  string input = 3;
  string session_id = 4;
  string user_id = 5;
  map<string, string> metadata = 6;

  reserved 7;
  reserved "capability";

  string workspace_id = 8;
  string daemon_runtime_id = 9;
  string acp_runtime = 10;
  string work_dir = 11;
}
```

字段命名可以在实现前再统一，但语义应保持：`DaemonRuntime` 负责连接和执行面身份；`RemoteAgent` 负责从 runtime 派生一个可被 ADK runner 调度的 agent；`DaemonTask` 负责传递本次调用才知道的运行上下文。

### 服务端流程

1. 用户在 workspace 内创建 `DaemonRuntime`。
2. 服务端创建 runtime 资源，并通过 `CreateDaemonRuntimeToken` 或等价 RPC 签发一次性可见的 daemon token。token 默认长期有效，语义跟现有 API token 一致。
3. UI 展示启动命令：`butter-daemon --url <grpc-url> --token <token>`。
4. daemon 使用 token 连接 `DaemonConnectorService.Connect`。服务端校验 token，注册到 `Registry`，键为 `workspace_id + daemon_runtime_id`。
5. 如果该 runtime 已有在线连接，服务端拒绝第二个连接。runtime 断开后允许用同一个 token 重连。
6. 用户基于该 runtime 创建 `RemoteAgent(protocol=DAEMON, daemon_runtime_id=..., acp_runtime=...)`；第一版 `acp_runtime` 只能是 `opencode` 或 `codex`。
7. 普通 `Agent` 仍通过 `remote_agent_ids` 引用这个 remote agent。
8. `internal/agent/resolveRemoteAgents()` 为 DAEMON remote agent 构造 ADK `agent.Agent`，bridge 中持有 `workspace_id + daemon_runtime_id + acp_runtime`。
9. 运行时，bridge 根据 session id 生成 `/tmp` 下的临时 `work_dir`，构造 `DaemonTask` 并投递给对应 runtime connection。后续可把 `work_dir` 来源替换为 invocation/context metadata。

### Daemon 端职责

`cmd/butter-daemon` 重构后优先支持最小启动参数：

```bash
butter-daemon --url localhost:9090 --token bt_daemon_runtime_xxx
```

daemon 本地不再用 `daemon_id/name/allowed_capabilities` 声明 agent 能力。它连接成功后作为 runtime 执行面接收任务，并根据 `DaemonTask.acp_runtime` 选择内置 ACP adapter：

| `acp_runtime` | 默认命令建议 | 说明 |
|---------------|--------------|------|
| `opencode` | `opencode acp` | 现有 ACP executor 的默认 profile |
| `codex` | `codex-acp` | 参考 `zed-industries/codex-acp`；也可通过 `npx @zed-industries/codex-acp` 启动 |

如果后续需要自定义命令、环境变量或权限策略，可以再引入本地 optional config 或 server-side runtime policy；但当前已确认的产品目标是 daemon 可先用 `url + token` 启动。

### ADK Agent 适配

daemon-backed `RemoteAgent` 仍必须实现为 ADK agent。由于 ADK `agent.Agent` 接口有未导出的 internal 方法，外部包仍通过 `agent.New(agent.Config{Run: ...})` 构造：

```go
bridge := daemon.NewRuntimeBridge(registry, workspaceID, runtimeID, acpRuntime)
adkAgent, err := bridge.BuildAgent(remoteAgentName, description)
```

`Run` 行为保持第一阶段简单模型：

1. 从 `InvocationContext.UserContent()` 提取用户输入。
2. 根据 session id 在 `/tmp` 下生成或复用临时 `work_dir`。
3. 按 `workspace_id + daemon_runtime_id` 查找在线 runtime。
4. 下发 `DaemonTask{acp_runtime, work_dir, input}`。
5. 等待 `COMPLETED/FAILED/CANCELLED`。
6. 成功时 yield 一个最终 `session.Event`。

### 待确认问题

- 当前 `/tmp` 临时 `work_dir` 是由 server 创建并以绝对路径下发给 daemon；因此第一阶段隐含要求 daemon 与 server 在同一主机/容器文件系统内运行，或在相同绝对路径挂载共享 volume。独立容器/远程主机部署在未共享该路径时会因本地 executor 无法进入目录而失败；后续需要把 `work_dir` 来源改为 daemon-local 创建或显式 workspace mount 配置。
- `/tmp` 临时 `work_dir` 暂时不清理；需要记录后续 GC 需求，包括 TTL、手动清理、按 workspace/session 清理、失败任务是否保留现场。
- `codex-acp` 的权限策略、文件 callback、terminal callback 与现有 ACP executor 的具体映射细节。

## 目标架构（参考 agent-gw 模块拆分）

| 模块 | 职责 | 设计建议 |
|------|------|----------|
| Access Layer | 接收 Telegram / HTTP 请求 | 统一转成内部 Task Request |
| Agent Gateway | 统一鉴权、路由、会话、任务状态 | 对外只暴露 /tasks、/agents、/sessions 等统一接口 |
| Agent Adapter | 屏蔽不同 agent 差异 | 定义 Invoke() / Stream() / Cancel() / GetState() 接口 |
| Execution Plane | 真正执行任务 | ADK Go 走本地执行；daemon agent 走 client 长连接转发 |
| State & Queue | 任务持久化与异步调度 | 用 Redis/MongoDB 保存 session、task、event、result |

---

## 现有架构分析

### 已有的分层

Butter 当前并非完全无分层。实际已存在以下结构：

```
Access Layer (已有，但未统一抽象)
├── Telegram Poller     → runner.Run()
├── Discord Poller      → runner.Run()
├── A2A Handler         → runner.Run()
├── Twirp ReplySession  → runner.Run()
└── Cron Scheduler      → runner.Run()

Execution Engine (runner.Service，已承担部分 Gateway 职能)
├── Agent 注册表 (map[name]agent.Agent)
├── Proto Agent 热重载 (ReloadProtoAgents)
├── Model Override + 缓存
├── Session 管理 (create/get/clear via ADK session.Service)
├── Runner 池 (map[channel:agent:model]*adkrunner.Runner)
└── Plugin 系统 (ContextGuard, Langfuse, CompactionNotifier)

Agent Building (internal/agent/)
├── NewFromProto() → LLM/Loop/Sequential/Parallel
├── MCP Server 解析 (inline + registry by ID)
└── Remote Agent 解析 (仅 A2A 协议)

Config Management (ConfigStore + ConfigRuntime)
├── 存储后端: Memory / MongoDB
├── CRUD: Agents, MCPServers, RemoteAgents, Channels
├── 热重载: ReloadRunner → ReloadProtoAgents + Reload Channels
└── Twirp 服务: AgentService, MCPServerService, RemoteAgentService, ChannelService

Persistence
├── Session: MongoDB (adk_sessions, adk_events)
├── Memory: MongoDB (adk_memories)
├── Config: MongoDB (config_agents, config_mcpservers, config_remoteagents, config_channels)
└── Cron: MongoDB (cron_jobs, cron_executions)
```

### 当前请求流（以 Telegram 为例）

```
Telegram Update
  → Poller.handleUpdate()
    → 检查 allowlist + triggers
    → deriveSessionID (USER/CHAT scope)
    → getActiveAgent (Redis 或 channel 默认)
    → getActiveModel (Redis 或 channel 默认)
    → buildMessageParts (text + photo)
    → 构造 ContextInfo
    → runner.Run(ctx, agentName, parts, modelOverride, ctxInfo, onEvent, onCompaction)
      → 查找 agent → 处理 model override → getOrCreateRunner
      → ADK runner.Run(userID, sessionID, msg, config)
      → 收集 events → 返回 response text
    → sendReply (MarkdownV2)
```

### 已有的 Remote Agent 概念

Proto 中已定义 `RemoteAgent` 和 `RemoteAgentProtocol`：

```protobuf
message RemoteAgent {
  string id = 1;
  string name = 2;
  string url = 3;
  RemoteAgentProtocol protocol = 4;
}

enum RemoteAgentProtocol {
  REMOTE_AGENT_PROTOCOL_UNSPECIFIED = 0;
  REMOTE_AGENT_PROTOCOL_A2A = 1;
}
```

`internal/agent/agent.go` 中的 `resolveRemoteAgents()` 只处理 A2A 协议，通过 HTTP 主动调用远端。这种模式要求远端有公网可访问的 endpoint，**不适用于 daemon 场景**（client 通常在 NAT/防火墙后）。

### Runner.Run() 签名

```go
func (s *Service) Run(
    ctx context.Context,
    agentName string,
    parts []*genai.Part,
    modelOverride string,
    ctxInfo *agentsv1.ContextInfo,
    onEvent EventCallback,           // func(evt *session.Event)
    onCompaction CompactionCallback,  // func(agentName string)
) (string, error)
```

关键特征：**同步阻塞，返回最终文本**。对于 ADK agent 通常秒级返回，但 daemon agent 任务可能持续数分钟到数小时。

---

## Daemon Agent 的核心挑战

| 问题 | 说明 | 影响 |
|------|------|------|
| **连接方向** | Client 在 NAT 后，必须 client 主动连接 server | 不能复用 A2A（HTTP 主动调用）模式 |
| **长时执行** | opencode 任务可能运行数分钟到数小时 | runner.Run() 的同步阻塞模式不适用 |
| **流式回传** | 用户需要看到执行进度 | 需要 partial output 机制 |
| **断线恢复** | Daemon 可能断线重连 | 任务状态需持久化，断线后可恢复 |
| **多 Daemon 路由** | 同一 agent 类型可能有多个 daemon 实例 | 需要负载分配和健康检查 |

---

## 设计方案：增量实施

避免一次性大规模重构现有工作代码，采用增量方式：先在现有架构上扩展 daemon 能力，后续按需向目标架构演进。

### Phase 1：Workspace-scoped Daemon Execution Plane

**目标**：daemon 作为 workspace-scoped DB 配置资源存在；daemon client 通过专用 credential 连接 server，接收任务，执行并返回结果。不改动现有 runner/channel 流程。

#### 1.1 Proto 定义

新增 `proto/agents/v1/daemon.proto`：

```protobuf
syntax = "proto3";
package agents.v1;

import "google/protobuf/timestamp.proto";

// Daemon 注册时携带的元信息
message DaemonInfo {
  string daemon_id = 1;              // Daemon 实例唯一 ID
  string name = 2;                   // 人类可读名称
  repeated string capabilities = 3;  // 支持的执行能力 e.g. ["opencode", "claude-code"]
  map<string, string> labels = 4;    // 路由标签 e.g. {"repo": "butter", "env": "prod"}
  string version = 5;
  string os = 6;
  repeated string executors = 7;
  string workspace_id = 8;           // 可自报；服务端以 daemon credential 绑定的 workspace 为准
}

// Server 下发给 Daemon 的任务
message DaemonTask {
  string task_id = 1;
  string agent_name = 2;
  string input = 3;
  string session_id = 4;
  string user_id = 5;
  map<string, string> metadata = 6;
  string capability = 7;              // 执行能力路由键，daemon 据此选择 executor
  string workspace_id = 8;            // 任务所属 workspace
}

// Daemon 回传的任务进度/结果
message DaemonTaskUpdate {
  string task_id = 1;
  DaemonTaskStatus status = 2;
  string output = 3;                 // partial or final output
  string error = 4;                  // 错误信息 (FAILED 时)
  google.protobuf.Timestamp timestamp = 5;
}

enum DaemonTaskStatus {
  DAEMON_TASK_STATUS_UNSPECIFIED = 0;
  DAEMON_TASK_STATUS_ACCEPTED = 1;   // Daemon 已接收
  DAEMON_TASK_STATUS_RUNNING = 2;    // 执行中，output 为中间输出
  DAEMON_TASK_STATUS_COMPLETED = 3;  // 完成，output 为最终结果
  DAEMON_TASK_STATUS_FAILED = 4;     // 失败，error 有详情
  DAEMON_TASK_STATUS_CANCELLED = 5;  // 已取消
}

// Daemon 发给 Server 的消息（client→server 流方向）
message DaemonMessage {
  oneof message {
    DaemonInfo register = 1;          // 首条消息：注册
    DaemonTaskUpdate task_update = 2; // 后续消息：任务更新
  }
}

// Server 发给 Daemon 的消息（server→client 流方向）
message ServerMessage {
  oneof message {
    DaemonTask task = 1;              // 下发任务
    CancelTask cancel = 2;           // 取消任务
  }
}

message CancelTask {
  string task_id = 1;
}

// Daemon 长连接服务
service DaemonConnector {
  // 双向流：Daemon 连接后注册自身，然后持续交换任务和结果
  rpc Connect(stream DaemonMessage) returns (stream ServerMessage);
}
```

扩展现有 `agent.proto`：

```protobuf
enum RemoteAgentProtocol {
  REMOTE_AGENT_PROTOCOL_UNSPECIFIED = 0;
  REMOTE_AGENT_PROTOCOL_A2A = 1;
  REMOTE_AGENT_PROTOCOL_DAEMON = 2;    // 新增
}

// RemoteAgent 增加可选字段，url 移除 proto 级 min_len 校验
message RemoteAgent {
  string id = 1 [(validate.rules).string.min_len = 1];
  string name = 2 [(validate.rules).string.min_len = 1];
  string url = 3;                      // A2A 时必填（应用层校验），DAEMON 时可空
  RemoteAgentProtocol protocol = 4;
  string daemon_capability = 5;        // DAEMON 时必填（应用层校验）：匹配 DaemonInfo.capabilities
}
```

#### 1.2 Server 端新增模块

```
internal/runtime/daemon/
├── registry.go         # DaemonRegistry: 管理在线 daemon 连接
├── connection.go       # Connection: 封装单个 daemon 的双向流通信
├── bridge.go           # Bridge: 实现 ADK agent 接口，桥接到 daemon
└── grpc_handler.go     # gRPC DaemonConnector 服务实现
```

**DaemonRegistry**：

```go
type Registry struct {
    mu    sync.RWMutex
    conns map[string]map[string]*Connection  // workspace_id → daemon_id → connection
}

// Register 在 daemon 连接时调用
func (r *Registry) Register(conn *Connection)
// Unregister 在 daemon 断开时调用
func (r *Registry) Unregister(workspaceID, daemonID string)
// FindByCapability 在指定 workspace 查找具有指定能力的可用 daemon
func (r *Registry) FindByCapability(workspaceID, capability string) *Connection
// ListConnected 列出指定 workspace 的在线 daemon；workspaceID 为空时列出全部（admin/内部使用）
func (r *Registry) ListConnected(workspaceID string) []*DaemonInfo
```

**Connection**：

```go
type Connection struct {
    Info          *DaemonInfo
    WorkspaceID   string
    sendCh        chan *ServerMessage    // server → daemon
    activeTasks   sync.Map              // task_id → chan *DaemonTaskUpdate
    ConnectedAt   time.Time
}

// SendTask 将任务推送给 daemon，返回 result channel
func (c *Connection) SendTask(task *DaemonTask) (<-chan *DaemonTaskUpdate, error)
// CancelTask 发送取消指令
func (c *Connection) CancelTask(taskID string) error
// DispatchUpdate 将 daemon 回传的 update 分发到对应 task 的 channel
func (c *Connection) DispatchUpdate(update *DaemonTaskUpdate)
```

**Bridge** — 关键适配：让 daemon agent 可以作为 ADK sub-agent 被 runner 调用。

> **注意**：ADK v1.1.0 的 `agent.Agent` 接口包含未导出的 `internal()` 方法，外部包无法直接实现。
> 必须通过 `agent.New(agent.Config{Run: ...})` 构造 custom agent。

```go
// Bridge 将 daemon 执行包装为 ADK agent
type Bridge struct {
    registry   *Registry
    capability string         // 匹配 DaemonInfo.capabilities
}

// BuildAgent 通过 agent.New() 构造 ADK agent，Run 函数闭包引用 bridge 的 registry 和 capability
func (b *Bridge) BuildAgent(name, description string) (agent.Agent, error) {
    return agent.New(agent.Config{
        Name:        name,
        Description: description,
        Run:         b.daemonRunFunc,
    })
}

// daemonRunFunc 实现 func(agent.InvocationContext) iter.Seq2[*session.Event, error]
func (b *Bridge) daemonRunFunc(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
    return func(yield func(*session.Event, error) bool) {
        // 1. 从 InvocationContext.UserContent() 提取用户输入文本
        input := extractTextFromContent(ctx.UserContent())

        // 2. 查找可用 daemon
        conn := b.registry.FindByCapability(b.capability)
        if conn == nil {
            yield(nil, fmt.Errorf("no daemon available for capability %q", b.capability))
            return
        }

        // 3. 构造 DaemonTask（包含 capability 字段供 daemon 路由到 executor）
        task := &DaemonTask{
            TaskId:     uuid.NewString(),
            AgentName:  ctx.Agent().Name(),
            Input:      input,
            Capability: b.capability,
            SessionId:  ctx.Session().ID(),
            UserId:     ctx.Session().UserID(),
        }

        // 4. 发送任务，等待结果
        resultCh, err := conn.SendTask(task)
        if err != nil {
            yield(nil, err)
            return
        }

        var lastOutput string
        for {
            select {
            case update := <-resultCh:
                switch update.Status {
                case DAEMON_TASK_STATUS_COMPLETED:
                    event := session.NewEvent(ctx.InvocationID())
                    event.Author = ctx.Agent().Name()
                    event.LLMResponse.Content = genai.NewContentFromText(update.Output, genai.RoleModel)
                    yield(event, nil)
                    return
                case DAEMON_TASK_STATUS_FAILED:
                    yield(nil, fmt.Errorf("daemon task failed: %s", update.Error))
                    return
                case DAEMON_TASK_STATUS_RUNNING:
                    lastOutput = update.Output
                }
            case <-ctx.Done():
                conn.CancelTask(task.TaskId)
                if lastOutput != "" {
                    event := session.NewEvent(ctx.InvocationID())
                    event.Author = ctx.Agent().Name()
                    event.LLMResponse.Content = genai.NewContentFromText(lastOutput, genai.RoleModel)
                    yield(event, ctx.Err())
                } else {
                    yield(nil, ctx.Err())
                }
                return
            }
        }
    }
}
```

**gRPC Handler**：

```go
type GRPCHandler struct {
    agentsv1.UnimplementedDaemonConnectorServer
    registry *Registry
    tokenRepo apitoken.Repository
    runtimeRepo configrepo.DaemonRuntimeRepository
}

func (h *GRPCHandler) Connect(stream agentsv1.DaemonConnector_ConnectServer) error {
    // 1. 接收首条消息（必须是 register）
    // 2. 鉴权（从 gRPC metadata 提取 daemon credential）
    // 3. 使用 credential 绑定的 workspace_id + daemon_runtime_id 覆盖 register
    // 4. 创建 Connection，注册到 Registry
    // 5. 启动两个 goroutine:
    //    - sendLoop: conn.sendCh → stream.Send()
    //    - recvLoop: stream.Recv() → conn.DispatchUpdate()
    // 6. 等待 stream 关闭或 ctx 取消
    // 7. Unregister + 清理 activeTasks（通知等待方 daemon 已断开）
}
```

#### 1.3 集成到现有 Agent 构建流程

在 `internal/agent/agent.go` 的 `resolveRemoteAgents()` 中扩展：

```go
func resolveRemoteAgents(pb *agentsv1.Agent, registry []agentsv1.RemoteAgent, daemonRegistry *daemon.Registry) ([]agent.Agent, error) {
    // ... existing lookup logic ...

    for _, id := range cfg.GetRemoteAgentIds() {
        ra := byID[id]

        switch ra.GetProtocol() {
        case agentsv1.REMOTE_AGENT_PROTOCOL_A2A:
            if ra.GetUrl() == "" {
                return nil, fmt.Errorf("remote agent %q: A2A protocol requires non-empty url", ra.GetName())
            }
            a, err := remoteagent.NewA2A(...)
            agents = append(agents, a)

        case agentsv1.REMOTE_AGENT_PROTOCOL_DAEMON:
            if ra.GetDaemonCapability() == "" {
                return nil, fmt.Errorf("remote agent %q: DAEMON protocol requires non-empty daemon_capability", ra.GetName())
            }
            if daemonRegistry == nil {
                return nil, fmt.Errorf("remote agent %q: daemon registry not available", ra.GetName())
            }
            workspaceID := ra.GetWorkspaceId()
            if workspaceID == "" {
                workspaceID = pb.GetWorkspaceId()
            }
            bridge := daemon.NewBridge(daemonRegistry, workspaceID, ra.GetDaemonCapability())
            a, err := bridge.BuildAgent(ra.GetName(), fmt.Sprintf("Daemon agent: %s", ra.GetName()))
            if err != nil {
                return nil, fmt.Errorf("remote agent %q: building daemon bridge: %w", ra.GetName(), err)
            }
            agents = append(agents, a)
        }
    }
    return agents, nil
}
```

需要的改动：
- `runner.NewService()` 接收 `*daemon.Registry` 参数
- `NewFromProto()` 接收 `*daemon.Registry` 参数并传递给 `resolveRemoteAgents()`
- `daemon.Registry` 在 `cmd/butter/main.go` 中创建（顶层单例），同时传给：
  1. gRPC handler（`DaemonConnector` 服务实现）
  2. `StartChannels()` → `runner.NewService()`（新增参数）
- 这确保了 gRPC handler 和 runner 看到的是同一个 registry 实例

#### 1.4 gRPC Server 注册

在 `internal/app/routes.go` 中注册 gRPC 服务：

```go
// SetupRoutes 中新增
grpcServer := grpc.NewServer(/* interceptors */)
agentsv1.RegisterDaemonConnectorServer(grpcServer, daemonGRPCHandler)

// Gin 路由中挂载 gRPC（通过 cmux 或独立端口）
```

注意：Butter 当前使用 Gin + Twirp（HTTP），gRPC 需要独立端口或使用 cmux 多路复用。推荐独立端口，通过 `AppConfig` 新增 `GRPCPort` 配置。

#### 1.5 Client 端：Daemon

```
cmd/butter-daemon/
├── main.go           # CLI 入口，读取配置
├── connector.go      # gRPC 连接 + 自动重连（指数退避）
└── executor/
    ├── executor.go   # Executor 接口
    ├── acp.go        # 通用 ACP stdio executor
    └── shell.go      # 通用 shell command executor
```

**Executor 接口**：

```go
type Executor interface {
    // Execute 执行任务，通过 onUpdate 回传进度
    Execute(ctx context.Context, task *DaemonTask, onUpdate func(*DaemonTaskUpdate)) error
    // Capability 返回此 executor 支持的能力名称
    Capability() string
}
```

**ACP Executor（现状）**：

```go
type ACPConfig struct {
    Capability       string            `yaml:"capability"`
    Command          string            `yaml:"command"`
    Args             []string          `yaml:"args"`
    Env              map[string]string `yaml:"env"`
    WorkDir          string            `yaml:"work_dir"`
    PermissionPolicy string            `yaml:"permission_policy"` // deny | allow
    FS               ACPFSConfig       `yaml:"fs"`
    Terminal         bool              `yaml:"terminal"`
}

type ACPFSConfig struct {
    Read               bool `yaml:"read"`
    Write              bool `yaml:"write"`
    AllowAbsolutePaths bool `yaml:"allow_absolute_paths"`
}
```

当前 `ACPExecutor` 使用 `github.com/coder/acp-go-sdk` 通过 stdio 启动 ACP agent，并按
`Initialize → NewSession → Prompt` 生命周期执行任务。`capability` 仍是 server 路由到
daemon executor 的键；opencode 不再由 `opencode.go` 直调 CLI，而是配置成
`command: opencode` + `args: ["acp"]`。executor 支持 ACP permission、file
callbacks、terminal callbacks、任务取消与进程清理；文件回调默认限制在 `work_dir`
内，并会解析 symlink 防止通过工作区内链接逃逸。

**服务端 DaemonRuntime**（通过 `DaemonService` 写入 workspace 配置仓库）：

```yaml
id: "dev-machine-1"
name: "orvice-dev"
description: "Local development worker"
labels:
  repo: "butter"
  env: "dev"
```

随后调用 `CreateDaemonRuntimeToken` 为该 `daemon_runtime_id` 签发 worker credential。credential token 为 `API_TOKEN_KIND_DAEMON`，scope 为 `daemon:connect`，并绑定 workspace + daemon_runtime_id。

**Daemon 本地配置**（YAML 或 flags）：

```yaml
server: "butter.example.com:9090"  # gRPC 地址
credential: "bt_xxx"                # daemon credential
daemon_id: "dev-machine-1"
name: "orvice-dev"
labels:
  repo: "butter"
  env: "dev"
executors:
  acp:
    - capability: opencode
      command: opencode
      args: ["acp"]
      work_dir: "/home/user/workspace/butter"
      permission_policy: deny
      fs:
        read: true
        write: true
      terminal: true
  shell:
    work_dir: "/home/user/workspace/butter"
```

旧的本地配置形态仍可兼容：

```yaml
executors:
  opencode:
    work_dir: "/home/user/workspace/butter"
    binary: "opencode"
```

daemon 启动时会把它转换成 `capability: opencode`、`command: opencode`、
`args: ["acp"]` 的 ACP profile；新配置应直接使用 `executors.acp`。

**连接流程**：

```
启动 → 解析配置
     → gRPC Dial (server:port, 带 TLS 或 insecure)
     → DaemonConnector.Connect() 建立双向流
     → 发送 DaemonMessage{register: DaemonInfo{...}}
     → recvLoop:
         ← ServerMessage{task: DaemonTask}
           → 启动 goroutine: executor.Execute()
           → 回传 DaemonTaskUpdate (ACCEPTED → RUNNING → COMPLETED/FAILED)
         ← ServerMessage{cancel: CancelTask}
           → cancel task context
     → 断线: log + 指数退避重连 (1s, 2s, 4s, ..., 30s max)
```

---

### Phase 2：Task State 持久化

**目标**：Daemon 任务有完整的状态追踪，支持查询进度、断线恢复。

#### 2.1 Task Store

```
internal/runtime/daemon/
└── task_store.go     # 任务状态持久化
```

```go
type TaskStore interface {
    Create(ctx context.Context, task *Task) error
    Get(ctx context.Context, taskID string) (*Task, error)
    UpdateStatus(ctx context.Context, taskID string, status DaemonTaskStatus, output string) error
    ListByDaemon(ctx context.Context, daemonID string) ([]*Task, error)
    ListPending(ctx context.Context) ([]*Task, error)  // 断线恢复用
}

type Task struct {
    ID           string
    AgentName    string
    DaemonID     string
    Input        string
    Output       string
    Status       DaemonTaskStatus
    SessionID    string
    UserID       string
    ContextInfo  *agentsv1.ContextInfo
    CreatedAt    time.Time
    UpdatedAt    time.Time
    CompletedAt  *time.Time
}
```

存储后端：MongoDB collection `daemon_tasks`，与现有 mongo 基础设施复用。

#### 2.2 查询 API

在现有 Twirp 服务中新增或扩展：

```protobuf
service DaemonService {
  rpc ListDaemons(ListDaemonsRequest) returns (ListDaemonsResponse);
  rpc GetDaemonTask(GetDaemonTaskRequest) returns (GetDaemonTaskResponse);
  rpc ListDaemonTasks(ListDaemonTasksRequest) returns (ListDaemonTasksResponse);
  rpc CancelDaemonTask(CancelDaemonTaskRequest) returns (CancelDaemonTaskResponse);
}
```

#### 2.3 断线恢复

Daemon 重连后：
1. Server 检查 `daemon_tasks` 中该 daemon 的 RUNNING 状态任务
2. 将这些任务标记为 FAILED（daemon 端已丢失执行上下文）
3. 或者，如果 daemon 支持恢复，重新下发任务

---

### Phase 3：异步执行 + 通知回调

**目标**：Access Layer 提交 daemon 任务后不阻塞，任务完成后主动通知。

#### 3.1 问题

当前 `runner.Run()` 是同步的。Telegram Poller 调用后阻塞等待结果再 sendReply。对 daemon 长任务，这会导致：
- Telegram 消息处理阻塞
- HTTP 请求超时
- 无法展示中间进度

#### 3.2 方案：回调通知模式

不重构 runner.Run()，而是在 daemon bridge 层处理异步：

```go
// Bridge.RunAsync 异步版本，用于长任务
func (b *Bridge) RunAsync(
    ctx context.Context,
    task *DaemonTask,
    onProgress func(update *DaemonTaskUpdate),  // 进度回调
    onComplete func(output string, err error),   // 完成回调
) (taskID string, err error)
```

**Telegram Poller 适配**：

```go
func (p *Poller) handleMessage(ctx context.Context, msg *models.Message) {
    agent := p.getActiveAgent(...)

    if isDaemonAgent(agent) {
        // 异步模式
        p.sendReply(ctx, msg, "Task submitted, I'll notify you when done.")
        bridge.RunAsync(ctx, task,
            func(update) { p.sendProgressUpdate(ctx, msg, update) },  // 可选进度推送
            func(output, err) { p.sendReply(ctx, msg, output) },      // 完成通知
        )
    } else {
        // 同步模式（现有逻辑不变）
        resp, _ := runner.Run(...)
        p.sendReply(ctx, msg, resp)
    }
}
```

判断是否为 daemon agent：可通过 `runner.Service` 暴露方法查询 agent 的执行类型。

#### 3.3 进度更新

Daemon 执行期间的 partial_output 可以：
- Telegram：编辑原消息（edit message）或发送新消息
- API：通过 SSE 或 WebSocket 推送
- Twirp：通过轮询 GetDaemonTask 查询

---

### Phase 4（可选）：Agent Adapter 统一抽象

当同时支持 ADK agent 和 daemon agent 后，如果需要进一步解耦，可以引入 Adapter 层：

```go
type AgentAdapter interface {
    Invoke(ctx context.Context, task *Task) (string, error)
    Stream(ctx context.Context, task *Task, onEvent func(*TaskEvent)) error
    Cancel(ctx context.Context, taskID string) error
    GetState(ctx context.Context) (*AdapterState, error)
}
```

此时 `runner.Service` 内部按 agent 类型委托给不同 adapter：
- ADK Agent → ADKAdapter（包装现有 ADK runner 逻辑）
- Daemon Agent → DaemonAdapter（包装 Bridge）

**这一步非必须**，只有当 agent 类型继续增多（如容器化执行、Serverless 函数等）时再引入。

### Phase 5（可选）：Agent Gateway 统一入口

将 Telegram/Discord/A2A/Twirp 等入口统一收敛到 Gateway：

```go
type Gateway struct {
    runner    *runner.Service
    taskStore TaskStore
}

func (g *Gateway) Submit(ctx context.Context, req *TaskRequest) (*TaskResponse, error)
func (g *Gateway) SubmitAsync(ctx context.Context, req *TaskRequest) (taskID string, err error)
func (g *Gateway) Stream(ctx context.Context, req *TaskRequest, onEvent func(*TaskEvent)) error
```

所有入口只负责协议转换→TaskRequest，由 Gateway 统一处理鉴权、路由、会话、任务跟踪。

**这一步同样非必须**，现有架构中各入口直接调用 runner.Run() 已经工作良好。只有当入口数量增多或需要统一任务审计时再引入。

---

## 端到端流程

### ADK Agent（现有，不变）

```
Telegram → Poller.handleMessage()
  → runner.Run("assistant", parts, model, ctxInfo, onEvent, onCompaction)
    → ADK runner 本地执行 (LLM call + MCP tools)
  ← response text
  → sendReply()
```

### Daemon Agent（Phase 1，同步短任务）

```
Telegram → Poller.handleMessage()
  → runner.Run("coding-agent", parts, model, ctxInfo, onEvent, nil)
    → agent 树中的 daemon sub-agent 被 ADK 调度
    → Bridge.Run()
      → Registry.FindByCapability(workspace_id, "opencode")
      → Connection.SendTask(DaemonTask{workspace_id: ...})
        → gRPC stream → Daemon Client
        → Daemon: ACP executor 调用 opencode acp
        → DaemonTaskUpdate (RUNNING → COMPLETED)
      ← final output
    ← response text
  → sendReply()
```

### Daemon Agent（Phase 3，异步长任务）

```
Telegram → Poller.handleMessage()
  → 识别 daemon agent
  → sendReply("Task submitted")
  → Bridge.RunAsync(task, onProgress, onComplete)
    → Registry.FindByCapability(workspace_id, "opencode")
    → Connection.SendTask(DaemonTask{workspace_id: ...})
    → (后台) Daemon Client 执行
    → DaemonTaskUpdate (RUNNING, partial output)
      → onProgress → Telegram editMessage("Progress: ...")
    → DaemonTaskUpdate (COMPLETED, final output)
      → onComplete → Telegram sendReply("Done: ...")
```

---

## 文件改动清单

### 必须改动（Phase 1）

| 位置 | 改动 | 说明 |
|------|------|------|
| `proto/agents/v1/daemon.proto` | **新增** | Daemon 通信协议定义 |
| `proto/agents/v1/agent.proto` | 修改 | RemoteAgentProtocol 加 DAEMON，RemoteAgent 加 daemon_capability，url 移除 min_len 校验 |
| `buf.gen.yaml` | 可能修改 | 确保 gRPC plugin 生成 daemon 服务代码 |
| `internal/runtime/daemon/` | **新增** | registry.go, connection.go, bridge.go, grpc_handler.go |
| `internal/agent/agent.go` | 修改 | resolveRemoteAgents() 增加 DAEMON case |
| `internal/runtime/runner/runner.go` | 修改 | NewService() 接收 daemon.Registry 参数 |
| `internal/app/channels.go` | 修改 | StartChannels() 接收 daemon.Registry 参数 |
| `internal/config/config.go` | 修改 | AppConfig 增加 GRPCPort |
| `cmd/butter/main.go` | 修改 | 创建 daemon.Registry（顶层单例），启动 gRPC server，传 registry 给 StartChannels |
| `cmd/butter-daemon/` | **新增** | Daemon client binary |

### 可选改动（Phase 2-3）

| 位置 | 改动 | 说明 |
|------|------|------|
| `internal/runtime/daemon/task_store.go` | 新增 | 任务状态 MongoDB 持久化 |
| `proto/agents/v1/daemon_service.proto` | 新增 | Daemon 管理 Twirp API |
| `internal/application/daemon_service.go` | 新增 | Twirp 服务实现 |
| `internal/channel/telegram/poller.go` | 修改 | 异步模式 + 进度推送 |
| `internal/channel/discord/poller.go` | 修改 | 同上 |

### 不需要改动

| 位置 | 原因 |
|------|------|
| `internal/runtime/session/` | Session 由 ADK runner 管理，daemon bridge 层无需介入 |
| `internal/runtime/memory/` | 同上 |
| `internal/store/config/` | RemoteAgent CRUD 已支持，只需 proto 字段扩展 |
| `internal/application/remoteagent_service.go` | CRUD 逻辑不变，proto 字段自动支持 |
| `internal/runtime/cron/` | Cron 通过 runner.Run() 调用，daemon agent 对 cron 透明 |

---

## 实施建议

1. **正式资源模型优先**：当前基线是 workspace-scoped `DaemonRuntime`、runtime token、registry workspace 隔离和 ACP runtime dispatch 同步演进。
2. **gRPC 端口**：建议独立端口（如 9090），避免与 Gin HTTP server 冲突。go.mod 中已有 `google.golang.org/grpc v1.79.3`。
3. **鉴权**：daemon gRPC 只接受 `API_TOKEN_KIND_DAEMON` + `daemon:connect` credential，通过 gRPC metadata 传递 `authorization: Bearer <credential>`；root token 与普通 API token 不进入 daemon connector。
4. **测试策略**：Phase 1 完成后用集成测试验证：mock daemon client → gRPC 连接 → 收到任务 → 返回结果 → 验证 runner 拿到输出。
5. **Phase 4/5 按需引入**：除非 agent 类型超过 3 种或入口超过 5 个，否则不必急于引入 Adapter/Gateway 层。
