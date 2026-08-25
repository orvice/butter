# pi RPC / SDK 集成调研：在 butter-box 中实现 pi API server

日期：2026-08-25
来源：`earendil-works/pi` `packages/coding-agent/docs/rpc.md`（协议事实均核对自该文档）、pi.dev/docs/latest/sdk、butter-box 现有代码、butter `internal/runtime/opencode` bridge。

> **时效备注**：本文的「butter 侧接入」章节（RemoteAgent 协议、unary SendMessage、一次性会话）已被后续设计取代——最终决策见 `docs/adr/0011-pi-agents-backed-by-butterbox.md`（`AGENT_TYPE_PI` + ButterBox 资源 + 异步 SubmitMessage/GetTurn + 持久会话映射）。协议事实部分（pi RPC/SDK 章节）仍然有效。

## 背景

目标：让 butter 能把部署在远端服务器上的 pi coding agent 当作一个 agent runtime 使用。pi-web（agegr/pi-web）只是 pi 的浏览器 UI，接口私有且未文档化，不适合作为对接面；正确的对接点是 pi 自身的 RPC 模式。butter-box（`github.com/orvice/butter-box`）的容器镜像已预装 `pi` 和 `pi-web`，且已有 HTTP server（MCP endpoint + pi-web 反向代理 + 进程 supervisor），是承载 pi API server 的自然位置。

## pi RPC 协议要点

启动：`pi --mode rpc [--provider <p>] [--model <pattern>] [--session <path|id>] [--session-dir <dir>] [--name <n>] [--no-session]`

- **传输**：stdin 收命令、stdout 出响应和事件，严格 JSONL（只以 `\n` 分帧，容忍尾部 `\r`；不能用会按 U+2028/U+2029 分行的通用 reader）。
- **关联**：命令可带 `id`，响应回带同一 `id`；`bash_execution_update` 事件也回带发起命令的 `id`。
- **命令面**（节选）：`prompt`（支持 `images`、流式期间需 `streamingBehavior: steer|followUp`）、`steer`、`follow_up`、`abort`、`new_session`、`switch_session`、`fork`、`clone`、`get_state`、`get_messages`、`get_entries`（带 `since` 游标，可跨重启增量拉取）、`get_last_assistant_text`、`set_model` / `get_available_models`、`set_thinking_level`、`compact` / `set_auto_compaction`、`set_auto_retry`、`bash`、`get_session_stats`、`set_session_name`、`get_commands`（扩展命令 / prompt 模板 / skill，`/name` 方式经 `prompt` 调用）。
- **事件流**：`agent_start` → 若干 turn（`turn_start`、`message_start`、`message_update`（text/thinking/toolcall 增量）、`message_end`、`tool_execution_start/update/end`、`turn_end`）→ `agent_end`（可能跟随 retry/compaction/排队消息）→ **`agent_settled`**（完全静止，判定"这轮跑完了"必须等它，不能只等 `agent_end`）。另有 `compaction_*`、`auto_retry_*`、`queue_update` 等。
- **Extension UI 子协议**：扩展可能发 `extension_ui_request`（`select`/`confirm`/`input`/`editor` 为阻塞对话，需要客户端回 `extension_ui_response`；`notify`/`setStatus` 等 fire-and-forget）。headless 服务端若不处理对话类请求会造成挂起，必须至少自动回 `{"cancelled": true}`。
- **错误**：失败命令回 `success:false` + `error`；解析失败回 `command:"parse"`。
- **会话存储**：默认持久化在 `~/.pi/agent` 的 session 目录（JSONL，append-only 树，entry id 稳定）。`--session <path|id>` 可重新挂载既有会话；`--no-session` 为临时模式。**pi-web 读同一份数据**，所以 API server 驱动的会话自动能在 pi-web 里浏览 / 导出 / fork。

## pi SDK（`@earendil-works/pi-coding-agent`，Node/TypeScript）

RPC 之外的另一条嵌入路径：同进程使用 SDK。要点（核对自 pi.dev/docs/latest/sdk）：

- **核心 API**：`createAgentSession()` 工厂 → `AgentSession`（生命周期、消息历史、模型状态）；`AgentSessionRuntime`（新建/切换/分支/导入会话）；`ModelRuntime`（模型目录与凭证，优先级：运行时覆盖 > auth.json > 环境变量）。
- **prompt 语义**：`session.prompt()` 的 Promise 在整轮 run（含自动 retry）完全结束后才 resolve——等价于 RPC 的 `agent_settled`，不需要自己判定静止点。流式输出仍靠 `session.subscribe()` 订阅与 RPC 相同形状的事件。`steer()` / `followUp()` / `abort()` / `dispose()` 与 RPC 命令一一对应。
- **会话持久化**：`SessionManager.inMemory()` / `create(cwd)` / `continueRecent(cwd)` / `open(filePath)`——底层就是同一套 `~/.pi/agent` JSONL 树，pi-web 同样可见。
- **SDK 独有能力**（RPC 没有的）：
  - `customTools`（`defineTool`）：向 pi 注入宿主自定义工具；
  - `resourceLoader` / `systemPromptOverride`：程序化控制 system prompt、扩展、skills 的装载；
  - 扩展工厂直接在宿主进程内注册事件钩子（如程序化审批危险命令，而不是 headless 下一律 `cancelled:true`）；
  - 单进程可创建多个 `AgentSession`（文档提到 "create multiple instances"，但没有并发安全性的正式承诺）。
- **运行模式导出**：`InteractiveMode`（TUI）、`runPrintMode`（单次）、`runRpcMode`（RPC 模式本身就是 SDK 之上的一层包装）。**SDK 没有内置 HTTP server**，web 集成要自己包。
- **约束**：仅 Node/TypeScript；npm 包无稳定性声明，API 面比 RPC 协议大得多，升级面也大。

### RPC vs SDK：对 butter-box 的取舍

| 维度 | Go 内嵌 RPC 子进程 | Node SDK sidecar 服务 |
|---|---|---|
| 语言 | 与 butter-box 同为 Go，单二进制 | 引入 Node 服务（butter-box 需 supervisor + 反代，模式同 pi-web） |
| 会话成本 | 每会话一个 pi 进程（Node，内存不低） | 单进程多 `AgentSession` |
| 完成判定 | 自己等 `agent_settled` | `prompt()` resolve 即完成 |
| 扩展 UI 对话 | 必须自动应答避免挂死 | 宿主可实现真实审批策略 |
| 自定义工具注入 | 不可能 | `customTools` 直接注入 |
| 兼容面 | 只耦合 JSONL 协议（有专门文档和兼容性说明） | 耦合整个 SDK API（无稳定性承诺） |
| 隔离性 | 会话间进程隔离 | 一个会话崩溃可能拖垮整个 sidecar |

**结论：butter-box v1 用 Go + RPC 子进程**——与仓库语言一致、协议面小且有文档承诺、会话间天然隔离；个人 VM 场景下进程开销可接受（配全局进程数上限）。SDK sidecar 作为升级路径保留，触发条件是出现这些需求之一：大量并发会话、把 butter-box 的沙箱工具注入给 pi（`customTools`）、需要程序化的危险命令审批策略。届时也可以走折中方案：写一个极薄的 Node 入口用 SDK 复刻 RPC 的 JSONL 协议并加私有便利命令（RPC 模式本身就是 `runRpcMode(runtime)` 一行），Go 侧代码不变。

## butter 侧现状（决定对接形状）

- `RemoteAgent` 已支持三种协议：`A2A`、`DAEMON`（ACP，v1 白名单 `opencode`/`codex`）、`OPENCODE_HTTP`（`proto/agents/v1/agent.proto`）。
- `OPENCODE_HTTP` 的实现是 `internal/runtime/opencode.Bridge`：每次 invocation 一次性 create session → 同步 sendMessage 拿合并文本 → 取消时 abort，包成 ADK `agent.Agent`。这是最薄、已验证的远端 runtime 模式，pi 的接入照此镜像即可。

## 方案：butter-box 内实现 pi API server

### butter-box 现状

单个 `http.Server`（`internal/app/run.go`）：`/healthz`、`{MCP_HTTP_PATH}`（bearer auth）、`/`（可选的 pi-web 反代 + `superviseProcess` 子进程托管）。新增 pi API 只需在 mux 上挂 `/api/pi/` 前缀（比 `/` 更具体，天然优先于 pi-web 反代），并复用已有的 bearer auth 与进程托管模式。

### 进程模型

**每个活跃会话一个 `pi --mode rpc` 子进程**：

- 创建会话 → spawn `pi --mode rpc --name <n> [--provider/--model]`，从 `get_state` 读出 pi 自己的 `sessionId`/`sessionFile`，以 pi 的 sessionId 作为 API 会话 ID。
- 空闲超时（如 30 分钟）杀进程释放内存；session 文件仍在磁盘上。
- 对已不在内存中的会话 ID 发消息 → 用 `pi --mode rpc --session <id>` 重新挂载。会话映射因此无需自己的持久化——pi 的 session 目录就是事实来源。
- stdout 读取：自实现 `\n` 分帧（bufio.Scanner 加大 buffer，含 base64 图片的行可能很大）；`extension_ui_request` 对话类一律自动回 `cancelled:true`；每条命令带自增 `id` 做请求/响应关联。
- 一个进程同一时刻只处理一个 prompt：会话正在流式时收到新消息，默认答 409（或调用方显式传 `steer`/`followUp` 转成排队命令）。

### HTTP API 面（全部走 bearer auth，token 独立于 MCP_AUTH_TOKEN 或复用）

```
POST   /api/pi/sessions                {name?, provider?, model?, thinkingLevel?}
                                        → {sessionId, sessionFile, model}
GET    /api/pi/sessions                活跃 + 磁盘上的会话列表
GET    /api/pi/sessions/{id}           get_state + get_session_stats 汇总
POST   /api/pi/sessions/{id}/messages  {message, images?, stream?, streamingBehavior?}
        stream=false：send prompt → 等 agent_settled → get_last_assistant_text
                      → {text, usage, stopReason}
        stream=true ：SSE 透传 pi 事件（message_update / tool_execution_* /
                      agent_settled），与 butter AG-UI sink 的消费习惯一致
POST   /api/pi/sessions/{id}/abort
DELETE /api/pi/sessions/{id}           杀进程；?purge=true 才删 session 文件
```

同步路径注意：等待信号必须是 `agent_settled`；`agent_end` 后可能还有自动 retry、overflow compaction 重试或排队的 follow-up。

### butter 侧接入（第二步，改动很小）

镜像 opencode 模式：

1. proto 增加 `REMOTE_AGENT_PROTOCOL_PI_HTTP`（url + 可选 model/provider 字段）。
2. `internal/runtime/pi` 新增 Bridge：每 invocation `POST /sessions` → 同步 `POST /messages` → ctx 取消时 `POST /abort`，产出单条 model 事件——与 `opencode.Bridge` 一比一对应。
3. `internal/agent/agent.go` 的 RemoteAgent switch 加一个 case。

一次性会话（opencode 平价）足够第一版；因为 pi 会话持久化 + `--session` 重挂载，后续要做"butter session ↔ pi session"的有状态映射也有现成地基。

### 附带收益

- butter-box 已把 pi-web 反代在 `/` 上：API server 跑出来的会话直接出现在 pi-web 里，人可以旁观、导出、fork butter 驱动的会话——机器入口（API）和人的入口（pi-web）共享同一份 `~/.pi/agent` 数据。
- 可选：在现有 MCP server 上加一个 `PiPrompt` 工具（一次性会话、同步返回文本），让 butter 的任何 agent 不改 butter 代码就能把 pi 当工具调——但作为"runtime"语义，RemoteAgent bridge 是正路。

## 风险与开放问题

- **协议稳定性**：pi 迭代很快（文档明确提到 `message_update` 刚移除过 cumulative `message` 字段）。butter-box 镜像里 pi 版本应 pin 住，升级时跑一遍协议冒烟测试。
- **供应商凭证**：pi 用自己的凭证体系（`~/.pi/agent` / 环境变量）。容器需注入 `ANTHROPIC_API_KEY` 等，或挂载 pi 的 auth 配置。
- **并发**：多个 butter invocation 并发到同一 butter-box 时，一次性会话模式天然并行（每 invocation 各一个进程）；需要一个全局进程数上限防止打爆 VM。
- **Extension 对话**：自动 `cancelled:true` 意味着需要人确认的扩展（如危险命令确认）在 headless 下总是走保守分支；如需放行策略，后续可加每会话的 auto-approve 配置。
