# A2UI React 组件库调研：哪些可以替代自写组件？

调研日期：2026-10-24。目标是减少 Butter A2UI 组件实现和维护，而不是更换聊天框架。依据上游 README、源码、npm registry 发布元数据，以及 Butter 已安装的包；没有仅凭项目名称或宣传中的 “A2UI support” 判断兼容性。

## 结论

1. **有现成组件可用，首选其实已安装：`@a2ui/react/v0_9` 的 18 个官方基础组件。** Butter 目前只用了它的渲染框架和组件工厂，自写了显示层；不必继续把所有标准组件都自己实现。可以按需导入官方 `Card`、`Row`、`Column`、`Divider` 等，组合成现有 `butter-basic-v1` catalog。[S1][S2]
2. **最合适的方向是“官方标准组件 + 少量 Butter 扩展/覆盖”，不是再引入一套 renderer。** 保留自定义 `KeyValue`、`Status`，以及承担表单状态和服务端字段错误的三个表单组件；为了维持纯文本约束，现阶段也保留 `Text`。[B1][B2]
3. **当前 10 个组件里，保守地可先复用 4 个，不是删掉整个自研层。** 现有基础控件已经复用项目 UI；复杂代码主要是表单交互而不是 Card 的样式。短期代码删除量有限，真正收益是以后增加 `List`、`Tabs`、`CheckBox` 等时不再逐个实现展示层。官方基础集本身没有 Table 或 Chart。[S1][B1]
4. **社区确实有更丰富的现成组件库，但没有核实到一个可无缝替换 Butter 当前 catalog、同时维持现有业务语义的成熟 shadcn A2UI v0.9.1 包。** 已检查的 shadcn 方案要么是旧协议/旧草案，要么实际使用另一种 JSON 协议。MUI 方案则已明确实现 v0.9.1，但会引入第二套设计系统或 renderer。[S3–S9]
5. **不要因为换组件库就撤掉 Go 侧白名单。** 现成前端组件不替代 `render_ui` 的服务端校验、只读卡片策略、表单与 Interrupt 的绑定、持久化和 fallback。[B2][B3]

**实施状态（同日更新）：** 用户已确认按上述最小范围实施。`catalog.tsx` 已改为直接复用官方四个布局组件，`surface.module.css` 提供局部主题映射，`Row` 的薄包装补回原有自动换行并约束子项宽度（官方默认不换行，无权重长内容已在 390px 浏览器探针中复现溢出）；绑定与子项解析仍复用官方实现。保留其他组件、空函数集和现有 wire format，不新增或升级依赖。新增 `catalog.test.ts` 锁定复用范围及白名单，`agui-a2ui-layout.spec.ts` 覆盖桌面/手机布局、长文本与明暗主题。下文 §3.3 记录的是实施前的调研探针，不是完整浏览器回归结果。

## 1. Butter 的约束和评估标准

当前前端使用 React 19、Tailwind CSS 4，`@a2ui/react` / `@a2ui/web_core` 均固定为 `0.12.0`，入口为 `/v0_9`，Butter 的消息版本是 `v0.9.1`，catalog ID 为 `butter-basic-v1`。[B1]

必须区分三类库：

- **同一运行时的现成 catalog/component implementations**：有机会只换展示层。
- **另一套 A2UI renderer**：协议相同也可能要重接 surface store、action、恢复、错误处理。
- **普通 UI 库或其他 Generative UI 协议**：有现成控件，但不能直接吃当前 A2UI 消息，仍需适配。

模型卡片当前只能使用 `Card`、`Row`、`Column`、`Text`、`Divider`、`KeyValue`、`Status`。`TextField`、`ChoicePicker`、`Button` 只用于服务端构造的 Human Input 表单；表单还承担 pending/submitting/answered、thread lock、字段错误和可访问性提示。[B1][B2]

## 2. 候选比较

下表版本为调研日 npm `latest`，**不是 Butter 已安装版本**；包版本号也不等于 A2UI 协议版本。[S10]

| 候选 | 现成组件 | 协议/运行时情况 | 对 Butter 的判断 |
| --- | --- | --- | --- |
| 官方 `@a2ui/react`，latest `0.14.0`，本地 `0.12.0` | 18 个基础组件，可单独导入 | 官方文档列 v0.9.1 为稳定支持；本地 0.12.0 已有完整基础集 | **首选；不必为了复用组件先升级** |
| `@copilotkit/a2ui-renderer` `1.79.1` | 基础 catalog、catalog 创建/过滤 API | 已发布包固定依赖 `@a2ui/web_core: 0.10.4`；有自身 Provider/Renderer | 可单独使用，不强制换整套 CopilotKit，但没有明显理由为同类基础组件再换 renderer |
| `@a2ui-sdk/react` `0.4.0`（easyops） | shadcn/Radix/Tailwind 标准组件 | README 明说 v0.9 实现基于 2026-01-12 草案，推荐 v0.8；并非已核实兼容当前 v0.9.1 | 风格契合，但不建议直接替换 |
| `@a2ui-bridge/react-shadcn` `0.1.0` / Mantine 适配器 | README 声称分别有 76/77 个适配器 | 文档和类型使用 `beginRendering`、`surfaceUpdate`、`literalString` 等旧消息；shadcn 包 peer 为 Tailwind `^3.4.0` | 数量丰富，但不是当前 v0.9.1 的直接替代品 |
| `@yessglory/generative-mui-react` `1.0.1` | 18 个基础组件；扩展集有 Table、图表、Progress 等 | 明确支持 v0.9.1；自有 `SurfaceStore` 和 registry；需 MUI 7、Emotion，图表另需 MUI X Charts | **若丰富表格/图表优先于统一设计系统，可做独立 POC** |
| `@language-lit/material3-expressive-a2ui` `0.2.1` | 18 个基础组件 + 9 个 Material 扩展 | catalog 可接官方 React surface，有上游互操作测试；peer web_core 仅 `^0.10.7 || ^0.11.0` | 方向最接近可插拔 catalog，但**不包含 Butter 的 0.12.0 peer 范围**，且需 Material 设计系统 |
| `@boteai/a2ui-render` / `a2ui-comp-preset` `0.0.3` | Lit 基础组件及 12 个 preset，包括表格、图表、Metric | React + Lit 桥接；已发布包 React peer 是 `^17.x`，web_core 为 `^0.9.2`，preset 依赖 AntD 4 和 Recharts 2 | 功能值得参考，但发布依赖与当前栈差距大，不能按 README 当作 React 19 即插即用 |
| `@json-render/shadcn` `0.21.3` | 文档列出 36 个现成组件，包括 Table、Grid、Progress 和各种输入 | 自己的 spec/registry 格式；A2UI 集成页明确是 concept，需要自己构建 renderer，示例还是旧 A2UI 消息 | **不是 A2UI v0.9.1 组件包**；只有考虑增加另一种 UI 协议时才值得评估 |

来源：官方 [S1][S2]；CopilotKit [S3]；easyops [S4]；Bridge [S5]；两个 Material 方案 [S6][S7]；Bote [S8]；json-render [S9]；发布依赖 [S10]。

另查到 `jacobsimionato/a2ui-shadcn`，但当前仓库只有 README 和 LICENSE，没有组件源码或 package manifest；不能算可使用的库。[S11]

## 3. 官方基础集：可以直接复用多少？

### 3.1 已安装的包就有这些组件

从本地 `@a2ui/react@0.12.0` 的导出声明、source map 内源码及运行时 catalog 实际枚举确认：[S1]

- 布局：`Row`、`Column`、`List`、`Card`、`Tabs`、`Modal`、`Divider`。
- 内容：`Text`、`Image`、`Icon`、`Video`、`AudioPlayer`。
- 输入：`Button`、`TextField`、`CheckBox`、`ChoicePicker`、`Slider`、`DateTimeInput`。

`basicCatalog` 自带的是标准 catalog ID，不是 `butter-basic-v1`。所以不能简单把 store 里的 `[butterCatalog]` 改成 `[basicCatalog]`，否则现有 surface 的 catalog ID 和两个 Butter 扩展都对不上。正确方向是**从包里挑选实现，重新组合 Butter catalog**。[S1][B1]

组合示意（不是已提交的实现）：

```tsx
import {
  Card,
  Row,
  Column,
  Divider,
} from '@a2ui/react/v0_9'
import { Catalog } from '@a2ui/web_core/v0_9'

const butterCatalog = new Catalog(
  'butter-basic-v1',
  '0.9.1',
  [
    Card, Row, Column, Divider,
    // 保留已有的 Text、KeyValue、Status 和三个表单实现。
  ],
  [], // 不因为复用组件就自动引入整个标准函数集。
)
```

### 3.2 不能不加区分地替换的部分

| 当前实现 | 建议 | 原因 |
| --- | --- | --- |
| Card / Row / Column / Divider | 优先验证官方实现 | 标准属性和组件引用形式相同；注意布局默认值和视觉差异 |
| Text | 暂留 Butter 实现 | Butter 明确要求纯文本；官方 body 支持 MarkdownContext 渲染管线，无 renderer 时才退回文本，不应无意改变富文本语义 |
| KeyValue / Status | 保留小型扩展 | 官方基础 catalog 没有这两个同名组件 |
| TextField / ChoicePicker / Button | 保留业务包装，后续再看能否复用内部控件 | 官方表单检查与 Butter 的服务器字段错误、锁定、提交/已回答状态不是同一套契约 |
| A2UIStore / SurfaceBoundary / snapshot 恢复 | 保留 | 是 Butter 的聊天和持久化适配，不是基础组件绘制代码 |

官方 Card 等使用 `--a2ui-*` CSS variables；可把颜色、间距、边框、圆角映射到 Butter 的设计 tokens，不必为统一视觉重新实现整个组件。仍需浏览器测试暗色模式、嵌套布局与表单尺寸；并非改几个变量就保证完全一致。[S1][B1]

### 3.3 本地兼容性探针

在 `front/` 直接通过 Node 导入已安装的 0.12.0 包，进行了以下处理器级检查：

1. 将官方 `Card`、`Column`、`Row`、`Divider`、`Text` 注册到 ID 为 `butter-basic-v1`、版本为 `0.9.1` 的新 catalog。
2. 输入 `v0.9.1` 的 `createSurface` / `updateComponents` / `updateDataModel`。
3. 断言 surface 存在、五个组件已进入 `componentsModel.componentsMap`、Card 实现就是官方导出。
4. 更新 `/value`，断言数据从 `Before` 变为 `After`。

**结果：通过。** 说明组合 catalog 不要求换协议或先升级到 0.14.0，但这不是 DOM 渲染、交互或视觉测试。

也尝试了 `renderToStaticMarkup`，本地 0.12.0 的 `A2uiSurface` 在 SSR 路径报 `Missing getServerSnapshot`。因此不能把该探针描述成“已成功服务端渲染”。Butter 当前是 Vite 客户端应用；本轮没有浏览器渲染验证，也没有安装所有候选库或跑端到端迁移测试。

## 4. 更丰富的组件：优先看什么？

### 如果保持现有 A2UI 协议和设计体系

优先复用官方标准组件，业务扩展继续做小型适配。**表格和图表仍是缺口**：官方 18 个基础组件并不提供它们。可复用项目已有基础 UI 或专用图表库写有限适配，但不能把“用了现成图表库”误称为“不需要 A2UI 组件适配”。[S1][B1]

### 如果明确需要现成表格、图表和更多交互组件

`generative-mui` 是本轮最值得验证的 A2UI 原生候选：协议明确为 v0.9.1，扩展组件已有实现。代价是新 renderer/store、MUI/Emotion 依赖与主题适配；仍要对接 Butter 的 `butter.a2ui` 事件、revision/seq、删除、快照恢复、fallback、表单 resume，并保持模型不能任意定义 action 的边界。应做一个隔离的结果卡片 POC，不宜直接替换表单链路。[S6][B2][B3]

Material 3 Expressive 的优势是可把 catalog 接到官方 surface，而不是必须换 store；但其已发布 peer 范围不支持当前 web_core 0.12.0。需要先取得上游兼容支持或验证后贡献修复，不能仅用忽略 peer dependency 的安装方式当作兼容证据。[S7][S10]

### 如果追求更大的 shadcn 生成式组件生态

`json-render/shadcn` 与当前 React 19 / Tailwind 4 更接近，也有较丰富的现成组件；但它使用自己的 spec、状态绑定和 action 系统。它的 A2UI 文档不是一个已经交付的 v0.9.1 消息处理器。迁移意味着新增转换层或改变 UI 协议，会波及 Go 校验、模型工具契约和已保存 surface，不能当成一次简单换库。[S9][S10][B3]

## 5. 建议的最小验证范围

先不换 renderer、不改 wire format，也不开放更多模型权限：

1. 用官方 `Card` / `Row` / `Column` / `Divider` 替换四个标准实现，保留现有 catalog ID 和其余组件。
2. 把必要的 `--a2ui-*` tokens 映射到 Butter 主题。
3. 回归流式创建/更新/删除、刷新恢复、fallback，以及表单 submitting/answered/thread lock/字段错误。
4. 用嵌套 Row/Column、长键值、移动端和暗色模式做视觉比较，判断代码减少是否值得样式适配成本。
5. 若目标是表格/图表，再单独比较“用现有 UI 写薄适配”与“引入 generative-mui”；不要为尚未确定的需求一次注册全部组件、函数和 action。

新增任何模型可用组件，仍要同步更新 `internal/a2ui/catalog.go`、工具说明和测试；若增加新的输入类型，还要同步表单 proto/构建/校验及客户端提交逻辑。现成 React 控件只能省掉其中的展示层工作。[B2][B3]

## 6. 来源与核对版本

### 本地代码

- [B1] [`front/package.json`](../../front/package.json)、[`a2ui/catalog.tsx`](../../front/src/features/agui-chat/a2ui/catalog.tsx)、[`a2ui/protocol.ts`](../../front/src/features/agui-chat/a2ui/protocol.ts)、[`a2ui/store.ts`](../../front/src/features/agui-chat/a2ui/store.ts)、[`surface-view.tsx`](../../front/src/features/agui-chat/a2ui/surface-view.tsx)。
- [B2] [`internal/a2ui/catalog.go`](../../internal/a2ui/catalog.go)、[`form.go`](../../internal/a2ui/form.go)。
- [B3] [`internal/a2uitool/a2uitool.go`](../../internal/a2uitool/a2uitool.go)、[`internal/a2ui/card.go`](../../internal/a2ui/card.go)、[`persistence.go`](../../internal/a2ui/persistence.go)。

### 上游一手资料

- [S1] 官方 React：[README](https://github.com/a2ui-project/a2ui/blob/409ded29ccac3eebc541979ae5320771703248e8/typescript/renderers/react/README.md)。本地固定版本的 [0.12.0 导出声明](https://unpkg.com/@a2ui/react@0.12.0/v0_9/index.d.ts)及 [source map（包含各组件源码）](https://unpkg.com/@a2ui/react@0.12.0/v0_9/index.js.map)；核对了 basic catalog、Card、Text、TextField、样式工具和 MarkdownContext。
- [S2] 官方 [renderers 兼容矩阵](https://github.com/a2ui-project/a2ui/blob/409ded29ccac3eebc541979ae5320771703248e8/docs/public/reference/renderers.md)、[社区 renderer 清单](https://github.com/a2ui-project/a2ui/blob/409ded29ccac3eebc541979ae5320771703248e8/docs/public/ecosystem/renderers.md)。社区清单只是发现入口，兼容性结论优先采用各项目源码和发布依赖；它仍把 easyops 列为仅 v0.8，而 easyops 自身称有旧 v0.9 草案，不能据此认定当前 v0.9.1 可用。
- [S3] CopilotKit：[renderer package](https://github.com/CopilotKit/CopilotKit/blob/b0c17366cd2eb3a7637230d139d3c9b7dc0e6bdf/packages/a2ui-renderer/package.json)、[公开 API](https://github.com/CopilotKit/CopilotKit/blob/b0c17366cd2eb3a7637230d139d3c9b7dc0e6bdf/packages/a2ui-renderer/src/react-renderer/index.ts)。
- [S4] easyops：[README 的 v0.9 草案警告、标准组件与扩展方式](https://github.com/easyops-cn/a2ui-sdk/blob/ec372f9a967497f24fe5513b75b3e25e0f23f70e/README.md)。
- [S5] A2UI Bridge：[README](https://github.com/southleft/a2ui-bridge/blob/31ed931a0d41102cffb4f58c0316a850ce425f10/README.md)、[协议类型](https://github.com/southleft/a2ui-bridge/blob/31ed931a0d41102cffb4f58c0316a850ce425f10/packages/core/src/types.ts)、[shadcn package](https://github.com/southleft/a2ui-bridge/blob/31ed931a0d41102cffb4f58c0316a850ce425f10/packages/react-shadcn/package.json)。
- [S6] generative-mui：[README 的架构、完整基础集和扩展组件表](https://github.com/yessGlory17/generative-mui/blob/260b0f264f0a5c1667743ddb40304747fc8b700b/README.md)、[React package](https://github.com/yessGlory17/generative-mui/blob/260b0f264f0a5c1667743ddb40304747fc8b700b/packages/generative-mui-react/package.json)。
- [S7] Material 3 Expressive A2UI：[README 的组件表、官方 surface 接法及兼容范围](https://github.com/Language-Lit/material3-expressive-a2ui/blob/3f5b7ec4f9fcb0a37d3a81582f592d7961146dcd/README.md)、[package](https://github.com/Language-Lit/material3-expressive-a2ui/blob/3f5b7ec4f9fcb0a37d3a81582f592d7961146dcd/package.json)、[官方 surface 互操作测试](https://github.com/Language-Lit/material3-expressive-a2ui/blob/3f5b7ec4f9fcb0a37d3a81582f592d7961146dcd/tests/runtime/official-surface.test.tsx)。上游测试不是本轮重跑结果。
- [S8] Bote：[README 的 React/Lit 架构及 preset 表](https://github.com/BoteAI/a2ui/blob/eb481292ecee4673706093ee3527b54dbe1b381c/README.md)、[renderer package](https://github.com/BoteAI/a2ui/blob/eb481292ecee4673706093ee3527b54dbe1b381c/packages/a2ui-render/package.json)、[preset package](https://github.com/BoteAI/a2ui/blob/eb481292ecee4673706093ee3527b54dbe1b381c/packages/a2ui-comp-preset/package.json)。GitHub 仓库标 Apache-2.0，而 npm 0.0.3 package 标 MIT，正式采用前还需核对许可文件。
- [S9] json-render：[shadcn README](https://github.com/vercel-labs/json-render/blob/c6f73d7d5f69acb80fc89faefd12c0865c8603d1/packages/shadcn/README.md)、[A2UI 文档的 Concept 提示与旧消息示例](https://github.com/vercel-labs/json-render/blob/c6f73d7d5f69acb80fc89faefd12c0865c8603d1/apps/web/content/docs/a2ui.mdx)。
- [S10] 2026-10-24 直接读取 npm registry，核对 dist-tags、发布日期、dependencies、peerDependencies 和 license：[官方 React](https://registry.npmjs.org/@a2ui%2freact)、[CopilotKit](https://registry.npmjs.org/@copilotkit%2fa2ui-renderer)、[easyops](https://registry.npmjs.org/@a2ui-sdk%2freact)、[Bridge shadcn](https://registry.npmjs.org/@a2ui-bridge%2freact-shadcn)、[generative-mui](https://registry.npmjs.org/@yessglory%2fgenerative-mui-react)、[Material 3](https://registry.npmjs.org/@language-lit%2fmaterial3-expressive-a2ui)、[Bote renderer](https://registry.npmjs.org/@boteai%2fa2ui-render)、[Bote preset](https://registry.npmjs.org/@boteai%2fa2ui-comp-preset)、[json-render shadcn](https://registry.npmjs.org/@json-render%2fshadcn)。这些 URL 会继续更新，正文记录的是本次查询值。
- [S11] [jacobsimionato/a2ui-shadcn 文件树](https://github.com/jacobsimionato/a2ui-shadcn/tree/b1467e98391d5864ad756468fae1f1b3679da4fb)。

许可概览：官方 A2UI、easyops、json-render 为 Apache-2.0；CopilotKit renderer、Bridge、generative-mui、Material 3 A2UI 的 package 元数据为 MIT；Bote 见上文差异。[S1][S3–S10]
