# Workspace Git 仓库与 Agent 内容管理 PRD

更新时间：2026-08-07

> **状态：Draft。** 本文记录已确认的产品边界、领域模型、用户流程、接口草案、迁移策略和验收标准。实现前允许调整字段编号和内部包名，但不得在未更新本文的情况下改变已经确认的行为语义。

## 1. 摘要

Butter 当前将完整 `Agent` 配置保存在 MongoDB 中。目标是在不把 Git API 放入 Agent 运行热路径的前提下，让 workspace 绑定一个 GitHub 或 GitLab 仓库，并将 Agent 的 description、prompt 等人工维护文本存入：

```text
agents/{agent-id}/
```

Git 是绑定 workspace 中 Agent 文本内容的唯一事实来源；MongoDB 保存运行配置、仓库绑定、同步状态、目录缓存和 last-known-good 内容快照。Runner 只读取 MongoDB 与内存中的有效配置，不在 session 或 invocation 期间调用 Git API。

这项能力同时推动 Agent 身份模型升级：Agent 使用 workspace 内唯一、不可变的 slug ID，所有 Agent 独立管理，父子及 workflow 关系使用 Agent ID 引用，不再嵌入完整子 Agent 配置。

## 2. 背景与问题

当前 Agent 以 `name` 作为 CRUD、runner、Channel、Cron、Automation 等功能的关联键，完整 proto 以 JSON 文本存入 `config_agents`。`sub_agents` 递归嵌入父 Agent，description、instruction 和运行参数由同一份 DB 文档管理。

这种模型存在以下问题：

- Prompt 缺少 Git 历史、review、diff 和标准协作流程。
- `name` 同时承担运行名称、展示名称和业务主键，重命名成本高。
- 嵌套 Agent 不能作为独立资源管理、授权和版本追踪。
- 将 Git 直接作为运行时读取源会使 session 受远端延迟、限流和故障影响。
- DB 与 Git 同时可写同一字段会产生无法可靠解决的双向同步冲突。

## 3. 目标

- Workspace 可绑定零个或一个 GitHub/GitLab repository location。
- 使用 Git 管理 Agent description、prompt 和 global prompt。
- 提供 provider-neutral 的目录浏览、文件读取和多文件提交接口。
- 默认直接 commit，可按绑定配置为 Pull Request / Merge Request 模式。
- 使用 DB cache 和 last-known-good 快照隔离 Git 与 Agent 运行热路径。
- Agent 使用稳定 slug ID，所有业务关联迁移为 Agent ID。
- Agent 成为独立实体，组合关系通过 ID 引用构建。
- 支持现有 Agent 分阶段设置 ID、迁移和按 workspace 导出到 Git。
- Git 或同步故障不能自动中断已经可运行的 Agent。

## 4. 非目标

- 第一版不把 model、MCP、Workflow graph 等运行配置迁移到 Git。
- 第一版不支持 Git 仓库内创建或删除 DB Agent 实体。
- 第一版不支持二进制 Agent 资源、Git submodule 或符号链接。
- 第一版不实现通用 Git IDE，也不允许修改绑定根目录外的源码和 CI 配置。
- 第一版不自动注册 Webhook。
- 第一版不在 session、invocation 或每轮 LLM 请求中访问 Git。
- 第一版不提供同一 Markdown 文件的自动文本合并。

## 5. 领域模型

### 5.1 Agent

每个 Agent 是 workspace-scoped 的独立实体：

```text
Agent
├── id                 immutable workspace-scoped slug
├── display_name       mutable UI label
├── type               LLM / Loop / Sequential / Parallel / Workflow
├── child_agent_ids    ordered references where applicable
├── config             model, MCP, runtime and workflow settings
├── lifecycle_status
└── workspace_id
```

约束：

- `id` 在 workspace 内唯一，创建后不可修改。
- `id` 用于 DB 主键、RPC 参数、Git 路径和业务关联。
- ADK 运行名称使用稳定的 Agent ID；`display_name` 不参与身份解析。
- 每个 Agent 最多被一个父 Agent 包含，整体关系构成森林。
- Agent 可独立管理和测试；成为子 Agent 不会自动成为 Channel、Cron、A2A 或 OpenAI API 入口。
- 删除后的 ID 默认保留，不允许创建另一个同 ID 的新实体。

### 5.2 Agent 关系

- LLM、Loop、Sequential 和 Parallel Agent 使用父侧有序 `child_agent_ids`。
- Sequential Agent 严格按 `child_agent_ids` 顺序执行。
- Workflow node 使用 `agent_id` 引用独立 Agent，不再按 name 引用。
- `parent_agent_id` 是反向查询结果，不作为第二份可写事实来源。
- 保存时验证引用存在、workspace 一致、单父约束和无循环。
- 删除仍被 Agent、Channel、Cron、Automation 等资源引用的 Agent 时返回冲突。

### 5.3 Workspace Repository Binding

一个 binding 表示 workspace 使用的 repository location：

```text
WorkspaceRepositoryBinding
├── workspace_id
├── git_host_id
├── repository
├── branch
├── root_path
├── write_mode          DIRECT_COMMIT | PULL_REQUEST
├── content_schema_version
├── encrypted_pat
├── observed_commit_sha
├── active_commit_sha
├── sync_status
├── webhook_status
├── last_synced_at
└── last_error
```

同一 repository location 可以被多个 workspace 绑定，不强制 root path 隔离。两个 workspace 指向相同实际路径且拥有相同 Agent ID 时，共享同一份 Git Agent Content；各自的 DB 运行配置仍然独立。

### 5.4 Agent Content

第一版目录协议：

```text
{root_path}/
└── agents/
    └── {agent-id}/
        ├── description.md
        ├── prompt.md
        ├── global-prompt.md
        └── *.md
```

固定映射：

| 文件 | Effective Agent 字段 | 规则 |
|---|---|---|
| `description.md` | `description` | 可选；缺失或空文件表示空描述 |
| `prompt.md` | `config.instruction` | LLM Agent 必填且非空；其他类型不参与运行 |
| `global-prompt.md` | `config.global_instruction` | 可选；缺失或空文件表示空值 |
| 其他 `*.md` | 无 | 可浏览和缓存，第一版不进入运行配置 |

目录中不保存 `id`、display name、model 等 DB 字段。Agent ID 由路径和 DB 实体共同确定，避免同一字段在 Git 与 DB 中双写。

### 5.5 Observed Revision 与 Active Revision

- Observed Revision 是 Butter 最近读取到的远端 HEAD。
- Active Revision 是最后一个通过完整校验并成功 reload runner 的 revision。
- 两者可以不同；新 revision 无效时继续运行 Active Revision。
- Invocation 记录本次使用的 `agent_config_commit_sha` 与 Agent display name 快照。

## 6. 内容所有权

### 6.1 未绑定 Workspace

- Agent 文本继续由 DB 管理。
- 现有 Agent CRUD 行为保持可用。

### 6.2 已绑定 Workspace

- Git 是 description、prompt 和 global prompt 的唯一事实来源。
- DB Agent 更新接口不得直接修改这些字段。
- DB 保存解析后的 active 内容快照和 repository cache。
- Effective Agent 由 DB 运行配置与 Active Agent Content 合并得到。

### 6.3 首次绑定

用户必须明确选择初始化模式：

- `EXPORT_CURRENT`：目标内容不能与待导出 Agent 路径冲突；将 DB 文本提交到 Git。
- `IMPORT_REPOSITORY`：只为 DB 中已存在且 ID 匹配的 Agent 导入内容。

不执行自动双向合并。共享根目录中的未知 Agent ID 被视为未认领目录，不阻止当前 workspace 同步。

### 6.4 解除绑定

- 将 Active Revision 的内容物化回 Agent DB 配置。
- 物化和 runner reload 成功后，才删除 binding 与 PAT。
- 不修改远端 Git 内容。
- 没有有效 active 快照时默认拒绝解绑。

## 7. 用户与权限

使用现有 workspace role：

| 操作 | owner | workspace admin | member | global admin |
|---|---:|---:|---:|---:|
| 查看 binding/status/cache | 是 | 是 | 是 | 是 |
| 浏览 Agent Markdown | 是 | 是 | 是 | 是 |
| 单独测试 Agent | 是 | 是 | 是 | 是 |
| 创建/更新 Agent 内容 | 是 | 是 | 否 | 是，需审计 |
| 配置或替换 PAT | 是 | 是 | 否 | 是，需审计 |
| 绑定/解绑仓库 | 是 | 是 | 否 | 是，需审计 |
| 接受 force-push 新基线 | 是 | 是 | 否 | 是，需审计 |
| Purge 未认领 Git 内容 | 是 | 是 | 否 | 是，需审计 |

Git commit 审计至少记录 Butter user ID、username、workspace ID、operation ID、请求基线 SHA 和实际父 SHA。

## 8. Git Host 与凭证

### 8.1 Git Host

- 支持 GitHub、GitLab 及其自托管部署。
- 系统管理员预先配置允许的 Git host 和 API base URL。
- Workspace 只能选择已允许的 host，不能提交任意 API URL。
- 该限制用于防止 SSRF 和内网探测。

### 8.2 PAT

- 第一版统一使用 Personal Access Token。
- 每个 workspace binding 独立保存 PAT，不共享凭证对象。
- PAT 使用部署级主密钥加密；DB 保存密文、指纹和更新时间。
- API 永不返回 PAT 明文。
- 绑定阶段探测 repository、branch、read/write 和 PR/MR 权限。
- PAT 失效时 binding 进入 `DEGRADED`，现有 Agent 继续使用缓存运行。

### 8.3 Provider Adapter

服务端使用 GitHub/GitLab provider API，不维护长期本地 clone。Provider adapter 至少提供：

- 读取 branch HEAD
- 列举 tree 和读取 blob
- 创建多文件 commit
- 创建 Pull Request / Merge Request
- 比较 revision 是否 fast-forward
- 获取 commit metadata

## 9. 目录与提交接口

### 9.1 读取

目录读取默认来自 DB repository cache，不在普通请求中访问 Git。

建议 RPC：

```text
GetWorkspaceRepositoryBinding
ListRepositoryEntries(path, revision)
GetRepositoryFile(path, revision)
GetRepositorySyncStatus
SyncWorkspaceRepository
```

读取响应包含：

- `commit_sha`
- `observed_commit_sha`
- `active_commit_sha`
- `is_active`
- entry path、kind、size、content hash

默认 revision 为最近观测到的远端快照；`revision=active` 用于查看运行时内容。显式刷新调用同步接口，不在 list/get 内隐式访问 Git。

### 9.2 写入

写入使用多文件 changeset：

```text
CommitRepositoryChanges
├── commit_message
├── base_commit_sha
└── changes[]
    ├── PUT(path, content)
    └── DELETE(path)
```

规则：

- 一组 changes 生成一个 Git commit。
- 默认 `DIRECT_COMMIT`；binding 可配置 `PULL_REQUEST`。
- 只能写 `{root_path}/agents/` 下的 Markdown。
- 提交前验证 resulting tree，不允许 Butter 主动提交无效 Agent Content。
- 分支 HEAD 已前进时采用目标路径 last-write-wins：基于最新 HEAD 创建 commit，无条件覆盖本次目标路径，保留其他路径变化。
- 不执行 force-push，不自动合并同一 Markdown 文件。
- 单文件写接口可以作为 changeset 的便捷封装。

### 9.3 内容限制

第一版默认限制：

- UTF-8 Markdown only
- 单文件 `256 KiB`
- 单 Agent `1 MiB`
- 单 workspace cache `20 MiB`
- 禁止绝对路径、`..`、符号链接和 submodule

限额可由服务端配置覆盖。超限内容可以存在于外部 Git，但不能进入 cache 或成为 Active Revision。

## 10. 同步与缓存

### 10.1 触发方式

- 用户手工配置 GitHub/GitLab Webhook。
- Butter 提供 callback URL、secret 和所需事件说明。
- Webhook 到达后异步触发对应 binding 同步。
- 提供手动同步 RPC。
- 定时对比远端 HEAD 作为 Webhook 丢失的兜底。

### 10.2 同步算法

1. 获取绑定分支 HEAD，记录 Observed Revision。
2. 读取受管 tree 和 Markdown blobs，写入 DB raw cache。
3. 只解析当前 workspace DB 中存在的 Agent ID；其他目录标记为 unclaimed。
4. 完整验证文件、Agent 类型、引用关系和内容限额。
5. 在单次发布操作中写入所有 Agent Content snapshots。
6. 构建 Effective Agents 并验证 runner 可加载。
7. 成功后原子切换 Active Revision 并 reload runner。
8. 失败时记录错误，保持现有 Active Revision。

同步任务按 binding 串行化，以 commit SHA 保证幂等。多实例部署必须使用共享 lease/lock，避免同一 binding 被并发发布。

### 10.3 运行时一致性

- Git API 不进入 session 或 invocation 热路径。
- Runner 只读取 active DB snapshot，并在内存中持有 Effective Agents。
- 已开始的 invocation 使用旧配置完成。
- 后续 invocation 使用新 Active Revision。
- Session 不永久锁定到旧 commit。

## 11. 状态与故障处理

### 11.1 Agent 生命周期

```text
PROVISIONING -> ACTIVE
ACTIVE -> DELETING -> DELETED
PROVISIONING/DELETING -> ERROR
legacy without complete IDs -> MIGRATION_REQUIRED
```

- 只有 `ACTIVE` Agent 进入正式 runner registry。
- `MIGRATION_REQUIRED` Agent 保留在 DB，但不能被调用或作为依赖使用。
- 父 Agent 的任一传递依赖缺少 ID 时，整个父入口进入 `MIGRATION_REQUIRED`。
- DELETED Agent 保留 tombstone，默认只能 restore，不能复用 ID 创建新实体。

### 11.2 跨 DB/Git 操作

创建、组合更新、解绑等操作使用 Saga 和 operation ID，不宣称 Mongo 与 Git 之间存在真实事务。

创建 Agent：

1. 生成或验证用户确认的 slug ID。
2. DB 创建 `PROVISIONING` 实体。
3. 将初始 description/prompt 作为单个 commit 写入 Git。
4. 同步、验证并生成 active snapshot。
5. 切换 `ACTIVE` 并 reload runner。

创建 LLM Agent 时必须同时提供有效的初始 prompt。

组合保存使用一个应用命令协调 DB patch 和 content changes，但底层所有权仍然分离：

```text
UpdateAgentConfiguration
├── agent_patch
├── content_changes
├── expected_agent_version
└── base_commit_sha
```

### 11.3 Git 故障

- PAT 失效、Git 服务不可用或仓库被删除时，binding 进入 `DEGRADED`。
- 禁止新 Git 写入，Agent 继续使用 last-known-good。
- 展示最后成功同步时间和快照年龄。
- 定时重试，并允许通过 NotifyGroup 告警。
- 新 binding 没有任何有效 snapshot 时不能激活 Agent。

### 11.4 无效 Revision

- Observed Revision 保留在 raw cache，便于用户查看和修复。
- Active Revision 不变。
- 同步状态返回按 path 和 Agent ID 定位的校验错误。
- 修复 commit 正常进入相同同步流程。

### 11.5 Force-push

- 检测到新 HEAD 不是当前 observed/active revision 的后代时，状态变为 `DIVERGED`。
- 不自动激活新基线。
- owner/admin 显式接受后，完整校验并设置新基线。

### 11.6 回滚

- 人工回滚必须形成新的 Git commit。
- 回滚操作恢复目标 revision 的受管 Agent Content，再经过正常同步和发布。
- 不允许长期仅在 DB 中激活旧 snapshot。

## 12. 删除、恢复与 Git 内容保留

- `DeleteAgent` 不删除 Git 目录。
- Agent 进入 `DELETED`，Git 内容成为 unclaimed content。
- `RestoreAgent` 恢复原实体，并重新解析保留的 Git 内容。
- 相同 Agent ID 默认不可重新创建。
- owner/admin 可显式执行 `PurgeAgentRepositoryContent`。
- Purge 前必须确认没有任何 workspace Agent 使用对应实际 repository path。

多个 workspace 共享同一路径时，任何内容更新都会在各 workspace 下一次成功同步后生效。共享表示共享文本内容，不表示共享 DB 运行配置或权限。

## 13. Agent ID 准备与迁移

### 13.1 兼容版本

先发布不改变运行行为的兼容版本：

```text
SetAgentID(agent_name, id)
GetAgentMigrationReadiness()
```

要求：

- UI 根据 display name 建议 slug，用户确认最终 ID。
- 服务端验证格式、workspace 唯一性和保留字。
- ID 一旦设置，不允许通过普通 Agent update 修改。
- 当前 runner 仍按 legacy name 工作。
- readiness 返回 `READY`、`MISSING_ID`、`CONFLICT` 和依赖问题。

### 13.2 V2 数据迁移

仓库当前没有正式 migration framework，因此新增独立迁移命令，不在普通服务启动中隐式执行不可逆改写：

```bash
butter-migrate agents-v2 dry-run
butter-migrate agents-v2 apply
butter-migrate agents-v2 verify
```

迁移内容：

1. 将具有 ID 的 legacy Agent 和嵌套 Agent 展开为独立文档。
2. 将父子关系转换为有序 ID 引用。
3. 将 Workflow node agent name 转为 Agent ID。
4. 将 Channel、Cron、Automation 等当前引用补写为 Agent ID。
5. 历史 Invocation 等记录保留 name 快照；可靠匹配时补写 ID。
6. 验证单父、无环、workspace scope 和 runner build。
7. 未准备完成的 Agent 标记为 `MIGRATION_REQUIRED`，不阻止其他 Agent 或 workspace 启动。

迁移观察期内保留 legacy name 字段和旧数据，便于回滚；稳定后再执行 cleanup migration，不维护长期双写。

### 13.3 Git 内容迁移

Agent V2 身份迁移与 Git binding 分开执行。每个 workspace 在准备 PAT 和 repository location 后单独选择：

- `EXPORT_CURRENT`
- `IMPORT_REPOSITORY`

只有完成初始 commit、回读、验证和 active snapshot 发布后，workspace 才切换为 Git-owned content。

## 14. API 变更范围

建议新增或调整以下 ConnectRPC surface。最终名称可在 proto 设计阶段细化。

### AgentService

- `SetAgentID`
- `GetAgentMigrationReadiness`
- Agent CRUD 改为以 `id` 定位
- `RestoreAgent`
- `UpdateAgentConfiguration`
- `WorkflowNode.agent_id`
- Channel、Cron、Automation 等关联字段改为 `agent_id`
- Invocation 增加 `agent_id`、`agent_display_name`、`agent_config_commit_sha`

### WorkspaceRepositoryService

- `GetBinding`
- `BindRepository`
- `UpdateBinding`
- `UnbindRepository`
- `ValidateBinding`
- `ReplacePAT`
- `GetSyncStatus`
- `SyncRepository`
- `AcceptRepositoryBaseline`
- `ListEntries`
- `GetFile`
- `CommitChanges`
- `RollbackContent`
- `PurgeAgentContent`

所有 workspace-scoped RPC 继续从 `X-Workspace-ID` 上下文派生 workspace，不信任请求体中的 workspace ID。

## 15. 可观测性与审计

必须记录：

- binding 创建、更新、解绑和 PAT 替换
- Webhook 验证失败
- observed/active SHA 变化
- 同步耗时、读取文件数、缓存字节数和校验错误数
- commit/PR/MR 创建结果
- Saga operation 状态和补偿结果
- force-push 检测与人工接受
- Agent migration readiness 和 migration result

建议指标：

- `repository_sync_total{provider,status}`
- `repository_sync_duration_seconds`
- `repository_cache_bytes{workspace}`
- `repository_active_revision_age_seconds`
- `repository_operations_total{kind,status}`
- `agents_migration_required{workspace}`

日志与 API 响应不得包含 PAT 或完整敏感 header。

## 16. 验收标准

### Agent 身份与关系

- 可为 legacy Agent 设置 workspace-unique slug ID，并检测冲突。
- V2 中所有业务关联使用 Agent ID。
- 每个 Agent 可独立 CRUD 和测试。
- 循环、跨 workspace 引用和多父引用被拒绝。
- 未设置 ID 的 Agent 进入 `MIGRATION_REQUIRED`，不会拖垮其他 Agent。

### Repository Binding

- owner/admin 可使用 PAT 绑定 GitHub、GitLab 或允许的自托管 host。
- 普通 member 无法修改 binding、PAT 或 Git 内容。
- PAT 不以明文出现在 DB 查询结果、日志或 RPC 响应中。
- 支持 manual Webhook 配置、手动同步和定时对账。

### 内容与缓存

- Agent 运行期间不调用 Git API。
- 目录 list/get 默认从 DB cache 返回，并附带 observed/active SHA。
- LLM Agent 缺少有效 `prompt.md` 时新 revision 不生效。
- 同一 changeset 生成一个 commit。
- 分支 HEAD 前进时只覆盖请求目标路径，不影响其他路径。
- 无效 revision 不替换 last-known-good。

### 运行时

- 已运行 invocation 在配置发布期间正常完成。
- 新 invocation 使用新 Active Revision。
- Invocation 可追溯到 Agent ID 和 commit SHA。
- Git/PAT 故障时，已有 Agent 继续使用缓存运行。

### 生命周期与迁移

- 创建 Agent 时 DB/Git 任一步失败都会留下可重试、可诊断状态。
- DeleteAgent 不删除 Git 内容，RestoreAgent 可恢复 tombstone。
- Unbind 成功前先将 active 内容物化回 DB。
- 可通过 dry-run/apply/verify 完成 V2 migration。
- 每个 workspace 可独立完成 Git 内容导入或导出。

## 17. 分阶段实施建议

### Phase 0：Agent ID 准备

- 新增可选 Agent ID、`SetAgentID` 和 readiness API。
- Dashboard 提供 slug 建议、冲突提示和迁移进度。
- 不改变当前 runner 行为。

### Phase 1：Agent V2 身份与关系

- 独立 Agent 文档和有序 ID 引用。
- 所有关联迁移为 Agent ID。
- 引入 lifecycle status、tombstone 和 Effective Agent resolver。
- 提供 migration dry-run/apply/verify。

### Phase 2：Repository Binding 与只读缓存

- Git host 配置、binding、PAT 加密和 provider adapters。
- 手动 Webhook 配置、同步 worker、repository cache。
- 目录 list/get、observed/active status。

### Phase 3：Git 写入与 Agent 生命周期 Saga

- changeset commit、默认 direct commit、可选 PR/MR。
- CreateAgent、UpdateAgentConfiguration、unbind 和 rollback Saga。
- last-write-wins 目标路径语义。

### Phase 4：Workspace 内容迁移与运维完善

- `EXPORT_CURRENT` / `IMPORT_REPOSITORY`。
- force-push baseline acceptance。
- NotifyGroup 告警、指标和审计视图。
- legacy 字段 cleanup migration。

## 18. 实现前仍需确定的参数

以下项目不改变产品语义，可在实现阶段通过配置或 proto 细化：

- Agent ID 的精确长度、保留字和 slug 正则。
- 定时对账默认周期。
- repository raw cache 的 revision 保留数量和清理周期。
- Git commit message 模板和 bot author 展示格式。
- Saga operation 的默认重试次数与超时。
- `PULL_REQUEST` 模式的 branch 命名规则。
