# Plane Agent Runs：官方文档协议基线

状态：`DOC_BASED_DRAFT`。本文依据 Plane 官方文档和固定版本官方 SDK 编写。文档明确内容可作为实现基线；未验证内容不得被描述为真实 Plane 验收通过。

## 证据标签与目标

- `DOC_CONFIRMED`：官方文档或 SDK 明确给出。
- `LIVE_UNVERIFIED`：公开资料不足，需目标实例验证。
- `SYNTHETIC_ONLY`：仅用于本地协议测试，不证明 Plane 真实行为。

默认官方 Cloud API origin 是 `https://api.plane.so`。Agent Runs 按官方资料属于 Beta 能力。目标 edition、workspace、App installation、secret 来源、实际 webhook delivery 和运行限制仍属于 `LIVE_UNVERIFIED`。

官方来源：

- [Agents overview](https://developers.plane.so/dev-tools/agents/overview)
- [Building an agent](https://developers.plane.so/dev-tools/agents/building-an-agent)
- [Choose token flow](https://developers.plane.so/dev-tools/build-plane-app/choose-token-flow)
- [App webhooks](https://developers.plane.so/dev-tools/build-plane-app/webhooks)
- [Signals and content payload](https://developers.plane.so/dev-tools/agents/signals-content-payload)
- [Best practices](https://developers.plane.so/dev-tools/agents/best-practices)
- [Plane Node SDK](https://github.com/makeplane/plane-node-sdk/tree/6a31147dc382b6362afb87f8b1fedaa711e7a8f8)

## 安装与认证

| 项目 | 暂定契约 | 证据 |
| --- | --- | --- |
| App 能力 | App 需启用 App Mentions，并配置 Setup URL、Redirect URI、Webhook URL。 | `DOC_CONFIRMED` |
| Bot 安装 | callback 取得 `app_installation_id`；安装后以 installation 绑定 workspace 和 App。 | 前半项 `DOC_CONFIRMED`，绑定完整性 `LIVE_UNVERIFIED` |
| Bot token | 使用 `client_credentials` 并携带 `app_installation_id`，不假设 refresh-token 流程。 | `DOC_CONFIRMED` |
| OAuth state | Bot callback 是否保证 state 往返、过期和重放保护。 | `LIVE_UNVERIFIED`；Butter 必须 fail closed |
| Webhook secret | 来源、作用域、编码和轮换方式。 | `LIVE_UNVERIFIED` |

## Webhook 事件

| 事件 | 文档形状 | Butter 暂定行为 |
| --- | --- | --- |
| Create | `type=agent_run`、`action=created`、Run 对象。 | 只幂等初始化和核对 Run，不启动 Agent；不假设存在 `promptContext`。 |
| Prompt | `type=agent_run_activity`、`action=prompted`，正文位于 activity content。 | 先验证 Run、workspace、App 和 actor 归属，再执行；首个 Prompt 才启动。 |
| Stop | activity `signal=stop`。 | 走控制路径，不作为用户文字；空正文行为为 `LIVE_UNVERIFIED`。 |

Create 与 Prompt 的完整 headers、签名、delivery 重投、缺字段响应和重复响应属于 `LIVE_UNVERIFIED`。实现不得以单安装猜测归属，也不得把 Run creator 当成每条 Prompt 的 actor。

## Run 与 Activity

官方资料和 SDK 列出的 Run 状态为：

```text
created, in_progress, awaiting, completed, stopping, stopped, failed, stale
```

SDK 暴露的 API 为：

```text
POST/GET /api/v1/workspaces/{slug}/runs/
GET     /api/v1/workspaces/{slug}/runs/{run_id}/
POST/GET /api/v1/workspaces/{slug}/runs/{run_id}/activities/
GET     /api/v1/workspaces/{slug}/runs/{run_id}/activities/{activity_id}/
```

Agent 可创建的 Activity 类型为 `thought`、`action`、`response`、`elicitation` 和 `error`；不能创建 `prompt`。`action.parameters` 是字符串映射。Activity POST 的幂等键、成功后的 Activity ID 语义、超时后的查询核对、429/Retry-After 和正文上限属于 `LIVE_UNVERIFIED`。

因此 Butter 要把执行状态和投递状态分开，并先保存冻结 payload：

- 明确成功：记录 `DELIVERED`；
- 明确拒绝且确认未创建：记录可控失败；
- 超时、断连、含糊 5xx 或 POST 后进程崩溃：记录 `UNKNOWN`，不自动重复发送；
- 没有可靠幂等或查询能力时，UNKNOWN 只能由有权限的人确认后受控补发。

## Elicitation 与终态

`elicitation` 可使 Run 等待输入。官方示例包含 select options，但真实回答字段、option ID/label 回传和恢复行为仍为 `LIVE_UNVERIFIED`。首版只承诺纯文本 elicitation，select 保持关闭。

Run 终态后 follow-up 是否复用原 Run、创建新 Run 或拒绝，官方资料没有统一保证，保持 `LIVE_UNVERIFIED`。`external_link` 虽出现在 Create 模型中，SDK 没有展示创建后的 Run 更新方法，不得假设可以回填。

## 实现约束

后续实现必须：

1. 对 webhook raw body 验签并在持久化接受前完成归属和权限校验；
2. 将 Create、Prompt 和 Stop 分成独立事件语义，使用业务 ID 做原子去重；
3. 以持久化接受顺序保证同一 Run 的 FIFO，并用 fenced holder 协调 Stop 与正常完成竞态；
4. 保存完整脱敏的 Activity 投递 payload，UNKNOWN 不自动重发；
5. 将纯文本 elicitation 作为默认降级，select 只有实测确认后才启用；
6. 不把 synthetic fixture、API 2xx、配置完成或代码合并当作真实 Plane 验收证据；
7. 日志和 fixture 不写入 token、secret、签名、完整上下文或未脱敏正文。

真实目标 edition/version、安装 state、secret、Create/Prompt/follow-up/Stop headers 和 payload、actor/归属、Activity response/idempotency、429、终态续聊、error 持久性、纯文本恢复和 select 回传仍需后续实例验证。本文冻结的是文档依据的实现基线，不替代 `live_redacted` 证据。
