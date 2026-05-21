# Butter

Butter is a workspace-aware AI agent orchestration service built on the [Butterfly](https://butterfly.orx.me) framework and powered by [Google ADK (Agent Development Kit)](https://google.golang.org/adk). It lets you manage agent workflows through workspace-scoped runtime APIs, expose them through HTTP/Twirp, chat channels, cron jobs, and daemon-backed remote execution, and operate them from a dashboard.

## Features

- **Multi-agent orchestration** — Manage LLM, Loop, Sequential, and Parallel agent workflows through runtime APIs
- **Workspace multi-tenancy** — Scope agents, channels, MCP servers, remote agents, model providers, cron jobs, API tokens, and invocation history by workspace
- **Runtime dashboard APIs** — Manage auth, workspaces, agents, channels, model providers, notify groups, cron jobs, health/activity feeds, daemon state, and API tokens
- **MCP server integration** — Connect agents to external tools via Model Context Protocol (MCP) servers
- **Channel-based delivery** — Bind agents to messaging platforms (Telegram, Discord) through configurable channels
- **Remote agent support** — Delegate work to remote agents via A2A (Agent-to-Agent) protocol or daemon reverse connections
- **Session management** — Persistent conversation sessions backed by MongoDB
- **Streaming chat API** — Run agents over a Server-Sent Events endpoint for dashboard-style chat UIs
- **Object storage** — Store avatars/static assets and optional ADK artifacts in S3-compatible backends
- **Observability** — Built-in OpenTelemetry tracing and optional Langfuse integration
- **Cron scheduling** — Automated agent execution via configurable cron jobs backed by MongoDB
- **Built-in system agent** — Administrative agent for managing agents and cron jobs, automatically inherits chat model
- **Protobuf-defined API/config** — Agent, channel, notify group, workspace, auth, dashboard, cron, daemon, and token contracts are defined as protobuf messages

## Architecture

```text
┌─────────────────────────────────────────────────────┐
│                    cmd/butter                        │
│         (HTTP/Twirp + channels + daemon gRPC)        │
└──────────┬──────────────────────────┬────────────────┘
           │                          │
    ┌──────▼──────┐           ┌───────▼────────┐
    │ HTTP/Twirp   │           │    Channels    │
    │ SSE/A2A/API  │           │ (TG, Discord)  │
    └──────┬──────┘           └───────┬────────┘
           │                          │
    ┌──────▼──────┐           ┌───────▼────────┐
    │  Application │           │    Runner      │
    │  (Twirp API) │           │  (ADK Agent)   │
    └──────┬──────┘           └───────┬────────┘
           │                          │
    ┌──────▼──────┐           ┌───────▼────────┐
    │ Repositories │           │   Runtime      │
    │ Config/Auth  │           │ Session/Cron   │
    │ Workspace    │           │ Daemon/Memory  │
    └─────────────┘           └────────────────┘
```

### Key Packages

| Package | Description |
|---|---|
| `cmd/butter` | Server entry point — wires config, services, handlers, channels, cron, and daemon gRPC |
| `cmd/butter-daemon` | Daemon client that reverse-connects to the server and executes delegated tasks |
| `internal/app` | Application bootstrap and wiring, split by concern (routes, runtime, channels, cron, system agent) |
| `internal/config` | `AppConfig` for Butterfly/runtime settings plus the flattened runtime config snapshot |
| `internal/agent` | `NewFromProto()` factory — converts proto agent configs into ADK agent instances |
| `internal/agent/system` | Built-in system agent for administrative operations (agent queries, cron management) |
| `internal/application` | Twirp RPC server implementations (auth, workspace, agent, session, cron, dashboard, daemon, API token APIs) |
| `internal/channel` | Channel manager and platform adapters (Telegram, Discord) |
| `internal/runtime/runner` | Executes agent invocations with session, model override, and context management |
| `internal/runtime/cron` | Cron scheduler for automated agent execution |
| `internal/runtime/daemon` | Daemon registry, gRPC bridge, connection tracking, and diagnostics |
| `internal/runtime/session` | Session storage backends (MongoDB) |
| `internal/runtime/memory` | Memory storage backends (MongoDB) |
| `internal/repo` | Repository interfaces and memory/Mongo implementations for config, auth, workspace, API token, and invocation data |
| `internal/handler/http` | Gin HTTP handlers |
| `internal/service` | Business logic layer |
| `pkg/proto/agents/v1` | Generated Go code from protobuf definitions (**do not edit**) |

### Proto Definitions

Located in `proto/agents/v1/`:

- **`agent.proto`** — Agent tree configuration: `Agent`, `AgentConfig`, runtime settings, MCP server config, remote agent config, model providers, notify groups, and context guard settings
- **`agentchannel.proto`** — Platform bindings: `AgentChannel`, triggers, delivery, session binding, and Telegram/Discord config
- **`agent_service.proto`** — Agent/MCP/model provider/remote agent/channel/session RPC services and invocation messages
- **`auth.proto`**, **`workspace.proto`**, **`api_token.proto`** — Dashboard auth, workspace membership, and workspace-bound API token services
- **`cron.proto`**, **`dashboard.proto`**, **`daemon.proto`** — Cron scheduling, dashboard overview/activity APIs, and daemon gRPC/service messages
- **`context.proto`** — Request context metadata propagated through runner, channels, API calls, A2A, MCP, and workspace-aware runtime records

## Getting Started

### Prerequisites

- Go 1.25+
- [buf CLI](https://buf.build/) (for protobuf generation)
- MongoDB (for sessions, auth, config repositories, workspaces, cron/invocation history)
- Redis (for dashboard auth sessions and per-channel active agent/model selection)
- Optional S3-compatible object storage (for avatar/static uploads and ADK artifact persistence)

### Setup

1. **Clone and configure**

```bash
cp .env.example .env
# Edit .env with your settings
```

2. **Install dependencies**

```bash
go mod tidy
```

3. **Run the service**

```bash
export $(grep -v '^#' .env | xargs)
go run ./cmd/butter
```

4. **Verify it's running**

```bash
curl http://127.0.0.1:8080/ping
# {"service":"butter","message":"pong"}
```

If auth repositories, DB API tokens, or `apiToken` are configured, API endpoints other than `/ping` and `AuthService.Login` require Bearer auth. Workspace-scoped RPCs also require `X-Workspace-ID` for user-session and root-token callers; API-token callers use the workspace bound to the token.

```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "X-Workspace-ID: <workspace-id>" \
  -H "Content-Type: application/json" \
  -d '{"page_size":20}' \
  http://127.0.0.1:8080/api/agents.v1.AgentService/ListAgents
```

### Configuration

The service is bootstrapped via a YAML file pointed to by `BUTTERFLY_CONFIG_FILE_PATH`. Persistent agent/runtime configuration is stored in MongoDB by default and managed through Twirp APIs; `storage_backend: memory` is available for local, process-local runs. At runtime, the config store flattens all workspaces into the `AppConfig` snapshot used by the runner, channel manager, and cron scheduler.

The YAML file configures:

- **Initial runtime settings** — HTTP greeting, storage backend, daemon gRPC port, and root `apiToken`
- **Model Providers** — Named provider/model aliases for Gemini, OpenAI-compatible providers, and runtime model override
- **Auth and Workspaces** — Initial admin bootstrap, dashboard sessions, workspace selection via `X-Workspace-ID`
- **MongoDB / Redis** — Persistent repositories, sessions, memory, invocation history, cron history, auth sessions, and channel state
- **Storage** — Mongo/Redis runtime stores, optional S3-backed `static` uploads and `artifact` persistence
- **apiToken** — Optional root bearer token for ops/daemon/API compatibility; `/ping` remains public for health checks

See [docs/api.md](docs/api.md), [docs/app.md](docs/app.md), [docs/architecture.md](docs/architecture.md), and [docs/storage.md](docs/storage.md) for the full API, product, architecture, and object storage details.

### Daemon Client

`cmd/butter-daemon` reverse-connects to the server's daemon gRPC endpoint (default `:9090`) and executes tasks for daemon remote agents:

```bash
go run ./cmd/butter-daemon -config daemon.yaml
```

The daemon uses the root `apiToken` as gRPC `authorization` metadata.

### Protobuf Code Generation

```bash
# Generate Go/TS code and inject custom Go struct tags
make buf

# Lint proto files
buf lint
```

### Running Tests

```bash
go test ./...
```

## Environment Variables

| Variable | Description |
|---|---|
| `BUTTERFLY_CONFIG_TYPE` | Config source type (e.g., `file`) |
| `BUTTERFLY_CONFIG_FILE_PATH` | Path to the YAML config file |
| `BUTTERFLY_TRACING_PROVIDER` | OpenTelemetry tracing provider (`http`) |
| `BUTTERFLY_TRACING_ENDPOINT` | Tracing collector endpoint |

Most runtime settings, including MongoDB, Redis, auth bootstrap, model providers, daemon gRPC port, and S3-backed storage, are configured in `config.yaml`.

## License

See [LICENSE](LICENSE) for details.
