# Pi MCP Code Mode 在 Butter 中的可行性调研

调研日期：2026-11-10

## 结论

**可行，但不适合把 Pi 的 Code Mode 直接当作 Butter ADK 的工具层来实现。** Pi 1.5.0 已经内置 MCP 与 Code Mode：MCP server 的工具默认以 `codemode` exposure 暴露，模型通过 `codemode` JavaScript 脚本发现、组合、过滤并调用工具。Pi agent 当前在 Butter 中运行于 ButterBox 的 PiService / `pi --mode rpc` 会话内，因此这些 MCP server 必须在 ButterBox 的 Pi 配置或扩展中可见。

当前 Butter 的边界是有意设计成这样的：普通 LLM Agent 可以挂载 Butter 管理的 MCP server；`AGENT_TYPE_PI` 是绑定 ButterBox 和工作目录的叶子 Agent，写入时拒绝 `mcp_servers`、`mcp_server_ids`、skills、instruction 等字段。这个边界仍然合理，因为 Pi 的 MCP 连接发生在 Box 进程一侧，Butter 侧的 MCP 凭证和连接不能自动跨进程传给 Pi。

建议分两阶段：

1. **短期不在 Butter Runner 中复制 Code Mode。** 在 ButterBox 侧用项目级 `.pi/mcp.json` 或用户级 `~/.pi/agent/mcp.json` 配置 MCP；保留 PI agent 的“Box 负责工具和指令”的契约。该路径无需修改 Butter API，也不需要 fork Pi。
2. **若需要 Butter Dashboard 管理 MCP 并供 Pi 使用，再增加受控的 Box 配置同步能力。** 同步时按 workspace、ButterBox、working directory 隔离，凭证只在 Box 的安全存储或进程环境中落地；不要把 MCP token 放进 PiAgentConfig 或普通 Agent proto。

## Pi Code Mode 的实际行为

Pi 官方文档和本机安装的 `@earendil-works/pi-coding-agent` 1.5.0 描述了以下契约：

- MCP server 配置来自用户级 `~/.pi/agent/mcp.json` 和项目级 `.pi/mcp.json`；项目配置要求项目已经被信任。
- Pi 支持 streamable HTTP 和 stdio MCP server；当前文档明确拒绝 SSE transport。服务器通过 `/mcp` 或 `pi mcp` 命令管理，也可以由扩展调用 `pi.registerMcpServer()` 注册到当前会话。
- 每个 MCP 工具注册为 `mcp__<server>__<tool>`。server 的 exposure 可以是：`direct`（直接出现在模型工具列表）、`deferred`（由 tool_search 延迟加载）、`codemode`（默认，通过 Code Mode 脚本调用）或 `hidden`。
- `codemode` 是一个受限 QuickJS 脚本工具。脚本通过 `tools.<name>(args)` 调用其他工具，通过 `searchTools()`、`describeTool()` 或 `ALL_TOOLS` 发现工具；脚本输出才进入模型上下文。脚本可以并行调用、筛选大结果，还可以调用分类器和图像模型。
- Pi 会在 MCP server 连接后自动启用 Code Mode；也可以在设置中加入 `defaultTools: ["+codemode"]`。MCP 工具通过 Code Mode 调用时仍经过 Pi 的 tool pipeline，扩展的权限和 tool_call/tool_result 钩子继续生效。
- MCP 工具结果超过模型展示限制时，Pi 会截断文本给模型，但 Code Mode 脚本能收到完整结果并在返回模型前压缩或筛选。单个 Pi 进程默认最多两个 Code Mode worker，可通过 `PI_CODEMODE_MAX_WORKERS` 调整。
- Pi SDK 不会自动加载这些 CLI 内置扩展。若未来 Box 采用 SDK，需要显式加入 `createMcpExtension()`、`createCodemodeExtension()` 和 `createToolSearchExtension()`，并在 session 启动后绑定扩展。

来源：本机 Pi 文档 `docs/mcp.md`、`docs/codemode.md`、`docs/sdk.md`、`docs/cli.md`，以及 `@earendil-works/pi-coding-agent` 1.5.0 的 `package.json`。

## Butter 当前接线

Butter 已经具备 workspace-scoped MCP 资源：

- `MCPServerService` 提供 MCP server CRUD、探活、工具列表和 OAuth 流程。
- LLM Agent 的 `AgentConfig` 可以通过 `mcp_servers` 或 `mcp_server_ids` 挂载这些工具。
- `AGENT_TYPE_PI` 的 `PiAgentConfig` 只包含 `butterbox_id`、`working_dir`、Pi provider/model/thinking level 和运行时限。
- `internal/runtime/pibox` 通过 ButterBox 的 PiService 提交异步 turn、等待 cursor settlement，并将 Pi session 绑定到 Butter session；Pi 进程和工作目录在 Box 侧。
- 当前 PI policy 会拒绝 Pi Agent 上的 Butter instruction、MCP、skill、file mount、context guard、memory 和 remote-agent 字段。仓库文档与 `proto/agents/v1/agent.proto` 都把 Box 侧 `AGENTS.md` / `.pi/` 作为 Pi 行为配置位置。

来源：`butter/docs/adr/0011-pi-agents-backed-by-butterbox.md`、`butter/docs/research/pi-rpc-integration.md`、`butter/AGENTS.md`、`butter/proto/agents/v1/agent.proto`、`butter/config.yaml`。

## 可选方案

| 方案 | 可行性 | 适用范围 | 主要问题 |
| --- | --- | --- | --- |
| Box 工作目录 `.pi/mcp.json` | 高 | 单 Box / 单工作目录预配置 | Dashboard 无法统一管理；需要 Box 运维或文件同步能力 |
| Box 用户级 `~/.pi/agent/mcp.json` | 高 | Box 上多个项目共享只读 MCP | workspace 隔离和 token 生命周期更难管理 |
| Pi extension 调用 `pi.registerMcpServer()` | 高 | 动态按 session 注册、复杂策略 | 需要把扩展部署到 Box，并定义配置/凭证注入协议 |
| Butter 生成 Box 的 MCP 配置 | 中高 | Dashboard 管理 MCP 后供 Pi 使用 | 需要新 Box RPC、加密落盘、版本/回滚、workspace 与 cwd 校验 |
| 在 Butter Runner 里执行 Pi 的 Code Mode | 低 | 试图让 ADK 统一控制 Pi MCP | 需要重复 Pi 的 QuickJS/tool exposure/权限语义，且 MCP 连接仍在 Box；与现有架构重复 |
| 把 Butter MCP endpoint 直接暴露给 Pi | 中 | 让 Pi 连接 Butter 已有 MCP server | 需要短期授权、Box 到 Butter 网络、run/workspace 路由和 token 轮换；不能直接复用 Butter 的 ADK toolset |

## 推荐的最小实现

如果目标只是“Pi Agent 能使用 MCP Code Mode”，无需改 Butter：

1. 在 ButterBox 镜像或工作目录中安装/启用 Pi 1.5.0 的 MCP/Code Mode 内置扩展。
2. 将 MCP 配置放到 Box 的 `.pi/mcp.json`，使用 streamable HTTP 或 stdio；配置项目 trust。
3. 对敏感 server 使用 Box 的环境变量或 Pi MCP OAuth 存储，不把 token 写进 Butter 的 `PiAgentConfig`。
4. 用 Pi RPC 的 `get_state`、`get_entries` 和一次真实/确定性 smoke test 验证：server 连接、Code Mode 工具发现、工具结果回写和 session 重挂载。
5. 在 Butter 中只补充 Box 状态/能力探测，例如报告 MCP/Code Mode 是否已启用和版本，而不把工具定义复制进 Butter Agent proto。

如果目标是“Dashboard 中配置的 Butter MCP server 自动出现在 Pi Agent 中”，建议新建 Box 侧能力，而不是放宽 PI agent 的 `mcp_server_ids` 校验：

- Butter 解析 Agent 所引用的 MCP server，向 Box 下发不含明文 token 的配置引用或一次性授权；
- Box 负责把凭证写入 Pi 可读的安全位置，并按 `(workspace, butterbox, working_dir)` 隔离；
- Pi 启动/重建 session 时读取该配置，使用默认 `codemode` exposure；
- 配置变化带版本号并支持撤销，Box 删除或失效授权后，Pi 重新连接时不能继续使用旧凭证；
- 不支持的 transport（Pi 当前拒绝 SSE）在 Butter 配置校验或 Box capability probe 阶段明确报错；
- 运行时仍由 `pibox` 负责 turn 的生命周期、取消和持久化，Code Mode 只属于 Pi 内部工具调度。

## 与现有 ADR 的关系和风险

ADR-0011 记录“Pi deliberately has no MCP support”，但这个判断与当前本机 Pi 1.5.0 官方文档不一致，应该修订为“Pi 的 MCP/skills/instructions 由 Box 侧管理，Butter 不直接注入”。Pi 已支持 MCP 不代表 Butter 可以安全地把 workspace MCP 凭证直接传入 Pi；多租户隔离、secret rotation、Box 网络可达性和项目 trust 仍是主要工程风险。

Pi 的 Code Mode worker、QuickJS 沙箱、工具 exposure 和 MCP 连接生命周期都属于 Pi 私有运行时契约。Butter 只应依赖 Box 暴露的稳定 capability/configuration API，不应在 Go runner 中复刻这些实现细节。

## 验证范围

本次验证读取了 Butter 当前源码、proto、ADR、已有 Pi RPC 调研，以及本机安装的 Pi 1.5.0 官方文档和 SDK 文档。未连接真实 MCP server，未修改 Butter/ButterBox 代码，也未运行需要外部 provider 凭证的 Pi turn；因此“Box 配置同步”仍是设计建议，不是已验证实现。
