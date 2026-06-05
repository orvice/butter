# Butter

Butter is a workspace-aware AI agent orchestration service built on the
[Butterfly](https://butterfly.orx.me) framework and powered by
[Google ADK](https://google.golang.org/adk). It runs configurable agent
workflows, exposes them through HTTP/RPC and chat channels, and provides a
dashboard-oriented backend for operating agents across workspaces.

## What It Does

- Orchestrates LLM, sequential, parallel, and loop-style agent workflows.
- Scopes agents, model providers, channels, cron jobs, tokens, sessions, and
  execution history by workspace.
- Connects agents to MCP servers, remote agents, chat channels, and daemon
  workers.
- Persists sessions, memory, auth, workspace data, cron jobs, and invocation
  history with MongoDB.
- Uses Redis for dashboard auth sessions and runtime channel selections.
- Supports streaming chat, scheduled runs, object storage, tracing, and optional
  Langfuse integration.

## Getting Started

### Prerequisites

- Go 1.25+
- MongoDB
- Redis
- [buf CLI](https://buf.build/) when regenerating protobuf code
- Optional S3-compatible storage for uploaded assets and ADK artifacts

### Configure

Copy the environment template and point it at the YAML config you want to run:

```bash
cp .env.example .env
```

The sample `.env.example` uses `BUTTERFLY_CONFIG_FILE_PATH=./config/butter.yaml`.
You can either place your runtime config there or change the variable to an
existing file such as `./config.yaml`.

Most service settings live in the YAML config, including auth bootstrap,
storage, tracing, and the optional root `apiToken`. Workspace-scoped agent,
MCP, remote agent, channel, model provider, cron, API token, and daemon
configuration is stored through the runtime config repository.

### Run

```bash
go mod tidy
export $(grep -v '^#' .env | xargs)
go run ./cmd/butter
```

Verify the service is up:

```bash
curl http://127.0.0.1:8080/ping
```

Expected response:

```json
{"service":"butter","message":"pong"}
```

When auth or API tokens are enabled, application endpoints require an
`Authorization: Bearer <token>` header. Workspace-scoped requests should also
send `X-Workspace-ID: <workspace-id>`.

## Daemon Worker

Butter can delegate work to daemon-backed remote agents. Start a daemon client
with:

```bash
go run ./cmd/butter-daemon --url grpc://localhost:9090 --token bt_daemon_runtime_secret
```

Before starting a worker, create a workspace-scoped `DaemonRuntime`, then issue
a runtime token for it. The token is a dedicated `API_TOKEN_KIND_DAEMON` token
with `daemon:connect` scope; it is accepted only by the daemon gRPC endpoint and
cannot call the HTTP API. The token determines the authoritative workspace and
daemon runtime id during registration.

Daemon-backed remote agents choose a runtime plus an ACP runtime (`opencode` or
`codex`). The server sends that choice in each task. The daemon has built-in ACP
profiles for `opencode acp` and `codex-acp`; custom local config can override or
extend those profiles:

```yaml
executors:
  acp:
    - runtime: opencode
      command: opencode
      args: ["acp"]
      permission_policy: deny
      fs:
        read: true
        write: true
      terminal: true
  shell:
    work_dir: /path/to/repo
```

`work_dir` is currently provided by the server per session under `/tmp` for ACP
tasks. Legacy `executors.opencode` config and `capability` in ACP profiles are
still accepted, but new configs should use `executors.acp[].runtime`.

## Development

### Frontend local development

The dashboard frontend lives in [front](front/). For local development, the
recommended setup is to keep `VITE_API_BASE_URL` empty and let the Vite dev
server proxy `/api` and `/ping` to the backend using `VITE_DEV_PROXY_TARGET`:

```env
VITE_API_BASE_URL=
VITE_DEV_PROXY_TARGET=http://localhost:8080
```

Then run:

```bash
cd front
cp .env.example .env.local
npm install
npm run dev
```

See [front/README.md](front/README.md) for the full local proxy behavior and
direct API call alternative.

### Backend development

Run tests:

```bash
go test ./...
```

Regenerate protobuf code when proto definitions change:

```bash
buf generate
```

The `Makefile` also provides:

```bash
make build
make buf
```

## Documentation

- [docs/app.md](docs/app.md) - product and capability overview
- [docs/architecture.md](docs/architecture.md) - system architecture
- [docs/api.md](docs/api.md) - API reference
- [docs/storage.md](docs/storage.md) - object storage, static assets, and artifacts
- [docs/project-structure.md](docs/project-structure.md) - repository layout

## License

See [LICENSE](LICENSE) for details.
