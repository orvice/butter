# Butter Desktop macOS 第一阶段：范围、验收与里程碑

- 状态：**范围已确认**（2026-11-09 产品决策），技术方案仍按实现结果校准。
- 依据：[`docs/research/mygo-butter-desktop.md`](research/mygo-butter-desktop.md)（MyGo 能力与风险核实）、[ADR-0016](adr/0016-agui-runs-outlive-their-request.md)（Detached Run）、[ADR-0002](adr/0002-interrupt-state-derived-from-session-events.md)（Interrupt）、[ADR-0014](adr/0014-a2ui-surfaces-over-agui.md)（A2UI）。
- 关系：本文只定义 **macOS 桌面客户端第一阶段**。#470 / ADR-0017 的 Flutter 移动端决策不变；两者互不取代。

## 1. 决策

Butter Desktop 第一阶段是一个 **macOS-only 的原生客户端**：MyGo（Go，原生 UI，不使用 WebView），账号密码登录，连接现有的远程 Butter server。

据此明确的事实：

- 不使用 WebView、不嵌入 React 页面、不自建 Web 壳；界面全部由 MyGo 原生 UI 绘制。
- 不提供窗口级的"打开 Web Dashboard"作为功能替代。
- 桌面端不直连 MongoDB / Redis / ButterBox，不引入本地 Agent runtime 或单机存储模式。
- 第一阶段之外的能力（Windows/Linux、OAuth、管理面）见 §4，属于后续阶段，而不是"第一阶段做不完的欠账"。

## 2. 第一阶段范围（必须有）

### 2.1 客户端与身份

- 单窗口应用：会话列表 + 聊天区 + 设置（服务器地址、账号与登出、主题、版本信息）。
- 登录使用 `AuthService.Login`（用户名 + 密码）。工作区来自 `LoginResponse.workspaces`，之后所有请求带 `Authorization: Bearer` 与 `X-Workspace-ID`。
- token 只进 macOS Keychain（`mygo.Secrets`），密码不落盘、不进日志、不进 crash 报告。
- 会话过期（默认 7 天，无 refresh）时回到登录页，保留服务器地址与工作区选择，不清空未发送的本地文本（若实现保留）。
- 修改服务器地址必须先停止观察、清理旧 token 与缓存，再连接新地址；旧 token 不得发往新地址。
- 登录失败按状态区分：401 用户名或密码错误、403 用户被禁用、400 缺字段、`FailedPrecondition` 表示服务端 auth store 不可用（不是用户错误）。**不做自动重试**（服务端没有登录限流，自动重试只会放大问题）。

### 2.2 聊天

- 文本与图片输入，遵循服务端限制（JPEG/PNG/GIF/WebP、单张 10 MiB、最多 10 张、合计 20 MiB、请求体 32 MiB）；图片走 inline data，不发 URL source。
- 渲染 Markdown、代码块（含复制）、表格、工具调用参数/结果与状态；单条消息或单个 Surface 渲染失败不能拖垮整个 thread。
- 历史 + UI Snapshot 恢复；历史图片按服务端返回的内联数据恢复。
- **Detached Run**：新 thread 请求 `forwardedProps.butterRun = {detach: true}`；离开页面、关窗、断网只停止观察，不发送 Stop，也不自动重发 POST。
- **重连**：重新进入时读取 thread，发现 `running` 后 attach 同一 `runId`；attach 从 `RUN_STARTED` 全量重放，因此重置本次 run 的临时基线再折叠事件，不把重放文本追加到断线残留文本上。
- **fallback**：attach 返回 204、流中断、`runId` 不符、收到 `CUSTOM butter.fallback`、两次读到的 run 基线不一致时，改读 thread；退避 1s 起、最多 5s，读到不再 `running` 后用历史 + UI Snapshot 替换临时结果。已知被截断的日志不反复 attach。
- **Stop**：`POST …/stop`；202 视为已接受（继续用终态/线程读取校准），204 视为没有可停的运行。服务端拒绝时不显示"已停止"。
- 失败或停止后保留 `lastRun` 提示与可恢复输入；再次发送是新 run，UI 要说明可能重复外部副作用。
- 共享状态：`STATE_SNAPSHOT` 覆盖、RFC 6902 `STATE_DELTA` 应用，服务端为准；不认识的补丁不算成功。
- Interrupt：普通 composer 不带 `resume`（服务端按 FIFO 回答最老的 Interrupt）；显式回答问题或提交表单只携带一个 addressed `resume`；不发送 `cancelled`。

### 2.3 A2UI

- 协商 `forwardedProps.butterA2UI`，处理 `butter.a2ui` envelope：revision、创建/更新/删除、与历史消息关联、UI Snapshot 恢复。
- 只读卡片：Text / Card / Divider / KeyValue / Status（加布局节点）。
- Human Input Form：TextField / ChoicePicker / Button，字段顺序、必填、长度、固定选项；提交 token/revision/Interrupt 绑定；提交中禁止重复点击；400/409/422 分别提示不可提交/已过期/字段错误；成功后锁定旧表单。
- 未知组件或未知 action 安全降级为纯文本提示，绝不执行模型自定义命令、HTML 或任意链接。

### 2.4 质量门槛

- 中文输入法：组合期间 Enter/Esc 不误提交；中文换行、选区、复制正常。
- 键盘可达：发送、切换会话、Stop、表单提交都能只用键盘完成。
- VoiceOver：消息、按钮、输入框、表单字段有可读的名称与状态。
- 深浅主题、高 DPI、系统字体缩放、窗口尺寸变化布局正确。
- 上下文隔离：切换工作区/账号/服务器后，旧请求、旧 Surface、旧草稿不得回填到新上下文。

## 3. 非目标（第一阶段明确不做）

- Windows / Linux 客户端。
- OAuth 登录与原生深链回跳（服务端 `redirect_uri` 固定，需要单独设计 native handoff）。
- 管理面页面：Workspace、Agent、Workflow、Model Provider、MCP、Remote Agent、ButterBox、文件空间、Session 管理、Cron、Automation、Notify Group、Telegram、Linear、Memory、Git 绑定、Token、Daemon/Operations。
- 自动更新通道（第一阶段手工分发 DMG；签名更新密钥与清单后续再接）。
- 本地 Agent runtime、离线队列、多服务器/多账号、通知与后台常驻、Forum、Skills 独立管理、全局管理员功能。

> 说明：这意味着第一阶段的"对等"只对**聊天路径**成立。若后续要求桌面端与 Web Dashboard 功能对等，需要在第二阶段单独评估，且会把工期与 MyGo pre-1.0 的升级风险一起放大。

## 4. 后续阶段候选（按价值排序）

1. OAuth 登录 + native handoff / PKCE / 深链回跳，复用 `internal/repo/oauthhandoff`（#448 已实现服务端一次性兑换）。
2. 管理面：先做只读（概览、Session 列表与详情、Invocation），再做可写（Agent/Workflow）。
3. Windows / Linux 客户端：需要补 D3D11/WARP、GTK3、libsecret 降级与三平台打包验证。
4. 自动更新（签名清单 + delta），与分发渠道一起定。

## 5. 验收门槛

| 项 | 门槛 | 判定方式 |
| --- | --- | --- |
| 构建与分发 | `mygo build` 出 universal `.app` + DMG，Developer ID 签名 + notarize + staple；在**非构建机**的干净 Mac 上双击可开；`spctl -a -vv` 通过；`minimumSystemVersion` 12.0 | 手工验收 + `spctl` 输出记录 |
| 登录 | 正确/错误密码、禁用用户、会话过期、Logout 撤销五条路径；token 只出现在 Keychain；密码不出现在日志与内存快照 | 契约测试 + 手工验收 |
| 协议穿透 | Connect client 调通 `Ping`/`AuthService`/`WorkspaceService`；AG-UI 的发送、detach、attach、stop 端到端通过；重复 `runId` 被 409 拒绝且客户端不重发 | 真实服务端 + 本地 fixtures |
| 重连与恢复 | 关窗 / 断网 / 杀进程后重进：读回历史并 attach 同一 run；`running` 消失后以历史 + UI Snapshot 校准；全程无重复 POST、无重复文本 | 手工 + 故障注入测试 |
| A2UI | revision、过期(409)、字段错误(422)、重复提交、未知组件 5 类用例 | 契约测试 |
| 长回复 | 25 KB 单条回复：滚动 p95 帧时 ≤ 16.7 ms；retained heap 增量 ≤ 2 MB；滚动后跨消息选区与复制不丢 | 性能记录；不达标时按 §6 降级 |
| 中文与可访问性 | 输入法组合不误提交；中文换行/选区/复制正确；VoiceOver 可读可操作；全键盘可达 | 手工验收清单 |
| 崩溃与日志 | 无 crash；日志不含 token、密码、消息正文、图片 payload | 日志抽查 |

## 6. 已知风险与预设处置

| 风险 | 影响 | 处置 |
| --- | --- | --- |
| MyGo #1471：长回复整段建树 + 选区不跨回收 | 长回复与跨消息复制的验收 | 进入 M2 前二选一：等上游修复；或自建 block 级窗口化并接受选区语义退化；都不接受则把该项降级为已知缺陷写入 release notes，并在验收记录里标注 |
| MyGo pre-1.0 breaking changes | 升级成本 | pin commit（首版 `v0.8.3` = `5add3817c74bb7505d0b473bc5fb28abe2201b99`）；每季度评审一次升级，按 `docs/ui/migration.md` 评估 |
| Go 工具链 | 构建 | Butter `go 1.27.1` → MyGo 要求 `go 1.27.3`；随引入依赖一起 bump，并同步 CI |
| 7 天会话无 refresh | 体验 | 过期前提示、过期后回登录页；不做伪造续期 |
| Apple 签名/公证资源 | 发布 | 需 Developer ID 与 notarytool 凭据；未准备好时 M0 出口降级为"本机可运行的 .app"，但 M3 不可降级 |
| 移植量 | 排期 | fold / follower / A2UI 状态层约 1.6k 行 TS 等价逻辑；按 `run-fold.test.ts` 逐用例移植 |

## 7. 里程碑与出口条件

- **M0 骨架 / 构建 / 登录**
  独立 `desktop/` Go module 与独立 CI（参考 `mobile/` 的 `macos-15` 先例）；窗口骨架、设置页、`AuthService.Login`、Keychain 存取、工作区选择、401/403/过期处理。
  *出口*：签名并可运行的 `.app`（公证可延后到 M3）+ 五条登录路径通过。
- **M1 协议穿透 / 只读聊天**
  Connect client 与 AG-UI 客户端传输层；attach 现有 thread，渲染历史、工具调用、图片；不做发送。
  *出口*：能打开 Web 端创建的 thread 并正确显示，无需发送新消息。
- **M2 完整聊天 / A2UI**
  发送（文本 + 图片）、Detached Run、attach/fallback、Stop、Interrupt 回复、A2UI 卡片与表单。
  *出口*：§5 的协议、重连、A2UI、长回复四行全绿（或已按 §6 显式降级）。
- **M3 发布验收**
  中文、VoiceOver、键盘、主题与高 DPI；公证与干净机器安装；release notes（含已知缺陷）。
  *出口*：§5 全表通过或已记录降级；DMG 可交付。

## 8. 仓库与构建边界

- 新增 `desktop/`，**独立 Go module**（不进服务端 `go.mod`），独立 CI job；Go/Web 的日常构建与测试不依赖桌面端。
- 不新增桌面端专用后端 CRUD；桌面端只消费现有 ConnectRPC、REST 上传与 AG-UI/SSE。
- 客户端状态层按职责分层：传输（Connect + SSE）、凭证与上下文、运行折叠、A2UI、界面、平台能力；界面不直接调用 HTTP。

## 9. 依赖的人力与资源

- Apple Developer Program 账号、Developer ID Application 证书；`xcrun notarytool store-credentials` 生成的公证凭据。
- macOS CI runner（GitHub Actions `macos-15` 或自托管）。
- 一台非构建机的干净 Mac 用于发布验收。
- DMG 的分发位置（GitHub Release 或内部渠道）。
