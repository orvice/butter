# Dashboard 后端接口缺口分析

更新时间：2026-06-13

> **状态：历史快照（全部 closed）。**
>
> 这份文档是在 PR #25 开工前对照 Stitch 设计稿做的差距清单，列出了 13 项后端缺口与 5 个分阶段实施分组。截至 2026-05-16，文档中提出的接口与字段全部已落地，对应 RPC 见 `docs/api.md`，功能总览见 `docs/app.md`。
> 2026-06-13 起，Operations 不再只有 cron 自动化：`AutomationService`、`automation_runs`、`automation_step_runs` 与 Automations dashboard 页面已补齐，Cron 也扩展了 timeout / retry / concurrency / notify / output policy 和更丰富的 execution 状态。本文仍只作为历史 gap 记录。
>
> 保留本文作为当时的决策记录与 RPC 命名依据；不再用作工作清单。新增能力请直接更新 `api.md` / `app.md` / `architecture.md`。

基于 Stitch 设计稿 `Butter Agent Dashboard`（project `6407848364648104779`）逐屏比对当前后端能力，列出缺失的 RPC/HTTP 接口与建议的 proto 改动。

设计稿包含 6 个功能屏：

| Screen | 主要内容 |
|---|---|
| Dashboard Overview | 实时指标、Cron 执行图表、系统健康、活动流、daemon 握手事件 |
| Agent Orchestration | Agent 列表、运行状态、子 agent 树、调用日志、Stop Execution、Hot-reload |
| MCP & Remote Integrations | MCP server 状态/工具数、tool whitelist、Remote Agent 状态 |
| Execution & Daemon Monitoring | Daemon 列表、active tasks、cancel、bridge 诊断 |
| Channels & Connectivity | Channel 状态/last poll、Restart/Resume、API token 管理 |
| Operations & Sessions | Automation CRUD/Run history、Cron CRUD/Run now、Session 过滤、Memory/Event log、Langfuse 链接 |

## 1. 现状覆盖矩阵

| 设计需求 | 当前是否支持 | 说明 |
|---|---|---|
| Agent CRUD | ✅ | `AgentService` |
| MCP Server CRUD | ✅ | `MCPServerService` |
| Remote Agent CRUD | ✅ | `RemoteAgentService` |
| Channel CRUD | ✅ | `ChannelService` |
| Automation CRUD + 运行历史 | ✅ | `AutomationService` + `ListAutomationRuns` / `ListAutomationStepRuns` |
| Cron CRUD + 历史 | ✅ | `CronJobService` + `ListCronExecutions` |
| Session CRUD + Reply | ✅ | `SessionService`，含过滤/分页 |
| Daemon 连接（gRPC） | ✅ | `DaemonConnectorService.Connect`（仅供 daemon client 接入） |
| 健康/Status | ✅ | `DashboardService.GetOverview` 实时探活 MongoDB / Redis / Runner |
| Dashboard 聚合指标 | ✅ | `DashboardService.GetOverview` |
| Cron 执行图表数据 | ✅ | `GetCronExecutionTimeseries` |
| 活动流 / 最近事件 | ✅ | `GetActivityFeed` + `invocations` |
| Agent 运行状态/last_run | ✅ | `GetAgentRuntimeStatus` / `ListAgentRuntimeStatuses` |
| 调用日志 / 调用历史 | ✅ | `ListAgentInvocations` |
| 手动触发 Agent / Automation / Cron | ✅ | `InvokeAgent` / `RunAutomationNow` / `RunCronJobNow` |
| 取消运行中任务 | ✅ | `CancelAgentInvocation` / `CancelDaemonTask` |
| Daemon 列表 / 详情 | ✅ | `DaemonService.ListDaemons` / `GetDaemon` |
| Daemon 任务监控 | ✅ | `DaemonService.ListDaemonTasks` |
| Bridge 诊断指标 | ✅ | `DaemonService.GetBridgeDiagnostics` |
| MCP 连通性 / ListTools | ✅ | `GetMCPServerStatus` / `ListMCPTools` |
| Channel 运行状态 / Restart | ✅ | `GetChannelStatus` / `RestartChannel` / `PauseChannel` / `ResumeChannel` |
| API Token 多 token 管理 | ✅ | `APITokenService` + daemon credential tokens |
| Langfuse trace URL | ✅ | `SessionEvent.trace_url` |

## 2. 建议新增 / 扩展的接口

### 2.1 DashboardService（新增）

聚合首屏所需的所有统计数据。

```proto
service DashboardService {
  rpc GetOverview(GetOverviewRequest) returns (GetOverviewResponse);
  rpc GetActivityFeed(GetActivityFeedRequest) returns (GetActivityFeedResponse);
  rpc GetCronExecutionTimeseries(GetCronExecutionTimeseriesRequest)
      returns (GetCronExecutionTimeseriesResponse);
}

message GetOverviewRequest {
  string environment = 1; // production | staging | development
}

message GetOverviewResponse {
  Counts counts = 1;       // 总数 + 趋势
  HealthSummary health = 2;
  DaemonHandshake latest_daemon_handshake = 3;
}

message Counts {
  int32 active_agents = 1;
  int32 mcp_servers = 2;
  int32 connected_daemons = 3;
  int32 active_sessions = 4;
  // 各自的 delta（百分比 / 绝对差）
  string agents_trend = 5;
  string mcp_trend = 6;
  string daemons_trend = 7;
  string sessions_trend = 8;
}

message HealthSummary {
  ComponentHealth mongodb = 1;
  ComponentHealth redis = 2;
  ComponentHealth runner = 3;
}

message ComponentHealth {
  enum Status { UNKNOWN = 0; HEALTHY = 1; DEGRADED = 2; DOWN = 3; }
  Status status = 1;
  string detail = 2;          // 区域 / 节点 / 错误信息
  google.protobuf.Timestamp checked_at = 3;
  int64 latency_ms = 4;
}

message DaemonHandshake {
  string daemon_id = 1;
  string os = 2;
  repeated string capabilities = 3;
  google.protobuf.Timestamp timestamp = 4;
}

message GetActivityFeedRequest {
  int32 limit = 1;          // default 50
  string cursor = 2;
}

message ActivityEvent {
  string id = 1;
  string kind = 2;          // invocation | execution_completed | warning | error
  string actor = 3;         // agent name / webhook id
  string message = 4;
  google.protobuf.Timestamp timestamp = 5;
}

message GetCronExecutionTimeseriesRequest {
  enum Range { RANGE_1D = 0; RANGE_7D = 1; RANGE_30D = 2; }
  Range range = 1;
  string job_name = 2;      // 可选，过滤单个 job
}

message GetCronExecutionTimeseriesResponse {
  repeated TimeBucket buckets = 1;
}

message TimeBucket {
  google.protobuf.Timestamp start = 1;
  int32 success = 2;
  int32 error = 3;
}
```

### 2.2 AgentService 扩展

```proto
service AgentService {
  // 现有 CRUD ...
  rpc GetAgentRuntimeStatus(GetAgentRuntimeStatusRequest)
      returns (GetAgentRuntimeStatusResponse);
  rpc InvokeAgent(InvokeAgentRequest) returns (InvokeAgentResponse);
  rpc ListAgentInvocations(ListAgentInvocationsRequest)
      returns (ListAgentInvocationsResponse);
  rpc CancelAgentInvocation(CancelAgentInvocationRequest)
      returns (CancelAgentInvocationResponse);
  rpc ReloadAgents(ReloadAgentsRequest) returns (ReloadAgentsResponse);
}

message AgentRuntimeStatus {
  string name = 1;
  enum State { IDLE = 0; RUNNING = 1; FAILED = 2; }
  State state = 2;
  google.protobuf.Timestamp last_run_at = 3;
  string last_invocation_id = 4;
  int32 in_flight = 5;
}

message InvokeAgentRequest {
  string agent_name = 1;
  string input = 2;
  string app_name = 3;
  string user_id = 4;
  string session_id = 5;    // 可选，空则新建临时 session
  string model_override = 6;
}

message InvokeAgentResponse {
  string invocation_id = 1;
  string session_id = 2;
  string response = 3;
}

message Invocation {
  string id = 1;
  string agent_name = 2;
  string session_id = 3;
  string app_name = 4;
  string user_id = 5;
  string status = 6;         // running | succeeded | failed | cancelled
  string input = 7;
  string output = 8;
  string error = 9;
  google.protobuf.Timestamp started_at = 10;
  google.protobuf.Timestamp finished_at = 11;
  int64 latency_ms = 12;
}
```

需要在 `runner.Service.Run` 周围加 invocation 持久化（新增 `internal/runtime/invocation` 仓库），并在 MongoDB 中存 `invocations` collection。

### 2.3 DaemonService（新增）

把 workspace-scoped `DaemonConfig`、daemon credential 签发和在线 registry 暴露成 RPC/HTTP 接口。

```proto
service DaemonService {
  rpc ListDaemonConfigs(ListDaemonConfigsRequest) returns (ListDaemonConfigsResponse);
  rpc GetDaemonConfig(GetDaemonConfigRequest) returns (GetDaemonConfigResponse);
  rpc CreateDaemonConfig(CreateDaemonConfigRequest) returns (CreateDaemonConfigResponse);
  rpc UpdateDaemonConfig(UpdateDaemonConfigRequest) returns (UpdateDaemonConfigResponse);
  rpc DeleteDaemonConfig(DeleteDaemonConfigRequest) returns (DeleteDaemonConfigResponse);
  rpc CreateDaemonCredential(CreateDaemonCredentialRequest)
      returns (CreateDaemonCredentialResponse);
  rpc ListDaemons(ListDaemonsRequest) returns (ListDaemonsResponse);
  rpc GetDaemon(GetDaemonRequest) returns (GetDaemonResponse);
  rpc ListDaemonTasks(ListDaemonTasksRequest) returns (ListDaemonTasksResponse);
  rpc CancelDaemonTask(CancelDaemonTaskRequest) returns (CancelDaemonTaskResponse);
  rpc GetBridgeDiagnostics(GetBridgeDiagnosticsRequest)
      returns (GetBridgeDiagnosticsResponse);
}

message DaemonStatus {
  string daemon_id = 1;
  string name = 2;
  repeated string capabilities = 3;
  string version = 4;
  string os = 5;
  string remote_addr = 6;
  enum State { UNKNOWN = 0; ONLINE = 1; IDLE = 2; OFFLINE = 3; }
  State state = 7;
  google.protobuf.Timestamp connected_at = 8;
  google.protobuf.Duration uptime = 9;
  int32 active_tasks = 10;
  repeated string executors = 11; // shell / open-code / claude-code ...
  string workspace_id = 12;
}

message DaemonTaskStatusView {
  string task_id = 1;
  string daemon_id = 2;
  string agent_name = 3;
  DaemonTaskStatus status = 4; // 复用 daemon.proto 中的枚举
  string current_step = 5;
  int32 progress = 6;          // 0~100
  google.protobuf.Timestamp started_at = 7;
  google.protobuf.Duration elapsed = 8;
}

message BridgeDiagnostics {
  double cpu_percent = 1;
  int64 memory_used_bytes = 2;
  int64 memory_limit_bytes = 3;
  repeated LatencyPoint latency = 4; // 用于 1h 折线图
}
```

实现层需要：

- `daemon.Registry` 增加 `Get`、并为 `DaemonInfo` 扩展 `version`/`os`/`remote_addr`/`connected_at` 字段（修改 `daemon.proto` 中的 `DaemonInfo`）。
- `daemon.Connection` 暴露 active task 列表。
- 新增 `internal/runtime/daemon/metrics.go` 输出 bridge 诊断。
- 新增 `config_daemons` 持久化 workspace daemon config；daemon credential 使用 `API_TOKEN_KIND_DAEMON` + `daemon:connect` scope。

### 2.4 MCPServerService 扩展

```proto
service MCPServerService {
  // 现有 CRUD ...
  rpc GetMCPServerStatus(GetMCPServerStatusRequest)
      returns (GetMCPServerStatusResponse);
  rpc ListMCPTools(ListMCPToolsRequest) returns (ListMCPToolsResponse);
  rpc TestMCPServer(TestMCPServerRequest) returns (TestMCPServerResponse);
}

message MCPServerStatus {
  string id = 1;
  enum State { UNKNOWN = 0; CONNECTED = 1; DISCONNECTED = 2; ERROR = 3; }
  State state = 2;
  int32 tool_count = 3;
  string error = 4;
  google.protobuf.Timestamp checked_at = 5;
}

message MCPTool {
  string name = 1;
  string description = 2;
  string server_id = 3;
  bool allowed = 4;  // 是否在 tool_filter 白名单内
}
```

### 2.5 RemoteAgentService 扩展

```proto
service RemoteAgentService {
  // 现有 CRUD ...
  rpc GetRemoteAgentStatus(GetRemoteAgentStatusRequest)
      returns (GetRemoteAgentStatusResponse);
  rpc TestRemoteAgent(TestRemoteAgentRequest) returns (TestRemoteAgentResponse);
}

message RemoteAgentStatus {
  string id = 1;
  RemoteAgentProtocol protocol = 2;
  enum State { UNKNOWN = 0; ACTIVE = 1; IDLE = 2; UNREACHABLE = 3; }
  State state = 3;
  string detail = 4;
  google.protobuf.Timestamp last_seen = 5;
}
```

### 2.6 ChannelService 扩展

```proto
service ChannelService {
  // 现有 CRUD ...
  rpc GetChannelStatus(GetChannelStatusRequest) returns (GetChannelStatusResponse);
  rpc RestartChannel(RestartChannelRequest) returns (RestartChannelResponse);
  rpc PauseChannel(PauseChannelRequest) returns (PauseChannelResponse);
  rpc ResumeChannel(ResumeChannelRequest) returns (ResumeChannelResponse);
}

message ChannelStatus {
  string name = 1;
  AgentChannelType type = 2;
  enum State { UNKNOWN = 0; LIVE = 1; PAUSED = 2; ERROR = 3; }
  State state = 3;
  google.protobuf.Timestamp last_poll_at = 4;
  string detail = 5;
  int64 messages_24h = 6;
}
```

`channel.Manager` 需新增 `Status(name)` / `Restart(name)` / `Pause(name)` / `Resume(name)` 方法。

### 2.7 APITokenService（新增）

设计稿要求“多 token 管理 + 标签 + Revoke / Regenerate”。当前已落地 `ListAPITokens` / `CreateAPIToken` / `RevokeAPIToken`，并扩展 `kind` / `scopes` / `expires_at` / `daemon_id` 支持 daemon credential；`RegenerateAPIToken` 仍未作为单独 RPC 实现，可通过 revoke + create 完成。

```proto
service APITokenService {
  rpc ListAPITokens(ListAPITokensRequest) returns (ListAPITokensResponse);
  rpc CreateAPIToken(CreateAPITokenRequest) returns (CreateAPITokenResponse);
  rpc RevokeAPIToken(RevokeAPITokenRequest) returns (RevokeAPITokenResponse);
}

message APIToken {
  string id = 1;
  string name = 2;
  string prefix = 3;          // 展示用前缀，例如 bt_live_8f92
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp last_used_at = 5;
  bool revoked = 6;
  APITokenKind kind = 7;
  repeated string scopes = 8;
  google.protobuf.Timestamp expires_at = 9;
  string daemon_id = 10;
  string workspace_id = 100;
}

message CreateAPITokenResponse {
  APIToken token = 1;
  string secret = 2; // 只在创建时返回一次
}
```

`AuthMiddleware` 从 token collection 校验普通 API token（`API_TOKEN_KIND_USER` + `api:*`），daemon connector 独立校验 daemon credential（`API_TOKEN_KIND_DAEMON` + `daemon:connect`）。

### 2.8 CronJobService 扩展

```proto
service CronJobService {
  // 现有 CRUD + ListCronExecutions ...
  rpc RunCronJobNow(RunCronJobNowRequest) returns (RunCronJobNowResponse);
}

message RunCronJobNowResponse {
  CronExecution execution = 1;
}
```

实现：直接调用 `cron.Scheduler.RunNow(jobName)`，绕过 schedule 立即执行并写入 executions。

### 2.9 SessionService 扩展

设计稿的 Session Explorer 需要按 channel / user / date range 过滤，并在 Session Event 上挂 trace URL。

```proto
message ListSessionsRequest {
  // 现有 app_name / user_id 保持兼容
  string app_name = 1;
  string user_id = 2;
  // 新增
  string channel = 3;        // 等价 app_name；与现有兼容
  google.protobuf.Timestamp start_time = 4;
  google.protobuf.Timestamp end_time = 5;
  int32 page_size = 6;
  string page_token = 7;
}

message ListSessionsResponse {
  repeated SessionInfo sessions = 1;
  string next_page_token = 2;
  int32 total = 3;
}

message SessionEvent {
  // 现有字段 ...
  string trace_id = 7;
  string trace_url = 8;   // Langfuse URL，由服务端拼装
}

message SessionDetail {
  // 现有字段 ...
  google.protobuf.Struct memory_context = 3; // 长期记忆快照
}
```

需要：

- 在 ADK runner 中把 Langfuse trace_id 写入 session event metadata，由服务端转成 URL。
- `MemoryService.Get` 暴露给 SessionService 用于回填 `memory_context`。

### 2.10 Health 扩展（HTTP）

将 `GET /status` 拓展为返回 MongoDB / Redis ping 结果，或新增 `GET /healthz` / `GET /readyz`。

```json
{
  "service": "butter",
  "components": {
    "mongodb": {"status": "healthy", "latency_ms": 4},
    "redis":   {"status": "healthy", "latency_ms": 1},
    "runner":  {"status": "healthy"}
  },
  "storage": { ... 原有内容 ... }
}
```

## 3. 数据持久化新增

| Collection | 用途 | 触发写入位置 |
|---|---|---|
| `invocations` | Agent / Daemon 任务调用记录 | `runner.Service.Run` finally |
| `automations` | Workspace 自动化定义 | AutomationService |
| `automation_runs` | 自动化 run 历史 | automation.Engine |
| `automation_step_runs` | 自动化 step-run 历史 | automation.Engine |
| `activity_events` | 仪表盘活动流 | runner / channel / daemon callback |
| `api_tokens` | 多 token 管理 | APITokenService |
| `config_daemons` | Workspace daemon 配置 | DaemonService |
| `channel_metrics` | 渠道消息计数（24h） | 渠道 poller 周期上报 |

## 4. 实施建议（分阶段）

1. **Phase 1 — Read-only 监控**：DashboardService、DaemonService、Channel/MCP/RemoteAgent Status、Status 扩展。前端可立即接通。
2. **Phase 2 — 可写控制**：InvokeAgent、CancelAgentInvocation、RunCronJobNow、ChannelService Restart/Pause/Resume、CancelDaemonTask、TestMCPServer/TestRemoteAgent。
3. **Phase 3 — 身份与历史**：APITokenService、Invocations 持久化、SessionService 过滤/分页、Langfuse trace 链接。

## 5. 与现有约束的兼容

- 所有新增 RPC 仍走 `/api/agents.v1.*` 路径，复用 `APITokenAuthMiddleware`。
- 新增 proto 需在 `proto/agents/v1` 下定义，运行 `make buf` 重新生成 `pkg/proto`。
- `daemon.proto` 中 `DaemonInfo` 加字段为兼容修改（新字段编号 5+，旧 daemon 不需要立即升级）。
- 配置层（`internal/store/config`）目前只管理静态配置；运行时状态走独立 runtime registry，不进入 ConfigStore，避免误触发 reload。
