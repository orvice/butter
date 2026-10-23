# Pi 接入 Butter A2UI：可行性、接线方案与限制

调研日期：2026-10-23。**调研建议，不是已接受的设计或已实现功能。**

## 结论

**可以接，不需要 fork Pi，也不需要把 ButterBox 从 Go + Pi RPC 改成 Node SDK。** Pi extension 可以注册模型可调用的 `render_ui`；Butter 的 Pi bridge 本来就是一个 ADK 自定义 agent，可以把经过校验的卡片写成 ADK session event，复用当前 A2UI 展示链路。[B1][B2][P1]

建议分开处理两种需求：

1. **只读结果卡片：优先做。** 若要达到现有 ADK `render_ui` 的语义——校验失败反馈给本轮模型、服务端分配 surface ID、成功后可继续更新——推荐 **Pi extension + ButterBox 工具请求/回执中继 + Butter 侧校验和持久化**。
2. **人工输入表单：不要顺手混进第一版。** Pi 的阻塞扩展弹窗不是 ADK 的可持久化 Workflow Interrupt。先用已有 Workflow Human Input 节点编排 Pi；若需要 Pi 自己发起提问、审批，再单独设计生命周期和恢复协议。[B3][B7][P2]

一个更便宜但语义较弱的选项是：Pi 把结构化卡片提案写到工具结果 `details`，Butter 通过现有 `ListEntries` 读取并导入。它适合“任务末尾展示摘要”，**不能仅凭提案已写入 Pi，就告诉模型“卡片已在 Butter 渲染成功”**。

## 1. 来源与验证范围

本次对照：

- Butter：`0375fd41e2f224c10867c64d1e4335f26c89c7f0`。
- 本机 ButterBox 工作树：`f54e085381623551a19410015e21ebf64c3b27dd`。
- Butter `go.mod` 实际依赖的 ButterBox：`5326bff87105`，**比上述工作树旧**。两者都有 `ListEntries`；只有较新工作树已有本文引用的 `GetCapabilities`（目前仅声明原子模型选择能力）。不能把较新源码当成 Butter 已经接入的接口。[B8]
- Pi 官方文档、声明和实现：本机 pi-web 内嵌包 **1.2.0**，对应上游 tag `v1.2.0`（Git ref SHA `06a466935536f8c9bcacb2d5efc0fc021be1bdff`）。本机独立 `pi` CLI 是 **1.3.0**；下述确定性探针显式启动 1.2.0，不混用版本结论。[P1–P5]
- A2UI 的规范背景沿用已有[一手来源调研](a2ui-adk-structured-output.md)：Butter 当前协议是 v0.9.1，catalog 是 `butter-basic-v1`，不是借这次接入升级协议。

**实际运行了一个无外部模型的 Pi RPC 探针；没有实现 Butter/ButterBox 接线，没有跑真实模型、浏览器或跨 Pod 端到端验证。**

### 确定性探针

在 `/tmp/pi-a2ui-probe/` 创建临时 extension 和 Python RPC 客户端：

- 用 `registerProvider` 注册本地确定性假 provider；第一轮发出 `render_ui` tool call，下一轮输出普通文本。无 HTTP 请求、无外部模型调用。
- 用 `registerTool` 注册 `render_ui`，返回结构化 `details.butterA2UI`，并调用 `pi.appendEntry()`。
- 正常发送 `prompt`，等待 `agent_settled`，调用 `get_entries`；退出进程后用 session 文件重挂载再读取。
- 额外通过扩展命令调用 `ctx.ui.custom()` 和 `ctx.ui.confirm()`；客户端按 Box 当前行为回复 `cancelled: true`。
- 使用独立的临时配置目录、显式 `--extension`、关闭自动发现和联网启动行为；没有修改用户配置。

| 验证项 | 结果 |
|---|---|
| RPC 模式加载扩展并执行模型工具调用 | 成功，有 `tool_execution_start/end` 和后续模型回复 |
| 工具结果 `details` 经 `get_entries` 读取 | 完整保留 |
| `appendEntry` 的 `customType/data` | 完整保留，出现 `entry_appended` |
| 进程退出、重挂载后 | 9 个 entry 全部保留，包含工具 details 和 custom entries |
| `get_entries(since)` | 返回游标之后的新 entry |
| RPC 下 `ctx.ui.custom()` | 返回 `undefined`，未执行 TUI factory |
| `ctx.ui.confirm()` | 发出 `extension_ui_request`；取消后返回 `false` |

该探针证明 Pi 的基础能力，不证明真实模型的出卡率，也不证明 Box → Butter → Chat 已连通。原始结果暂存于 `/tmp/pi-a2ui-probe/result.json`；临时文件不是仓库长期测试资产。

## 2. 当前链路：哪些已经有，哪里没有接

```text
Chat（协商 butterA2UI）
  → Butter AG-UI handler（绑定、线程 lease、Detached Run）
  → ADK runner
  → pibox.Bridge（自定义 ADK agent）
  → ButterBox SubmitMessage / GetTurn
  → pi --mode rpc
```

### 2.1 Pi 端有工具扩展能力，但没有 A2UI 专用协议

Pi 官方 extension 的 `registerTool` 支持描述、TypeBox 参数、执行函数、结构化 `details` 和执行模式；RPC 模式也加载 extensions。官方 todo 样例从 active branch 的工具结果 details 重建状态。[P1][P4]

需要区分三种“UI”：

- `renderCall/renderResult` 和 `ctx.ui.custom()` 是 **Pi 终端 UI**，不是 A2UI JSON，也不是可发给 React 的远程控件。
- `extension_ui_request` 是 **RPC 弹窗子协议**：select/confirm/input/editor 等待响应，notify/widget 等是通知。
- Butter 的 **A2UI** 是受 catalog 限制的声明式卡片，通过 AG-UI `CUSTOM butter.a2ui` 发送。[P1][P2][B3]

所以不能安装一个 Pi TUI 插件，就期待 Chat 自动显示同样的 UI。

### 2.2 ButterBox 有可恢复的结构化 transcript，但 GetTurn 只返回文本

`ListEntries` 原样返回 entry JSON，不丢弃 tool-result `details`、custom entries 或 custom-message details。Pi 的 `get_entries` 包含压缩前历史及废弃分支，稳定 entry ID 可用于增量读取。[B4][P3][P5]

但当前：

- `GetTurn` 的 `TurnResult` 只提供最终 assistant 文本、stop reason 和统计信息。
- `StreamMessage` 能转发普通 Pi 运行事件，但其调用承担发送消息和等待运行的生命周期；它不是 Detached Run 可恢复的事件日志。
- `process.dispatch` 会自动取消所有阻塞 `extension_ui_request`，并丢弃其他 UI 通知；它们不进入事件订阅者。
- `SubmitMessage` 后台监督等到 `agent_settled`，但不把途中事件保存为 Butter 可消费的运行日志。[B4]

### 2.3 Butter 已有通用卡片存储和前端，但 Pi bridge 只产出最终文本

`pibox.Bridge.run` 当前的输出：新建 Pi session 时写 binding `StateDelta`，运行结束后 yield 一条文本事件；`await` 看到 `running=true` 就继续等，不读取 entries 或工具请求。[B1]

现有 LLM 路径则是：

```text
render_ui
  → a2ui.Apply（整批校验、服务端 ID / revision）
  → tool 的 StateDelta
  → ADK AppendEvent
  → aguiSink.emitUI
  → CUSTOM butter.a2ui
  → Detached Run Log / SSE
  → 前端 A2UIStore / React renderer
```

卡片存在 `butter:a2ui:card:*`，值为 JSON 字符串；snapshot、历史定位、attach 重放均已有实现。因此 Pi 接入应复用这条链路，而不是另开前端解析器或卡片数据库。[B2][B3][B5]

## 3. 方案比较

以下都是设计选项，不是现有接口。

| 方案 | 改动与好处 | 限制 / 判断 |
|---|---|---|
| A. 在 Pi 正文输出 A2UI JSON，Butter 或浏览器提取 | 接线看似最少 | 混淆正文与协议；历史、其他入口、错误纠正复杂。浏览器不能成为唯一校验者。**不推荐作为正式方案** |
| B. extension 工具把提案写入 `details`，Butter 拉 `ListEntries` 导入 | 不必先新增 Box 公共 RPC；支持最终结果恢复；可做最小 PoC | Butter 校验结果不会自动回到同一 Pi 工具调用；surface handle 需映射；重复校验与状态归属容易分裂。**适合明确限缩的摘要卡片版本** |
| C. extension 工具请求经 Box 中继，Butter 执行并回执 | 共用现有校验、ID、限额、状态；本轮模型能修正参数；不要求 Box 反连 Butter | 需新增每 turn 能力配置、工具请求读取和应答；实现重试、取消、幂等。**推荐生产路径** |
| D. extension 通过 HTTP/MCP 直接调用 Butter | 天然请求/响应，可把错误返回模型 | 需要 Box → Butter 网络、短期授权及精确 run 路由；不能为了写卡片重拿已被该 run 持有的线程 lease。可行，但运维和鉴权成本高于 C |
| E. Box 改成 Node SDK host，以 `customTools` 执行宿主回调 | 宿主直接掌控工具 | 引入新的进程和部署形态；RPC + extension 已能提供所需工具。**没有必要仅为 A2UI 重构运行时** |

### 为什么 B 不等同于现有 `render_ui`

例如 Pi 工具先返回 `created`，Butter 随后发现第 21 张卡片超额：Pi 模型已经继续运行，无法从这个工具结果得知失败。若 Pi 先生成自己的 `surface_id`，而 Butter 又由 `a2ui.Apply` 生成真正 ID，还需持久化映射，支持后续更新、删除、重复导入与 session 重建。[B2][B5]

若选择 B，建议契约明确为 **“提交卡片提案”**，第一版只支持创建只读摘要，返回 `queued` 而非 `rendered`；Butter 侧最终校验，失败保留文字。不要把它包装成与 ADK 一样的完整 `render_ui`。

## 4. 推荐方案 C：让 Pi 工具调用穿过已有卡片 seam

### 4.1 数据流

```text
Pi 模型调用 render_ui(args)
  → 受信任 extension.execute 等待宿主结果
  → Box 本地工具中继（如 UDS；具体实现待验证）
  → Butter 主动拉取该 turn 的待处理工具请求
  → 校验 run / 调用身份 / Card Policy
  → a2ui.Apply(currentCards, batch, ButterAnchor, limits)
  → yield 非 partial 的 ADK 工具结果事件 + 卡片 StateDelta
  → ADK session 存储成功
  → sink / Run Log / Chat 按既有链路展示
  → Butter 应答 Box，extension 返回 {surface_id, revision, status}
  → Pi 模型继续回答或修正调用
```

**Pi 不直接向 stdout 打印自定义 A2UI 记录。** stdout 是 Pi 自己的 JSONL 协议；扩展应使用明确定义的本地传输，而不是与 RPC framing 争用输出。[P3]

### 4.2 模块分工

- **Pi extension：薄 adapter。** 注册工具、提供与 Butter 同源的描述/参数、处理取消、把宿主结果交回模型。不要复制另一套完整卡片状态机。工具串行执行，避免同一 Pi 回合对同一卡片乱序更新。[P1][P4]
- **ButterBox：工具中继 module。** 只管理 Pi session/turn、请求关联、背压、等待、应答和失联清理；不理解 `Card`、不计算 UI revision。
- **Butter：卡片权威。** 共享 `a2ui.Apply`；把当前 `a2uitool` 中依赖 ADK tool context 的 adapter 与通用执行逻辑分开，让 ADK LLM 和 Pi bridge 成为两个 adapter。
- **前端：沿用现有 Interface。** 继续消费 `butter.a2ui` 和 snapshot；无需知道这张卡片来自哪个运行时。仍需验证新事件的 message anchor、失败工具展示、刷新和重放行为。[B2][B3][B5]

这次已有两个真实调用方，值得抽出共享 seam；不必先做一个能桥接所有远端工具的通用 RPC 平台。

### 4.3 需要新增的 Box Interface

建议最小范围；名称仅为示意：

1. 能力发现：声明版本化的 host-tool relay 能力。Box 支持中继、目标 Pi session 已加载兼容 extension，是两件事；后者也必须确认。
2. `SubmitMessage` 的 **per-turn** host-tool 配置：是否开放 `render_ui`、版本/catalog、有效策略、唯一 turn nonce。与 prompt 提交共享 run slot，避免配置落到另一轮。
3. 可恢复读取待处理 host calls：可扩展 `GetTurn` 或另设读取 RPC；每项有 turn ID、call ID、工具名和参数。
4. 按 turn/call 应答结果或工具错误；重复应答幂等，跨 turn 应答拒绝。

**关键：不能只给现有 GetTurn 增加一个字段就结束。** 当前长轮询等到 settle；Pi 工具若正等 Butter 的应答，run 不会 settle。读取必须在出现 host call 时立即唤醒，且“没有新请求”和“工具请求已被处理”要有清楚语义。[B1][B4]

旧 Box 对新增 protobuf 字段可能静默忽略，必须能力协商；不支持时明确走文本降级，或按产品要求拒绝本轮 A2UI，不能悄悄启用一个无人应答的工具。[B8]

### 4.4 复用现有持久化时的一个隐藏障碍

**只 yield 一个含卡片 StateDelta 的空事件，不足以保证实时显示。**

ADK `Event.IsFinalResponse()` 会把无 function call/response 的非 partial 事件判为 final；Butter runner 当前不把普通 final event 交给 `onEvent`。这类事件可能持久化了，却没到 `aguiSink.emitUI`，直到刷新 snapshot 才看到卡片。[B6]

因此需要：

- 将 Pi `render_ui` 调用及结果正规化为 ADK FunctionCall / FunctionResponse，卡片 delta 随结果事件持久化；或
- 明确扩展 runner 的“必须交给 sink 的状态事件”规则，并测试对其他入口的影响。

不要将卡片事件标为 partial 来绕过过滤：ADK 不持久化 partial event，违反先存后发。[B6]

`yield` 返回也不能脱离上下文被当成通用“提交回执”。当前直接运行路径先 AppendEvent 再向调用方 yield，但插件及 composite/workflow 的事件处理也参与其中；应通过最小接线测试确认回执点，持久化失败绝不返回成功。[B6]

### 4.5 Card Policy 与叶子 Agent 规则需要显式修订

当前 PI/CURSOR 共用 `rejectBoxOwnedFields`，均拒绝 `result_cards`；前端也隐藏两者的策略配置。`treeBuilder` 虽然计算父级继承策略，但现有 `PiAgentBuilder` 只收到 proto，没有收到已解析的 policy。[B2][B7]

若 Pi 支持卡片：

- 把它定义为 **Butter Chat 的呈现能力**，而不是把 Pi 的 instruction / MCP / skills 配置搬回 Butter。
- 允许 PI 的 `result_cards`，但保留 CURSOR 的现有限制，不可直接删除共用 helper 中的一项而意外同时放开 Cursor。
- 从 tree builder 把同一份有效 Card Policy 传给 Pi adapter，保留“父级 DISABLED 只能收紧、presentation 最近显式值优先”的规则；不要在 handler 按名字再遍历一次树。
- 非协商运行不暴露工具；禁用后旧卡片和 snapshot 仍可读，本轮不能继续修改。
- `PREFERRED` 仍只改变工具描述，不污染 Pi 的持久 instruction。[B3]

这影响 ADR-0011 / ADR-0014 的现有约束，实施前应记录明确的修订决定；本调研不替代该决定。

## 5. 故障、安全与恢复要求

以下是方案需要满足的契约，不是现状已提供的保证。

1. **先持久化，再成功回执。** 同一逻辑写入保存 card delta 和 call 的结果标记。应答丢失后重读不能重新创建卡片或再加 revision。幂等键至少包含 Pi session、turn nonce、call ID；同一 ID 携不同参数要拒绝。
2. **不要宣称天然 exactly-once。** 现有 Mongo `AppendEvent` 先插 event、再更新 session state，不是一个事务。新增回执/重放必须覆盖中间失败，考虑从已存 event 恢复结果或其他幂等策略，而非仅靠内存 map。[B9]
3. **身份来自 Butter。** principal/workspace/agent/thread/run/message anchor、surface ID、revision 均由 Butter 决定。Pi 参数和 entry 的自称身份不可信。更新卡片必须属于该绑定会话；不接受任意 HTML/脚本/URL/actions，沿用当前 catalog 与限额。[B2][B3][B5]
4. **每 turn 隔离，结束即撤销。** 会话级环境变量或全局 MCP 配置不足以区分 AG-UI、Telegram、cron、pi-web。旧 turn 的迟到应答不能碰下一轮。Box 本地凭证不能是 Butter root token；同进程扩展仍具 OS 权限，不能将本地 socket 鉴权误称为对恶意代码的沙箱。[P1]
5. **不向外投递 raw transcript。** `details` 不进模型上下文不代表保密；里面可能含敏感工具数据。仅桥接允许的工具和有限 payload，不把整个 Pi session entries 发给浏览器。
6. **取消与失联必须释放等待。** Stop、max-run timeout、lease 丢失、Box 重启、Butter Pod 消失时，仍走现有 AbortSession 取消路径；中继另需有界等待、过期和资源清理，不能让 Pi 永久占住 run slot。浏览器断开只 detach，不应取消工具。[B1][B4]
7. **避免重复运行。** 只重试 poll/ack；不自动重试可能已被接受的 SubmitMessage。已落盘的卡片保留，run 失败不伪造事务回滚。
8. **并发要在权威侧保序。** Pi 工具可设 sequential，但 Parallel Agent 下可能有多个 Pi/LLM 叶子共同写卡片；线程 lease 只序列化不同 run，不序列化一个 run 内所有工具。需对卡片更新和幂等结果统一排序。[P1][B3]

### 若采用方案 B，另外有三个 transcript 陷阱

- `get_entries(since)` 是 **append-order 全历史**，不是 active branch。`leafId` 是树指针，不一定是最后追加的 entry。导入必须明确分支策略、按 ID 去重，不要把 leaf 当成通用日志尾游标。[P3][P5]
- Box `GetTurn` 从 cursor 后选最后一个 assistant，没有独立的不可变 turn 结果上界；不能把别的后续消息或 pi-web 分支误认为当前 Butter 运行的卡片。[B4]
- Pi 先发公开 `message_end` 再 append session entry；刚启动、尚无 conversation 时部分 custom entries 也可能仍只在内存。实时通知不等于磁盘持久化回执。[P5]

## 6. 为什么交互表单应另做

当前 Butter 表单在 Workflow 暂停时，把字段规则、submit token、revision 和 Interrupt ID 固定到 session event；运行结束，lease 释放，稍后提交被校验后触发一次新运行。pending 由 session events 推导，无第二个 pending store。[B3]

Pi 的 `ctx.ui.select/confirm/input/editor` 则是在当前扩展调用栈中等待 Promise：

- Box 现在直接自动取消；
- 仅改为转发，会让当前 Pi run 和进程继续活着；用户不回答时等待多久、Stop 怎么做、重启后 Promise 如何恢复，都没有自动答案；
- Butter 现有提交路径要求一个真实的 ADK Interrupt；展示一个形似表单的 A2UI surface，不会自动产生可 resume 的执行状态。[P2][B4]

**近期做法：** 用 Workflow Human Input 节点包住 Pi 工作步骤，复用已有表单与恢复机制。

**若做 Pi 原生交互：** 必须先选择“活 run 内应答”或“终止本轮并持久化 continuation，提交时新起一轮”。前者需要新的 in-flight input/Stop 协议，后者需要 Box/Pi 可恢复执行契约。两者都不是把 `extension_ui_response` 改名成 AG-UI `resume`。

审批也必须 fail closed：取消、过期、丢失不能视为允许。只读卡片支持不应顺带开放模型任意按钮和执行动作。

## 7. 建议实施顺序与验收

### 第一步：最小接线 PoC（不做交互表单）

- 显式加载受信任 Pi extension，注册 `render_ui`；使用确定性假 provider。
- 验证一条 host call 从 Box 被 Butter 拉到，`a2ui.Apply` 失败能回到同一 Pi 工具调用；成功返回真实 surface ID。
- 证明卡片在 Pi 最终文字结束前到达 Chat；session 存储失败时不能显示或返回成功。
- 验证 composite/workflow 下事件存储与回执时序。若必须先走方案 B，应明确降格为摘要提案，不承诺完整工具语义。

### 第二步：生产化只读卡片

| 范围 | 主要改动 |
|---|---|
| `orvice/butter-box` | extension 打包和版本固定；本地工具中继；能力发现、per-turn 配置、请求读取/回执；取消与超时测试 |
| Butter `internal/a2uitool` / `internal/a2ui` | 共享执行 seam；工具参数与描述同源；限额与失败计数保持一致 |
| Butter `internal/runtime/pibox` | 带策略的构建参数；host-call poll/ack；事件正规化与幂等 |
| Butter Agent 校验/前端配置 | PI 卡片策略；保留 CURSOR 拒绝；父级策略继承 |
| Chat 与文档 | 主要复用；补实时、snapshot、history、attach、Stop 的集成用例；修订相关 ADR/API 文档 |

最低验收：

- 新建、跨 turn 更新、删除；同一 call 重复投递不重复写入。
- 非法组件、悬空引用、超额整批拒绝，错误返回模型；连续失败上限保持一致。
- 没协商、旧绑定、策略禁用、旧 Box/缺 extension 时安全降级，不挂死。
- 父级 DISABLED、多个并行叶子、跨 workspace/agent/thread、过期回执隔离。
- 断线重连、snapshot、history anchor 与 attach Run Log 重放一致。
- Stop、deadline、Pod/Box 重启；提交应答和工具回执丢失；Mongo event/state 中间失败。
- 真实 Pi 模型测试工具选择率和参数质量，不能拿假 provider 探针当生成质量结论。

### 第三步：单独定义 Pi 原生提问/审批需求

在确定是否真的需要 Pi 扩展内部阻塞交互后，再写 spec；不作为只读卡片的依赖。

## 8. 对旧调研的修正

[早期 Pi RPC 调研](pi-rpc-integration.md) 把“RPC 自定义工具注入”写为“不可能”。应缩小为：**RPC 命令面没有 SDK 的 `customTools` 参数，但 RPC 子进程可加载 extension，由 extension 注册工具。** 本次已通过真实工具调用探针验证；因此“接 A2UI 必须改 SDK sidecar”不成立。[P1][P4]

## 一手来源索引

Butter 链接指向本次工作树；代码结论以 §1 的 commit 为准。Pi 链接固定 `v1.2.0`，Box 链接固定本次调查的较新工作树。

- **[B1]** [`internal/runtime/pibox/bridge.go`](../../internal/runtime/pibox/bridge.go)：`BuildAgent`、`run`、`submit`、`await`、`yieldBinding`、`abort`。
- **[B2]** [`internal/a2uitool/a2uitool.go`](../../internal/a2uitool/a2uitool.go)：`Tools`、`render`、`apply`、工具参数/描述；[`internal/agent/agent.go`](../../internal/agent/agent.go)：`treeBuilder.build`、`newLLMAgent`。
- **[B3]** [ADR-0014](../adr/0014-a2ui-surfaces-over-agui.md)、[ADR-0011](../adr/0011-pi-agents-backed-by-butterbox.md)、[ADR-0016](../adr/0016-agui-runs-outlive-their-request.md)。
- **[B4]** ButterBox [`pi.proto`](https://github.com/orvice/butter-box/blob/f54e085381623551a19410015e21ebf64c3b27dd/proto/butterbox/pi/v1/pi.proto)、[`process.go`](https://github.com/orvice/butter-box/blob/f54e085381623551a19410015e21ebf64c3b27dd/internal/pi/process.go) `dispatch`、[`manager.go`](https://github.com/orvice/butter-box/blob/f54e085381623551a19410015e21ebf64c3b27dd/internal/pi/manager.go) `SubmitWithModel` / `finishDetached` / `Turn` / `turnResult` / `Entries` / `waitSettled`、[`service.go`](https://github.com/orvice/butter-box/blob/f54e085381623551a19410015e21ebf64c3b27dd/internal/pi/service.go) `ListEntries` / `StreamMessage`。
- **[B5]** [`internal/a2ui/card.go`](../../internal/a2ui/card.go) `Apply` / `DefaultLimits`；[`state.go`](../../internal/a2ui/state.go)；[`agui_sink.go`](../../internal/handler/http/agui_sink.go) `emitUI`；[`agui_history.go`](../../internal/handler/http/agui_history.go)；[`front/src/features/agui-chat/a2ui/store.ts`](../../front/src/features/agui-chat/a2ui/store.ts)。
- **[B6]** [ADK v2.5.0 `runner/runner.go`](https://github.com/google/adk-go/blob/v2.5.0/runner/runner.go#L725-L776)、[`session/session.go`](https://github.com/google/adk-go/blob/v2.5.0/session/session.go#L218-L232) `IsFinalResponse`；Butter [`runner.go`](../../internal/runtime/runner/runner.go) `run` 的 `onEvent` 过滤；[`streamorch/run.go`](../../internal/runtime/streamorch/run.go)。
- **[B7]** [`internal/agent/pi.go`](../../internal/agent/pi.go) `PiAgentBuilder` / `rejectBoxOwnedFields`；[`result-cards-config.ts`](../../front/src/features/agents/result-cards-config.ts) `supportsResultCards`；[`internal/runtime/interrupt`](../../internal/runtime/interrupt)。
- **[B8]** [`go.mod`](../../go.mod)；[Butter 依赖版本的 Box proto](https://github.com/orvice/butter-box/blob/5326bff87105/proto/butterbox/pi/v1/pi.proto)；[较新 Box 的能力声明](https://github.com/orvice/butter-box/blob/f54e085381623551a19410015e21ebf64c3b27dd/proto/butterbox/pi/v1/pi.proto)。
- **[B9]** [`internal/runtime/session/mongo/mongo.go`](../../internal/runtime/session/mongo/mongo.go) `AppendEvent`。
- **[P1]** Pi [Extensions](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/extensions.md)、[声明](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/src/core/extensions/types.ts)：`ToolDefinition`、`executionMode`、动态工具、进程权限。
- **[P2]** Pi [RPC Extension UI](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/rpc-extension-ui.md)、[Terminal UI](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/tui.md)。
- **[P3]** Pi [RPC](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/rpc.md)、[RPC Commands](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/rpc-commands.md) `get_entries`、[JSON Event Stream](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/json.md)。
- **[P4]** Pi [hello tool](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/examples/extensions/hello.ts)、[todo](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/examples/extensions/todo.ts)、[SDK](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/sdk.md)、[Custom Providers](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/custom-provider.md)。
- **[P5]** Pi [Message Types](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/message-types.md)、[Session Format](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/docs/session-format.md)、[`agent-session.ts`](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/src/core/agent-session.ts) 的 `_handleAgentEvent`、[`session-manager.ts`](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/src/core/session-manager.ts) 的 `_hasConversation` / `_persist`、[`rpc-mode.ts`](https://github.com/earendil-works/pi/blob/v1.2.0/packages/coding-agent/src/modes/rpc/rpc-mode.ts) 的 `get_entries`。
