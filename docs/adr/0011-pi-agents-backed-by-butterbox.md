# ADR-0011: Pi agents are a first-class agent type backed by a ButterBox resource

- Status: Accepted
- Date: 2026-08-25
- Issue: #307 (butter), orvice/butter-box#4 (PiService), orvice/butter-box#6 (cwd), orvice/butter-box#9 (async turn API)
- Builds on: ADR-0005 / ADR-0008 (credentials outside the public model), research in `docs/research/pi-rpc-integration.md`

## Context

We want butter agents backed by the [pi coding agent](https://github.com/earendil-works/pi) running on ButterBox VMs. butter-box exposes pi over ConnectRPC (`PiService`): one supervised `pi --mode rpc` process per active session, pi's own session IDs, idle stop + `--session` re-attach, sessions visible in pi-web.

Two prior integration shapes were considered and rejected along the way:

- **pi-web as the integration surface** — it is a human UI with a private,
  undocumented protocol. The machine entry point is pi's RPC protocol (via
  butter-box); pi-web stays the human window onto the same session data.
- **A RemoteAgent protocol (like OPENCODE_HTTP)** — a RemoteAgent is only
  reachable through a host LLM agent that transfers to it, which costs an
  extra model round-trip per turn and does not match the product shape:
  "create an agent that *is* pi on this box, bound to this directory".

A pi run also breaks two assumptions the opencode bridge gets away with:
runs are long (tens of minutes, not one HTTP request), and continuity
matters (pi holds its own conversation, tools state, and working tree).

## Decision

### 1. ButterBox is a workspace-scoped resource; the token never enters the proto

`ButterBox` (name, base URL, enabled, `credential_state`) with
`ButterBoxService` CRUD and a status probe (healthz + active-session count).
Workspace-scoped like MCPServer/RemoteAgent — a box carries a filesystem and
provider credentials, which is tenant data; sharing one box across
workspaces would break isolation at the disk. The bearer token is encrypted
through a credential seam (secretbox), following ADR-0005/0008 rather than
opencode's plaintext `password` precedent.

Deleting a ButterBox is refused while any PI agent references it, listing
the agent IDs. **butter never deletes data on the box**: deleting a PI agent
or a butter chat session leaves pi session files in place — they are
history, visible in pi-web, owned by the box.

### 2. `AGENT_TYPE_PI` is a leaf agent type with a nested config

`AGENT_TYPE_PI = 6` with `PiAgentConfig pi` nested in `AgentConfig`
(`butterbox_id`, `working_dir`, `provider`, `model`, `thinking_level`,
`max_run_seconds`). Nested-config-per-type follows the Workflow precedent;
`type = PI ⇒ config.pi` required, `child_agent_ids` forbidden (pi is a
leaf). A PI agent may appear anywhere an agent goes — sub-agent, workflow
node, Sequential/Parallel member — with no special casing: workflow node
timeouts and `max_run_seconds` compose through the shared ctx-cancel path.

Model identity is pi's, not butter's: `provider`/`model` are passed through
to the box (dropdown fed by `PiService.GetAvailableModels`, empty = box
default), ModelProviders are not involved, and Telegram model switching is
locked by the existing empty-selectable-list semantics.

The same boundary applies to the whole behavior surface: **a PI agent's
tools and instructions are configured on the box, not in butter.** pi
deliberately has no MCP support (its tool surface is built-ins + extensions
+ skills, loaded from `~/.pi/agent/` and the working directory's `.pi/`,
gated by project trust), and behavior customization lives in the working
directory's `AGENTS.md` / `SYSTEM.md`. Accordingly, `type = PI` **rejects on
write** — rather than silently ignoring — `instruction`,
`global_instruction`, `mcp_servers`/`mcp_server_ids`, `skills`,
`file_mounts`, `context_guard`, and `remote_agent_ids`, with an error that
points at the box-side configuration (a dangling MCP binding that never
takes effect is worse than a refusal). pi's own context management replaces
butter's context guard. Supporting `instruction` via pi's
`--append-system-prompt` was considered and deferred: it would add a
butter-box API field and drag the instruction into the session-recreate
key; box-side `AGENTS.md` is the single place behavior is defined.

### 3. One pi session per (butter session × agent), keyed in session state

The bridge stores `{pi_session_id, butterbox_id, working_dir}` in ADK
session state under `pibox:<agent_id>`. Each turn sends only the current
user input — pi holds its own history. On mismatch (the agent was repointed
to another box or directory) or `not_found` (the box lost the session), the
bridge **abandons and recreates**; sessions are never migrated, because a
session's file references are only honest in the directory it ran in.
Switching agents in one butter session and back resumes the pi session,
matching Telegram's existing per-agent-session semantics.

### 4. Turns are asynchronous: submit + poll a cursor, never hold a connection

The unary `SendMessage` shape ties the run's life to one idle HTTP
connection — the form most likely to be killed by any proxy in between, and
a dropped connection would *abort* a 20-minute run rather than merely lose a
response. Instead the bridge drives butter-box's async API
(orvice/butter-box#9):

- `SubmitMessage` accepts the prompt and returns an **entries cursor**
  (pi's session leaf id — stable across box restarts);
- `GetTurn(cursor, wait_seconds)` long-polls for settlement; completion is
  judged by entries after the cursor, so a box restart mid-run yields an
  honest "did not finish" instead of a stale previous answer;
- ctx cancel or the `max_run_seconds` deadline (default 1800s, 0 =
  unlimited) → `AbortSession`. The deadline exists to bound runaway loops
  that would otherwise hold a Telegram session lease and a box process slot
  forever; long legitimate runs raise the number.

`StreamMessage` is demoted to a best-effort observation stream (future
AG-UI tool activity); the turn result never depends on it. This is a
connection-robustness decision, not end-to-end resumability: butter's own
runner turn is in-memory and does not survive a butter restart.

### 5. Capacity and pauses fail fast

A box at `PI_API_MAX_SESSIONS` rejects the next activation; the bridge maps
ResourceExhausted/Busy to actionable user-facing errors rather than queueing
inside the runner. pi extension UI dialogs remain auto-cancelled on the box
(headless), so approval-style pi extensions take their conservative branch;
mapping them onto butter's Interrupt system is explicitly deferred.

## Consequences

- Registering a box once and referencing it from agents means a token
  rotation touches one resource, and the dashboard can show per-box health
  and session watermark.
- The state-keyed session mapping adds no persistence: session events remain
  the single source of truth (consistent with ADR-0002's spirit).
- pi session files accumulate on the box by design; cleanup is a box-side
  concern (`DeleteSession(purge)` exists for future tooling).
- Client stubs are imported from the `github.com/orvice/butter-box` module —
  one proto source of truth, at the cost of a cross-repo Go dependency.
- Deferred, tracked separately: streaming partials into chat/AG-UI,
  extension-dialog → Interrupt mapping, a directory-picker RPC.
