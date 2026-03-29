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
go build ./cmd/butter

# Run tests
go test ./...

# Generate protobuf code (requires buf CLI)
buf generate

# Lint protos
buf lint
```

## Architecture

Module: `go.orx.me/apps/butter`

A service skeleton built on `butterfly.orx.me/core` (Butterfly framework) with an agent system powered by Google ADK (`google.golang.org/adk`).

**Layers:**
- `cmd/butter/main.go` — Entry point. Wires config, repos, services, handlers, and registers Gin routes via Butterfly's `core.New()`.
- `internal/config/` — `AppConfig` holds `[]agentsv1.Agent` and `[]agentsv1.AgentChannel` loaded from YAML by Butterfly.
- `internal/handler/http/` — Gin HTTP handlers.
- `internal/service/` — Business logic.
- `internal/repo/` — Data access abstractions.
- `internal/agent/` — `NewFromProto()` factory: converts proto `agentsv1.Agent` configs into ADK agent instances (LLM, Loop, Sequential, Parallel).
- `pkg/agent/` — Thin wrapper around ADK `agent.Agent`.
- `pkg/proto/agents/v1/` — Generated Go code from protos. **Do not edit.**

**Proto definitions** live in `proto/agents/v1/`:
- `agent.proto` — Agent tree config: `Agent`, `AgentConfig`, `LLMAgentConfig`, `MCPServer`, workflow agent configs (Loop, Sequential, Parallel).
- `agentchannel.proto` — Platform bindings: `AgentChannel`, triggers, delivery, Telegram config.

Code generation is configured via `buf.gen.yaml` (outputs to `pkg/proto/`). Plugins: protobuf-go, gRPC, gRPC-Gateway, ConnectRPC, validate, Twirp.

**Config** is loaded by Butterfly from `config/butter.yaml` via env var `BUTTERFLY_CONFIG_FILE_PATH`. Tracing uses OpenTelemetry (`BUTTERFLY_TRACING_PROVIDER`, `BUTTERFLY_TRACING_ENDPOINT`).
