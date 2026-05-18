# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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
- `internal/app/` — Application bootstrap and wiring. Split by concern: `routes.go` (HTTP/Twirp setup), `channels.go` (orchestration), `runtime.go` (MongoDB/Redis/Langfuse init), `cron.go` (scheduler init), `system_agent.go` (built-in agent registration).
- `internal/config/` — `AppConfig` holds `[]agentsv1.Agent` and `[]agentsv1.AgentChannel` loaded from YAML by Butterfly.
- `internal/handler/http/` — Gin HTTP handlers.
- `internal/application/` — Twirp RPC server implementations (agent, session, cron, MCP server, remote agent services).
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

Code generation is configured via `buf.gen.yaml` (outputs to `pkg/proto/`). Plugins: protobuf-go, gRPC, gRPC-Gateway, ConnectRPC, validate, Twirp.

**Config** is loaded by Butterfly from the YAML file pointed to by `BUTTERFLY_CONFIG_FILE_PATH`. The repository sample is `config.yaml`; deployments may copy it to `config/butter.yaml` or another path. Tracing uses OpenTelemetry (`BUTTERFLY_TRACING_PROVIDER`, `BUTTERFLY_TRACING_ENDPOINT`).
