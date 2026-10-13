# ADR-0017: Flutter 作为完整 Butter 客户端，复用服务端协议与运行语义

- Status: Proposed
- Spec 总 Issue：[GitHub #470](https://github.com/orvice/butter/issues/470)
- 关联：[Flutter App Spec](../specs/flutter-app.md)
- Builds on: ADR-0002（Interrupt）、ADR-0014（A2UI）、ADR-0016（Detached Run）

Butter 将在当前仓库加入支持 Android 与 iOS 的 Flutter App，目标是覆盖现有 Dashboard 的普通用户与 Workspace 管理操作，而非仅聊天；全局管理员功能明确不在范围内。提案是在 `mobile/` 独立构建原生界面，复用 Butter 的身份、Workspace、配置、Session 与运行记录：管理走现有 ConnectRPC，聊天走 AG-UI/SSE 与 Butter 的 A2UI/Detached Run 扩展，不新增移动端专用 CRUD 或另一套聊天执行协议。移动系统可能随时暂停网络或杀进程，因此运行属于服务端，App 只观察；回前台读取 thread 并 attach，显式 Stop 才结束 run。

已确认产品边界：Android 与 iOS 均支持；不包含全局管理员功能；Workflow 使用图查看 + 结构化编辑，不要求拖拽画布；暂不支持多个服务器地址，用户在 App 内填写并保存唯一服务端地址，不依赖构建时固定地址。暂不做 Forum 与 Skills 独立管理，Agent 关联已有 Skill 的能力保留。读取 Git Host 与安装已有全局 MCP 预设仍属于 Workspace 使用能力，不包括修改全局资源。

本 ADR 的技术方案仍待评审，Dart SDK、OAuth 回跳及其余待定事项尚未确认。产品范围与验收见 Spec；不在此固定尚未验证的库。

## 为什么这样划分

- 同仓库便于 proto、API 契约与客户端变更协同评审，但 Flutter 构建不能成为 Go/Web 日常开发的前置依赖。
- 完整功能需要大量触屏表单与原生生命周期处理。WebView 封装虽然更快，却不能达到原生界面目标，还会把 Web 的认证存储与页面生命周期问题带入 App。
- 既有 AG-UI thread 可在 Web/App 之间共享。重新设计 mobile chat 或直连 ButterBox 会分裂历史、权限、Stop 与恢复语义，并绕过 Butter 的凭证边界。
- 只用 `StreamAgent` 的同步请求不足以表达已交付的 Detached Run、Run Log、A2UI 和跨 Pod Stop；移动端频繁断连更需要 ADR-0016 的服务端运行模型。

## 约束

1. `proto/agents/v1/` 仍是 RPC schema 单一来源。优先验证生成的 Dart protobuf + Connect binary unary；不把普通 Dart gRPC stub 视为 Connect 客户端。若工具链不兼容，再评审集中式 Connect JSON adapter，禁止页面各自拼接业务 HTTP。
2. 登录使用用户 session token，安全存储在系统凭证存储中；以服务器/用户/Workspace 隔离请求与缓存，角色权限最终由服务端执行。修改唯一服务端地址时清理原凭证与缓存并重新登录，新地址验证不携带旧 token；不因此 Stop 原服务器的 Detached Run。App 不提供全局管理入口，即使登录者是全局 admin；平台 endpoint 等配置继续由 Web/运维管理，App 展示缺失原因。
3. 聊天请求 detach；离开页面/后台/断线只停止观察。发送响应不明确时读取确认，不自动重发；失败不自动 rerun。attach 重放重建本次 run 的临时结果，fallback 用最终 Thread History 与 UI Snapshot 校准。
4. 完整版提供当前 `butter-basic-v1` Result Card 与 Human Input Form 的原生渲染。第三方 A2UI renderer 或有限 catalog 自实现必须通过同一套契约；只有文字 fallback 不算功能对等。
5. 外部授权使用系统浏览器。当前登录 Provider 使用配置的固定 callback，原生回跳方案需验证；如需后端一次性兑换/回跳能力，独立设计安全与 Web 兼容性，不能直接把长期 token 放在 deep link 中。
6. 客户端不得直接执行 Pi/Cursor/Daemon，也不得建立离线副作用队列。移动生命周期恢复不要求后台常驻服务、永久 SSE 或后台 Agent 执行。

## Considered Options

- **WebView 包装 Dashboard**：可快速覆盖页面，但不满足原生交互目标；不作为正式完整客户端方案。
- **独立仓库**：可隔离发布权限，但 schema/后端/客户端需要额外跨仓库版本协调；本次在当前仓库交付，并保持构建隔离。
- **移动专用 BFF/REST CRUD**：会重复授权、校验、optional 字段与资源生命周期规则；没有真实能力缺口前不引入。
- **只做聊天 App**：规模较小，但与用户要求的完整功能不符；可作为交付阶段，不能替代最终范围。
- **只用同步 RPC 聊天或直连 ButterBox**：无法复用完整的 thread binding、A2UI、服务端运行与跨 Pod 恢复；不采用。
- **手机后台维持运行**：OS 不保证连接和进程存活；采用服务端 Detached Run 与前台重建，而非依赖后台常驻。

## Consequences

- Web 与 App 共享持久化业务状态，但各自维护临时 UI；需要跨端契约测试，尤其是运行重放、表单和上下文隔离。
- 两套界面均需维护。功能对等靠可审计的操作矩阵与测试，不靠复制 React 代码或嵌入 Web。
- Dart Connect、AG-UI、A2UI 与原生授权有前置验证成本；验证不通过时先修正 adapter/回跳方案，不削弱验收要求。
- 需要独立 Flutter CI、iOS macOS 构建环境及双平台签名/发布流程；Go 与 Web 构建保持原样。
- 新 API/字段上线时需评估两个客户端的兼容性；未知字段/enum 不得因旧版移动端保存而被静默清空。
- 手机离线时只保留明确的本地 UI 状态，不宣称运行已停止、配置已保存或副作用已安全重试。
