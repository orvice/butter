# 项目目录结构文档

更新时间：2026-05-17

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
│   ├── dashboard-api-gap.md
│   ├── design-daemon-agent.md
│   ├── project-structure.md
│   ├── structure-review.md
│   └── superpowers/
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
│       │   ├── agents/
│       │   ├── api-tokens/
│       │   ├── channels/
│       │   ├── cron/
│       │   ├── daemons/
│       │   ├── mcp-servers/
│       │   ├── remote-agents/
│       │   ├── sessions/
│       │   └── dashboard.tsx
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
│   ├── application/             # RPC service implementations.
│   │   │                        # Each service is split into:
│   │   │                        #   <svc>_service.go  — business logic (Twirp-shaped signatures)
│   │   │                        #   <svc>_connect.go  — ConnectRPC adapter (delegates via WrapUnary)
│   │   ├── agent_service.go
│   │   ├── agent_connect.go
│   │   ├── agentfile_service.go
│   │   ├── agentfile_connect.go
│   │   ├── apitoken_service.go
│   │   ├── apitoken_connect.go
│   │   ├── auth_service.go
│   │   ├── auth_oauth.go
│   │   ├── auth_connect.go
│   │   ├── channel_service.go
│   │   ├── channel_connect.go
│   │   ├── cron_service.go
│   │   ├── cron_connect.go
│   │   ├── daemon_service.go
│   │   ├── daemon_connect.go
│   │   ├── dashboard_service.go
│   │   ├── dashboard_connect.go
│   │   ├── forum_service.go
│   │   ├── forum_connect.go
│   │   ├── globalmcp_service.go
│   │   ├── globalmcp_connect.go
│   │   ├── mcpserver_service.go
│   │   ├── mcpserver_connect.go
│   │   ├── modelprovider_service.go
│   │   ├── modelprovider_connect.go
│   │   ├── notifygroup_service.go
│   │   ├── notifygroup_connect.go
│   │   ├── remoteagent_service.go
│   │   ├── remoteagent_connect.go
│   │   ├── runtime_mutation.go
│   │   ├── session_service.go
│   │   ├── session_connect.go
│   │   ├── workspace_service.go
│   │   └── workspace_connect.go
│   ├── transport/
│   │   └── connectx/            # WrapUnary adapter, Twirp→Connect error mapping,
│   │                            # snake_case JSON codec (HandlerOptions)
│   ├── channel/
│   │   ├── manager.go           # ChannelStatus + RuntimeState
│   │   ├── discord/
│   │   └── telegram/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   └── http/                # /ping, /a2a, /status, APITokenAuthMiddleware
│   ├── repo/
│   │   ├── apitoken/            # interface + memory + mongo (workspace-scoped)
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── auth/                # users + auth_sessions
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── config/              # workspace-scoped CRUD + AcrossWorkspaces listings
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── invocation/          # interface + memory + mongo
│   │   │   ├── memory/
│   │   │   ├── mongo/
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
│       ├── agent_service.proto
│       ├── agentchannel.proto
│       ├── api_token.proto
│       ├── auth.proto
│       ├── context.proto
│       ├── cron.proto
│       ├── daemon.proto
│       ├── dashboard.proto
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
├── buf.gen.yaml                 # go + connect + grpc-gateway + validate + twirp (unused, see docs/connectrpc-followups.md) + bufbuild/es
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
- `internal/application/`：RPC 服务实现（Agent / AgentFile / MCPServer / GlobalMCPServer / ModelProvider / NotifyGroup / RemoteAgent / Channel / Session / Cron / Dashboard / Daemon / APIToken / Auth / Forum / Workspace）。每个服务一对文件：`*_service.go` 写业务逻辑（当前签名仍是 Twirp 风格，错误用 `twirp.NewError`），`*_connect.go` 是 ConnectRPC 适配器，通过 `connectx.WrapUnary` 把方法包成 `agentsv1connect.XxxServiceHandler`，并在出口把 `twirp.Error` 翻译成 `connect.Error`。彻底去 Twirp 依赖的计划见 `docs/connectrpc-followups.md`。
- `internal/transport/connectx/`：ConnectRPC 共享 plumbing。`WrapUnary` 把 Twirp 风格签名包成 Connect handler 方法；`TwirpErrorToConnect` 做错误码翻译；`HandlerOptions()` 返回固定的 codec/option 列表，强制 JSON 输出用 `UseProtoNames=true`（snake_case），与原 Twirp wire format 一致。
- `internal/workspace/`：workspace context 包，提供 `WithID` / `FromContext` / `HeaderName="X-Workspace-ID"` / `DefaultSlug="default"`。
- `internal/repo/workspace/`：`workspaces` + `workspace_members` 仓库（memory + mongo），支撑 `WorkspaceService` 和 auth middleware 的成员校验。
- `internal/channel/`：渠道适配与渠道管理（Telegram、Discord），含 `RuntimeState` 探活。
- `internal/runtime/`：运行时能力 —— `runner`（含 invocation 记录与 cancel 注册）、`cron`（含 RunJobNow / 时序聚合）、`daemon`（registry / connection / bridge / grpc_handler / metrics）、`session`、`memory`。
- `internal/repo/`：仓库层。`config/` 是配置仓库（memory + mongo）；新增 `apitoken/` 与 `invocation/`，同样 memory + mongo 双实现。
- `front/`：Vite + React 19 dashboard。`src/api/` 是类型化的 ConnectRPC 客户端（`@connectrpc/connect-web`，一服务一文件），`src/api/transport.ts` 提供共享 transport + 注入 Authorization / X-Workspace-ID 的 interceptor，`src/api/_proto-bridge.ts` 提供 Timestamp/Duration/bigint 转换工具；`src/gen/` 是 buf 生成的 TS proto 类型（含 service definitions），`src/pages/` 一个目录一屏。
- `proto/`：Proto 定义源文件（agent + cron + daemon + dashboard + api_token + agentchannel + context）。
- `pkg/proto/`：Proto 生成代码（Go + Connect + grpc + validate，Twirp 仍生成但未被运行时代码引用）；不要手改。
- `.github/workflows/`：CI。后端走 `docker-publish.yml`，前端独立 `front-publish.yml`（`paths: front/**` 过滤），均推 ghcr 并 cosign 签名。
- `docs/`：项目文档。系统架构见 `architecture.md`；API 契约见 `api.md`；功能总览见 `app.md`。

## 维护建议

- 新增模块优先放在现有分层下，避免在 `internal/` 根目录继续平铺。
- `pkg/proto/` 与 `front/src/gen/` 均为生成代码目录，手动变更应在 `proto/` 中进行后 `make buf` 重新生成。
- 新增 RPC service：在 `proto/agents/v1/*.proto` 定义 service + messages → `make buf` 生成代码 → 在 `internal/application/` 加 `<svc>_service.go`（业务逻辑）和 `<svc>_connect.go`（按现有 adapter 模板，每个 RPC 一行 `connectx.WrapUnary`）→ 在 `internal/app/routes.go` 用 `agentsv1connect.NewXxxServiceHandler(...)` 创建 handler 并挂在 `/api/agents.v1.XxxService/*`（注意 `http.StripPrefix("/api", ...)`）→ 前端在 `front/src/api/` 加文件，`makeClient(XxxService)` 拿到类型化 client。
- 新增 MongoDB collection：在 `internal/repo/` 下加同名子包，提供 interface + memory + mongo 两实现，在 `internal/app/channels.go` 按 `storage_backend` 选择后端并注入 BootstrapResult。
- 结构变更后同步更新本文件，保证文档与仓库一致。
