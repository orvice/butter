# Flutter release handoff (#469)

This document records reproducible build steps, product decisions still required, known limitations, Spec/ADR alignment, and what **cannot** be verified in the Linux CI agent environment.

## Reproducible build (development artifacts)

Prerequisites: Flutter **3.38.9** / Dart **3.10.8**, Buf **1.71.0** (codegen only), Go from repo root (contract fixture only). Android: JDK 17 + SDK. iOS: macOS + Xcode + CocoaPods.

```bash
export PATH="/home/coder/flutter-3.38.9/bin:/tmp/butter-470-tools:$PATH"
cd mobile
flutter pub get --enforce-lockfile
bash tool/check.sh all          # format, analyze, unit/widget, contract
bash tool/generate.sh --check   # proto drift (CI job `codegen`)
flutter build apk --debug --no-pub
# macOS only:
flutter build ios --simulator --debug --no-codesign --no-pub
```

CI: `.github/workflows/mobile.yml` runs `format`, `analyze`, `test`, `codegen`, `contract`, Android debug APK, and unsigned iOS simulator build on `macos-15`.

**Signing:** No release keystores, provisioning profiles, or App Store credentials belong in this repository. Android `.gitignore` excludes `*.jks` / `*.keystore`. Development builds use debug signing / `--no-codesign`.

## Product owner decisions (not invented here)

Per Spec §10 / #470, confirm before store or internal production rollout:

| Topic | Current dev placeholder | Needs PO confirmation |
| --- | --- | --- |
| Android application ID | `dev.orvice.butter_mobile` | Production ID |
| iOS bundle ID | `dev.orvice.butterMobile` | Production ID |
| Display name | Butter Dev | Store/marketing name |
| Minimum OS versions | SDK defaults (API 24 / iOS 13) | Supported floor |
| Languages | English UI only in tree | zh/en requirement |
| Tablet | Adaptive layout only | Formal tablet acceptance? |
| Distribution | — | Play/App Store vs enterprise |
| Signing owner | — | Who holds keys & CI secrets |
| macOS CI | Workflow exists | Access to runners for release builds |
| OAuth return scheme | `ee.orz.butter` dev scheme | Verified App Links / universal links |
| Native OAuth | Server `auth.native_return_uris` | Operator config per environment |

## Known limitations & release blockers

- **Not verified on physical devices** in this environment (see checklist below).
- **Not verified** against multi-Pod Butter + Redis for cross-client AG-UI attach/Stop/fallback.
- **Not verified** against production S3 avatar upload or real mem0/Git/Telegram/Linear production endpoints.
- Spec status remains **Draft**; ADR-0017 remains **Proposed** (`docs/adr/0017-flutter-app-as-a-full-butter-client.md`) while implementation exists—see discrepancies below.
- Historical note (#451): intermittent local test failure was reported then not reproduced at 116 passed; treat any recurrence as a release blocker until root-caused.

## Spec / ADR vs implementation discrepancies

| Document statement | Implementation reality |
| --- | --- |
| ADR-0017 Proposed; “do not create Flutter project yet” (older spec snapshot) | `mobile/` is implemented through tickets #447–#468 |
| Spec §8 requires zh/en (proposal pending) | App strings are English-only |
| Spec §8 tablet target TBD | Layout adapts; no separate tablet acceptance |
| ADR: no global admin UI | Matches: no `/admin` routes; global admin gets extra **Overview** metrics only |
| ADR: Skills not managed | Matches: `ListSkills` for Agent pick only |
| ADR: single server | Matches: one address in secure/prefs storage |
| Spec §5 cross-client attach matrix | Covered by contract tests + dashboard e2e on Web; **App↔Web live matrix not run here** |

## Verification status (honest)

### Automated in CI / `check.sh all` (this environment)

- Dart format, analyzer, full unit/widget suite, Go-backed Connect contract tests, proto codegen drift, Android debug build; iOS simulator build on macOS workflow (not executed on Linux agent).

### Not verified here (human / staging required)

| Area | Why |
| --- | --- |
| Web + App same AG-UI thread, attach, Stop, fallback, form recovery | Needs Redis + multi-Pod + two clients |
| Resend-only Telegram/Linear on production delivery path | Needs real worker + network |
| Android/iOS images, secure storage, backup/restore | Needs devices |
| Background, lock screen, process kill, foreground recovery | Needs devices |
| OAuth / MCP / Linear browser return to App | Needs devices + configured URLs |
| Screen reader, dynamic type, keyboard on real OS | Needs devices (widget tests are smoke only) |
| Production signing & store upload | Needs keys and PO process |

## Manual release checklist (outline)

Script outline: `mobile/tool/manual_release_checks.sh` (comments only—run steps manually).

1. **Staging server**: HTTPS Butter with AG-UI, native OAuth URIs, test workspace.
2. **Install** release-candidate APK/IPA (signed by release pipeline, not debug).
3. **Auth**: password + OAuth round trip; verify token in Keystore/Keychain; logout clears.
4. **Server change**: change address → confirm old token not sent to new host (`client_context_test` documents intent).
5. **Workspace roles**: member vs admin vs owner on Telegram/Linear/Memory/API tokens/Git binding.
6. **Chat**: send text+image; background app; return and attach; Stop; failed run notice + restore input.
7. **Cross-client**: start run on Web, attach from App (same user/thread); reverse direction.
8. **Integrations**: Telegram test send + resend-only; Linear resend-only (no rerun).
9. **Accessibility**: TalkBack/VoiceOver on chat composer and one large form.
10. **Record**: device model, OS version, build SHA, pass/fail per step; file issues for any GAP in `parity-matrix.md`.

## Native smoke (dev fixture only)

See `mobile/README.md` § Tests and device acceptance for `integration_test/login_test.dart` against `go run ./mobile/tool/contractserver` with `BUTTER_ALLOW_HTTP=true`. Do not point at production user data.
