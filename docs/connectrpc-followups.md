# ConnectRPC 迁移：后续工作

> Phase 0/1/2/3 已落地（见 [`migration-connectrpc.md`](migration-connectrpc.md)）。
> Phase 3 在 commit `2cbfdfd` 完成，Twirp 依赖、生成产物和所有 `twirp.X`
> 调用点已彻底从 runtime 移除（38k+ 行删除）。
> 本文列出剩余的、可独立成 PR 的后续项。

## 1. 运行时验证（**必做，最高优先级**）

整个迁移目前只验证了编译期和单元测试，**没有实地启动 dashboard 跑过一次**。
应至少完成下列手测路径，确认运行时字段映射、时间戳格式、enum 数值等不出意外：

1. `cp .env.example .env && export $(grep -v '^#' .env | xargs)`
2. `go run ./cmd/butter` + `cd front && npm run dev`
3. 登录 → 看 sidebar 用户名/头像（验证 `User.display_name` / `avatar_url`）
4. Workspace 选择 → 进入仪表板（验证 `connected_daemons` / `active_sessions`
   数字渲染）
5. 列表页：Agents / MCP Servers / Channels / Cron Jobs / Daemons —— 各看一个，
   重点关注：
   - 时间戳列（`created_at` / `updated_at` / `last_run_at`） — 验证 `tsToISO`
     输出在前端被 `formatDate` 正确解析
   - 枚举字段（Agent 类型、Channel 状态、MCP transport） — 验证字符串 ↔
     数字映射
   - bigint 字段（`size_bytes` / `latency_ms`） — 验证 `bigintToNumber` 不丢精度
6. 创建/编辑：
   - 创建一个 Model Provider（最简单：name + type + api_key）
   - 创建一个 MCP Server（带 oauth2）
   - 创建一个 Cron Job
   - 编辑一个 Agent（**最复杂**：嵌套 config / sub_agents / mcp_servers / file_mounts）
7. Chat：发一条消息，确认 `AgentService.StreamAgent` server-stream 正常
   （`started` → `text_delta` / `run_event` → `final`；Stop 用 AbortSignal +
   `CancelAgentInvocation`）
8. 安装一个 Global MCP Server preset（验证 Phase 2 切换的
   `GlobalMCPServerService.InstallGlobalMCPServer`）

**预期风险点**：
- `agents.ts` / `global-mcp-servers.ts` 用 `protojson` round-trip 处理嵌套，
  可能在边界值（空数组、`undefined` vs `null`、unset oneof 字段）上出意外
- `Timestamp` 在前端 form 里如果保留为 ISO string 直接提交回服务端，
  protojson 接受；但如果有地方读 `createdAt` 当 Date 对象用，会炸

## 2. Phase 3 已完成 ✅

后端 `internal/application/*.go` 的 251 处 `twirp.X` 错误构造全部替换为
`connect.NewError` 或 `connectx.{RequiredArgument,InvalidArgument,NotFound,
Internal,InternalWith}` helper。生成的 `*.twirp.go` 文件、`go.mod` 里的
`github.com/twitchtv/twirp` 依赖、`buf.gen.yaml` 里的 `protoc-gen-twirp`
插件项、`connectx.TwirpErrorToConnect` / `twirpCodeToConnect` 全部移除。

剩余结构（**有意保留**）：

- `connectx.HandlerOptions()` 仍强制 JSON 输出为 snake_case
  （`UseProtoNames=true`）。**前端浏览器调用现已切换到 binary protobuf**
  （`useBinaryFormat: true` in `front/src/api/transport.ts`），所以 JSON
  codec 主要是给 `curl` 调试和未来非浏览器调用方留的兜底。要彻底去掉这个
  codec hack 仍需全前端做 camelCase 重命名 + 撤掉 binary 切换，单独
  PR 更合适。

## 3. 原生 Connect 签名 ✅ 完成（commit `0f7218e`）

`connectx.WrapUnary` 适配器 + `internal/application/<svc>_connect.go` 文件
全部删除。Service 方法签名改为原生 ConnectRPC 形式
`(ctx, *connect.Request[Req]) (*connect.Response[Res], error)`，
直接实现 `agentsv1connect.XxxServiceHandler` 接口；
`routes.go` 通过 `agentsv1connect.NewXxxServiceHandler(svc, ...)` 挂载，
不再需要中间适配器。

副作用：
- `internal/application/*_test.go` 全部更新为
  `svc.Method(ctx, connect.NewRequest(&Req{...}))` 调用形式，
  响应字段通过 `resp.Msg.GetX()` 访问。
- `globalmcp_service.go::InstallGlobalMCPServer` 调用
  `s.mcpSvc.CreateMCPServer` 也跟着用 `connect.NewRequest` + `.Msg`。
- 净删除 **810 行**（15 个 adapter 文件 + WrapUnary 实现 + 测试 helper）。

原生 Connect 签名是 server-streaming RPC 的前提（`func(ctx, req, stream *connect.ServerStream[T]) error`）。
Dashboard chat 已用此形式实现 `AgentService.StreamAgent`（`agent_stream.go`），
替代已删除的 `POST /api/chat/stream` SSE handler。

## 4. 文档维护

- ✅ `docs/api.md` 错误格式、协议描述已更新；新增 `GlobalMCPServerService` 章节
- ✅ `docs/api.md`：`StreamAgent` 替代 SSE `chat/stream`；REST 列表与 dashboard wire 格式同步
- ✅ `docs/architecture.md` / `docs/app.md` / `docs/project-structure.md` / `docs/migration-connectrpc.md`：Connect 挂载与 chat/upload 边界同步
- ✅ `docs/architecture.md` / `docs/app.md` / `docs/project-structure.md`
  全部扫过并改写到 ConnectRPC 表述
- ✅ `AGENTS.md` / `CLAUDE.md` 已更新
- ✅ `docs/security-review.md` / `docs/frontend-required-apis.md` 已扫
- ✅ 历史文档（`design-daemon-agent.md` / `structure-review.md` /
  `dashboard-api-gap.md`）加了 stale notice 头部
- ✅ `docs/migration-connectrpc.md` 顶部加进度状态表

## 5. 杂项

### pnpm-lock.yaml / pnpm-workspace.yaml

Phase 2 期间 `pnpm install` 自动生成了 `front/pnpm-lock.yaml` 和
`front/pnpm-workspace.yaml`，但项目实际用 npm（`front/Dockerfile` 跑 `npm ci`，
`front/package-lock.json` 才是真实 lockfile）。已在 commit `be7a404` 修复：

- `npm install` 同步 `package-lock.json` 写入两个新 `@connectrpc/*` 依赖
- 删除 pnpm 残留文件
- `.gitignore` 加 `front/pnpm-lock.yaml` / `pnpm-workspace.yaml` / `yarn.lock`
  显式拒绝跨包管理器污染

### Daemon gRPC 单独端口

`internal/app/grpc.go` 现在为 `DaemonConnectorService` 起独立的 `:9090` gRPC
监听。**与本次迁移无关**，daemon 协议是原生 gRPC，浏览器侧也不需要它。如果未来想统一到一个端口，
得给主 HTTP 加 h2c 支持，不在本计划范围。

### 公共路径白名单

`internal/handler/http/auth.go::isPublicPath` 用硬编码字符串 match：

```go
case "/ping",
    "/api/agents.v1.AuthService/Login",
    "/api/agents.v1.AuthService/ListOAuthProviders",
    "/api/agents.v1.AuthService/BeginOAuthFlow",
    "/api/agents.v1.AuthService/CompleteOAuthFlow",
    "/api/mcp/oauth/callback":
```

如果未来给 AuthService 加新的"无需登录"方法，记得同步这个列表。可以考虑改成
Connect Interceptor 形式判定（基于 service 元信息），但当前规模够用，
保持现状成本最低。
