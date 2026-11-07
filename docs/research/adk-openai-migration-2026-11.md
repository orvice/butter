# ADK 原生 OpenAI / Responses API 迁移复核

调研与发布后复核：2026-11-02。升级前 Butter 基线：`1961a3930dcaf511052345c6b3177f15ce6477f2`。

## 结论（v2.7.0 发布后更新）

**可以将 Chat Completions 和 Responses API 都统一到 ADK 原生 `model/openaimodel`。不再需要等待媒体补丁的正式版本。**

用户提醒后重新核对：**v2.7.0 已于 2026-11-02 10:37:02 UTC 发布**，明确包含图片／文件输入 [PR #1543]。tag 对应提交正是此前验证过的 **`be8b100d0006ff9d9bb2a185a54deeb7f5b7a047`**。首次调研时“最新稳定版 v2.6.0 不包含媒体支持、需要等待”的判断已被这次发布取代。

已先将 Butter 的依赖从 ADK **v2.5.0 升级至 v2.7.0**，随后按用户确认完成原生适配器切换和 Provider 协议配置。

落地方案：

1. 两种协议都使用原生适配器，不维护第三方／原生两套 OpenAI 模型实现。
2. Provider 新增 `openai_api`，新建／编辑页可选，列表显示；旧记录未设置时仍使用 Chat Completions。
3. Responses 必须显式选择；不自动探测或失败后切换协议。真实目标 provider 仍需独立验收。

来源：[v2.7 release]、[v2.7 module origin]、[PR #1543]、[原生构造器]。

## 1. 当前实现与迁移接缝

- 升级后 `go.mod`：ADK `v2.7.0`、genai `v1.75.0`、openai-go `v3.82.0`。Butter 的 Go `1.27.1` 已满足 ADK 的最低 Go `1.26.10` 要求。
- `internal/agent/model.go` 已统一创建 `openaimodel.NewModel`，显式传递所选 API。空配置覆盖 ADK 的默认 Responses，保留 Butter 的 Chat Completions 默认行为。
- `proto/agents/v1/agent.proto` 的 `ModelProvider.openai_api` 接受 `chat_completions` / `responses` / 空值，仅适用于 `type="openai"`；未知值和非 OpenAI 类型的非空值在保存和构建时拒绝。
- `internal/userinput/userinput.go` 的统一入口接收文字和 JPEG/PNG/GIF/WebP 内联图片，产物为 `genai.Part.InlineData`。原生媒体转换必须覆盖新图片和历史图片，不能只验证纯文字。
- `internal/agent/agent.go:newLLMAgent` 通过 `model.LLM` 消费模型。构造点替换无需重写 runner、AG-UI 或 session 数据模型，但仍须验证协议行为和旧会话续跑。
- `adk-utils-go v0.22.0` 还承载 Langfuse 和 ContextGuard；替换 OpenAI 适配器不等于可删除整个依赖。

这不涉及 Butter 对外的 OpenAI 兼容 HTTP API：`internal/handler/http/openai.go` 仍只提供 models 和 Chat Completions。更换出站模型协议不会自动增加入站 `/responses`。

来源：上述 Butter 文件、[当前第三方适配器]、[v2.7 go.mod]。

## 2. 版本与功能对照

| 能力 | 原第三方 v0.22.0 | 原生 v2.5.0 | 原生 v2.6.0 | 原生 v2.7.0 |
| --- | --- | --- | --- | --- |
| Chat Completions | 是 | 是，显式选择 | 是，显式选择 | 是，显式选择 |
| Responses | 旧包没有 | 是，默认 | 是，默认 | 是，默认 |
| 普通多轮文字 | 支持 | 支持 | 支持 | 支持 |
| `InlineData` 图片 | 支持 | **拒绝** | **拒绝** | **已转换** |
| `InlineData` PDF 等文件 | 支持部分 MIME | **拒绝** | **拒绝** | **已转换**，服务端仍须支持具体格式 |
| ADK function tools、结果回填 | 支持 | 支持 | 支持 | 支持 |
| 流式文字／工具调用转换 | 支持 | 有实现及上游测试 | 有实现及上游测试 | 有实现及上游测试 |

表中“支持”指适配能力，不代表已通过生产 provider 验收。

### 多轮问题早已修复

2026-08 的旧调研指出 assistant 历史被写成 `input_text`。issue #1197 已于 2026-08-31 关闭；v2.3.0 源码已经有 `newOutputMessage` 和多轮回归测试。不能继续沿用“原生适配器第二轮必然 400”的旧结论。

来源：[issue #1197]、[v2.3 历史转换]、[Responses 请求转换]。

### 媒体补丁在 v2.7.0 首次进入正式版

v2.5.0 / v2.6.0 的两种 API 路径都会在 HTTP 前拒绝 `InlineData` / `FileData`。历史中的图片也会参与后续请求转换，所以续问是纯文字并不能绕过旧版缺口。

v2.7.0 已实现：

- Responses：内联图片 → `input_image` + base64 data URL；文件 → `input_file`；已上传文件 ID、图片／文件 URL 也有转换。
- Chat Completions：内联图片及图片 URL → `image_url`；内联文件和已上传文件有转换。
- detail 固定 `auto`；转换支持不代表模型支持任意 MIME、`gs://` URI 或输入尺寸。
- 两条原生路径仍拒绝 audio/video；Chat Completions 不支持图片 file ID、非图片文件 URL；不支持的角色／工具结果位置会明确拒绝媒体。

Butter 统一用户入口目前只接受文字和图片，所以第三方 inline audio 能力不是该入口当前的产品承诺；其他工具／artifact 来源仍需单独核对。

来源：[PR #1543]、[媒体分类]、[Responses 请求转换]、[Chat 请求转换]、`internal/userinput/userinput.go`。

## 3. 协议选择和行为边界

### 显式选择，旧配置保持语义

原生 `ClientConfig.API` 接受 `APIChatCompletions`、`APIResponses`，**零值默认 Responses**。只换构造器且不填写 API 会悄悄改变所有兼容 provider 的请求路径。

已新增 provider 级协议字段，旧记录未设置时解析为 Chat Completions。同一 endpoint 的不同模型如需不同协议，可拆成两个 provider，不引入模型级覆盖。Model override 构建时保留 alias，避免多个 Provider 的相同实际 Model ID 混淆 endpoint、凭据或协议；已有回归覆盖此场景及 warmed cache。

不按域名、模型前缀或一次 404 猜测协议，也不在失败后自动换协议重试：请求可能已经被接受并产生费用或服务端副作用。

来源：[原生构造器]、Butter `ModelProvider` 定义。

### 工具仍由 Butter / ADK 执行

MCP、memory tools、`render_ui` 等继续走 ADK function tools。原生适配器支持 declarations、call ID、结果回填；Responses 的工具参数显式 `strict:false`，Chat Completions 使用非 strict 默认值。不能宣称改用 Responses 后自动获得严格参数保证。

也不会自动获得 OpenAI 托管的 web search、file search、remote MCP 或 Codex：该转换器只接受 function tools，没有这些原生 Responses 工具的通用映射。

来源：[工具转换]、[原生文档]、Butter `internal/agent/agent.go`。

### 推理控制有映射，但不保留原始 reasoning 状态

原生包能映射 `ThinkingConfig` 到 reasoning effort，Responses 可以请求 summary。但历史 thought／reasoning 被丢弃，没有回传原始 reasoning item / encrypted content，也没有使用 `previous_response_id`。它仍通过 ADK 历史重建请求，不等价于 OpenAI 推荐的完整 reasoning-state continuation，不能据此承诺质量、延迟或 token 成本必然改善。

来源：[原生文档]、[Responses 请求转换]、[OpenAI 迁移指南]。

### 配置、存储和凭据

- 原生包对无法翻译的 genai 配置明确报错。Responses 不映射 stop sequences、frequency/presence penalty、seed；Chat Completions 可以映射；两者均拒绝 top-k。
- strict JSON 输出 schema 会补全 required、关闭 additionalProperties；不要把 `render_ui` 的松散参数直接改成 strict output schema。Butter 当前 LLM 构造器也没有配置这一输出 schema。
- 请求构建器不主动设置 `store=false`。官方说明 Responses 默认存储，新的 Chat Completions 账户也可能默认存储。须明确策略；可评估通过 `ClientConfig.Options` + `option.WithJSONSet("store", false)` 注入并做 wire test。禁用存储不等同于获得 ZDR。
- 空 key / base URL 会继承进程环境变量。自定义网关要有显式凭据边界，避免进程级 OpenAI key 被发往不相关 endpoint。
- `openaimodel` 仍标为 experimental，应固定版本并保留自己的协议回归测试。

来源：[原生文档]、[原生构造器]、[Chat 请求转换]、[Responses 请求转换]、[schema 处理]、[OpenAI 迁移指南]、[SDK JSON option]。

## 4. v2.5.0 → v2.7.0 的框架升级影响

不只模型包发生变化；`go mod tidy` 同时按 ADK 约束更新 genai、OpenAI SDK、OTel、AWS SDK 和 Google API 等传递依赖，没有执行全量 `go get -u`。

- **模型调用上限**：`agent.RunConfig{}` 从无该限制变为默认每次 invocation 最多 500 次模型调用，超出返回 `agent.ErrLLMCallsLimitExceeded`。Butter 现有 runner 未覆盖此字段，采用新默认；可用 `ADK_MAX_LLM_CALLS` 调整，`-1` 禁用。此处不擅自改成无限制。
- **遥测**：默认切到新版 GenAI semconv，事件及 span 内容格式和根层级改变。采集链路依赖旧格式时可设置 `ADK_TELEMETRY_SCHEMA_VERSION_OPT_IN=otel_semconv_1_36`，上游承诺至少支持至 2027-03-31；本轮不自动覆盖运维环境。两项变量已说明在 `.env.example`。
- **可选指令占位符**：仅 missing-state / missing-artifact 被当作空值，基础设施错误现在传播。Butter 的 Mongo state 已返回 `session.ErrStateKeyNotExist`，S3 artifact 的缺失路径已包装 `fs.ErrNotExist`，无需为这两个 missing sentinel 做补丁。
- **工作流与工具集**：跨越 v2.6.0，Loop/Sequential 会在子 agent 失败时停止、toolsets 每个 step 重评估；v2.7.0 还修复 join 恢复、Loop 取消、callback artifact/output_key 持久化等。应覆盖工作流暂停续跑、动态工具和 AG-UI 的回归。
- **并行 Human Input 图的兼容性**：全套测试暴露原 FIFO fixture 的两个输出终点。ADK v2.5.0 已要求最多一个终点产生输出，但旧版恢复时丢失已完成分支，掩盖了该图的问题。v2.7.0 恢复两个分支后正确返回 `ErrMultipleTerminalOutputs`。测试已改为 Join → 汇总 agent，并新增两个回答都保留、各节点只执行一次、第二个回答前不汇总的断言，连续 10 次通过。生产 runner 未做错误吞掉或图自动改写；已有确实会同时产生多个输出终点的并行图也应检查并显式汇合。互斥 Router 的多个终点不等于同时产生多个输出。来源：[v2.5 终点规则]、[v2.7 状态恢复]。
- **Langfuse**：第三方插件仍保留；源代码核对不等于已在真实 Langfuse 后端验收新版遥测。

来源：[v2.6 release]、[v2.7 release]、[RunConfig]、Butter `internal/runtime/runner/runner.go`、`internal/runtime/session/mongo/mongo.go`、`pkg/adkutils/s3_artifact.go`。

## 5. 验证

### 适配器请求级验证

独立探针在 `/tmp/butter-openai-research/probe/probe_test.go`，使用 `httptest` 捕获真实适配器的 HTTP 请求并返回固定 JSON / SSE，不读取生产 provider 配置或使用真实凭据。

覆盖三种适配器（迁移前第三方、原生 Chat、原生 Responses）× 六种场景：多轮文字、内联图片、历史图片的纯文字续问、PDF、function-call/result ID 回填、流式文字。媒体字节只用于转换探针，不代表通过服务端内容校验。

| ADK 版本 | 实际结果 |
| --- | --- |
| v2.5.0 | 18/18 断言满足，其中 6 个原生媒体场景验证的是 HTTP 前拒绝 |
| v2.6.0 | 18/18 断言满足，同样包含 6 个预期拒绝，不代表媒体可用 |
| 原 main 快照 `be8b100d` | 18/18 成功；两条原生路径都发送图片、历史图片和 PDF |
| **正式 tag v2.7.0** | **重新运行 18/18 成功**，无需 source replace |

正式版命令：`EXPECT_NATIVE_MEDIA=1 GOGC=10 GOMEMLIMIT=768MiB go test -p=1 -gcflags=github.com/openai/openai-go/v3=-l -modfile=tag.mod -mod=mod -vet=off -v ./...`，在临时 module 中执行。主线验证还执行了 `go test ./model/openaimodel/...`，原生包及 completions、responses、shared 共 4 个 package 通过。

默认编译 SDK 曾受内存限制被杀，最终仅对 SDK 根包关闭内联并使用较低 GC 阈值完成。这是功能验证，不是性能或生产构建资源验收。

日志：`/tmp/butter-adk27/wire-probes.log`；此前对照日志在 `/tmp/butter-openai-research/`。

### Butter 升级验证

执行仓库全套 Go 测试：`GOGC=10 GOMEMLIMIT=768MiB go test -p=1 -gcflags=github.com/openai/openai-go/v3=-l ./...`，**最终全部通过，退出码 0**；日志为 `/tmp/butter-adk27/tests-final.log`。首次运行发现上述 FIFO fixture 不符合单一终点规则；修正 fixture 并强化断言后，定向 `-run '^TestRun_WorkflowHumanInputFIFO$' -count=10` 通过，再重跑全套通过。`go mod verify` 也通过。

### 原生适配器与 Provider 配置开发后的回归

- `internal/agent/model_openai_test.go`：通过 Butter 的真实 provider resolver，验证默认／显式 Chat／Responses × 流式／非流式 × 历史图片／PDF，实际执行两次模型调用完成工具调用 → 结果回填 → 最终回答，检查分片参数、call ID、系统指令、实际 Model ID 和图片 data URL；另验证未知协议、错误 provider 类型和不自动回退。
- `internal/application/modelprovider_api_test.go`：CRUD 回显、保存前校验、空配置兼容及 reload 失败回滚。
- config repo conformance + Mongo JSON 测试：新字段持久化及旧记录缺字段解码。
- `internal/app/modelprovider_api_integration_test.go`：经真实 RPC service / ConfigRuntime reload，把已 warmed 的默认与 override runner 从 Chat 切到 Responses，再切回，保留同一会话的图片及两种协议的回答历史。
- `internal/runtime/runner/model_protocol_test.go`：不同 Provider 使用相同实际 Model ID 时，alias override 及其 cache 仍选择正确协议。原来的提前 alias → Model ID 行为会在此测试失败，现已修正。
- `front/e2e/model-provider-edit.spec.ts`：5 个场景通过，包括模型行编辑、两种协议保存回显、新建、切换 Gemini 清理及表单校验。容器 PIDs 限制使常规多进程浏览器启动不稳定，实际使用构建产物静态服务和单进程 Chromium，逐 case 独立执行；没有修改仓库 Playwright 配置。
- 全套 Go 测试再次通过，日志 `/tmp/butter-provider-all-go-final.log`。Frontend build、e2e TypeScript 检查、变更文件 ESLint、`buf lint` 通过。默认 Chat fake 改为按请求返回 JSON 或 SSE，原有 AG-UI model override / 标题回归也通过。

没有真实 OpenAI／兼容网关、生产 Mongo／Redis／S3 或 Langfuse 验收，不能把本地测试等同于生产环境已验证。

## 6. 真实部署仍需验收
- 真实目标 endpoint：单轮／多轮、图片后续问、流式与最终答案、分片工具参数、并行工具 ID 匹配、429/5xx/断流／空输出／max tokens 和 Stop。
- Butter：MCP、memory、A2UI `render_ui`、AG-UI frontend tools、Human Input resume、ContextGuard 和 Langfuse。
- 数据保留、Authorization 目标及旧会话续跑明确验证。
- 协议变更通过修改 provider 配置生效；发布回滚前验证旧版本能读已有会话，不在同一 turn 内自动双协议重试。

## 来源

[v2.7 release]: https://github.com/google/adk-go/releases/tag/v2.7.0
[v2.7 module origin]: https://github.com/google/adk-go/tree/v2.7.0
[v2.7 go.mod]: https://github.com/google/adk-go/blob/v2.7.0/go.mod
[v2.6 release]: https://github.com/google/adk-go/releases/tag/v2.6.0
[PR #1543]: https://github.com/google/adk-go/pull/1543
[issue #1197]: https://github.com/google/adk-go/issues/1197
[v2.3 历史转换]: https://github.com/google/adk-go/blob/v2.3.0/model/openaimodel/request.go
[v2.5 终点规则]: https://github.com/google/adk-go/blob/v2.5.0/workflow/scheduler.go#L656-L676
[v2.7 状态恢复]: https://github.com/google/adk-go/blob/v2.7.0/workflow/persistence.go#L137-L182
[当前第三方适配器]: https://github.com/achetronic/adk-utils-go/blob/v0.22.0/genai/openai/openai.go
[原生构造器]: https://github.com/google/adk-go/blob/v2.7.0/model/openaimodel/openaimodel.go
[原生文档]: https://github.com/google/adk-go/blob/v2.7.0/model/openaimodel/doc.go
[媒体分类]: https://github.com/google/adk-go/blob/v2.7.0/model/openaimodel/internal/shared/media.go
[Responses 请求转换]: https://github.com/google/adk-go/blob/v2.7.0/model/openaimodel/internal/responses/request.go
[Chat 请求转换]: https://github.com/google/adk-go/blob/v2.7.0/model/openaimodel/internal/completions/request.go
[工具转换]: https://github.com/google/adk-go/blob/v2.7.0/model/openaimodel/internal/responses/tools.go
[schema 处理]: https://github.com/google/adk-go/blob/v2.7.0/model/openaimodel/internal/shared/schema.go
[RunConfig]: https://github.com/google/adk-go/blob/v2.7.0/agent/run_config.go
[OpenAI 迁移指南]: https://developers.openai.com/api/docs/guides/migrate-to-responses
[SDK JSON option]: https://github.com/openai/openai-go/blob/v3.82.0/option/requestoption.go
