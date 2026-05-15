# Butter API Reference

## Authentication

All endpoints except `GET /ping` require Bearer token authentication:

```
Authorization: Bearer <token>
```

Token is configured via `apiToken` in `config.yaml`. Returns `401 Unauthorized` on failure.

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

**Request:** `{}`

**Response:**

```json
{
  "agents": [
    {
      "name": "assistant",
      "description": "A general-purpose assistant",
      "type": "AGENT_TYPE_LLM",
      "config": {
        "model": "flash",
        "instruction": "You are a helpful assistant."
      }
    }
  ]
}
```

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

#### RemoteAgent Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier (required) |
| `name` | string | Human-readable name (required) |
| `url` | string | Endpoint URL (required) |
| `protocol` | enum | `REMOTE_AGENT_PROTOCOL_A2A` |

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

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `app_name` | string | Channel/app name |
| `user_id` | string | User ID |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `sessions` | SessionInfo[] | Matching sessions |

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

#### SessionDetail Object

| Field | Type | Description |
|-------|------|-------------|
| `session` | SessionInfo | Session info |
| `events` | SessionEvent[] | Session events |

#### SessionEvent Object

| Field | Type | Description |
|-------|------|-------------|
| `event_id` | string | Unique event ID |
| `invocation_id` | string | Invocation identifier |
| `author` | string | Event author/role |
| `branch` | string | Branch identifier |
| `content_json` | string | Content as JSON (genai.Content) |
| `timestamp` | timestamp | Event timestamp |

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
