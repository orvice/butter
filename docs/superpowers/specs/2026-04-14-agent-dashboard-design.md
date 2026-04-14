# Agent Management Dashboard — Design Spec

## Overview

A React-based management dashboard for Butter's agent platform. Provides CRUD management for all 5 API services (Agents, MCP Servers, Remote Agents, Sessions, Cron Jobs) plus a Cron execution monitoring dashboard with charts and timelines.

**Target audience:** External users. Polished UI with clear onboarding.

## Tech Stack

- **Build:** Vite
- **UI:** React 18 + TypeScript
- **Routing:** React Router 6
- **Components:** shadcn/ui + Tailwind CSS
- **Data fetching:** TanStack Query 5
- **Charts:** Recharts
- **JSON editor:** @monaco-editor/react
- **Theme:** next-themes (dark + light toggle, system preference detection)
- **Toast:** Sonner
- **Icons:** lucide-react
- **Forms:** react-hook-form + zod validation
- **Language:** English first, i18n-ready (no i18n library initially, but string literals centralized in constants)

## Authentication

Token-based authentication matching the current API mechanism (single Bearer token configured in `config.yaml`).

### Flow

1. User visits any page → check `localStorage` for token
2. No token → redirect to `/login`
3. Login page: single input field + "Connect" button
4. On submit: store token, call `POST /api/agents.v1.AgentService/ListAgents` (authenticated endpoint) to validate
5. Success response → token is valid, redirect to `/`
6. 401 or error → clear token, show error message
6. Any API returning 401 → clear token, redirect to `/login`

Future extensibility: auth logic is isolated in `hooks/use-auth.ts` and `api/client.ts`, making it straightforward to swap in a user system later.

## Layout

Fixed left sidebar navigation + right content area. Login page has its own standalone layout.

### Sidebar

| Icon (lucide) | Label | Route |
|---|---|---|
| LayoutDashboard | Dashboard | `/` |
| Bot | Agents | `/agents` |
| Server | MCP Servers | `/mcp-servers` |
| Globe | Remote Agents | `/remote-agents` |
| MessageSquare | Sessions | `/sessions` |
| Clock | Cron Jobs | `/cron` |

Sidebar bottom: theme toggle button + logout (clear token).

### Theme

Dark and light themes with system preference detection. Toggle in sidebar footer. Implemented via `next-themes` (works with Vite, no Next.js dependency). shadcn/ui provides built-in dark mode CSS variables.

## Routes

```
/login                    → Login page (no sidebar)
/                         → Dashboard (Cron execution monitor)
/agents                   → Agent list
/agents/create            → Create Agent
/agents/:name/edit        → Edit Agent
/mcp-servers              → MCP Server list
/mcp-servers/create       → Create MCP Server
/mcp-servers/:id/edit     → Edit MCP Server
/remote-agents            → Remote Agent list
/remote-agents/create     → Create Remote Agent
/remote-agents/:id/edit   → Edit Remote Agent
/sessions                 → Session list (filterable by app_name, user_id)
/sessions/detail?app=X&user=Y&session=Z → Session detail (event list)
/cron                     → Cron Job list
/cron/create              → Create Cron Job
/cron/:name/edit          → Edit Cron Job
/cron/:name/executions    → Execution records for a single job
```

## Project Structure

```
front/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.ts
├── components.json              # shadcn/ui config
├── public/
├── src/
│   ├── main.tsx                 # Entry point
│   ├── App.tsx                  # Router + ThemeProvider + QueryClientProvider
│   ├── api/
│   │   ├── client.ts            # Twirp HTTP client wrapper
│   │   ├── agents.ts            # AgentService hooks
│   │   ├── mcp-servers.ts       # MCPServerService hooks
│   │   ├── remote-agents.ts     # RemoteAgentService hooks
│   │   ├── sessions.ts          # SessionService hooks
│   │   └── cron.ts              # CronJobService hooks
│   ├── components/
│   │   └── ui/                  # shadcn/ui generated components
│   ├── layouts/
│   │   └── dashboard-layout.tsx # Sidebar + content shell
│   ├── pages/
│   │   ├── login.tsx
│   │   ├── dashboard.tsx
│   │   ├── agents/
│   │   │   ├── list.tsx
│   │   │   ├── create.tsx
│   │   │   └── edit.tsx
│   │   ├── mcp-servers/
│   │   │   ├── list.tsx
│   │   │   ├── create.tsx
│   │   │   └── edit.tsx
│   │   ├── remote-agents/
│   │   │   ├── list.tsx
│   │   │   ├── create.tsx
│   │   │   └── edit.tsx
│   │   ├── sessions/
│   │   │   ├── list.tsx          # filterable by app_name, user_id
│   │   │   └── detail.tsx        # uses query params: app, user, session
│   │   └── cron/
│   │       ├── list.tsx
│   │       ├── create.tsx
│   │       ├── edit.tsx
│   │       └── executions.tsx
│   ├── hooks/
│   │   └── use-auth.ts
│   ├── lib/
│   │   ├── utils.ts             # shadcn cn() helper
│   │   └── constants.ts
│   └── types/
│       └── api.ts               # TypeScript types from proto definitions
```

## API Layer

### Twirp Client (`api/client.ts`)

A thin wrapper around `fetch` for Twirp RPC calls:

- Base URL from `VITE_API_BASE_URL` environment variable
- All requests: `POST` with `Content-Type: application/json` and `Authorization: Bearer <token>`
- Response parsing: JSON body, Twirp error format `{ code, msg }` thrown as typed errors
- 401 responses: clear token from `localStorage`, redirect to `/login`

### TanStack Query Hooks

Each service file exports query and mutation hooks. Example for `agents.ts`:

- `useAgents()` → `useQuery(["agents"], listAgents)`
- `useAgent(name)` → `useQuery(["agents", name], () => getAgent(name))`
- `useCreateAgent()` → `useMutation(createAgent, { onSuccess: invalidate ["agents"] })`
- `useUpdateAgent()` → `useMutation(updateAgent, { onSuccess: invalidate ["agents"] + ["agents", name] })`
- `useDeleteAgent()` → `useMutation(deleteAgent, { onSuccess: invalidate ["agents"] })`

Same pattern for all 5 services.

### Query Key Design

```
["agents"]                          → ListAgents
["agents", name]                    → GetAgent
["mcp-servers"]                     → ListMCPServers
["mcp-servers", id]                 → GetMCPServer
["remote-agents"]                   → ListRemoteAgents
["remote-agents", id]               → GetRemoteAgent
["sessions", { appName, userId }]   → ListSessions
["sessions", { appName, userId, sessionId }] → GetSession
["cron-jobs"]                       → ListCronJobs
["cron-jobs", name]                 → GetCronJob
["cron-executions", { jobName, page }] → ListCronExecutions
```

### Query Configuration

- List data: `staleTime: 30_000` (30s)
- Dashboard data: `refetchInterval: 60_000` (auto-refresh every 60s)
- Mutations: invalidate relevant query keys on success, toast on error

### Error Handling

- Global `QueryClient.onError`: display error `msg` via sonner toast
- 401 in fetch wrapper: clear token, redirect `/login`
- Network errors: "Connection failed" toast

## CRUD Pages — Shared Patterns

All 5 resources follow the same UI patterns for consistency.

### List Page

- Header: page title + "Create" button (right-aligned)
- shadcn `DataTable` with resource-specific columns
- Each row: actions dropdown (Edit, Delete)
- Delete: `AlertDialog` confirmation
- Empty state: illustration + "Create your first X" CTA
- Loading: skeleton placeholders

### Column Definitions

| Resource | Columns |
|---|---|
| Agents | Name, Type (badge), Description, A2A (badge), Actions |
| MCP Servers | ID, Name, Transport (badge), URL/Command, Actions |
| Remote Agents | ID, Name, URL, Protocol (badge), Actions |
| Sessions | Session ID, App Name, User ID, Last Update, Actions |
| Cron Jobs | Name, Schedule, Agent, Timezone, Enabled (switch), Actions |

### Create/Edit Page

- Breadcrumb navigation + page title
- shadcn `Form` + react-hook-form + zod validation
- Grouped fields in `Card` components: Basic Info, Configuration, Advanced
- Fixed bottom action bar: Cancel + Save
- Success: toast + redirect to list page

### Agent Edit — Dual Mode (Form + JSON)

Top tab bar toggles between "Form" and "JSON" views. Switching syncs data bidirectionally.

**Form mode (default):**
- Basic Info: name, description, type (select), enable_a2a (switch)
- Model Config: model (input), instruction (textarea)
- MCP Config: mcp_server_ids (multi-select from ListMCPServers)
- Remote Agents: remote_agent_ids (multi-select from ListRemoteAgents)
- Advanced: context_guard settings, include_contents, output_key, schemas
- Sub-agents: read-only list (name + type badges). Full editing requires JSON mode.

**JSON mode:**
- Monaco editor with full Agent JSON
- Syntax highlighting, validation, auto-format

## Dashboard Page (`/`)

Cron execution monitoring dashboard. Four components:

### 1. Stats Cards (top row, 4-column grid)

| Card | Value | Source |
|---|---|---|
| Total Executions (24h) | Count | Client-side count from executions list |
| Success Rate | Percentage | success / total from executions |
| Active Cron Jobs | Count | ListCronJobs filtered by `enabled: true` |
| Avg Duration | Seconds | Mean of `finished_at - started_at` |

Comparison hints (e.g., "+12% from yesterday") are omitted in v1 since the API lacks time-range filtering. Can be added when the backend supports it.

### 2. Execution Timeline (stacked bar chart)

- Recharts `BarChart` with hourly buckets over last 24 hours
- Stacked bars: green (success) + red (error)
- X-axis: time labels (00:00, 06:00, 12:00, 18:00, Now)
- Legend: Success / Error

### 3. Status Breakdown (donut chart)

- Recharts `PieChart` (donut)
- Two segments: success (green), error (red)
- Center label: overall success rate percentage

### 4. Recent Executions (table)

- Last 10 execution records
- Columns: Job Name, Agent, Status (badge), Duration, Started At, Output (truncated)
- "View all →" link navigates to `/cron` executions view

### Data Source

All dashboard data comes from two API calls:
- `ListCronJobs` (for active job count)
- `ListCronExecutions` (no job_name filter, fetch recent pages)

Stats and charts are computed client-side by aggregating the execution records. Auto-refresh every 60 seconds.

## TypeScript Types (`types/api.ts`)

Hand-written types mirroring proto message definitions. Key types:

```typescript
// Enums
type AgentType = "AGENT_TYPE_UNSPECIFIED" | "AGENT_TYPE_LLM" | "AGENT_TYPE_LOOP" | "AGENT_TYPE_SEQUENTIAL" | "AGENT_TYPE_PARALLEL"
type MCPServerTransport = "MCP_SERVER_TRANSPORT_UNSPECIFIED" | "MCP_SERVER_TRANSPORT_STDIO" | "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP" | "MCP_SERVER_TRANSPORT_SSE"
type RemoteAgentProtocol = "REMOTE_AGENT_PROTOCOL_UNSPECIFIED" | "REMOTE_AGENT_PROTOCOL_A2A"
type CronDeliveryType = "CRON_DELIVERY_TYPE_UNSPECIFIED" | "CRON_DELIVERY_TYPE_LOG" | "CRON_DELIVERY_TYPE_WEBHOOK" | "CRON_DELIVERY_TYPE_CHANNEL"
type CronExecutionStatus = "CRON_EXECUTION_STATUS_UNSPECIFIED" | "CRON_EXECUTION_STATUS_SUCCESS" | "CRON_EXECUTION_STATUS_ERROR"

// Core models
interface Agent { name, description, sub_agents, labels, metadata, config, type, enable_a2a }
interface AgentConfig { model, instruction, global_instruction, mcp_servers, mcp_server_ids, remote_agent_ids, context_guard, include_contents, output_key, input_schema_json, output_schema_json, max_iterations }
interface MCPServer { id, name, transport, command, args, env, url, headers, tool_filter, metadata }
interface RemoteAgent { id, name, url, protocol }
interface SessionInfo { session_id, app_name, user_id, state, last_update_time }
interface SessionDetail { session, events }
interface SessionEvent { event_id, invocation_id, author, branch, content_json, timestamp }
interface CronJob { name, schedule, agent_name, input, timezone, enabled, delivery, metadata }
interface CronDelivery { type, webhook_url, channel_name, chat_id }
interface CronExecution { id, job_name, agent_name, status, input, output, started_at, finished_at }
```

## State Management

**No global state library.** Two categories:

- **Server state:** TanStack Query (all API data, caching, refetching)
- **Client state:** React Context for auth (token) and theme; component-local state for form inputs, UI toggles

## Non-Goals (Out of Scope)

- Interactive session chat (ReplySession API)
- Real-time WebSocket/SSE streaming
- User management / RBAC
- Agent channel (Telegram/Discord) configuration UI
- i18n library integration (strings are centralized but no runtime translation)
- Mobile-responsive layout (desktop-first)
