# 项目目录结构文档

更新时间：2026-06-03

```text
butter/
├── cmd/
│   ├── butter/
│   │   └── main.go
│   └── butter-daemon/
│       ├── main.go
│       ├── connector.go
│       └── executor/
├── config/
├── docs/
│   ├── api.md
│   ├── app.md
│   ├── architecture.md
│   ├── connectrpc-followups.md
│   ├── dashboard-api-gap.md
│   ├── design-daemon-agent.md
│   ├── frontend-required-apis.md
│   ├── migration-connectrpc.md
│   ├── postgres-migration-analysis.md
│   ├── project-structure.md
│   ├── security-review.md
│   ├── storage.md
│   └── structure-review.md
├── front/
│   ├── Dockerfile
│   ├── nginx.conf
│   ├── .dockerignore
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── api/                 # Typed ConnectRPC clients (one file per service)
│       │                        # + transport.ts (shared interceptors)
│       │                        # + _proto-bridge.ts (Timestamp/Duration helpers)
│       ├── gen/                 # protoc-gen-es output (@bufbuild/protobuf)
│       │   ├── agents/v1/
│       │   └── validate/
│       ├── hooks/
│       ├── layouts/
│       ├── lib/
│       ├── pages/
│       │   ├── admin/
│       │   ├── agent-files/
│       │   ├── agents/
│       │   ├── api-tokens/
│       │   ├── channels/
│       │   ├── chat/
│       │   ├── cron/
│       │   ├── daemons/
│       │   ├── forum/
│       │   ├── mcp-servers/
│       │   ├── model-providers/
│       │   ├── notify-groups/
│       │   ├── remote-agents/
│       │   ├── sessions/
│       │   ├── users/
│       │   ├── auth-oauth-callback.tsx
│       │   ├── dashboard.tsx
│       │   ├── integrations.tsx
│       │   ├── login.tsx
│       │   ├── operations.tsx
│       │   ├── profile.tsx
│       │   └── workspaces.tsx
│       └── types/
├── internal/
│   ├── agent/
│   │   ├── agent.go             # NewFromProto + ProbeMCPServer
│   │   ├── model.go
│   │   ├── model_test.go
│   │   └── system/
│   ├── app/
│   │   ├── channels.go          # bootstrap (mongo, redis, runner, channels, repos)
│   │   ├── config_runtime.go
│   │   ├── config_store.go
│   │   ├── cron.go
│   │   ├── grpc.go
│   │   ├── routes.go            # ConnectRPC + HTTP + auth wiring
│   │   ├── runtime.go
│   │   └── system_agent.go
│   ├── application/             # RPC service implementations. Each
│   │   │                        # `<svc>_service.go` uses native ConnectRPC
│   │   │                        # signatures and is handed straight to
│   │   │                        # agentsv1connect.NewXxxServiceHandler.
│   │   ├── agent_service.go
│   │   ├── agent_stream.go        # AgentService.StreamAgent server-stream
│   │   ├── agentfile_service.go
│   │   ├── apitoken_service.go
│   │   ├── auth_service.go
│   │   ├── auth_oauth.go
│   │   ├── channel_service.go
│   │   ├── cron_service.go
│   │   ├── daemon_service.go
│   │   ├── dashboard_service.go
│   │   ├── forum_service.go
│   │   ├── globalmcp_service.go
│   │   ├── mcpserver_service.go
│   │   ├── modelprovider_service.go
│   │   ├── notifygroup_service.go
│   │   ├── remoteagent_service.go
│   │   ├── runtime_mutation.go
│   │   ├── session_service.go
│   │   └── workspace_service.go
│   ├── transport/
│   │   └── connectx/            # connect.Error helpers (RequiredArgument
│   │                            # /NotFound/InternalWith/…) + snake_case
│   │                            # JSON codec (HandlerOptions)
│   ├── channel/
│   │   ├── manager.go           # ChannelStatus + RuntimeState
│   │   ├── discord/
│   │   └── telegram/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   └── http/                # /ping, /status, /a2a, /api/uploads/*, auth middleware
│   ├── repo/
│   │   ├── agentfile/           # agent file spaces + files (workspace-scoped)
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── apitoken/            # interface + memory + mongo (workspace-scoped)
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── auth/                # users + auth_sessions (Redis wrapper in production)
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── config/              # workspace-scoped CRUD + AcrossWorkspaces listings
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── forum/               # forum threads + posts (workspace-scoped)
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── invocation/          # interface + memory + mongo
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── mcpoauth/            # MCP OAuth2 token store
│   │   │   └── repository.go
│   │   ├── oauthstate/          # OAuth state nonce store (used by auth + MCP OAuth)
│   │   │   └── repository.go
│   │   ├── workspace/           # workspaces + workspace_members
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   └── health.go
│   ├── runtime/
│   │   ├── cron/                # scheduler + repo (job + execution, workspace-scoped) + ListByTimeRange
│   │   ├── daemon/              # registry, connection (snapshots/cancel),
│   │   │                        # bridge, grpc_handler, metrics
│   │   ├── memory/mongo/
│   │   ├── runner/              # Service.Run, InvocationRecorder, CancelInvocation
│   │   └── session/mongo/       # CountSessions
│   ├── service/
│   │   ├── health.go
│   │   └── status.go
│   └── workspace/               # ctx propagation: WithID / FromContext / HeaderName
├── openspec/
│   ├── changes/
│   └── specs/
├── pkg/
│   ├── agent/
│   └── proto/
│       └── agents/
├── proto/
│   └── agents/v1/
│       ├── agent.proto
│       ├── agent_file.proto
│       ├── agent_service.proto
│       ├── agentchannel.proto
│       ├── api_token.proto
│       ├── auth.proto
│       ├── context.proto
│       ├── cron.proto
│       ├── daemon.proto
│       ├── dashboard.proto
│       ├── forum.proto
│       └── workspace.proto
├── .github/
│   └── workflows/
│       ├── buf.yml
│       ├── docker-publish.yml   # backend image → ghcr.io/<owner>/<repo>
│       ├── front-publish.yml    # frontend image → ghcr.io/<owner>/<repo>-front
│       └── go.yml
├── .claude/
├── .codex/
├── .kilocode/
├── .env.example
├── AGENTS.md
├── CLAUDE.md
├── buf.gen.yaml                 # go + connect + grpc-gateway + validate + bufbuild/es
├── buf.lock
├── buf.yaml
├── config.yaml
├── Dockerfile                   # Go backend (distroless static)
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
└── README.md
```

## 目录说明

- `cmd/`：进程入口。`butter` 是服务端；`butter-daemon` 是通过 gRPC 反连服务端的 daemon client（自报 version / os / executors）。
- `internal/app/`：应用装配与初始化（路由、gRPC、运行时、配置仓库、渠道、Cron、系统 Agent、token / workspace 仓库选择、初始 admin 与 default workspace bootstrap、Langfuse host 透传）。
- `internal/application/`：RPC 服务实现（Agent / AgentFile / MCPServer / GlobalMCPServer / ModelProvider / NotifyGroup / RemoteAgent / Channel / Session / Cron / Dashboard / Daemon / APIToken / Auth / Forum / Workspace）。每个服务一个 `*_service.go`，方法签名是原生 ConnectRPC 形式 `(ctx, *connect.Request[Req]) (*connect.Response[Res], error)`，直接满足 `agentsv1connect.XxxServiceHandler` 接口，由 `routes.go` 通过 `agentsv1connect.NewXxxServiceHandler(svc, ...)` 挂载。错误用 `connect.NewError` 或 `connectx` helper 构造。
- `internal/transport/connectx/`：ConnectRPC 共享 plumbing。`RequiredArgument` / `InvalidArgument` / `NotFound` / `Internal` / `InternalWith` 是 `connect.Error` 的常用构造短手；`HandlerOptions()` 含 snake_case JSON codec（`UseProtoNames=true`）供 curl/非浏览器调用；dashboard 浏览器默认 binary protobuf（`front/src/api/transport.ts`）。
- `internal/workspace/`：workspace context 包，提供 `WithID` / `FromContext` / `HeaderName="X-Workspace-ID"` / `DefaultSlug="default"`。
- `internal/repo/workspace/`：`workspaces` + `workspace_members` 仓库（memory + mongo），支撑 `WorkspaceService` 和 auth middleware 的成员校验。
- `internal/channel/`：渠道适配与渠道管理（Telegram、Discord），含 `RuntimeState` 探活。
- `internal/runtime/`：运行时能力 —— `runner`（含 invocation 记录与 cancel 注册）、`cron`（含 RunJobNow / 时序聚合）、`daemon`（registry / connection / bridge / grpc_handler / metrics）、`session`、`memory`。
- `internal/repo/`：仓库层。`config/` 是配置仓库（memory + mongo）；新增 `apitoken/` 与 `invocation/`，同样 memory + mongo 双实现。
- `front/`：Vite + React 19 dashboard。`src/api/` 是类型化的 ConnectRPC 客户端（`@connectrpc/connect-web`，一服务一文件）；`uploads.ts` 是唯一仍用裸 `fetch` + multipart 的 API 模块。`src/api/transport.ts` 提供共享 transport（binary protobuf）+ Authorization / X-Workspace-ID interceptor；`chat.ts` 调用 `StreamAgent` server-stream。`src/gen/` 是 buf 生成的 TS proto 类型；`src/pages/` 一目录一屏。
- `proto/`：Proto 定义源文件（agent + cron + daemon + dashboard + api_token + agentchannel + context）。
- `pkg/proto/`：Proto 生成代码（Go + Connect + grpc + grpc-gateway + validate）；不要手改。Twirp 生成产物已在 ConnectRPC Phase 3 移除。
- `.github/workflows/`：CI。后端走 `docker-publish.yml`，前端独立 `front-publish.yml`（`paths: front/**` 过滤），均推 ghcr 并 cosign 签名。
- `docs/`：项目文档。系统架构见 `architecture.md`；API 契约见 `api.md`；功能总览见 `app.md`。

## 维护建议

- 新增模块优先放在现有分层下，避免在 `internal/` 根目录继续平铺。
- `pkg/proto/` 与 `front/src/gen/` 均为生成代码目录，手动变更应在 `proto/` 中进行后 `make buf` 重新生成。
- 新增 RPC service：在 `proto/agents/v1/*.proto` 定义 service + messages → `make buf` 生成代码 → 在 `internal/application/` 加 `<svc>_service.go`，每个 RPC 方法签名为 `func (s *XxxServiceServer) Method(ctx context.Context, req *connect.Request[agentsv1.YyyRequest]) (*connect.Response[agentsv1.YyyResponse], error)`，方法体访问 `req.Msg.GetX()`，返回 `connect.NewResponse(&agentsv1.YyyResponse{...})` → 在 `internal/app/routes.go` 用 `agentsv1connect.NewXxxServiceHandler(svc, connectOpts...)` 创建 handler，并挂到 `/api/agents.v1.XxxService/*`（`http.StripPrefix("/api", handler)`）→ 前端在 `front/src/api/` 加文件，`makeClient(XxxService)` 拿到类型化 client。
- 新增 MongoDB collection：在 `internal/repo/` 下加同名子包，提供 interface + memory + mongo 两实现，在 `internal/app/channels.go` 按 `storage_backend` 选择后端并注入 BootstrapResult。
- 结构变更后同步更新本文件，保证文档与仓库一致。
