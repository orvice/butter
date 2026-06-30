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
Every `Agent`, `AgentChannel`, `MCPServer`, `RemoteAgent`, `ModelProvider`, `NotifyGroup`, `AgentFileSpace`, `AgentFile`, `ForumThread`, `ForumPost`, `CronJob`, `APIToken`, `Invocation`, and `CronExecution` belongs to exactly one workspace. Repo CRUD methods take `workspaceID string` as the first parameter; RPC services derive it from the request context via `internal/workspace.FromContext`. Clients select the active workspace with the `X-Workspace-ID` HTTP header; the auth middleware validates the caller's membership (global admins bypass the check). `AuthService`, `WorkspaceService`, `DashboardService`, and `DaemonService` do not require a workspace header; `SessionService` session CRUD is app/user/session scoped, but `ReplySession` should include `X-Workspace-ID` so the runner resolves agents in the intended workspace. On startup `application.BootstrapDefaultWorkspace` ensures a `default` workspace exists and adds all known users as owners. Repos also expose `*AcrossWorkspaces` listings used by the runtime layers (runner, channel manager, cron scheduler) that operate on the flat global view — agent names are therefore expected to be unique across workspaces in this iteration.

**Layers:**
- `cmd/butter/main.go` — Entry point. Wires config, services, handlers, and registers Gin routes via Butterfly's `core.New()`.
- `internal/app/` — Application bootstrap and wiring. Split by concern: `routes.go` (HTTP + ConnectRPC route setup), `channels.go` (orchestration), `runtime.go` (MongoDB/Redis/Langfuse init), `cron.go` (scheduler init), `system_agent.go` (built-in agent registration).
- `internal/config/` — `AppConfig` holds `[]agentsv1.Agent` and `[]agentsv1.AgentChannel` loaded from YAML by Butterfly.
- `internal/handler/http/` — Gin HTTP handlers.
- `internal/application/` — RPC service implementations (agent, session, cron, MCP server, remote agent, …). Each service has a `*_service.go` with the business logic. Service methods use native ConnectRPC signatures (`func(ctx, *connect.Request[Req]) (*connect.Response[Res], error)`) and satisfy `agentsv1connect.XxxServiceHandler` directly — `routes.go` hands the service straight to `agentsv1connect.NewXxxServiceHandler(svc, ...)`. Errors are constructed via `connect.NewError` or the `connectx` helpers below.
- `internal/transport/connectx/` — Shared ConnectRPC plumbing: `connect.Error` constructor helpers (`RequiredArgument` / `InvalidArgument` / `NotFound` / `Internal` / `InternalWith`) and the snake_case JSON codec installed via `HandlerOptions()` so the wire format stays compatible with the pre-migration JSON output.
- `internal/service/` — Business logic.
- `internal/repo/` — Data access abstractions.
- `internal/store/config/` — In-memory CRUD store for agent/MCP/remote-agent configurations.
- `internal/agent/` — `NewFromProto()` factory: converts proto `agentsv1.Agent` configs into ADK agent instances (LLM, Loop, Sequential, Parallel).
- `internal/runtime/runner/` — Agent runner service managing per-channel ADK runners.
- `internal/runtime/cron/` — Cron scheduler for automated agent execution.
- `internal/runtime/session/` — Session persistence (MongoDB implementation).
- `internal/runtime/memory/` — Memory persistence (MongoDB implementation).
- `internal/channel/` — Platform channel implementations (Telegram, Discord).
- `pkg/agent/` — Thin wrapper around ADK `agent.Agent`.
- `pkg/proto/agents/v1/` — Generated Go code from protos. **Do not edit.**

**Proto definitions** live in `proto/agents/v1/`:
- `agent.proto` — Agent tree config: `Agent`, `AgentConfig`, `LLMAgentConfig`, `MCPServer`, workflow agent configs (Loop, Sequential, Parallel).
- `agentchannel.proto` — Platform bindings: `AgentChannel`, triggers, delivery, Telegram config.

Code generation is configured via `buf.gen.yaml` (outputs to `pkg/proto/`). Plugins: protobuf-go, gRPC, gRPC-Gateway, ConnectRPC, validate, and bufbuild/es for the frontend. Twirp generation and runtime dependencies were removed in ConnectRPC Phase 3.

**Config** is loaded by Butterfly from the YAML file pointed to by `BUTTERFLY_CONFIG_FILE_PATH`. The repository sample is `config.yaml`; deployments may copy it to `config/butter.yaml` or another path. Tracing uses OpenTelemetry (`BUTTERFLY_TRACING_PROVIDER`, `BUTTERFLY_TRACING_ENDPOINT`).

## Documentation

Docs directory layout:

- `docs/api.md` — App developer API reference and handoff doc: auth, workspace headers, ConnectRPC URL/field conventions, TypeScript Connect-Web examples, REST uploads (`/api/uploads/*`), `AgentService.StreamAgent` chat stream, and errors.
- `docs/migration-connectrpc.md` — Twirp → ConnectRPC migration plan + status (phases 0–3.5, chat `StreamAgent` complete).
- `docs/connectrpc-followups.md` — Post-migration follow-ups (runtime smoke test, wire-format notes).
- `docs/app.md` — Product/function overview in Chinese, including workspace multi-tenancy, agent orchestration, model management, MCP tools, remote agents, daemon execution, and channel entry points.
- `docs/architecture.md` — System architecture overview covering multi-tenancy, process entry, layered structure, startup wiring, agent construction, and runner execution flow.
- `docs/dashboard-api-gap.md` — Dashboard backend API gap analysis, including current coverage, recommended API extensions, persistence additions, phased implementation, and compatibility notes.
- `docs/design-daemon-agent.md` — Daemon Agent design proposal with background, goals, architecture analysis, core challenges, incremental implementation plan, end-to-end flow, and file change list.
- `docs/project-structure.md` — Project directory structure documentation and maintenance guidance.
- `docs/storage.md` — S3 object storage + static asset / avatar upload configuration and HTTP endpoints.
- `docs/structure-review.md` — Directory structure review with strengths, issues, and refactoring recommendations such as renaming, bootstrap split, and runtime organization.

## Agent skills

### Issue tracker

Issues live in GitHub Issues (`orvice/butter`, via the `gh` CLI); external PRs are also a triage surface. See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical roles using default label strings (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.
