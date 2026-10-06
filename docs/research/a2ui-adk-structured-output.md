# A2UI 与 ADK 深度集成：需要结构化输出吗？

调研日期：2026-10-01。只用一手来源：

- A2UI 仓库 [a2ui-project/a2ui@102ec1a](https://github.com/a2ui-project/a2ui/tree/102ec1a0497510eede5dc1938d6d1cc6042b3370)（2026-10-01）：规范 JSON 与文档、Python agent SDK、`samples/` 下全部 ADK 样例、`eval/`。a2ui.org 由该仓库的 `docs/public` 生成；首页、v0.9.1 和 v1.0 规范页已抽查，与仓库一致。
- a2ui.org 的 "Any Agent Framework" 指南所指向的 [ag-ui-protocol/ag-ui@015baf5](https://github.com/ag-ui-protocol/ag-ui/tree/015baf5dc1504898193f86943b6d85950bacb86d)。
- 模块缓存中的 ADK Go v2.1.0、genai v1.63.0、adk-utils-go v0.22.0、jsonschema-go v0.4.3。下文行号与各自的 tag 一致；adk-go main（`60216c0`）已复核。
- [adk.dev](https://adk.dev/agents/llm-agents/)：只用于 Python/Java 的语义，引用处会标注。
- Gemini 与 OpenAI 的官方文档。Butter 只接了两类 `model.LLM`：ADK `model/gemini`，以及经 adk-utils-go 调用的 OpenAI Chat Completions（`internal/agent/model.go:128-151`）。Anthropic 只出现在 ButterBox 的 pi/opencode 配置中，不经过 ADK 模型调用，因此不在本文范围内。

**未做的事：** 没有调用任何模型，也没有运行 A2UI 样例或 Butter。唯一运行过的是 scratchpad 里的一个探针：用 jsonschema-go v0.4.3 推断 `render_ui` 的参数 schema（见 §3.5）。上游代码的"行为"结论都来自读代码。

**升级到 ADK Go v2.5.0 后的复核（同日）：** §3 的结论不变。

- `OutputSchema` 仍是 `*genai.Schema`。注释改成了"会注入 `set_model_response`，模型仍可调用其他工具"（[llmagent.go L314-L318](https://github.com/google/adk-go/blob/v2.5.0/agent/llmagent/llmagent.go#L314-L318)）。
- Gemini API 的回退判断（`internal/llminternal/googlellm/variant.go`）和 ADK 自带的校验器（`internal/utils/schema_utils.go`）与 v2.1.0 逐字相同。`OutputKey` 仍存字符串，校验仍是 TODO（[llmagent.go L521-L551](https://github.com/google/adk-go/blob/v2.5.0/agent/llmagent/llmagent.go#L521-L551)）。
- jsonschema-go 仍是 v0.4.3，所以 `jsonschema_description` 仍被忽略。`MaxLLMCalls` 仍只在 `LiveRunConfig` 里。
- 新增的是 ADK 原生 `model/openaimodel`（实验性），v2.5.0 起也支持 Chat Completions。它把 ResponseSchema/ResponseJsonSchema 映射成 strict `response_format` 时，会给每个对象补 `additionalProperties: false`，并把全部字段设为 required（[schema.go L80-L108](https://github.com/google/adk-go/blob/v2.5.0/model/openaimodel/internal/shared/schema.go#L80-L108)）；函数工具则固定 `strict: false`。Butter 仍用 adk-utils-go，所以 §3.6 的 OpenAI 一列仍然适用。

## 结论

- **不需要。** 这里的"结构化输出"指 OutputSchema、response_schema 或 constrained decoding。A2UI 不依赖它，官方也明确不走这条路。v0.9 把协议从 "Structured Output First" 改成了 "Prompt First"，理由是结构化输出的 schema 子集限制了 catalog 的表达力；代价是必须做生成后校验和纠错（[v0.9 evolution L7-L10](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9/docs/evolution_guide.md#L7-L10)、[v0.9.1 协议 L31-L40](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L31-L40)）。官方 Python SDK、全部 ADK 样例和 AG-UI 的 A2UI 工具链，都没有用 response schema 约束 A2UI 信封。仓库里出现 `response_schema` 的几处，约束的都只是小的业务数据，再由代码拼出 A2UI（§2.3）。
- **Butter 现有的 `render_ui` 方向是对的。** 它的做法是"工具调用 + 整批服务端校验 + 把错误回给模型"，与官方 `send_a2ui_json_to_client` 和 AG-UI `render_a2ui` 同型。差别在于：Butter 用宽松的结构化参数，官方 SDK 用一个 JSON 字符串参数。按代码对比，Butter 的校验比官方 SDK 现状更严格，因为官方 toolset 在 2026-09-23（#2715）之后丢弃了组件校验结果（§2.2）。
- **不要用 ADK Go v2.1.0 的 `OutputSchema` 承载 A2UI 信封。** 原因有四：
  - `*genai.Schema` 没有 `oneOf/$ref/const/additionalProperties`。
  - Butter 每个 LLM agent 都挂了 toolset，在 Gemini API 上必然落入 `set_model_response` 回退。
  - ADK Go 自带的校验器不支持 `anyOf`。
  - OpenAI 适配器把 schema 映射成 `strict: true`，却不补 `additionalProperties: false`。

  更根本的是，邻接表的引用完整性（root、悬空 id、父子约束）超出了 JSON Schema 的表达范围，约束解码代替不了服务端校验（§3、§4）。
- **结构化输出真正划算的位置是"服务端模板 + 小而平的数据 schema"（§4 方案 d）。** AG-UI 的 fixed schema 模式、A2UI SDK 的 macros 和社区样例都这么做，而且这种 schema 落在所有 provider 的 strict 子集之内。其次可以给 `render_ui` 换上从 catalog 派生的精确参数 schema（方案 b'）。但在 Butter 当前的两个适配器上，它带来的只是 ADK 侧的前置校验和更好的提示，得不到约束解码：两个适配器都没开 strict，也没用 VALIDATED 模式。
- **两处需要代码侧修正或实测：**
  1. `renderArgs` 用的 `jsonschema_description` tag 不被 jsonschema-go 识别，字段描述没有进入工具声明（探针已证实）。**已在 #374 修复。**
  2. AG-UI 的 ADK 适配器注释称，Gemini 函数调用会把"无属性的 object 数组"填成 `{}`，所以它改用了 JSON 字符串参数。`render_ui.messages` 的形状与此接近，需要在真实 Gemini 上验证。**OpenAI 的 `gpt-5.6-luna` 已实测，没有这个问题（§5.1）；Gemini 暂未测。**

## 1. A2UI v0.9.1 用法要点（附 v1.0 对生成侧有影响的差异）

- **消息与生命周期。** 服务端发往客户端的信封只有四种，每条恰好含一个键：
  - `createSurface`：`surfaceId` 和 `catalogId` 必填，可带 `theme`、`sendDataModel`；创建后二者固定，要改须先删除。
  - `updateComponents`。
  - `updateDataModel`：`path` 缺省为 `/`，即整体替换；省略 `value` 表示删除。
  - `deleteSurface`。

  组件中必须有一个 `id: "root"`（[协议 L173-L281](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L173-L281)）。v0.9.1 相对 v0.9 只有两处改动：MIME 统一为 `application/a2ui+json`，surfaceId 唯一性放宽到"当前活跃的 surface 之间"（[evolution L7-L10](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/evolution_guide.md#L7-L10)）。
- **扁平邻接表。** 组件是平铺列表，用 id 引用子节点，可以按任意顺序发送。客户端先缓冲，等 root 出现再渲染，并跳过无效引用（[L320-L328](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L320-L328)）。官方给的理由是：平铺结构加 `component` 判别字段，比嵌套或动态键更容易让 LLM 稳定生成，也便于流式（[v0.9 evolution L126-L131](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9/docs/evolution_guide.md#L126-L131)、[components.md L5-L18](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/docs/public/concepts/components.md#L5-L18)）。自定义 catalog 必须用 `ComponentId`/`ChildList` 作引用类型，校验器才会检查被引用的组件是否存在（[L160-L171](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L160-L171)）。
- **数据模型与绑定。** 绑定用 JSON Pointer。`ChildList` 模板（`{componentId, path}`）会为数组每一项建一个子作用域，作用域内不以 `/` 开头的路径是相对路径。输入组件的双向绑定只改客户端本地模型，只有触发 action 时才把值带回服务端。若设了 `sendDataModel`，整份模型会随每条客户端消息放进传输元数据（[L399-L593](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L399-L593)）。
- **动态值与客户端函数。** `Dynamic*` 类型可以是字面量、`{path}` 或 `FunctionCall`。catalog 同时定义了一组函数：`required/regex/length/numeric/email/formatString/formatNumber/formatCurrency/formatDate/pluralize/openUrl/and/or/not`。输入组件和按钮用 `checks` 做客户端校验，按钮的 checks 不通过时会被禁用。`${...}` 插值只允许出现在 `formatString` 里（[L595-L779](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L595-L779)、[v0.9 evolution L222-L235](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9/docs/evolution_guide.md#L222-L235)）。
- **action 与回传。** 按钮的 `action.event{name, context}` 发往服务端，`action.functionCall` 调用本地函数。客户端发往服务端的消息有两种：`action{name, surfaceId, sourceComponentId, timestamp, context}`，以及 `error`（标准格式为 `VALIDATION_FAILED{surfaceId, path, message}`）（[L353-L393](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L353-L393)、[L796-L861](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L796-L861)）。
- **catalog 的声明与选择。**
  - catalog 是一份 JSON Schema；`catalogId` 只是标识，运行时不会去下载。
  - 客户端在每条消息的元数据里按偏好顺序给出 `supportedCatalogIds`。也可以带 `inlineCatalogs`，前提是 agent 声明了 `acceptsInlineCatalogs`。
  - agent 在 `createSurface` 时选定一个，在该 surface 的生命周期内锁定；没有匹配项就不发 UI。
  - 校验分两段：agent 发送前一次，客户端收到后一次（[catalogs.md L290-L475](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/docs/public/concepts/catalogs.md#L290-L475)）。
- **主题。** v0.9.1 基础 catalog 的 `theme` 只有 `primaryColor`、`iconUrl`、`agentDisplayName` 三项。后两项用来标明 surface 来自哪个 agent，编排器应改写或核验，以防冒充（[L711-L725](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L711-L725)）。
- **流式与消息边界。** 传输层必须保序，必须明确分隔每条 JSON 信封（JSONL 的行、WebSocket 帧或 SSE 事件），还必须能携带元数据（[L84-L93](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L84-L93)）。关于 AG-UI，a2ui.org 只说"由 AG-UI/CopilotKit 处理"，没有给出事件格式（[transports.md L44-L48](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/docs/public/concepts/transports.md#L44-L48)）。AG-UI 自己的 `@ag-ui/a2ui-middleware` 用的是 `ACTIVITY_SNAPSHOT` 加 `activityType: "a2ui-surface"`（[index.ts L56、L932-L976](https://github.com/ag-ui-protocol/ag-ui/blob/015baf5dc1504898193f86943b6d85950bacb86d/middlewares/a2ui-middleware/src/index.ts#L932-L976)）。这与 Butter 的 `CUSTOM butter.a2ui` 不同，但与结构化输出无关。
- **v1.0 候选版中影响生成侧的差异**（[v1.0 evolution L9-L98](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v1_0/docs/evolution_guide.md#L9-L98)、[v1.0 协议 L176-L188、L552-L564、L760-L808](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v1_0/docs/a2ui_protocol.md#L552-L564)）：
  - `createSurface` 可以直接带 `components` 和 `dataModel`，一条消息就能给出完整 UI。
  - 删除了 `theme`。
  - `updateDataModel.value` 改为必填，用 `null` 表示删除。
  - surfaceId 恢复为"在渲染器生命周期内全局唯一"。
  - catalog 新增 `instructions` 字段（取代 `rules.txt`），以及 `allowedParents/allowedChildren`。规范说明加后者的原因是"JSON Schema 无法在平铺的 id 引用上约束子组件类型"。
  - 为了让 catalog 能可靠地转成"LLM 友好的 DSL"，v1.0 禁止 catalog 自定义 `$defs`，并限制 `$ref` 的目标。
  - 新增双向函数调用：`callRendererFunction/agentFunctionResponse/callAgentFunction/rendererFunctionResponse`。

  仓库里另有 Express、Atom、Elemental 等紧凑 DSL 提案，声称比 JSON 少 55–70% 的输出 token，可以按行流式解析，目前由环境变量开关控制（[a2ui_express.md L3-L16](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/proposals/express/a2ui_express.md#L3-L16)）。可见演进方向是更紧凑的文本格式，而不是 schema 约束解码。

## 2. A2UI 自己的 agent 侧如何让 LLM 产出 A2UI

### 2.1 规范给出的循环与理由

规范给出的是三步循环：Prompt → Generate → Validate。Prompt 里放期望的 UI、含 catalog 的 A2UI JSON Schema 和合法示例；结果不合法时，把错误按 `VALIDATION_FAILED{surfaceId, path, message}` 报回给 LLM，让它自纠（[v0.9.1 L781-L816](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9_1/docs/a2ui_protocol.md#L781-L816)）。v1.0 沿用这个循环，并增加了 `UNALLOWED_PARENT/CHILD` 两个错误码（[L1282-L1304](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v1_0/docs/a2ui_protocol.md#L1282-L1304)）。

v0.8 规范自称"为结构化输出模式优化"（[v0.8 L634](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_8/docs/a2ui_protocol.md#L634)）。v0.9 evolution 记录了放弃它的原因：

- v0.8.1 为了适配 strict JSON mode 和函数调用（原文："which is also a form of Structured Output"），采用了"深嵌套和特定包装结构……常让 LLM 费解"。
- v0.9 改为把 schema 内嵌进 prompt，偏向 LLM 擅长的普通 JSON。例如用对象表示 map，而不是键值对数组。
- 难以用 JSON Schema 表达的条件约束，改为用自然语言写进 `rules.txt`（[L7-L10、L59-L66、L178-L192](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/specification/v0_9/docs/evolution_guide.md#L178-L192)）。

### 2.2 Python agent SDK（`a2ui-agent-sdk`）

- **prompt 里放什么，有多大。**
  - 入口是 `DirectJsonFormat`；`A2uiSchemaManager` 已只是它的弃用别名（[manager.py L15-L55](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/schema/manager.py#L15-L55)）。
  - catalog 按客户端能力选择，优先级依次为：合并 inline catalog、第一个双方都支持的、默认（[format.py L153-L253](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/inference_formats/direct_json/format.py#L153-L253)）；还可以按组件和消息类型裁剪。
  - `render_as_llm_instructions()` 把 server_to_client、common_types、catalog 三份 schema 以紧凑 JSON 原样放进 `---BEGIN A2UI JSON SCHEMA---` 块（[catalog.py L427-L452](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/schema/catalog.py#L427-L452)）。
  - 再拼上工作流规则（每个块包在 `<a2ui-json>…</a2ui-json>` 里，root 放第一个，父组件先于子组件，便于流式；[constants.py L91-L110](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/schema/constants.py#L91-L110)）和示例。
  - 体量（本地测量，紧凑 JSON 计）：v0.9.1 基础 catalog 未裁剪时，三份 schema 共约 3.95 万字符（4,509 + 5,780 + 29,240）；restaurant_finder 的 v0.9 示例另有约 15 KB。
  - 官方 eval 基线（direct_json，gemini-3.5-flash，thinking 0，v1.0 数据集 51 例，单次运行）：输入 token 中位数 12,522，输出 958，schema 通过率 96.1%（[run_meta.json](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/eval/baselines/direct_json/run_meta.json)）。
- **如何取出。** 有两条路：
  - 文本中的 `<a2ui-json>` 标签块。缺闭合标签或块为空时报错（[parser.py L27-L69](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/inference_formats/direct_json/parser.py#L27-L69)）。
  - 工具 `send_a2ui_json_to_client`。它唯一的参数 `a2ui_json` 被声明为 **STRING**；schema 和示例由工具在 `process_llm_request` 里追加到 system instruction。模块文档自称这"只是捕获 A2UI JSON 的一种方式"（[toolset L21、L221-L292](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/adk/send_a2ui_to_client_toolset.py#L221-L292)）。

  两条路的结果都会转成 A2A `DataPart`（[part_converter.py L88-L143](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/adk/a2a/part_converter.py#L88-L143)）。
- **如何校验。**
  - `parse_and_fix` 只修弯引号和尾逗号（[payload_fixer.py L24-L95](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/parser/payload_fixer.py#L24-L95)）。
  - `DirectJsonParser.compile` 只在注入了 validator 时才校验，源码里留着 TODO（[parser.py L106-L128](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/inference_formats/direct_json/parser.py#L106-L128)）。
  - `A2uiCatalog.validate_components` 是**返回**错误列表，而不是抛异常（[catalog.py L256-L275](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/schema/catalog.py#L256-L275)）。
  - [#2715](https://github.com/a2ui-project/a2ui/commit/8d74a873da2b5278cba254ac53e87fae6f23dd87) 把 toolset 和样例里原先会抛异常的 `validator.validate(...)` 换成了它，但调用方都没有检查返回值（[toolset L307](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/adk/send_a2ui_to_client_toolset.py#L307)、[restaurant_finder L336](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/samples/agent/adk/restaurant_finder/agent.py#L336)）。
  - 单测用 mock 让它抛异常，因此没有暴露这个问题（[test L262-L278](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/tests/adk/test_send_a2ui_to_client_toolset.py#L262-L278)）。
  - 读代码的结论：在 HEAD 上，组件级的 schema 错误既不会让工具报错，也不会触发样例重试；只有 JSON 解析错误会。
  - 样例还用 `remove_strict_validation` 去掉了 schema 里所有的 `additionalProperties/unevaluatedProperties: false`（[common_modifiers.py L18-L34](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/schema/common_modifiers.py#L18-L34)），方向与 strict 结构化输出正好相反。
- **重试。** SDK 本身没有重试逻辑（grep 无命中）。工具失败时返回 `{"error": ...}`，交给模型处理。官方样例在 agent 外层手写了"最多 2 次"：失败后用 "Your previous response was invalid… You MUST…" 重新提问，次数用尽就回一句文字道歉（[restaurant_finder L222-L414](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/samples/agent/adk/restaurant_finder/agent.py#L222-L414)）。
- **流式。** `DirectJsonStreamParser` 在文本流中找标签，补全半截 JSON（只给 `text/label` 等"可截断键"补引号，URL 不补），按 root 可达性增量产出组件，并对每条消息做 Draft 2020-12 校验（[streaming.py L314-L559](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/python/a2ui_agent/src/a2ui/inference_formats/direct_json/streaming.py#L314-L559)）。流式模式下组件先推给客户端，整段校验要等流结束后才做（[restaurant_finder L281-L359](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/samples/agent/adk/restaurant_finder/agent.py#L281-L359)）。

### 2.3 样例逐一对照（`samples/` 下全部 ADK 目录，外加仓库里所有出现 response schema 的地方）

| 样例 | 生成机制 | 校验 | 重试 |
|---|---|---|---|
| `agent/adk/restaurant_finder`（官方，Gemini） | prompt 带全量 schema 和示例，文本输出 `<a2ui-json>`；按 UI_DESCRIPTION 选示例当模板 | `parse_response` + `validate_components`（返回值被丢弃） | 外层重问 1 次；流式增量推送 |
| `agent/adk/custom-components-example`（官方，LiteLLM） | 同上，另外接受 inline catalog | 同上 | 同上 |
| `community/.../rizzcharts` | 工具 `send_a2ui_json_to_client(a2ui_json: string)`；catalog 从 A2A 元数据选出后存进 session state，每次请求注入 | toolset（同上） | 无；prompt 要求出错时向用户道歉（[agent.py L76](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/samples/community/agent/adk/rizzcharts/python/agent.py#L55-L79)） |
| `community/.../gemini_enterprise` v0_8/v0_9 | 同 restaurant_finder | 同上 | 外层 1 次 |
| `community/.../orchestrator` 的子 agent | instruction 里写死整张表单，模型原样回显（[subagent_front_desk.py L64-L118](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/samples/community/agent/adk/orchestrator/subagent_front_desk.py#L64-L118)） | `A2uiPartConverter` 解析 | 无 |
| `community/.../mcp_app_proxy`、`file_upload_summarizer` | 工具函数在代码里拼好 A2UI，返回 `validated_a2ui_json`；`commentate_pong_game` 只让一个嵌套 LLM 写一句话，再放进 `updateDataModel`（[tools.py L392-L434](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/samples/community/agent/adk/mcp_app_proxy/tools.py#L392-L434)） | 无需校验（服务端构造） | 不需要 |
| `community/.../personalized_learning` | 工具内嵌套一次 Gemini 调用，用 `response_mime_type="application/json"`（JSON mode，无 schema），示例是 v0.8 格式（[agent.py L598-L656](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/samples/community/agent/adk/personalized_learning/agent.py#L598-L656)） | 只有 `json.loads` | 无 |
| `community/client/lit/personalized_learning/deploy.py` | `response_schema` 只约束 `{front, back, category}` 卡片数组，随后"programmatically"拼出 A2UI（[L1109-L1168](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/samples/community/client/lit/personalized_learning/deploy.py#L1109-L1168)） | 构造出来即合法 | 不需要 |
| `community/mcp/a2ui-in-mcpapps` 的 smart_editor（不用 ADK） | `response_schema` 约束 2–3 个控件规格（type 枚举 + label），再拼出 A2UI（[L42-L135](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/samples/community/mcp/a2ui-in-mcpapps/server/smart_editor_agent.py#L42-L135)） | 失败时用默认控件 | 无 |
| `community/.../mcp-apps-in-a2ui-sample`（目录在 adk 下，实为 FastAPI） | 返回静态的 v0.8 消息 | — | — |
| `eval/` 的 subagent_tool 策略 | 主模型调用 `a2ui_specialist(input)`，子调用使用带 schema 的 prompt，从文本标签中取出结果（[L44-L100](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/eval/a2ui_eval/strategies/subagent_tool.py#L44-L100)）。eval 的其余策略也全是文本格式（direct/express/elemental/atom） | `parse_response` | 解析失败时把错误交给主模型 |
| `tools/editor`（Gemini） | prompt-first；代码里 `responseMimeType` 和 `responseJsonSchema: {array of ServerToClientMessage}` 被注释掉了（[gemini.ts L127-L134](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/tools/editor/middleware/gemini.ts#L127-L134)） | — | — |

### 2.4 AG-UI 一侧（a2ui.org 指南所指向的实现）

a2ui.org 的指南说：CopilotKit 会把客户端 catalog 转发到服务端，并注入 `generate_a2ui` 工具；在"固定 schema 流程"中，由 agent 自己返回 `a2ui_operations`（[a2ui-with-any-agent-framework.md L348-L364](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/docs/public/guides/a2ui-with-any-agent-framework.md#L348-L364)）。AG-UI 自己的说明把两种模式分得很清楚（[SKILL.md L42-L46、L104-L108](https://github.com/ag-ui-protocol/ag-ui/blob/015baf5dc1504898193f86943b6d85950bacb86d/skills/ag-ui-a2ui-integration/SKILL.md#L42-L108)、[runtime 说明 L56-L104](https://github.com/ag-ui-protocol/ag-ui/blob/015baf5dc1504898193f86943b6d85950bacb86d/skills/ag-ui-a2ui-integration/references/a2ui-runtime-and-renderer.md#L56-L104)）：

- **fixed schema**：后端工具返回 `a2ui_operations`，"每次只有数据变化"。规则原文是"不要让模型发明组件树"。
- **dynamic schema**：`generate_a2ui` 运行一个被**强制调用** `render_a2ui` 的子 agent，参数以流式产出，以便渐进绘制；toolkit 在"校验并重试之后"才提交结果。

ADK 适配器的具体实现（[a2ui_tool.py L301-L366](https://github.com/ag-ui-protocol/ag-ui/blob/015baf5dc1504898193f86943b6d85950bacb86d/integrations/adk-middleware/python/src/ag_ui_adk/a2ui_tool.py#L301-L366)、[recovery.py L18-L100](https://github.com/ag-ui-protocol/ag-ui/blob/015baf5dc1504898193f86943b6d85950bacb86d/sdks/python/a2ui_toolkit/ag_ui_a2ui_toolkit/recovery.py#L18-L100)、[toolkit L130-L160](https://github.com/ag-ui-protocol/ag-ui/blob/015baf5dc1504898193f86943b6d85950bacb86d/sdks/python/a2ui_toolkit/ag_ui_a2ui_toolkit/__init__.py#L130-L160)）：

- 子请求使用 `FunctionCallingConfigMode.ANY`，只允许调用 `render_a2ui`。
- 共享定义里 `components` 是 `array<object>`，但 ADK/Gemini 版本把它改成了 JSON **字符串**。注释给出的原因是："Gemini 的函数调用严格填充有类型的参数，对无属性的 object 数组会给出空 `{}`"。
- schema 借用 A2UI SDK 的 `render_as_llm_instructions` 渲染进 system prompt。
- 校验内容包括结构、catalog 成员、必填属性和绝对绑定路径。不通过时把错误附在 prompt 后面重新生成，默认最多尝试 3 次。

这里的"结构化"仅限于函数调用本身，没有用 response schema。

### 2.5 对早先说法的核对

这些说法来自 #350 之前的两份草稿笔记，草稿没有入库。

- "v0.9.1 是 prompt-first"：成立。
- "官方 Agent Development 指南用 ADK + `A2uiSchemaManager` 把 schema 与示例放进 system prompt"：实质成立，但该指南已经过时：
  - 它导入的 `a2ui.strategies.schema` 在 HEAD 中不存在；
  - 默认版本是 v0.8；
  - 文中描述的 `render/update` 操作不是 v0.9 的消息名；
  - 结尾仍是 TODO；
  - 第 10 行 "Use LLM structured output or prompts" 只是宽泛说法（[agent-development.md L10、L101-L186](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/docs/public/guides/agent-development.md#L101-L186)）。

  介绍页 "LLMs generate as structured output" 的说法也是如此（[what-is-a2ui.md L35](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/docs/public/introduction/what-is-a2ui.md#L35)），应以规范为准。
- "CopilotKit 注入 `generate_a2ui` 工具"：成立，机制见 §2.4。
- #350 的 spec 在 Out of Scope 里写了"v1.0 及其 actionResponse RPC"。这一说法来自 a2ui.org 首页和 overview 的过时摘要（[index.md L27](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/docs/public/index.md#L27)、[overview.md L50-L52](https://github.com/a2ui-project/a2ui/blob/102ec1a0497510eede5dc1938d6d1cc6042b3370/docs/public/concepts/overview.md#L50-L52)）。v1.0 规范页和 JSON schema 里既没有 `actionResponse`，也没有 `surfaceProperties`；v1.0 的 RPC 是 §1 列出的四种消息。

## 3. ADK Go v2.1.0 的结构化输出语义

### 3.1 OutputSchema、InputSchema、OutputKey

- `InputSchema`（agent 被当作工具调用时的入参）和 `OutputSchema` 都是 `*genai.Schema`。注释写的是"设置 OutputSchema 后 agent 只能回复，不能使用工具、RAG 或 agent transfer"（[llmagent.go L304-L311](https://github.com/google/adk-go/blob/v2.1.0/agent/llmagent/llmagent.go#L304-L311)）。`OutputKey` 把最终回复存进 session state（[L329-L334](https://github.com/google/adk-go/blob/v2.1.0/agent/llmagent/llmagent.go#L329-L334)）。
- 如何落到请求上：在不需要回退时，`basicRequestProcessor` 设置 `Config.ResponseSchema` 和 `ResponseMIMEType = "application/json"`。ADK Go 从不设置 `ResponseJsonSchema`。`ModeTask` 跳过这一步，改由 `finish_task` 工具的 `Parameters` 承载（[basic_processor.go L46-L56](https://github.com/google/adk-go/blob/v2.1.0/internal/llminternal/basic_processor.go#L46-L56)、[finish_task_tool.go L57-L137](https://github.com/google/adk-go/blob/v2.1.0/internal/workflowinternal/finish_task_tool.go#L57-L137)）。
- `*genai.Schema` 是 OpenAPI 3.0 的子集。它有 `anyOf/enum/items/properties/required/nullable/propertyOrdering/min*/max*/pattern`，没有 `oneOf/allOf/$ref/$defs/const/additionalProperties`（[types.go L1846-L1911](https://github.com/googleapis/go-genai/blob/v1.63.0/types.go#L1846-L1911)）。对照之下，A2UI 的 catalog 依赖 `allOf + unevaluatedProperties: false + const + discriminator`，common_types 依赖 `oneOf` 和递归的 `FunctionCall`。

### 3.2 与工具、子 agent、transfer 同时存在时

- **Gemini 模型有回退。** 回退条件是：agent 有 `Tools` 或 `Toolsets`，模型名以 `gemini-` 开头，且不是"Vertex AI + Gemini ≥ 2.0"。满足时不设 ResponseSchema，改为：
  - 注入 `set_model_response` 工具，其 `ParametersJsonSchema` 就是 OutputSchema；
  - 追加一句"必须用它给出最终答复"的指令；
  - 工具的函数响应再被转成一条最终文本事件。

  代码见 [outputschema_processor.go L32-L145](https://github.com/google/adk-go/blob/v2.1.0/internal/llminternal/outputschema_processor.go#L32-L145)、[variant.go L76-L105](https://github.com/google/adk-go/blob/v2.1.0/internal/llminternal/googlellm/variant.go#L76-L105)、[base_flow.go L647-L657](https://github.com/google/adk-go/blob/v2.1.0/internal/llminternal/base_flow.go#L647-L657)。adk-python 的 main 分支同样只在 Vertex AI 上认为二者可以组合（[_capabilities.py L30-L51](https://github.com/google/adk-python/blob/3d1d12ded0043ecae6f9607fb5a1c91c6cfa9027/src/google/adk/models/_capabilities.py#L30-L51)）。
- **一手来源之间有冲突。** Gemini 文档称"Gemini 3 可以把结构化输出与函数调用组合使用"（[structured-output](https://ai.google.dev/gemini-api/docs/structured-output)）；adk.dev 也写"包括 Gemini 3.0 在内的特定模型支持"，否则回退到一个"可能不可靠"的函数工具，并建议"用子 agent 单独做输出格式化"。这两处都与 ADK 代码不一致，adk-go main 至今未改。
- **非 Gemini 模型名**（例如 OpenAI 适配器）：直接设置 ResponseSchema，同时照常发送工具。
- **transfer**：`transfer_to_agent` 由排在后面的 `AgentTransferRequestProcessor` 追加，这一步不检查 OutputSchema；前面判断回退时的 `hasTools` 也不把 transfer 计算在内（[base_flow.go L79-L96](https://github.com/google/adk-go/blob/v2.1.0/internal/llminternal/base_flow.go#L79-L96)、[agent_transfer.go L69-L97](https://github.com/google/adk-go/blob/v2.1.0/internal/llminternal/agent_transfer.go#L69-L97)）。按代码推断，在 Gemini API 上"设了 OutputSchema、有子 agent、没有其他工具"时，会把 ResponseSchema 和函数声明同时发出（未运行验证）。

### 3.3 OutputKey 存的是什么；流式行为

- **普通聊天流程**中，`maybeSaveOutputToState` 存的是非 thought 文本拼成的**字符串**。设了 OutputSchema 时只会跳过空白块，校验和反序列化仍是 TODO（[llmagent.go L510-L543](https://github.com/google/adk-go/blob/v2.1.0/agent/llmagent/llmagent.go#L510-L543)）。adk.dev 所说的"设了 output_schema 就存解析后的 dict/Map"只适用于 Python、Java 和 Kotlin。
- **作为 workflow 节点运行时**才会调用 `ValidateOutputSchema`，并把解析后的对象写入 OutputKey；**作为工具（agenttool）运行时**同样校验，解析结果作为工具结果返回（[llm_agent_wrapper.go L157-L226](https://github.com/google/adk-go/blob/v2.1.0/agent/llmagent/llm_agent_wrapper.go#L157-L226)、[agent_tool.go L212-L219](https://github.com/google/adk-go/blob/v2.1.0/tool/agenttool/agent_tool.go#L212-L219)）。
- **流式**：partial 事件照常产出，但不写入 state（[base_flow.go L612-L620](https://github.com/google/adk-go/blob/v2.1.0/internal/llminternal/base_flow.go#L612-L620)）。Gemini 流式输出的结构化结果是可以拼接的部分 JSON。如果走的是 `set_model_response`，最终答复是一次函数调用，因此不会有文本流。

### 3.4 ADK Go 自带的校验器很浅

`ValidateMapOnSchema/matchType` 只认 `Type` 字段：

- 没有 `Type` 的 `anyOf` 节点会直接报 "unsupported type"；
- 未在 schema 中声明的键一律报错；
- `ValidateOutputSchema` 只接受顶层是对象（[schema_utils.go L27-L138](https://github.com/google/adk-go/blob/v2.1.0/internal/utils/schema_utils.go#L27-L138)）。

`set_model_response`、workflow 节点和 agenttool 用的都是这个校验器。所以任何用 `anyOf` 表达"组件联合"的 OutputSchema，在 ADK 自己这一层就会失败。

### 3.5 functiontool 的参数 schema（函数调用本身也是一种受约束的输出）

- `functiontool.New` 用 jsonschema-go 从参数类型推断 schema，也可以用 `Config.InputSchema` 覆盖。schema 以 `ParametersJsonSchema`（而不是 `Parameters`）发出（[function.go L37-L47、L159-L199](https://github.com/google/adk-go/blob/v2.1.0/tool/functiontool/function.go#L159-L199)）。
- `Run` 先用解析后的 schema 校验参数，再调用 handler（[convert.go L27-L56](https://github.com/google/adk-go/blob/v2.1.0/internal/typeutil/convert.go#L27-L56)）。失败时以 `{"error": ...}` 返回给模型（[base_flow.go L1233-L1276](https://github.com/google/adk-go/blob/v2.1.0/internal/llminternal/base_flow.go#L1233-L1276)）。
- jsonschema-go 只识别 `jsonschema` 这个 struct tag（[infer.go L79、L329-L337](https://github.com/google/jsonschema-go/blob/v0.4.3/jsonschema/infer.go#L329-L337)）。

探针做法：把 `renderArgs` 原样复制到 scratchpad，用 jsonschema-go v0.4.3 的 `jsonschema.For` 推断。得到的 schema 是：

`{"type":"object","properties":{"surface_id":{"type":"string"},"messages":{"type":["null","array"],"items":{"type":"object","additionalProperties":true}},"fallback":{"type":"string"}},"required":["messages"],"additionalProperties":false}`

可见三个 `jsonschema_description` 都没有进入声明（`internal/a2uitool/a2uitool.go:58-62`，#374 已改为 `jsonschema`）；A2UI 的组件词汇只出现在工具描述里（约 2.2K 字符，`description()`，`:130-149`）。此外，非 live 运行没有 LLM 调用次数上限：`MaxLLMCalls` 只存在于 `LiveRunConfig`（[live.go L48](https://github.com/google/adk-go/blob/v2.1.0/agent/live.go#L48)）。工具报错后模型何时停止自纠，完全由模型自己决定；Butter 仓库里也没搜到针对单次 invocation 的上限。

### 3.6 Butter 的两个适配器是否真的执行这些 schema

| | Gemini（ADK `model/gemini` → genai） | OpenAI（adk-utils-go v0.22.0，Chat Completions） |
|---|---|---|
| 工具参数 | `ParametersJsonSchema` 原样透传（[gemini.go L129、L145](https://github.com/google/adk-go/blob/v2.1.0/model/gemini/gemini.go#L129-L145)）。ADK 不设 `ToolConfig`，所以是默认的 AUTO 模式；文档只对 VALIDATED 模式写了 "ensures function schema adherence"（[function-calling](https://ai.google.dev/gemini-api/docs/function-calling)、[types.go L336-L357](https://github.com/googleapis/go-genai/blob/v1.63.0/types.go#L336-L357)） | 优先取 `ParametersJsonSchema`，把类型名改成小写、补上 `properties`，**不设 `strict`**（[openai.go L536-L590](https://github.com/achetronic/adk-utils-go/blob/v0.22.0/genai/openai/openai.go#L536-L590)）。OpenAI 文档写明 "Chat Completions requests remain non-strict by default"，即只是尽力而为（[function-calling](https://developers.openai.com/api/docs/guides/function-calling)） |
| 有 OutputSchema、无工具 | 走 `ResponseSchema`，由服务端按 OpenAPI 子集约束 | 映射成 `response_format: json_schema, strict: true`。但 `convertSchema` 只保留 type/description/required/enum/properties/items，丢掉 anyOf、nullable 等，也不补 `additionalProperties: false`（[openai.go L314-L335、L641-L682](https://github.com/achetronic/adk-utils-go/blob/v0.22.0/genai/openai/openai.go#L641-L682)）。OpenAI strict 要求每个对象都有 `additionalProperties: false` 且所有字段 required，否则拒绝请求（[structured-outputs](https://developers.openai.com/api/docs/guides/structured-outputs)）。按文档推断会被拒，未实测 |
| 有 OutputSchema、有工具 | Butter 默认用 Gemini API backend（[client.go L222-L249](https://github.com/googleapis/go-genai/blob/v1.63.0/client.go#L222-L249)），因此走 `set_model_response` | ResponseSchema 与工具一起发送，问题同上一行 |
| `ResponseJsonSchema` | genai 支持一套更完整的 JSON Schema 子集：含 `$defs/$ref/anyOf`，`oneOf` 按 anyOf 处理，支持 `additionalProperties`；环引用只做有限展开，而且只能用在非必填属性上（[types.go L2829-L2844](https://github.com/googleapis/go-genai/blob/v1.63.0/types.go#L2829-L2844)）。但 ADK Go 不用这个字段 | 适配器不读取这个字段 |

**provider 限额备查。**

- OpenAI strict：支持 string/number/boolean/integer/object/array/enum/anyOf，以及 `$defs` 和递归。根节点必须是对象，且不能是 anyOf。上限为 5000 个属性、10 层嵌套；属性名、定义名和枚举值的总长不超过 12 万字符；枚举值总数不超过 1000 个。不支持 `allOf/not/if-then-else/dependent*`（出处同上）。
- Gemini：支持 anyOf 和 `$ref` 递归；"过大或嵌套过深的 schema 可能被拒"；输出"在语法上是正确的 JSON，但仍须在应用中校验取值"（[structured-output](https://ai.google.dev/gemini-api/docs/structured-output)）。

## 4. 对 Butter 的判断（本节是仓库判断，事实依据见上文）

先列 Butter 的现状：

- 只读 catalog 有 7 个组件，没有函数调用和模板；DynString 只允许字面量或 `{path}`（`internal/a2ui/catalog.go:50-84, 188-206`）。
- 服务端负责 createSurface 和各种限额（`internal/a2ui/card.go:29-34, 86-155`）。
- catalog 是服务端内置的，客户端只做选择，所以"每次运行协商出不同 schema"目前不是问题。
- 每个 LLM agent 都挂着 `aguitool` 和 `a2uitool` 两个 toolset（`internal/agent/agent.go:162-176`）。
- proto 里已有 `output_schema_json` 字段，但没有接线（`proto/agents/v1/agent.proto:178-181`）。
- 运行使用 SSE 流式（`internal/runtime/runner/runner.go:1030,1044`），文本增量直接变成 `TEXT_MESSAGE_CONTENT`（`internal/handler/http/agui_sink.go:164-169`）。

| 方案 | 有利 | 不利 |
|---|---|---|
| (a) 在正文里输出 prompt-first JSON 再解析（官方样例的主路径） | 是官方主路径；可以增量渲染 | 原始 JSON 会进入正文、对话历史，以及 Telegram 等其他入口；与"先持久化再发送"冲突；纠错需要整轮重跑；prompt 要放全量 schema（官方约 4 万字符，另加示例） |
| (b) 工具调用 + 校验 + 把错误回给模型（现状） | 与官方 toolset、AG-UI `render_a2ui` 同型；ADK 会自动把错误回给模型；工具只在协商成功的运行中出现；提示约 2.2K 字符 | 参数是开放对象，provider 不做约束；字段描述被丢弃；Gemini 填 `{}` 的风险尚待验证；没有显式的重试上限 |
| (b') 同一个工具，换成从 catalog 派生的精确参数 schema | 7 个组件都是不递归的分支，用 anyOf + enum 就能表达；ADK 会先用 jsonschema-go 校验，返回带路径的错误；枚举值成为模型看得见的约束 | 两个适配器都不开 strict 或 VALIDATED，得不到约束解码。要开 strict 须修改 adk-utils-go，而且任意结构的 `updateDataModel.value` 必须改成键值对数组——这正是 v0.9 有意放弃的 v0.8 形状。Gemini 的 VALIDATED 模式作用于整个请求里的所有工具（包括 MCP 工具） |
| (c) 独立的 UI composer agent + OutputSchema | adk.dev 本身就建议"用子 agent 单独做输出格式化" | genai.Schema 表达不了 A2UI；Butter 必挂 toolset，在 Gemini API 上必走 `set_model_response`；ADK 校验器不支持 anyOf；OpenAI 适配器的 strict 缺 `additionalProperties`；多一次模型调用；composer 输出的 JSON 会经 SSE 流进对话，需要在 sink 里隐藏；引用、root 和限额仍然要由服务端校验 |
| (d) 服务端模板，模型只填一个小 schema | 是 AG-UI fixed schema 和 A2UI macros 的官方路线；平面 schema 落在所有 provider 的 strict 子集之内；版式确定、提示最小、校验简单；Butter 的表单已经是这种做法 | 表达力受模板限制；需要维护模板版本和历史渲染（即 #381 的方案 C） |

**直接回答：** 生成 A2UI 不需要结构化输出。

- 保留 (b)。
- 如果卡片的一致性不够，下一步做 (d)：用工具参数承载模板数据（必要时开 strict），而不是 ADK 的 `OutputSchema`。
- (b') 只在实测证明宽松参数会被模型填坏时再做。
- 如果将来确实要用 OutputSchema，只在没有工具、没有 transfer 的叶子 agent 上配平面 schema，并自行校验 OutputKey 里存的字符串。

**建议的验证顺序：**

1. 把 `jsonschema_description` 改成 `jsonschema`（或者显式设置 `InputSchema`），确认工具声明里出现了字段描述。**已完成：#374。**
2. 分别在真实 Gemini 和 OpenAI 上跑一组 `render_ui`，统计参数被填成 `{}`、缺少 root、出现悬空引用的比例；并为单次 invocation 内的连续失败设一个上限。**上限已在 #374 加上，测试工具在 #380；OpenAI 已测（§5.1），Gemini 暂未测。**
3. 如果需要更稳定的版式，加 1–2 个模板工具（例如"摘要卡片：title + 键值列表 + 状态"），与现状对比使用率和失败率。**待定，见 #381。**

## 5. 实测：gpt-5.6-luna

### 5.1 `render_ui` 调用质量（2026-10-02）

第 1 步已经在 #374 完成（字段描述进入工具声明，连续失败 3 次后停用）。之后用 `internal/a2uitool/probe_test.go`（#380）跑了一次真实模型：

- **环境：** ADK v2.5.0。经 Butter 自己的 `NewFromProto` 构建 agent，provider 类型 `openai`，即 adk-utils-go v0.22.0 适配器。模型 `gpt-5.6-luna`，走 OpenAI 兼容端点，SSE 流式。
- **场景：** 5 个场景各 3 次，共 18 轮。4 个单轮卡片场景（部署摘要、任务状态、套餐对比、值班），外加 1 个两轮场景：先建事故卡片，再更新它。agent 指令明确鼓励用卡片。

| 指标 | 结果 |
|---|---|
| 调用了 `render_ui` 的轮数 / 显示出卡片的轮数 | 18 / 18 |
| `render_ui` 调用 / 首次即合法 | 18 / 18 |
| 含 `{}` 的消息、`messages` 不是数组、触发失败上限 | 0、0、0 |
| 第二轮更新复用了同一张卡片 | 3 / 3，只发送变化的组件，并带上 `surface_id` |
| 有文字回答的轮数 | 18 / 18 |

- 用到的组件：KeyValue 69 次、Column 18、Text 18、Card 15、Status 12。没用过 Row、Divider，也没用 `{path}` 数据绑定。全部消息都是 `updateComponents`。
- 每轮耗时 15–47 秒，套餐对比最长。每轮两次模型调用：一次工具调用，一次文字回答。

**结论：** 对这个模型，宽松的 `messages` 参数没有问题，不需要改成 JSON 字符串参数，也不需要精确的 catalog schema。这次也实际验证了 #373 升级后（openai-go 3.64）adk-utils-go 的流式工具调用。Gemini 没有测，"填成 `{}`"的风险仍然只来自 AG-UI 适配器的代码注释。

### 5.2 卡片策略的 presentation：AUTO 与 PREFERRED（2026-10-06，#442）

卡片策略（ADR-0014 的 Card Policy 修订）设为 `PREFERRED` 时，只在 `render_ui` 工具描述第一段的末尾加一句 `a2uitool.PreferredHint`：

> Prefer a card: whenever your answer has structured results (key facts, a status, a list of results), show them in a card as well as in text.

`AUTO` 下的描述与之前逐字节相同。为了只看这句话的作用，用 `internal/a2uitool/probe_test.go` 里的 `TestRenderUIPresentationProbe` 做了一组对照：

- **环境：** 与 §5.1 相同（ADK v2.5.0、Butter 的 `NewFromProto`、provider 类型 `openai`、`gpt-5.6-luna`、SSE 流式），4 轮并行。
- **指令：** 中性，不提卡片："You are an operations assistant in a chat app. Answer the user's questions briefly." 两组之间唯一的差别就是那句提示。
- **场景：** 8 个"边界"单轮提示，用文字或卡片回答都说得过去：SLA 要点、三次构建耗时、服务所在区域、磁盘增长趋势、三个待处理 PR、密钥轮换步骤、liveness 与 readiness probe 的区别、p95 的含义。每个提示在 `AUTO` 和 `PREFERRED` 下各跑 5 次，共 80 轮。

| 场景 | AUTO | PREFERRED |
|---|---|---|
| open-prs（三个 PR，各有作者和状态） | 0/5 | 3/5 |
| 其余 7 个场景 | 0/35 | 0/35 |
| **合计（显示出卡片的轮数）** | **0/40** | **3/40** |

- 两组都没有运行错误，40/40 轮都有文字回答。文字长度的中位数分别是 218 和 200 个字符。
- `PREFERRED` 下的 3 次 `render_ui` 调用都首次合法，都新建了一张卡片：一个 Card，里面一个 Column、一个标题 Text，每个 PR 一个 KeyValue。
- 每轮耗时的中位数：`AUTO` 9.0 秒，`PREFERRED` 10.1 秒。出卡片的 3 轮要 13–21 秒，因为多了一次模型调用。
- 在这个样本量下，3/40 对 0/40 并不显著（Fisher 精确检验，单侧 p≈0.12）。

**结论：** 对这个模型，`PREFERRED` 是很弱的提示。指令不提卡片时，`AUTO` 在这些边界提示上一张卡片也没出；`PREFERRED` 只让最像列表的那个提示有时出卡片，键值事实、步骤和概念解释仍然只用文字。这符合设计：提示从不强制卡片，也不进入 instruction。但如果作者希望某个 Agent 稳定地出卡片，只靠 `PREFERRED` 不够。作为参照，§5.1 里指令明确要求卡片、提示本身也更适合卡片时是 18/18。

## 未核实或互相矛盾之处

- 实测只覆盖 `gpt-5.6-luna`（§5），没有测 Gemini。以下两点仍只来自文档和代码注释：Gemini 对 `additionalProperties: true` 的开放对象数组是否同样会填 `{}`；OpenAI 是否会拒绝适配器生成的 strict schema。
- §5.2 的对照每组只有 40 轮，`PREFERRED` 的效果有多大（3/40）还不能外推；换模型或换提示句之后要重测。
- "A2UI SDK 丢弃了组件校验结果"是读代码（`102ec1a`）得出的，没有执行样例验证。
- Gemini 3 能否在 Gemini API 上组合结构化输出与函数调用：Gemini 文档和 adk.dev 说可以，而 ADK Go（v2.1.0 和 main）与 adk-python main 的代码都按不可以处理。
- a2ui.org 内部也有不一致：首页和 overview 对 v1.0 的摘要（`actionResponse`、`surfaceProperties`）与 v1.0 规范不符；Agent Development 指南与当前 SDK 不符；介绍页把生成方式写成 "structured output"，与规范的 prompt-first 不一致。
- eval 基线只是一个模型、一个 epoch、温度 0 下的一次运行，96.1% 这个数字不能外推。
