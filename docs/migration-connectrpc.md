# API 迁移：Twirp → ConnectRPC / gRPC-Web

> 现状分析 + 分阶段改造计划。覆盖后端 RPC 服务、HTTP 鉴权中间件、前端调用层。

## 0. 进度状态（2026-06-02 收尾）

| 阶段 | 状态 | 说明 |
|---|---|---|
| Phase 0 — 公共设施 | ✅ 完成 | `connectx.WrapUnary` 适配器 + Twirp→Connect 错误码映射 + OPTIONS 预检放行 |
| Phase 1 — 后端切换 | ✅ 完成 | 14 个原 Twirp 服务全部 Connect 化；新增 `GlobalMCPServerService` 替代原 `/api/global-mcp-servers/*` REST |
| Phase 1.5 — JSON codec 修复 | ✅ 完成 | 强制 `UseProtoNames=true` 输出 snake_case，避免前端 130+ 处 snake_case 字段访问静默失效 |
| Phase 2 — 前端切换 | ✅ 完成 | 15 个 `front/src/api/*.ts` 全部用 `@connectrpc/connect-web` 类型化客户端，`twirpFetch` 删除 |
| Phase 3 — 去 Twirp 依赖 | ✅ 完成 (commit `2cbfdfd`) | 251 处 `twirp.X` 替换为 `connect.NewError` / `connectx.X` helper；`*.twirp.go` 38k+ 行删除；`go.mod` 移除 `github.com/twitchtv/twirp`；`buf.gen.yaml` 删 `protoc-gen-twirp` 插件；`TwirpErrorToConnect` 删除，`WrapUnary` 简化为 forward |
| Phase 3.5 — 原生 Connect 签名 | ✅ 完成 (commit `0f7218e`) | Service 方法签名改为 `(ctx, *connect.Request[Req]) (*connect.Response[Res], error)`；15 个 `*_connect.go` adapter + `connectx.WrapUnary` 全删；service 直接满足 `agentsv1connect.XxxServiceHandler`；净删 810 行 |
| Chat streaming | ✅ 完成 | `POST /api/chat/stream`（SSE）删除；dashboard 改用 `AgentService.StreamAgent` Connect server-stream（`internal/application/agent_stream.go`，`front/src/api/chat.ts`） |

测试现状：后端 383 个测试通过，前端 `tsc --force` 全量类型检查通过。非 Connect 调用仅保留 `uploads.ts`（multipart REST）；chat 已纳入 Connect（`StreamAgent` server-stream）。

后续工作清单：[`connectrpc-followups.md`](connectrpc-followups.md)。

---

> 下面第 1–5 节保留迁移前的现状分析、方案和执行计划，用于历史追溯；
> 当前实现状态以第 0 节、[`connectrpc-followups.md`](connectrpc-followups.md)
> 和 [`project-structure.md`](project-structure.md) 为准。

## 1. 现状梳理

### 1.1 后端 RPC 注册（Twirp）
- 入口 `cmd/butter/main.go` 通过 Butterfly 注入一个 Gin `router`，由 `internal/app/routes.go::SetupRoutes` 装配。
- 14 个 RPC 服务全部以 Twirp 形式挂载在 `/api` 前缀下：
  ```go
  pathPrefix := twirp.WithServerPathPrefix("/api")
  authTwirp := agentsv1.NewAuthServiceServer(authSvcServer, pathPrefix)
  r.Any(authTwirp.PathPrefix()+"*path", gin.WrapH(authTwirp))
  ```
- 服务实现位于 `internal/application/*.go`，签名为：
  ```go
  func (s *AuthServiceServer) Login(ctx context.Context, req *agentsv1.LoginRequest) (*agentsv1.LoginResponse, error)
  ```
  错误用 `twirp.NewError(...)` / `twirp.RequiredArgumentError(...)` 等构造，共 **254 处**调用、跨 24 个文件。

### 1.2 ConnectRPC 生成代码已就绪
- `buf.gen.yaml` 已开启 `buf.build/connectrpc/go:v1.18.1`。
- 生成产物在 `pkg/proto/agents/v1/agentsv1connect/`：`auth.connect.go` / `agent_service.connect.go` / …，每个服务都有 `NewXxxServiceHandler(svc, opts...) (string, http.Handler)`。
- 默认 mount path 是 `/agents.v1.XxxService/`（**没有 `/api` 前缀**）。
- `go.mod` 已包含 `connectrpc.com/connect v1.19.1`，无新增依赖成本。
- 但 ConnectHandler 期望的服务接口签名为：
  ```go
  Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error)
  ```
  和现有 Twirp 签名**不兼容**。

### 1.3 单独的 gRPC 服务
- `internal/app/grpc.go` 起了另一个独立的 `google.golang.org/grpc` 服务（默认 `:9090`），目前**只承载 `DaemonConnectorService`**（daemon 反向连接）。
- 主 HTTP（gin）端口和 gRPC 端口分离，两者不共用监听。

### 1.4 HTTP 中间件
- 入口：`internal/handler/http/auth.go::AuthMiddleware`，作为 Gin global middleware（`routes.go:334`）。
- 关键行为：
  1. `isPublicPath` 白名单放行 `/api/agents.v1.AuthService/Login` 等若干路径（**路径硬编码带 `/api`**）。
  2. 读 `Authorization: Bearer <token>`，依次尝试：
     - 用户 session（`auth.Repository.LookupSession`）
     - root token（常量时间比较）
     - API token（`apitoken.Repository.Lookup`）
  3. 命中后用 `c.Request.WithContext(...)` 写入 `user`/`admin`/`workspace_id`，下游通过 `auth.UserFromContext` / `wsctx.FromContext` 读取。
  4. `applyWorkspaceHeader` 读 `X-Workspace-ID`，校验成员关系后写入 context。
- 这个中间件运行在 Gin 层，**在 Twirp/Connect handler 之前**，对协议无感知。Connect 通过 `r.Context()` 拿到同一个 context。✅ 迁移时**无需重写鉴权逻辑**，只需关注 `isPublicPath` 的字符串匹配。
- 兼容路径补充：
  - `/api/global-mcp-servers/*` 仍是普通 Gin REST（注册于 `routes.go::registerGlobalMCPServerRoutes`）。
  - `/api/mcp/oauth/callback`、健康/状态/上传/A2A/聊天 SSE 等也都是 Gin handler，不受影响。
  - `writeError(c, err)` 在 REST 路由里把 `twirp.Error` 翻译成 HTTP，迁移后需要同时支持 `connect.Error`（或让相关 service 改成返回 connect 错误）。

### 1.5 前端调用层
- 已生成 `front/src/gen/agents/v1/*_pb.ts`（`@bufbuild/es`，**只有 message，没有 service client**）。
- `front/package.json` 有 `@bufbuild/protobuf` 但**没有** `@connectrpc/connect` / `@connectrpc/connect-web`。
- 所有调用走 `front/src/api/client.ts::twirpFetch`，自己 `fetch` 到 `/api/agents.v1.XxxService/Method`，body 是 JSON，header `Authorization` + `X-Workspace-ID`。
- 错误结构 `TwirpError = { code, msg }`，401 时清 token 跳登录。

### 1.6 路径前缀差异（迁移核心冲突点）
| 协议 | 现状 URL | Connect 默认 URL |
|---|---|---|
| Twirp | `/api/agents.v1.AuthService/Login` | — |
| Connect | — | `/agents.v1.AuthService/Login` |

同一 URL **不能同时**给 Twirp 和 Connect 处理：两者都接受 `POST application/json`，但 body shape 完全相同（恰好），错误响应却不同（Twirp `{"code","msg"}` vs Connect `{"code","message"}`）。所以同 URL 同时挂会导致协议 negotiation 含糊。推荐保留 `/api` 前缀作为 Connect 的最终落点，按服务粒度切换。

---

## 2. 改造方案

### 2.1 推荐方案：Connect 复用 `/api` 前缀 + 按服务渐进切换 + 适配器减少改动

- **保留 `/api` 前缀作为最终 URL**，前端可保持现有 URL 不变，只需替换 client 实现。
- 通过 `http.StripPrefix("/api", connectHandler)` 把 Connect handler 挂在 `/api/agents.v1.XxxService/*`，覆盖 Twirp 同名路由（按服务逐个切换，迁完一个删一个 Twirp）。
- 为减少 `internal/application/*.go` 改动量，写一个**通用适配器**把 Twirp 风格签名包成 Connect handler：
  ```go
  // 伪代码
  func wrapUnary[Req, Res any](fn func(context.Context, *Req) (*Res, error)) func(context.Context, *connect.Request[Req]) (*connect.Response[Res], error) {
      return func(ctx context.Context, req *connect.Request[Req]) (*connect.Response[Res], error) {
          resp, err := fn(ctx, req.Msg)
          if err != nil {
              return nil, twirpErrToConnect(err)
          }
          return connect.NewResponse(resp), nil
      }
  }
  ```
  这样每个 service 写一个 `XxxServiceConnectAdapter` 实现 `agentsv1connect.XxxServiceHandler` 接口，逐字段调用 `wrapUnary`。254 处 `twirp.NewError` **暂时保留**，由 `twirpErrToConnect` 统一映射。
- 后续阶段再把 `application/*.go` 改成原生 Connect 签名、把 `twirp.NewError` 全部替换成 `connect.NewError`。

### 2.2 协议覆盖范围
Connect-Go 的 `NewXxxServiceHandler` 一个 endpoint **同时支持** 三种协议：
- Connect（JSON / Protobuf）— 前端默认
- gRPC-Web（`application/grpc-web`，`application/grpc-web+proto`）— 浏览器可用
- 原生 gRPC — 需 HTTP/2 + h2c

Gin 跑在 HTTP/1.1 上，所以**浏览器侧拿到的是 Connect + gRPC-Web**，原生 gRPC 仍走 `internal/app/grpc.go` 那个独立 `:9090` 端口（daemon 用）。无需改 daemon 通道。

### 2.3 中间件改动点
- `isPublicPath` 的硬编码路径**保留 `/api` 前缀即可不变**。
- 加一项：Connect 浏览器请求会发 `OPTIONS` 预检（CORS）。当前 `AuthMiddleware` 不放行 `OPTIONS`，要么在 `isPublicPath` 加 `c.Request.Method == http.MethodOptions` 的早返回，要么在 Butterfly 的 CORS 中间件里处理。如果同源部署可以不管，跨域部署必须处理。
- Connect 客户端会带 header `Connect-Protocol-Version: 1`、`Connect-Timeout-Ms` 等；gRPC-Web 还要求暴露/转发 `Grpc-Status`、`Grpc-Message`、`Grpc-Status-Details-Bin` 等 trailer headers。如果走反向代理（Nginx/Cloudflare）需要 allowlist。
- `routes.go::writeError` 目前依赖 `twirp.Error` 分支：迁移期间继续兼容，新增 `if cerr, ok := err.(*connect.Error); ok { ... }` 分支。

### 2.4 前端改动
两条路：
- **A. 最小改动**：保留 `twirpFetch`，只调整 URL（不变）和错误字段：`err.message` 替代 `err.msg`。代价：丢失类型安全，未来要再改。
- **B. 推荐**：引入 `@connectrpc/connect@^2` + `@connectrpc/connect-web@^2`，在 `buf.gen.yaml` 加 `@connectrpc/connect-es` 插件生成 `*_connect.ts`，把 `front/src/api/*.ts` 改造成：
  ```ts
  const transport = createConnectTransport({ baseUrl: "/api", interceptors: [authInterceptor] });
  const authClient = createPromiseClient(AuthService, transport);
  ```
  鉴权 / workspace header 通过 Interceptor 注入，401 通过 `ConnectError.code === Code.Unauthenticated` 处理。

### 2.5 与现有 SSE/流式 endpoint 的关系（历史）
- 迁移当时 `chat_stream` / `a2a` 是自定义 handler，不在 Twirp 范围内。
- **后续已完成**：dashboard chat 已从 `POST /api/chat/stream`（SSE）迁到
  `AgentService.StreamAgent`（Connect server-stream）。`a2a` 仍为 Gin JSON-RPC，未改。

---

## 3. 分阶段执行计划

> 每一阶段都能独立合入并通过 `go test ./...` + 前端构建。

### 阶段 0 — 准备 & 公共设施（1 PR，约半天）
- [ ] 在 `internal/application` 或新建 `internal/transport/connect/` 下加：
  - `errors.go`：`func twirpErrToConnect(err error) *connect.Error`（按 `twirp.ErrorCode` switch）
  - `wrap.go`：`func wrapUnary[Req, Res any](...)` 适配器
- [ ] `routes.go::writeError` 增加 `connect.Error` 分支（前向兼容，REST 路由继续可用）。
- [ ] `internal/handler/http/auth.go::isPublicPath`：放行 `OPTIONS` 预检（如果未在 Butterfly CORS 中处理）。
- [ ] 增加单元测试覆盖 `twirpErrToConnect` 的 code 映射。

### 阶段 1 — 后端双挂载（按服务渐进，每服务 1 PR）
建议顺序（按风险从低到高）：
1. **AuthService**（公共路径多，需谨慎处理 `isPublicPath`）
2. WorkspaceService、APITokenService（数据少、用例简单）
3. AgentService / MCPServerService / ModelProviderService / NotifyGroupService / RemoteAgentService / ChannelService（配置类 CRUD）
4. ForumService / SessionService / CronService（带运行时依赖）
5. DashboardService / AgentFileService / DaemonService（聚合 / 大文件 / 跨 gRPC）

对每个服务：
- 加一个 `XxxConnectAdapter` 包住 `application.XxxServiceServer`。
- 在 `routes.go` 用 `http.StripPrefix("/api", connectHandler)` 挂到 `/api/agents.v1.XxxService/*`，**同时移除**该服务的 Twirp 挂载（这一刻就完成切换；Connect 默认 JSON codec 与原前端 `twirpFetch` body 兼容，零前端改动）。
- 跑后端集成测试 + 在前端用现有 `twirpFetch` 手测一次主要操作。
- 错误响应字段从 `{code,msg}` 变成 `{code,message}`：要么同步改前端 `client.ts`（把 `err.msg` 加 fallback 读 `err.message`），要么在适配器里塞一个临时 response 包装。**建议在阶段 0 就先把前端 `client.ts` 改成同时读 `message || msg`**，这样阶段 1 切换前端零感知。

### 阶段 2 — 前端切换到 Connect 客户端（1～2 PR）
- [ ] `buf.gen.yaml` 增加 `@connectrpc/connect-es` 插件，重新生成 `*_connect.ts`。
- [ ] `front/package.json` 加 `@connectrpc/connect` + `@connectrpc/connect-web`。
- [ ] 新建 `front/src/api/transport.ts` 集中创建 `createConnectTransport({ baseUrl: "/api" })` + auth/workspace interceptor。
- [ ] 按服务逐个把 `front/src/api/*.ts` 从 `twirpFetch` 改成 `createPromiseClient(...)`，类型化 input/output。
- [ ] 删 `twirpFetch`，保留 `apiFetch`（REST 用）。

### 阶段 3 — 原生化服务签名（可选，去 Twirp 依赖）
- [ ] 把 `internal/application/*.go` 的方法签名从 `(ctx, *Req) (*Res, error)` 改成 `(ctx, *connect.Request[Req]) (*connect.Response[Res], error)`，直接实现 Connect handler 接口。
- [ ] 把 254 处 `twirp.NewError` 全替换为 `connect.NewError(connect.CodeXxx, errors.New(...))`。
- [ ] 删除 `wrapUnary` 适配器。
- [ ] `buf.gen.yaml` 移除 `protoc-gen-twirp` 插件，删除 `pkg/proto/agents/v1/*.twirp.go`。
- [ ] 从 `go.mod` 移除 `github.com/twitchtv/twirp`。

### 阶段 4 — 可选优化
- [ ] 把 daemon gRPC 服务也挪到 Connect handler（统一 `:8080`）；如果一定要支持原生 gRPC 客户端，则给主 HTTP 配置 h2c（`golang.org/x/net/http2/h2c`）。
- [ ] 用 Connect Interceptor 实现统一 logging / tracing / workspace 注入，替代部分 gin middleware 逻辑（**注意**：这会让逻辑分散到两层，未必更好——除非未来 gRPC 直连绕过 gin 时需要）。

---

## 4. 风险与对策

| 风险 | 说明 | 对策 |
|---|---|---|
| URL 重叠 | Twirp 与 Connect 不能共用同 URL | 按服务一次性切换，不并存 |
| 错误字段差异 | 前端读 `msg` → Connect 是 `message` | 阶段 0 在 `client.ts` 同时兼容 `message ?? msg` |
| CORS / OPTIONS 预检 | `AuthMiddleware` 默认拒绝 OPTIONS | 在 `isPublicPath` 放行 OPTIONS 或交给 Butterfly CORS |
| `twirp.Error` 类型断言遗留 | `routes.go::writeError`、测试中可能直接断言 | 阶段 0 加 `connect.Error` 分支；阶段 3 才能彻底删 |
| 中间件 context 传递 | Connect handler 通过 `r.Context()` 取 context | Gin `WrapH` 已传递修改后的 request，无需改 |
| 测试覆盖 | `application/*_test.go` 直接调用方法，签名变化会全挂 | 阶段 1-2 保持原签名（用适配器）；阶段 3 改签名时同步重写测试 |
| daemon 服务 | 走独立 gRPC `:9090`，依赖原生 gRPC client | 本期不改 daemon；与本迁移解耦 |
| 公共路径白名单硬编码 | `isPublicPath` 路径含 `/api` 前缀 | 保留 `/api` 前缀部署即可，无需改 |

---

## 5. 第一步建议

如果同意此计划，建议第一个 PR 落地：
1. 新建 `internal/transport/connect/wrap.go` + `errors.go`（适配器 + 错误映射）。
2. `front/src/api/client.ts` 错误字段双兼容（`message ?? msg`）。
3. `internal/handler/http/auth.go` 放行 OPTIONS。
4. 单元测试：错误码映射表。

第二个 PR 起按服务粒度切换（推荐从 `AuthService` 起步，因为它最敏感、迁完最大幅度验证路径白名单/CORS）。
