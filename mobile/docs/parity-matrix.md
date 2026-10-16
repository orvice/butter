# Web → API → App → test parity matrix (#469)

Audit date: 2026-10-16. Branch: `flutter-470-t469`. Evidence is from `front/src/routes`, proto/services, and **grep-verified** mobile code/tests—not README claims alone.

Legend:

| Status | Meaning |
| --- | --- |
| **OK** | App implements the Web operation with structured UI + tests |
| **READ** | Read/list/detail only; write missing or Web-only |
| **GAP** | In scope but missing, partial, or not honestly CRUD |
| **N/A** | Explicitly out of scope for the Flutter product |

## Excluded from App scope (N/A)

| Web area | Web routes (examples) | Why excluded |
| --- | --- | --- |
| Forum | `/forum/*` | Spec §1 / #470 |
| Skills management | (Web skills UI) | Spec §3; Agent editor may **ListSkills** for pick-only |
| Global administration | `/admin/*`, `/users` | Spec §1; Git Host CRUD, global MCP CRUD, platform Telegram/Linear endpoints, user admin |
| Multi-server | — | Single saved server address only |
| Legacy web-chat / generic channels | — | Retired (#409, #273) |
| `StreamAgent` / async Invocation watch | — | Removed #410; App uses AG-UI |

## Auth, server, profile (#447–#449)

| Web operation | RPC / REST | App entry | Tests | Status |
| --- | --- | --- | --- | --- |
| Verify server | `GET /ping`, `ListOAuthProviders` | Server form | `client_context_test`, `contract_test` | OK |
| Password login / logout | `AuthService` | Login / sign out | `client_context_test`, `contract_test` | OK |
| Native OAuth | `BeginOAuthFlow(native)`, `RedeemOAuthHandoff` | OAuth buttons | `native_oauth_contract_test` | OK (no device browser leg) |
| Profile / password | `AuthService`, uploads | My profile | `management_*`, `profile_widget_test` | OK |
| Avatar upload | `POST /api/uploads/avatar` | Profile | `management_context_test` | **GAP**: production S3 success not verified on device |
| Appearance / display prefs | local prefs | Settings | `preferences_test`, `settings_widget_test` | OK |
| Workspace list/select/create | `WorkspaceService` | Workspace picker | `management_*`, `contract_test` | OK |
| Workspace edit/delete/members | `WorkspaceService` | Manage Workspace | `management_*`, `contract_test` | OK (owner/global admin write; ws admin cannot edit workspace metadata) |

## Overview, operations, tokens (#468)

| Web operation | RPC / REST | App entry | Tests | Status |
| --- | --- | --- | --- | --- |
| Dashboard overview (global) | `DashboardService` | Overview (admin user) | `ops_contract_test` | OK (labeled server-wide) |
| Workspace-scoped overview | List*/Get* aggregates | Overview | `overview_controller_test`, `ops_contract_test` | OK |
| Invocations list/cancel | `ListAgentInvocations`, `CancelAgentInvocation` | Operations | `ops_contract_test` | OK |
| Personal API tokens | `APITokenService` (USER kind) | API tokens | `ops_contract_test`, `management_permission_matrix_test` (controller gates) | OK |
| Integrations hub / copy URLs | derived + navigation | Integrations | manual via `ops_pages` | OK |

## Chat, threads, reconnect, A2UI (#460–#462, #461)

| Web operation | RPC / REST | App entry | Tests | Status |
| --- | --- | --- | --- | --- |
| Detached AG-UI run, images | `POST /api/agui/:agent_id` | Chat | `chat_run_context_test`, `chat_contract_test` | OK |
| Thread list/rename/delete | `SessionService` + AG-UI | Chat threads | `chat_threads_*`, `chat_widget_test` | OK |
| Attach / Stop / fallback | AG-UI GET run, `/stop`, history | Chat reconnect | `chat_reconnect_test`, `chat_reconnect_contract_test` | OK in fixture; **GAP**: multi-Pod + Redis cross-client |
| A2UI cards / Human Input | `butter.a2ui`, UI snapshot | Chat surfaces | `a2ui_*`, `run_fold_test` | OK in tests; **GAP**: device rendering/a11y |
| Agent avatar in chat | — | Initial letter only | — | **GAP**: Web shows configured avatar |
| Frontend tools | — | Not declared | `run_request_test` | N/A (by design) |

## Agents, workflow, lifecycle, content (#451–#452, #459)

| Web operation | RPC / REST | App entry | Tests | Status |
| --- | --- | --- | --- | --- |
| List/get/create/edit agents (LLM/composite) | `AgentService` | Agents hub | `agent_*`, `workflow_draft_test` | OK |
| Workflow structured edit | `AgentConfig.workflow` | Agent editor | `workflow_draft_test`, `workflow_runtime_test.go` | OK (no drag canvas) |
| Pi/Cursor leaves | `AgentService` + box | Agent editor | `box_*` | OK |
| Tombstone/restore/purge/operations | `AgentService` lifecycle RPCs | Agent detail | `agent_lifecycle_*` | OK |
| Git Agent Content / CAS save | `UpdateAgentConfiguration`, commits | Edit Agent / Git content | `agent_lifecycle_contract_test` | **GAP**: real Git-bound deployment overlay not verified here |
| Skill association | `ListSkills` + config | Agent editor checkboxes | `agent_association_test` | READ (pick only) |

## Model providers, MCP, remote, daemon, boxes (#450, #453–#455)

| Web operation | RPC / REST | App entry | Tests | Status |
| --- | --- | --- | --- | --- |
| Model provider CRUD | `ModelProviderService` | Model providers | `model_provider_*` | OK |
| MCP server CRUD, probe, OAuth | `MCPServerService`, OAuth HTTP | MCP servers | `mcp_*`, `mcp_contract_test` | OK; **GAP**: real IdP / system browser on device |
| Install global MCP preset | `GlobalMCPServerService` | Browse presets | `mcp_contract_test` | READ (install only; no global CRUD) |
| Remote agent CRUD + daemon bind | `RemoteAgentService` | Remote Agents | `remote_agent_*`, `daemon_*` | OK |
| ButterBox CRUD, catalogs | `ButterBoxService` | ButterBoxes | `box_*` | OK |
| Daemon runtime/tasks/tokens | `DaemonService`, `APITokenService` | Daemon execution | `daemon_*` | OK; **GAP**: plain member can mint worker token but not revoke (server) |

## Files, memory, Git (#456–#458)

| Web operation | RPC / REST | App entry | Tests | Status |
| --- | --- | --- | --- | --- |
| File space/file CRUD, versions | `AgentFileService` | File spaces | `file_*`, `files_contract_test` | OK (text UTF-8; no CAS) |
| Agent mounts / associations | `AgentService` | Agent editor | `agent_association_test`, `files_runtime_test.go` | OK |
| Workspace memory config + browse | `WorkspaceMemoryConfigService`, `WorkspaceMemoryService` | Memory | `memory_*` | OK |
| Agent memory / result cards | `Agent.config` | Agent editor | `memory_draft_test`, `agent_contract_test` | OK |
| Git binding onboarding/detach | `WorkspaceRepoBindingService` | Git repository | `repo_binding_*` | OK (fake git provider in contract) |
| List Git hosts | `GitHostService` | Host picker | `repo_binding_contract_test` | READ (list/get only) |

## Sessions, cron, automation, notify (#463–#464, #467)

| Web operation | RPC / REST | App entry | Tests | Status |
| --- | --- | --- | --- | --- |
| Session explorer + reply | `SessionService` | Sessions | `session_*` | OK |
| Cron CRUD, run now, WAITING_INPUT reply | `CronJobService`, `ReplySession` | Cron jobs | `cron_*` | OK |
| Automation CRUD, manual run | `AutomationService` | Automations | `automation_*` | OK |
| Automation triggers (webhook/forum/…) | — | Editor (disabled) | `automation_draft_test` | **GAP**: triggers not executable on server; UI disables |
| Notify group CRUD | `NotifyGroupService` | Notify Groups | `notify_group_*` | OK |

## Telegram, Linear (#465–#466)

| Web operation | RPC / REST | App entry | Tests | Status |
| --- | --- | --- | --- | --- |
| Channel/destination CRUD, test send | Telegram services | Telegram hub | `telegram_*` | OK |
| Processing resend-only | `ResendTelegramReply` | Processing detail | `telegram_contract_test` | OK |
| Platform webhook base (global admin) | `TelegramAdminService` | Hint on list | `telegram_contract_test` | READ (members see hint only) |
| Linear app/install/processing | Linear services | Linear | `linear_*` | OK |
| Linear public base URL admin | `LinearAdminService` | Banner / copy fields | `linear_widget_test` | READ |

## Permission visibility (#469)

| Check | Evidence |
| --- | --- |
| Member vs owner vs workspace admin gates | `management_permission_matrix.dart` (table), `management_permission_matrix_test.dart`, `management_permission_matrix_widget_test.dart`, `management_*` |
| Global admin without `/admin` navigation | `management_permission_matrix.dart` (`excludedGlobalAdminNavLabels`), `management_permission_matrix_test.dart`, `management_widget_test` (no Create user) |
| Context isolation (late response, workspace switch, server change) | `agent_context_test`, `chat_threads_context_test`, `model_provider_context_test`, `client_context_test` (in-flight list after `setServer`) |

## Summary of GAPs (in-scope, not done or partial)

1. **Cross-environment**: Web↔App on same thread with multi-Pod Redis (attach, Stop, fallback, form recovery, resend-only)—requires live deployment; not verified in CI.
2. **Device-only**: secure storage, background/lock/kill, OAuth browser return, images, accessibility, iOS ATS with real HTTPS.
3. **Chat**: remote Agent avatar image; optional camera entry undecided.
4. **Git Agent Content**: Active Revision overlay on real Git-bound server (#451 follow-up in ticket-469 notes).
5. **Avatar**: multipart to production object storage.
6. **Automation**: non-manual/schedule triggers disabled in UI (server does not run them).
7. **Daemon worker tokens**: member mint vs owner revoke asymmetry (server behavior, documented in README).
8. **Product/release**: App ID, min OS, languages, tablet target, distribution, signing owner, macOS CI access—**pending PO** (see `release.md`).

Detailed per-ticket mappings remain in `mobile/README.md` sections; this matrix is the audit rollup for #469.
