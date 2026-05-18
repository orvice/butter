# PostgreSQL Migration Analysis

This note summarizes Butter's current database usage and which parts are good
candidates for a PostgreSQL backend.

## Current Storage Shape

Butter currently uses MongoDB as the default persistent store and Redis for
short-lived channel/auth state. `storage_backend` accepts `mongo` or `memory`,
but the runtime always connects MongoDB for ADK sessions, ADK memory, cron, auth
users, and dashboard health checks.

MongoDB collections in active use:

| Area | Collections | Usage pattern |
| --- | --- | --- |
| Workspace metadata | `workspaces`, `workspace_members` | CRUD, unique slug, membership lookup by `(workspace_id, user_id)` and by `user_id`. |
| Dashboard users | `users` | CRUD-like user management, unique username, bcrypt password hash. |
| Dashboard auth sessions | `auth_sessions` in Mongo implementation, Redis wrapper in production path | Token hash lookup, TTL expiry, revoke/touch. Production boot wraps Mongo users with Redis session storage. |
| API tokens | `api_tokens` | Workspace-scoped list, lookup by `secret_hash`, revoke, `last_used_at` touch. |
| Runtime config | `config_agents`, `config_mcpservers`, `config_remoteagents`, `config_channels`, `config_modelproviders` | Workspace-scoped CRUD. Current Mongo doc stores `workspace_id`, `name`, and full protojson `spec`. |
| Cron jobs | `cron_jobs` | Workspace-scoped CRUD, composite id `workspace_id:name`, mostly scalar fields plus metadata. |
| Cron executions | `cron_executions` | Append execution rows, list by workspace/job, list by time range for dashboard aggregation. |
| Invocations | `invocations` | Upsert by invocation id, list/filter by workspace/agent/session, recent activity feed. Current Mongo doc stores full protojson `spec` and filters with regex over JSON. |
| ADK sessions | `adk_sessions`, `adk_events` | Session row plus append-only ordered events; recent-event queries; JSON content and mutable session state. |
| ADK memory | `adk_memories` | Re-ingest session memories, text search by app/user/query, JSON content and metadata. |

Redis is still used for:

- Dashboard auth sessions in the production `authredis.Store` wrapper.
- Telegram/Discord active agent, model selection, and debug toggles.

## Best PostgreSQL Candidates

### High Fit

These are relational, scalar, and already accessed through repository
interfaces. They should be migrated first.

| Area | Why it fits PostgreSQL | Suggested shape |
| --- | --- | --- |
| `workspaces` / `workspace_members` | Natural relational model with uniqueness and join-like membership checks. | `workspaces(id primary key, slug unique, ...)`; `workspace_members(workspace_id references workspaces, user_id, role, primary key(workspace_id, user_id))`; index `(user_id)`. |
| `users` | Scalar fields, unique username, simple update/read paths. | `users(id primary key, username unique, password_hash, role, disabled, created_at, updated_at)`. |
| `api_tokens` | Scalar lookup/update workload. | `api_tokens(id primary key, workspace_id, name, prefix, secret_hash unique, revoked, created_at, last_used_at)` plus indexes on `(workspace_id)` and `(secret_hash) where revoked = false`. |
| `cron_jobs` | Structured config with stable scalar fields. | `cron_jobs(workspace_id, name, schedule, agent_name, input, timezone, enabled, delivery fields, metadata jsonb, primary key(workspace_id, name))`. |
| `cron_executions` | Append/list/time-range analytics are stronger in PostgreSQL. | `cron_executions(id primary key, workspace_id, job_name, agent_name, status, input, output, started_at timestamptz, finished_at timestamptz)`; indexes `(workspace_id, job_name, started_at desc)` and `(started_at)`. |

### Medium Fit

These can move to PostgreSQL with `jsonb`, but the schema decision matters.

| Area | Migration approach | Notes |
| --- | --- | --- |
| Runtime config collections | Store canonical protojson in `jsonb` plus generated/query columns for `workspace_id`, `name`, and entity-specific id. | Current CRUD is simple and workspace-scoped, so moving is easy. Avoid opaque `text` JSON; `jsonb` enables validation, indexing, and future queries. |
| `invocations` | Split frequently filtered fields into columns, keep full invocation as `jsonb`. | The current Mongo implementation filters by regex over `spec`, which is a strong signal this should become columns: `workspace_id`, `agent_name`, `session_id`, `status`, `started_at`, `finished_at`, plus `spec jsonb`. This is a good PostgreSQL target once the query contract is made explicit. |
| ADK sessions/events | Use relational session/event tables with `jsonb` for state/content. | Fits PostgreSQL well if event ordering and state-delta updates are implemented carefully. Use `(app_name, user_id, session_id)` unique key and `(app_name, session_id, timestamp)` event index. |

### Lower Fit / Defer

| Area | Reason to defer | PostgreSQL option |
| --- | --- | --- |
| ADK memory search | Current behavior depends on Mongo text search over `content_text`; quality and ranking semantics will change. | Use PostgreSQL full-text search (`tsvector`, GIN index) for simple keyword search, or introduce a vector/embedding store if semantic memory search is desired. Migrate after deciding search requirements. |
| Redis-backed channel selections/debug toggles | These are ephemeral, small, and latency-sensitive. | Keep in Redis unless the product needs durable audit/history. PostgreSQL is not a better default for this state. |
| Redis auth session cache | TTL and touch-heavy token lookup are a good Redis workload. | Keep Redis as the hot session store. If DB-backed sessions are required, PostgreSQL can mirror `auth_sessions` with `expires_at` index and periodic cleanup, but that changes operational behavior. |

## Recommended Migration Order

1. Add a PostgreSQL connection/config path and a backend selector that is not named `storage_backend=mongo|memory` only. Keep repository interfaces stable.
2. Implement PostgreSQL repositories for `workspace`, `auth` users, `apitoken`, `cron` jobs/executions.
3. Move runtime config repositories to PostgreSQL using `jsonb` plus scalar identity columns.
4. Redesign `invocations` storage around indexed scalar fields instead of JSON regex filtering, then migrate it.
5. Move ADK sessions/events if operationally needed; validate event ordering, recent-event pagination, and state-delta updates against ADK behavior.
6. Decide memory search direction before moving `adk_memories`: PostgreSQL full-text for keyword search, or a vector-capable store for semantic search.

## Implementation Notes

- Prefer adding `internal/repo/<area>/postgres` packages next to existing
  `mongo` and `memory` implementations.
- Use `pgx` or `database/sql` consistently across all new repositories.
- Add migrations rather than relying on runtime index creation. The current
  Mongo code creates indexes in repository constructors/bootstrap; PostgreSQL
  should have explicit schema migrations.
- Preserve workspace scoping in every migrated table. Most runtime/dashboard
  APIs assume `workspace_id` is available for filtering and authorization.
- Convert timestamp fields to `timestamptz`; avoid storing Unix seconds where
  dashboard aggregation or ordering is required.
- Treat protojson payloads as compatibility envelopes, not as the primary query
  surface. Any field used in a filter, sort, or uniqueness rule should be a
  first-class column.
