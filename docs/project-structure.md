# 项目目录结构文档

更新时间：2026-06-02

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
│   ├── frontend-required-apis.md
│   ├── postgres-migration-analysis.md
│   ├── project-structure.md
│   ├── security-review.md
│   ├── storage.md
│   └── structure-review.md
├── front/
│   ├── Dockerfile
│   ├── nginx.conf
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── api/                 # twirpFetch helpers (one file per service)
│       ├── gen/                 # protoc-gen-es output (@bufbuild/protobuf)
│       │   ├── agents/v1/
│       │   └── validate/
│       ├── hooks/
│       ├── layouts/
│       ├── lib/
│       ├── pages/
│       │   ├── admin/           # global-mcp-servers
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
│   │   └── system/
│   ├── agentfiletool/           # ADK toolset for mounted agent files
│   ├── app/
│   │   ├── channels.go          # bootstrap (mongo, redis, runner, channels, repos)
│   │   ├── config_runtime.go
│   │   ├── config_store.go
│   │   ├── cron.go
│   │   ├── grpc.go
│   │   ├── routes.go            # Twirp + HTTP + auth + global MCP wiring
│   │   ├── runtime.go
│   │   └── system_agent.go
│   ├── application/             # Twirp service implementations
│   │   ├── agent_service.go
│   │   ├── agentfile_service.go
│   │   ├── apitoken_service.go
│   │   ├── auth_service.go
│   │   ├── auth_oauth.go
│   │   ├── channel_service.go
│   │   ├── cron_service.go
│   │   ├── daemon_service.go
│   │   ├── dashboard_service.go
│   │   ├── forum_service.go
│   │   ├── mcpserver_service.go
│   │   ├── modelprovider_service.go
│   │   ├── notifygroup_service.go
│   │   ├── remoteagent_service.go
│   │   ├── runtime_mutation.go
│   │   ├── session_service.go
│   │   └── workspace_service.go
│   ├── auth/                    # auth context helpers
│   │   └── provider/            # GitHub/Google OAuth providers
│   ├── authn/                   # shared HTTP + gRPC/grpc-web auth resolver
│   ├── channel/
│   │   ├── manager.go
│   │   ├── discord/
│   │   └── telegram/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   └── http/                # /ping, /status, /a2a, /api/chat/stream, uploads, auth middleware
│   ├── mcpoauth/                # MCP OAuth discovery, flow, token resolver
│   ├── notify/                  # cron notify-group delivery
│   ├── repo/
│   │   ├── agentfile/           # file space metadata + S3/memory content backends
│   │   ├── apitoken/
│   │   ├── auth/                # users (mongo) + sessions (redis)
│   │   ├── config/              # workspace CRUD + AcrossWorkspaces + global MCP presets
│   │   ├── forum/
│   │   ├── invocation/
│   │   ├── mcpoauth/
│   │   ├── oauthstate/
│   │   ├── workspace/
│   │   └── health.go
│   ├── runtime/
│   │   ├── cron/
│   │   ├── daemon/
│   │   ├── memory/mongo/
│   │   ├── runner/
│   │   └── session/mongo/
│   ├── service/
│   │   ├── health.go
│   │   ├── status.go
│   │   └── upload.go
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
│       ├── agent_file.proto
│       ├── agentchannel.proto
│       ├── api_token.proto
│       ├── auth.proto
│       ├── context.proto
│       ├── cron.proto
│       ├── daemon.proto
│       ├── dashboard.proto
│       ├── forum.proto
│       └── workspace.proto
├── .github/workflows/
│   ├── buf.yml
│   ├── docker-publish.yml
│   ├── front-publish.yml
│   └── go.yml
├── AGENTS.md
├── CLAUDE.md
├── buf.gen.yaml
├── config.yaml
├── Dockerfile
├── go.mod
├── Makefile
└── README.md
```

## 目录说明

- `cmd/`：进程入口。`butter` 是服务端；`butter-daemon` 是通过 gRPC 反连服务端的 daemon client。
- `internal/app/`：应用装配（路由、gRPC、运行时、配置仓库、渠道、Cron、系统 Agent、各 repo 选择与 default workspace bootstrap）。
- `internal/application/`：Twirp 服务实现（Agent / AgentFile / MCPServer / ModelProvider / NotifyGroup / RemoteAgent / Channel / Session / Cron / Forum / Dashboard / Daemon / APIToken / Auth / Workspace）。
- `internal/authn/`：HTTP 与 gRPC/grpc-web 共用的鉴权 + workspace 解析（`Resolver`）。
- `internal/auth/provider/`：Dashboard OAuth provider 注册（GitHub、Google 等）。
- `internal/workspace/`：workspace context 包（`X-Workspace-ID`）。
- `internal/repo/`：持久化抽象。`config/` 含 workspace 配置与 global MCP preset；`auth/` 用户走 MongoDB、session 走 Redis；另有 forum、agentfile、mcpoauth、oauthstate 等集合。
- `internal/channel/`：Telegram、Discord 渠道适配与 manager。
- `internal/runtime/`：`runner`（invocation 记录 + cancel）、`cron`、`daemon`、`session`、`memory`。
- `internal/handler/http/`：REST/A2A/SSE/upload/global MCP/MCP OAuth callback；Twirp 仍走 `/api`。
- `front/`：Vite + React dashboard。`src/api/` 为 twirpFetch wrapper，`src/gen/` 为 buf 生成的 TS proto。
- `proto/`：Proto 源文件；`pkg/proto/` 与 `front/src/gen/` 为生成代码，不要手改。
- `docs/`：项目文档。架构见 `architecture.md`；API 契约见 `api.md`；功能总览见 `app.md`；对象存储见 `storage.md`。

## 维护建议

- 新增模块优先放在现有分层下，避免在 `internal/` 根目录继续平铺。
- `pkg/proto/` 与 `front/src/gen/` 均为生成代码目录，手动变更应在 `proto/` 中进行后 `make buf` 重新生成。
- 新增 Twirp service：proto 定义 → `make buf` → `internal/application/` 加 server → `internal/app/routes.go` 挂载 → 前端 `front/src/api/` 加 wrapper。
- 新增 MongoDB collection：在 `internal/repo/` 下加子包（interface + memory + mongo），在 `internal/app/channels.go` 按 `storage_backend` 注入。
- 结构变更后同步更新本文件、`architecture.md` 与 `AGENTS.md`，保证文档与仓库一致。
