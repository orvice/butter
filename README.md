# Butter

Butter is a configuration-driven AI agent orchestration service built on the [Butterfly](https://butterfly.orx.me) framework and powered by [Google ADK (Agent Development Kit)](https://google.golang.org/adk). It allows you to define agent workflows via YAML/protobuf config and expose them through various channels like Telegram.

## Features

- **Multi-agent orchestration** — Define LLM, Loop, Sequential, and Parallel agent workflows in config
- **MCP server integration** — Connect agents to external tools via Model Context Protocol (MCP) servers
- **Channel-based delivery** — Bind agents to messaging platforms (currently Telegram) through configurable channels
- **Remote agent support** — Delegate work to remote agents via A2A (Agent-to-Agent) protocol
- **Session management** — Persistent conversation sessions backed by MongoDB
- **Observability** — Built-in OpenTelemetry tracing and optional Langfuse integration
- **Protobuf-defined config** — Agent and channel configurations are defined as protobuf messages, generated via `buf`

## Architecture

```text
┌─────────────────────────────────────────────────┐
│                   cmd/butter                     │
│            (Butterfly app bootstrap)             │
└──────────┬──────────────────────────┬────────────┘
           │                          │
    ┌──────▼──────┐           ┌───────▼───────┐
    │  HTTP Handler│           │   Channels    │
    │  (Gin routes)│           │  (Telegram)   │
    └──────┬──────┘           └───────┬───────┘
           │                          │
    ┌──────▼──────┐           ┌───────▼───────┐
    │   Service    │           │    Runner     │
    │   Layer      │           │  (ADK Agent)  │
    └──────┬──────┘           └───────┬───────┘
           │                          │
    ┌──────▼──────┐           ┌───────▼───────┐
    │    Repo      │           │    Session    │
    │   Layer      │           │   (MongoDB)   │
    └─────────────┘           └───────────────┘
```

### Key Packages

| Package | Description |
|---|---|
| `cmd/butter` | Entry point — wires config, repos, services, handlers, and starts channels |
| `internal/config` | `AppConfig` holding `[]Agent` and `[]AgentChannel` loaded from YAML |
| `internal/agent` | `NewFromProto()` factory — converts proto agent configs into ADK agent instances |
| `internal/channel` | Channel manager and platform adapters (Telegram) |
| `internal/runner` | Executes agent invocations with session and context management |
| `internal/session` | Session storage backends (MongoDB) |
| `internal/handler/http` | Gin HTTP handlers |
| `internal/service` | Business logic layer |
| `internal/repo` | Data access abstractions |
| `pkg/proto/agents/v1` | Generated Go code from protobuf definitions (**do not edit**) |

### Proto Definitions

Located in `proto/agents/v1/`:

- **`agent.proto`** — Agent tree configuration: `Agent`, `AgentConfig`, `LLMAgentConfig`, `MCPServer`, workflow agent configs (Loop, Sequential, Parallel), remote agent config, and context guard settings
- **`agentchannel.proto`** — Platform bindings: `AgentChannel`, triggers, delivery, and Telegram-specific config

## Getting Started

### Prerequisites

- Go 1.25+
- [buf CLI](https://buf.build/) (for protobuf generation)
- MongoDB (for session storage)

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

If `apiToken` is configured in `config.yaml`, API endpoints other than `/ping` require:

```bash
curl -H "Authorization: Bearer <token>" http://127.0.0.1:8080/a2a/<agent-name> 
```

### Configuration

The service is configured via a YAML file pointed to by `BUTTERFLY_CONFIG_FILE_PATH`. The config defines:

- **Agents** — The agent tree: model, instructions, tools (MCP servers), sub-agents, and workflow type
- **Agent Channels** — Bindings between agents and delivery platforms (e.g., Telegram bot token, allowed chats)
- **apiToken** — Optional shared bearer token for protecting HTTP API routes; `/ping` remains public for health checks

### Protobuf Code Generation

```bash
# Generate Go code from proto definitions
buf generate

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

## License

See [LICENSE](LICENSE) for details.
