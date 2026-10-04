# 项目目录结构文档

更新时间：2026-10-04

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
│   ├── agents/              # issue tracker, triage labels, domain-doc notes for agent skills
│   ├── api.md
│   ├── app.md
│   ├── architecture.md
│   ├── daemon.md
│   ├── dashboard-api-gap.md
│   ├── design-daemon-agent.md
│   ├── design-dashboard-chat-async.md
│   ├── postgres-migration-analysis.md
│   ├── prd-workspace-agent-git-content.md
│   ├── project-structure.md
│   ├── repo-binding-migration.md
│   ├── research/
│   ├── storage.md
│   └── adr/
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
│       ├── routes/                # TanStack Router file-based routes
│       ├── features/              # Resource and page implementations
│       ├── components/
│       ├── context/
│       ├── stores/
│       ├── lib/
│       ├── styles/
│       └── types/
├── internal/
│   ├── a2ui/                    # A2UI v0.9.1 layer: butter-basic-v1 catalog, result cards, Human Input forms
│   ├── a2uitool/                # render_ui tool for A2UI-capable AG-UI runs
│   ├── agent/
│   │   ├── agent.go             # NewFromProto + ProbeMCPServer
│   │   ├── model.go
│   │   ├── model_test.go
│   │   └── system/
│   ├── agentcontent/            # Agent Content parsed and validated from the repository cache
│   ├── agentfiletool/           # agent_files_* tools over mounted Agent File spaces
│   ├── aguitool/                # AG-UI client-declared frontend tools, per run
│   ├── app/
│   │   ├── channels.go          # bootstrap (mongo, redis, runner, channels, repos)
│   │   ├── config_runtime.go
│   │   ├── config_store.go
│   │   ├── cron.go
│   │   ├── h2c.go
│   │   ├── invocations.go       # process liveness + stale invocation sweep (#390)
│   │   ├── reconciler.go
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
│   │   ├── automation_service.go
│   │   ├── butterbox_service.go       # ButterBoxes (ADR-0011)
│   │   ├── channel_service.go
│   │   ├── cron_service.go
│   │   ├── daemon_service.go
│   │   ├── dashboard_service.go
│   │   ├── forum_service.go
│   │   ├── githost_service.go         # platform Git host allowlist
│   │   ├── globalmcp_service.go
│   │   ├── linear_app_service.go       # Linear Apps (ADR-0015)
│   │   ├── linear_install.go           # install flow, OAuth callback, installations
│   │   ├── linear_references.go        # Agent reference guard
│   │   ├── linear_admin_service.go     # platform Linear base URL
│   │   ├── linear_processing_service.go
│   │   ├── mcpserver_service.go
│   │   ├── modelprovider_service.go
│   │   ├── notifygroup_service.go
│   │   ├── remoteagent_service.go
│   │   ├── repobinding_service.go     # workspace repository binding, sync, publish (ADR-0005, ADR-0007)
│   │   ├── runtime_mutation.go
│   │   ├── session_service.go
│   │   ├── skill_service.go
│   │   ├── telegram_admin_service.go
│   │   ├── telegram_channel_service.go
│   │   ├── telegram_destination_service.go
│   │   ├── telegram_processing_service.go
│   │   ├── workspace_service.go
│   │   ├── workspace_memory_service.go   # WorkspaceMemoryConfig (ADR-0013)
│   │   └── workspace_memories_service.go # list, search, delete memories
│   ├── auth/
│   │   └── provider/            # dashboard sign-in OAuth providers (GitHub, Google)
│   ├── eventqueue/              # protocol-neutral Redis Streams hand-off (Telegram + Linear)
│   ├── linearapi/               # Linear OAuth + GraphQL client; lineartest/ fake
│   ├── redact/                  # best-effort credential redaction for outgoing text
│   ├── gitprovider/             # provider-neutral Git hosting API seam (GitHub, GitLab REST)
│   ├── mcpoauth/                # OAuth2 for MCP servers: discovery, authorization flow, tokens
│   ├── mcpserver/               # MCP server exposing read-only workspace tools
│   ├── mem0/                    # minimal mem0 OSS REST client (ADR-0013)
│   ├── memorytool/              # search_memory / add_memory tools (ADR-0013)
│   ├── notify/                  # Notify Group delivery
│   ├── redislease/              # Redis-backed renewable lease electing one holder
│   ├── secretbox/               # AES-GCM encryption for persisted credentials
│   ├── skilltool/               # workspace skills as an ADK skill source (ADR 0004)
│   ├── telegramapi/             # Telegram Bot API seam, one client per decrypted token
│   ├── telegramqueue/           # Redis Streams hand-off from Telegram receive to workers
│   ├── telegramsend/            # the single outbound path to Telegram
│   ├── testsupport/
│   │   ├── openaifake/          # scripted OpenAI-compatible backend for tests
│   │   └── tooltest/            # checks tool declarations, e.g. parameter descriptions
│   ├── transport/
│   │   └── connectx/            # connect.Error helpers (RequiredArgument
│   │                            # /NotFound/InternalWith/…) + snake_case
│   │                            # JSON codec (HandlerOptions)
│   ├── channel/
│   │   └── manager.go           # Legacy AgentChannel reporter; starts no transport
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   └── http/                # /ping, /status, /a2a, /api/uploads/*, auth middleware,
│   │                            # AG-UI (agui*.go: run, detached runs + fan-out, A2UI,
│   │                            # UI snapshot, thread history),
│   │                            # OpenAI-compatible API, repository webhook, workspace MCP,
│   │                            # Telegram + Linear webhooks, Linear OAuth callback
│   ├── repo/
│   │   ├── agentcontent/        # validated Agent Content snapshots by repository revision
│   │   ├── agentfile/           # agent file spaces + files (workspace-scoped)
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── agentop/             # durable Agent lifecycle operation (Saga) records
│   │   ├── apitoken/            # interface + memory + mongo (workspace-scoped)
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── auth/                # users + auth_sessions (Redis wrapper in production)
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── butterbox/           # workspace ButterBoxes; access token behind the credential seam
│   │   ├── config/              # workspace-scoped CRUD + AcrossWorkspaces listings
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── cryptokey/           # database-backed master key for credential encryption
│   │   ├── forum/               # forum threads + posts (workspace-scoped)
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   ├── githost/             # platform allowlist of Git hosts
│   │   ├── inputpart/           # multimodal input parts stored apart from events
│   │   ├── invocation/          # interface + memory + mongo; owner stamp + stale sweep
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   ├── repotest/        # Repository conformance suite (memory + mongo)
│   │   │   ├── repository.go
│   │   │   └── stale.go         # StaleSweeper: fail runs whose owning process is gone
│   │   ├── linear/              # Linear Apps + Installations, credential seam (repotest/)
│   │   ├── linearprocessing/    # Linear processing records, ADR-0009 claim (repotest/)
│   │   ├── linearsetting/       # platform Linear settings
│   │   ├── linearstate/         # single-use Linear install states (hashed, TTL)
│   │   ├── mcpoauth/            # MCP OAuth2 token store
│   │   │   └── repository.go
│   │   ├── memoryconfig/        # per-workspace mem0 connection (WorkspaceMemoryConfig)
│   │   ├── oauthstate/          # OAuth state nonce store (used by auth + MCP OAuth)
│   │   │   └── repository.go
│   │   ├── repobinding/         # zero-or-one workspace repository binding; PAT via credential seam
│   │   ├── repocache/           # workspace repository cache snapshots from sync
│   │   ├── skill/               # skills: Mongo metadata + ContentStore for SKILL.md (ADR 0004)
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   ├── repotest/        # Repository conformance suite (memory + mongo)
│   │   │   └── repository.go
│   │   ├── telegram/            # Telegram Channels + Destinations, credential seam
│   │   ├── telegramprocessing/  # Telegram processing records (ADR-0009)
│   │   ├── telegramsetting/     # platform Telegram settings (Webhook base URL)
│   │   ├── workspace/           # workspaces + workspace_members
│   │   │   ├── memory/
│   │   │   ├── mongo/
│   │   │   └── repository.go
│   │   └── health.go
│   ├── runtime/
│   │   ├── asyncrun/              # Dashboard async invocation coordinator
│   │   ├── automation/            # Automation engine, scheduler and repositories
│   │   ├── butterboxconn/         # ButterBox connection resolution for pibox + cursorbox
│   │   ├── cron/                # scheduler + repo (job + execution, workspace-scoped) + ListByTimeRange
│   │   ├── cursorbox/             # AGENT_TYPE_CURSOR bridge to a ButterBox (ADR-0012)
│   │   ├── daemon/              # registry, connection (snapshots/cancel),
│   │   │                        # bridge, grpc_handler, metrics
│   │   ├── interrupt/            # Pending/Resume workflow interrupt seam
│   │   ├── linear/               # Linear receiver, worker, orchestrator, session coordination
│   │   ├── linearconn/           # Linear installation token source (refresh under a lease)
│   │   ├── liveness/             # per-process liveness keys (Redis) + in-process registry
│   │   ├── mem0memory/            # mem0-backed ADK memory.Service (ADR-0013)
│   │   ├── memoryconn/            # WorkspaceMemoryConfig → mem0 client, per call
│   │   ├── memoryhook/            # Memory Recall + Capture around each turn
│   │   ├── opencode/             # OpenCode HTTP RemoteAgent bridge
│   │   ├── pibox/                 # AGENT_TYPE_PI bridge to a ButterBox (ADR-0011)
│   │   ├── runner/              # Service.Run, InvocationRecorder, CancelInvocation
│   │   ├── runstate/             # AG-UI run state next to the thread lease (Redis + in-process)
│   │   ├── session/mongo/        # CountSessions and ADK event persistence
│   │   ├── sessionguard/          # serializes turns within one session across Pods
│   │   ├── streamorch/            # Shared streaming orchestration
│   │   └── telegram/             # Telegram receiver, poller, worker and router
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
│       ├── agent_operation.proto
│       ├── agent_file.proto
│       ├── agent_service.proto
│       ├── agentchannel.proto
│       ├── api_token.proto
│       ├── auth.proto
│       ├── context.proto
│       ├── cron.proto
│       ├── daemon.proto
│       ├── dashboard.proto
│       ├── automation.proto
│       ├── content.proto
│       ├── butterbox.proto
│       ├── githost.proto
│       ├── linear.proto
│       ├── forum.proto
│       ├── repobinding.proto
│       ├── skill.proto
│       ├── telegram.proto
│       ├── workspace_memory.proto
│       └── workspace.proto
├── .github/
│   └── workflows/
│       ├── buf.yml
│       ├── docker-publish.yml   # backend image → ghcr.io/<owner>/<repo>
│       ├── daemon-publish.yml   # daemon image → ghcr.io/<owner>/<repo>-daemon
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

- `cmd/`：进程入口。`butter` 是服务端；`butter-daemon` 是通过 `/api` ConnectRPC 反连服务端的 daemon client（自报 version / os / executors）。
- `internal/app/`：应用装配与初始化（路由、运行时、配置仓库、渠道、Cron、系统 Agent、token / workspace 仓库选择、初始 admin 与 default workspace bootstrap、Langfuse host 透传）。
- `internal/application/`：RPC 服务实现（Agent/lifecycle、AgentFile、Skill、MCPServer、GlobalMCPServer、ModelProvider、NotifyGroup、RemoteAgent、legacy Channel、Telegram、Linear、Session、Cron、Automation、Dashboard、Daemon、APIToken、Auth、Forum、Workspace、GitHost、ButterBox、WorkspaceRepoBinding）。每个服务一个 `*_service.go`，方法签名是原生 ConnectRPC 形式 `(ctx, *connect.Request[Req]) (*connect.Response[Res], error)`，直接满足 `agentsv1connect.XxxServiceHandler` 接口，由 `routes.go` 通过 `agentsv1connect.NewXxxServiceHandler(svc, ...)` 挂载。错误用 `connect.NewError` 或 `connectx` helper 构造。
- `internal/transport/connectx/`：ConnectRPC 共享 plumbing。`RequiredArgument` / `InvalidArgument` / `NotFound` / `Internal` / `InternalWith` 是 `connect.Error` 的常用构造短手；`HandlerOptions()` 含 snake_case JSON codec（`UseProtoNames=true`）供 curl/非浏览器调用；dashboard 浏览器默认 binary protobuf（`front/src/api/transport.ts`）。
- `internal/workspace/`：workspace context 包，提供 `WithID` / `FromContext` / `HeaderName="X-Workspace-ID"` / `DefaultSlug="default"`。
- `internal/repo/workspace/`：`workspaces` + `workspace_members` 仓库（memory + mongo），支撑 `WorkspaceService` 和 auth middleware 的成员校验。
- `internal/channel/`：只保留 legacy `AgentChannel` 报告器；Telegram 适配与运行时位于 `internal/runtime/telegram/`。
- `internal/runtime/`：运行时能力 —— `runner`（含 invocation 记录与 cancel 注册）、`cron`（含 RunJobNow / 时序聚合）、`daemon`（registry / connection / bridge / grpc_handler / metrics）、`session`、`interrupt`、`streamorch` / `asyncrun`、`automation`、`telegram`、`linear` / `linearconn`、`sessionguard`、`liveness`（进程存活键，供 invocation 遗留清理判断 owner 是否已退出）、`runstate`（AG-UI 运行在 thread 租约旁记录的 run state，只在持有租约期间续期），Workspace Memory 的 `memoryhook` / `mem0memory` / `memoryconn`，以及 ButterBox 的 `pibox` / `cursorbox` / `butterboxconn`。
- Agent 工具：`a2uitool`（`render_ui`）、`aguitool`（AG-UI 客户端工具）、`agentfiletool`、`memorytool`、`skilltool`；A2UI 的 catalog、卡片与表单逻辑在 `internal/a2ui/`。
- `internal/repo/`：仓库层。除 `config/`、`apitoken/`、`invocation/` 外，还包含 Agent Content/lifecycle operation、Agent Files、Skills、Telegram resources/processing/settings、Linear Apps/installations/processing/settings/install states、ButterBoxes、Git hosts/repo bindings/cache、input parts、OAuth、forum、workspace、auth、cryptokey 等 memory/mongo 实现。
- `front/`：Vite + React 19 dashboard。TanStack Router 路由在 `src/routes/`，资源实现位于 `src/features/`；`src/api/` 是类型化的 ConnectRPC 客户端，`uploads.ts` 是唯一仍用裸 `fetch` + multipart 的 API 模块。`src/gen/` 是 buf 生成的 TS proto 类型。
- `proto/`：Proto 定义源文件，包含 Agent/lifecycle、Telegram、Linear、Automation、Skills、Git binding、Daemon、Dashboard、Workspace、Forum、Cron、Content 等服务和消息。
- `pkg/proto/`：Proto 生成代码（Go + Connect + grpc + grpc-gateway + validate）；不要手改。Twirp 生成产物已在 ConnectRPC Phase 3 移除。
- `.github/workflows/`：CI。后端走 `docker-publish.yml`，前端独立 `front-publish.yml`（`paths: front/**` 过滤），均推 ghcr 并 cosign 签名。
- `docs/`：项目文档。系统架构见 `architecture.md`；API 契约与外部 App 开发接入说明见 `api.md`；功能总览见 `app.md`。

## 维护建议

- 新增模块优先放在现有分层下，避免在 `internal/` 根目录继续平铺。
- `pkg/proto/` 与 `front/src/gen/` 均为生成代码目录，手动变更应在 `proto/` 中进行后 `make buf` 重新生成。
- 新增 RPC service：在 `proto/agents/v1/*.proto` 定义 service + messages → `make buf` 生成代码 → 在 `internal/application/` 加 `<svc>_service.go`，每个 RPC 方法签名为 `func (s *XxxServiceServer) Method(ctx context.Context, req *connect.Request[agentsv1.YyyRequest]) (*connect.Response[agentsv1.YyyResponse], error)`，方法体访问 `req.Msg.GetX()`，返回 `connect.NewResponse(&agentsv1.YyyResponse{...})` → 在 `internal/app/routes.go` 用 `agentsv1connect.NewXxxServiceHandler(svc, connectOpts...)` 创建 handler，并挂到 `/api/agents.v1.XxxService/*`（`http.StripPrefix("/api", handler)`）→ 前端在 `front/src/api/` 加文件，`makeClient(XxxService)` 拿到类型化 client。
- 新增 MongoDB collection：在 `internal/repo/` 下加同名子包，提供 interface + memory + mongo 两实现，在 `internal/app/channels.go` 按 `storage_backend` 选择后端并注入 BootstrapResult。
- 结构变更后同步更新本文件，保证文档与仓库一致。
