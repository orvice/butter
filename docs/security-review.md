# Security Review — butter

Date: 2026-05-23
Reviewer: Craft Agent (Claude Opus 4.7)
Scope: Repository `go.orx.me/apps/butter` at commit `809f2dd` (main).

This document records the findings of a code-level security review of the
butter service. Findings are grouped by severity. Each entry lists the
affected file, the issue, an exploit scenario, and a recommended fix.

The service is a multi-tenant agent orchestration platform built on the
Butterfly framework and Google ADK. It exposes ConnectRPC services
(Connect / gRPC-Web / gRPC on the same URL) and Gin HTTP routes,
integrates with Telegram/Discord, runs MCP servers (HTTP/SSE
transports — stdio was removed in commit `9f19aad`), and persists
configurations/sessions in MongoDB and Redis.

---

## Critical

### C1. RemoteAgent URL not validated → SSRF

- **File:** `internal/application/remoteagent_service.go:61-92`
  (`CreateRemoteAgent`, `UpdateRemoteAgent`)
- **Issue:** RemoteAgent URLs used by the A2A protocol are stored without
  scheme/host validation. `MCPServer` uses `validateHTTPURL()` in
  `internal/application/mcpserver_service.go:95-103`; the same check is
  missing here.
- **Exploit:** A workspace member creates a RemoteAgent pointing at
  `http://localhost:27017` (MongoDB), `http://169.254.169.254` (AWS
  metadata), or any internal endpoint. Agent invocations that contact the
  remote agent become an SSRF primitive.
- **Fix:** Call `validateHTTPURL()` on `req.GetRemoteAgent().GetUrl()` in
  both Create and Update.

### C2. ModelProvider `base_url` not validated → SSRF

- **File:** `proto/agents/v1/agent.proto:337-338` (`base_url` field) and
  the corresponding `internal/application/modelprovider_service.go`
  Create/Update handlers.
- **Issue:** OpenAI-compatible `base_url` accepted without validation.
- **Exploit:** Attacker points `base_url` at an internal service or
  attacker-controlled host. All future LLM calls leak prompts/system
  instructions/credentials to that endpoint, or probe the internal
  network.
- **Fix:** Apply `validateHTTPURL()` (or equivalent) to `base_url` on
  create/update.

### C3. `InvokeAgent` input not size-limited → DoS

- **File:** `internal/application/agent_service.go:219-282`
- **Issue:** `req.GetInput()` is forwarded into runner/session storage
  without a length check. Connect/gRPC request body limits are
  framework-level defaults; the application layer adds no protection.
- **Exploit:** Repeated multi-megabyte requests exhaust memory and disk
  (sessions are persisted). A single attacker workspace can DoS the
  service.
- **Fix:** Reject inputs above a configurable limit (default 1MB is
  generous for chat) with `connect.NewError(connect.CodeInvalidArgument, …)`
  (or, while the migration adapter is still in place,
  `twirp.InvalidArgumentError`).

---

## High

### H1. Invocation list query uses regex on protojson — NoSQL/tenancy risk

- **File:** `internal/repo/invocation/mongo/mongo.go:62-80` (`List`)
- **Issue:** The repo builds `bson.M{"spec": bson.M{"$regex": ...}}`
  against the protojson-encoded document, using a custom `regexEscape()`
  helper. Even with escaping, regex over JSON is fragile and
  performance-poor; any missed metacharacter or unicode edge case can let
  workspace/agent filters be bypassed.
- **Exploit:** Crafted workspace or agent identifiers leak invocations
  from other tenants in the listing endpoint.
- **Fix:** Project `workspace_id`, `agent_name`, `session_id`,
  `created_at` into top-level BSON fields with proper indexes. Use exact
  match (`{"workspace_id": wsID}`) instead of `$regex`.

### H2. Admin context flag has multiple sources — proliferation risk

- **File:** `internal/handler/http/auth.go:85-91` and
  `internal/repo/auth/repository.go:76-90`
- **Issue:** `auth.WithAdmin(ctx)` is set from session role lookups, root
  token usage, and a legacy "no repos configured" bootstrap path. Every
  new code path that handles authentication risks promoting a non-admin
  user accidentally.
- **Exploit:** Any future regression that wires `WithAdmin` into a
  non-admin code branch silently grants global admin powers
  (cross-workspace install of MCP presets, listing all workspaces, etc.).
- **Fix:** Centralize admin determination behind a single helper that
  takes the authenticated principal and returns a boolean; reject context
  flags that aren't set by that helper. Add unit tests that attempt to
  spoof admin context.

### H3. Global MCP preset install: admin bypasses workspace membership

- **File:** `internal/app/routes.go:436-471` (install global MCP server
  endpoint)
- **Issue:** Line 450 lets admins skip the workspace-membership check and
  install presets into any workspace. Combined with H2, a compromised
  admin session is a one-step path to planting a malicious MCP server
  (with SSRF URL — see C1/C2) in any tenant.
- **Exploit:** Compromised or spoofed admin installs a preset whose URL
  points at attacker infrastructure or internal services. Any agent in
  the target workspace that invokes the MCP server becomes a beachhead.
- **Fix:** Add audit logging for every admin install; consider requiring
  workspace-owner co-sign for installation into workspaces the admin is
  not a member of.

### H4. CronJob does not validate agent ownership at create time

- **File:** `internal/application/cron_service.go:73-90`
- **Issue:** The CronJob is stored with `agent_name`. There is no
  pre-flight check that the agent exists in the caller's workspace. If
  the runtime cron layer resolves agents via
  `*AcrossWorkspaces` queries (agent names are globally unique in this
  iteration), a user in workspace A can schedule the agent of
  workspace B.
- **Exploit:** Cross-tenant invocation; execution logs may also leak via
  the cron-execution listings.
- **Fix:** Resolve the agent during create/update
  (`scheduler.GetAgent(ctx, wsID, name)`); store and reference the
  resolved `agent_id` rather than the name.

---

## Medium

### M1. Incomplete secret redaction when serving installed MCP presets

- **File:** `internal/application/mcpserver_service.go:148-158`
  (`redactInstalledGlobalMCPSecret`)
- **Issue:** Redacts `ClientSecret` only when `oauth != nil`. Other auth
  shapes (API keys, bearer tokens, custom headers) can still surface in
  responses.
- **Fix:** Implement a dedicated redaction routine that knows the full
  set of secret-bearing fields; call it from every read path.

### M2. Telegram `BotToken` interpolated into URL without validation

- **File:** `internal/notify/notifier.go:81-91`
- **Issue:** `"https://api.telegram.org/bot" + token + "/sendChatAction"`
  with no validation. A malicious config can inject characters that break
  out of the URL path (CRLF, slashes, query separators).
- **Fix:** Validate against `^[0-9]+:[A-Za-z0-9_-]+$` and reject
  otherwise.

### M3. API token lookup not defensive about workspace binding

- **File:** `internal/handler/http/auth.go:106-125`
- **Issue:** Trusts the repo to return the correct `workspace_id`.
  Defence-in-depth would assert it is non-empty and matches expectations.
- **Fix:** After lookup, assert `stored.WorkspaceId != ""` and reject
  otherwise.

---

## Low

### L1. Workspace header validation fails silently

- **File:** `internal/handler/http/auth.go:136-163`
- **Issue:** A failed workspace-membership check logs a warning and
  proceeds without setting workspace. Downstream errors are generic and
  mask the actual reason.
- **Fix:** Return a typed error context distinguishing "no such
  workspace" from "no membership".

### L2. Legacy dev fallback promotes everyone to admin

- **File:** `internal/handler/http/auth.go:66-72`
- **Issue:** When `rootToken == ""` and both repos are `nil`, the
  middleware treats every request as admin. A misconfigured production
  deployment (e.g., repos fail to init) silently opens a backdoor.
- **Fix:** Loudly log an ERROR if this path is taken and `ENV != "dev"`;
  consider panicking on startup.

---

## Informational / Positive

- `subtle.ConstantTimeCompare` used for root token comparison
  (`internal/handler/http/auth.go:99`).
- Stdio MCP transport removed in commit `9f19aad` — eliminated arbitrary
  command execution exposure from user-supplied configs.
- Session secret uses `crypto/rand` with 32 bytes
  (`internal/application/auth_service.go:481-486`).
- Default session TTL is 7 days
  (`internal/application/auth_service.go:33-37`); consider 24h plus
  refresh-token rotation for better blast-radius control.

---

## Suggested order of remediation

1. **Now:** C1, C2 — apply existing `validateHTTPURL` to RemoteAgent and
   ModelProvider.
2. **This sprint:** C3 — input size limit on `InvokeAgent` (tracked in
   the worktree spun off from this review).
3. **This sprint:** H1 — restructure invocation storage to use top-level
   indexed fields.
4. **Follow-up:** H2, H3 — centralize admin determination and audit
   admin install actions.
5. **Follow-up:** H4 — make CronJob bind to `agent_id`.
6. **Backlog:** M-tier hardening and L-tier UX/error messages.
