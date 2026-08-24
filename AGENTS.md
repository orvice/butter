# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Build & Run

```bash
# Install dependencies
go mod tidy

# Run the service (requires env vars, see .env.example)
cp .env.example .env && export $(grep -v '^#' .env | xargs)
go run ./cmd/butter

# Build
make build

# Run tests
go test ./...

# Generate protobuf code and inject custom tags (requires buf CLI and protoc-go-inject-tag)
make buf

# Lint protos
buf lint
```

## Architecture

Module: `go.orx.me/apps/butter`

A service skeleton built on `butterfly.orx.me/core` (Butterfly framework) with an agent system powered by Google ADK (`google.golang.org/adk/v2`).

**Workspaces (multi-tenancy):**
Every `Agent`, `AgentChannel`, `MCPServer`, `RemoteAgent`, `ModelProvider`, `NotifyGroup`, `AgentFileSpace`, `AgentFile`, `ForumThread`, `ForumPost`, `CronJob`, `APIToken`, `Invocation`, and `CronExecution` belongs to exactly one workspace. Repo CRUD methods take `workspaceID string` as the first parameter; RPC services derive it from the request context via `internal/workspace.FromContext`. Clients select the active workspace with the `X-Workspace-ID` HTTP header; the auth middleware validates the caller's membership (global admins bypass the check). `AuthService`, `WorkspaceService`, `DashboardService`, and `DaemonService` do not require a workspace header; `SessionService` session CRUD is app/user/session scoped, but `ReplySession` should include `X-Workspace-ID` so the runner resolves agents in the intended workspace. On startup `application.BootstrapDefaultWorkspace` ensures a `default` workspace exists and adds all known users as owners. Repos also expose `*AcrossWorkspaces` listings used by the runtime layers (runner, channel manager, cron scheduler) that operate on the flat global view — agent names are therefore expected to be unique across workspaces in this iteration.

**Layers:**
- `cmd/butter/main.go` — Entry point. Wires config, services, handlers, and registers Gin routes via Butterfly's `core.New()`.
- `internal/app/` — Application bootstrap and wiring. Split by concern: `routes.go` (HTTP + ConnectRPC route setup), `channels.go` (orchestration), `runtime.go` (MongoDB/Redis/Langfuse init), `cron.go` (scheduler init), `system_agent.go` (built-in agent registration).
- `internal/config/` — `AppConfig` is the YAML-loaded startup config. Its `Agents` / `Channels` / `ModelProviders` / `MCPServerConfigs` / `RemoteAgents` fields are tagged `yaml:"-"` because they are not inputs but the flattened runtime snapshot that `app.ConfigStore.SyncToConfig` rebuilds from the config store on every reload; those resources are created through the RPC services and live in the database.
- `internal/handler/http/` — Gin HTTP handlers. Protocol adapters live here alongside their wire format: `a2a.go` (A2A, gated on `enable_a2a`), `openai.go` (OpenAI-compatible, gated on `enable_openai_api`), and `agui.go` + `agui_sink.go` (AG-UI, gated on `enable_agui`). AG-UI is `POST /api/agui/:agent_id` streaming AG-UI events over SSE; its `aguiSink` implements `streamorch.Sink` and derives the `RUN_FINISHED{outcome:interrupt}` signal from `session.Event.RequestedInput` in-stream. Sessions are `agui-{threadId}` scoped to the caller and serialized across Pods by a Redis lease (`internal/runtime/sessionguard` on `internal/redislease`; a busy thread answers 409 before the stream opens, lease loss cancels the run). Client-declared frontend tools ride the run context into a per-invocation long-running ADK toolset (`internal/aguitool`): the run ends after `TOOL_CALL_*` with the call pending, and trailing tool-role messages answer it, validated by `toolCallId` against `interrupt.PendingToolCalls`. Shared state is server-authoritative: the client's `state` mirror is validated and corrected with `STATE_SNAPSHOT`, in-run writes stream as RFC 6902 `STATE_DELTA`, and a trailing session re-read covers `output_key` writes. Only the trailing client message is forwarded, and `resume[].status:"cancelled"` stays rejected (see `docs/api.md` and `docs/research/ag-ui-integration.md`). The dashboard ships a hand-rolled AG-UI chat client (`front/src/features/agui-chat` + `front/src/api/agui.ts`) beside the classic chat path.
- `internal/application/` — RPC service implementations (agent, session, cron, MCP server, remote agent, …). Each service has a `*_service.go` with the business logic. Service methods use native ConnectRPC signatures (`func(ctx, *connect.Request[Req]) (*connect.Response[Res], error)`) and satisfy `agentsv1connect.XxxServiceHandler` directly — `routes.go` hands the service straight to `agentsv1connect.NewXxxServiceHandler(svc, ...)`. Errors are constructed via `connect.NewError` or the `connectx` helpers below.
- `internal/transport/connectx/` — Shared ConnectRPC plumbing: `connect.Error` constructor helpers (`RequiredArgument` / `InvalidArgument` / `NotFound` / `Internal` / `InternalWith`) and the snake_case JSON codec installed via `HandlerOptions()` so the wire format stays compatible with the pre-migration JSON output.
- `internal/service/` — Business logic.
- `internal/repo/` — Data access abstractions.
- `internal/store/config/` — In-memory CRUD store for agent/MCP/remote-agent configurations.
- `internal/agent/` — `NewFromProto()` factory: converts proto `agentsv1.Agent` configs into ADK agent instances (LLM, Loop, Sequential, Parallel, Workflow). Workflow agents are directed graphs of nodes and edges (see `workflow.go`, `workflow_router.go`, `workflow_human_input.go`).
- `internal/runtime/runner/` — Agent runner service managing per-channel ADK runners. `workflow_resume.go` gates implicit resume to workflow-bearing agents; the pending-Interrupt derivation and FIFO reply-matching it relies on live in `internal/runtime/interrupt`.
- `internal/runtime/interrupt/` — The single seam for a paused workflow's human-input state: `Pending` derives unanswered Interrupts from session events (FIFO, oldest-first) and `Resume` rewraps a plain-text reply as the oldest Interrupt's `FunctionResponse`. Consumed by the runner (implicit resume) and, via `TurnResult.Pending`, the cron scheduler (WAITING_INPUT finalization). Holds no state — session events remain the single source of truth (ADR-0002).
- `internal/runtime/cron/` — Cron scheduler for automated agent execution.
- `internal/runtime/session/` — Session persistence (MongoDB implementation).
- `internal/runtime/memory/` — Memory persistence (MongoDB implementation).
- `internal/channel/` — Platform channel implementations (Telegram, Discord).
- `pkg/agent/` — Thin wrapper around ADK `agent.Agent`.
- `pkg/proto/agents/v1/` — Generated Go code from protos. **Do not edit.**

**Proto definitions** live in `proto/agents/v1/`:
- `agent.proto` — Agent tree config: `Agent`, `AgentConfig`, `LLMAgentConfig`, `MCPServer`, workflow agent configs (Loop, Sequential, Parallel), and the graph Workflow Agent config (`WorkflowConfig`, `WorkflowNode`, `WorkflowEdge`, `WorkflowNodeKind`, `WorkflowRetryConfig`).
- `agentchannel.proto` — **Deprecated** platform bindings: `AgentChannel`, triggers, delivery, legacy Telegram/Discord config. Retained so historical records decode; nothing starts them (#273). Use `telegram.proto`.
- `cron.proto` — CronJob, CronExecution (including `WAITING_INPUT` status for workflow pauses), CronJobService.
- `skill.proto` — Skill (agentskills.io bundle metadata) and `SkillResource`; SkillService CRUD plus the resource RPCs (`ListSkillResources` / `GetSkillResource` / `PutSkillResource` / `DeleteSkillResource`).
- `telegram.proto` — Telegram Channels and Destinations (issue #264). A `TelegramChannel` is one Bot transport: immutable ID/key/Bot ID, receive mode (`WEBHOOK` default, `LONG_POLLING`), independent inbound/outbound desired state, and an optimistic `revision`. A `TelegramDestination` is one exact address under a Channel — `chat_id` plus optional `message_thread_id`, both immutable — carrying the inbound routing and interaction policy (`TelegramDestinationConfig`). Telegram-native IDs are canonical decimal strings so JSON clients keep `int64` precision, and an absent `message_thread_id` means "not a Topic", distinct from any real thread. The Bot Token is deliberately not a proto field: it is encrypted through the credential seam on `internal/repo/telegram` under a database-backed master key (`internal/repo/cryptokey` + `secretbox.Keyring`), so `credential_state` only reports whether a usable one exists. Telegram API access goes through `internal/telegramapi` (a `Factory` per decrypted token; no client cache, so a rotation takes effect on the next call).
  Outbound Telegram delivery is unified behind `internal/telegramsend`: every proactive and interactive send takes a Destination ID, always carries the Destination's optional `message_thread_id`, converts Markdown centrally with a plain-text fallback, and honors Telegram `retry_after`. The only sanctioned raw-address path is `Sender.SendRaw`, reserved for the transport-level `/where` command. Notify Group Telegram targets and Cron `TELEGRAM_DESTINATION` delivery reference a Destination ID; the legacy raw fields and `CRON_DELIVERY_TYPE_CHANNEL` are rejected on write.
  Inbound Telegram runs on Redis Streams (`internal/telegramqueue`) as durable queue infrastructure, not a cache. The public callback `POST /api/telegram/webhook/:channel_id` is reachable on every Pod, bypasses workspace auth, and validates the per-Channel secret in constant time before parsing; HTTP 200 means the event was atomically deduplicated by `(channel_id, update_id)` and appended to the Stream by one Lua script, so a queue failure answers 503 rather than losing the update. `internal/runtime/telegram` holds the routing seam: `Router` matches the exact `(channel, chat, thread)` address, freezes the Destination policy and revision into the event, recognizes `/where` *before* Destination matching so an unconfigured chat can identify itself, and ignores unknown addresses. A Redis lease (`telegramqueue.Lease`) elects the single Pod that reconciles Webhook registration; the callback URL is derived from the global base URL (`TelegramAdminService`, global admins only) and the immutable Channel ID rather than stored.
  `internal/runtime/telegram.DecideInteraction` is the pure policy seam for inbound turns: it rejects Bot-authored, anonymous `sender_chat`, channel-post, automatic-forward, edited, and service messages; applies `allowed_user_ids` (empty admits every real user) and `controller_user_ids` (controllers must also be admitted); evaluates the trigger mode against Telegram entities; strips only this Bot's mention; and derives the session as `tg:{channel}:{destination}:{subject}:{agent}` so two topics never share history and switching Agents never inherits another's conversation. `Orchestrator` re-reads Channel and Destination before invoking, so a Destination disabled after acceptance produces no reply, and routes every response — agent output, command answers, and errors — through the unified sender so it stays in the originating Forum Topic. Workflow Interrupt resume is inherited from the runner's implicit FIFO behavior (ADR-0002) because the turn is dispatched with the Destination-scoped session ID.
  Runtime Agent/Model selection is persisted in Redis (`telegram.RedisPreferenceStore`, keyed per Destination *and session subject*), so a choice survives restarts and any Pod sees it. `ResolveEffective` applies a stored choice only while the current selectable list still allows it — a revoked candidate falls back to the default and is cleared. An empty selectable list locks selection. Switching Agent moves the session (histories stay separate, switching back resumes); switching Model does not. `internal/application/telegram_references.go` blocks deleting an Agent or Model — including dropping a model from a provider — while a Destination names it as a default or candidate, and the error lists the Destination IDs.
  Photos are queued as metadata only — Redis holds the `file_id`, never bytes — and downloaded by the worker immediately before invocation (`telegram.DownloadPhoto`), which enforces the 20 MiB limit and the supported image MIME set *before* the Agent runs; a download failure answers in the topic instead of letting the Agent respond to an image it never saw. Media groups are deliberately not aggregated: each update is its own interaction. Long responses go through `telegramsend.NewDelivery` + `DeliverSegments`, which splits on rune boundaries preferring paragraph/line/word breaks, edits the processing placeholder into the first segment, sends later segments in order, carries the topic on every segment, quotes only the first, falls back to plain text per segment, and leaves unsent segments pending so a retry resumes without re-running the Agent.
  Each accepted update gets one `TelegramProcessingRecord` (`internal/repo/telegramprocessing`, unique on `(channel_id, update_id)`, 30-day TTL) whose state answers one question: may the Agent have run? Everything before that boundary retries automatically; `FAILED_UNCERTAIN` after it is dead-lettered for a person and never re-run (ADR-0009). The complete response is persisted before delivery starts, so `ResendTelegramReply` resends only unsent/failed segments and never invokes the Agent — there is deliberately no rerun action. A Redis session lease serializes turns within one derived session while leaving unrelated sessions parallel.
  Long Polling (`telegram.Poller`) is an alternative *transport*, not an alternative pipeline: a `getUpdates` batch feeds the same Router, snapshot, Redis Stream, and workers as a webhook callback. One Pod per Channel is elected by a Redis lease; the offset lives in Redis (`telegram.RedisOffsetStore`) and advances only past updates that were durably accepted or explicitly ignored, with a Lua script fencing commits by lease ownership so a stale leader cannot confirm work the new one is still doing. The webhook reconciler deletes any stale registration for a `LONG_POLLING` Channel, because Telegram answers `getUpdates` with 409 while a webhook is set.
  The legacy generic `AgentChannel` runtime is retired (#273): `channel.Manager` starts nothing and only reports leftover records, `ChannelService` mutations return `Unimplemented` while reads and delete remain for cleanup, Cron `CHANNEL` delivery is not executed, and YAML no longer configures Telegram or Discord. Legacy records are deliberately not migrated — a generic channel has a chat allowlist, not an exact address, so there is no honest Destination to derive.
- `githost.proto` — GitHost: platform-admin-configured allowlist of Git endpoints (GitHub/GitLab kinds, API base URLs); GitHostService (reads open to authenticated users, mutations global-admin only).
- `repobinding.proto` — WorkspaceRepoBinding: zero-or-one per-workspace repository binding (host, repository, branch, root path, write mode, validation status); WorkspaceRepoBindingService (member read, owner/admin manage). The PAT is deliberately not a proto field — it is encrypted (`internal/secretbox`, key from `git.encryption_key`) and stored through a separate credential seam on `internal/repo/repobinding` (ADR-0005). Git access goes through the provider-neutral `internal/gitprovider` contract (GitHub + GitLab REST adapters, no clone/CLI). `OnboardWorkspaceRepository(mode)` does the one-time DB↔Git reconciliation on first bind — `EXPORT_CURRENT` (DB content → validated commit → publish) or `IMPORT_REPOSITORY` (adopt matching Agent IDs) — and `DeleteWorkspaceRepoBinding` is a safe detach: it materializes the Active Revision back into DB fields and reloads runtime before removing the binding/PAT, refusing without a valid snapshot unless a `RepoBindingDetachRecovery` is chosen (ADR-0007; `internal/application/repobinding_onboarding.go`).

Code generation is configured via `buf.gen.yaml` (outputs to `pkg/proto/`). Plugins: protobuf-go, gRPC, gRPC-Gateway, ConnectRPC, validate, and bufbuild/es for the frontend. Twirp generation and runtime dependencies were removed in ConnectRPC Phase 3.

**Config** is loaded by Butterfly from the YAML file pointed to by `BUTTERFLY_CONFIG_FILE_PATH`. The repository sample is `config.yaml`; deployments may copy it to `config/butter.yaml` or another path. Tracing uses OpenTelemetry (`BUTTERFLY_TRACING_PROVIDER`, `BUTTERFLY_TRACING_ENDPOINT`).

## Documentation

Docs directory layout:

- `docs/api.md` — App developer API reference and handoff doc: auth, workspace headers, ConnectRPC URL/field conventions, TypeScript Connect-Web examples, REST uploads (`/api/uploads/*`), `AgentService.StreamAgent` chat stream, the AG-UI endpoint (`POST /api/agui/:agent_id`), and errors.
- `docs/app.md` — Product/function overview in Chinese, including workspace multi-tenancy, agent orchestration (LLM/Loop/Sequential/Parallel/Workflow), Workflow Agent graph nodes and human-in-the-loop pauses, model management, MCP tools, remote agents, daemon execution, cron WAITING_INPUT, and channel entry points.
- `docs/architecture.md` — System architecture overview covering multi-tenancy, process entry, layered structure, startup wiring, agent construction (including Workflow Agent graph building and validation), runner execution flow with implicit workflow interrupt resume, and cron WAITING_INPUT handling.
- `docs/dashboard-api-gap.md` — Dashboard backend API gap analysis, including current coverage, recommended API extensions, persistence additions, phased implementation, and compatibility notes.
- `docs/design-daemon-agent.md` — Daemon Agent design proposal with background, goals, architecture analysis, core challenges, incremental implementation plan, end-to-end flow, and file change list.
- `docs/design-dashboard-chat-async.md` — Dashboard asynchronous chat design and invocation observer contract.
- `docs/project-structure.md` — Project directory structure documentation and maintenance guidance.
- `docs/storage.md` — S3 object storage + static asset / avatar upload configuration and HTTP endpoints.
- `docs/postgres-migration-analysis.md` — PostgreSQL migration analysis and storage-shape notes.
- `docs/prd-workspace-agent-git-content.md` — Draft PRD for workspace Git-backed Agent Content.
- `docs/research/adk-go-v2.1-openai.md` — ADK Go v2.1 OpenAI compatibility research.
- `docs/research/ag-ui-integration.md` — Research verifying the AG-UI integration (issue #286) against the real SDK and ADK v2.1.0: the untagged Go SDK, ADK's lack of native AG-UI support, why the interrupt outcome and addressed resume need no plumbing change, and why `STEP_*` events have no source event.
- `docs/adr/0001-workflow-graph-as-nodes-and-edges-proto.md` — ADR: Workflow graphs as explicit nodes + edges in proto; phase-1 node kinds.
- `docs/adr/0002-interrupt-state-derived-from-session-events.md` — ADR: Pending interrupts derived from session events, FIFO implicit resume.
- `docs/adr/0003-cron-workflow-pause-notify-and-wait.md` — ADR: Cron + Human Input → WAITING_INPUT, notify question, resume via ReplySession.
- `docs/adr/0004-skill-metadata-in-mongo-content-in-s3.md` — ADR: Skill metadata in Mongo, content in S3; skills addressed by name, no versioning in v1.
- `docs/adr/0005-repo-binding-pat-outside-public-model.md` — ADR: Repository binding PATs are not proto fields; encrypted per binding via a dedicated repository credential seam.
- `docs/adr/0006-agent-lifecycle-saga.md` — ADR: Agent lifecycle (create/save/delete-tombstone/restore/purge) as a durable, idempotent Saga (`internal/repo/agentop` + `internal/application/agent_saga.go`) composing the #183 mutation seam and the #217 content-commit seam; soft-delete tombstones + runner exclusion of non-ACTIVE agents.
- `docs/adr/0007-workspace-repo-onboarding-and-safe-detachment.md` — ADR: Workspace onboarding (`OnboardWorkspaceRepository` EXPORT_CURRENT/IMPORT_REPOSITORY) and safe detachment (DeleteWorkspaceRepoBinding materializes the Active Revision into DB fields, reloads, then removes binding+PAT; refuses without a valid snapshot unless a recovery path is chosen). Composes the #216 publication seam and the `ApplyToProto` overlay in reverse; no new persistence.
- `docs/adr/0008-telegram-channel-and-destination-resources.md` — ADR: Telegram Channels are Bot transports and Destinations are exact addresses (`chat_id` + optional `message_thread_id`); pinned Bot identity, index-enforced global Bot uniqueness and per-Channel address uniqueness, immutable addresses, decimal-string Telegram IDs, and database-backed master-key credential encryption.
- `docs/adr/0009-telegram-retry-boundary-and-resend.md` — ADR: the Telegram retry boundary is "may the Agent have run?"; processing-record state machine, output persisted before delivery, resend-without-rerun, per-session lease, and 30-day retention.
- `docs/adr/0010-agent-id-as-logical-persistence-key.md` — ADR: (workspace_id, agent_id) is the Agent repository's logical primary key; legacy Mongo `_id` values stay as opaque physical identifiers ($set updates, no rewrite), name uniqueness moves to an explicit index, migration RPCs retire in favor of the read-only `VerifyAgentIDCutover` verifier, and removal of the internal ADK runtime `Agent.name` is explicitly deferred.
- `docs/adr/0011-pi-agents-backed-by-butterbox.md` — ADR: Pi agents as a first-class `AGENT_TYPE_PI` leaf type backed by a workspace-scoped ButterBox resource (token via credential seam); one pi session per (butter session × agent) keyed in ADK session state, abandon-and-recreate on repoint; asynchronous turns (SubmitMessage + GetTurn entries-cursor long-poll, never a held connection); fail-fast capacity; butter never deletes data on the box.
- `docs/repo-binding-migration.md` — Repository binding migration & operations guide: onboarding/offboarding procedures, failure recovery, and cross-provider acceptance. Its Agent-ID rollout steps are historical — that migration completed and its RPCs are retired (ADR-0010).

## Agent skills

### Issue tracker

Issues live in GitHub Issues (`orvice/butter`, via the `gh` CLI); external PRs are also a triage surface. See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical roles using default label strings (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.
