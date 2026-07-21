# Butter API Reference

This document is written for app and dashboard developers who need to call
Butter directly. It covers the wire-level conventions that matter when copying
the API into another client project. The canonical schemas are the protobuf
files in `proto/agents/v1/*.proto`; this page explains how those RPCs are
mounted, authenticated, and used from browser/mobile/server apps.

## App Developer Quick Start

### Base URL and RPC URL shape

Use your deployment origin as `BUTTER_BASE_URL`. All ConnectRPC calls are under
`/api`:

```
{BUTTER_BASE_URL}/api/agents.v1.<Service>/<Method>
```

Examples:

```
POST https://butter.example.com/api/agents.v1.AuthService/Login
POST https://butter.example.com/api/agents.v1.AgentService/ListAgents
POST https://butter.example.com/api/agents.v1.AgentService/StreamAgent
```

REST-only endpoints (`/ping`, `/status`, `/a2a/*`, `/api/uploads/*`,
`/api/workspaces/:workspace_id/mcp`) are documented separately below.

### Minimum app flow

1. Call `AuthService.Login` or complete the OAuth flow. These public methods do
   not require a Bearer token.
2. Store `LoginResponse.token` as the app session token.
3. Let the user pick a workspace from `LoginResponse.workspaces`.
4. For workspace-scoped calls, send both:

```
Authorization: Bearer <token>
X-Workspace-ID: <workspace-id>
```

5. Use `AgentService.StreamAgent` for chat streaming, `InvokeAgent` for a
   one-shot response, and `/api/uploads/*` for multipart uploads.

### Request headers

| Header | Required | Use |
|--------|----------|-----|
| `Authorization: Bearer <token>` | Required except public endpoints | Dashboard session token, root token, or API token |
| `X-Workspace-ID: <workspace-id>` | Required for workspace-scoped RPCs when using a user session or root token | Selects the active workspace |
| `Content-Type: application/json` | Required for plain JSON HTTP calls | Connect JSON request body |
| `Content-Type: application/proto` | Used by generated Connect clients with binary protobuf | Smaller/faster payloads; not convenient for curl |
| `Connect-Protocol-Version: 1` | Sent by Connect-Web clients | Allowed by CORS and safe to include |

The server CORS middleware allows `Authorization`, `Content-Type`,
`X-Workspace-ID`, and `Connect-Protocol-Version`.

### Field naming

Generated protobuf clients use their language-native field names. In the
current TypeScript client that means camelCase, for example `agentName` and
`modelOverride`.

Plain JSON HTTP calls should use proto field names, which are snake_case:
`agent_name`, `model_override`, `display_name`, `expires_at`. The server also
accepts camelCase on JSON input for compatibility, but JSON responses use
snake_case.

### Plain JSON examples

Login:

```bash
curl -sS "$BUTTER_BASE_URL/api/agents.v1.AuthService/Login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"secret"}'
```

Response:

```json
{
  "token": "session-token",
  "user": {
    "id": "u-123",
    "username": "admin",
    "display_name": "Admin",
    "role": "admin"
  },
  "expires_at": "2026-06-11T08:00:00Z",
  "workspaces": [
    { "id": "w-123", "name": "Default", "slug": "default" }
  ]
}
```

List agents in a workspace:

```bash
curl -sS "$BUTTER_BASE_URL/api/agents.v1.AgentService/ListAgents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUTTER_TOKEN" \
  -H "X-Workspace-ID: $BUTTER_WORKSPACE_ID" \
  -d '{"page_size":50}'
```

Invoke an agent once:

```bash
curl -sS "$BUTTER_BASE_URL/api/agents.v1.AgentService/InvokeAgent" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BUTTER_TOKEN" \
  -H "X-Workspace-ID: $BUTTER_WORKSPACE_ID" \
  -d '{
    "agent_name": "assistant",
    "input": "Summarize today'\''s channel activity.",
    "app_name": "external-app",
    "user_id": "u-123"
  }'
```

### TypeScript Connect-Web client

Install the Connect runtime and generate TypeScript protobuf code from
`proto/agents/v1/*.proto` with Buf or your preferred protobuf pipeline:

```bash
npm install @connectrpc/connect @connectrpc/connect-web @bufbuild/protobuf
```

Client setup:

```ts
import { Code, ConnectError, createClient, type Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthService } from "./gen/agents/v1/auth_pb";
import { AgentService } from "./gen/agents/v1/agent_service_pb";

const BUTTER_BASE_URL = "https://butter.example.com";

const authInterceptor: Interceptor = (next) => async (req) => {
  const token = localStorage.getItem("butter_token");
  const workspaceId = localStorage.getItem("butter_workspace_id");
  if (token) req.header.set("Authorization", `Bearer ${token}`);
  if (workspaceId) req.header.set("X-Workspace-ID", workspaceId);
  return next(req);
};

const transport = createConnectTransport({
  baseUrl: `${BUTTER_BASE_URL}/api`,
  useBinaryFormat: true,
  interceptors: [authInterceptor],
});

export const authClient = createClient(AuthService, transport);
export const agentClient = createClient(AgentService, transport);

export function isAuthError(err: unknown): boolean {
  return err instanceof ConnectError && err.code === Code.Unauthenticated;
}
```

Login and store the selected workspace:

```ts
const login = await authClient.login({ username: "admin", password: "secret" });
localStorage.setItem("butter_token", login.token);

const workspace = login.workspaces[0];
if (workspace) {
  localStorage.setItem("butter_workspace_id", workspace.id);
}
```

Stream chat:

```ts
const stream = agentClient.streamAgent({
  agentName: "assistant",
  message: "What changed in this workspace?",
  appName: "external-app",
  userId: "u-123",
  sessionId: "",
});

for await (const msg of stream) {
  switch (msg.event.case) {
    case "started":
      console.log("invocation", msg.event.value.invocationId);
      break;
    case "textDelta":
      appendAssistantText(msg.event.value.text);
      break;
    case "runEvent":
      console.log("agent event", msg.event.value.contentJson);
      break;
    case "final":
      console.log("final", msg.event.value.response);
      break;
  }
}
```

Abort/cancel a stream from the browser:

```ts
const controller = new AbortController();
const stream = agentClient.streamAgent(
  { agentName: "assistant", message: "Run a long task" },
  { signal: controller.signal },
);

controller.abort();
```

If you received a `started.invocation_id`, you can also call
`AgentService.CancelAgentInvocation` to cancel server-side work.

## Authentication

All endpoints require Bearer token authentication except `GET /ping`,
`OPTIONS` preflights, the MCP OAuth callback, and the public
`AuthService.Login` / `AuthService.ListOAuthProviders` /
`AuthService.BeginOAuthFlow` / `AuthService.CompleteOAuthFlow` RPCs.

```
Authorization: Bearer <token>
```

Three token sources are accepted by `AuthMiddleware` (tried in order):

1. **Dashboard user session** — issued by `AuthService.Login`. The middleware looks up the hashed token in **Redis** (key `butter:auth:session:<sha256(token)>`) and asynchronously updates `last_used_at`.
2. **Root token** — the single value of `apiToken` in `config.yaml`. Compared with constant-time. Intended for ops / CLI.
3. **DB-stored API tokens** — managed at runtime via `APITokenService` and daemon credential issuance. Stored as `sha256` hashes; only the prefix is visible. User API tokens are `kind=API_TOKEN_KIND_USER` with `api:*` scope and are bound to one workspace. Daemon credentials are `kind=API_TOKEN_KIND_DAEMON` with `daemon:connect` scope and are only accepted by `DaemonConnectorService.Connect`. Successful auth updates `last_used_at` asynchronously.

`401 Unauthorized` on failure.

## Workspace selection

Most product data belongs to a workspace. Clients select the active workspace
via:

```
X-Workspace-ID: <workspace-id>
```

The header is required for most methods on these app-facing services:

| Service | Scope |
|---------|-------|
| `AgentService` | Agent config, invocation, runtime status, chat stream |
| `AgentFileService` | Agent file spaces and files |
| `MCPServerService` | Workspace MCP server config, status, OAuth, tool listing |
| `ModelProviderService` | Workspace model provider config |
| `NotifyGroupService` | Workspace notification groups |
| `RemoteAgentService` | Workspace remote agent config/status |
| `ChannelService` | Workspace channel config/status/control |
| `AutomationService` | Workspace automation definitions, runs, and step runs |
| `CronJobService` | Workspace cron jobs and executions |
| `ForumService` | Workspace forum threads/posts and agent replies |
| `APITokenService` | Tokens are created/listed/revoked within the selected workspace |
| `DaemonService` | Workspace daemon configs, credentials, online daemon/task views |

The header is not required for `AuthService`, `WorkspaceService`,
or `DashboardService`. `SessionService` creates, reads, lists,
and deletes sessions by `app_name` + `user_id` + `session_id`; include
`X-Workspace-ID` when calling `ReplySession` so the runner resolves agents in
the intended workspace. `GlobalMCPServerService` list/create/update/delete are
global/admin operations; `InstallGlobalMCPServer` installs into the current
workspace unless an admin explicitly passes `workspace_id`.

Resolution rules:

- A **user session** must be a member of the workspace identified by `X-Workspace-ID`; global `admin` users bypass the membership check.
- An **API token** ignores any caller-supplied header and uses the workspace bound to the token at creation time.
- The **root token** accepts the `X-Workspace-ID` header verbatim.

If the header is missing on a workspace-scoped RPC the server returns
`failed_precondition` with message
`workspace required (set X-Workspace-ID header)`.

On login, `AuthService.Login` returns the user's accessible workspaces in `LoginResponse.workspaces`; the dashboard uses that list to render the workspace picker before storing the chosen id in `X-Workspace-ID`.

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

### Uploads

All upload endpoints use the same auth middleware as the rest of `/api`.
They require S3-compatible static storage to be enabled with `static.s3_bucket`
in config. See [storage.md](storage.md) for full object storage setup,
authorization details, and response examples.

Uploads intentionally stay as **REST multipart** endpoints (not ConnectRPC):
browsers upload via `FormData`, and the dashboard persists the returned URL
through `AuthService.UpdateProfile` (user avatars) or agent metadata updates
(agent icons). See [storage.md](storage.md) §6.

#### Upload current user avatar

```
POST /api/uploads/avatar
```

Multipart form:

| Field | Required | Description |
|-------|----------|-------------|
| `file` | yes | Avatar image bytes |
| `content_type` | no | Overrides the form part content type; must be PNG, JPEG, GIF, or WebP |

Uploads an avatar for the authenticated user and returns:

```json
{
  "key": "butter/avatars/user/u-123/20260518123045-9f3a1b2c.png",
  "url": "https://cdn.example.com/butter/avatars/user/u-123/20260518123045-9f3a1b2c.png",
  "content_type": "image/png",
  "size": 24580
}
```

#### Upload avatar for owner

```
POST /api/uploads/avatar/:owner_kind/:owner_id
```

Uploads an avatar for another owner. Admins may upload for any owner. Non-admin
users may upload their own `user` avatar and may upload `agent` icons in their
current workspace.

#### Upload static asset

```
POST /api/uploads/static
```

Admin only. Multipart form:

| Field | Required | Description |
|-------|----------|-------------|
| `file` | yes | Asset bytes |
| `name` | no | Object name; defaults to the uploaded filename |
| `content_type` | no | Overrides the form part content type |

Stores the asset under `<static.key_prefix>/static/<name>` and returns the same
response shape as avatar uploads.

### Workspace MCP Endpoint

```
ANY /api/workspaces/:workspace_id/mcp
```

Exposes a workspace-scoped MCP server over HTTP, forwarding requests to the
underlying workspace MCP service. Accepts all HTTP methods (`GET`, `POST`, etc.)
as required by the MCP streamable-HTTP transport. Requires Bearer token auth;
the caller must be a member of `:workspace_id`.

---

## RPC Endpoints

All RPC endpoints are served by [ConnectRPC](https://connectrpc.com) handlers
mounted under `/api/<package>.<Service>/<Method>`. Every endpoint
simultaneously speaks three protocols on the same URL:

- **Connect** — JSON or binary protobuf, the default for the dashboard.
- **gRPC-Web** — `application/grpc-web` / `application/grpc-web+proto`, usable
  from browsers without HTTP/2.
- **gRPC / HTTP/2-compatible Connect streams** — long-lived streams such as
  `DaemonConnectorService.Connect` use the h2c listener at
  `http://localhost:8081/api` for local cleartext HTTP/2, or an HTTPS endpoint
  that preserves HTTP/2 in production.

For the dashboard, **`application/proto` (binary protobuf)** is the canonical
wire format (`useBinaryFormat: true` in `front/src/api/transport.ts`). The
server also accepts Connect JSON and gRPC-Web on the same URLs. When using
JSON, responses use proto field names (snake_case) via `connectx.HandlerOptions()`
(`UseProtoNames=true`); both snake_case and camelCase are accepted on input.

---

### AuthService

Manages dashboard user authentication. `Login` is public; all other methods require a Bearer token. User management methods are admin-only.

#### Login

```
POST /api/agents.v1.AuthService/Login
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `username` | string | Required |
| `password` | string | Required |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `token` | string | Dashboard session token |
| `user` | User | Authenticated dashboard user |
| `expires_at` | timestamp | Session expiry |
| `workspaces` | Workspace[] | Workspaces this user can access |

#### Me

```
POST /api/agents.v1.AuthService/Me
```

**Request:** `{}`

**Response:** `{ "user": User }`

#### Logout

```
POST /api/agents.v1.AuthService/Logout
```

Revokes the current session token.

**Request:** `{}`
**Response:** `{}`

#### ListUsers

```
POST /api/agents.v1.AuthService/ListUsers
```

Admin only.

**Request:** `{}`

**Response:** `{ "users": User[] }`

#### CreateUser

```
POST /api/agents.v1.AuthService/CreateUser
```

Admin only.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `username` | string | Required |
| `password` | string | Required |
| `display_name` | string | Optional |
| `role` | string | User role, e.g. `admin` |
| `disabled` | bool | Whether the user starts disabled |

**Response:** `{ "user": User }`

#### UpdateUserPassword

```
POST /api/agents.v1.AuthService/UpdateUserPassword
```

Admin only.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Required user id |
| `password` | string | Required new password |

**Response:** `{ "user": User }`

#### SetUserDisabled

```
POST /api/agents.v1.AuthService/SetUserDisabled
```

Admin only.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Required user id |
| `disabled` | bool | New disabled state |

**Response:** `{ "user": User }`

#### UpdateProfile

```
POST /api/agents.v1.AuthService/UpdateProfile
```

Updates the authenticated user's display name and avatar. Avatar URLs
are typically produced by `POST /api/uploads/avatar` (see
[storage.md](storage.md)); pass an empty `avatar_url` to clear it.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `display_name` | string | Required, non-empty |
| `avatar_url` | string | Optional. Empty string clears the stored avatar. |

**Response:** `{ "user": User }`

#### ChangePassword

```
POST /api/agents.v1.AuthService/ChangePassword
```

Changes the authenticated user's own password. Requires a password-backed user
session; OAuth-only users may not have a current password hash.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `current_password` | string | Required current password |
| `new_password` | string | Required new password |

**Response:** `{ "user": User }`

Returns `permission_denied` when `current_password` is incorrect.

#### ListOAuthProviders

```
POST /api/agents.v1.AuthService/ListOAuthProviders
```

Public endpoint (no Bearer token required). Returns the OAuth providers
configured on the server (e.g. `github`).

**Request:** `{}`

**Response:** `{ "providers": OAuthProvider[] }`

| OAuthProvider | Type | Description |
|-------|------|-------------|
| `name` | string | Provider key, e.g. `github` |
| `display_name` | string | Human-readable label, e.g. `GitHub` |

#### BeginOAuthFlow

```
POST /api/agents.v1.AuthService/BeginOAuthFlow
```

Public endpoint. Returns the authorization URL to redirect the user to, plus
a `state` token that must be passed back to `CompleteOAuthFlow`.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `provider` | string | Provider key (e.g. `github`) |
| `redirect_uri` | string | Client callback URI stored with the OAuth state |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `authorize_url` | string | Redirect the user here |
| `state` | string | Opaque state token |

#### CompleteOAuthFlow

```
POST /api/agents.v1.AuthService/CompleteOAuthFlow
```

Public endpoint. Exchanges the provider authorization code for a session, mirroring `LoginResponse`.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `provider` | string | Provider key used by `BeginOAuthFlow` |
| `code` | string | Authorization code from provider |
| `state` | string | State token from `BeginOAuthFlow` |

**Response:** Same shape as `Login` response (`token`, `user`, `expires_at`, `workspaces`).

---

| User | Type | Description |
|-------|------|-------------|
| `id` | string | User id |
| `username` | string | Login username |
| `display_name` | string | Display name |
| `avatar_url` | string | Avatar URL (CDN-aware, see [storage.md](storage.md)) |
| `email` | string | Email address, usually populated by OAuth providers |
| `provider` | string | OAuth provider name; empty for password users |
| `external_id` | string | Provider-issued external account id |
| `role` | string | User role |
| `disabled` | bool | Whether login is disabled |
| `created_at` | timestamp | RFC3339 timestamp |
| `updated_at` | timestamp | RFC3339 timestamp |

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

#### StreamAgent

```
POST /api/agents.v1.AgentService/StreamAgent
```

**Server-streaming** RPC for the dashboard chat UI. Replaces the removed
`POST /api/chat/stream` SSE endpoint. The client opens a Connect server stream,
sends one `StreamAgentRequest`, then reads `StreamAgentResponse` messages until
the server closes the stream after `final` or aborts with a `connect.Error`.

Requires the same Bearer token as other `/api` RPCs. Non-admin callers must set
`X-Workspace-ID` so the runner invocation is workspace-scoped (same rule as
`InvokeAgent`).

**Request (`StreamAgentRequest`):**

| Field | Type | Description |
|-------|------|-------------|
| `agent_name` | string | Required |
| `message` | string | User prompt. Required when `parts` is empty; ignored when `parts` is set |
| `app_name` | string | ADK app name; defaults to `"api"` |
| `user_id` | string | ADK user id; defaults to `"api"` |
| `session_id` | string | Reuse an existing session; empty creates `chat-<uuid>` |
| `model_override` | string | Optional model alias or full name |
| `parts` | `InputPart[]` | Multimodal input (text + inline images). When non-empty it is used as the user input and `message` is ignored; when empty, `message` is used as before |

**Multimodal input (`InputPart`, defined in `agents/v1/content.proto`):**

Each `InputPart` sets exactly one of:

| Field | Type | Description |
|-------|------|-------------|
| `text` | string | Plain text content. Max 1 MiB per part (same cap as `message`) |
| `inline_data` | `InlineData` | Inline image: `mime_type` (string) + `data` (bytes; Connect-Web sends a `Uint8Array` with `useBinaryFormat: true`) |

Validation happens in the application layer; violations abort the stream with
`invalid_argument`:

- `mime_type` must be one of `image/jpeg`, `image/png`, `image/gif`, `image/webp`.
- A single image is capped at 10 MiB; a request carries at most 10 images.
- The combined payload of all parts (text + image bytes) is capped at 20 MiB.

Image-only requests (no text part) are accepted.

**Stream messages (`StreamAgentResponse.event` oneof):**

| Variant | Description |
|---------|-------------|
| `started` | First message. `invocation_id`, `session_id`, `agent_name`. Use `invocation_id` with `CancelAgentInvocation`. |
| `text_delta` | Partial assistant text chunk (`text`). |
| `run_event` | Full ADK `session.Event` mirror: `event_id`, `author`, `branch`, `partial`, `final_response`, `content_json`, `timestamp`. |
| `final` | Terminal success. `response` text; server closes the stream after this. |

Terminal failures are **`connect.Error`** on the RPC (e.g. `failed_precondition`
when the runner is unavailable or workspace header is missing), not an in-stream
error payload.

The dashboard calls this via `front/src/api/chat.ts::streamChat` using
`agentClient.streamAgent(...)` with an optional `AbortSignal` for stop/cancel.

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
| `name` | string | Unique name (required, cannot be "user"; must be globally unique across workspaces while the runtime keeps a flat view) |
| `description` | string | Description for LLM delegation |
| `sub_agents` | Agent[] | Nested sub-agents |
| `labels` | map\<string,string\> | Routing/indexing labels |
| `metadata` | map\<string,string\> | Custom annotations |
| `config` | AgentConfig | Execution settings (see below) |
| `type` | enum | `AGENT_TYPE_LLM`, `AGENT_TYPE_LOOP`, `AGENT_TYPE_SEQUENTIAL`, `AGENT_TYPE_PARALLEL`, `AGENT_TYPE_WORKFLOW` |
| `enable_a2a` | bool | Expose via A2A protocol |
| `workspace_id` | string | Owning workspace (server-enforced from `X-Workspace-ID` on writes; returned on reads) |

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
| `file_mounts` | AgentFileMount[] | Agent Files spaces mounted into the built-in `agent_files_*` tools |
| `include_contents` | enum | `LLM_INCLUDE_CONTENTS_DEFAULT`, `LLM_INCLUDE_CONTENTS_NONE` |
| `output_key` | string | Session state key for output |
| `input_schema_json` | string | Input JSON schema |
| `output_schema_json` | string | Output JSON schema |
| `max_iterations` | uint32 | Max loop iterations (LOOP type only) |
| `workflow` | WorkflowConfig | Workflow graph config (WORKFLOW type only) |

#### AgentFileMount Object

| Field | Type | Description |
|-------|------|-------------|
| `space_id` | string | Agent Files space to mount |
| `mount_path` | string | Virtual path exposed to the agent, such as `/docs` |
| `permission` | enum | `AGENT_FILE_MOUNT_PERMISSION_READ`, `AGENT_FILE_MOUNT_PERMISSION_READ_WRITE`, or `AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE` |

#### WorkflowConfig Object

Declares a Workflow Agent's directed graph. See [ADR 0001](adr/0001-workflow-graph-as-nodes-and-edges-proto.md).

| Field | Type | Description |
|-------|------|-------------|
| `nodes` | WorkflowNode[] | Steps of the graph; names must be unique and not `"START"` |
| `edges` | WorkflowEdge[] | Directed connections; use `"START"` in `from` for entry edges |

#### WorkflowNode Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique node name within the graph |
| `kind` | enum | `WORKFLOW_NODE_KIND_AGENT`, `WORKFLOW_NODE_KIND_HUMAN_INPUT`, `WORKFLOW_NODE_KIND_ROUTER`, `WORKFLOW_NODE_KIND_JOIN` |
| `agent` | string | AGENT nodes: name of a sub-agent to run |
| `question` | string | HUMAN_INPUT nodes: the question presented to the human |
| `parallel_worker` | bool | AGENT nodes only: fan-out concurrently over list-typed input |
| `retry` | WorkflowRetryConfig | Retry policy for failed activations |
| `timeout_seconds` | int32 | Per-activation timeout; 0 means no timeout |

#### WorkflowEdge Object

| Field | Type | Description |
|-------|------|-------------|
| `from` | string | Source node name, or `"START"` for entry edges |
| `to` | string | Target node name (must reference a declared node) |
| `route` | string | Route label; empty means the edge always fires |
| `is_default` | bool | Catch-all edge for unmatched Router output; mutually exclusive with `route` |

#### WorkflowRetryConfig Object

| Field | Type | Description |
|-------|------|-------------|
| `max_attempts` | int32 | Max attempts including the first; 0 or 1 means no retries |
| `initial_delay_seconds` | int32 | Delay before the first retry |
| `max_delay_seconds` | int32 | Cap on delay between retries |
| `backoff_factor` | double | Multiplier applied to delay per attempt |

**Workflow Agent example (approval flow with Human Input):**

```json
{
  "agent": {
    "name": "approval",
    "type": "AGENT_TYPE_WORKFLOW",
    "sub_agents": [
      {"name": "draft", "config": {"model": "flash", "instruction": "Draft content."}},
      {"name": "publish", "config": {"model": "flash", "instruction": "Publish content."}}
    ],
    "config": {
      "workflow": {
        "nodes": [
          {"name": "draft", "kind": "WORKFLOW_NODE_KIND_AGENT", "agent": "draft"},
          {"name": "ask", "kind": "WORKFLOW_NODE_KIND_HUMAN_INPUT", "question": "Approve this draft?"},
          {"name": "publish", "kind": "WORKFLOW_NODE_KIND_AGENT", "agent": "publish"}
        ],
        "edges": [
          {"from": "START", "to": "draft"},
          {"from": "draft", "to": "ask"},
          {"from": "ask", "to": "publish"}
        ]
      }
    }
  }
}
```

**Workflow pause/resume:** when a workflow reaches a HUMAN_INPUT node, the turn ends with the question as the reply text. The next plain-text message on the same session is automatically taken as the answer (implicit FIFO resume, [ADR 0002](adr/0002-interrupt-state-derived-from-session-events.md)). This works identically via `StreamAgent`, `ReplySession`, or channel messages. Delete the session to abandon a paused workflow.

**Validation rules:** `CreateAgent`/`UpdateAgent` reject graphs with unknown node references, duplicate node names, AGENT nodes referencing undeclared sub-agents, HUMAN_INPUT nodes without a question, Router nodes without a default outgoing edge, routed/default edges targeting JOIN nodes, and graphs without a START entry edge.

---

### AgentFileService

Manages workspace-scoped Agent Files spaces and text files. Agents can only use
spaces mounted in `Agent.config.file_mounts`; direct service calls still require
the caller's workspace via `X-Workspace-ID`.

Endpoints:

| Method | Path |
|--------|------|
| `ListAgentFileSpaces` | `POST /api/agents.v1.AgentFileService/ListAgentFileSpaces` |
| `GetAgentFileSpace` | `POST /api/agents.v1.AgentFileService/GetAgentFileSpace` |
| `CreateAgentFileSpace` | `POST /api/agents.v1.AgentFileService/CreateAgentFileSpace` |
| `UpdateAgentFileSpace` | `POST /api/agents.v1.AgentFileService/UpdateAgentFileSpace` |
| `DeleteAgentFileSpace` | `POST /api/agents.v1.AgentFileService/DeleteAgentFileSpace` |
| `ListAgentFiles` | `POST /api/agents.v1.AgentFileService/ListAgentFiles` |
| `GetAgentFile` | `POST /api/agents.v1.AgentFileService/GetAgentFile` |
| `WriteAgentFile` | `POST /api/agents.v1.AgentFileService/WriteAgentFile` |
| `DeleteAgentFile` | `POST /api/agents.v1.AgentFileService/DeleteAgentFile` |
| `SearchAgentFiles` | `POST /api/agents.v1.AgentFileService/SearchAgentFiles` |

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
| `workspace_id` | string | Workspace the invocation ran under |

---

### SkillService

Manages workspace-scoped Skills — agentskills.io bundles (a `SKILL.md` plus
optional resource files) that agents attach by listing Skill names in
`Agent.config.skills`. All methods require the caller's workspace via
`X-Workspace-ID`. A Skill is addressed by its spec-validated `name` (unique per
workspace); there is no separate generated ID (ADR 0004).

Endpoints:

| Method | Path |
|--------|------|
| `ListSkills` | `POST /api/agents.v1.SkillService/ListSkills` |
| `GetSkill` | `POST /api/agents.v1.SkillService/GetSkill` |
| `CreateSkill` | `POST /api/agents.v1.SkillService/CreateSkill` |
| `UpdateSkill` | `POST /api/agents.v1.SkillService/UpdateSkill` |
| `DeleteSkill` | `POST /api/agents.v1.SkillService/DeleteSkill` |
| `ListSkillResources` | `POST /api/agents.v1.SkillService/ListSkillResources` |
| `GetSkillResource` | `POST /api/agents.v1.SkillService/GetSkillResource` |
| `PutSkillResource` | `POST /api/agents.v1.SkillService/PutSkillResource` |
| `DeleteSkillResource` | `POST /api/agents.v1.SkillService/DeleteSkillResource` |

**Limits (enforced by the application layer):**

- `SKILL.md` document ≤ 256 KiB by default (configurable via `skills.max_skill_md_bytes`).
- One resource ≤ 10 MiB — fixed, hard-aligned with ADK's per-resource read cap; not configurable.
- ≤ 100 resources per skill by default (configurable via `skills.max_resources_per_skill`). Overwriting an existing path does not count against the cap.

**Resource paths** must fall under `references/`, `assets/`, or `scripts/` after
`path.Clean`. Traversal attempts (`../`, embedded `..`), backslashes, and
absolute paths are rejected with `invalid_argument` on every resource method.

#### Skill Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Spec-validated name; the sole identifier. Must equal the frontmatter `name` |
| `description` | string | Parsed from frontmatter; drives model skill selection |
| `license` | string | Parsed from frontmatter |
| `compatibility` | string | Parsed from frontmatter |
| `metadata` | map<string,string> | Parsed from frontmatter |
| `allowed_tools` | string[] | Parsed from frontmatter; stored, not enforced |
| `size_bytes` | int64 | Size of the stored `SKILL.md` document |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |
| `workspace_id` | string | Owning workspace |

#### SkillResource Object

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Resource path relative to the skill root, e.g. `references/api.md` |
| `size_bytes` | int64 | Content size |
| `content_type` | string | Optional MIME type stored as metadata; not validated |
| `created_at` | timestamp | Stamped once; preserved across overwrites |
| `updated_at` | timestamp | |

#### ListSkills

```
POST /api/agents.v1.SkillService/ListSkills
```

**Request:** `{}`

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `skills` | Skill[] | All skills in the workspace, sorted by name |

#### GetSkill

```
POST /api/agents.v1.SkillService/GetSkill
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Skill name |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `skill` | Skill | Parsed metadata |
| `skill_md` | string | Full `SKILL.md` document (frontmatter + instructions) |

#### CreateSkill

```
POST /api/agents.v1.SkillService/CreateSkill
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Must equal the frontmatter `name` |
| `skill_md` | string | Full `SKILL.md` document |

**Response:** `{ "skill": Skill }`

Errors: `invalid_argument` (missing/oversized `skill_md`, malformed frontmatter,
or a frontmatter name mismatch), `already_exists` (name taken in the workspace).

#### UpdateSkill

```
POST /api/agents.v1.SkillService/UpdateSkill
```

Overwrites in place; the name is immutable, so the frontmatter `name` must match.

**Request:** `{ "name": "<name>", "skill_md": "<full document>" }`

**Response:** `{ "skill": Skill }` — Errors: `invalid_argument`, `not_found`.

#### DeleteSkill

```
POST /api/agents.v1.SkillService/DeleteSkill
```

Cascades to all of the skill's resource content.

**Request:** `{ "name": "<name>" }`

**Response:** `{}` — Errors: `not_found`.

#### ListSkillResources

```
POST /api/agents.v1.SkillService/ListSkillResources
```

Served from the Mongo path index (no content-store reads).

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `skill_name` | string | Skill the resources belong to |

**Response:** `{ "resources": SkillResource[] }` (sorted by path) — Errors: `not_found` (skill absent).

#### GetSkillResource

```
POST /api/agents.v1.SkillService/GetSkillResource
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `skill_name` | string | |
| `path` | string | Resource path, e.g. `references/api.md` |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `resource` | SkillResource | Metadata |
| `content` | bytes | Raw content. **Under the JSON codec this field is base64-encoded** |

Errors: `invalid_argument` (unsafe path), `not_found` (skill or resource absent).

#### PutSkillResource

```
POST /api/agents.v1.SkillService/PutSkillResource
```

Creates or overwrites a resource in place.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `skill_name` | string | |
| `path` | string | Must fall under `references/`, `assets/`, or `scripts/` |
| `content` | bytes | Raw content (base64 under the JSON codec); ≤ 10 MiB |
| `content_type` | string | Optional MIME type stored as metadata |

**Response:** `{ "resource": SkillResource }`

Errors: `invalid_argument` (unsafe path or oversized content), `not_found`
(skill absent), `resource_exhausted` (per-skill count cap reached).

#### DeleteSkillResource

```
POST /api/agents.v1.SkillService/DeleteSkillResource
```

Removes exactly one resource.

**Request:** `{ "skill_name": "<name>", "path": "<resource path>" }`

**Response:** `{}` — Errors: `invalid_argument` (unsafe path), `not_found`.

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
| `state` | enum | `STATE_CONFIGURED`, `STATE_CONNECTED`, `STATE_DISCONNECTED`, `STATE_ERROR` |
| `tool_count` | int32 | Tools exposed after applying `tool_filter` |
| `detail` | string | Error / context |
| `checked_at` | timestamp |  |

#### ListMCPTools

```
POST /api/agents.v1.MCPServerService/ListMCPTools
```

Enumerates tools across configured MCP servers.

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

#### StartMCPServerOAuth

```
POST /api/agents.v1.MCPServerService/StartMCPServerOAuth
```

Prepares a workspace-scoped OAuth2 authorization flow for an MCP server.
Returns the authorization URL to redirect the user to, plus a `flow_id` that
the frontend must hold until the user returns via the OAuth callback.

**Request:** `{ "mcp_server_id": "<server-id>" }`

**Response:** `{ "authorization_url": string, "flow_id": string }`

#### CompleteMCPServerOAuth

```
POST /api/agents.v1.MCPServerService/CompleteMCPServerOAuth
```

Exchanges an OAuth code+state (after the callback redirects) for a stored token
and persists the connection. The HTTP callback `GET /api/mcp/oauth/callback`
internally invokes this; clients may also call it directly.

**Request:** `{ "flow_id": string, "code": string }`

**Response:** `{ "status": MCPOAuthConnectionStatus }`

#### GetMCPServerOAuthStatus

```
POST /api/agents.v1.MCPServerService/GetMCPServerOAuthStatus
```

**Request:** `{ "mcp_server_id": "<server-id>" }`

**Response:** `{ "status": MCPOAuthConnectionStatus }`

| MCPOAuthConnectionStatus | Type | Description |
|-------|------|-------------|
| `server_id` | string | |
| `connected` | bool | Whether a valid token is stored |
| `expires_at` | timestamp | Token expiry if known |
| `detail` | string | Error or status message |

#### DisconnectMCPServerOAuth

```
POST /api/agents.v1.MCPServerService/DisconnectMCPServerOAuth
```

Revokes and deletes the stored OAuth token for the server.

**Request:** `{ "mcp_server_id": "<server-id>" }`

**Response:** `{}`

---

#### MCPServer Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier within a workspace |
| `name` | string | Server name (required) |
| `transport` | enum | `MCP_SERVER_TRANSPORT_STREAMABLE_HTTP`, `MCP_SERVER_TRANSPORT_SSE` |
| `url` | string | URL for HTTP/SSE transports |
| `headers` | map\<string,string\> | HTTP headers |
| `tool_filter` | string[] | Allowlist of exposed tools |
| `metadata` | map\<string,string\> | Custom metadata |
| `workspace_id` | string | Owning workspace |

---

### GlobalMCPServerService

Workspace-agnostic MCP server presets that admins curate and any
authenticated user can install into their workspace. List and install do
**not** require `X-Workspace-ID` (install resolves the target from the
header if present, or from the request body for admin cross-workspace
installs). Create / Update / Delete require role `admin`.

For non-admin callers the `auth.oauth2.client_secret` field is zeroed in
responses; the secret is still persisted server-side so the OAuth flow
can complete when the preset is installed into a workspace.

#### ListGlobalMCPServers

```
POST /api/agents.v1.GlobalMCPServerService/ListGlobalMCPServers
```

**Request:** `{}`

**Response:** `{ "mcp_servers": MCPServer[] }`

#### CreateGlobalMCPServer  *(admin only)*

```
POST /api/agents.v1.GlobalMCPServerService/CreateGlobalMCPServer
```

**Request:** `{ "mcp_server": MCPServer }` — `workspace_id` on the body
is ignored (presets are workspace-agnostic).

**Response:** `{ "mcp_server": MCPServer }`

#### UpdateGlobalMCPServer  *(admin only)*

```
POST /api/agents.v1.GlobalMCPServerService/UpdateGlobalMCPServer
```

**Request:** `{ "mcp_server": MCPServer }` — `mcp_server.id` is required.

**Response:** `{ "mcp_server": MCPServer }`

#### DeleteGlobalMCPServer  *(admin only)*

```
POST /api/agents.v1.GlobalMCPServerService/DeleteGlobalMCPServer
```

**Request:** `{ "id": "<preset-id>" }`

**Response:** `{}`

#### InstallGlobalMCPServer

```
POST /api/agents.v1.GlobalMCPServerService/InstallGlobalMCPServer
```

Clones the preset into a workspace and tags the resulting MCP server
with the source preset id (in `metadata`). Admins may install across
workspaces by passing `workspace_id`; the server audit-logs cross-tenant
installs.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Preset id (required) |
| `workspace_id` | string | Optional. Defaults to the `X-Workspace-ID` header. Admin only when it differs from the active workspace. |

**Response:** `{ "mcp_server": MCPServer }` — workspace-scoped clone with
`client_secret` redacted in the response.

---

### ModelProviderService

Manages LLM provider configurations. Requires `X-Workspace-ID` for user-session and root-token callers; API-token callers use the token's workspace.

#### ListModelProviders

```
POST /api/agents.v1.ModelProviderService/ListModelProviders
```

**Request:** `{}`

**Response:** `{ "model_providers": ModelProvider[] }`

#### GetModelProvider

```
POST /api/agents.v1.ModelProviderService/GetModelProvider
```

**Request:** `{ "name": "<provider-name>" }`

**Response:** `{ "model_provider": ModelProvider }`

#### CreateModelProvider

```
POST /api/agents.v1.ModelProviderService/CreateModelProvider
```

**Request:** `{ "model_provider": ModelProvider }`

**Response:** `{ "model_provider": ModelProvider }`

#### UpdateModelProvider

```
POST /api/agents.v1.ModelProviderService/UpdateModelProvider
```

**Request:** `{ "model_provider": ModelProvider }`

**Response:** `{ "model_provider": ModelProvider }`

#### DeleteModelProvider

```
POST /api/agents.v1.ModelProviderService/DeleteModelProvider
```

**Request:** `{ "name": "<provider-name>" }`

**Response:** `{}`

#### ModelProvider Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Human-readable provider name, e.g. `openai` |
| `type` | string | Provider backend, e.g. `gemini` or `openai` |
| `api_key` | string | API key or env-injected value |
| `base_url` | string | Base URL for OpenAI-compatible endpoints |
| `models` | ModelConfig[] | Models served by this provider |
| `workspace_id` | string | Owning workspace |

| ModelConfig | Type | Description |
|-------|------|-------------|
| `name` | string | Provider model identifier |
| `alias` | string | Short alias used by agents and channels |

---

### NotifyGroupService

Manages outbound notification groups for cron job delivery. Requires `X-Workspace-ID` for user-session and root-token callers; API-token callers use the token's workspace.

Endpoints:

| RPC | Path | Request | Response |
|-----|------|---------|----------|
| `ListNotifyGroups` | `POST /api/agents.v1.NotifyGroupService/ListNotifyGroups` | `{}` | `{ "notify_groups": NotifyGroup[] }` |
| `GetNotifyGroup` | `POST /api/agents.v1.NotifyGroupService/GetNotifyGroup` | `{ "name": "<group-name>" }` | `{ "notify_group": NotifyGroup }` |
| `CreateNotifyGroup` | `POST /api/agents.v1.NotifyGroupService/CreateNotifyGroup` | `{ "notify_group": NotifyGroup }` | `{ "notify_group": NotifyGroup }` |
| `UpdateNotifyGroup` | `POST /api/agents.v1.NotifyGroupService/UpdateNotifyGroup` | `{ "notify_group": NotifyGroup }` | `{ "notify_group": NotifyGroup }` |
| `DeleteNotifyGroup` | `POST /api/agents.v1.NotifyGroupService/DeleteNotifyGroup` | `{ "name": "<group-name>" }` | `{}` |

#### NotifyGroup Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique group name |
| `enabled` | bool | Disabled groups are skipped by cron delivery |
| `targets` | NotifyTarget[] | Telegram, Lark webhook, or Discord webhook targets |
| `metadata` | map<string,string> | Optional metadata |
| `workspace_id` | string | Owning workspace |

`NotifyTarget.type` is one of `NOTIFY_TARGET_TYPE_TELEGRAM`, `NOTIFY_TARGET_TYPE_LARK_WEBHOOK`, or `NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK`. Target-specific config is stored under `telegram`, `lark`, or `discord`.

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
* `REMOTE_AGENT_PROTOCOL_DAEMON`: resolves `daemon_runtime_id` against the daemon registry and dispatches with `acp_runtime`.

**Request:** `{ "id": "<id>" }`

**Response:** `{ "status": RemoteAgentStatus }`

| RemoteAgentStatus | Type | Description |
|-------|------|-------------|
| `id` | string |  |
| `protocol` | enum | Echoed from config |
| `state` | enum | `STATE_CONFIGURED`, `STATE_ACTIVE`, `STATE_IDLE`, `STATE_UNREACHABLE`, `STATE_ERROR` |
| `detail` | string | Error message or extra context |
| `serving_daemon_runtime_id` | string | For DAEMON protocol: id of the runtime currently serving the agent |
| `checked_at` | timestamp |  |
| `latency_ms` | int64 | Probe latency (A2A only) |

#### RemoteAgent Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier within a workspace (required) |
| `name` | string | Human-readable name (required) |
| `url` | string | Endpoint URL (required for A2A) |
| `protocol` | enum | `REMOTE_AGENT_PROTOCOL_A2A`, `REMOTE_AGENT_PROTOCOL_DAEMON` |
| `daemon_runtime_id` | string | Required for DAEMON protocol |
| `acp_runtime` | string | Required for DAEMON protocol; v1 supports `opencode` and `codex` |
| `workspace_id` | string | Owning workspace |

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

#### CreateChannel

```
POST /api/agents.v1.ChannelService/CreateChannel
```

Creates an `AgentChannel` configuration and reloads the channel manager.

**Request:** `{ "channel": AgentChannel }`

**Response:** `{ "channel": AgentChannel }`

#### UpdateChannel

```
POST /api/agents.v1.ChannelService/UpdateChannel
```

Updates an `AgentChannel` configuration and reloads the channel manager.

**Request:** `{ "channel": AgentChannel }`

**Response:** `{ "channel": AgentChannel }`

#### DeleteChannel

```
POST /api/agents.v1.ChannelService/DeleteChannel
```

Deletes an `AgentChannel` configuration and reloads the channel manager.

**Request:** `{ "name": "<channel-name>" }`

**Response:** `{}`

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

#### PauseChannel

```
POST /api/agents.v1.ChannelService/PauseChannel
```

Disables the channel and reloads the channel manager. Idempotent.

**Request:** `{ "name": "<channel-name>" }`
**Response:** `{ "channel": AgentChannel }`

#### ResumeChannel

```
POST /api/agents.v1.ChannelService/ResumeChannel
```

Enables the channel and reloads the channel manager. Idempotent.

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
| `message` | string | User message text. Ignored when `parts` is set; max 1 MiB (same cap as `StreamAgent.message`) |
| `model_override` | string | Optional model alias or name |
| `parts` | `InputPart[]` | Multimodal input (text + inline images). Same semantics and validation as `StreamAgent.parts` (see the `InputPart` table in the StreamAgent section); violations return `invalid_argument` |

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

### AutomationService

Manages workspace-scoped automation workflows. An automation has one trigger,
optional conditions, ordered steps, and an execution policy. Manual and schedule
triggers execute in the current implementation; webhook/forum/channel/daemon
trigger shapes are reserved for follow-up event routing.

#### ListAutomations

```
POST /api/agents.v1.AutomationService/ListAutomations
```

**Request:** `{}`

**Response:** `{ "automations": Automation[] }`

#### GetAutomation

```
POST /api/agents.v1.AutomationService/GetAutomation
```

**Request:** `{ "name": "<automation-name>" }`

**Response:** `{ "automation": Automation }`

#### CreateAutomation / UpdateAutomation

```
POST /api/agents.v1.AutomationService/CreateAutomation
POST /api/agents.v1.AutomationService/UpdateAutomation
```

**Request:** `{ "automation": Automation }`

**Response:** `{ "automation": Automation }`

Schedule-triggered automations are registered, rescheduled, or unscheduled as
part of create/update/delete. Disabled automations are persisted but not
scheduled.

**Example:**

```json
{
  "automation": {
    "name": "daily-summary-workflow",
    "enabled": true,
    "trigger": {
      "type": "AUTOMATION_TRIGGER_TYPE_SCHEDULE",
      "schedule": {
        "schedule": "0 9 * * *",
        "timezone": "Asia/Shanghai"
      }
    },
    "steps": [
      {
        "name": "summarize",
        "type": "AUTOMATION_STEP_TYPE_INVOKE_AGENT",
        "invoke_agent": {
          "agent_name": "assistant",
          "input": "Generate today's workspace summary."
        }
      },
      {
        "name": "notify",
        "type": "AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP",
        "send_notify_group": {
          "notify_group_name": "ops",
          "title": "Daily summary",
          "message": "The daily summary automation completed."
        }
      }
    ],
    "policy": {
      "timeout": "300s",
      "retry": { "max_attempts": 1, "backoff": "5s" },
      "concurrency": "AUTOMATION_CONCURRENCY_POLICY_SKIP",
      "max_output_bytes": 4096
    }
  }
}
```

#### DeleteAutomation

```
POST /api/agents.v1.AutomationService/DeleteAutomation
```

**Request:** `{ "name": "<automation-name>" }`

**Response:** `{ "automation": Automation }`

#### RunAutomationNow

```
POST /api/agents.v1.AutomationService/RunAutomationNow
```

Runs an enabled automation immediately with trigger type `MANUAL`. Conditions
are evaluated before steps; failed conditions produce a skipped run and no step
runs.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Automation name |
| `trigger_payload_json` | string | Optional JSON payload stored as a truncated preview and available to conditions |

**Response:** `{ "run": AutomationRun }`

#### ListAutomationRuns

```
POST /api/agents.v1.AutomationService/ListAutomationRuns
```

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `automation_name` | string | Filter by automation name (empty = all) |
| `page_size` | int32 | Max records per page |
| `page_token` | string | Pagination token from previous response |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| `runs` | AutomationRun[] | Run records (newest first) |
| `next_page_token` | string | Token for next page |

#### GetAutomationRun

```
POST /api/agents.v1.AutomationService/GetAutomationRun
```

**Request:** `{ "id": "<run-id>" }`

**Response:** `{ "run": AutomationRun }`

#### ListAutomationStepRuns

```
POST /api/agents.v1.AutomationService/ListAutomationStepRuns
```

**Request:** `{ "run_id": "<run-id>" }`

**Response:** `{ "step_runs": AutomationStepRun[] }` ordered by step order.

#### Automation Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique automation name within a workspace |
| `enabled` | bool | Disabled automations are persisted but not scheduled or manually runnable |
| `trigger` | AutomationTrigger | Manual, schedule, or reserved event trigger configuration |
| `conditions` | AutomationCondition[] | All conditions must pass before steps execute |
| `steps` | AutomationStep[] | Ordered linear workflow steps |
| `policy` | AutomationPolicy | Run-level timeout/retry/concurrency/output policy |
| `metadata` | map\<string,string\> | Custom metadata |
| `created_at` / `updated_at` | timestamp | Server-managed timestamps |
| `workspace_id` | string | Owning workspace |

#### AutomationTrigger Object

| Field | Type | Description |
|-------|------|-------------|
| `type` | enum | `MANUAL`, `SCHEDULE`, plus reserved `WEBHOOK`, `FORUM_EVENT`, `CHANNEL_EVENT`, `DAEMON_EVENT` |
| `schedule.schedule` | string | Cron expression or predefined schedule for schedule triggers |
| `schedule.timezone` | string | IANA timezone; defaults to UTC |
| `webhook.token` | string | Reserved stable token for webhook routing |
| `event.event_type` | string | Reserved event kind such as `forum.post.created` |
| `event.filters` | map\<string,string\> | Reserved event filters |

#### AutomationCondition Object

| Field | Type | Description |
|-------|------|-------------|
| `selector` | string | Dot selector such as `payload.kind` |
| `operator` | enum | `EQUALS`, `NOT_EQUALS`, `CONTAINS`, `REGEX_MATCH`, `EXISTS`, `NOT_EXISTS` |
| `value` | string | Comparison value when required by the operator |

#### AutomationStep Object

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Stable step name recorded in step-run history |
| `type` | enum | `INVOKE_AGENT`, `CALL_WEBHOOK`, `SEND_NOTIFY_GROUP`, `CREATE_FORUM_POST` |
| `invoke_agent` | object | `agent_name`, `input`, optional `model_override` |
| `call_webhook` | object | `url`, `method`, `payload_json`, `headers` |
| `send_notify_group` | object | `notify_group_name`, `title`, `message` |
| `create_forum_post` | object | `thread_id`, `body` |
| `policy` | AutomationPolicy | Optional step-level overrides |

#### AutomationPolicy Object

| Field | Type | Description |
|-------|------|-------------|
| `timeout` | duration | Run/step timeout; 0 uses server default behavior |
| `retry.max_attempts` | int32 | Retries after the initial attempt |
| `retry.backoff` | duration | Backoff between attempts |
| `concurrency` | enum | `SKIP`, `QUEUE`, `REPLACE`, `ALLOW`; `SKIP` preserves non-overlap defaults |
| `max_output_bytes` | int32 | Stored output preview cap; 0 uses server default |

#### AutomationRun Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique run ID |
| `automation_name` | string | Automation that ran |
| `trigger_type` | enum | Trigger that started the run |
| `status` | enum | `RUNNING`, `SUCCEEDED`, `FAILED`, `SKIPPED`, `CANCELLED` |
| `trigger_payload_json` | string | Truncated trigger payload preview |
| `error` | string | Failure reason |
| `started_at` / `finished_at` | timestamp | Run timing |
| `duration_ms` | int64 | Wallclock duration |
| `workspace_id` | string | Owning workspace |

#### AutomationStepRun Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique step-run ID |
| `run_id` | string | Parent automation run |
| `automation_name` | string | Automation that owns the step |
| `step_name` | string | Step name from the definition |
| `step_type` | enum | Action type |
| `status` | enum | `RUNNING`, `SUCCEEDED`, `FAILED`, `SKIPPED`, `CANCELLED` |
| `attempt_count` | int32 | Attempts used for this step |
| `input_json` / `output_json` | string | Stored input/output previews |
| `error` | string | Failure reason |
| `invocation_id` | string | Agent invocation ID when the step invoked an agent |
| `started_at` / `finished_at` | timestamp | Step timing |
| `duration_ms` | int64 | Wallclock duration |
| `order` | int32 | Step order in the workflow |
| `truncated` | bool | Whether output was truncated by policy |
| `workspace_id` | string | Owning workspace |

---

### CronJobService

Manages scheduled agent execution. Cron jobs are stored in MongoDB and remain
backward compatible with older clients that omit the reliability fields.

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
    },
    "timeout": "120s",
    "retry": {
      "max_attempts": 2,
      "backoff": "10s"
    },
    "concurrency_policy": "CRON_CONCURRENCY_POLICY_SKIP",
    "notify_on": "CRON_NOTIFY_ON_FAILURE",
    "max_output_bytes": 4096
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
| `name` | string | Unique name within a workspace (required) |
| `schedule` | string | Cron expression or predefined schedule (required) |
| `agent_name` | string | Agent to execute (required; resolved by name globally) |
| `input` | string | Message to send to agent |
| `timezone` | string | IANA timezone (default UTC) |
| `enabled` | bool | Whether the job is active |
| `delivery` | CronDelivery | Result delivery config |
| `timeout` | duration | Per-execution timeout; 0 means no cron-level timeout |
| `retry` | CronRetryPolicy | Retry attempts and backoff for failed invocations |
| `concurrency_policy` | enum | `SKIP`, `QUEUE`, `REPLACE`, `ALLOW`; default preserves previous skip-overlap behavior. Under `SKIP` and `QUEUE`, a job with a `WAITING_INPUT` execution also skips new triggers (recorded as `SKIPPED`) until the pending question is answered or abandoned — queueing behind a human answer would be unbounded |
| `notify_on` | enum | `ALWAYS`, `FAILURE`, `SUCCESS`; controls non-log delivery |
| `max_output_bytes` | int32 | Stored output preview cap; 0 uses server default |
| `metadata` | map\<string,string\> | Custom metadata |
| `workspace_id` | string | Owning workspace |

**Schedule formats:**
- Standard 5-field: `*/5 * * * *` (every 5 min), `0 9 * * *` (daily at 9:00)
- Predefined: `@every 30m`, `@daily`, `@hourly`, `@weekly`, `@monthly`

#### CronDelivery Object

| Field | Type | Description |
|-------|------|-------------|
| `type` | enum | `CRON_DELIVERY_TYPE_LOG`, `CRON_DELIVERY_TYPE_WEBHOOK`, `CRON_DELIVERY_TYPE_CHANNEL`, `CRON_DELIVERY_TYPE_NOTIFY_GROUP` |
| `webhook_url` | string | URL for WEBHOOK type |
| `channel_name` | string | AgentChannel name for CHANNEL type |
| `chat_id` | string | Target chat ID for CHANNEL type |
| `notify_group_name` | string | NotifyGroup name for NOTIFY_GROUP type |

#### CronRetryPolicy Object

| Field | Type | Description |
|-------|------|-------------|
| `max_attempts` | int32 | Retries after the initial attempt; 0 disables retries |
| `backoff` | duration | Backoff between retry attempts; 0 retries immediately |

#### CronExecution Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique execution ID |
| `job_name` | string | Cron job name |
| `agent_name` | string | Agent that was executed |
| `status` | enum | `SUCCESS`, `ERROR`, `SKIPPED`, `CANCELLED`, `WAITING_INPUT` |
| `input` | string | Input message sent |
| `output` | string | Agent output preview or error message |
| `started_at` | timestamp | Execution start time |
| `finished_at` | timestamp | Execution end time; unset while status is `WAITING_INPUT` |
| `error` | string | Failure/cancellation error text |
| `duration_ms` | int64 | Wallclock duration |
| `attempt_count` | int32 | Attempts used for the logical execution |
| `trigger_type` | enum | `SCHEDULE` or `MANUAL` |
| `skipped_reason` | string | Reason populated when status is `SKIPPED` |
| `truncated` | bool | Whether output was truncated by `max_output_bytes` |
| `session_app_name` | string | Set when status is `WAITING_INPUT`: `app_name` for `SessionService.ReplySession` |
| `session_user_id` | string | Set when status is `WAITING_INPUT`: `user_id` for `SessionService.ReplySession` |
| `session_id` | string | Set when status is `WAITING_INPUT`: `session_id` for `SessionService.ReplySession` |
| `workspace_id` | string | Workspace that owns the parent cron job |

**Waiting executions (approval-style jobs):** a cron-run Workflow Agent that pauses on a Human Input node records its execution as `WAITING_INPUT` and delivers the node's question through the job's delivery target together with the session coordinates above. Answer by calling `SessionService.ReplySession` with those coordinates and the job's `agent_name`; the workflow resumes and the execution reaches a terminal state. To abandon instead, delete the session via `SessionService.DeleteSession` with the same coordinates: the execution transitions to `CANCELLED` with the reason recorded in `error`, and the cancellation is delivered per the job's `notify_on` policy (cancellations count as failures). Sessions removed outside that RPC (e.g. direct database cleanup) are not reconciled. Otherwise a paused execution waits indefinitely.

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
| `daemon_runtime_id` | string |  |
| `name` | string |  |
| `acp_runtimes` | string[] | Supported ACP runtimes |
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

Workspace-scoped daemon runtime configuration, token issuance, and control over
connected daemons. A daemon must have a stored `DaemonRuntime` in the active
workspace before a runtime token can be issued or a worker connection can be
accepted.

> `DaemonConnectorService` in `proto/agents/v1/daemon.proto` is the daemon worker's bidirectional streaming API (`Connect`) used for task dispatch and progress updates. It is mounted under the same `/api` ConnectRPC prefix as the dashboard APIs, and the local server exposes it on the dedicated `:8081` h2c listener. It authenticates daemon runtime tokens itself and is intended for `cmd/butter-daemon`, not dashboard / ops CRUD clients.

#### ListDaemonRuntimes

```
POST /api/agents.v1.DaemonService/ListDaemonRuntimes
```

**Response:** `{ "runtimes": DaemonRuntime[] }`

#### GetDaemonRuntime

```
POST /api/agents.v1.DaemonService/GetDaemonRuntime
```

**Request:** `{ "id": "<daemon-runtime-id>" }`

**Response:** `{ "runtime": DaemonRuntime }`

#### CreateDaemonRuntime / UpdateDaemonRuntime

```
POST /api/agents.v1.DaemonService/CreateDaemonRuntime
POST /api/agents.v1.DaemonService/UpdateDaemonRuntime
```

**Request:** `{ "runtime": DaemonRuntime }`

`id` and `name` are required. The server writes `workspace_id`, `created_at`,
and `created_by`.

#### DeleteDaemonRuntime

```
POST /api/agents.v1.DaemonService/DeleteDaemonRuntime
```

**Request:** `{ "id": "<daemon-runtime-id>" }`

#### CreateDaemonRuntimeToken

```
POST /api/agents.v1.DaemonService/CreateDaemonRuntimeToken
```

Issues a daemon runtime token for an existing runtime in the active workspace.
The returned secret should be passed to `cmd/butter-daemon --token`; the server
uses it to derive the authoritative workspace and daemon runtime id.

**Request:** `{ "daemon_runtime_id": "<id>", "name": "<label>"?, "ttl": "720h"? }`

**Response:** `{ "token": APIToken, "secret": "bt_<...>" }`

| DaemonRuntime | Type | Description |
|-------|------|-------------|
| `id` | string | Workspace-local runtime id |
| `name` | string | Display name |
| `description` | string |  |
| `labels` | map\<string,string\> | Operator metadata |
| `created_at` | timestamp |  |
| `created_by` | string | User id that created the config |
| `workspace_id` | string | Owning workspace |

#### ListDaemons

```
POST /api/agents.v1.DaemonService/ListDaemons
```

**Response:** `{ "daemons": DaemonStatus[] }`

#### GetDaemon

```
POST /api/agents.v1.DaemonService/GetDaemon
```

**Request:** `{ "daemon_runtime_id": "<id>" }`

**Response:** `{ "daemon": DaemonStatus }`

| DaemonStatus | Type | Description |
|-------|------|-------------|
| `daemon_runtime_id` | string |  |
| `name` | string |  |
| `acp_runtimes` | string[] | Supported ACP runtimes |
| `labels` | map\<string,string\> |  |
| `state` | enum | `STATE_ONLINE`, `STATE_IDLE`, `STATE_OFFLINE` |
| `connected_at` | timestamp |  |
| `uptime` | duration |  |
| `active_tasks` | int32 |  |
| `version` | string | Reported at handshake |
| `os` | string | e.g. `linux-amd64` |
| `executors` | string[] | Informational executor names |
| `remote_addr` | string | Peer address captured by the server |
| `workspace_id` | string | Owning workspace |

#### CancelDaemonTask

```
POST /api/agents.v1.DaemonService/CancelDaemonTask
```

Cancels a task on any connected daemon (or the supplied `daemon_runtime_id`). Returns `not_found` if no online daemon is tracking the task.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `task_id` | string | Required |
| `daemon_runtime_id` | string | Optional hint |

**Response:** `{ "daemon_runtime_id": "<id>" }` (the runtime that received the cancel)

#### ListDaemonTasks

```
POST /api/agents.v1.DaemonService/ListDaemonTasks
```

Tasks currently in flight across connected daemons.

**Request:** `{ "daemon_runtime_id": "<id>"? }`

**Response:** `{ "tasks": DaemonTaskInFlight[] }`

| DaemonTaskInFlight | Type | Description |
|-------|------|-------------|
| `task_id` | string |  |
| `daemon_runtime_id` | string |  |
| `daemon_name` | string |  |
| `acp_runtime` | string | ACP runtime used to dispatch the task |
| `started_at` | timestamp | When the task was dispatched |
| `elapsed` | duration | Wallclock since dispatch |
| `current_step` | string | Latest progress label reported by the daemon |
| `progress` | int32 | 0–100; 0 when unknown |
| `agent_name` | string | Agent that triggered the dispatch |
| `workspace_id` | string | Owning workspace |

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

### DaemonConnectorService

Daemon client connection protocol. This is a bidirectional ConnectRPC streaming
service under `/api`; local cleartext workers should use
`http://localhost:8081/api`, while production workers can use an HTTPS/HTTP2
base URL such as `https://butter.example.com/api`.

Reverse proxies in front of this path must allow long-lived streaming requests:
disable request/response buffering for `/api/agents.v1.DaemonConnectorService/`
and set read/send timeouts high enough for an idle daemon connection.

#### Connect

```
POST /api/agents.v1.DaemonConnectorService/Connect
```

Establishes a long-lived stream. The daemon sends a `register` message first, then sends `task_update` messages. The server sends task assignments and cancellation requests on the response stream.

The client must include `authorization: Bearer <daemon-runtime-token>`. The server validates the hashed token before accepting the registration: the token must be `API_TOKEN_KIND_DAEMON`, carry `daemon:connect`, belong to a workspace, and reference an existing `DaemonRuntime`. The token's workspace and `daemon_runtime_id` are authoritative; if the daemon self-reports a different `workspace_id` or runtime id, the server overwrites/rejects accordingly. If the daemon does not report `acp_runtimes`, the server defaults it to `opencode` and `codex`.

**Client stream (`ConnectRequest`):**

| Field | Type | Description |
|-------|------|-------------|
| `register` | DaemonInfo | First message only; registers daemon runtime id, name, acp runtimes, labels, version, OS, executor names, and optional workspace id |
| `task_update` | DaemonTaskUpdate | Subsequent task lifecycle, output, error, current step, and progress updates |

**Server stream (`ConnectResponse`):**

| Field | Type | Description |
|-------|------|-------------|
| `task` | DaemonTask | Task assignment, including agent name, input, session/user ids, metadata, daemon runtime id, acp runtime, and work dir |
| `cancel` | CancelTask | Cancellation request by `task_id` |

---

### WorkspaceService

Manages workspaces and their memberships. No `X-Workspace-ID` header required — the service is the workspace selector itself.

#### ListWorkspaces

```
POST /api/agents.v1.WorkspaceService/ListWorkspaces
```

Returns the workspaces the caller can access. Global `admin` users get every workspace; other users only see their memberships.

**Request:** `{}`

**Response:** `{ "workspaces": Workspace[] }`

#### GetWorkspace

```
POST /api/agents.v1.WorkspaceService/GetWorkspace
```

**Request:** `{ "id": "<workspace-id>" }`

**Response:** `{ "workspace": Workspace }`

#### CreateWorkspace

```
POST /api/agents.v1.WorkspaceService/CreateWorkspace
```

Creates a new workspace; the caller is added as the initial `owner`.

**Request:**

| Field | Type | Description |
|-------|------|-------------|
| `workspace.name` | string | Required display name |
| `workspace.slug` | string | Required URL-safe slug (must be unique) |
| `workspace.description` | string | Optional |

**Response:** `{ "workspace": Workspace }` — `id`, `created_at`, `updated_at` filled by server.

#### UpdateWorkspace

```
POST /api/agents.v1.WorkspaceService/UpdateWorkspace
```

Updates workspace metadata. The workspace is matched by `workspace.id`.

**Request:** `{ "workspace": Workspace }`

**Response:** `{ "workspace": Workspace }`

#### DeleteWorkspace

```
POST /api/agents.v1.WorkspaceService/DeleteWorkspace
```

Permanently deletes a workspace and removes every `workspace_members` row for that workspace.

**Request:** `{ "id": "<workspace-id>" }`

**Response:** `{}`

#### ListWorkspaceMembers

```
POST /api/agents.v1.WorkspaceService/ListWorkspaceMembers
```

**Request:** `{ "workspace_id": "<workspace-id>" }`

**Response:** `{ "members": WorkspaceMember[] }`

#### AddWorkspaceMember / UpdateWorkspaceMember / RemoveWorkspaceMember

```
POST /api/agents.v1.WorkspaceService/AddWorkspaceMember
POST /api/agents.v1.WorkspaceService/UpdateWorkspaceMember
POST /api/agents.v1.WorkspaceService/RemoveWorkspaceMember
```

**Add / Update request:**

| Field | Type | Description |
|-------|------|-------------|
| `workspace_id` | string | Required |
| `user_id` | string | Required |
| `role` | string | `owner` / `admin` / `member`; Add defaults to `member`, Update requires non-empty |

**Remove request:** `{ "workspace_id": "...", "user_id": "..." }`

**Response:** `{ "member": WorkspaceMember }` for Add/Update; `{}` for Remove.

#### Workspace Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Server-assigned UUID |
| `name` | string | Display name |
| `slug` | string | URL-safe unique identifier |
| `description` | string | Free-form |
| `created_at` | timestamp |  |
| `updated_at` | timestamp |  |

#### WorkspaceMember Object

| Field | Type | Description |
|-------|------|-------------|
| `workspace_id` | string |  |
| `user_id` | string |  |
| `role` | string | `owner` / `admin` / `member` |
| `created_at` | timestamp |  |

---

### ForumService

Manages forum threads and posts within a workspace. Requires `X-Workspace-ID`.

| RPC | Path | Notes |
|-----|------|-------|
| `ListThreads` | `POST /api/agents.v1.ForumService/ListThreads` | Paginated; filter by label |
| `ListThreadLabels` | `POST /api/agents.v1.ForumService/ListThreadLabels` | Labels in use across threads |
| `GetThread` | `POST /api/agents.v1.ForumService/GetThread` | Includes posts |
| `CreateThread` | `POST /api/agents.v1.ForumService/CreateThread` | Creates thread with optional initial post |
| `UpdateThread` | `POST /api/agents.v1.ForumService/UpdateThread` | Updates title/labels |
| `DeleteThread` | `POST /api/agents.v1.ForumService/DeleteThread` | Deletes thread and all posts |
| `CreatePost` | `POST /api/agents.v1.ForumService/CreatePost` | Appends a post to a thread |
| `DeletePost` | `POST /api/agents.v1.ForumService/DeletePost` | Deletes a single post |
| `InvokeAgentInThread` | `POST /api/agents.v1.ForumService/InvokeAgentInThread` | Invokes an agent in thread context |

---

### APITokenService

Manages API bearer tokens. The plaintext secret is only returned at create time; subsequent reads expose the prefix only. Each token is scoped to one workspace (taken from `X-Workspace-ID` at create time); authentication automatically scopes the request to that workspace.

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
| `workspace_id` | string | Workspace the token authenticates into |
| `kind` | enum | `API_TOKEN_KIND_USER` or `API_TOKEN_KIND_DAEMON` |
| `scopes` | string[] | `api:*` for user API tokens, `daemon:connect` for daemon credentials |
| `expires_at` | timestamp | Optional expiry; empty means no expiry |
| `daemon_runtime_id` | string | Set for daemon runtime tokens |
| `created_at` | timestamp |  |
| `last_used_at` | timestamp | Updated async after successful auth |
| `revoked` | bool | Revoked tokens cannot authenticate |

---

## Error Handling

RPC endpoints return errors in the standard Connect envelope:

```json
{
  "code": "not_found",
  "message": "agent \"foo\" not found"
}
```

The HTTP status code mirrors the gRPC mapping (Connect / gRPC-Web also
encode the code via response headers / trailers). Common codes:

- `invalid_argument` (HTTP 400) — request validation failed
- `unauthenticated` (HTTP 401) — missing or invalid bearer token
- `permission_denied` (HTTP 403) — insufficient role
- `not_found` (HTTP 404) — resource missing
- `already_exists` (HTTP 409) — name/id collision
- `failed_precondition` (HTTP 400) — service not yet initialized or
  X-Workspace-ID header missing on a scoped RPC
- `internal` (HTTP 500) — unexpected server error

REST endpoints (`/ping`, `/status`, `/a2a/*`, `/api/uploads/*`)
return errors as `{"error": "..."}` with a conventional HTTP status code.

TypeScript client handling:

```ts
import { Code, ConnectError } from "@connectrpc/connect";

try {
  await agentClient.listAgents({});
} catch (err) {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.Unauthenticated:
        redirectToLogin();
        break;
      case Code.PermissionDenied:
        showToast("You do not have access to this action.");
        break;
      case Code.FailedPrecondition:
        showToast(err.message); // often missing X-Workspace-ID
        break;
      default:
        showToast(err.message);
    }
  }
}
```

For app developers, the most common integration mistakes are:

- Missing `Authorization` after login.
- Missing `X-Workspace-ID` on workspace-scoped calls.
- Sending JSON field names from generated TypeScript types (`agentName`) while
  expecting snake_case JSON responses (`agent_name`).
- Trying to upload files through ConnectRPC instead of the REST multipart
  endpoints under `/api/uploads/*`.
