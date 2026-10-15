# Butter mobile — development client (#447–#451)

Independent Android/iOS Flutter app: configure **one** Butter address → password or OAuth sign-in → choose Workspace → paginated authorized Agent list. Workspace/member management, self-service profile/avatar/password, appearance settings, Model Providers and LLM/Loop/Sequential/Parallel/Workflow Agent configuration are available too. This is not the complete Flutter product. There are no global administration, chat, Skills management or server-list screens.

## Develop

Pinned toolchain: **Flutter 3.38.9 / Dart 3.10.8**, **Buf 1.71.0**, Go from the root `go.mod` (only for the server contract fixture). Install Android SDK/JDK 17 for Android, or Xcode/CocoaPods on macOS for iOS. Existing Go/Web builds do **not** depend on Flutter.

```sh
cd mobile
flutter pub get --enforce-lockfile
flutter run                     # choose Android/iOS; enter address in the App
bash tool/check.sh format
bash tool/check.sh analyze
bash tool/check.sh test
bash tool/check.sh contract      # production Go services on a disposable loopback server
bash tool/generate.sh --check    # regenerate into a temp tree and detect drift
flutter build apk --debug
# macOS only:
flutter build ios --simulator --debug --no-codesign
```

`bash tool/generate.sh` regenerates `lib/gen/` from the root proto schema and locked Buf dependencies. Dart protobuf and Connect generators are resolved from `pubspec.lock`; neither gRPC stubs nor hand-authored wire models are used. Commit generated output; do not hand-edit or reformat it. `.github/workflows/mobile.yml` separately checks format, analysis, tests/contracts, generation drift and both native builds.

### Development settings, not release decisions

- Android `dev.orvice.butter_mobile`, iOS `dev.orvice.butterMobile`; displayed development name is Butter Dev.
- SDK-created development OS baselines (Android API 24, iOS 13) are **not** approved product minimums. Plugin constraints still apply. Confirm IDs, minimum OS, name, languages, distribution/signing and macOS CI access before a production release.
- No certificates, signing credentials or production tokens are in this project. Android debug signing / unsigned iOS simulator builds are development artifacts only.

## Protocol decisions and evidence

- `connectrpc 2.0.0`, `protobuf 6.1.0`, `protoc_plugin 25.1.0`; native `dart:io` HTTP/1.1, **Connect binary unary**, `application/proto`, `/api/agents.v1.Service/Method`. Not gRPC. No JSON adapter was necessary.
- `tool/contractserver` registers production Connect handlers with `connectx.HandlerOptions()`, real application Login/Me/Logout/ListWorkspaces/ListAgents, real auth/Workspace middleware and `/ping`. Only database persistence is replaced by ephemeral repositories. `test/contract_test.dart` calls those handlers from Dart, including invalid password, unauthorized Workspace and revocation, both at root and at a reverse-proxy `/butter` prefix. It proves actual backend protocol interoperability, not deployment/device acceptance. The fixture cannot bind non-loopback addresses and never executes an Agent. `tool/contractserver/workflow_runtime_test.go` (`go test ./mobile/tool/contractserver/`) is the exception for #452: it saves Workflows shaped like the editor's output through those same handlers, reads the stored records back, and runs them on the production runner against a fake OpenAI backend. It asserts that a Router takes its route or default edge, that a fan-out reaches the Join with every branch, that a Parallel Worker runs once per list item, and that a Human Input node pauses with its question and form and resumes into its successor; it also asserts that the server refuses the graphs the editor flags. It does not render the form in the App (separate ticket).
- Address verification calls `/ping` **and** the generated, public `ListOAuthProviders` RPC without credentials. The second check distinguishes a Butter RPC endpoint from an arbitrary web page. Auth/Workspace requests carry no Workspace header; Agent requests carry the selected `X-Workspace-ID`.
- HTTPS is default, normal OS/Dart certificate validation stays enabled, redirects are refused, userinfo/query/fragment and dot paths are rejected. A deployment prefix is supported; enter the public root/prefix, not `/api`. Proxies must serve `/ping` and Connect beneath that prefix without redirects.
- Generated protobuf messages remain intact in memory: optional absence/explicit zero and unknown fields survive list decoding. Agent editing (#451) copies the complete generated message and changes only edited fields; tests cover optional/int64, nested unknown fields and future enum retention. Unknown type/lifecycle and legacy embedded children block editing; unknown Include Contents/ContextGuard strategy is shown as retained and has no rewriting control. Server validation still wins.
- One request scope is disposable. Switching server, account or Workspace closes its HTTP client and increments a generation. Late responses cannot restore old identities or lists. Secure-store writes/deletes are serialized, including writes that outlive a cancelled login. Entering the background closes observation/request scope; foreground recovery performs reads only, never repeats Login or side effects.
- No automatic retries, refresh tokens, root/API-token login or persistent Agent cache. 401/Unauthenticated clears sensitive state and secure credentials; 403 remains a permission error. Logout clears locally first and explicitly reports unconfirmed remote revocation. Storage failures remain visible and fail closed; if the OS refuses deletion, the app cannot honestly promise that the OS record was erased.

## OAuth sign-in (#448)

The provider only ever returns to the server's one configured `redirect_url`, which is a web page, so the App cannot be the callback and `BeginOAuthFlow.redirect_uri` is not honored. The App uses a PKCE-bound one-time handoff (`docs/api.md` § Native App sign-in): `BeginOAuthFlow(native{return_uri, code_challenge})` → system browser (`url_launcher`, external application; only `http(s)` URLs are opened) → the web callback page → `ee.orz.butter:/oauth/callback?handoff=…&state=…` → `RedeemOAuthHandoff(…, code_verifier)`. The session token never appears in a URL; the handoff is single-use, expires in 2 minutes and is useless without the verifier that never left the App.

- The pending flow (server, provider, state, verifier, start time) lives in secure storage so it survives the OS killing the App during the browser leg, and expires after 10 minutes. Cancel, a timeout, a stray link with another state or path, a duplicate callback and a late response from a replaced flow all leave the user signed out with a clear message and never repeat a side-effecting call.
- Links arrive through `app_links` (its stream includes the cold-start link). Flutter's own deep-link routing is disabled in the Android manifest and iOS `Info.plist`, so the scheme reaches only this handler.
- **Operator step:** the server must list the return URI in `auth.native_return_uris` (exact match). The default is empty, which disables native sign-in; the App then reports that this server does not support it. `ee.orz.butter` is a **development** scheme, not an approved release identifier. A custom scheme can be claimed by another installed app, which is why PKCE binds redemption to this App instance; choose the release scheme (or verified App/Universal Links) with the other release decisions above.
- Not exercised here: iOS builds and real-device browser round trips (this Linux environment has neither). The Go↔Dart contract test (`test/native_oauth_contract_test.dart`) drives the real handlers with a fake provider only.

## Credential boundary

`flutter_secure_storage 10.0.0`: Android Keystore-backed encrypted storage; iOS Keychain with `first_unlock_this_device` (not migratable to another device), Keychain entitlements in every app build configuration. Android application backup/transfer is disabled. Only the server address and non-sensitive device appearance settings are in ordinary preferences. The secure record contains its server binding and token; a mismatched server record is deleted before any authorized request. Passwords are never persisted, printed or placed in URLs, and the password field is cleared on submission. No shared server/account session pool exists.

Changing the address requires confirmation, closes the old scope and clears credentials/cache **before** verifying the new address. This does not call Stop or revoke the old server's Detached Run. Future run observers must live under the same context lifetime.

## Workspace, self-service and appearance (#449)

| Web operation | API/storage | App entry | Tests |
| --- | --- | --- | --- |
| Workspace list/create/edit/delete | WorkspaceService | Workspace picker, Create Workspace, Manage | `management_context_test`, `management_widget_test`, real Go `contract_test` |
| Member list/add/role/remove | WorkspaceService | Workspace details; existing user ID, not admin-only ListUsers | same tests |
| My display name/avatar/password/logout | AuthService + `POST /api/uploads/avatar` multipart `file` | My profile | `management_context_test`, `profile_widget_test`, `management_widget_test`, `contract_test` |
| Theme/display preferences | device `butter.appearance` JSON, no identity/secret | Appearance and display | `preferences_test`, `settings_widget_test` |

Permissions come from a fresh GetWorkspace + ListWorkspaceMembers read. Current server rules allow any signed-in user to create a Workspace, members to read, and **owners** (or global admins) to edit/delete and manage members. Workspace admin alone cannot write these resources. Unknown member roles are displayed but not offered in the role editor. No user-directory/admin RPC, user creation/disable or another user's password reset is used.

Edits copy the generated Workspace message, retaining untouched timestamps/unknown fields. Profile avatar optional absence keeps the existing value, explicit empty removes it. The API has no Workspace CAS or last-owner guard; the App does not claim either. Delete confirms Workspace/membership removal and loss of access, not guaranteed deletion of all scoped resources (the current backend does not cascade those collections).

Gallery selection uses `image_picker 1.2.3`, with iOS photo-library purpose text and no camera entry. Selection/preview is local until Save, accepts PNG/JPEG/GIF/WebP by signature and caps reads at 10 MiB; the server's configured limit still wins. Multipart requests carry the user's Bearer token, no Workspace header, refuse redirects and bound responses. Upload followed by UpdateProfile is not atomic: if interrupted, a stored object may be orphaned; no upload or profile write is automatically replayed. A 503 can mean avatar storage is not configured; the form explains that dependency.

All account/Workspace replacements unmount management routes' draft subtrees; foreground recovery refreshes permissions and clears inaccessible management data. Denials preserve safe fields, passwords clear on submission, and interrupted writes retain an unconfirmed-outcome notice through foreground reads. Back warns on unsaved drafts; destructive/access-changing operations name the target and require confirmation. Theme defaults to system; system/light/dark and showing Agent identifiers are device preferences, restored after restart. System font scaling is not overridden, lists/forms scroll above the keyboard and icon actions have reader labels.

The real Go contract fixture now covers Workspace/member CRUD and member/admin denial, profile optional avatar semantics, password checks and revocation at both root and `/butter`. It also exercises the production upload handler's storage-not-configured response. Successful multipart bodies, photo-picker/OS preferences and failure drafts are checked using external-boundary fixtures. A phone-sized widget test exercises 2× system text and reader labels; this is **not** Android/iOS physical-device accessibility acceptance. Language/localization, tablet acceptance and release identifiers/OS/distribution remain undecided. Avatar S3 success, device photo permissions/restart and iOS build still need native/deployment verification.

## Model providers (#450)

| Web operation | API/storage | App entry | Tests |
| --- | --- | --- | --- |
| List / view providers | ModelProviderService List/Get, `X-Workspace-ID` | Agents page → Model providers | `model_provider_context_test`, `model_provider_widget_test`, real `model_provider_contract_test` |
| Create / edit (type, base URL, API key, models with ID, alias, context window) | Create / Update (the whole object is replaced) | Add provider, Edit provider | `provider_draft_test`, widget and contract tests |
| Delete | Delete, after a confirmation naming the provider | Provider details | same tests, plus `mobile/tool/contractserver/main_test.go` |
| Refused change because Telegram Destinations use an alias | `FailedPrecondition` with the alias and Destination IDs | Shown on the form or the delete dialog | contract tests against the real `TelegramReferenceGuard` |

Any member of the Workspace may use these calls: the server has no owner/admin check here, so the App does not invent one. The provider belongs to the selected Workspace; switching Workspace, account or server drops the list, the detail and unsaved drafts.

**Credentials.** The API exposes no `credential_set`, write-only field or connection test, and returns the key in plain text, so the App does not pretend otherwise. It never displays the key: lists and details show only "API key: Set / Not set", logs and errors do not carry it and drafts are not persisted. Editing offers Keep (default, the stored value is untouched), Replace (an obscured field) and Remove (asks for confirmation). Nothing is auto-retried.

**Model fields.** Model ID is what the provider is called with; Alias is the name Agents and Telegram select; Context window is the **input** capacity in tokens (0 to 4,294,967,295). Empty and 0 both mean unset, and the server then uses its own default. It is not the maximum output tokens, and the provider API has no output-limit field, so the App offers none. (An Agent's `context_guard.max_tokens` is likewise an input-context override, not an output limit.) The form needs at least one model, like the Web page.

**Unchanged fields.** An edit starts from a deep copy of the stored message and each model row keeps its own original, so unedited and unknown protobuf fields survive an update, also when rows are removed or reordered. If a save reply is lost the App reports an unconfirmed outcome and reloads instead of replaying the write.

Limits: key rotation policies and a live connection test do not exist in this ticket. The widget tests include a phone-sized 2× text run with reader labels; that is **not** device accessibility acceptance, and no Android/iOS device or release verification was done for this ticket.

## LLM and composite Agents (#451)

| Web operation | API | App entry | Tests |
| --- | --- | --- | --- |
| Paginated Agent list / fresh detail / lifecycle + runtime status | ListAgents / GetAgent / GetAgentRuntimeStatus, by immutable Agent ID | Agents → row | `agent_context_test`, `agent_contract_test` |
| Create / edit LLM, Loop, Sequential, Parallel | CreateAgent / UpdateAgent | Create Agent / detail → Edit Agent | `agent_widget_test`, real `agent_contract_test` |
| View / edit Workflow Agents: AGENT, HUMAN_INPUT, ROUTER, JOIN nodes, START / route / default edges, Parallel Worker, retry and timeout, Human Input forms (text / single-choice fields, order, options, required, limits) | `AgentConfig.workflow` + `child_agent_ids` | Agent detail graph; Edit Agent → structured node, edge and form cards (no drag canvas, no JSON textarea) | `workflow_draft_test` (the draft model, which view and editor share); `tool/contractserver` `go test` saves editor-shaped Workflows through the production handlers and runs them |
| LLM model/alias, instruction/global instruction, schemas, output key, include contents, transfer flags | AgentConfig | Structured LLM fields | widget, draft and contract tests |
| ContextGuard Off / Threshold / Sliding Window | ContextGuardConfig | LLM editor | same tests; real backend rejects negative, wrong-strategy and non-LLM settings |
| Ordered independent child references / labels / metadata / API exposure | child_agent_ids, map rows, interface switches | Structured rows, ID picker, reorder/remove controls | widget and real contract relationship validation |
| Associate/remove existing Skills | SkillService.ListSkills + AgentConfig.skills | Agent editor checkboxes (no Skill CRUD/content/resource management) | widget and real contract catalog/round-trip tests |

Unbound Agent configuration writes are available to Workspace members under the existing backend rules. The App does not invent an owner-only restriction. The immutable ID is the only lookup/reference key; the display name is never resolved. Runtime name is supplied on create and server-controlled thereafter. Existing type is retained by this editor (type selection is offered on create); changing existing type stays with Web to avoid discarding future type-specific fields. Blocked lifecycle states, unknown types, historical embedded-child configurations and a Pi/Cursor leaf that holds children or Butter-side behavior fields (the server refuses them on write, so a record that has them is legacy) are read-only (composite saves would discard embedded children). Pi and Cursor leaves and ButterBoxes were added by #453 (next section). No deletion, restore, purge or operation-management action is provided for Agents.

Model choices include every configured provider's alias (model ID when no alias is set), keep custom/current values, and never cache or display provider keys. Child choices load every Agent page; choosing an ID does not embed or modify that resource. The server enforces existence, Workspace scope, cycles and single-parent ownership. Existing unavailable references remain visible and can be removed, not silently dropped.

**Advanced fields.** Input/output schema text is validated as JSON on edit; it is stored configuration, not a claim of runtime structured-output support. Labels and metadata have separate key/value rows, not a whole-Agent JSON editor. ContextGuard's Threshold has only an Agent Context Override in input tokens (0 inherits Model Context Capacity); Sliding Window has only its content-entry limit (0 defaults to 20). Neither sets maximum output tokens. Switching a known strategy clears the other strategy's field; leaving it unchanged preserves the message including unknown fields. Known backend ContextGuard errors have specific safe explanations; arbitrary server payloads are not echoed. Existing Skills are selectable/removable by name from the read-only Workspace catalog; missing previously selected names remain until deliberately removed. Unedited MCP/Remote/files/Memory/Card Policy/runtime associations remain intact for their separate tickets.

**Git boundary.** Every explicit save first reads GetWorkspaceRepoBinding; an unavailable read never falls back to a DB-only write. Create includes initial Agent Content. An update uses UpdateAgentConfiguration only once the binding has an Active Revision (`active_commit_sha`), matching backend Git ownership. Bound-but-not-onboarded content remains database-owned and editable by members through UpdateAgent. Git-owned updates carry the original int64 version, the active SHA and only changed description/prompt/global-prompt file actions (empty means DELETE). A response that silently ignores DB-path content changes is treated as unconfirmed, not success. CAS conflicts ask for a reload, never force overwrite. Unbound UpdateAgent does not accept a client-observed version, so the App does not promise stale-editor CAS there. Production-service Dart tests cover unbound CRUD, bound-but-not-onboarded member edits, read-only Skill catalogs and real validators at root and `/butter`; bound routing/version/path mapping is tested at the controllable HTTP boundary. A real Git-bound deployment and operation-recovery/device acceptance are **not** established by these tests.

Drafts live only in memory, warn before leaving, and unmount when the server/account/Workspace changes. All fields stay mounted in a scrolling form so off-screen errors cannot bypass validation. Background closes requests; foreground only reads. Ambiguous writes remain explicitly unconfirmed, are never replayed, and the draft refuses another Save until the user leaves and explicitly reloads the Agent. The detail reload reconciles the notice; foreground reads alone do not dismiss it. A phone-sized 2× system-text test covers the large form, not physical-device accessibility acceptance.

## ButterBoxes and Pi / Cursor Agents (#453)

A ButterBox is an agent VM that runs Pi and Cursor leaf Agents. The phone never connects to a Box and never runs anything on one by itself: every call goes to the Butter server, which holds the Box token, probes the Box and asks it for catalogs. A user run happens only when a person sends a chat message.

| Web operation | API | App entry | Tests |
| --- | --- | --- | --- |
| List / view boxes (name, URL, enabled, token set/updated) | ButterBoxService List/Get, `X-Workspace-ID` | Agents page → ButterBoxes → row | `box_controller_test`, `box_widget_test`, real `box_contract_test` |
| Register (name, base URL, enabled, optional token) / edit | Create / Update | Register ButterBox, Edit | same; `tool/contractserver/boxes_runtime_test.go` |
| Token: keep / replace / clear | Create's `token`; `SetButterBoxToken` (empty clears); Update never carries one | Edit keeps it; Replace access token and Clear access token are separate actions (a blank replacement is refused locally) | same |
| Test connection | `GetButterBoxStatus` — an unreachable box is data, not an error | Test connection (explicit; never on open) | same |
| Delete | `DeleteButterBox`; refused with `FailedPrecondition` naming the referencing Agents | Delete after a confirmation | same, against the real reference guard |
| Create / edit a Pi leaf: box, working directory, provider, model, thinking level, run limit | `AgentConfig.pi` via CreateAgent / UpdateAgent | Create Agent → type PI; Edit Agent | `box_leaf_draft_test`, `box_widget_test`, `box_contract_test` |
| Create / edit a Cursor leaf: box, working directory, model, mode, run limit | `AgentConfig.cursor` | type CURSOR | same |
| Model choice | `ListButterBoxModels` (Pi) / `ListButterBoxCursorModels` (Cursor), loaded only when the person taps **Load models from this box** | picker, or typed by hand | same |
| Run a leaf | the existing Chat entry point (`POST /api/agui/:agent_id`) | Chat → leaf | `box_contract_test`, `boxes_runtime_test.go` |

**Run limit.** `max_run_seconds` is `optional`, and the three forms are different on the wire: *Server default* leaves it absent (the server applies 1800 seconds), *Unlimited* sends an explicit `0`, *Set seconds* sends a whole number from 1 to 2147483647. A stored explicit `0` shows as Unlimited and survives an unrelated edit; an invalid number blocks the save instead of becoming one of the others.

**Leaves have no Butter-side behavior.** Their instructions, tools, skills, memory and cards live on the Box, so the editor shows none of those fields, drops any left by an earlier type, and the server rejects them anyway (`InvalidArgument`, checked in `box_contract_test`). Instruction, MCP, Skill, Memory, Card Policy and children are therefore never written for a leaf. A model chosen on one box is cleared when the Agent is pointed at another; the working directory is kept as typed. Unknown fields of the stored Pi/Cursor message are preserved.

**Catalogs.** The status text of an unreachable Box is the server's raw probe error, which can carry whatever the Box or an intermediary returned, so the App only classifies it (token rejected, token unusable, capacity, timeout, cannot connect, otherwise "did not answer") and never displays it. Pi catalog failures (capacity, unreachable) and Cursor failures (`CURSOR_API_KEY` missing or invalid, a Box that does not serve CursorService, unreachable; the Cursor call may take up to 90 seconds) are mapped to messages that say what to do and always leave the model fields open for typing by hand. A disabled box is marked and cannot take a new binding; an Agent already bound to a missing or disabled box stays unchanged unless it is repointed.

**Session continuity (server contract, shown in the editor).** A conversation keeps one session on the Box per (conversation × Agent). Changing the Agent's ButterBox or working directory makes the next message start a fresh session; the old one is never moved or deleted on the Box. Model, provider, thinking level and mode apply to sessions created afterwards. This is the server's behavior and is covered by `boxes_runtime_test.go` and `box_contract_test.dart`.

**Evidence and its limits.** `tool/contractserver/boxes.go` serves the production ButterBoxService, the production AgentService leaf validation, the real runner and the real `pibox` / `cursorbox` bridges. Only the Box itself is an in-process test double that serves the real generated Pi and Cursor Connect handlers (profiles at `/fake-box/ok`, `/capacity`, `/nokey`; a closed port is the unreachable one). It proves the transport, validation, token handling, catalog and run error mapping, continuity and repointing. It is **not** a production Box, a real Cursor/Pi model run, a TLS Box or a device run. Dart widget tests drive a controllable HTTP boundary (`test/support/box_server.dart`) for timing and failure shapes, and the real semantics are asserted in the contract tests.

Writes follow the app-wide rule: an ambiguous answer (a lost reply, a background close mid-write) is shown as unconfirmed, never repeated, and further writes stay off until a read reconciles it. Only a read that **began after** the write ended counts, so a slow read already in flight when a write was lost neither settles it nor replaces what the write left with an earlier snapshot; a read of one Box settles only a write aimed at that Box (a lost create is settled by the list). The same fence applies to a read that fails: once a write has stored newer state, an older read's failure (including a 401 it carries) is dropped instead of clearing the stored Box or marking the list failed, while a failure of the current read still reports and a current 401 still signs out. The token is held only in its dialog or field, sent once, cleared after sending, and never logged, displayed or read back (the server only reports that one is set).

## Chat and threads (#460)

Agents → **Chat** opens the thread list; a thread is one AG-UI conversation (`agui-{threadId}`) with the existing UI binding, so the dashboard sees the same threads.

| Capability | Protocol | Tests |
| --- | --- | --- |
| New thread, send | `POST /api/agui/:agent_id` over plain `dart:io` SSE, always `forwardedProps.butterRun = {detach: true}`; only this turn's trailing user message (text + inline images) is sent, never history | `sse_test`, `run_request_test`, `run_fold_test`, `chat_run_context_test` |
| History / restore | `GET …/threads/:thread_id/messages` (bounded read), images restored as attachments | `history_test`, `chat_widget_test` |
| Thread list, search, pagination, rename | `SessionService.ListSessions` (`appName=agui`, Workspace-scoped; unbound sessions are never listed) and `UpdateSessionTitle` | `threads_test`, `chat_threads_context_test`, `chat_widget_test` |
| Images | jpeg/png/gif/webp, ≤10 MiB each, ≤10 per message, request ≤32 MiB, checked before reading and before sending | `run_request_test`, `chat_widget_test` |
| Markdown, code copy, tables, tool status/results, Agent name and avatar | rendered locally; remote images are never fetched | `chat_widget_test` |

Delivery rules: the SSE parser accepts arbitrary UTF-8 chunking, LF/CRLF/CR, multi-line `data:`, comments (heartbeats) and a bounded event size. A POST is **never** retried. A request that was refused (400 `thread_unbound`, 403, 404, 409, 413, 503) or never connected returns the turn to the composer; one whose outcome is unknown shows an `unconfirmed` notice and is not resent. A dropped stream shows "connection lost — the run may still be going" with a manual **Refresh** (one history read); it is never shown as a Stop. Leaving the page, suspending the app, or changing identity/Workspace only detaches (closes that run's socket).

`test/chat_contract_test.dart` (via `tool/contract.sh`) drives `ClientContext` against the production AG-UI handler, runner, `SessionServiceServer` and detached-run machinery in `tool/contractserver/chat.go`; only the model (a loopback OpenAI-compatible echo) and persistence are replaced. It covers a detached two-send conversation and history in a fresh session, Markdown, an image round trip, server-normalized rename, Workspace isolation, and the `thread_unbound` / `run_exists` refusals, at root and `/butter`. `contract.sh` runs `flutter test` without `--no-pub`; on a Flutter SDK other than the pinned one it re-resolves `pubspec.lock`, so restore it with `git checkout` before committing.

Scope notes: tool-role results only answer client-declared frontend tools, and this App declares none, so a run request carries exactly one user message. An Agent shows its name and an initial, not its configured avatar image, so chat fetches no remote image.

Not provided yet (labelled in the UI): reconnecting to a live run after leaving or a dropped stream, **Stop**, and A2UI forms/cards (an Interrupt shows a "waiting for input" notice only). Workflow Human Input cannot be answered in the App.

## Tests and device acceptance

Unit/controller/widget tests use public application interfaces with fake **external** HTTP/OS-storage boundaries. They cover first setup, persistence, real request headers, permission/empty/error/pagination UI, admin without global navigation, expiry, relogin/server/Workspace races, slow credential writes, vault failure and generated-field retention. The separate Go contract fixture is a real-service HTTP integration test.

Native smoke test (disposable fixture account only):

```sh
# Terminal 1, repo root:
go run ./mobile/tool/contractserver -addr 127.0.0.1:8080
# Terminal 2, Android emulator/device attached by adb:
adb reverse tcp:8080 tcp:8080
cd mobile
flutter test integration_test/login_test.dart -d <device-id> \
  --dart-define=BUTTER_TEST_SERVER=http://127.0.0.1:8080 \
  --dart-define=BUTTER_ALLOW_HTTP=true
# iOS simulator on the same Mac: use its simulator ID and the same URL.
# Prefer a trusted HTTPS tunnel if simulator ATS rejects local HTTP.
```

HTTP is accepted only with an explicit `BUTTER_ALLOW_HTTP=true` **debug** build; release code never permits it. Android's debug-only manifest supports that development mode; iOS has no global ATS exception or TLS bypass. Do not expose the fixture publicly. The native test exercises the actual platform vault, UI forms, real HTTP services, restoration via a fresh context and logout. It deletes development credentials; do not point it at a real user's installation. For ordinary manual login, enter credentials in the App, not build flags.

Cold Android builds use a 2 GiB Gradle heap and a 512 MiB Metaspace cap, one worker and in-process project Kotlin compilation. Allow additional RAM for Flutter, the SDK's compiler and the OS. Earlier local verification built a warm-cache arm64 debug APK (`flutter build apk --debug --target-platform android-arm64 --no-pub`) under a 2 GiB container limit; this does **not** establish that a clean multi-ABI build fits that limit. Production-handler contracts (including path prefixes), analysis and generation drift checks also passed.

This Linux environment cannot run Xcode/iOS builds or native device acceptance. CI provides a macOS unsigned simulator build check, but **an unexecuted workflow is not a passing iOS build**. Android/iOS physical-device vault, backup/restore, app restart and accessibility checks remain required before release. Do not mark those acceptance items complete based on plugin mocks or this README.

## Known gaps for later tickets

- Connect streaming and A2UI are not implemented. AG-UI chat (#460) covers detached runs, history, images and threads, but not live re-attach, Stop or A2UI (see Chat and threads). Agent lifecycle (delete/restore/purge) and the remaining association editors (MCP, Remote Agents, files, Memory, Card Policy) are separate tickets; ButterBoxes and Pi/Cursor leaves are covered by #453. Avatar multipart success is covered at a controllable HTTP boundary, not a production S3 deployment.
- No refresh mechanism exists on the server. Reauthentication uses password or OAuth sign-in; there is no refresh-token rotation.
- Product identifiers, production OS support, distribution/signing, languages and real-device acceptance are still open. The proposed ADR/spec remain untouched and Proposed/Draft.
