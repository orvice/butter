# MongoDB → PostgreSQL Migration Plan

This document records the analysis and recommended plan for migrating Butter's
database layer from MongoDB to PostgreSQL, based on actual usage patterns.

## Decision Framework

| Dimension | Favor PostgreSQL | Favor MongoDB |
|---|---|---|
| Schema stability | Fixed structure, explicit fields | Dynamic schema, arbitrary JSON |
| Write pattern | Low-medium CRUD, needs transactions | High-frequency append-only, event streams |
| Query pattern | Exact lookup, relational joins, pagination | Full-text search, time-range scans |
| Data growth | Bounded (config/users/relations) | Unbounded (logs/events/conversations) |
| Consistency needs | Strong consistency, FK constraints | Eventual consistency acceptable |

## Collections to Migrate to PostgreSQL (20 collections)

### Users & Auth

| Collection | Access Pattern | Why PG |
|---|---|---|
| `users` | CRUD, lookup by username/external_id, count | Core relational entity referenced everywhere |
| `auth_sessions` | Lookup by token_hash + expiry check, low write | Pairs with users table; TTL via `WHERE expires_at > now()` |
| `oauth_states` | Create → atomic consume (FindOneAndDelete) | Extremely low volume, ephemeral |

### Workspaces

| Collection | Access Pattern | Why PG |
|---|---|---|
| `workspaces` | CRUD, unique slug constraint | Classic relational entity |
| `workspace_members` | Many-to-many (workspace ↔ user), compound unique | Relationship table; benefits from FK constraints |

### API Tokens

| Collection | Access Pattern | Why PG |
|---|---|---|
| `api_tokens` | Create, lookup by secret_hash (hot path), revoke, touch | Fixed schema, indexed lookup, low write volume |

### Configuration (7 collections)

| Collection | Access Pattern | Why PG |
|---|---|---|
| `config_agents` | List by workspace, Get by (workspace, name), CRUD | Config data; JSONB `spec` column + composite unique |
| `config_mcpservers` | Same pattern | Same rationale |
| `config_global_mcpservers` | List all, Get by id, CRUD | Same rationale |
| `config_remoteagents` | Same pattern | Same rationale |
| `config_channels` | Same pattern | Same rationale |
| `config_modelproviders` | Same pattern | Same rationale |
| `config_notifygroups` | Same pattern | Same rationale |

All config collections share the same `configDoc` structure: `{_id, workspace_id, name, spec}`.
They have extremely low write frequency and small data volume (< 50 rows per workspace).

### MCP OAuth

| Collection | Access Pattern | Why PG |
|---|---|---|
| `mcp_oauth_connections` | Upsert, Get, MarkState (atomic update), Delete | Fixed schema, low frequency; `INSERT ON CONFLICT` replaces upsert |

### Forum

| Collection | Access Pattern | Why PG |
|---|---|---|
| `forum_threads` | CRUD, paginated list with sort/filter, transactions, Distinct labels | Relational parent entity; PG transactions are simpler |
| `forum_posts` | CRUD, paginated list by thread, parent-child with threads | Relational child entity; benefits from FK to threads |

The forum uses MongoDB multi-document transactions (`CreatePostAndMarkThreadProcessing`).
PG native `BEGIN/COMMIT` with `UPDATE ... RETURNING` is more reliable.

### Cron Jobs

| Collection | Access Pattern | Why PG |
|---|---|---|
| `cron_jobs` | List by workspace, List all, CRUD by (workspace, name) | Configuration data, not event data |

### Agent File Metadata

| Collection | Access Pattern | Why PG |
|---|---|---|
| `agent_file_spaces` | List by workspace, Get, CRUD | Metadata only (content in S3); relational hierarchy |
| `agent_files` | List by (workspace, space), Get, Write (upsert-like) | Parent-child with spaces |
| `agent_file_versions` | Insert on write, query by (workspace, space, path, version) | 1:N with files; monotonic version numbers |

## Collections to Keep on MongoDB (5 collections)

### ADK Sessions & Events

| Collection | Access Pattern | Why Mongo |
|---|---|---|
| `adk_sessions` | Create, Get, List, delta-update state (`$set: state.key`) | Dynamic `map[string]any` state with partial updates |
| `adk_events` | High-frequency AppendEvent per conversation turn | Append-only, arbitrary JSON content (genai.Content) |

Key factors:
- Every agent conversation turn appends an event (high write frequency)
- Session state uses `$set: {"state." + key: value}` for atomic partial updates
- Event content schema is dictated by the LLM SDK and evolves independently
- Data volume grows proportionally to total agent interactions (unbounded)

### ADK Memory

| Collection | Access Pattern | Why Mongo |
|---|---|---|
| `adk_memories` | Batch delete+insert (re-ingestion), `$text` search with score | Full-text search with relevance ranking on semi-structured content |

Key factors:
- MongoDB `$text` index with `$meta: textScore` is already working
- Content is free-form text extracted from `genai.Content`
- Re-ingestion pattern (delete all for session → insert many) suits Mongo
- Future path may involve vector search (MongoDB Atlas Search)

### Invocations

| Collection | Access Pattern | Why Mongo |
|---|---|---|
| `invocations` | Upsert twice per agent run (start + end), paginated list by time | High-frequency operational telemetry with spec blob |

Key factors:
- Highest write frequency in the system (2× per agent run)
- Operational telemetry, not core business data (low consistency needs)
- Unbounded growth; may need TTL or sharding in future
- protojson spec blob has zero overhead in Mongo vs JSONB parse cost in PG

### Cron Executions

| Collection | Access Pattern | Why Mongo |
|---|---|---|
| `cron_executions` | Append-only per cron tick, time-range queries | Operational log data, same nature as invocations |

## Architecture: Hybrid PG + Mongo

```
┌─────────────────────────────────────────────────────┐
│                   Butter Service                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌─────────────────────┐  ┌──────────────────────┐  │
│  │   PostgreSQL (PG)   │  │   MongoDB (Mongo)    │  │
│  │                     │  │                      │  │
│  │  • workspaces       │  │  • adk_sessions      │  │
│  │  • workspace_members│  │  • adk_events        │  │
│  │  • users            │  │  • adk_memories      │  │
│  │  • auth_sessions    │  │  • invocations       │  │
│  │  • oauth_states     │  │  • cron_executions   │  │
│  │  • api_tokens       │  │                      │  │
│  │  • config_* (×7)    │  └──────────────────────┘  │
│  │  • mcp_oauth_conns  │                            │
│  │  • forum_threads    │  ┌──────────────────────┐  │
│  │  • forum_posts      │  │      Redis           │  │
│  │  • cron_jobs        │  │  • session cache     │  │
│  │  • agent_file_*     │  │  • channel state     │  │
│  └─────────────────────┘  └──────────────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Configuration Changes

```yaml
# config.yaml additions
postgres_uri: "postgres://user:pass@localhost:5432/butter?sslmode=disable"
storage_backend: "postgres"  # new option; mongo collections above still use mongo_uri
```

The `storage_backend` field gains a new value `"postgres"` (or `"postgres+mongo"` to be
explicit). When set to `"postgres"`:
- Core business repos (workspace, auth, apitoken, config, forum, etc.) use PG
- Runtime repos (adk_sessions, adk_events, adk_memories, invocations, cron_executions) still use MongoDB via `mongo_uri`

## Implementation Plan

### Phase 1: Foundation

1. Add `postgres_uri` to `AppConfig`
2. Add `connectPostgres()` to `internal/app/runtime.go`
3. Introduce migration framework (recommend `pressly/goose` or `golang-migrate/migrate`)
4. Write SQL migrations for all 20 PG tables

### Phase 2: Simple Repos (field-level mapping)

These repos have explicit `xxxDoc` structs with `bson` tags → map directly to PG columns.

1. `internal/repo/workspace/pg/pg.go`
2. `internal/repo/auth/pg/pg.go`
3. `internal/repo/apitoken/pg/pg.go`
4. `internal/repo/oauthstate/pg/pg.go`
5. `internal/repo/mcpoauth/pg/pg.go`

### Phase 3: Spec-Blob Repos (JSONB + index columns)

These repos store protojson in a `spec` field with denormalized query fields.

6. `internal/repo/config/pg/pg.go`
7. `internal/repo/invocation/pg/pg.go` (if decided to migrate later)
8. `internal/repo/agentfile/pg/pg.go`

### Phase 4: Complex Repos

9. `internal/repo/forum/pg/pg.go` — transactions + proto direct storage
10. `internal/runtime/cron/pgrepo.go` — cron_jobs only (executions stay Mongo)

### Phase 5: Integration

11. Update `internal/app/channels.go` switch-case to add `"postgres"` backend
12. Update `internal/app/config_store.go` for PG config repo
13. Update health check to ping both PG and Mongo

### Phase 6: Testing

14. Add PG integration tests (can reuse patterns from memory implementation tests)
15. Add `docker-compose.yml` dev setup with PG + Mongo + Redis

## SQL Schema Examples

### workspaces

```sql
CREATE TABLE workspaces (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspace_members (
    workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      TEXT NOT NULL,
    role         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);
```

### users & auth

```sql
CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL DEFAULT '',
    avatar_url    TEXT NOT NULL DEFAULT '',
    email         TEXT NOT NULL DEFAULT '',
    provider      TEXT NOT NULL DEFAULT '',
    external_id   TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT '',
    disabled      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_provider_external
    ON users(provider, external_id)
    WHERE provider != '' AND external_id != '';

CREATE TABLE auth_sessions (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    token_hash  TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ,
    revoked     BOOLEAN NOT NULL DEFAULT FALSE
);
```

### config (generic pattern for all 7 tables)

```sql
CREATE TABLE config_agents (
    id           TEXT PRIMARY KEY,  -- composite: workspace_id:name
    workspace_id TEXT NOT NULL,
    name         TEXT NOT NULL,
    spec         JSONB NOT NULL,
    UNIQUE(workspace_id, name)
);
```

### forum

```sql
CREATE TABLE forum_threads (
    id           TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    title        TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT '',
    labels       TEXT[] NOT NULL DEFAULT '{}',
    metadata     JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_forum_threads_ws_updated ON forum_threads(workspace_id, updated_at DESC);
CREATE INDEX idx_forum_threads_ws_status ON forum_threads(workspace_id, status);

CREATE TABLE forum_posts (
    id           TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    thread_id    TEXT NOT NULL REFERENCES forum_threads(id) ON DELETE CASCADE,
    author       TEXT NOT NULL DEFAULT '',
    content      TEXT NOT NULL DEFAULT '',
    metadata     JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_forum_posts_thread ON forum_posts(workspace_id, thread_id, created_at);
```

## Estimated Effort

| Category | Files | Complexity |
|---|---|---|
| SQL migrations (20 tables) | ~10 migration files | Medium |
| PG implementations (8 packages) | ~10 .go files | Medium-High |
| Startup/config changes | ~5 files | Low |
| Integration tests | ~8 _test.go files | Medium |

Total: ~35 files new/modified, ~2500-3500 lines of code.

## Open Questions

1. **Composite ID pattern**: Keep `workspace_id:name` as single TEXT PK, or switch to
   composite PK `(workspace_id, name)`? Recommend composite PK for PG.
2. **Migration tooling**: `goose` vs `golang-migrate`? Both work; `goose` has simpler Go
   integration.
3. **Dual-write period**: Do we need a dual-write phase for zero-downtime cutover, or is
   a stop-the-world migration acceptable given the data volumes are small?
