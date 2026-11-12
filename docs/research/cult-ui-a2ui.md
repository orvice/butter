# Cult UI 与 Butter A2UI 扩展调研

调研日期：2026-11-12。问题是 Cult UI 能否作为 Butter 的 A2UI 组件来源，以及怎样扩展而不破坏现有协议、安全边界和恢复语义。

## 结论

Cult UI 值得作为 **React/shadcn 的视觉实现来源**，不适合作为 Butter 的 A2UI renderer、catalog 或 wire protocol。它是一个 MIT licensed 的 shadcn registry：每个条目把 TSX 源码复制进应用，组件的 props、事件和本地状态由应用自己负责；条目没有 A2UI component schema、catalog ID、surface store 或 A2UI 数据绑定。

因此建议：

1. 保留 `@a2ui/react` 0.12.0、`@a2ui/web_core` 0.12.0、`butter-basic-v1` 和现有 `butter.a2ui` 事件契约。
2. 先用一个纯展示组件做隔离 POC。优先候选是 `fluid-ai-workloads`（静态 SVG）或 `analytics-chart`（数据点和悬浮提示）；把 Cult UI 源码当作本地实现，外面包一层 Butter A2UI component implementation。
3. 新组件必须同时加入 Go 侧白名单、明确的属性 schema、模型可用性策略和前端 catalog；不得因为安装 Cult UI 就把整个 registry 或任意 React props 暴露给模型。
4. `PromptComposer`、`PromptLibrary`、`AI Instructions`、`ChoicePoll` 和带 commit/action 的卡片属于普通 dashboard/chat UI 或服务端表单候选，不应直接注册为模型可生成的 Result Card。

## Cult UI 实际提供什么

- 官网和仓库描述为 150+ 个 animated shadcn/ui 组件，目标运行时是 React + Tailwind CSS v4；多数动画使用 `motion`。
- 安装方式是 `npx shadcn@latest add @cult-ui/<name>`。CLI 将源码写入 `components/ui`，并按 registry 条目安装依赖；更新需要显式 `--overwrite` 后人工审 diff。它不是一个带稳定组件 ABI 的 npm runtime 包。
- 组件依赖应用已有的 `cn` helper、`@/` alias、shadcn theme tokens。少数条目使用 Next.js 专用的 `next/image` / `next/link`，Vite 项目需要手工替换。
- Cult UI 还提供 shadcn MCP server，帮助编码 agent 搜索和安装 registry 条目。MCP 只解决开发期源码发现，不会让运行时消息变成 A2UI。
- 仓库与 registry 页面声明 MIT；采用前仍应锁定源码 commit、审查每个条目的依赖和许可证，并避免把远端 registry 当作生产构建的隐式输入。

## 与 Butter 当前 A2UI 的边界

Butter 当前前端是 React 19 + Vite + Tailwind 4，固定 `@a2ui/react` / `@a2ui/web_core` 0.12.0。`front/src/features/agui-chat/a2ui/catalog.tsx` 只注册显式实现，`internal/a2ui/catalog.go` 对模型消息执行独立白名单和属性校验。

ADR-0014 定义的 Result Card 是只读的：模型不能收集输入、定义 action 或上传组件；Text 还必须保持纯文本。Human Input Form 由服务端构造，负责 pending/submitting/answered、字段错误、token/revision 和 resume。Cult UI 的普通 React API（`onChange`、`onCommit`、`onValueChange`、本地 `useState`、任意 `className`/style）不能直接映射到这个契约。

可以把两者的关系分成三层：

| 用法 | 判断 | 原因 |
| --- | --- | --- |
| 直接把 Cult UI registry 当 A2UI catalog | 不可行 | registry 条目没有 A2UI schema、组件 ID、数据绑定或 surface 生命周期 |
| 复制一个 Cult UI 视觉实现，在 A2UI wrapper 中暴露受限 props | 可行 | wrapper 可把动态值限制为已校验的字段，把展示层和协议层分开 |
| 在普通 Dashboard/Chat 页面使用 Cult UI | 可行 | 组件保留自己的 React state 和回调，不进入模型生成的 surface |
| 用 Cult UI 替换 `@a2ui/react` renderer 或 `butter.a2ui` 协议 | 不建议 | 会重新实现 revision/seq、删除、快照恢复、fallback、表单 resume 和隔离逻辑 |

## 候选组件

| Cult UI 条目 | 适合作为 A2UI Result Card | 需要注意 |
| --- | --- | --- |
| `fluid-ai-workloads` | 高 | 主要是 SVG 展示；只暴露有限的 `width`/`height` 和固定文案，避免把 SVG 属性任意透传 |
| `analytics-chart` | 中高 | 适合只读数据图；需要限制点数、数值范围和 label 长度，并把 hover tooltip 当客户端展示状态 |
| `code-block` | 中 | 展示部分可复用；应先去掉或由宿主控制 copy、tab 切换等 action，代码长度和语言必须受限 |
| `halo-progress` | 中 | 适合状态展示，但 registry 条目依赖 Base UI/Motion；先核对动画、无障碍和 bundle 成本 |
| `agent-suggest-card-stack` | 低 | 有进度/错误状态和 commit button；需要拆出无 action 的只读变体才能进入 Result Card |
| `choice-poll` | 低（Result Card） | `onValueChange` 和键盘选择是输入语义，应复用 Butter server-built Form，而不是让模型生成 |
| `prompt-composer` | 不属于 A2UI | 它是带 attach/send/loading 回调的聊天输入；可作为 dashboard composer 的视觉候选 |
| `prompt-library` / `ai-instructions` | 不属于 A2UI | 带搜索、创建、保存和选择状态，适合 agent 设置页或聊天工具栏 |

`analytics-chart` 的 registry 条目本身没有额外 npm dependency；`fluid-ai-workloads` 也没有额外 dependency。相较之下，`prompt-composer` 需要 `@base-ui/react`、`@paper-design/shaders-react`、`lucide-react`、`motion`，`choice-poll` 需要 Radix controllable state、`lucide-react` 和 CVA，不能把“复制一个文件”当成零成本集成。

## 推荐的扩展步骤

以 `AnalyticsChart` 为例，第一版只允许服务端生成一个只读图表：

1. 固定 Cult UI 源码 commit，复制实现到 Butter 的前端源码，并将 `@/lib/utils`、theme tokens 和 Vite/Next.js 差异处理成可审查的本地代码。
2. 在 Go catalog 增加一个明确的 `AnalyticsChart` spec：有限长度的数据点数组、有限的日期/label/value 类型和上限；拒绝 URL、HTML、任意 SVG/JS、事件名和 style/className。
3. 在 React catalog 中用 A2UI 的动态值解析结果填充该 wrapper；图表 hover、焦点和 tooltip 只能是客户端临时状态，不得产生 AG-UI action 或写回 session。
4. 为兼容旧客户端，优先新建版本化 catalog ID（例如 `butter-basic-v2`）并保留 v1；只有声明 v2 的客户端才能收到新组件。不能让仍声明 v1 的旧前端收到未知组件。
5. 增加 Go 校验/拒绝测试、catalog 注册测试、流式 create/update/delete、快照恢复和 fallback 测试，再补桌面/移动、明暗主题、长 label 和大量数据点的浏览器测试。
6. 对表单控件继续走现有 server-built Form 路径；如果将来需要 Cult UI 的视觉样式，只复用其内部 DOM/CSS，保留 Butter 的字段规则、错误显示、提交锁和 token/revision 校验。

在没有明确表格/图表需求前，不建议引入整个 Cult UI registry、替换 A2UI renderer 或一次性开放所有组件。现有 Butter 已有 `recharts`，图表 POC 应先比较薄 wrapper 的维护成本与直接复用现有图表实现，再决定是否保留 Cult UI 的视觉代码。

## 来源

- Cult UI 官网、安装与 MCP 文档：<https://www.cult-ui.com/docs>, <https://www.cult-ui.com/docs/installation.md>, <https://www.cult-ui.com/docs/mcp-server.md>
- Cult UI 仓库与 README（MIT、source-copy registry）：<https://github.com/nolly-studio/cult-ui>
- 条目文档与 registry JSON：<https://www.cult-ui.com/docs/components/analytics-chart.md>, <https://www.cult-ui.com/docs/components/fluid-ai-workloads.md>, <https://www.cult-ui.com/r/analytics-chart.json>, <https://www.cult-ui.com/r/prompt-composer.json>
- A2UI 上游架构与 catalog 原则：<https://github.com/google/A2UI/blob/main/README.md>
- Butter 当前实现：`front/package.json`、`front/components.json`、`front/src/features/agui-chat/a2ui/catalog.tsx`、`internal/a2ui/catalog.go`、`docs/adr/0014-a2ui-surfaces-over-agui.md`、`docs/research/a2ui-component-libraries.md`
