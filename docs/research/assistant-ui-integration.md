# 全面接入 assistant-ui：Butter dashboard 与第三方前端

调研日期：2026-10-04。

> **后续决定（2026-10-04）**：只保留 AG-UI Chat，它迁到 `/chat`；主聊天和异步 Invocation API 一并退役。切换前必须做到刷新后实时续流。计划见 PRD #389。下文"结论"第 1 条"不要合并"的建议因此不再适用；§6 的其余条目已拆进该 PRD 的子 issue。

## 考察的版本

| 包 | Butter 当前（`front/package.json:16-20`） | 最新发布（npm，2026-10-04） |
|---|---|---|
| `@assistant-ui/react` | 0.15.16（2026-08-20） | 0.15.23（2026-10-02） |
| `@assistant-ui/core`（传递依赖） | 0.3.15 | 0.3.22 |
| `@assistant-ui/react-ag-ui` | 0.0.56（2026-08-20） | 0.0.63（2026-10-02），依赖 `@ag-ui/client` `^0.0.59` |
| `@ag-ui/client` | 0.0.58（2026-08-14） | 1.0.1（2026-09-29）；1.0.0 发布于 2026-09-17 |
| `@assistant-ui/react-google-adk` | 未使用 | 0.0.33（只用于 §6 P3 的探针） |
| `@a2ui/react` | 0.12.0 | 不属于 assistant-ui，本文不评估它的升级 |

assistant-ui 的 main（`1f77d04`，2026-10-04）已经把版本号改成 react 0.15.24、react-ag-ui 0.0.64，但这两个版本还没发布，`npm view` 返回 404。

## 来源

只用一手来源。

- **assistant-ui 仓库的两个发布提交。**
  - [`75e3ef7`](https://github.com/assistant-ui/assistant-ui/tree/75e3ef71beb5dc99f6fc624624d3d61b307c8599)：`@assistant-ui/react@0.15.16` 和 `@assistant-ui/react-ag-ui@0.0.56` 两个 tag 都指向它。
  - [`3542d60`](https://github.com/assistant-ui/assistant-ui/tree/3542d602272a62eddeb8989befc910841c267022)：0.15.23 和 0.0.63 的 tag 都指向它。
  - 下文的 `@0.0.56` 和 `@0.0.63` 分别指这两个提交。
  - `front/node_modules` 里 react、core、react-ag-ui 三个包的 `src/` 与 `75e3ef7` 用 `diff -r` 比对，没有差异。npm 上 0.0.63 的 `src/` 也与 `3542d60` 一致。所以文中给出的行号同样适用于已安装的代码。
- **assistant-ui 文档**（<https://www.assistant-ui.com/docs>）。按 `llms.txt` 的索引逐页抓取 `.md` 版本，抓取时间是 2026-10-04。文档站从 main 构建，内容可能比最新发布版更新。
- **AG-UI 仓库。**
  - 协议文档和 TS SDK 取自 tag `@ag-ui/client@1.0.1`，即提交 [`ec9f4fa`](https://github.com/ag-ui-protocol/ag-ui/tree/ec9f4fa68b950d0b440087cd94530c0815b81115)。
  - 0.0.58 没有 tag，改用它的发布提交 [`2ac1dd9c`](https://github.com/ag-ui-protocol/ag-ui/tree/2ac1dd9c961c35a92a4683a89b1eb248be91b660)。已核对：0.0.58 的 dist 和这个提交的源码里有同一条错误文案。
  - 抽查过 docs.ag-ui.com 的 interrupts 页，与仓库里的文档一致。
- **npm registry**：`npm view <pkg> versions dist-tags time --json`。
- **Butter 代码**：`212f98f`。此外读了模块缓存中的 ADK Go v2.5.0，以及 Butter 依赖的 AG-UI Go SDK（`v0.0.0-20260813165816-6691ac50b34a`）。

## 实际运行过的东西

全部在 scratchpad 里完成，没有改动 `front/`。

1. **升级 spike。** 把 `front/` 复制成三份：
   - `pinned`：用原锁文件 `npm ci`；
   - `latest`：react 0.15.23 + react-ag-ui 0.0.63 + `@ag-ui/client` 0.0.59；
   - `agui1`：同 `latest`，但 `@ag-ui/client` 用 1.0.1。

   三份都跑了 `tsc -b`。`pinned` 和 `latest` 还跑了 `vitest run`，以及 11 个聊天相关的 Playwright 文件，共 38 个用例，后端用仓库自带的 fixture 模拟。
2. **两个 Playwright 探针**，只放在 scratchpad 里，在 `pinned` 和 `latest` 上各跑一次：
   - AG-UI Chat 的两种 Interrupt 场景；
   - 主聊天如何渲染持久化的事件。
3. **三个 Node 探针：**
   - 把 Butter 前端工具的事件序列喂给真实的 `RunAggregator`；
   - 用三个版本的 `@ag-ui/client` `HttpAgent` 解析与 Butter 输出同形的 SSE；
   - 用 `@assistant-ui/react-google-adk` 的 `AdkEventAccumulator` 处理 Butter 持久化下来的事件。

没有启动 Butter 后端，也没有调用模型。

## 结论

1. **两个聊天各自的 runtime 都选对了。不要换，也不要现在合并。**
   - 主聊天用 `useExternalStoreRuntime` 包住异步 Invocation API。assistant-ui 为"服务端事件日志是唯一来源"推荐的正是这种接法（[Claude Managed Agents 集成页](https://www.assistant-ui.com/docs/runtimes/claude-managed-agents)）。
   - AG-UI Chat 用官方的 `useAgUiRuntime`。
   - 合并成一个 runtime 受后端所限。AG-UI 端点的 run 随请求断开而取消，只接收文本，而且要求 agent 打开 `enable_agui`。主聊天依赖的离线续跑、图片输入和失败后重试，只有 Invocation API 提供（§2 第 4 行，§5）。
   - 因此"全面接入"能落地的含义是三件事：升级、两个聊天共享一层视图组件、修复已经复现的 bug。
2. **先一起升级三件套，`@ag-ui/client` 留在 0.0.x。**
   - 升到 `@assistant-ui/react` 0.15.23、`@assistant-ui/react-ag-ui` 0.0.63、`@ag-ui/client` 0.0.59 后，`tsc -b` 没有错误，85 个单测和 38 个聊天 e2e 全部通过。
   - 不要单独把 `@ag-ui/client` 升到 1.x。react-ag-ui 0.0.63 仍依赖 `^0.0.59`，混装会出现两份 client，`tsc` 报 TS2322。上游的 [assistant-ui#8517](https://github.com/assistant-ui/assistant-ui/issues/8517) 仍未关闭。
3. **AG-UI Chat 有两个已复现的 bug，升级修不掉。**
   - 有两个及以上未决 Interrupt 时，在文本回答框作答会直接抛出 `missing responses for open interrupts`，什么请求也不发。
   - 有未决 Interrupt 时在 composer 里发消息，消息被清空，但既没有发出，也没有报错。
   - 根因：AG-UI 规范、`@ag-ui/client` 和 assistant-ui 三层都要求一次 resume 覆盖全部未决 Interrupt，而 Butter 只给表单提交做了绕行（§1.3，§4.2）。
4. **`a2ui/agent.ts` 里的 resume 改写在最新版里仍然需要，升级也不会让它失效。** 要删掉它，得先让 Butter 后端接受 `status:"cancelled"`，而这需要先定下这个状态的语义（§6 P2-8）。
5. **主聊天的消息转换跟不上 assistant-ui 的 part 模型。**
   - 持久化事件里 `thought` 部分的文本被当成正文显示。
   - 用户上传的图片在历史里不显示；只有图片、没有文字的那一轮会整条消失。
   - 工具调用和结果按工具名 FIFO 配对，没有用 id。
   - assistant-ui 现成的 `reasoning`、`image` part 和 toolkit 渲染器可以直接接上（§1.3，§2）。
6. **assistant-ui 已经原生支持 A2UI，但约定与 Butter 不兼容。**
   - assistant-ui 识别 AG-UI 的 `ACTIVITY_SNAPSHOT{activityType:"a2ui-surface"}`，转换成自己的 generative UI 渲染；用户动作走 `forwardedProps.a2uiAction`。
   - Butter 用的是 `CUSTOM butter.a2ui` 事件、自有 catalog，以及绑定到 Interrupt 的表单。
   - dashboard 继续用 `@a2ui/react`（§4.6）。
7. **对第三方应用来说，AG-UI 端点是唯一能用上 assistant-ui 大部分能力的协议，但开放之前后端要先补齐。**
   - 前端工具要求 react-ag-ui ≥0.0.58。
   - 用 API token 调用的请求全部落到同一个用户 `agui-user`。
   - 会话 ID 的归属已在 #388 收紧：同一 app 内一个会话 ID 只属于一个用户，AG-UI 对别人的 `threadId` 或跨 workspace 的线程返回 403。
   - 客户端断线时 run 随之取消；用户发送的图片被丢弃。
   - 其他协议：OpenAI 兼容端点只适合无状态的纯文本问答；A2A 端点与 `@assistant-ui/react-a2a` 不兼容；Assistant Cloud 会与 Butter 的服务端会话重复存一份（§5）。

## 1. 现状盘点

### 1.1 主聊天 `front/src/features/chat`

**runtime 与后端调用**

- 用 `useExternalStoreRuntime`（`butter-runtime.ts:610-617`）。
- 消息来自两处：`useLiveSession` 轮询得到的持久化事件，以及 `WatchAgentInvocation` 推送的 `run_event` 和 `text_delta`（`butter-runtime.ts:231-400`）。
- 发送用 `SubmitAgentInvocation`，带 `request_id` 保证幂等（`:534-546`）；停止用 `CancelAgentInvocation`（`:570-590`）。
- `isRunning` 由 Butter 自己的 run 状态决定。

**用到的 assistant-ui API**

- `AssistantRuntimeProvider`。
- `ThreadPrimitive.Root/Viewport/Empty/Messages/If`。
- `MessagePrimitive.Root/Content/If`。
- `ComposerPrimitive.Root/Input/Send/Cancel`。
- `ActionBarPrimitive.Root/Copy`。
- `unstable_useComposerInput`，只用了其中的 `setText`（`aui-chat-window.tsx:535`）。
- 类型 `ThreadMessageLike` 和 `AppendMessage`。

**手写、但 assistant-ui 有现成对应物的部分**

- **事件到消息的转换**（`message-converter.ts`）：
  - 只识别 text、functionCall、functionResponse 三种 part。
  - 调用和结果按工具名 FIFO 配对（`:98-127`），不看 `functionCall.id`。
  - `thought` 和 `inlineData` 不做处理（`lib/session-events.ts:133-145`）。
  - 内容为空的事件整条丢掉（`message-converter.ts:80`）。
- **Markdown**：`react-markdown` 加 `remark-gfm`，外加十几个组件覆盖（`aui-chat-window.tsx:760-859`），没有代码高亮。
- **工具渲染**：只有一个 `tools.Fallback` 组件，其中把 `adk_request_input` 特殊处理成 Human Input 卡片（`:633-749`）。
- **图片附件**：
  - `useImageAttachments` 负责拖放、粘贴、文件选择、校验和预览。
  - composer 外面单独渲染一条缩略图。
  - `onNew` 从 ref 里取出文件，转成 `InputPart`（`butter-runtime.ts:490-517`）。

**assistant-ui 没有对应物、应保留的部分**

- 失败或停止后的提示，以及 "Restore input"（`aui-chat-window.tsx:550-603`）。
- 新聊天草稿页 `DraftComposer`。它是普通的 `<textarea>`，不经过 runtime。
- 全局侧栏的会话列表 `components/layout/nav-chat-history.tsx`：按"今天 / 7 天内 / 更早"分组，支持搜索、重命名和删除。

**两处"绕行"**

- **`identityConvert`**（`butter-runtime.ts:632-637`）。
  - 注释说它是为了绕过 core 的一个 bug，实际上它是类型契约的要求。`ExternalStoreAdapter<T>` 规定：`T` 不是 `ThreadMessage` 时，`convertMessage` 必填（[external-store-adapter.ts L75-77、L224-228@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/core/src/runtimes/external-store/external-store-adapter.ts#L224-L228)）。文档也写明它只在使用 ThreadMessage 类型时可省（[external-store](https://www.assistant-ui.com/docs/runtimes/custom/external-store)，"Not needed if using ThreadMessage type"）。
  - 如果不传，core 会把 `messages` 直接当作 `ThreadMessage` 使用（[external-store-thread-runtime-core.ts L275-305](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/core/src/runtimes/external-store/external-store-thread-runtime-core.ts#L275-L305)）。之后 `state` getter 读取 `message.metadata.unstable_state` 时就会崩溃（[base-thread-runtime-core.ts L105-114](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/core/src/runtime/base/base-thread-runtime-core.ts#L105-L114)）。
- **`RuntimeErrorBoundary`**：错误消息里含 `unstable_state` 或 `isOptimistic` 时，直接吞掉这个渲染错误（`aui-chat-window.tsx:605-631`）。

### 1.2 AG-UI Chat `front/src/features/agui-chat`

**runtime 与线程**

- 调用方式是 `useAgUiRuntime({agent, onError, adapters:{history}})`（`index.tsx:397-401`）。
- `agent` 是 `ButterAGUIAgent extends HttpAgent`（`a2ui/agent.ts`）。`threadId` 写在 HttpAgent 上。
- workspace、user、agent、thread 中任何一项变化，组件都按 `key` 整体重新挂载（`index.tsx:263-264`）。

**用到的 assistant-ui API**

这里用到的比任务描述里 grep 出的要多：

- primitives、`AssistantRuntimeProvider`、`ThreadHistoryAdapter`。
- `useAuiState`（`index.tsx:646`、`:667`）。
- `fromThreadMessageLike` 和 `ExportedMessageRepository` 类型（`history.ts:1-5`、`:171`）。
- react-ag-ui 的 `useAgUiInterrupts`、`useAgUiSubmitInterruptResponses`、`useAgUiSteerAway` 和 `useAgUiState`（`index.tsx:21-27`）。
- 渲染表：`tools.by_name.render_ui`、`tools.Fallback` 和 `data.by_name['butter.a2ui']`（`:734-743`）。

`unstable_state` 并没有作为 API 被调用。它只出现在主聊天的注释里，以及错误边界的字符串匹配中。

**手写部分**

- **线程列表**：
  - `thread-list.tsx`；
  - `threads.ts`，按隐藏的 UI Binding 过滤线程；
  - `thread-pointer.ts`，把当前线程记在 localStorage。
- **A2UI**：自己的 store，以及基于 `@a2ui/react` 和 shadcn catalog 的渲染。
- **Interrupt 相关**：表单提交桥（`index.tsx:485-510`）和文本 Interrupt 的回答框（`:771-839`）。
- **共享状态面板**（`:841-871`）。
- **与主聊天重复的组件**，这是第二套：
  - Markdown（`:763-769`）；
  - 工具视图（`:587-640`）；
  - composer（`:873-900`）。

**绕行**

- `ButterAGUIAgent` 覆盖了 `prepareRunAgentInput` 和 `requestInit`（`a2ui/agent.ts:39-66`），见 §4.2。
- `useThreadHistory` 用一个 mounted ref，防止 StrictMode 的预演卸载打断历史加载（`index.tsx:319-365`）。

**死代码**

- `src/api/agui.ts` 里的 `runAGUIAgent` 和 `applyAGUIStateDelta` 没有任何调用方。自提交 `1b2f46d`（2026-08-21，"replace hand-written AG-UI client with official assistant-ui runtime"）起就是这样。
- `src/api/chat.ts` 里的 `streamChat` 也没有调用方。
- 因此 `docs/research/ag-ui-integration.md` 中"dashboard 用手写 SSE 解析、没有 AG-UI npm 依赖"的描述已经过时。

**重复**：前端现在有三套 Markdown 渲染（主聊天、AG-UI Chat、`components/markdown-content.tsx`）、两套工具视图、两套 composer。

### 1.3 实测：现有 e2e 没有覆盖的路径

| 场景 | 结果（`pinned` 与 `latest` 相同，除非另注） | 依据 |
|---|---|---|
| AG-UI Chat：一次 run 以两个 Interrupt 结束，在第一个文本框里作答 | 没有发出任何请求。页面抛出未处理异常 `[agui] submitInterruptResponses: missing responses for open interrupts: int-2` | 见下方说明 1 |
| AG-UI Chat：有一个未决 Interrupt 时，在 composer 输入后回车 | 只有第一次请求。composer 被清空，消息没有出现，控制台也没有任何错误 | 有未决 Interrupt 时，`append` 会抛错（[AgUiThreadRuntimeCore.ts L367-376、L464-469@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L464-L469)）。这个错误没有反馈给用户：探针既没有捕获到 pageerror，也没有 console error |
| 主聊天：持久化事件包含 `thought:true` 的文本、用户的 `inlineData` 图片、一个已完成的工具调用，以及一个未答复的 `adk_request_input` | thought 文本当正文显示；用户图片不显示（页面上 0 个 `<img>`）；Human Input 卡显示 "Human Input"，而不是 "Waiting for input" | 问题在转换层（§1.1），与 assistant-ui 版本无关 |
| 把 Butter 前端工具的事件序列喂给 `RunAggregator`：`TOOL_CALL_*` 之后是 `RUN_FINISHED{outcome:success}`，没有 `TOOL_CALL_RESULT` | 0.0.56 的最终状态是 `complete`；0.0.63 是 `requires-action/tool-calls` | §4.5 |
| 0.0.58、0.0.59 和 1.0.1 三个版本的 `HttpAgent` 解析与 Butter 同形的 SSE（含 `id:` 行、CUSTOM 和 interrupt outcome），然后再发一次不带 resume 的 run | 三个版本都完整收到 11 个事件，`pendingInterrupts=[int-7]`。第二次 run 都被客户端拒绝：`Thread has 1 pending interrupt(s) not addressed by resume: int-7` | ADR-0002 规定纯文本消息会回答最早的未决 Interrupt。任何版本的 `HttpAgent` 都发不出这样的请求 |
| 1.0.1 解析 Butter 的 `RUN_ERROR`（带 `runId`） | 0.0.59 原样交付。1.0.1 交付前剥掉 `runId`，并警告 `Removed unrecognised material at '/runId'` | 1.0 schema 的 RunErrorEvent 没有 `runId` 字段，`unevaluatedProperties:false`。影响不大 |

说明 1（第一行）：

- `InterruptPrompts` 只提交一条回答（`index.tsx:787-795`）。
- `submitInterruptResponses` 却要求覆盖所有未决 Interrupt（[L518-557@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L518-L557)，[L506-539@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L506-L539)）。
- 对协商了 A2UI 的客户端，`RUN_FINISHED` 会列出线程上全部未决 Interrupt（`agui_sink.go:323-346`）。
- 推论：只要文本 Interrupt 之外还有任何一个未决 Interrupt，就会触发这个错误，带表单的 Interrupt 也算。这一点是按同一代码路径推断的，探针没有单独跑这种组合。

## 2. 能力映射

分类有四种：**已使用**、**可用未用**、**后端阻塞**、**不适用**。工作量和风险两项是仓库判断，不是核实结果：S 指不超过 1 天，M 指 2–5 天，L 指一周以上。

| 能力 | assistant-ui 提供什么 | Butter 现状 | 分类 | 工作量/风险（判断） |
|---|---|---|---|---|
| ExternalStore runtime | 消息由宿主持有，runtime 只回调 `onNew`、`onCancel` 等（[external-store](https://www.assistant-ui.com/docs/runtimes/custom/external-store)） | 主聊天 | 已使用 | — |
| AG-UI runtime | `useAgUiRuntime` + `HttpAgent`（[ag-ui overview](https://www.assistant-ui.com/docs/runtimes/ag-ui/overview)） | AG-UI Chat | 已使用 | — |
| LocalRuntime / DataStream / AssistantTransport | Local 由 runtime 自己持有状态；后两种要求后端讲 AI SDK data stream 或 assistant-transport 协议（[pick-a-runtime](https://www.assistant-ui.com/docs/runtimes/pick-a-runtime)） | 未用 | 不适用：Butter 的会话在服务端，再加一套协议没有收益 | — |
| 用一个 runtime 同时服务两个聊天 | — | 两套 | 后端阻塞。AG-UI 的 run 以请求为作用域（`agui.go:186-195`），只接收文本（`:654-682`），并要求 `enable_agui`（`:321`）；`SubmitAgentInvocation` 只对已登录用户开放（`agent_async.go:88-91`） | L/高 |
| Google ADK runtime | `useAdkRuntime` 直连 ADK REST 接口（`/run_sse` 和 `/apps/{app}/users/{user}/sessions`），原生处理 `adk_request_input`、长运行工具、thought 和图片（[google-adk api](https://www.assistant-ui.com/docs/runtimes/google-adk/api)，[hooks](https://www.assistant-ui.com/docs/runtimes/google-adk/hooks)） | 未用 | 后端阻塞：Butter 不暴露 ADK REST。ADK Go 的 `adkrest` 每个请求都自建 runner（[runtime.go L359](https://github.com/google/adk-go/blob/v2.5.0/server/adkrest/controllers/runtime.go#L359)），会绕过 Butter 的 runner | L/高；其中 `AdkEventAccumulator` 可以单独复用（§6 P3） |
| 线程列表 | `ThreadListPrimitive`；`useRemoteThreadListRuntime`，可包任意 runtime hook；AG-UI 的 `adapters.threadList`（experimental）（[threads](https://www.assistant-ui.com/docs/runtimes/concepts/threads)） | 两套手写：全局侧栏，以及 AG-UI 页内列表 | 可用未用 | M/低；收益小，分组、搜索和 Binding 过滤都还得自己保留 |
| 工具 UI | toolkit：`defineToolkit` + `Tools({toolkit})`，其中 `type:"backend"` 只负责渲染。`makeAssistantToolUI` 等已标 deprecated（[tool-ui](https://www.assistant-ui.com/docs/tools/tool-ui)，[toolkit 迁移](https://www.assistant-ui.com/docs/migrations/toolkit-tools)）；0.15.16 已有 toolkit | `MessagePrimitive.Content` 的 `tools.by_name` / `Fallback` 组件表 | 可用未用 | S/低。`type:"backend"` 的条目不会当作前端工具发给 agent（[schema-utils.ts L170-176@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/assistant-stream/src/core/tool/schema-utils.ts#L170-L176)） |
| 前端工具（在浏览器执行） | 带 `execute` 的 toolkit 条目 | 未用 | AG-UI Chat：后端已支持（`internal/aguitool`），需要 react-ag-ui ≥0.0.58（§4.5）。主聊天：后端阻塞，Invocation API 没有 tools 参数 | S/中 |
| HITL：AG-UI Interrupt | `useAgUiInterrupts`、`useAgUiSubmitInterruptResponses`、`useAgUiSteerAway`，文档标为 experimental（[runtime-options](https://www.assistant-ui.com/docs/runtimes/ag-ui/runtime-options)） | 在用，带绕行和两个 bug | 已使用 | 修 bug：S |
| HITL：工具审批 | `reason:"tool_call"` 的 Interrupt 投影成工具 part 上的 approval | — | 不适用：Butter 没有工具确认 | — |
| 生成式 UI / A2UI | `@assistant-ui/react-generative-ui`：`a2ui-surface` activity 转成 `present` 工具调用，再加 `useAgUiSendA2uiAction`（[tools/a2ui](https://www.assistant-ui.com/docs/tools/a2ui)） | `CUSTOM butter.a2ui` + `@a2ui/react` + 绑定 Interrupt 的表单 | 约定不兼容，保留现状（§4.6） | 整体替换 L/高；额外发一份只读卡片 M/中 |
| 附件 | `AttachmentAdapter`（`SimpleImage` / `SimpleText` / `Composite`）、`ComposerPrimitive.AddAttachment/Attachments/AttachmentDropzone`，粘贴自动加附件默认开启（[attachments](https://www.assistant-ui.com/docs/guides/attachments)） | 主聊天手写 `useImageAttachments`；AG-UI Chat 没有附件 | 主聊天：可用未用。AG-UI：后端阻塞，只取文本 | 主聊天 M/中；AG-UI 后端 M |
| Markdown / 代码 | `@assistant-ui/react-markdown`（`MarkdownTextPrimitive`，适合流式）、`react-streamdown`、`react-syntax-highlighter`，以及 registry 组件 `markdown-text` | 三套 `react-markdown` 实现，没有代码高亮 | 可用未用 | S/低 |
| shadcn registry | `npx assistant-ui add thread …`，按 `components.json` 的 style 选择 Radix 或 Base UI 风格（[base-ui](https://www.assistant-ui.com/docs/base-ui)，[cli](https://www.assistant-ui.com/docs/cli)） | 未用。Butter 的 style 是 `new-york`，对应 Radix | 可用未用 | M/中：默认的 Thread 自带 Edit、Reload、BranchPicker、Feedback、Speak 等按钮，Butter 后端大多不支持 |
| 分支 / 编辑 / 重新生成 | `onEdit`、`onReload`、`BranchPicker`。ExternalStore 按回调是否存在决定开关（[core L292-304@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/core/src/runtimes/external-store/external-store-thread-runtime-core.ts#L292-L304)） | 未开 | 后端阻塞：会话是线性的，ADK Go v2.5.0 没有 rewind；AG-UI 只转发最后一条消息（`agui.go:510-561`）。但 `useAgUiRuntime` 默认提供 onEdit 和 onReload（[L245-252@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/useAgUiRuntime.ts#L245-L252)） | L |
| reasoning part | 推理内容 part，以及 registry 组件 `reasoning` | 主聊天把 thought 当正文显示；AG-UI 的 sink 直接丢弃 thought | 主聊天：可用未用。AG-UI：后端阻塞，需要发出 `REASONING_*` 事件 | 前端 S；后端 S–M |
| source part（引用来源） | `source` part | 无 | 后端阻塞：不导出 grounding metadata | — |
| image / file part | `image` / `file` part | 用户图片在历史里不显示 | 可用未用 | S |
| data part | CUSTOM 事件转成 data part | `butter.a2ui` 在用 | 已使用 | — |
| Suggestions | 静态 `Suggestions(...)`，或由 runtime 推送 `suggestions`（[suggestions](https://www.assistant-ui.com/docs/guides/suggestions)） | 无 | 可用未用：缺产品数据，例如每个 agent 的起始问题 | S |
| Feedback | `FeedbackAdapter` 和 `ActionBarPrimitive.FeedbackPositive/Negative` | 无 | 后端阻塞：没有存储反馈的接口 | M |
| 语音 | `WebSpeechSynthesisAdapter` 等（[speech](https://www.assistant-ui.com/docs/guides/speech)） | 无 | 可用未用，纯前端 | S；价值低 |
| 消息排队 | `unstable_enableMessageQueue`（AG-UI） | 无 | 可用未用（unstable）。可以避免同一线程上的 409，也会在有未决 Interrupt 时把消息留在队列里，而不是丢掉（[useAgUiRuntime.ts L159-171@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/useAgUiRuntime.ts#L159-L171)） | S |
| 条件渲染 | `AuiIf` | 用的是已 deprecated 的 `ThreadPrimitive.If` 和 `MessagePrimitive.If` | 可用未用 | S |
| 写入 composer | `useAui().composer().setText()` / `addAttachment()`（[hooks/state](https://www.assistant-ui.com/docs/api-reference/hooks/state)） | `unstable_useComposerInput().setText` | 可用未用 | S |
| Assistant Cloud | 托管线程、历史和遥测 | — | 不适用（§5.5） | — |
| MCP Apps | `mcp-apps` activity，或工具结果里的 `_meta` | — | 后端阻塞：Butter 不转发 MCP 的 `_meta` | — |

## 3. 版本与升级

### 3.1 两个版本之间的变化（只列与 Butter 相关的）

**`@assistant-ui/react` 0.15.17–0.15.23**（[CHANGELOG@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react/CHANGELOG.md)）

全部是补丁版本，没有删除 Butter 用到的导出。值得一提的有：

- #6192（0.15.17）：Radix 改为从 `radix-ui/internal` 引入。
- #7981（0.15.22）：运行时状态类型改名为 `ThreadRuntimeState` 等，旧名标为 deprecated，并说明 0.16 起旧名将指向 store 状态。Butter 没有用这些类型。
- #8030（0.15.22）：附件还在上传时，消息也会先显示出来。
- #8082（0.15.23）：修复 `MessagePrimitive.If user={false}` 等写法，并迁移到 `AuiIf`。

**`@assistant-ui/core` 0.3.16–0.3.22**（[CHANGELOG@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/core/CHANGELOG.md)）

- **0.3.16 #6324**：在 ExternalStore 的 `convertMessage` 路径上，含未完成工具调用的消息改为报告 `requires-action`，以前是 `complete`。
  - 这是一处行为变化，会影响主聊天里尚未答复的 `adk_request_input`。
  - 在 e2e 和探针里都没有观察到可见差异，因为 `HumanInputView` 只判断 `running`（`aui-chat-window.tsx:646-748`）。
- **0.3.21 #8338**：在 cloud 线程列表下，ExternalStore runtime 默认把消息复制一份到 Assistant Cloud。Butter 不用 cloud，不受影响。

**`@assistant-ui/react-ag-ui` 0.0.57–0.0.63**（[CHANGELOG@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-ag-ui/CHANGELOG.md)）

- **0.0.57**：
  - #6278：history adapter 晚于首次加载出现时，也会加载。
  - #6339：流式快照不再包含空的文本 part。
- **0.0.58**：
  - **#6501**：运行成功结束时，未完成的工具调用保持待处理状态，见 §4.5。
  - #6580：resume 的转发交给上游 client，react-ag-ui 自己的 resume shim 不再需要。
  - #6608：建模了 subagent 协议。
  - #6660、#6663、#6664：遇到畸形事件或快照时报出诊断信息，并拒绝处理。
- **0.0.59**：
  - **#7021**：新增 `resumeTranscript`（`"full"` 或 `"appended"`）选项。
  - #6949：忽略已被取代的线程加载。
  - #6962：开始新的 run 时，通过 agent 取消被取代的那一次。
- **0.0.60**：
  - #7494：`addToolResult` 会保留 `modelContent`。
  - #7710、#7076：重载后，同一个 turn 的多条记录会合并回一条。
- **0.0.61**：
  - #8063：保留没有渲染器的 activity；A2UI 的重建操作存进 `artifact.a2ui`。
  - #8139：data URL 不是 base64 的附件，改按 `url` source 发送。
  - #8282：Stop 晚于完成时，消息保持"完成"状态。
  - #7989：新建线程前先清空当前线程。
- **0.0.63**：
  - #8466：编辑消息时创建兄弟分支，不再删除原来的 turn。
  - #8518：fast refresh 和 StrictMode 重放不再拆掉 runtime。
  - #8087：`ACTIVITY_DELTA` 应用到对应的 activity 上。

公开的 hooks 没有变（`hooks.ts` 两版逐字相同），只是新增了类型和选项。

**`@ag-ui/client` 0.0.59**

0.0.59 的提交范围是 `2ac1dd9c..9f39302`。它的 CHANGELOG 从 1.0.0 才开始记，所以只能按提交看，主要有：

- subagent 协议；
- 事件和消息上的 `metadata`（resume entry 也有）；
- 不再发送值为 `null` 的可选字段；
- `RUN_ERROR` 之后可以开始新的 run（f875cb0b）。

1.0 见 §3.3。

### 3.2 升级实测

| 组合 | `tsc -b` | `vitest run` | 11 个聊天 e2e 文件（38 用例） |
|---|---|---|---|
| pinned（0.15.16 / 0.0.56 / 0.0.58） | 0 错误 | 11 个文件 85 个用例通过 | 38/38 通过 |
| latest（0.15.23 / 0.0.63 / 0.0.59） | 0 错误 | 85/85 通过 | 38/38 通过 |
| agui1（0.15.23 / 0.0.63 / **1.0.1**） | 1 个错误 | 未跑 | 未跑 |

参与 e2e 的文件是 chat-aui、chat-cancellation、chat-draft、chat-retry、chat-session-resolution、chat-sidebar-delete、chat-watch、agui-chat、agui-a2ui、agui-threads 和 workflow-human-input-form。

`agui1` 的错误如下。`node_modules` 中出现两份 client：顶层是 1.0.1，`@assistant-ui/react-ag-ui/node_modules` 下是 0.0.59。

```
src/features/agui-chat/index.tsx(398,5): error TS2322: Type 'ButterAGUIAgent' is not assignable to type 'AbstractAgent'.
  … Type 'FileSource' is not assignable to type '{ value: string; type: "url"; mimeType?: string | undefined; }'.
```

### 3.3 `@ag-ui/client` 1.0

- **兼容性。** 迁移文档说明，0.x 的 agent 可以对接 1.0 的 client，1.0 的 agent 也可以对接 0.x 的 client。线上格式不变（[migrating-to-1-0.mdx L17-38@ec9f4fa](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/docs/migrating-to-1-0.mdx#L17-L38)）。
- **破坏性变化。**
  - 事件上未知的属性会被剥掉并给出警告；`RunAgentInput` 上的未知键也会被剥掉（L106-121）。
  - client 每次都会发送 `protocolVersion:"1.0"`（L204-208）。
  - 内容 part 改名为 `ContentPart`、`ImagePart` 等，并新增 `FileSource`；`ToolMessage.content` 可以是数组（L167-200）。
  - 1.0.0 的 CHANGELOG 还列了 enforcement pipeline、`MESSAGES_SNAPSHOT` 的顺序语义，以及帧缓冲上限（[CHANGELOG](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/sdks/typescript/packages/client/CHANGELOG.md)）。
- **对 Butter 后端的影响。**
  - Go SDK 的 `RunAgentInput.UnmarshalJSON` 只读取已知的键（[types.go L365-400](https://github.com/ag-ui-protocol/ag-ui/blob/6691ac50b34a/sdks/community/go/pkg/core/types/types.go#L365-L400)），所以 `protocolVersion` 没有影响。
  - Butter 发出的 `RUN_ERROR` 带有 `runId`，1.0 的 client 会把它剥掉并警告（§1.3）。
- **对 Butter 前端的影响。** 要等 react-ag-ui 迁到 1.0。#8517 自 2026-09-28 起一直开着；从讨论看，这需要维护者批准一个 react-ag-ui 的 major 版本。

### 3.4 Butter 依赖的 unstable 与 deprecated API

| API | 用在哪里 | 0.15.16 / 0.0.56 | 0.15.23 / 0.0.63 | 替代 |
|---|---|---|---|---|
| `unstable_useComposerInput` | `aui-chat-window.tsx:535` | 已导出，标注 `@deprecated Under active development and might change without notice`（[useComposerInput.ts L52-77](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react/src/unstable/useComposerInput.ts#L52-L77)） | 文件没有变化 | `useAui().composer().setText()` |
| `metadata.unstable_state` | 只出现在注释和错误边界的字符串匹配中 | 在 [stability](https://www.assistant-ui.com/docs/runtimes/concepts/stability) 页列为 unstable | 同左 | 先验证不需要后，删掉错误边界里的特殊处理 |
| `fromThreadMessageLike` | `history.ts:171` | `@deprecated This API is experimental`（[thread-message-like.ts L110-113](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/core/src/runtime/utils/thread-message-like.ts#L110-L113)） | 同左 | `ExportedMessageRepository.fromArray` / `fromBranchableArray`（[message-repository.ts L24-70@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/core/src/runtime/utils/message-repository.ts#L24-L70)） |
| `ThreadPrimitive.If` / `MessagePrimitive.If` | 两个聊天 | deprecated，建议改用 `AuiIf`（[ThreadIf.ts L34](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react/src/primitives/thread/ThreadIf.ts#L34)） | 同左 | `AuiIf` |
| AG-UI 的 interrupt hooks | AG-UI Chat | 文档标为 experimental | 同左 | 把调用集中在一处（现在基本已是这样） |
| `MessagePrimitive.Content` | 两个聊天 | 是 `Parts` 的别名 | 同左 | 可以改名为 `Parts`，不是必须 |

相关的政策：

- **弃用政策**（[deprecation-policy](https://www.assistant-ui.com/docs/migrations/deprecation-policy)）：
  - 带 `unstable_`、`experimental_` 前缀或属于 internal 的 API 可以随时删除。
  - Primitives（`AttachmentPrimitive` 除外）属于 stable，弃用前至少提前 3 个月通知。
  - Runtime API、消息类型、附件 API 属于 beta，弃用通知期不到 1 个月。
- **stability 页**建议依赖 `unstable_` 的项目锁定版本，并把调用点隔离起来。Butter 现在用的是精确版本，没有 `^`，符合这个建议。

## 4. `@assistant-ui/react-ag-ui` 适配器

### 4.1 threadId

- runtime 不管理线程 id，取 `this.agent.threadId || "main"`（[L1403@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L1403)，[L1491@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L1491)）。
- `HttpAgent` 的默认 threadId 是一个新的 UUID（0.0.58 `agent.ts` L109；[1.0.1 L236](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/sdks/typescript/packages/client/src/agent/agent.ts#L236)）。
- Butter 为每个线程新建一个 `ButterAGUIAgent`，并让组件整体重新挂载（`index.tsx:104-116`、`:263-264`），与上面的设计一致。
- `adapters.threadList` 仍是 experimental。新版修复了切换线程时的竞态（#6949、#7989），但 Butter 不需要改。

### 4.2 RUN_FINISHED 的 interrupt outcome 与 resume

**客户端的行为**

- `outcome.type==="interrupt"` 时，最后一条 assistant 消息的状态设为 `requires-action/interrupt`，interrupts 写进 `metadata.custom.agui.interrupts`（[run-aggregator.ts L151-158@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/adapter/run-aggregator.ts#L151-L158)）。
- 有未决 Interrupt 时，`append` 和 `reload` 都会抛错（[L367-414、L464-469](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L367-L414)）。
- `submitInterruptResponses` 要求每个未决 Interrupt 各有一条回答（L518-557）。
- `steerAway` 把没有给出回答的 Interrupt 一律填成 `cancelled`（[L680-740](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L680-L740)）。
- 以上三点在 0.0.63 中没有变化（[L453-539、L673-736](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L453-L539)）。

**规范与 `@ag-ui/client` 的要求**

- AG-UI 规范第 3 条：一个 resume 数组必须覆盖全部未决 Interrupt，不支持部分 resume。
- 第 4 条：线程上有未决 Interrupt 时，任何 `RunAgentInput` 都必须带 resume。收到不合规输入的 agent 应当发出 `RunError`（[interrupts.mdx L121-142@ec9f4fa](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/docs/concepts/interrupts.mdx#L121-L142)）。
- `@ag-ui/client` 在 `onInitialize` 中再检查一次覆盖（[0.0.58 L401-418](https://github.com/ag-ui-protocol/ag-ui/blob/2ac1dd9c961c35a92a4683a89b1eb248be91b660/sdks/typescript/packages/client/src/agent/agent.ts#L401-L418)，[1.0.1 L579-600](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/sdks/typescript/packages/client/src/agent/agent.ts#L579-L600)）。探针证实三个版本都会拒绝。

**Butter 服务端的行为**

- 拒绝 `cancelled`（`agui.go:466-471`）。
- 接受只回答一部分 Interrupt 的 resume（`docs/api.md` "Human-in-the-loop"）。
- 纯文本按隐式 FIFO 回答最早的 Interrupt（ADR-0002）。
- 规范第 3、4 条的要求与这里正好相反。

**`ButterAGUIAgent` 的做法与升级后的情况**

- 做法：先在 `prepareRunAgentInput` 里用 `cancelled` 占位，让 client 的覆盖检查通过；再在 `requestInit` 里只发送表单的那一条（`a2ui/agent.ts:39-66`）。
- 升级后：
  - 0.0.58 #6580 删掉了 react-ag-ui 自己的 `installResumeShim`（[L1431-1450@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L1431-L1450)），改由上游转发 resume。0.0.58、0.0.59、1.0.1 三个版本的 `prepareRunAgentInput` 都会原样转发 `parameters.resume`（[1.0.1 L546-577](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/sdks/typescript/packages/client/src/agent/agent.ts#L546-L577)）。
  - `requestInit` 仍是可以覆盖的 protected 方法，而且在 1.0.1 中调用时机晚于出站校验（[http.ts L53-64、L91-103](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/sdks/typescript/packages/client/src/agent/http.ts#L53-L103)）。
  - 因此这个覆盖在升级后仍然有效（e2e 证实），也仍然需要。

**`resumeTranscript: "appended"`（0.0.59 #7021）**

- Butter 在 resume 请求里本来就不读 `messages`：只要构造出了 FunctionResponse，`aguiInputParts` 就直接返回（`agui.go:517-553`）。
- 所以可以打开这个选项来缩小请求体，但收益很小。

**`Interrupt.reason`**

- Butter 用的是 `human_input`（`agui_sink.go:24`）。
- 规范建议：自定义 reason 用 `<framework>:<name>` 形式加命名空间；或者用核心值 `input_required`，并给出 `responseSchema`（[interrupts.mdx L168-192](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/docs/concepts/interrupts.mdx#L168-L192)）。
- assistant-ui 只对 `tool_call` 做投影，所以改名不影响 dashboard。

### 4.3 STATE_SNAPSHOT 与 STATE_DELTA

- `STATE_SNAPSHOT` 整体替换本地状态。`STATE_DELTA` 用 fast-json-patch 先校验、再应用（[L1668-1689@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L1668-L1689)）。下一次 run 把本地快照作为 `state` 发回（L1411）。
- 这与 Butter "服务端权威 + 有分歧时发更正快照" 的模型吻合（`agui.go:402-439`）。
- 第三方需要知道：用 `useAgUiSetState` 改的本地状态，会在下一次 run 被服务端快照覆盖。

### 4.4 CUSTOM

- 每个 CUSTOM 事件按到达顺序变成一个 `{type:"data", name, data:value}` part（[L216-225@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/adapter/run-aggregator.ts#L216-L225)，[L380-389@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-ag-ui/src/runtime/adapter/run-aggregator.ts#L380-L389)）。
- 文档补充了两点：data part 每次 run 都会重置；它们不会回传给 agent（[runtime-options "Custom events"](https://www.assistant-ui.com/docs/runtimes/ag-ui/runtime-options#custom-events)）。
- Butter 只用 `data.by_name['butter.a2ui']` 来确定卡片在消息中的位置。surface 的状态由订阅 `httpAgent.onCustomEvent` 的独立 store 维护（`index.tsx:303-317`）。两个版本行为一致。

### 4.5 前端工具

**发送与 Butter 端的处理**

- runtime 把 model context 里带 `execute` 的工具转成 `tools`（L1413）。`type:"backend"` 和已禁用的工具不发送（pinned 依赖的 `assistant-stream` 0.3.39 `schema-utils.ts:137`；@0.0.63 L170-176）。
- Butter 端：`tools` 变成 aguitool 的长运行工具。模型调用后，Butter 发出 `TOOL_CALL_*`，以 `outcome:success` 结束，不发 `TOOL_CALL_RESULT`（`agui_sink.go:384-389`；`docs/api.md` "Frontend tools"）。

**0.0.56 与 0.0.58 的差别**

- 0.0.56：只要 outcome 是 success，不管还有没有未完成的调用，状态都是 `complete`（[run-aggregator.ts L159-170@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/adapter/run-aggregator.ts#L159-L170)）。
- `addToolResult` 之后，只有状态为 `requires-action/tool-calls` 的消息才会续跑（[maybeCompleteAfterToolResults L936-963](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L936-L963)）。所以在 0.0.56 下，工具结果永远不会发回 Butter。
- 0.0.58 #6501 改为：只要还有未完成的调用，状态就是 `requires-action/tool-calls`（[L302-316@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-ag-ui/src/runtime/adapter/run-aggregator.ts#L302-L316)）。
- 这一差别已由探针证实（§1.3）。
- 结论：Butter 的前端工具协议要求 react-ag-ui ≥0.0.58。dashboard 没有注册前端工具，所以这个问题至今没有暴露。

**续跑与自动取消**

- 续跑请求默认发送完整的 transcript，最后几条是 tool 消息。Butter 只认末尾连续的 tool 消息，并按会话里未决的调用逐条校验（`agui.go:563-636`），两者对得上。
- `autoCancelPendingToolCalls` 默认开启。用户改发新消息时，runtime 会给未完成的调用填一个错误结果，与新消息一起发出（[runtime-options](https://www.assistant-ui.com/docs/runtimes/ag-ui/runtime-options#auto-cancelling-pending-tool-calls)）。
- 这些 tool 消息不在末尾，Butter 会忽略它们，服务端的那次调用因此仍然未决。后果没有验证，见 §7。

### 4.6 A2UI

**assistant-ui 自带的 A2UI 支持**

- 0.0.56 已经原生处理 `ACTIVITY_SNAPSHOT{activityType:"a2ui-surface"}`：把 operations 归并成 surface，生成 `toolName:"present"`、`toolCallId:"a2ui:<surfaceId>"` 的工具调用 part（[run-aggregator.ts L54、L313、L402-413@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/adapter/run-aggregator.ts#L402-L413)），再由 `@assistant-ui/react-generative-ui` 渲染。
- 按钮动作通过 `useAgUiSendA2uiAction`，以 `forwardedProps.a2uiAction.userAction` 发回（L1421-1423；[tools/a2ui](https://www.assistant-ui.com/docs/tools/a2ui)）。
- 这是 `@ag-ui/a2ui-middleware` 的约定。middleware 会把动作改写成一个合成的 `log_a2ui_event` 工具调用（[index.ts L56、L251-286、L925-976@ec9f4fa](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/middlewares/a2ui-middleware/src/index.ts#L251-L286)）。

**与 Butter 的三处差别**

- **事件形状**：Butter 用 CUSTOM，对方用 ACTIVITY_SNAPSHOT。
- **渲染器**：assistant-ui 把 A2UI 基础 catalog 映射成自己的 generative UI 词汇表，未知组件默认跳过（文档 "Component mapping"），Butter 的 `KeyValue` 和 `Status` 会因此丢失。
- **表单语义**：Butter 的表单绑定到一个 Interrupt，带 token 和 revision，由服务端校验后作为 resume 送达（`docs/api.md` "Human Input forms"）；对方只是一次普通的工具调用。

**结论**

- dashboard 保留 `@a2ui/react`。
- 如果想让现成的 assistant-ui 或 CopilotKit 客户端也能看到卡片，可以在后端为只读卡片额外发一份 `a2ui-surface` activity，只用基础 catalog 的组件。表单不做互通。

### 4.7 历史加载

- runtime 在挂载时调用一次 `adapters.history.load()`，它返回 `ExportedMessageRepository`，可以附带 `state` 和 `unstable_resume`（[__internal_load L322-365@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L322-L365)）。
- Butter 的 `append` 是空实现。文档明确说，只有后端自己持久化对话时这样做才安全（runtime-options "Loading conversation history"）。Butter 正是这种情况。
- 官方还提供 `fromAgUiMessages`：
  - 它把 AG-UI messages 转换成 assistant-ui 消息；
  - 它会从 assistant 消息的 `metadata.custom.agui.interrupts` 恢复 interrupt 状态（[conversions.ts L583-592@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-ag-ui/src/runtime/adapter/conversions.ts#L583-L592)）；
  - 如果 Butter 的历史端点在最后一条 assistant 消息上带上这个 metadata，就可以换用它。AG-UI 1.0 的 message 有开放的 `metadata` 字段。换用后 `history.ts` 只需保留放置 surface 的逻辑，收益中等。
- 新版的相关修复：#6278、#6949、#7710、#7759。#8518 用 `useReplaySafeEffect`（[useAgUiRuntime.ts L332@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-ag-ui/src/useAgUiRuntime.ts#L332)）让 StrictMode 重放不再触发 detach 时的取消。这可能让 Butter 的 mounted-ref 防护变得多余，但没有验证。

### 4.8 取消

- `core.cancel()` 先调用 `agent.abortRun()`（HttpAgent 中止 fetch），再中止本地 run，聚合器把消息标为 `incomplete/cancelled`（[L416-428@0.0.56](https://github.com/assistant-ui/assistant-ui/blob/75e3ef71beb5dc99f6fc624624d3d61b307c8599/packages/react-ag-ui/src/runtime/AgUiThreadRuntimeCore.ts#L416-L428)）。
- 组件卸载时，`detachRuntime()` 同样会取消（L209-214）。
- Butter 端：连接断开即取消 run，并释放租约（`agui.go:186-195`）。AG-UI Chat 在删除当前线程、切换线程、换 agent、新建线程之前都会主动调用 `abortRun()`（`index.tsx:220`、`:427`、`:445`、`:449`）。
- 新版修复了 Stop 与完成之间的竞态（#8282）。main 上还有尚未发布的 #8325、#8326。
- 对比主聊天：主聊天的 run 不会因离开页面而取消，AG-UI 路径没有这种能力。

### 4.9 Butter 的绕行和手写部分在升级后的去留

| 绕行或手写部分 | 升级到 0.15.23 / 0.0.63 后 | 什么时候可以删 |
|---|---|---|
| `ButterAGUIAgent` 的 resume 改写 | 仍然需要，仍然有效（e2e 全部通过） | Butter 后端接受 `cancelled` 之后（§6 P2-8） |
| `InterruptPrompts` 只提交一条回答 | 仍然是 bug | 改走 `resumeNextRunWith`（§6 P0-1） |
| `history.ts` 手工拼 repository，并用 `fromThreadMessageLike` | 仍然可用 | 可改用 `fromAgUiMessages` 和 `ExportedMessageRepository.fromArray` |
| `useThreadHistory` 的 mounted ref | 无害，可能已多余（#8518） | 验证之后 |
| `src/api/agui.ts` 手写的 SSE 客户端 | 死代码 | 现在就可以删 |
| 主聊天的 `identityConvert` | 类型契约要求，保留；改正注释 | — |
| 主聊天 `RuntimeErrorBoundary` 的特殊处理 | 可能已多余 | 验证之后 |
| `unstable_useComposerInput` | 仍然导出，代码没有变化 | 换成 `useAui().composer().setText()` |

## 5. Butter 作为第三方 assistant-ui 应用的后端

| Butter 协议 | 可接的 assistant-ui runtime | 线程持久化 | 工具调用 | Interrupt / 人工输入 | 共享状态 | 附件 | A2UI |
|---|---|---|---|---|---|---|---|
| AG-UI `POST /api/agui/:agent_id` | `useAgUiRuntime` + `HttpAgent`（react-ag-ui ≥0.0.58） | 服务端会话 `agui-{threadId}`。历史需要自写 history adapter 调 `GET …/threads/:id/messages`；线程列表要走 ConnectRPC `ListSessions(app_name="agui")` | 后端工具会显示；前端工具可用 | 支持按 id 寻址的 resume。客户端强制一次覆盖全部未决 Interrupt，Butter 却拒绝 `cancelled`。表单只能用 Butter 的扩展 | 可用，服务端权威 | 被丢弃，只取文本 | 只有 Butter 自己的扩展，没有现成的渲染器 |
| OpenAI 兼容 `/api/v1/chat/completions` | LocalRuntime + 自写的 ChatModelAdapter；或在服务端经 OpenAI 兼容 provider 接 AI SDK runtime | 无。每个请求都新建会话，整段历史压成一段文本 | 不暴露，也不接受 `tools` | 无法恢复 | 无 | 只取文本 | 无 |
| ConnectRPC：`StreamAgent`、`InvokeAgent`、`SessionService`；`SubmitAgentInvocation` 只对登录用户开放，API token 因而也用不上 `WatchAgentInvocation` | ExternalStoreRuntime + 自写的事件折叠（同 dashboard）；线程列表用 `useRemoteThreadListRuntime` | 有。API token 可以为任意 `user_id` 建立、列出和删除会话 | `run_event` 里有完整的 `content_json` | 只能按 FIFO 用文本回答 | 无 | `InputPart` 图片 | 无 |
| A2A `/a2a/:agent_ref` | `@assistant-ui/react-a2a` 要求 A2A v1.0 | — | — | — | — | — | 不兼容 |

### 5.1 AG-UI：可用，但开放前后端要先补齐

**能端到端打通的部分**

- 流式文本和后端工具调用的展示。
- 前端工具，要求 react-ag-ui ≥0.0.58。
- 共享状态。
- 按 id 回答 Interrupt：不过必须照抄 `ButterAGUIAgent` 的覆盖写法，因为 assistant-ui 自带的 `steerAway` 一定会带上 `cancelled`，Butter 会返回 400。
- 用历史端点恢复线程。

**Butter 端缺的东西**

1. **API token 调用方的主体。**
   - API token 认证只往上下文里放 token id，不放用户（`auth.go:152-195`），于是 `aguiUserID` 回落到常量 `agui-user`（`agui.go:687-697`）。
   - 结果：同一个 token、不同的 token，乃至不同 workspace 的调用方，在 ADK 里都是同一个用户，只能靠 threadId 区分。
   - 第三方的 BFF 没法按终端用户隔离。ConnectRPC 路径可以，因为它允许指定 `user_id`。
2. **会话 ID 的归属（已在 #388 修复）。**
   - 会话事件按 `(app_name, session_id)` 存储，Telegram 的 Destination 会话靠它让群里成员共享一段对话。
   - #388 起，同一 app 内一个会话 ID 只属于一个用户（`internal/runtime/sessionshare`），只有 Telegram 的对话路径可以共享。
   - AG-UI 对别人持有的 `threadId`，或调用方自己在另一个 workspace 的线程，在打开流之前返回 403。第三方应为每个新线程生成新的 UUID。
3. **断线即取消**（`agui.go:186-195`），没有重连机制。`@ag-ui/client` 有 `connectAgent()` 钩子，但 `HttpAgent` 默认没有实现它（[agent.ts L410-411@ec9f4fa](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/sdks/typescript/packages/client/src/agent/agent.ts#L410-L411)）。
4. **图片。**
   - Go SDK 能解析 `image`、`audio`、`video`、`document` 部件（[types.go L51-62](https://github.com/ag-ui-protocol/ag-ui/blob/6691ac50b34a/sdks/community/go/pkg/core/types/types.go#L51-L62)），但 `latestAGUIUserText` 只取文本（`agui.go:654-682`）。
   - react-ag-ui 发送的图片附件形如 `{type:"image", source:{type:"data",…}}`，会被静默丢弃。只有图片的消息直接得到 400。
5. **规范偏差。** Butter 拒绝 `cancelled`，接受部分 resume，并允许文本隐式回答 Interrupt，详见 §4.2。另外：
   - Interrupt 前没有发出 `MESSAGES_SNAPSHOT`。规范要求在 interrupt 之前发出恢复所需的状态（[interrupts.mdx L144-154](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/docs/concepts/interrupts.mdx#L144-L154)），不过 assistant-ui 并不依赖它。
   - 不发出 `REASONING_*` 事件。
6. **编辑与重新生成。**
   - `useAgUiRuntime` 默认提供 onEdit 和 onReload（0.0.63 的 L266-273），registry 里的 Thread 组件也默认带 Edit、Reload 和 BranchPicker。
   - 而 Butter 只会追加最后一条用户消息（`agui.go:510-561`），服务端历史会与客户端分支分叉。第三方应该隐藏这些按钮。
7. **接入门槛与浏览器直连。**
   - 每个 agent 都要单独打开 `enable_agui`。
   - CORS 会反射任意 Origin，允许的请求头有 `Authorization`、`Content-Type`、`X-Workspace-ID` 和 `Connect-Protocol-Version`（`auth.go:218-228`），所以浏览器可以直连。
   - 但直连就等于把 Butter token 放进浏览器，应该经过 BFF。

### 5.2 OpenAI 兼容端点：只适合无状态的纯文本

- 请求只读取 `model`、`messages`、`stream` 三个字段（`openai.go:62-66`），非文本 part 被忽略（`:76-106`）。
- 每个请求都新建会话 `openai-<uuid>`，用户固定为 `openai-user`（`:195-205`）。整段历史被压成形如 "[User] …" 的一段文本（`:341-360`）。
- 流式出错时返回 `finish_reason:"error"`（`:300`），这不是 OpenAI 的标准值；`usage` 恒为 0（`:234`）。
- 在 assistant-ui 这边，可以用 LocalRuntime 加一个调用此端点的 ChatModelAdapter，由客户端自己持有历史。
- Workflow 遇到 Human Input 暂停时，问题会以文本形式返回，但下一个请求已经是新会话，所以无法恢复。

### 5.3 ConnectRPC：数据最全，但没有现成的 runtime

- **API token 的限制。**
  - `SubmitAgentInvocation` 要求已登录用户（`agent_async.go:88-91`）。而且依赖这条异步路径时，部署不能超过一个副本（`:93-99`；`docs/api.md` "Async failure…"）。
  - 所以第三方用 API token 时，只能用三类接口：
    - 同步的 `StreamAgent`，请求断开时随之取消；
    - `InvokeAgent`；
    - `SessionService`：API token 可以指定本 workspace 内任意 `user_id`（`docs/api.md` "Session access"）。
- **接入方式。** 可以照 assistant-ui 的 Claude Managed Agents 集成来做：
  - 把事件日志折叠成消息，交给 ExternalStore runtime；
  - 会话列表接到 `useRemoteThreadListRuntime`。
- **缺的东西。**
  - 公开的客户端 SDK 和事件转换器：dashboard 的转换器是内部代码。
  - 面向 API token 或服务主体的异步 Invocation。
  - 在 ConnectRPC 上按 id 回答 Interrupt：现在只有 AG-UI 支持。

### 5.4 A2A：不兼容

- **Butter 端**：只实现了 JSON-RPC `tasks/send`（`a2a.go:99-102`）。
  - agent card 在 `/.well-known/agent.json`，并且声明 `streaming:false`（`:47`、`:182-192`）。
  - 只取第一个文本 part（`:194-201`）。
  - task id 直接当作 session id，用户固定为 `a2a`（`:111-125`）。
- **`@assistant-ui/react-a2a` 一侧**：要求 A2A v1.0 的 HTTP+JSON 绑定，也就是 `POST /message:stream` 或 `/message:send`，以及 `/.well-known/agent-card.json`（[client-and-hooks](https://www.assistant-ui.com/docs/runtimes/a2a/client-and-hooks)）。
- **可能的出路**：ADK Go v2.5.0 带有 `server/adka2a/v2`（基于 a2a-go/v2 v2.5.0），也许能用来升级，本文没有评估。
- **Google ADK runtime 也不可行。**
  - `@assistant-ui/react-google-adk` 直连 ADK REST（[AdkClient.ts L86-87@0.0.63](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-google-adk/src/AdkClient.ts#L86-L87)）。
  - 它回答 `adk_request_input` 时把答案包成 `{result}`（[hooks.ts L95-105](https://github.com/assistant-ui/assistant-ui/blob/3542d602272a62eddeb8989befc910841c267022/packages/react-google-adk/src/hooks.ts#L95-L105)）。ADK Go v2.5.0 会拆开 `result`、`response`、`payload` 三种包装（[persistence.go L504-525](https://github.com/google/adk-go/blob/v2.5.0/workflow/persistence.go#L504-L525)），所以在语义上与 Butter 的 Human Input 兼容。
  - 但 Butter 没有暴露 ADK REST，见 §2。

### 5.5 Assistant Cloud

- Assistant Cloud 是托管服务，按活跃用户计价：Free 每月 200 个，超过即停；Pro 每月 50 美元，含 500 个（[pricing](https://www.assistant-ui.com/docs/cloud/pricing)）。
- AG-UI 通过 `useCloudThreadListRuntime` 接入 Cloud，后者提供线程、消息历史、附件和反馈；被包装的 runtime 自己提供的 adapter 会被保留（[custom-thread-list](https://www.assistant-ui.com/docs/cloud/custom-thread-list)）。
- 对 Butter 来说，这会在服务端会话之外再存一份对话（ADR-0002 规定会话事件是唯一来源），还会把对话内容发给第三方 SaaS。dashboard 不应接入。
- 第三方如果为了分析而接入，应保留 Butter 的 history adapter 作为对话的来源。

## 6. 建议（仓库判断，不是核实结果）

### P0：bug 与隔离

1. **fix(agui-chat): 有多个未决 Interrupt 时，文本回答发不出去**
   - 做法：文本回答改走与表单相同的 `resumeNextRunWith` + `steerAway` 路径。
   - 验收：新增 e2e，两个未决 Interrupt 时回答其中一个，请求里只带一条 resume entry，并且 run 继续执行。
   - 估计：S。
2. **fix(agui-chat): 有未决 Interrupt 时，composer 发的消息被静默丢弃**
   - 二选一：
     - 禁用 composer，并提示用户先回答问题；
     - 把文本当作最早那个 Interrupt 的回答发出，与 ADR-0002 一致。
   - 也可以考虑打开 `unstable_enableMessageQueue`，让消息至少保留在队列里。
   - 验收：e2e 覆盖这种情况。
   - 估计：S。
3. **会话 ID 的归属**：已由 #388 完成，同时修复 #383（`StreamAgent` / `InvokeAgent` 改走会话访问规则）。

### P1：升级与前端整理

4. **chore(front): 一次性升级三件套**
   - 升到 `@assistant-ui/react` 0.15.23、`@assistant-ui/react-ag-ui` 0.0.63、`@ag-ui/client` 0.0.59，保持精确版本。
   - 用本文跑过的 11 个 e2e 文件作门禁。
   - 写下规则：`@ag-ui/client` 跟随 react-ag-ui 的依赖版本，在 #8517 解决前不升到 1.x。
   - 估计：S。
5. **feat(chat): 主聊天消息转换补齐 part 类型**
   - `thought` 转成 `reasoning` part，折叠显示或隐藏。
   - 用户的 `inlineData` 转成 `image` part。
   - 工具调用按 `functionCall.id` 配对。
   - 未答复的 `adk_request_input` 显示为等待状态。
   - 给转换器补单测。
   - 估计：S–M。
6. **refactor(front): 两个聊天共用一层 assistant-ui 视图组件**
   - toolkit 的 `type:"backend"` 渲染器，覆盖 `adk_request_input`、`render_ui` 和通用回退。
   - 一个 Markdown 组件，用 `@assistant-ui/react-markdown` 或 registry 的 `markdown-text`。
   - 改用 `AuiIf` 和 `useAui().composer().setText()`。
   - 删除 `src/api/agui.ts` 的死代码和 `streamChat`。
   - 改正 `identityConvert` 的注释；验证之后去掉错误边界的特殊处理。
   - 估计：M。
7. **feat(chat): 主聊天改用 `AttachmentAdapter`**
   - 替换 `useImageAttachments`：限额在 `add()` 中校验；`onNew` 把附件转成 `InputPart`；Restore input 改用 `composer.addAttachment`。
   - 估计：M。风险中等，需要回归重试和恢复两条流程。

### P2：需要决策的事项，以及给第三方开放 AG-UI

8. **ADR：AG-UI resume 中 `cancelled` 的语义**
   - **A. 维持拒绝。** 第三方必须照抄 `ButterAGUIAgent`。
   - **B. 接受，含义是"本次不回答，保持未决"，并在 `RUN_FINISHED` 中再次报告。** 这偏离了规范对 `cancelled` 的定义，但现成的 assistant-ui 立即可用，`a2ui/agent.ts` 的大部分也可以删掉。
   - **C. 实现真正的放弃。** Human Input 节点收到取消信号，走一条专门的边。符合规范，但需要做 Workflow 设计。
   - 倾向：以 C 为目标，短期维持 A。
9. **feat(agui): 区分 API token 调用方的主体**
   - 例如：只对 API token 开放 `forwardedProps.butterSubject`，并写进会话的 UI Binding。
   - 至少用 `token:<id>` 代替常量 `agui-user`。
   - 估计：M。
10. **feat(agui): 接收 `image` 部件**
    - 只接受 data source，复用 `InputPart` 的 MIME 白名单和大小限额；url source 暂时拒绝。
    - 估计：M。
11. **docs(api): 新增"用 assistant-ui 接入 Butter"一节**，内容包括：
    - 要求 react-ag-ui ≥0.0.58；
    - `ButterAGUIAgent` 式的覆盖写法；
    - 隐藏 Edit、Reload 和 BranchPicker；
    - 线程列表用 `ListSessions(app_name="agui")`；
    - 用历史端点实现 history adapter；
    - 由 BFF 持有 token。
    - 估计：S。

### P3：值得评估，但先不排期

12. **用 `@assistant-ui/react-google-adk` 的 `AdkEventAccumulator` 作主聊天的事件折叠器。**
    - 探针：把 Butter 持久化的 `content_json` 直接喂进去，`inlineData` 得到 `image` part，`thought` 得到 `reasoning` part。
    - 限制一：工具结果是单独的 `tool` 消息，需要经 `useExternalMessageConverter` 合并。
    - 限制二：Butter 的 `SessionEvent` 不带 `long_running_tool_ids` 和 `actions`，不扩展 proto 就识别不出人工输入。
13. **让 AG-UI 发出 `REASONING_*` 事件**，与第 5 条的 reasoning 显示配套。
14. **为只读卡片额外发一份 `a2ui-surface` activity**，供现成客户端渲染。
15. **长期**：让 AG-UI 的 run 跑在 asyncrun 上并支持重连，之后再考虑把主聊天迁到 AG-UI runtime。这是把两个聊天合成一个 runtime 的唯一途径。
16. **A2A v1.0**：评估 `adka2a/v2`。

### 不建议

- 接入 Assistant Cloud。
- 用 `react-generative-ui` 替换 `@a2ui/react`。
- 迁到 `ThreadListPrimitive`：收益小。
- 暴露 ADK REST 接口。
- 单独把 `@ag-ui/client` 升到 1.x。

## 7. 未核实或互相矛盾之处

**未核实**

- 所有 e2e 都用仓库的 fixture 模拟后端，没有跑真实的 Butter。
  - 前端工具的完整链路没有测过：真实模型调用前端工具，assistant-ui 执行它，再续跑。只在聚合器这一层验证了状态。
  - `autoCancelPendingToolCalls` 发出的 tool 消息不在末尾，被 Butter 忽略后，服务端那次长运行调用会怎样，没有验证。例如 ADK 会不会补一个响应，OpenAI 类 provider 会不会报"缺少 tool 结果"。
- 在 Butter 当前的模型配置下，真实的 Gemini 或 OpenAI 是否会产生 `thought` 部件，没有在真实模型上验证。探针只证明了：只要存在 thought，主聊天就会把它当正文显示。后端有多处显式过滤 `Thought`（`internal/runtime/streamorch/classify.go:37`、`internal/runtime/runner/runner.go:1299`、`internal/handler/http/agui_history.go:183`），可见这类部件确实会出现。
- "文本 Interrupt 与带表单的 Interrupt 同时未决"这种组合，是按同一条代码路径推断的，探针只跑了两个都是文本 Interrupt 的情况。
- Assistant Cloud 能否自托管：threads 概念页提到了 self-host options，但 cloud 文档的各页里没有找到对应说明。
- react-ag-ui 何时迁到 `@ag-ui/client` 1.0 不确定。#8517 仍未关闭，而且需要维护者批准一个 major 版本。

**文档与源码不一致**

- assistant-ui 的 Google ADK hooks 文档说 `adk_request_input` "is emitted only by ADK Python 2.0+ … ADK JS has no equivalent"。但 ADK Go v2.5.0 同样会发出它（[request_input.go L36](https://github.com/google/adk-go/blob/v2.5.0/workflow/request_input.go#L36)）。
- AG-UI runtime 文档的功能表把 Interrupts 标为 "Experimental (`unstable_*` API)"。可是推荐使用的 hooks（`useAgUiInterrupts` 等）并没有 `unstable_` 前缀；带前缀的是 runtime 上两个已经 deprecated 的方法。
- assistant-ui 文档站从 main 构建，个别描述对应的可能是尚未发布的版本，例如 #8325 和 #8326。

**Butter 自己的文档或注释与代码不一致**

- `butter-runtime.ts` 的注释说 `identityConvert` 是为了绕过 core 的 bug。但 assistant-ui 的类型和文档都表明，对 `ThreadMessageLike` 来说 `convertMessage` 本来就是必填的。
- `docs/research/ag-ui-integration.md` 说 dashboard 用手写的 SSE 客户端、没有 AG-UI npm 依赖。这已被 `1b2f46d` 取代。

**AG-UI 规范与 Butter 的行为相反**

- 规范第 3、4 条要求一次 resume 覆盖全部未决 Interrupt；线程有未决 Interrupt 时，新输入必须带 resume，否则 agent 应发出 RunError。Butter 却接受部分 resume，也允许用纯文本隐式回答。

## References

- assistant-ui
  - 文档（2026-10-04 抓取）：
    - 选型与稳定性：[pick-a-runtime](https://www.assistant-ui.com/docs/runtimes/pick-a-runtime)、[stability](https://www.assistant-ui.com/docs/runtimes/concepts/stability)、[deprecation-policy](https://www.assistant-ui.com/docs/migrations/deprecation-policy)
    - runtime 与线程：[ag-ui runtime-options](https://www.assistant-ui.com/docs/runtimes/ag-ui/runtime-options)、[external-store](https://www.assistant-ui.com/docs/runtimes/custom/external-store)、[threads](https://www.assistant-ui.com/docs/runtimes/concepts/threads)
    - 工具与 A2UI：[tools/a2ui](https://www.assistant-ui.com/docs/tools/a2ui)、[tool-ui](https://www.assistant-ui.com/docs/tools/tool-ui)、[toolkit 迁移](https://www.assistant-ui.com/docs/migrations/toolkit-tools)
    - UI 与组件：[attachments](https://www.assistant-ui.com/docs/guides/attachments)、[base-ui](https://www.assistant-ui.com/docs/base-ui)、[cli](https://www.assistant-ui.com/docs/cli)、[hooks/state](https://www.assistant-ui.com/docs/api-reference/hooks/state)
    - 其他 runtime：[google-adk api](https://www.assistant-ui.com/docs/runtimes/google-adk/api)、[google-adk hooks](https://www.assistant-ui.com/docs/runtimes/google-adk/hooks)、[a2a client-and-hooks](https://www.assistant-ui.com/docs/runtimes/a2a/client-and-hooks)、[claude-managed-agents](https://www.assistant-ui.com/docs/runtimes/claude-managed-agents)
    - Cloud：[cloud](https://www.assistant-ui.com/docs/cloud)、[cloud/custom-thread-list](https://www.assistant-ui.com/docs/cloud/custom-thread-list)、[cloud/pricing](https://www.assistant-ui.com/docs/cloud/pricing)
    - 索引：<https://www.assistant-ui.com/llms.txt>
  - 源码：[`75e3ef7`](https://github.com/assistant-ui/assistant-ui/tree/75e3ef71beb5dc99f6fc624624d3d61b307c8599)（react 0.15.16 / react-ag-ui 0.0.56）、[`3542d60`](https://github.com/assistant-ui/assistant-ui/tree/3542d602272a62eddeb8989befc910841c267022)（react 0.15.23 / react-ag-ui 0.0.63）
  - issue：[#8517](https://github.com/assistant-ui/assistant-ui/issues/8517)
- AG-UI
  - 协议文档：[interrupts](https://docs.ag-ui.com/concepts/interrupts)（[源文件@ec9f4fa](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/docs/concepts/interrupts.mdx)）、[migrating-to-1-0](https://github.com/ag-ui-protocol/ag-ui/blob/ec9f4fa68b950d0b440087cd94530c0815b81115/docs/migrating-to-1-0.mdx)
  - TS client：[1.0.1](https://github.com/ag-ui-protocol/ag-ui/tree/ec9f4fa68b950d0b440087cd94530c0815b81115/sdks/typescript/packages/client)、[0.0.58 发布提交](https://github.com/ag-ui-protocol/ag-ui/tree/2ac1dd9c961c35a92a4683a89b1eb248be91b660/sdks/typescript/packages/client)
  - [a2ui-middleware@ec9f4fa](https://github.com/ag-ui-protocol/ag-ui/tree/ec9f4fa68b950d0b440087cd94530c0815b81115/middlewares/a2ui-middleware)
- ADK Go v2.5.0：[workflow](https://github.com/google/adk-go/tree/v2.5.0/workflow)、[server/adkrest](https://github.com/google/adk-go/tree/v2.5.0/server/adkrest)、[server/adka2a/v2](https://github.com/google/adk-go/tree/v2.5.0/server/adka2a/v2)
- Butter
  - 前端：`front/src/features/chat/*`、`front/src/features/agui-chat/*`、`front/src/api/{agui,chat}.ts`、`front/src/lib/session-events.ts`
  - 后端：`internal/handler/http/{agui,agui_sink,agui_a2ui,agui_history,openai,a2a,auth}.go`、`internal/aguitool/aguitool.go`、`internal/application/agent_async.go`、`internal/runtime/session/mongo/mongo.go`
  - 文档：`docs/api.md`、`docs/research/ag-ui-integration.md`、`docs/adr/0002-interrupt-state-derived-from-session-events.md`、`docs/adr/0014-a2ui-surfaces-over-agui.md`
