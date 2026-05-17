# Butter API Reference

## Authentication

All endpoints except `GET /ping` require Bearer token authentication:

```
Authorization: Bearer <token>
```

Two token sources are accepted by `APITokenAuthMiddleware`:

1. **Root token** — the single value of `apiToken` in `config.yaml`. Compared with constant-time. Intended for ops / CLI.
2. **DB-stored tokens** — managed at runtime via `APITokenService` (see below). Stored as `sha256` hashes; only the prefix is visible. Successful auth updates `last_used_at` asynchronously.

`401 Unauthorized` on failure.

## REST Endpoints

### Health Check

```
GET /ping
```

Returns server health status. No authentication required.

### Service Status

```
GET /status
```

Returns runtime status for the config storage backend. Requires Bearer token authentication when `apiToken` is configured.

**Response:**

```json
{
  "service": "butter",
  "storage": {
    "configured_backend": "mongo",
    "active_backend": "mongo",
    "persistent": true,
    "mongo_database": "butter",
    "collections": {
      "agents": 3,
      "mcp_servers": 2,
      "remote_agents": 1,
      "channels": 2
    }
  }
}
```

### A2A Protocol

#### Get Agent Card

```
GET /a2a/:agent_name/.well-known/agent.json
```

Returns agent metadata for A2A discovery. Only available for agents with `enable_a2a: true`.

**Response:**

```json
{
  "name": "assistant",
  "description": "A general-purpose assistant",
  "url": "http://host/a2a/assistant",
  "version": "0.1.0",
  "capabilities": {
    "streaming": false
  }
}
```

#### Send Task

```
POST /a2a/:agent_name
```

Sends a task to an agent using JSON-RPC 2.0.

**Request:**

```json
{
  "jsonrpc": "2.0",
  "id": "req-1",
  "method": "tasks/send",
  "params": {
    "id": "task-1",
    "message": {
      "role": "user",
      "parts": [{ "type": "text", "text": "Hello" }]
    }
  }
}
```

**Response:**

```json
{
  "jsonrpc": "2.0",
  "id": "req-1",
  "result": {
    "id": "task-1",
    "status": { "state": "completed" },
    "output": {
      "role": "agent",
      "parts": [{ "type": "text", "text": "Hi there!" }]
    }
  }
}
```

---

## Twirp RPC Endpoints

All Twirp endpoints use `POST` with path pattern `/api/<package>.<Service>/<Method>`.

Request/response bodies are JSON. Content-Type: `application/json`.

---

### AgentService

Manages agent configurations.

#### ListAgents

```
POST /api/agents.v1.AgentService/ListAgents
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `page_size` | int32 | Page size (default 50) |
| `page_token` | string | Opaque token returned by previous response |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `agents` | Agent[] | Page of agents |
| `next_page_token` | string | Token for next page (empty if last) |
| `total` | int32 | Total agents across all pages |

#### GetAgent

```
POST /api/agents.v1.AgentService/GetAgent
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Agent name |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `agent` | Agent | Agent configuration |

#### CreateAgent

```
POST /api/agents.v1.AgentService/CreateAgent
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `agent` | Agent | Agent to create |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `agent` | Agent | Created agent |

#### UpdateAgent

```
POST /api/agents.v1.AgentService/UpdateAgent
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `agent` | Agent | Agent to update (matched by name) |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `agent` | Agent | Updated agent |

#### DeleteAgent

```
POST /api/agents.v1.AgentService/DeleteAgent
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Agent name to delete |

**Response:** `{}`

#### InvokeAgent

```
POST /api/agents.v1.AgentService/InvokeAgent
```

One-shot agent run. If `session_id` is empty an ephemeral id `invoke-<uuid>` is generated.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `agent_name` | string | Required |
| `input` | string | Required input text |
| `app_name` | string | Defaults to `"api"` |
| `user_id` | string | Defaults to `"api"` |
| `session_id` | string | Reuse an existing session; empty creates a new one |
| `model_override` | string | Optional model alias or full name |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | The session id used (echoed back) |
| `response` | string | Final agent response text |

#### CancelAgentInvocation

```
POST /api/agents.v1.AgentService/CancelAgentInvocation
```

Cancels an in-flight invocation by id. The invocation transitions to `FAILED` with a cancellation error.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `invocation_id` | string | Required |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `cancelled` | bool | True if the invocation was found and signalled |

#### ReloadAgents

```
POST /api/agents.v1.AgentService/ReloadAgents
```

Triggers the runtime hot-reload of the agent registry. Same operation that runs implicitly after a config mutation; exposed for the dashboard "Hot-reload" button.

**Request:** `{}`

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `reloaded_at` | timestamp | Server wallclock when the reload returned |

#### GetAgentRuntimeStatus

```
POST /api/agents.v1.AgentService/GetAgentRuntimeStatus
```

Returns the latest invocation state for an agent.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Agent name |

**Response:** `{ "status": AgentRuntimeStatus }` (see below).

#### ListAgentRuntimeStatuses

```
POST /api/agents.v1.AgentService/ListAgentRuntimeStatuses
```

Batched variant; if `names` is empty, returns statuses for all configured agents.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `names` | string[] | Optional filter list |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `statuses` | AgentRuntimeStatus[] | Status for each requested agent |

#### ListAgentInvocations

```
POST /api/agents.v1.AgentService/ListAgentInvocations
```

Returns persisted invocation records, optionally filtered.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `agent_name` | string | Optional filter |
| `session_id` | string | Optional filter |
| `page_size` | int32 | Defaults to 20 |
| `page_token` | string | Pagination token |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `invocations` | Invocation[] | Page of records (newest first) |
| `next_page_token` | string | Empty if last page |
| `total` | int32 | Total matching records |

#### Agent Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique name (required, cannot be "user") |
| `description` | string | Description for LLM delegation |
| `sub_agents` | Agent[] | Nested sub-agents |
| `labels` | map\<string,string\> | Routing/indexing labels |
| `metadata` | map\<string,string\> | Custom annotations |
| `config` | AgentConfig | Execution settings (see below) |
| `type` | enum | `AGENT_TYPE_LLM`, `AGENT_TYPE_LOOP`, `AGENT_TYPE_SEQUENTIAL`, `AGENT_TYPE_PARALLEL` |
| `enable_a2a` | bool | Expose via A2A protocol |

#### AgentConfig Object

| Field | Type | Description |
|-------|------|-------------|
| `model` | string | Model identifier or alias (e.g. "flash", "gemini-2.5-pro") |
| `instruction` | string | Agent behavior instruction |
| `global_instruction` | string | Root-level instruction for entire agent tree |
| `mcp_servers` | MCPServer[] | Inline MCP server configs |
| `mcp_server_ids` | string[] | References to shared MCP servers |
| `remote_agent_ids` | string[] | References to shared remote agents |
| `context_guard` | ContextGuardConfig | Context window management |
| `include_contents` | enum | `LLM_INCLUDE_CONTENTS_DEFAULT`, `LLM_INCLUDE_CONTENTS_NONE` |
| `output_key` | string | Session state key for output |
| `input_schema_json` | string | Input JSON schema |
| `output_schema_json` | string | Output JSON schema |
| `max_iterations` | uint32 | Max loop iterations (LOOP type only) |

#### AgentRuntimeStatus Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Agent name |
| `state` | enum | `AGENT_RUNTIME_STATE_IDLE`, `AGENT_RUNTIME_STATE_RUNNING`, `AGENT_RUNTIME_STATE_FAILED` |
| `last_run_at` | timestamp | Most recent invocation finished_at (or started_at if still running) |
| `last_invocation_id` | string | Id of the most recent invocation |
| `in_flight` | int32 | Number of currently RUNNING invocations for this agent |

#### Invocation Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique id (matches `ContextInfo.uuid`) |
| `agent_name` | string |  |
| `app_name` | string | Channel / app the invocation was triggered from |
| `user_id` | string |  |
| `session_id` | string |  |
| `status` | enum | `INVOCATION_STATUS_RUNNING`, `INVOCATION_STATUS_SUCCEEDED`, `INVOCATION_STATUS_FAILED` |
| `input` | string | Truncated to 4096 chars |
| `output` | string | Truncated to 4096 chars |
| `error` | string | Set when FAILED |
| `started_at` | timestamp |  |
| `finished_at` | timestamp |  |
| `latency_ms` | int64 |  |
| `model_override` | string |  |
| `source` | string | `ContextSource` enum string |

---

### MCPServerService

Manages MCP (Model Context Protocol) server configurations.

#### ListMCPServers

```
POST /api/agents.v1.MCPServerService/ListMCPServers
```

**Request:** `{}`

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `mcp_servers` | MCPServer[] | All MCP servers |

#### GetMCPServer

```
POST /api/agents.v1.MCPServerService/GetMCPServer
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | MCP server ID |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `mcp_server` | MCPServer | MCP server configuration |

#### CreateMCPServer

```
POST /api/agents.v1.MCPServerService/CreateMCPServer
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `mcp_server` | MCPServer | MCP server to create |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `mcp_server` | MCPServer | Created MCP server |

#### UpdateMCPServer

```
POST /api/agents.v1.MCPServerService/UpdateMCPServer
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `mcp_server` | MCPServer | MCP server to update |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `mcp_server` | MCPServer | Updated MCP server |

#### DeleteMCPServer

```
POST /api/agents.v1.MCPServerService/DeleteMCPServer
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | MCP server ID to delete |

**Response:** `{}`

#### GetMCPServerStatus

```
POST /api/agents.v1.MCPServerService/GetMCPServerStatus
```

Probes the configured MCP transport and reports liveness + tool count.

**Request:** `{ "id": "<server-id>" }`

**Response:** `{ "status": MCPServerStatus }`

| MCPServerStatus | Type | Description |
|-------|------|-------------|
| `id` | string |  |
| `name` | string |  |
| `state` | enum | `STATE_CONFIGURED` (STDIO, no probe), `STATE_CONNECTED`, `STATE_DISCONNECTED`, `STATE_ERROR` |
| `tool_count` | int32 | Tools exposed after applying `tool_filter` (or filter size if probing skipped) |
| `detail` | string | Error / context |
| `checked_at` | timestamp |  |

#### ListMCPTools

```
POST /api/agents.v1.MCPServerService/ListMCPTools
```

Enumerates tools across configured MCP servers. STDIO transports are skipped and surfaced in `errors`.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `server_id` | string | Optional. When empty, all servers are probed. |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `tools` | MCPTool[] | Tools with per-tool `allowed` verdict from `tool_filter` |
| `errors` | map\<string,string\> | server_id → probe error |

| MCPTool | Type | Description |
|-------|------|-------------|
| `name` | string |  |
| `description` | string |  |
| `server_id` | string |  |
| `server_name` | string |  |
| `allowed` | bool | True if not filtered out by the server's `tool_filter` |

#### MCPServer Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `name` | string | Server name (required) |
| `transport` | enum | `MCP_SERVER_TRANSPORT_STDIO`, `MCP_SERVER_TRANSPORT_STREAMABLE_HTTP`, `MCP_SERVER_TRANSPORT_SSE` |
| `command` | string | Command for stdio transport |
| `args` | string[] | Command arguments |
| `env` | map\<string,string\> | Environment variables |
| `url` | string | URL for HTTP/SSE transports |
| `headers` | map\<string,string\> | HTTP headers |
| `tool_filter` | string[] | Allowlist of exposed tools |
| `metadata` | map\<string,string\> | Custom metadata |

---

### RemoteAgentService

Manages remote agent configurations (A2A protocol).

#### ListRemoteAgents

```
POST /api/agents.v1.RemoteAgentService/ListRemoteAgents
```

**Request:** `{}`

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `remote_agents` | RemoteAgent[] | All remote agents |

#### GetRemoteAgent

```
POST /api/agents.v1.RemoteAgentService/GetRemoteAgent
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Remote agent ID |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `remote_agent` | RemoteAgent | Remote agent configuration |

#### CreateRemoteAgent

```
POST /api/agents.v1.RemoteAgentService/CreateRemoteAgent
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `remote_agent` | RemoteAgent | Remote agent to create |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `remote_agent` | RemoteAgent | Created remote agent |

#### UpdateRemoteAgent

```
POST /api/agents.v1.RemoteAgentService/UpdateRemoteAgent
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `remote_agent` | RemoteAgent | Remote agent to update |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `remote_agent` | RemoteAgent | Updated remote agent |

#### DeleteRemoteAgent

```
POST /api/agents.v1.RemoteAgentService/DeleteRemoteAgent
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Remote agent ID to delete |

**Response:** `{}`

#### GetRemoteAgentStatus

```
POST /api/agents.v1.RemoteAgentService/GetRemoteAgentStatus
```

Probes the endpoint and reports liveness.

* `REMOTE_AGENT_PROTOCOL_A2A`: HTTP GET `<url>/.well-known/agent.json` (5-second timeout).
* `REMOTE_AGENT_PROTOCOL_DAEMON`: resolves `daemon_capability` against the daemon registry.

**Request:** `{ "id": "<id>" }`

**Response:** `{ "status": RemoteAgentStatus }`

| RemoteAgentStatus | Type | Description |
|-------|------|-------------|
| `id` | string |  |
| `protocol` | enum | Echoed from config |
| `state` | enum | `STATE_CONFIGURED`, `STATE_ACTIVE`, `STATE_IDLE`, `STATE_UNREACHABLE`, `STATE_ERROR` |
| `detail` | string | Error message or extra context |
| `serving_daemon_id` | string | For DAEMON protocol: id of the daemon currently serving the capability |
| `checked_at` | timestamp |  |
| `latency_ms` | int64 | Probe latency (A2A only) |

#### RemoteAgent Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier (required) |
| `name` | string | Human-readable name (required) |
| `url` | string | Endpoint URL (required for A2A) |
| `protocol` | enum | `REMOTE_AGENT_PROTOCOL_A2A`, `REMOTE_AGENT_PROTOCOL_DAEMON` |
| `daemon_capability` | string | Required for DAEMON protocol |

---

### ChannelService

Manages `AgentChannel` configurations (Telegram / Discord pollers) and exposes pause/resume/restart controls.

#### ListChannels

```
POST /api/agents.v1.ChannelService/ListChannels
```

**Request:** `{}`

**Response:** `{ "channels": AgentChannel[] }`

#### GetChannel

```
POST /api/agents.v1.ChannelService/GetChannel
```

**Request:** `{ "name": "<channel-name>" }`

**Response:** `{ "channel": AgentChannel }`

#### CreateChannel / UpdateChannel / DeleteChannel

Standard CRUD; payloads accept `{ "channel": AgentChannel }`. Delete uses `{ "name": "<channel-name>" }`. All mutations reload the channel manager.

#### GetChannelStatus

```
POST /api/agents.v1.ChannelService/GetChannelStatus
```

**Request:** `{ "name": "<channel-name>" }`

**Response:** `{ "status": ChannelStatus }`

| ChannelStatus | Type | Description |
|-------|------|-------------|
| `name` | string |  |
| `platform` | enum | `AGENT_CHANNEL_PLATFORM_TELEGRAM`, `AGENT_CHANNEL_PLATFORM_DISCORD` |
| `state` | enum | `STATE_LIVE`, `STATE_PAUSED`, `STATE_DISABLED`, `STATE_ERROR` |
| `last_poll_at` | timestamp | Server clock at status query (heartbeat not yet plumbed) |
| `detail` | string |  |

#### RestartChannel

```
POST /api/agents.v1.ChannelService/RestartChannel
```

Reloads the channel manager (bounces all pollers). Optional `name` field is echoed back.

**Request:** `{ "name": "<channel-name>"? }`
**Response:** `{ "channel": AgentChannel? }`

#### PauseChannel / ResumeChannel

Toggles the channel's `enabled` flag and reloads. Idempotent.

**Request:** `{ "name": "<channel-name>" }`
**Response:** `{ "channel": AgentChannel }`

---

### SessionService

Manages agent sessions (conversation state).

#### CreateSession

```
POST /api/agents.v1.SessionService/CreateSession
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `app_name` | string | Channel/app name |
| `user_id` | string | User ID |
| `session_id` | string | Custom session ID (auto-generated if empty) |
| `state` | object | Initial session state |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `session` | SessionInfo | Created session |

#### GetSession

```
POST /api/agents.v1.SessionService/GetSession
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `app_name` | string | Channel/app name |
| `user_id` | string | User ID |
| `session_id` | string | Session ID |
| `num_recent_events` | int32 | Max recent events (0 = all) |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `session_detail` | SessionDetail | Session with events |

#### ListSessions

```
POST /api/agents.v1.SessionService/ListSessions
```

Sessions returned newest-first by `last_update_time`. Date filtering is applied server-side.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `app_name` | string | Channel/app name (optional filter) |
| `user_id` | string | User ID (optional filter) |
| `start_time` | timestamp | Only sessions with last_update_time ≥ this |
| `end_time` | timestamp | Only sessions with last_update_time ≤ this |
| `page_size` | int32 | Page size (default 20) |
| `page_token` | string | Opaque base64 cursor returned by previous response |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `sessions` | SessionInfo[] | Matching sessions |
| `next_page_token` | string | Empty if last page |
| `total` | int32 | Total matching sessions across pages |

#### DeleteSession

```
POST /api/agents.v1.SessionService/DeleteSession
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `app_name` | string | Channel/app name |
| `user_id` | string | User ID |
| `session_id` | string | Session ID |

**Response:** `{}`

#### ReplySession

```
POST /api/agents.v1.SessionService/ReplySession
```

Sends a user message to an existing session and returns the agent response.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `agent_name` | string | Agent to invoke |
| `app_name` | string | Channel/app name |
| `user_id` | string | User ID |
| `session_id` | string | Session ID |
| `message` | string | User message text |
| `model_override` | string | Optional model alias or name |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `response` | string | Agent's response text |

#### SessionInfo Object

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | Session identifier |
| `app_name` | string | Channel/app name |
| `user_id` | string | User ID |
| `state` | object | Session state data |
| `last_update_time` | timestamp | Last update time |
| `turn_count` | int32 | Number of recorded events (populated by `GetSession`) |

#### SessionDetail Object

| Field | Type | Description |
|-------|------|-------------|
| `session` | SessionInfo | Session info |
| `events` | SessionEvent[] | Session events |
| `duration` | duration | Wallclock between first and last event |

#### SessionEvent Object

| Field | Type | Description |
|-------|------|-------------|
| `event_id` | string | Unique event ID |
| `invocation_id` | string | Invocation identifier |
| `author` | string | Event author/role |
| `branch` | string | Branch identifier |
| `content_json` | string | Content as JSON (genai.Content) |
| `timestamp` | timestamp | Event timestamp |
| `trace_id` | string | Mirrors `invocation_id`; consumers treat as opaque |
| `trace_url` | string | `<langfuse_host>/trace/<trace_id>` when Langfuse is configured |

---

### CronJobService

Manages scheduled agent execution. Cron jobs are stored in MongoDB.

#### ListCronJobs

```
POST /api/agents.v1.CronJobService/ListCronJobs
```

**Request:** `{}`

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `cron_jobs` | CronJob[] | All cron jobs |

#### GetCronJob

```
POST /api/agents.v1.CronJobService/GetCronJob
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Cron job name |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `cron_job` | CronJob | Cron job configuration |

#### CreateCronJob

```
POST /api/agents.v1.CronJobService/CreateCronJob
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `cron_job` | CronJob | Cron job to create |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `cron_job` | CronJob | Created cron job |

**Example:**

```json
{
  "cron_job": {
    "name": "daily-summary",
    "schedule": "0 9 * * *",
    "agent_name": "assistant",
    "input": "Generate a daily summary",
    "timezone": "Asia/Shanghai",
    "enabled": true,
    "delivery": {
      "type": "CRON_DELIVERY_TYPE_LOG"
    }
  }
}
```

#### UpdateCronJob

```
POST /api/agents.v1.CronJobService/UpdateCronJob
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `cron_job` | CronJob | Cron job to update (matched by name) |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `cron_job` | CronJob | Updated cron job |

#### DeleteCronJob

```
POST /api/agents.v1.CronJobService/DeleteCronJob
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Cron job name to delete |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `cron_job` | CronJob | Deleted cron job |

#### RunCronJobNow

```
POST /api/agents.v1.CronJobService/RunCronJobNow
```

Triggers an immediate execution of the named job, bypassing its schedule. The resulting `CronExecution` is persisted and delivered the same way scheduled runs are.

**Request:** `{ "name": "<job-name>" }`

**Response:** `{ "execution": CronExecution }`

#### ListCronExecutions

```
POST /api/agents.v1.CronJobService/ListCronExecutions
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `job_name` | string | Filter by job name (empty = all) |
| `page_size` | int32 | Max records per page (default 20) |
| `page_token` | string | Pagination token from previous response |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `executions` | CronExecution[] | Execution records (newest first) |
| `next_page_token` | string | Token for next page (empty if last page) |

#### CronJob Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique name (required) |
| `schedule` | string | Cron expression or predefined schedule (required) |
| `agent_name` | string | Agent to execute (required) |
| `input` | string | Message to send to agent |
| `timezone` | string | IANA timezone (default UTC) |
| `enabled` | bool | Whether the job is active |
| `delivery` | CronDelivery | Result delivery config |
| `metadata` | map\<string,string\> | Custom metadata |

**Schedule formats:**
- Standard 5-field: `*/5 * * * *` (every 5 min), `0 9 * * *` (daily at 9:00)
- Predefined: `@every 30m`, `@daily`, `@hourly`, `@weekly`, `@monthly`

#### CronDelivery Object

| Field | Type | Description |
|-------|------|-------------|
| `type` | enum | `CRON_DELIVERY_TYPE_LOG`, `CRON_DELIVERY_TYPE_WEBHOOK`, `CRON_DELIVERY_TYPE_CHANNEL` |
| `webhook_url` | string | URL for WEBHOOK type |
| `channel_name` | string | AgentChannel name for CHANNEL type |
| `chat_id` | string | Target chat ID for CHANNEL type |

#### CronExecution Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique execution ID |
| `job_name` | string | Cron job name |
| `agent_name` | string | Agent that was executed |
| `status` | enum | `CRON_EXECUTION_STATUS_SUCCESS`, `CRON_EXECUTION_STATUS_ERROR` |
| `input` | string | Input message sent |
| `output` | string | Agent output or error message |
| `started_at` | timestamp | Execution start time |
| `finished_at` | timestamp | Execution end time |

---

### DashboardService

Read-only aggregated metrics for the operator dashboard.

#### GetOverview

```
POST /api/agents.v1.DashboardService/GetOverview
```

**Request:** `{ "environment": "production"? }` (echoed back, not yet filtered on)

**Response:** `{ "counts": OverviewCounts, "health": HealthSummary, "latest_daemon_handshake": DaemonHandshake }`

| OverviewCounts | Type | Description |
|-------|------|-------------|
| `active_agents` | int32 |  |
| `mcp_servers` | int32 |  |
| `connected_daemons` | int32 |  |
| `remote_agents` | int32 |  |
| `channels` | int32 |  |
| `cron_jobs` | int32 |  |
| `active_sessions` | int32 | From `mongo session service.CountSessions` |

| HealthSummary | Type | Description |
|-------|------|-------------|
| `mongodb` | ComponentHealth | Real MongoDB ping with latency |
| `redis` | ComponentHealth | Real Redis ping |
| `runner` | ComponentHealth | Runner readiness |

| ComponentHealth | Type | Description |
|-------|------|-------------|
| `status` | enum | `STATUS_HEALTHY`, `STATUS_DEGRADED`, `STATUS_DOWN`, `STATUS_UNSPECIFIED` |
| `detail` | string | Db name or error message |
| `checked_at` | timestamp |  |
| `latency_ms` | int64 | Probe latency |

| DaemonHandshake | Type | Description |
|-------|------|-------------|
| `daemon_id` | string |  |
| `name` | string |  |
| `capabilities` | string[] |  |
| `connected_at` | timestamp |  |
| `os` | string | e.g. `linux-amd64` |

#### GetActivityFeed

```
POST /api/agents.v1.DashboardService/GetActivityFeed
```

Recent invocations rendered as a feed.

**Request:** `{ "limit": 20?, "page_token": ""? }` (limit capped at 200, defaults to 20)

**Response:** `{ "events": ActivityEvent[], "next_page_token": string }`

| ActivityEvent | Type | Description |
|-------|------|-------------|
| `id` | string | Invocation id |
| `kind` | string | `invocation` / `execution_completed` / `error` |
| `actor` | string | Agent name |
| `message` | string | Input or error message |
| `timestamp` | timestamp | finished_at or started_at |

#### GetCronExecutionTimeseries

```
POST /api/agents.v1.DashboardService/GetCronExecutionTimeseries
```

Aggregates `cron_executions` into time buckets for the Overview chart.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `range` | enum | `RANGE_1D` (1-hour buckets, default), `RANGE_7D` (1-day), `RANGE_30D` (1-day) |
| `job_name` | string | Optional filter |

**Response:** `{ "buckets": CronExecutionBucket[] }` — oldest first.

| CronExecutionBucket | Type | Description |
|-------|------|-------------|
| `start` | timestamp | Bucket start (aligned to bucket size) |
| `success` | int32 |  |
| `error` | int32 |  |

---

### DaemonService

Read-only views and control over connected daemons.

#### ListDaemons

```
POST /api/agents.v1.DaemonService/ListDaemons
```

**Response:** `{ "daemons": DaemonStatus[] }`

#### GetDaemon

```
POST /api/agents.v1.DaemonService/GetDaemon
```

**Request:** `{ "daemon_id": "<id>" }`

**Response:** `{ "daemon": DaemonStatus }`

| DaemonStatus | Type | Description |
|-------|------|-------------|
| `daemon_id` | string |  |
| `name` | string |  |
| `capabilities` | string[] |  |
| `labels` | map\<string,string\> |  |
| `state` | enum | `STATE_ONLINE`, `STATE_IDLE`, `STATE_OFFLINE` |
| `connected_at` | timestamp |  |
| `uptime` | duration |  |
| `active_tasks` | int32 |  |
| `version` | string | Reported at handshake |
| `os` | string | e.g. `linux-amd64` |
| `executors` | string[] | Informational executor names |
| `remote_addr` | string | Peer address captured by the server |

#### CancelDaemonTask

```
POST /api/agents.v1.DaemonService/CancelDaemonTask
```

Cancels a task on any connected daemon (or the supplied `daemon_id`). Returns `not_found` if no online daemon is tracking the task.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `task_id` | string | Required |
| `daemon_id` | string | Optional hint |

**Response:** `{ "daemon_id": "<id>" }` (the daemon that received the cancel)

#### ListDaemonTasks

```
POST /api/agents.v1.DaemonService/ListDaemonTasks
```

Tasks currently in flight across connected daemons.

**Request:** `{ "daemon_id": "<id>"? }`

**Response:** `{ "tasks": DaemonTaskInFlight[] }`

| DaemonTaskInFlight | Type | Description |
|-------|------|-------------|
| `task_id` | string |  |
| `daemon_id` | string |  |
| `daemon_name` | string |  |
| `capability` | string | Capability used to dispatch the task |
| `started_at` | timestamp | When the task was dispatched |
| `elapsed` | duration | Wallclock since dispatch |
| `current_step` | string | Latest progress label reported by the daemon |
| `progress` | int32 | 0–100; 0 when unknown |
| `agent_name` | string | Agent that triggered the dispatch |

#### GetBridgeDiagnostics

```
POST /api/agents.v1.DaemonService/GetBridgeDiagnostics
```

Process-level diagnostics for the daemon bridge.

**Request:** `{ "window": "1h"? }` (reserved; currently ignored — server returns a rolling in-memory sample window)

**Response:** `{ "diagnostics": BridgeDiagnostics }`

| BridgeDiagnostics | Type | Description |
|-------|------|-------------|
| `cpu_percent` | double | Avg CPU% since process start (from `runtime/metrics`) |
| `memory_used_bytes` | int64 | `MemStats.Sys` |
| `memory_limit_bytes` | int64 | 0 unless cgroup probe is added |
| `goroutines` | int32 | `runtime.NumGoroutine()` |
| `checked_at` | timestamp |  |
| `latency` | LatencyPoint[] | Rolling 60-sample ring buffer (oldest first) |

| LatencyPoint | Type | Description |
|-------|------|-------------|
| `timestamp` | timestamp |  |
| `latency_ms` | int64 |  |

---

### APITokenService

Manages API bearer tokens. The plaintext secret is only returned at create time; subsequent reads expose the prefix only.

#### ListAPITokens

```
POST /api/agents.v1.APITokenService/ListAPITokens
```

**Response:** `{ "tokens": APIToken[] }`

#### CreateAPIToken

```
POST /api/agents.v1.APITokenService/CreateAPIToken
```

Creates a new token of the form `bt_<48 hex chars>`. The hash is stored; the secret is returned **once**.

**Request:** `{ "name": "<label>" }`

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `token` | APIToken | Persisted token metadata (no secret) |
| `secret` | string | Plaintext secret. **Only returned at create time.** |

#### RevokeAPIToken

```
POST /api/agents.v1.APITokenService/RevokeAPIToken
```

**Request:** `{ "id": "<token-id>" }`

**Response:** `{ "token": APIToken }` — `revoked: true`.

#### APIToken Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Server-assigned identifier |
| `name` | string | Label |
| `prefix` | string | First 12 chars of the secret, for display only |
| `created_at` | timestamp |  |
| `last_used_at` | timestamp | Updated async after successful auth |
| `revoked` | bool | Revoked tokens cannot authenticate |

---

## Error Handling

Twirp endpoints return errors in the standard Twirp format:

```json
{
  "code": "not_found",
  "msg": "agent \"foo\" not found"
}
```

Common error codes:
- `not_found` - Resource does not exist
- `already_exists` - Resource name/ID already taken
- `failed_precondition` - Service not yet initialized
- `internal` - Unexpected server error
