# Butter mobile — development client (#447)

Independent Android/iOS Flutter app: configure **one** Butter address → password or OAuth sign-in → choose Workspace → paginated authorized Agent list. This is the foundation ticket, not the complete Flutter product. There are no global administration, chat, Skills management or server-list screens.

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
- `tool/contractserver` registers production Connect handlers with `connectx.HandlerOptions()`, real application Login/Me/Logout/ListWorkspaces/ListAgents, real auth/Workspace middleware and `/ping`. Only database persistence is replaced by ephemeral repositories. `test/contract_test.dart` calls those handlers from Dart, including invalid password, unauthorized Workspace and revocation, both at root and at a reverse-proxy `/butter` prefix. It proves actual backend protocol interoperability, not deployment/device acceptance. The fixture cannot bind non-loopback addresses and never executes an Agent.
- Address verification calls `/ping` **and** the generated, public `ListOAuthProviders` RPC without credentials. The second check distinguishes a Butter RPC endpoint from an arbitrary web page. Auth/Workspace requests carry no Workspace header; Agent requests carry the selected `X-Workspace-ID`.
- HTTPS is default, normal OS/Dart certificate validation stays enabled, redirects are refused, userinfo/query/fragment and dot paths are rejected. A deployment prefix is supported; enter the public root/prefix, not `/api`. Proxies must serve `/ping` and Connect beneath that prefix without redirects.
- Generated protobuf messages remain intact in memory: optional absence/explicit zero and unknown fields survive list decoding. This ticket does not write Agent configuration. Future editing must prove enum/optional/oneof/int64/bytes round-trips before offering save; retaining unknown bytes alone is not proof that a form handles future enum values safely.
- One request scope is disposable. Switching server, account or Workspace closes its HTTP client and increments a generation. Late responses cannot restore old identities or lists. Secure-store writes/deletes are serialized, including writes that outlive a cancelled login. Entering the background closes observation/request scope; foreground recovery performs reads only, never repeats Login or side effects.
- No automatic retries, refresh tokens, root/API-token login or persistent Agent cache. 401/Unauthenticated clears sensitive state and secure credentials; 403 remains a permission error. Logout clears locally first and explicitly reports unconfirmed remote revocation. Storage failures remain visible and fail closed; if the OS refuses deletion, the app cannot honestly promise that the OS record was erased.

## OAuth sign-in (#448)

The provider only ever returns to the server's one configured `redirect_url`, which is a web page, so the App cannot be the callback and `BeginOAuthFlow.redirect_uri` is not honored. The App uses a PKCE-bound one-time handoff (`docs/api.md` § Native App sign-in): `BeginOAuthFlow(native{return_uri, code_challenge})` → system browser (`url_launcher`, external application; only `http(s)` URLs are opened) → the web callback page → `ee.orz.butter:/oauth/callback?handoff=…&state=…` → `RedeemOAuthHandoff(…, code_verifier)`. The session token never appears in a URL; the handoff is single-use, expires in 2 minutes and is useless without the verifier that never left the App.

- The pending flow (server, provider, state, verifier, start time) lives in secure storage so it survives the OS killing the App during the browser leg, and expires after 10 minutes. Cancel, a timeout, a stray link with another state or path, a duplicate callback and a late response from a replaced flow all leave the user signed out with a clear message and never repeat a side-effecting call.
- Links arrive through `app_links` (its stream includes the cold-start link). Flutter's own deep-link routing is disabled in the Android manifest and iOS `Info.plist`, so the scheme reaches only this handler.
- **Operator step:** the server must list the return URI in `auth.native_return_uris` (exact match). The default is empty, which disables native sign-in; the App then reports that this server does not support it. `ee.orz.butter` is a **development** scheme, not an approved release identifier. A custom scheme can be claimed by another installed app, which is why PKCE binds redemption to this App instance; choose the release scheme (or verified App/Universal Links) with the other release decisions above.
- Not exercised here: iOS builds and real-device browser round trips (this Linux environment has neither). The Go↔Dart contract test (`test/native_oauth_contract_test.dart`) drives the real handlers with a fake provider only.

## Credential boundary

`flutter_secure_storage 10.0.0`: Android Keystore-backed encrypted storage; iOS Keychain with `first_unlock_this_device` (not migratable to another device), Keychain entitlements in every app build configuration. Android application backup/transfer is disabled. Only the server address is in ordinary preferences. The secure record contains its server binding and token; a mismatched server record is deleted before any authorized request. Passwords are never persisted, printed or placed in URLs, and the password field is cleared on submission. No shared server/account session pool exists.

Changing the address requires confirmation, closes the old scope and clears credentials/cache **before** verifying the new address. This does not call Stop or revoke the old server's Detached Run. Future run observers must live under the same context lifetime.

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

- Connect streaming, AG-UI/SSE, Detached Run/A2UI, uploads and complex configuration writes are not validated by this unary foundation ticket.
- No refresh mechanism exists on the server. Reauthentication is password login, not invented refresh-token rotation.
- Product identifiers, production OS support, distribution/signing, languages and real-device acceptance are still open. The proposed ADR/spec remain untouched and Proposed/Draft.
