# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Build & Run

```bash
# Install dependencies
go mod tidy

# Run the service (requires env vars, see .env.example)
cp .env.example .env && export $(grep -v '^#' .env | xargs)
go run ./cmd/butter

# Build
make build

# Run tests
go test ./...

# Generate protobuf code and inject custom tags (requires buf CLI and protoc-go-inject-tag)
make buf

# Lint protos
buf lint
```

## Architecture

Module: `go.orx.me/apps/butter`

A service skeleton built on `butterfly.orx.me/core` (Butterfly framework) with an agent system powered by Google ADK (`google.golang.org/adk`).

**Workspaces (multi-tenancy):**
Every `Agent`, `AgentChannel`, `MCPServer`, `RemoteAgent`, `ModelProvider`, `CronJob`, `APIToken`, `Invocation`, and `CronExecution` belongs to exactly one workspace. Repo CRUD methods take `workspaceID string` as the first parameter; Twirp services derive it from the request context via `internal/workspace.FromContext`. Clients select the active workspace with the `X-Workspace-ID` HTTP header; the auth middleware validates the caller's membership (global admins bypass the check). On startup `application.BootstrapDefaultWorkspace` ensures a `default` workspace exists and adds all known users as owners. Repos also expose `*AcrossWorkspaces` listings used by the runtime layers (runner, channel manager, cron scheduler) that operate on the flat global view — agent names are therefore expected to be unique across workspaces in this iteration.

**Layers:**
- `cmd/butter/main.go` — Entry point. Wires config, services, handlers, and registers Gin routes via Butterfly's `core.New()`.
- `cmd/butter-daemon/` — Daemon client that reverse-connects to the server's daemon gRPC endpoint (connector, executor).
- `internal/app/` — Application bootstrap and wiring. Split by concern: `routes.go` (HTTP/Twirp setup), `channels.go` (orchestration), `runtime.go` (MongoDB/Redis/Langfuse init), `cron.go` (scheduler init), `system_agent.go` (built-in agent registration), `config_runtime.go`/`config_store.go` (config repo selection), `grpc.go`.
- `internal/config/` — `AppConfig` holds `[]agentsv1.Agent` and `[]agentsv1.AgentChannel` loaded from YAML by Butterfly.
- `internal/handler/http/` — Gin HTTP handlers (`/ping`, `/a2a`, `/status`, `/api/chat/stream`, `/api/uploads/*`, global MCP preset routes, MCP OAuth callback, API token auth middleware via `internal/authn`).
- `internal/application/` — Twirp RPC server implementations (agent, agent file, session, cron, MCP server, remote agent, model provider, notify group, channel, forum, dashboard, daemon, API token, auth, workspace services).
- `internal/service/` — Business logic (health, status, upload).
- `internal/repo/` — Data access abstractions, each with memory + mongo implementations where applicable: `config/` (workspace-scoped CRUD + `*AcrossWorkspaces` + global MCP presets), `apitoken/`, `auth/` (users in mongo, sessions in redis), `invocation/`, `workspace/`, `forum/`, `agentfile/`, `mcpoauth/`, `oauthstate/`.
- `internal/agent/` — `NewFromProto()` factory: converts proto `agentsv1.Agent` configs into ADK agent instances (LLM, Loop, Sequential, Parallel); also `ProbeMCPServer`.
- `internal/runtime/runner/` — Agent runner service managing per-channel ADK runners (invocation recording, cancel registry).
- `internal/runtime/cron/` — Cron scheduler for automated agent execution.
- `internal/runtime/daemon/` — Daemon runtime: registry, connection, bridge, gRPC handler, metrics.
- `internal/runtime/session/` — Session persistence (MongoDB implementation).
- `internal/runtime/memory/` — Memory persistence (MongoDB implementation).
- `internal/channel/` — Platform channel implementations and channel manager (Telegram, Discord).
- `internal/workspace/` — Workspace context propagation: `WithID` / `FromContext` / `HeaderName` ("X-Workspace-ID").
- `internal/authn/` — Shared auth + workspace resolution for HTTP/Twirp and gRPC/grpc-web (`Resolver`, `HeaderSource`).
- `internal/auth/` + `internal/auth/provider/` — Auth context helpers and OAuth provider registry (GitHub, Google).
- `internal/agentfiletool/`, `internal/mcpoauth/`, `internal/notify/` — Agent file tooling, MCP OAuth, and notifications.
- `pkg/agent/` — Thin wrapper around ADK `agent.Agent`.
- `pkg/proto/agents/v1/` — Generated Go code from protos. **Do not edit.**

**Proto definitions** live in `proto/agents/v1/`:
- `agent.proto` — Agent tree config: `Agent`, `AgentConfig`, `LLMAgentConfig`, `MCPServer`, workflow agent configs (Loop, Sequential, Parallel).
- `agent_service.proto` — Agent management Twirp service.
- `agent_file.proto` — Agent file tooling messages.
- `agentchannel.proto` — Platform bindings: `AgentChannel`, triggers, delivery, Telegram config.
- `api_token.proto` — API token management.
- `auth.proto` — Authentication / user / session messages.
- `context.proto` — Shared context messages.
- `cron.proto` — Cron job and execution definitions.
- `daemon.proto` — Daemon client gRPC contract.
- `dashboard.proto` — Dashboard backend service.
- `forum.proto` — Forum threads/posts service.
- `workspace.proto` — Workspace and membership service.

Code generation is configured via `buf.gen.yaml` (outputs to `pkg/proto/`). Plugins: protobuf-go, gRPC, gRPC-Gateway, ConnectRPC, validate, Twirp.

**Config** is loaded by Butterfly from the YAML file pointed to by `BUTTERFLY_CONFIG_FILE_PATH`. The repository sample is `config.yaml`; deployments may copy it to `config/butter.yaml` or another path. Tracing uses OpenTelemetry (`BUTTERFLY_TRACING_PROVIDER`, `BUTTERFLY_TRACING_ENDPOINT`).

## Documentation

Docs directory layout:

- `docs/api.md` — API reference covering authentication, workspace selection, REST endpoints, Twirp RPC endpoints, and error handling.
- `docs/app.md` — Product/function overview in Chinese, including workspace multi-tenancy, agent orchestration, model management, MCP tools, remote agents, daemon execution, and channel entry points.
- `docs/architecture.md` — System architecture overview covering multi-tenancy, process entry, layered structure, startup wiring, agent construction, and runner execution flow.
- `docs/dashboard-api-gap.md` — Dashboard backend API gap analysis, including current coverage, recommended API extensions, persistence additions, phased implementation, and compatibility notes.
- `docs/design-daemon-agent.md` — Daemon Agent design proposal with background, goals, architecture analysis, core challenges, incremental implementation plan, end-to-end flow, and file change list.
- `docs/frontend-required-apis.md` — Follow-up list of backend semantics/fields the frontend still needs (gaps not yet covered by existing APIs).
- `docs/postgres-migration-analysis.md` — Analysis of current MongoDB/Redis storage usage and candidates for a PostgreSQL backend.
- `docs/project-structure.md` — Project directory structure documentation and maintenance guidance.
- `docs/security-review.md` — Code-level security review with findings grouped by severity (affected file, issue, exploit scenario, recommended fix).
- `docs/storage.md` — S3 object storage + static asset / avatar upload configuration and HTTP endpoints.
- `docs/structure-review.md` — Directory structure review with strengths, issues, and refactoring recommendations such as renaming, bootstrap split, and runtime organization.
