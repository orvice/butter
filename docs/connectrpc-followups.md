# ConnectRPC 迁移：后续工作

> Phase 0/1/2 已落地（见 [`migration-connectrpc.md`](migration-connectrpc.md)）。
> 本文列出尚未完成的、可独立成 PR 的后续项。

## 1. 运行时验证（**必做，最高优先级**）

整个迁移目前只验证了编译期和单元测试，**没有实地启动 dashboard 跑过一次**。
应至少完成下列手测路径，确认运行时字段映射、时间戳格式、enum 数值等不出意外：

1. `cp .env.example .env && export $(grep -v '^#' .env | xargs)`
2. `go run ./cmd/butter` + `cd front && pnpm dev`
3. 登录 → 看 sidebar 用户名/头像（验证 `User.display_name` / `avatar_url`）
4. Workspace 选择 → 进入仪表板（验证 `connected_daemons` / `active_sessions` 数字渲染）
5. 列表页：Agents / MCP Servers / Channels / Cron Jobs / Daemons —— 各看一个，
   重点关注：
   - 时间戳列（`created_at` / `updated_at` / `last_run_at`） — 验证 `tsToISO` 输出在前端被 `formatDate` 正确解析
   - 枚举字段（Agent 类型、Channel 状态、MCP transport） — 验证字符串 ↔ 数字映射
   - bigint 字段（`size_bytes` / `latency_ms`） — 验证 `bigintToNumber` 不丢精度
6. 创建/编辑：
   - 创建一个 Model Provider（最简单：name + type + api_key）
   - 创建一个 MCP Server（带 oauth2）
   - 创建一个 Cron Job
   - 编辑一个 Agent（**最复杂**：嵌套 config / sub_agents / mcp_servers / file_mounts）
7. Chat：发一条消息，确认 SSE 流式输出正常（这条没改过协议，但顺手验证）
8. 安装一个 Global MCP Server preset（验证新切换的 `GlobalMCPServerService.InstallGlobalMCPServer`）

**预期风险点**：
- `agents.ts` / `global-mcp-servers.ts` 用 `protojson` round-trip 处理嵌套，
  可能在边界值（空数组、`undefined` vs `null`、unset oneof 字段）上出意外
- `Timestamp` 在前端 form 里如果保留为 ISO string 直接提交回服务端，
  protojson 接受；但如果有地方读 `createdAt` 当 Date 对象用，会炸

## 2. Phase 3 — 彻底去 Twirp 依赖（机械、可分服务进行）

当前后端 `internal/application/*.go` 仍有 **251 处** `twirp.NewError` /
`twirp.RequiredArgumentError` / `twirp.NotFoundError` / `twirp.InternalErrorWith`
调用。通过 `connectx.WrapUnary` 在出口转码成 `connect.Error`，**功能正确**，
但 Twirp 依赖还在 `go.mod`，生成的 `*.twirp.go` 文件也还在。

### 操作步骤

1. 替换错误构造（机械转换表）：

   | Twirp 调用 | 替换为 |
   |---|---|
   | `twirp.NewError(twirp.InvalidArgument, msg)` | `connect.NewError(connect.CodeInvalidArgument, errors.New(msg))` |
   | `twirp.NewError(twirp.PermissionDenied, msg)` | `connect.NewError(connect.CodePermissionDenied, errors.New(msg))` |
   | `twirp.NewError(twirp.NotFound, msg)` | `connect.NewError(connect.CodeNotFound, errors.New(msg))` |
   | `twirp.NewError(twirp.AlreadyExists, msg)` | `connect.NewError(connect.CodeAlreadyExists, errors.New(msg))` |
   | `twirp.NewError(twirp.FailedPrecondition, msg)` | `connect.NewError(connect.CodeFailedPrecondition, errors.New(msg))` |
   | `twirp.NewError(twirp.Unauthenticated, msg)` | `connect.NewError(connect.CodeUnauthenticated, errors.New(msg))` |
   | `twirp.RequiredArgumentError("x")` | `connect.NewError(connect.CodeInvalidArgument, errors.New(\`x is required\`))` |
   | `twirp.InvalidArgumentError("x", reason)` | `connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s: %s", "x", reason))` |
   | `twirp.NotFoundError(name)` | `connect.NewError(connect.CodeNotFound, fmt.Errorf("%s not found", name))` |
   | `twirp.InternalErrorWith(err)` | `connect.NewError(connect.CodeInternal, err)` |

   建议先在一两个服务上跑通，再用 `sed` / `gopls rename`(可选) 批量。

2. 同步替换 service 测试里的 `twirp.Error` 类型断言（`internal/application/*_test.go`），
   改成 `connect.Error` 断言 / `connect.CodeOf(err)` 比较。

3. 当所有 `application/*.go` 不再 import `twirp` 后，**可以**删除 `connectx.WrapUnary`
   并把方法签名直接改成原生 Connect：
   ```go
   func (s *AuthServiceServer) Login(ctx context.Context, req *connect.Request[agentsv1.LoginRequest]) (*connect.Response[agentsv1.LoginResponse], error) {
       msg := req.Msg
       // ... 业务逻辑
       return connect.NewResponse(&agentsv1.LoginResponse{...}), nil
   }
   ```
   这样 `*_connect.go` 适配器也能删除（service 直接实现 `agentsv1connect.XxxServiceHandler`）。
   **可选**：也可以保留适配器一层，把 service 内部错误类型变成 `*connect.Error` 即可，
   两种风格各有取舍。

4. `buf.gen.yaml` 删除 `protoc-gen-twirp` 插件项。

5. 删除生成的 9 个 `pkg/proto/agents/v1/*.twirp.go` 文件
   （删后必须重新 `make buf` 一次确认无残留）。

6. `go mod tidy` 删除 `github.com/twitchtv/twirp` 依赖。

7. 删除 `internal/transport/connectx/errors.go` 里的 `TwirpErrorToConnect` / `twirpCodeToConnect`
   以及对应测试（`errors_test.go` 整个文件可删）。如果第 3 步保留了 `WrapUnary`，那 `wrap.go`
   也保留，但 `WrapUnary` 现在拿到的就是 `*connect.Error` 类型，可以直接 `return nil, err`
   省掉转码。

### 估算
- 单纯错误替换：1.5 小时机械工作，可写脚本辅助
- 测试断言更新：0.5 小时
- 删除生成文件 + 依赖 + 验证：15 分钟
- 合计：~2 小时一个 PR

## 3. 可选：原生 Connect 签名（深度清理）

如果团队偏好原生签名而非适配器层（Phase 3 第 3 步），需要再做：

- `internal/application/*_connect.go` 全部删除
- `internal/application/*_service.go` 的方法签名改为 `(ctx, *connect.Request[Req]) (*connect.Response[Res], error)`
- 业务代码访问 `req.Msg` 取请求体
- 收益：能直接读 Connect headers / trailers，
  方便后续做服务端 streaming（如果有需要把 chat SSE 收编到 Connect server-stream）
- 代价：所有现有 `internal/application/*_test.go` 也得改签名

这步不是必须的；适配器一层的运行时开销可以忽略。

## 4. 文档维护

- ✅ `docs/api.md` 错误格式、协议描述已更新（本轮）
- ✅ `docs/api.md` 已补 `GlobalMCPServerService` 章节
- 🔜 `docs/architecture.md` 仍可能提到 Twirp（未审计）
- 🔜 `docs/dashboard-api-gap.md` 如果还在用应该刷新
- 🔜 `AGENTS.md` / `CLAUDE.md` 里 "Twirp services" 表述应改成 "Connect services"

## 5. 杂项

### pnpm-lock.yaml / pnpm-workspace.yaml

Phase 2 期间 `pnpm install` 自动生成了 `front/pnpm-lock.yaml` 和
`front/pnpm-workspace.yaml`。项目 git 历史从未跟踪过 lockfile（Phase 2 commit 也没纳入），
当前状态：本地存在、未提交。

两种处理路径，选其一：
- **跟踪 lockfile**：把 `front/pnpm-lock.yaml` 加入 git；可重现安装。`pnpm-workspace.yaml`
  本会话生成的版本只有一行 `allowBuilds: msw: ...` 占位内容，可以丢弃。
- **继续不跟踪**：把 `front/pnpm-lock.yaml` 和 `pnpm-workspace.yaml` 加入 `.gitignore`
  显式表达策略。

### Daemon gRPC 单独端口

`internal/app/grpc.go` 现在为 `DaemonConnectorService` 起独立的 `:9090` gRPC 监听。
**与本次迁移无关**，daemon 协议是原生 gRPC，浏览器侧也不需要它。如果未来想统一到一个端口，
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
