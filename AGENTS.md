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
Every `Agent`, `AgentChannel`, `MCPServer`, `RemoteAgent`, `ModelProvider`, `CronJob`, `APIToken`, `Invocation`, and `CronExecution` belongs to exactly one workspace. Repo CRUD methods take `workspaceID string` as the first parameter; RPC services derive it from the request context via `internal/workspace.FromContext`. Clients select the active workspace with the `X-Workspace-ID` HTTP header; the auth middleware validates the caller's membership (global admins bypass the check). On startup `application.BootstrapDefaultWorkspace` ensures a `default` workspace exists and adds all known users as owners. Repos also expose `*AcrossWorkspaces` listings used by the runtime layers (runner, channel manager, cron scheduler) that operate on the flat global view — agent names are therefore expected to be unique across workspaces in this iteration.

**Layers:**
- `cmd/butter/main.go` — Entry point. Wires config, services, handlers, and registers Gin routes via Butterfly's `core.New()`.
- `internal/app/` — Application bootstrap and wiring. Split by concern: `routes.go` (HTTP + ConnectRPC route setup), `channels.go` (orchestration), `runtime.go` (MongoDB/Redis/Langfuse init), `cron.go` (scheduler init), `system_agent.go` (built-in agent registration).
- `internal/config/` — `AppConfig` holds `[]agentsv1.Agent` and `[]agentsv1.AgentChannel` loaded from YAML by Butterfly.
- `internal/handler/http/` — Gin HTTP handlers.
- `internal/application/` — RPC service implementations (agent, session, cron, MCP server, remote agent, …). Each service has a `*_service.go` with the business logic plus a `*_connect.go` adapter that exposes it as a ConnectRPC handler via `connectx.WrapUnary`. Implementations currently still construct errors with `twirp.NewError`; the adapter translates them into `connect.Error` on the wire. See `docs/connectrpc-followups.md` for the Phase 3 cleanup that drops the Twirp dependency.
- `internal/transport/connectx/` — Shared ConnectRPC plumbing: `WrapUnary` generic adapter, `TwirpErrorToConnect` code mapping, and the snake_case JSON codec installed via `HandlerOptions()` so the wire format stays compatible with the pre-migration Twirp output.
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

Code generation is configured via `buf.gen.yaml` (outputs to `pkg/proto/`). Plugins: protobuf-go, gRPC, gRPC-Gateway, ConnectRPC, validate, Twirp. (Twirp output is still generated but nothing in runtime code references it; slated for removal — see `docs/connectrpc-followups.md`.)

**Config** is loaded by Butterfly from the YAML file pointed to by `BUTTERFLY_CONFIG_FILE_PATH`. The repository sample is `config.yaml`; deployments may copy it to `config/butter.yaml` or another path. Tracing uses OpenTelemetry (`BUTTERFLY_TRACING_PROVIDER`, `BUTTERFLY_TRACING_ENDPOINT`).

## Documentation

Docs directory layout:

- `docs/api.md` — API reference covering authentication, workspace selection, REST endpoints, ConnectRPC endpoints, and error handling.
- `docs/migration-connectrpc.md` — Plan + status report for the Twirp → ConnectRPC migration (phases 0–2 complete).
- `docs/connectrpc-followups.md` — Outstanding work after the migration: runtime smoke test, Phase 3 Twirp removal, doc/config touch-ups.
- `docs/app.md` — Product/function overview in Chinese, including workspace multi-tenancy, agent orchestration, model management, MCP tools, remote agents, daemon execution, and channel entry points.
- `docs/architecture.md` — System architecture overview covering multi-tenancy, process entry, layered structure, startup wiring, agent construction, and runner execution flow.
- `docs/dashboard-api-gap.md` — Dashboard backend API gap analysis, including current coverage, recommended API extensions, persistence additions, phased implementation, and compatibility notes.
- `docs/design-daemon-agent.md` — Daemon Agent design proposal with background, goals, architecture analysis, core challenges, incremental implementation plan, end-to-end flow, and file change list.
- `docs/project-structure.md` — Project directory structure documentation and maintenance guidance.
- `docs/storage.md` — S3 object storage + static asset / avatar upload configuration and HTTP endpoints.
- `docs/structure-review.md` — Directory structure review with strengths, issues, and refactoring recommendations such as renaming, bootstrap split, and runtime organization.
