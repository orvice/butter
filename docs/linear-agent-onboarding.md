# Linear Agent 接入、验证与排错

本文面向 workspace owner / admin 与平台管理员，说明如何把一个已有 Butter Agent 接进
Linear：注册 OAuth App、安装、设置 allowlist、启用 inbound、做真实验证，以及出问题时
从哪里看证据。架构与实现细节见 [ADR-0015](adr/0015-linear-agent-sessions-as-an-entry-point.md)、
[docs/architecture.md](architecture.md#linear-执行流adr-0015) 和
[docs/api.md](api.md#linear-services)。

> **证据状态（重要）**：截止本文更新，本仓库只有代码、单元/HTTP 测试和脱敏 fixture
> 证明的协议结论（见
> [docs/research/linear-protocol-validation.md](research/linear-protocol-validation.md)），
> 还没有在真实 Linear 组织里执行过的 live validation。因此页面上的“已安装”“已启用”
> 都只是配置事实，不等于“已经在真实环境跑通”。未验证项集中列在
> [§8 剩余限制](#8-剩余限制)。

## 1. 页面上的五种事实，各自证明什么

Linear 接入没有 `connected` / `healthy` 这种总状态。Dashboard 和引导页只展示可以逐项
核对的事实。把它们混为一谈是排错时最常见的错误。

| 事实 | 含义 | 能证明 | 不能证明 |
|---|---|---|---|
| 已注册（Linear App） | 一个 workspace 资源：不可变 client ID、路由到的 Agent、只写 secret 已保存 | Butter 侧已有可用的 OAuth client 配置 | Linear 侧 OAuth App 的 callback / webhook URL 已填对；Linear 能访问 Butter |
| 已安装（Linear Installation） | 该 App 在某个 Linear 组织完成 `actor=app` 授权，token 已加密保存 | 至少一次 OAuth 安装成功；记录了 organization / app user / scopes | inbound 已开；webhook 可达；Linear 侧事件订阅正确 |
| 已启用 inbound | 服务端的 desired state：接受新的 Linear 投递 | 新 deliveries 会被处理 | 历史上跑通过；当前凭证有效；正在运行的 turn 会被停止（不会） |
| 最近已跑通 | 读取窗口内、属于当前 App + Installation + 当前 Agent 的一条成功记录 | 窗口内确实有一次被接受、Agent 执行、回复已投递的历史 | 现在健康；窗口外没出过问题；其他 Installation 也正常 |
| 当前故障 | 最近记录中的失败 / 拒绝 / 投递失败 / 不确定 / 凭证失效 | 具体这条投递发生了什么，下一步能做什么 | 是不是全局故障；未观察到事件时不等于没有事件 |

补充：

- **凭证状态**是 Installation 的字段，不是健康检查。`VALID` 表示最近一次已知结果可用；
  `NEEDS_REINSTALL` 表示 Linear 拒绝或撤销了刷新；`UNSPECIFIED`（以及未来未知值）显示
  “Credential state not recorded”，表示还没有任何 token 使用/刷新结论，**绝不显示为可用**。
- **App 凭证状态** `INCOMPLETE` / `COMPLETE` 只表示两个只写 secret 是否都已保存。
- **“最近已跑通”的判定边界**：记录必须是当前 App 和 Installation、当前绑定的 Agent、
  `created` 或非 Stop 的 `prompted`、`SUCCEEDED`、带 invocation ID，并且回复已投递。
  Stop 记录没有 invocation，不计入成功证据。旧 Agent、改配置前、被移除 Installation
  的记录会单独标注。

## 2. 前置条件

1. **Public base URL**：全局管理员在 Manage → Workspace → Administration →
   Public endpoints 的 Linear 卡片填写。必须是 `https` 且不带路径；仅本地开发允许
   `http://localhost`。每个 App 的 callback / webhook URL 都由它和不可变 App ID 派生，
   从不落库；没配置时安装会被拒绝（failed precondition）。
2. **Durable Redis**：与 Webhook Telegram Channel 相同的要求——持久化、无驱逐。Linear
   的 accepting dedupe、事件 Stream（`butter:linear:events`）、follow-up 列表、会话租约、
   Stop 标记和 OAuth install state 都依赖它。Redis 丢数据会丢掉已接受但未处理的投递和
   排队中的 follow-up；崩溃留下的 `PROCESSING` 记录会被下一个持有者标为
   `FAILED_UNCERTAIN`，不会重跑。
3. **Agent**：必须是 ACTIVE。一个 Linear App 固定路由到一个 Agent；App user 就是该
   Agent 在 Linear 里的身份，换 Agent 等于换身份，也会开新会话历史。
4. **权限**：workspace member 只读；owner / admin 管理 App、安装、secret、启用与补发；
   global admin 旁路。权限由服务端强制，不是按钮隐藏。
5. **Linear 管理员**：需要能在目标组织创建 OAuth App 并安装它。

## 3. 管理员接入步骤

### 步骤 0：平台配置

全局管理员设置 public base URL（见上）。Dashboard 会在 App 页面展示派生出的
callback URL 与 webhook URL，并显示“Set the Linear public base URL first”直到配置完成。

### 步骤 1：在 Linear 创建 OAuth App

在 Linear 的 Settings → API → OAuth applications 创建应用：

- 打开 **Agent session events**（App 必须能收到 Agent Session 事件）。
- 安装时 Butter 会请求 `read,write,app:assignable,app:mentionable`（
  `linearapi.InstallScopes`）；`app:assignable` / `app:mentionable` 才让 App user 可被
  issue 委派和在评论里 @ 到。
- 记录 client ID 与 client secret。webhook signing secret 用 Linear 提供的值，Butter
  只用它验签。

### 步骤 2：在 Butter 注册 App

两个入口操作的是**同一个资源**：

- Agent 编辑页 → **Connect Linear**（`/agents/<agent_id>/linear`）：Agent 由入口固定，
  表单里的 Agent 不可改，提交时总是带入口的 `agent_id`。
- Manage → **Linear Apps** → Register Linear App：可以自行选择 Agent。

填写 display name、client ID、client secret、webhook signing secret、allowlist、
max run time。secret 只写、不回填，不会进入 URL、localStorage 或日志。

### 步骤 3：把 URL 贴回 Linear

把 App 页面展示的 callback URL 与 webhook URL 填进 Linear OAuth App。
没有这一步，安装与投递都会失败，但 Butter 侧仍会显示“secret 已设置”。

### 步骤 4：安装到 Linear 组织

点 **Install to Linear** 打开 Linear 授权页，选择目标组织并批准。Linear 会重定向到公共
callback；成功后回到 App 页面并列出 Installation。一个 App 可以装进多个组织，每个组织
是一条独立 Installation。

OAuth 失败时页面按 `reason` 给出具体含义，且不会显示安装成功：

| reason | 含义 | 处理 |
|---|---|---|
| `denied` | 在 Linear 取消/拒绝授权 | 重新 install |
| `state_invalid` | state 过期（10 分钟）或已被使用 | 从 App 页面重新开始 |
| `app_missing` | state 指向的 App 已不存在 | 重新注册或选择正确 App |
| `exchange_failed` | Linear 拒绝 code 换 token | 核对 client ID / secret 与 callback URL |
| `identity_failed` | 取不到组织/app user 身份 | 检查 scopes 与 Linear 应用权限 |
| `storage_failed` | token 存储失败 | 检查数据库与主密钥，然后重试 |

### 步骤 5：allowlist、运行时长与启用

- **allowlist**：留空表示放行**已安装组织**的所有成员；非空时只允许列出的 Linear user ID。
  比较会规范化大小写和首尾空白；配置了 allowlist 但事件里没有 prompting user 时
  **fail closed**（不运行）。
- **风险**：被准入的用户可以驱动这个 Agent。对 Pi / Cursor Agent 来说，这等于能在其
  ButterBox 上执行命令，包括该 box 用户能访问的一切。空 allowlist 时页面会明确写出
  这一点。
- **max run time**：默认 1800 秒，显式 `0` 表示不限，上限 86400 秒。超时会以 `error`
  activity 结束。
- **启用 inbound**：要求两个 secret 都已保存，且 Redis 是持久化队列；否则保存会被拒绝。
  关闭 inbound 只停止接收**新**投递，不承诺中断已经在跑的 turn。

### 步骤 6：验证

引导页第 4 步与 App 页面都会显示同一份诊断。可见期间每 5 秒只读刷新，离开即停。
先在 Linear 里委派或 @ 一次，再回来看 Installation 区块的“最近已跑通”。

## 4. 真实验证与测量条件

这一节定义“跑通”的测量方式。没有真实证据前，不要把“配置完成”写成“已验证”。

### 4.1 两个时限

| 时限 | 从 | 到 | 怎么测 |
|---|---|---|---|
| Webhook 响应 ≤ 5 秒 | Linear 发起投递（payload `webhookTimestamp`） | Butter 返回 HTTP 200/其它状态 | 边缘代理 access log 的请求耗时，或对同一签名 body 的受控重放（重放会命中去重并返回 200，不产生新 invocation）；注意 5 秒是接收路径（验签 + 入队）的预算，不含 Agent 运行 |
| 确认活动 ≤ 10 秒 | 同上 `webhookTimestamp` | Linear 会话里出现确认 `thought`（“Picked up …”） | Linear 侧 activity 的时间戳；或 Butter 的 debug 日志与 Linear API 观察到的 activity 创建时间 |

注意：

- Processing record 的 `created_at` 是 worker 认领记录的时间，不是投递到达时间；
  `webhookTimestamp` 只用于 ±60 秒的 replay 窗口，不落库。Butter 不持久化单次投递
  延迟，因此这两个指标必须从边缘日志和 Linear 侧时间取得。
- 10 秒预算覆盖：验签入队、worker claim、处理记录 claim、准入、确认 thought。Redis
  拥塞、Pod 没有空闲槽位、或数据库慢都会吃掉它。
- 所有测量都要记录 UTC 时间、部署 commit、App ID、organization ID、session ID、
  delivery ID。

### 4.2 验证清单

逐项执行并记录结果（通过 / 失败 / 未验证）：

1. **Delegation**：把一次性 issue 委派给 App。5 秒内 webhook 得到 200；10 秒内出现
   确认 thought；随后出现 response。
2. **@mention**：在一次性 issue 评论里 @ App。确认是否开新 session、comment 文本是否
   成为 prompt、prompting user 解析是否正确。
3. **Follow-up**：在会话里追加消息。通过确认 / 追问上下文存在；删除旧 session 前后的
   行为符合预期。
4. **并发追问**：turn 运行中再发一条。立即回“Queued”，本轮结束后合并进下一轮，顺序
   保持 FIFO，不并发。
5. **Stop（跨 Pod）**：长 turn 中按 Stop。任何 Pod 收到都能取消；确认在 turn 真正停止
   后到达；已排队消息被丢弃；之后还能在同一会话继续。
6. **重投不重复**：用相同 `Linear-Delivery` 和相同 body 重放。返回 200、decision 为
   duplicate、只有一条 processing record、只跑一次 Agent。
7. **持久化后补发**：制造一次回复投递失败（例如暂时错误的 Linear token），确认记录
   是 `FAILED` 且 `output` 已持久化；在 Deliveries 里 Resend 成功且**没有**新的
   invocation。
8. **不确定结果不重跑**：turn 中途杀 Pod。下一个持有者把记录标为 `FAILED_UNCERTAIN`
   并在会话里报“cut short”；没有自动重跑；新消息继续同一会话。
9. **allowlist**：非 allowlist 用户触发时收到 error activity 且没有运行；不同用户
   follow-up 不能借用 creator 身份。
10. **凭证失效**：撤销 Installation 后触发，页面变为需重装，投递收到 error 而不是静默
    失败。
11. **读取窗口**：未产生事件时显示“当前记录范围内未发现”，不是“失败”；超过 200 条
    或 30 天后明确提示读取范围不足。
12. **Pi / Cursor**（若使用）：多轮共用同一 box session，Stop 走到 `AbortSession`。

### 4.3 证据格式与脱敏

每条证据只保留：UTC 时间、部署 commit、App ID、organization ID、session ID、
delivery ID、HTTP 状态、观察到的字段**路径**（不含完整值）、activity 类型、长度 /
哈希、结果。禁止提交：access/refresh token、client secret、webhook signing secret、
签名原文、完整 `promptContext`、完整 issue 正文、完整用户消息。fixture 只保留形状，
并标注 `capture_status`（例如 `synthetic_shape_only`），不把合成数据写成 live 证据。

## 5. 排错

### 5.1 接收路径

| 症状 | 可能原因 | 怎么确认 / 修复 |
|---|---|---|
| 会话一直 unresponsive，没有确认 | webhook URL 未填、App 未启用 inbound、组织未安装、Redis 不可用、Pod 无空闲槽位 | 看 webhook HTTP 状态；确认 App.inbound_enabled；看 `ListLinearInstallations` 是否含该组织；检查 durable Redis 与 worker 并发 |
| `401 unauthorized` | 签名不对、代理改写了 body、时钟偏移超过 ±60 秒 | 用原始 body 重新计算 HMAC；确认 secret 与 Linear App 一致；校时 |
| `200` 但没有反应 | 事件不是 `AgentSessionEvent`（issue 事件按设计忽略）、action 不是 `created`/`prompted`（未知 action 按设计忽略并返回 200）、inbound 已关、组织未安装、多 Installation 且 payload 无 `organizationId` | 看处理记录是否存在；对照 §5.2 的忽略条件 |
| `404 unknown app` | URL 里的 app ID 不对，或 App 已删除 | 从 App 页面复制正确 webhook URL |
| `413` | body 超过 1 MiB | 检查是否被代理放大或发送了异常 payload；签名覆盖整个 body，必须先完整读取 |
| `503` | Redis 入队失败、凭据读取/解密失败 | 检查 Redis 持久化与主密钥；Linear 会自动重投 |
| 安装回调总是失败 | callback URL 没填、base URL 配错、state 过期/复用 | 对照 §3 步骤 3 与 reason 表 |

### 5.2 会话内错误

| 会话里的 error | 原因 | 处理 |
|---|---|---|
| 用户未被准入 | allowlist 不含该 Linear user ID，或事件没有 prompting user 而非空 allowlist fail closed | 补充 allowlist；确认 Linear payload 中的 user 字段 |
| App 已禁用 / Agent 未加载 | 投递被接受后配置变了 | 重新启用；确认 Agent 仍是 ACTIVE |
| 超时 | 超过 `max_run_seconds` | 调整运行时长或优化 Agent |
| cut short | 持有者崩溃，turn 被标为 `FAILED_UNCERTAIN` | 不会自动重跑；发新消息继续 |
| 凭证需重装 | token 被撤销或刷新被拒 | 重新安装 Installation |

### 5.3 “配置完成但没有成功证据”

按顺序排除：

1. 是否真的从 Linear 触发过事件？没有事件时页面应显示“当前记录范围内未发现”，这是
   “未观察到”，不是失败。
2. 事件是否被拒绝（allowlist、App 禁用、组织未安装、未知 action）？被拒绝会留下记录。
3. 事件是否失败（Linear API 401、超时、投递失败）？失败会给出具体原因。
4. 记录是否属于当前 App + Installation + 当前 Agent？改绑 Agent、重装、改配置后旧记录
   不再算数。
5. 是否超出读取窗口（最近 200 条 / 30 天）？页面会明确提示。

### 5.4 回复已产生但没有送达

Deliveries 里找到 `FAILED` 且有持久化 output 的记录 → Resend。它只重发已持久化的回复，
不会重新运行 Agent。`FAILED_UNCERTAIN` 没有可靠的重发语义：让用户在同一会话发一条新
消息。没有“重跑”动作是设计决定（ADR-0009）。

## 6. 读取窗口与诊断可见性

- `ListLinearProcessingRecords` 按 app / status 过滤，最新在前，默认 50 条、上限 200 条；
  记录保留 30 天。页面上的“历史”只在这个窗口内成立。
- 诊断按 agent / App / Installation / organization / issue identifier / delivery /
  record / invocation / 时间关联，区分：配置未完成、已配置但 inbound 关闭、inbound 已启用
  （desired state）、未安装 / 最后已知可用 / 需重装、无成功证据、最近已跑通、
  失败 / 停止 / 不确定。
- 读权限失败会显示为读取失败，不会被当成“没有事件”。
- 验证视图只在可见时每 5 秒刷新；不看时不轮询。
- 凭证、token、secret、签名永不进入 proto、API 响应或页面；错误摘要由运行时清理。

## 7. 与其他入口的关系

- Telegram、Dashboard、AG-UI 等入口各自有独立的状态展示；Linear 接入不会改变它们的
  行为。`Manage → Linear Apps` 与 Agent 编辑页的 `Connect Linear` 操作同一个 App 资源。
- 删除被 Linear App 路由的 Agent 会被拒绝（failed precondition，列出 App ID），与
  Telegram Destination 的引用保护一致。
- 一个 Agent 可以被多个 Linear App 使用（每个 App 是 Linear 里一个独立身份）。

## 8. 剩余限制

以下内容在当前仓库中**未确认**或明确不支持，不能因为页面显示“已启用”就当作已验证：

1. **没有 live validation**。delegation、@mention、follow-up、Stop、重投、成功/失败
   回复、长回复的真实 payload 与 HTTP 行为都没有真实组织证据；`Linear-Delivery`、
   `organizationId`、prompting user 字段优先级、stop 形状、timestamp 单位、`8000` rune
   回复上限都只是代码 / fixture 结论。
2. **回复长度**：8000 runes 截断沿用 butter-box#20 的实现假设，Linear 的真实上限待确认。
3. **身份**：prompting user 只作为 Memory Capture 的 principal（`linear:<user_id>`），
   不是已验证的 Butter 用户；allowlist 依赖 payload 中的 user 字段，来源仍待 live 确认。
4. **组织解析**：缺少 `organizationId` 时仅在 App 只有一个 Installation 时回退；多
   Installation 无法可靠选择时按失败忽略。
5. **Pi / Cursor 进度**：在 pibox / cursorbox 产出工具事件前，只有确认 thought 和最终
   response，没有中间 action。
6. **不支持**：按 team / project / label 路由、Linear 作为出站目标（Cron / Notify
   Group）、session plan、`auth` 信号与账号关联、repository suggestions、图片附件；
   Agent 需要读写 issue 时请用 Linear 的 MCP server。
7. **平台统一 OAuth App**：v1 每个 workspace 自建 OAuth App；一个 App 只能路由一个
   Agent，改绑会开新会话历史。
8. **Human Input**：仅覆盖 Workflow 的 Human Input Interrupt；只有“唯一字段是单选”的
   表单才使用 `select` 信号，其余表单以字段说明纯文本追问。
9. **没有 health / connected 状态**：页面只给历史证据与最近故障，不承诺当前可用性。
10. **Redis 丢失不可恢复**：已接受但未处理的投递、排队 follow-up 和租约会丢；处理记录
    与已持久化的回复仍在数据库中，但不会自动重放。

## 9. 回归测试地图

验证入口行为时优先跑这些现有测试；它们都从公共入口或已声明的契约出发，不断言内部
goroutine 或 Redis key。

| 范围 | 测试 |
|---|---|
| Webhook 签名、时间窗、去重、组织解析、忽略语义 | `internal/handler/http/linear_webhook_test.go`、`linear_protocol_fixture_test.go` |
| OAuth 安装回调与失败原因 | `internal/handler/http/linear_oauth_test.go` |
| 确认 / 对话 / 追问 / Stop / 超时 / 长回复截断 / 凭证失败 | `internal/runtime/linear/orchestrator_test.go`、`followup_test.go`、`stop_test.go`、`recovery_test.go`、`humaninput_test.go` |
| 多 Pod 会话租约与队列 | `internal/runtime/linear/coordination_test.go`、`worker_test.go` |
| App / 投递记录的权限、revision、补发 | `internal/application/linear_app_service_test.go`、`linear_processing_service_test.go` |
| 引导式接入、诊断文案、跨入口一致性与启用保存 | `front/e2e/linear-agent-setup.spec.ts`、`linear-diagnostics.spec.ts`、`linear-onboarding-journey.spec.ts` |

fixture 与文档结论的一致性由 `TestStageOneFixturesAreAcceptedAtTheWebhook` 保证：它重放
`docs/research/fixtures/linear-agent-session/agent-session-events.json` 的脱敏请求形状，
并断言字段路径、组织路由与重投去重；fixture 的 `capture_status` 必须是
`synthetic_shape_only`，否则测试失败。
