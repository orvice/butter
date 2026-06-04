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

Most service settings live in the YAML config, including agents, model
providers, MCP servers, channels, auth bootstrap, storage, daemon settings, and
the optional root `apiToken`.

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
go run ./cmd/butter-daemon -config daemon.yaml
```

The daemon connects back to the server's daemon gRPC endpoint and uses the root
`apiToken` as authorization metadata.

Daemon executors are configured locally. ACP-compatible coding agents such as
opencode should be exposed through the generic ACP executor:

```yaml
server: localhost:9090
token: your-root-api-token
daemon_id: local-dev
name: Local Dev Daemon

executors:
  acp:
    - capability: opencode
      command: opencode
      args: ["acp"]
      work_dir: /path/to/repo
      permission_policy: deny
      fs:
        read: true
        write: true
      terminal: true
  shell:
    work_dir: /path/to/repo
```

Legacy `executors.opencode` config is still accepted and is translated to the
ACP form above (`opencode acp`), but new configs should use `executors.acp`.

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
