# Butter Flutter App Spec

- 状态：Draft，产品范围与 Ticket 拆分已确认；其它待定事项及技术方案仍待评审，本文不代表实现已完成。
- Spec 总 Issue：[GitHub #470](https://github.com/orvice/butter/issues/470)；实施子 Issue：#447–#469（23 张）。
- 基线：当前仓库 `47f9c77` 的 Web 路由、API 实现与 proto；后续开发需重新核对基线变化。
- 关联：[ADR-0017](../adr/0017-flutter-app-as-a-full-butter-client.md)、[API](../api.md)、[功能总览](../app.md)、[领域术语](../../CONTEXT.md)。

## 1. 目标与“完整功能”的定义

在当前仓库新增 Flutter 客户端，作为与 Web Dashboard 并列的正式客户端，而不是只提供聊天的 companion app。复用同一个 Butter 服务端、身份、Workspace、配置与运行记录，不建立第二套业务存储或执行系统。

**功能对等**指：当前 Dashboard 已交付且服务端支持的普通用户与 Workspace 管理操作，都能在 App 完成；**不包含全局管理员功能**。可以为触屏重新设计交互，不要求像素级复制 Web，也不能用“打开 Web 页面”代替核心管理界面。第三方授权和第三方控制台配置仍由系统浏览器完成。

完整范围允许分阶段交付，但全部必选项通过验收前不得宣称“完整版”。后端只有 schema、尚未实现的功能，以及已退役 API，不因移动端而重新启用。

### 平台范围

- **已确认：Android 与 iOS 均为正式目标。**
- 手机采用自适应布局；平板是否作为正式验收目标待确认。
- Flutter Web、macOS、Windows、Linux 暂不纳入；现有 `front/` 继续承担 Web。
- Android application ID、iOS bundle ID、最低 OS 版本、上架方式与签名责任人需在脚手架前确认。

### 不在本次范围

- Forum：暂不实现独立的 Forum Thread/Post 浏览、创建、回复或管理，不改变后端与 Web 已有 Forum 能力。
- Skills 独立管理：暂不实现 Skill CRUD、SKILL.md 编辑或 Skill Resource 管理；Agent 配置中关联已有 Skill 的能力保留，不改变后端 Skills 功能。
- 全局管理员功能：用户创建/停用/重置他人密码、Git Host CRUD、全局 MCP 预设 CRUD、Telegram/Linear 公网 endpoint 配置及全局诊断管理等；继续由 Web/运维配置。个人资料、本人密码修改与 Workspace 成员管理仍在范围内。
- 在手机上启动 Go 服务、ADK、Pi、Cursor 或 Daemon worker。
- 直连 ButterBox、mem0 或 Git Provider 执行业务；仍通过 Butter 的权限和凭证边界。
- 离线运行 Agent、离线写入队列、自动重跑失败操作。
- 新增语音、视频、任意文件聊天输入、原生推送、桌面客户端。
- 恢复旧 `web-chat`、generic Channel 写入或已移除的异步 Invocation API。

## 2. 角色、上下文与导航

以服务端权限为准，不能假设所有 workspace 成员都有配置写权限，也不能把全局 admin 与 workspace owner/admin 混为一谈。客户端控制入口可见性只是体验，服务端必须再次授权。

建议一级入口：概览、聊天、资源管理、运行与集成、设置；手机采用底部导航与分组列表，大屏可采用侧栏。管理入口覆盖范围内的 Workspace 能力，不提供全局管理入口；即使全局 admin 登录 App，也只提供范围内操作，不改变服务端既有权限。

客户端上下文至少区分：服务端地址、当前用户、Workspace。聊天再增加 Agent ID 与 threadId。Agent 的引用统一使用不可变 `agent_id`，显示名只能用于展示。

- 同一登录用户在 Web 与 App 可访问同一条绑定的 AG-UI thread，不迁移历史、不创建独立 `mobile-chat` 会话体系。
- **已确认：用户在 App 内填写唯一服务端地址，暂不支持多个服务器地址。** 首次使用先填写并验证地址，再进入登录；保存该地址供下次启动使用，不提供服务器列表、多服务器配置或快捷切换，不依赖构建时固定地址。
- 设置页允许修改唯一服务端地址；确认修改时必须停止旧上下文的观察、清理原登录凭证与缓存并重新登录，不能把旧 token 发送到新地址。验证新地址不携带旧凭证；修改地址不 Stop 原服务器上的 Detached Run。重新登录或切换 Workspace 时取消旧上下文的请求与观察，清除内存中的消息、Surface、共享状态和权限缓存；旧响应不得回填到新上下文。不维护多服务器/多账户会话池。
- 导航离开聊天只 detach 观察，不隐式 Stop 服务端 run。

## 3. 功能对等矩阵

下表为必选范围，不含全局管理员操作；每个模块实现时应建立具体的“Web 操作 → RPC/REST → App 页面 → 测试”清单。不得以一个 JSON 文本框代替整个复杂编辑器。

| 模块 | App 必须提供的操作与语义 | 主要接口/依据 |
| --- | --- | --- |
| 服务端与身份 | App 内填写、保存及修改唯一服务端地址，连接验证；密码登录、已配置 Provider 的 OAuth 登录、登出；会话失效后重新登录；不提供多服务器切换 | AuthService、`/ping` |
| 个人资料与设置 | 资料、头像上传、修改密码、主题、显示偏好；当前服务器与账户可辨识 | AuthService、uploads REST、Web profile/settings |
| Workspace | 列表、选择、创建、编辑、删除、成员管理；按服务端实际角色限制操作 | WorkspaceService |
| 概览 | 数量、健康、活动流、Cron 时序；不得把全局汇总错误标为当前 Workspace 数据 | DashboardService |
| 聊天与 thread | Agent 选择、文本/图片、历史、标题、重命名、删除、Detached Run、重连、Stop、失败输入恢复 | AG-UI、SessionService；见 §5 |
| Agent | 七种类型的创建/编辑；子 Agent 引用、模型、指令、MCP、Skills、文件挂载、Memory、Card Policy、ContextGuard、API 暴露设置等适用字段；状态与调用记录 | AgentService、agent.proto |
| Agent 生命周期与内容 | 与 Web 当前入口对等的 tombstone/恢复、operation 查询与重试、Git 内容编辑及冲突提示；删除与 purge 不混同 | AgentService、WorkspaceRepoBindingService；ADR-0006/0007 |
| Workflow 编辑 | 图查看 + 结构化节点/边编辑（已确认，无需拖拽画布）；Router/default route、Join、Parallel Worker、重试与超时、Human Input Form 字段和选项；验证错误 | WorkflowConfig；ADR-0001/0002/0014 |
| Model Provider | CRUD、模型与别名、连接参数、Model Context Capacity；遵循当前后端凭证处理，不自行改变 API | ModelProviderService |
| MCP | Workspace server CRUD、transport、工具白名单、探活、工具列表、OAuth 授权/断开、安装全局预设 | MCPServerService、GlobalMCPServerService |
| Remote Agent | A2A、OpenCode HTTP、Daemon 配置与状态，引用现有 runtime | RemoteAgentService |
| ButterBox | CRUD、测试连接、凭证保持/替换/清除（依具体 API）、Pi/Cursor 模型选择；引用阻止删除提示 | ButterBoxService；ADR-0011/0012 |
| 文件空间与文件 | 与 Web 对等的空间/文件 CRUD、查看、编辑、搜索、版本操作（按现有接口支持）；挂载配置 | AgentFileService、Web agent-files |
| Session 管理 | 过滤、分页、详情/事件/trace、重命名、删除、回复可授权会话；非 AG-UI session 不强行变成聊天 thread | SessionService、session access policy |
| Cron | CRUD、启停、立即运行、schedule/时区/策略/投递编辑、执行记录与详情、WAITING_INPUT 查看和通过授权 session 回复 | CronJobService、SessionService；ADR-0003 |
| Automation | CRUD、conditions、ordered steps、policy、手动运行、run 与 step-run 详情；未接入的 trigger 不显示为可用 | AutomationService |
| Notify Group | CRUD、所有已支持 target 类型、Telegram Destination 引用；拒绝 legacy raw Telegram 字段 | NotifyGroupService |
| Telegram | Channel 创建、同 Bot token 轮换、收发状态与预检、Webhook/Long Polling；Destination 精确地址、路由与交互策略、测试发送；processing 列表/详情与仅补发 | Telegram 四组服务；ADR-0008/0009 |
| Linear | App CRUD、secret 更新、准入名单、运行时限、Installation 安装/查看/移除、入站开关、processing 列表/详情与仅补发 | Linear 三组服务；ADR-0015 |
| Memory | Workspace 连接配置、测试、警告；Workspace/Agent Memory 列表、搜索、授权删除、截断提示；principal 按权限展示 | WorkspaceMemoryConfigService、WorkspaceMemoryService |
| Git Repository Binding | Git Host 选择、绑定配置、PAT 管理、验证、导入/导出 onboarding、状态/同步/发布、当前 Web 内容操作、安全 detach 与恢复选项 | WorkspaceRepoBindingService；ADR-0005/0007 |
| Token | API token 列表、创建、一次性 secret 展示、撤销；worker token 与普通 API token 不混用 | APITokenService、DaemonService |
| Daemon 与 Operations | runtime CRUD、token 签发、在线 daemon/任务、任务取消与诊断；Invocation 查看和授权 Stop/取消；实际取消能力由入口决定 | DaemonService、AgentService、Web operations |
| Integrations 与 Manage | 提供 Web 当前连接说明、入口聚合、可复制的 URL/配置；不把导航入口误当独立业务资源 | Web integrations/manage |

读取平台已配置的 Git Host、读取/安装已有全局 MCP 预设仍在范围内；修改这些全局资源不在范围内。Telegram/Linear 配置依赖全局公网 endpoint 时，只展示状态、缺失原因与联系平台管理员的提示，不在 App 提供配置入口，也不以 root token 绕过。

### Skills 范围（已确认）

暂不做 Skills 独立管理，即使后端已有完整 CRUD 与 Resource RPC，也不纳入本次 App 实现或验收。Agent 配置仍可关联已有 Skill；必要的只读 Skill 列表用于选择，不提供独立管理页、正文编辑或资源上传/删除入口。

## 4. 表单与管理交互

- 常见 CRUD 必须有列表、详情、明确的创建/编辑入口、服务端错误、加载/空状态和提交中状态；接口支持分页或过滤时必须可用。
- **已确认：Workflow 提供图查看 + 结构化节点/边编辑器**，小屏可采用分步表单；触屏拖拽画布不是验收要求。
- 七种 Agent 类型按服务端规则显示适用字段。Pi/Cursor 禁止配置 Butter instruction/MCP/Skills/Memory/Card Policy 等叶子限制，不因为 Flutter 表单通用化而发送非法字段。
- 更新保持未编辑字段；不能把未知 enum 或不识别字段静默覆盖为默认值。需建立安全往返测试，无法无损编辑时阻止保存并提示版本不兼容。
- 有 revision/CAS 的操作使用接口提供的版本，冲突时让用户刷新比较，不自动强制覆盖；没有并发控制的接口不宣称已支持 CAS。
- 区分密钥“保持不变”“替换”“清空”，保留 optional 字段的 absent 与 empty 差异。写后不尝试读取服务端 write-only secret。
- 删除、purge、token 撤销、运行、补发、入站开启、repo onboarding/detach 等有副作用操作需明确确认对象与风险；不可因网络超时自动执行第二次。
- 编辑页离开提示未保存变更；敏感草稿仅在内存，非敏感草稿持久化策略待确认。

## 5. 聊天、A2UI 与恢复契约

### 5.1 发送与呈现

聊天使用 `POST /api/agui/:agent_id`，普通管理使用 ConnectRPC。新 thread 按现有 AG-UI 路径建立 UI Binding，并请求 `forwardedProps.butterRun = {detach: true}`；不能先创建一个无 binding 的 Session 再假设它可 detach。

- 展示 Markdown、代码块/复制、表格、Agent 头像/名称、工具参数/结果与状态；一条消息或 Surface 渲染失败不得拖垮整个 thread。
- 支持系统图片选择、预览、移除和历史图片恢复；相机入口可后续确认。遵循 JPEG/PNG/GIF/WebP、单张 10 MiB、最多 10 张、合计 20 MiB、请求体 32 MiB 限制。
- 图片发送为 inline data，不发送 URL source；发送只携带尾部工具结果或本轮用户消息，不重发全部历史图片。支持纯图片消息；回答 Interrupt 仍遵循文字要求。
- 普通 composer 不带 `resume`，按服务端 FIFO 回答最老 Interrupt；显式问题输入与表单提交只携带一个 addressed resolved `resume`；不发送 `cancelled`。
- 暂不声明任何原生 frontend tools；需要新增时必须另行定义白名单、权限与执行规则。能够展示服务端工具不等于 App 会执行工具。

### 5.2 客户端状态模型

客户端区分线程读取状态与运行观察状态：加载、空闲、发送待确认、实时观察、恢复读取、attach 重放、fallback 等待、等待 Human Input、失败/停止提示、无权限/不存在。枚举命名可在实现时调整，但以下语义不可合并：

1. 每次新操作生成新的 runId；发送前保留本次 threadId/runId 与上下文的恢复标识，不保留凭证或完整图片 payload。
2. 接受状态不明确时，先读取 thread 和已知运行标识；禁止 POST 自动重试，禁止换一个 runId“重连”。若仍无法判定，只提示用户，不伪装为确定失败。
3. 读取历史与 UI Snapshot，按服务端 cut 展示；发现 `running` 后 attach 同一个 run。
4. attach 从 `RUN_STARTED` 全量重放，因此重置这次 run 的临时回复/状态基线，再折叠事件，不把重放文本追加到断线前的残留文本。
5. 只接收与期待 runId、当前上下文及观察 generation 匹配的流；异步旧结果不污染新页面。
6. attach 返回 204、流中断、runId 不符、`butter.fallback`，或两次 thread read 出现不一致的 run 基线时，重新读线程；不能拼接两个不同 run 的基线。
7. 无法完整观察当前 run 时，按现有 Web 契约进入 fallback：1 秒起、退避至最多 5 秒，读到不再 `running` 后，用历史与 UI Snapshot 替换临时结果。已知被截断的日志不反复 attach。
8. fallback 等待时禁用发送，仍允许编辑本地输入与 Stop；临时网络/502/503/504 可退避，409 与其他确定错误不无限重试。401 要重新登录，403 不解释成掉线。
9. `STATE_SNAPSHOT` 与 RFC 6902 `STATE_DELTA` 维护客户端镜像，服务端为权威；未识别的状态操作不能静默当成功处理。
10. `RUN_FINISHED` interrupt outcome、pending tool call、成功、`RUN_ERROR stopped` 和失败分别呈现；断开本地观察不等于失败或停止。

### 5.3 Stop、删除与移动端生命周期

- Stop 调用线程 `/stop`；202 表示接受、204 表示没有可停止运行，不表示已观测到终态。服务端拒绝时不显示“已停止”。
- Stop 接受后可停止本地流，但通过终态/线程读取校准最终结果，不继续接入同一旧流；fallback 等待中也如此。
- 删除通过 SessionService 的服务端 stop-and-delete 契约完成；失败不得本地永久移除历史。
- 锁屏、切到后台、OS 杀进程、网络切换均不能发送 Stop，也不承诺手机后台持续持有 SSE。回到前台后重新授权/读取/attach。
- 保留失败/停止的 `lastRun` 提示与 Restore input；明确再次发送会创建新 run，可能重复外部副作用。历史无法关联原图时只恢复可确认的文字；Interrupt/form/tool-result run 不伪造可恢复输入。

### 5.4 A2UI 完整支持

完整版必须支持当前 A2UI v0.9.1 与 `butter-basic-v1`，不能仅靠纯文本 fallback 宣称对等。

- 验证通过后才声明 `forwardedProps.butterA2UI`；实现 `butter.a2ui` envelope、revision、创建/更新/删除、历史 message 关联和 UI Snapshot 恢复。
- 组件覆盖当前 catalog：Text、Column、Row、Card、Divider、KeyValue、Status，以及服务端表单的 TextField、ChoicePicker、Button；渲染取值/binding 必须与服务端 catalog 相符。
- 模型 Result Card 保持只读；只接受服务端 Human Input Form 规定的提交 action，不执行模型自定义命令、HTML 或任意链接。
- 表单保留字段顺序、必填、长度、固定选项、提交 token/revision/Interrupt 绑定。提交中禁止双击；400/409/422 分别提示不可提交/过期/字段错误，失败保留草稿；成功锁定旧表单。
- 使用 revision 防止重复回放污染；切换上下文清空 Surface 与表单草稿；未识别 Surface 显示 fallback，不中断对话，也不开放未知 action。
- Flutter SDK 或手写有限 catalog renderer 都必须通过相同契约测试；`agentivity_ag_ui` 等库目前仅为候选，尚未验证兼容性。

## 6. 协议、认证与安全

### 6.1 传输边界

- `proto/agents/v1/` 是 RPC schema 单一来源。优先验证 Dart protobuf + Connect 客户端的 binary unary RPC；不把 gRPC stub 当成 Connect 支持的证据。
- 若 Dart Connect 工具链不满足现有服务，才评审集中式 Connect JSON adapter。JSON 路径须遵循 snake_case 响应、optional/oneof、枚举、int64、bytes 与 Timestamp，不在 feature 中散落手写 HTTP。
- SSE 必须支持带 Bearer/Workspace headers 的 POST 与 GET、任意 UTF-8 分片、LF/CRLF、注释 heartbeat、多行 data 和有界缓冲；重连采用 Butter thread/Run Log 契约，不假设 Last-Event-ID 增量续传。
- `/api/uploads/*` 继续 REST multipart；文件资源的其他 bytes/上传路径按对应 API，不统一误改成头像上传。
- Header 的 Workspace 要求按具体服务执行；Auth/Workspace/Dashboard 等例外与 Session access policy 不能被通用 interceptor 简化掉。
- 不新增 mobile-only CRUD 后端；原生登录回跳等真实协议缺口可作为独立后端变更，需测试并保持 Web 兼容。

### 6.2 身份与凭证

- 以用户登录 session 为正式入口，Workspace 管理功能不使用 root/API token 冒充用户；API token 管理不等于 API token 登录。
- Bearer token 存系统安全存储；服务端地址与非敏感偏好可普通存储。平台备份不得泄漏 token，日志/crash report 不记录密码、密钥、消息 payload 或授权 code。
- 401/Unauthenticated 清理该账户凭证和敏感缓存并引导登录；不虚构当前后端没有的 refresh token 机制。Logout 即使离线也必须清理本地，服务端失效结果如实提示。
- 默认 HTTPS；HTTP 仅在显式开发模式放行，不能全局跳过证书校验。地址不得带 userinfo、query 或 fragment；部署 path prefix 是否支持需验证。
- 外部授权走系统浏览器，避免嵌入式 WebView 和长期 token 的 URL 回跳；一次性凭证如需引入，必须绑定服务器、provider、state 与允许的回调地址。

### 6.3 已发现的登录 OAuth 缺口

`AuthService.BeginOAuthFlow` 存储 `redirect_uri`，但当前 Provider 的 authorize/code exchange 使用配置的固定 RedirectURL；不能假设其支持任意 App deep link。需验证当前 Web callback 与原生回跳可如何安全衔接，明确 code 的兑换方、state 校验、防重放、取消、超时和允许地址。

MCP OAuth 与 Linear 安装是不同授权流程，各自复用服务端回调和凭证管理；不得拿用户登录 OAuth 的设计直接套用。浏览器返回 App 后重新读取安装/授权状态。未解决登录回跳前不得标记 OAuth 验收完成。

## 7. 仓库与模块边界（提案）

```text
mobile/
  lib/
    app/                # 启动、导航、上下文与生命周期
    core/               # 传输、安全存储、错误、公共 UI
    gen/                # 生成 protobuf / service 绑定，禁止手改
    features/           # 按领域组织页面、状态与用例
  test/                 # 单元、widget、契约 fixtures
  integration_test/     # 原生关键路径
  android/
  ios/
```

页面只消费领域层用例；Connect/SSE transport、AG-UI run fold、A2UI surface store、上下文隔离分别具备可独立测试的接口。状态管理、路由、HTTP、Markdown 库先做短验证再锁版本，不在本 Spec 凭空宣称某 SDK 支持协议。

`mobile/` 独立依赖/构建/CI；不改变 Go/Web 的日常构建前提。生成配置需可重复执行并检测漂移，不手改 `pkg/proto/` 或 Web 生成物。

## 8. 验收与发布门槛

### 功能对等门槛

- §3 每个必选操作都有 App 实现与测试记录，按 member、workspace owner/admin 验证可见性及服务端拒绝；全局 admin 登录时也不得出现范围外的全局管理入口。
- 管理端不仅能查看：复杂 Workflow、Cron/Automation、Git onboarding/detach、凭证轮换和 Workspace 管理均可在手机上完成。
- 全局资源只允许范围内的读取/安装操作；缺少平台配置时如实提示，不要求在 App 修改全局配置才能完成验收。
- UI 未提供的能力、后端未实现的接口必须明确标为缺口，不能用入口截图替代验收。

### 契约测试矩阵

| 场景 | 必须证明 |
| --- | --- |
| Connect binary/JSON、分页与字段往返 | 服务端真实可调用；optional absent/empty、enum、int64、bytes、安全编辑不丢字段 |
| 两 Workspace 使用同一 Agent ID、退出后重新登录、修改唯一服务端地址 | 缓存、迟到响应、Surface、token、draft 不串上下文；旧 token 不发往新地址 |
| Web 发起/App attach；App 发起/Web attach | 同一 thread，无重复运行与文本，绑定/权限一致 |
| POST 响应丢失、409 busy/run_exists | 读取恢复，不自动新建 run，不重复执行工具 |
| 网络分片/断线、锁屏、后台、杀进程 | run 继续，前台恢复；手机不要求后台 SSE 永久在线 |
| 日志截断/过期/204、读与 run 开始竞态 | 正确 fallback，并用最终历史与 UI 替换临时结果 |
| Stop 成功/拒绝/超时/竞态、删除运行中 thread | 不误报停止，不本地删掉服务端仍存在的 thread |
| 多 Interrupt、表单过期/重复/字段错误 | 精确恢复一个，其他保持 pending，无 cancelled，草稿不丢 |
| A2UI 创建/更新/删除/replay/unknown | 正确 revision、message 关联、无重复 Surface、安全 fallback |
| Token 过期、成员移除、secret 修改冲突 | 重新登录/权限提示，不无限重试，不覆盖凭证 |
| Git 内容冲突、引用阻止删除、delivery 补发 | 展示真实失败原因；补发绝不 rerun Agent |
| 登录/MCP/Linear 浏览器授权 | 取消/超时/回跳/状态恢复均可用，长期凭证不进入 URL |

- 单元测试覆盖 SSE parser、run fold、state patch、Surface revision、错误映射与上下文 generation；widget 测试覆盖复杂表单和权限。
- 使用本地测试服务器/fixtures 跑确定性契约；另用真实 Butter 多 Pod + Redis 环境验证跨 Pod attach/Stop，不仅用 mock 宣称兼容。
- 至少一次 Android 与 iOS 真机验证图片、安全存储、生命周期、授权返回及可访问性；模拟器不替代所有真机检查。
- CI：format、analyze、unit/widget/contract、代码生成漂移、Android build；iOS 在 macOS runner 验证。发布签名凭证不得进仓库。
- 完整版支持中文/英文提案待确认；必须支持系统字体缩放、屏幕阅读器标签、键盘遮挡处理、深浅主题和大表单滚动。

## 9. 实施分阶段（不缩减最终范围）

1. **协议与缺口验证**：Dart Connect 真实调用、SSE 与 Run Log 恢复、A2UI catalog 及原生 OAuth；形成验证结果与后端依赖清单。
2. **基础与资源管理骨架**：脚手架、CI、安全登录、上下文、Workspace、导航、统一列表/表单/错误。
3. **完整聊天**：历史、图片、Detached Run、跨端 attach/Stop、A2UI、生命周期与失败恢复。
4. **Agent 与配置**：所有类型/Workflow、模型、MCP/Remote、ButterBox、文件、Memory、Git 内容与生命周期。
5. **集成与运行**：Telegram、Linear、Cron、Automation、通知、Session、Invocation、Daemon、Dashboard。
6. **完整管理与发布**：Token、个人设置与 Workspace 成员管理、对等矩阵审计、双平台真机、签名分发；不包含全局用户与平台设置管理。

每阶段验收仅代表该阶段完成；协议验证失败先修正设计，不能边做页面边默认协议可用。

## 10. 产品决策

### 已确认

1. Android 与 iOS 均支持。
2. 完整功能范围不包含全局管理员功能；保留普通用户与 Workspace 管理能力。
3. Workflow 采用图查看 + 结构化编辑，不要求触屏拖拽画布。
4. 暂不支持多个服务器地址；用户在 App 内填写、保存及修改唯一服务端地址。
5. 暂不做 Skills 独立管理；保留 Agent 关联已有 Skill。
6. 暂不做 Forum，不纳入当前 App 功能矩阵与验收。

### 待确认

1. 平板是否纳入正式验收目标？
2. 语言、App 名称/标识、最低 OS、内部分发还是商店、签名与 macOS CI 资源？
3. 非敏感离线草稿是否需要持久化？推送/生物识别解锁暂不在范围是否接受？

待以上产品决策与 §6 技术缺口评审后，把 Spec 与 ADR 状态推进；现在只增加文档，不创建 Flutter 工程或修改后端接口。
