# 用 MyGo 原生 UI 做 Butter Desktop：调研记录

调研日期：2026-11-08 至 2026-11-09。**这是调研记录，不是已接受的设计，也不是已实现的功能。** 第一阶段的范围决策见 [`docs/desktop-macos-phase1.md`](../desktop-macos-phase1.md)。

## 结论

- **技术路线成立。** MyGo 的原生 UI 是"Go 写界面、自己用 GPU 画"，不经 HTML/JS/WebView，因此"不使用 WebView"这条硬约束能满足；macOS 上还顺带拿到原生菜单、托盘、Keychain、VoiceOver 与签名/公证打包。
- **SoT 约定可复用。** Butter 的 proto 与生成的 Go ConnectRPC client 可直接用于桌面端；鉴权与 Web 相同（`Authorization: Bearer` + `X-Workspace-ID`）。
- **#525 里三处判断需要修正**（§4）：Go 侧 AG-UI SSE 并非从零（Butter 已依赖的 community Go SDK 自带客户端 SSE 与事件 decoder）；线程安全规则现在就能核对，不必留作待办；长回复渲染不是"待实测"，而是上游已记录的缺陷。
- **一个会正面冲击验收的上游缺陷**：#1471 —— 长 Markdown 回复在虚拟列表里整段建树，且跨消息文本选择在滚动后丢失（§6）。
- **第一阶段建议只做 macOS**，因为 Windows/Linux 需要额外的渲染后端、GTK3 与安全存储降级验证，而收益不能提前兑现。决策与验收见 [`docs/desktop-macos-phase1.md`](../desktop-macos-phase1.md)。

## 1. 来源与验证范围

本次对照的一手来源：

- MyGo：tag `v0.8.3` = commit `5add3817c74bb7505d0b473bc5fb28abe2201b99`（2026-11-07），含 `docs/`、`ui/`、`ui/editor`、`ui/markdown` 与 `examples/`。本文行号/文件路径都按该 commit。
- MyGo 的 open issues（2026-11-09 读取，共 16 个），其中 #1471、#1304、#105 与本文相关。
- Butter：`7c6a32f`。相关一手材料：`docs/api.md`（AG-UI、Detached Run、Run Log、A2UI）、ADR-0002/0014/0016、`internal/a2ui/catalog.go`、`internal/handler/http/agui_*.go`、`front/src/features/agui-chat/`、`go.mod`。

**未做的事：** 没有构建或运行任何 MyGo 应用，没有在 macOS / Windows / Linux 上跑过原生 UI，没有实测输入法、VoiceOver、GPU 后端；也没有对 Butter 发起 AG-UI 请求。所有"行为"结论都来自读文档、读代码与上游 issue 里给出的实测数据。未验证项逐条列在 §8。

## 2. Butter 侧可复用的部分

| 能力 | 现状 | 依据 |
| --- | --- | --- |
| RPC schema 与 Go 客户端 | 已有生成的 protobuf + Connect handler/client | `pkg/proto/agents/v1/agentsv1connect` |
| 鉴权 | `AuthService.Login`（公开，用户名+密码，bcrypt）→ session token；`Authorization: Bearer` + `X-Workspace-ID` | `docs/api.md:1676-1702`、`internal/application/auth_service.go:63` |
| 会话时长 | `sessionTTL` 默认 7×24h，无 refresh token；`Logout` 撤销；token 以 sha256 存 Redis | `internal/application/auth_service.go:46-50`、`docs/api.md:498` |
| AG-UI 运行 | `POST /api/agui/:agent_id`；`forwardedProps.butterRun={detach:true}` 走 Detached Run | `docs/api.md`、ADR-0016 |
| 重连 | `GET …/threads/:thread_id/run` attach，从 `RUN_STARTED` 全量重放，无日志时 204，不可跟随结束时发 `CUSTOM butter.fallback` | `docs/api.md:928-980`、ADR-0016 |
| Stop | `POST …/threads/:thread_id/stop`，202 = 接受、204 = 无可停 | `docs/api.md:880-912` |
| 历史与 UI Snapshot | `GET …/messages`、`GET …/ui`；运行中返回 `running` 并在 run 起点截断 | `docs/api.md:1408-1496` |
| A2UI | `butter-basic-v1`；模型只读组件 Text/Card/Divider/KeyValue/Status，表单组件 TextField/ChoicePicker/Button 仅由服务端表单生成 | `internal/a2ui/catalog.go:60-86` |

**Go 侧已有的 AG-UI 代码资产：** Butter `go.mod` 已 pin `github.com/ag-ui-protocol/ag-ui/sdks/community/go v0.0.0-20260813165816-6691ac50b34a`，服务端用其中的 `core/events` 与 `core/types`。该模块还带 `pkg/client/sse`（`Client.Stream` 返回 `Frame` 通道）与 `events.EventDecoder`。可复用，但必须按 Butter 的扩展语义核对：15s `: heartbeat` 注释、attach 的 204、`butter.fallback`、detach 的 runId 一次性（重复 `runId` 返回 409 `run_exists`）、Run Log 上限（50,000 条 / 8 MiB）。

服务端的 Go 侧重建逻辑（`internal/handler/http/agui_history.go`、`agui_state.go`、`agui_runlog.go`）可以作为客户端状态的对照基线，但它们服务于 HTTP 契约，不是可复用的客户端状态机。

## 3. MyGo 能力对照

| #525 的说法 | 核实 | 证据 |
| --- | --- | --- |
| 原生布局 | 成立 | flexbox + grid、`Min/MaxContent`、sticky、z-index |
| 文本输入、输入法 | 成立 | `ui.TextInput`/`TextArea` 走平台输入法；composition 期间 Enter/Esc 属于输入法，不提交；`Password(true)` 屏蔽剪贴板与输入法 |
| 虚拟列表、表格 | 成立（都只构建可视行） | `docs/ui/list.md`、`docs/ui/table.md` |
| Markdown 流式 | 成立，且是为 LLM 回复设计的 | `docs/ui/markdown.md`：增量解析、未闭合语法按"将成形态"渲染、token 成本与位置无关 |
| 代码/富文本编辑器 | 成立 | `ui/editor`（Zed 风格，只布局可视行）、`docs/ui/rich-text-editor.md` |
| 图片 | 成立 | `ui.Image`、`docs/ui/image.md` |
| 窗口/菜单/托盘/文件对话框 | 成立 | `docs/windows.md`、`docs/menus.md`（含 `mygo.NewTray`）、`docs/native.md` |
| 签名更新 | 成立 | `docs/updates.md`：`mygo keygen`，签名清单 + delta 更新 |
| 无障碍 | 成立 | `docs/ui/accessibility.md`：角色/命名/状态/live region 对齐 ARIA，VoiceOver/Orca/Narrator/NVDA，`Tester.Access` 可断言 |
| 国际化 | 成立 | `docs/i18n.md`：catalog JSON → 生成的 typed 函数 |
| 系统安全存储 | 成立，但有平台差异 | Keychain（`SecItem*`）/ libsecret Secret Service / Credential Manager；**Linux 无 libsecret 或没有 Secret Service 时返回 `ErrUnsupported`** |
| 深链 / 单实例 | 成立 | `deeplink.go`、`single_instance.go` |
| 纯 Go、无 cgo | 成立 | `README.md`；`go.mod` 只依赖 purego 等 |

**渲染后端：** macOS 用 Metal，Windows 用 Direct3D 11（无驱动时 WARP），Linux 用 GtkGLArea 里的 OpenGL，均可回落到 CPU 渲染（`MYGO_GPU=0`）。原生 UI 在 Linux 只需要 GTK3，不需要 WebKitGTK（`linux.webview: false` 时打包依赖也不含 webkit2gtk），这也是"不用 WebView"在 Linux 上同样成立的原因。

**线程安全（#525 列为待核对，现已确认）：** view 在窗口的主 goroutine 构建；后台结果用 `win.Update(func(){…})` 或 `mygo.RunOnMain` + `Invalidate` 提交；`Invalidate` 明确"safe from any goroutine"；桌面 API 的公开方法都可从任意 goroutine 调用（`docs/ui/views.md:190-212`、`docs/ui/migration.md:184`、`docs/architecture.md:208`）。所以 SSE 读取与 UI 状态提交可以严格分离，但状态字段仍要自己加锁或只在主线程写。

## 4. 对 #525 判断的三处修正

1. **"需要实现 Go 侧 AG-UI SSE 解码"被低估了工作量，但方向说反了。** 传输层与事件解码在已依赖的 SDK 里就有（§2）；真正要写的是 fold（事件 → 回复快照）、follower（Run Log + fallback 状态机）与 A2UI store/renderer。对照 Web 侧代码量：`run-fold.ts` 438 行、`run-follower.ts` 344 行、`a2ui/store.ts` 383 行、`a2ui/catalog.tsx` 408 行——约 1.6k 行 TS 的等价逻辑需要 Go 重写，并逐用例对齐 `run-fold.test.ts`。
2. **"线程安全规则需按锁定版本核对"可以删掉。** 规则在锁定版本里是明确写着的（§3），剩下的是应用自己的状态所有权问题。
3. **"中文输入/长列表只是待实测"不够。** 长回复渲染是上游已记录的缺陷（§6），需要在验收里给出显式处置，而不是等实测再说。中文输入反过来风险更低：文档与 GUI e2e 都覆盖了输入法组合，但真机实测仍是必须的。

## 5. 版本、工具链与发布

- MyGo 仍 pre-1.0（`v0.8.3`），`v0.8.0` 的 release notes 明确列了 breaking changes（例如 `HandleInput` 开始收到所有按键、`InputEvent.Code` 改语义、窗口失焦配色等），并维护 `docs/ui/migration.md`。**必须 pin 到具体 commit，不能跟 `main`**，升级要按迁移文档评估。
- MyGo 的 `go.mod` 要求 `go 1.27.3`，Butter 当前是 `go 1.27.1`。引入该依赖需要全仓 bump 工具链（`go.mod` 一行 + CI 的 Go 版本），属于小改动但会影响所有构建。
- macOS 打包：`mygo build` 出 `.app` + DMG；正式分发需要 Developer ID 签名与 notarytool 公证（`macos.signingIdentity`、`macos.notarize.keychainProfile`），未签名/未公证的包在别的机器上会被 Gatekeeper 拦。`minimumSystemVersion` 默认 12.0。
- Butter 仓库里 `mobile/` 已是"独立模块 + 独立 CI"的先例（`.github/workflows/mobile.yml` 用 `macos-15` 跑 iOS 构建）。桌面端建议同样自成模块，避免把 GUI 依赖塞进服务端 `go.mod`。

## 6. 上游缺陷：#1471（与平台无关）

一条回复在 `ui.List` 里就是一个 row，导致两个同源问题（issue 里的实测：25 KB 回复，149 个顶层 block，800×900，macOS arm64，Go 1.27.3）：

- **整段建树**：`markdown.View` 会把整条回复的所有 block 都建出来，retained heap +10.5 MB；只建首屏（前 2.1 KB）是 +1.0 MB。稳态每帧 184 次分配 / 8.6 KB，对比首屏方案的 20 次 / 0.8 KB。
- **选区不跨回收**：列表回收 row 后，`textSelectionBounds`/`syncTextSelection` 找不到选区端点，选择被清空——拖选后 `ScrollToEnd` 再复制得到空或错位文本；跨消息 Shift 点选同样丢。

这与聊天场景直接冲突，而 `docs/ui/markdown.md` 推荐的 chat 写法正是"一行一条消息"。可选处置：

1. **等上游**（issue 有提案，含两段可独立合并的改动）；
2. **自建 block 级窗口化**：把一条长回复拆成多行渲染，代价是跨段选择、脚注/引用定义与增量解析语义会退化；
3. **降级验收**：把"超长回复 + 跨消息选择"列为已知缺陷并写入 release notes。

相关但**不影响**本方案：#1304（缺少 `writing-mode: vertical-rl`）只关系到竖排文本，Butter 只做横排；它放在这里只作为"文本引擎仍在演进"的证据，以及锁定版本时升级预算的参照。#105（native UI roadmap）同为开放项。

## 7. 未验证清单

- 三平台（尤其 macOS）真机构建、运行、GPU 后端与高 DPI 表现；
- 中文输入法与 VoiceOver / Narrator / Orca 的实际朗读与操作；
- MyGo 的纯 Go HTTP 客户端跑 Butter SSE 的完整行为（理论无风险，但未跑）；
- 目标 GTK3 版本下的 Linux 运行与 libsecret 降级路径；
- #1471 是否会在目标 MyGo 版本上修复、以及自建窗口化的实际收益。

## 8. 结论

技术上可行，工作量主要在客户端状态层而不是渲染或传输。建议按 [`docs/desktop-macos-phase1.md`](../desktop-macos-phase1.md) 的 macOS 第一阶段推进，并把 #1471 的处置作为进入 M2 的前置决策。
