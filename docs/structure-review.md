# 目录结构评审

评审时间：2026-04-12

## 结论

当前项目的主干分层是合理的，`cmd`、`internal`、`proto`、`pkg/proto` 这几个层次清晰，整体符合 Go 服务项目的常见组织方式。

不过随着功能继续增加，目录职责边界已经开始出现一些模糊点，尤其体现在启动装配、RPC 实现命名、以及 runtime 相关模块的摆放上。当前结构还能继续维护，但建议尽早做一次小规模整理，避免后续演化成“目录名存在，但职责不稳定”的状态。

## 当前结构的优点

- `cmd/butter` 作为唯一入口，职责清晰。
- `proto/agents/v1` 与 `pkg/proto/agents/v1` 区分了定义和生成物，边界明确。
- `internal/channel`、`internal/agent`、`internal/runner`、`internal/cron` 这些核心能力已经按功能拆开，没有全部堆在一个目录里。
- `internal/session/mongo`、`internal/memory/mongo` 这种“接口语义 + 存储实现”的分层方向是对的。

## 主要问题

### 1. `internal/bootstrap` 过重

`internal/bootstrap/channels.go` 当前同时承担了以下职责：

- MongoDB 初始化
- Redis 初始化
- Langfuse plugin 初始化
- runner service 构建
- channel manager 构建与启动
- cron scheduler 构建与启动
- system agent 注册

这已经不是单纯的 bootstrap glue code，而是一个集成式启动中心。文件继续增长后，会让依赖关系和失败路径都变得更难理解。

### 2. `internal/service/configapi` 命名不准确

这个目录里的实现本质上是 Twirp server/transport adapter，而不是纯领域 service。

例如：

- `AgentServiceServer`
- `CronJobServiceServer`
- `SessionServiceServer`

这些类型本质上都是协议层入口。继续放在 `internal/service` 下，后续会让人误以为它们是业务服务，而不是 RPC 层实现。

### 3. `internal/repo/configstore` 语义偏移

`configstore.Store` 当前是线程安全的内存 CRUD store，不是典型 repository，也不是外部存储访问层。

如果以后真的引入数据库版 config repo，那么 `repo/configstore` 这个命名会变得更混乱：到底它是 repo，还是内存缓存，还是配置态 state store。

### 4. runtime 相关目录较平铺

目前以下目录是平铺在 `internal/` 下的：

- `runner`
- `cron`
- `session`
- `memory`

从职责上看，它们都属于“运行时基础设施”。随着后续增加 queue、event、workflow、scheduler 之类模块，`internal/` 根目录会越来越拥挤。

### 5. `channel` 平台目录可能重复膨胀

`internal/channel/telegram` 和 `internal/channel/discord` 下已经各自有：

- `poller`
- `selector`
- `status`
- `clear`
- `debug`
- `photo`

说明平台间已经出现明显的结构对称。现在规模还可控，但之后如果再加 Slack、Feishu、Discord webhook 等，重复实现会继续扩散。

## 调整建议

## 一、先做低风险重命名

### 1. `internal/service/configapi` 改为协议层目录

建议迁移为以下之一：

- `internal/transport/twirp`
- `internal/handler/twirp`

更推荐 `internal/transport/twirp`，因为它和 `internal/handler/http` 是并列关系更自然：一个是 HTTP handler，一个是 Twirp transport。

### 2. `internal/repo/configstore` 改为 store 语义

建议迁移为：

- `internal/store/config`
- 或 `internal/config/store`

更推荐 `internal/store/config`，语义最直接，后续如果补充 memory store、snapshot store、runtime store 也容易扩展。

## 二、拆分 bootstrap

建议把现在的 `internal/bootstrap` 收敛成“编排入口”，不要继续堆实现细节。

例如可拆成：

- `internal/app/routes.go`
- `internal/app/wire_runtime.go`
- `internal/app/wire_channels.go`
- `internal/app/wire_cron.go`

如果暂时不想改目录名，至少可以先在 `internal/bootstrap` 内做文件拆分：

- `routes.go`
- `runtime.go`
- `channels.go`
- `cron.go`
- `system_agent.go`

目标是让每个文件只负责一类初始化逻辑。

## 三、为 runtime 建立上层目录

建议中期整理为：

```text
internal/
  runtime/
    runner/
    cron/
    session/mongo/
    memory/mongo/
```

这样做的好处：

- `internal/` 根目录更干净
- runtime 基础设施一眼能归类
- 后续新增 execution、queue、eventbus 时有统一归属

## 四、收敛 channel 平台共性

可以考虑新增：

```text
internal/channel/
  common/
  telegram/
  discord/
```

把以下可复用概念上提：

- selector 接口
- debug/status/clear 这类控制命令抽象
- 通用消息格式转换辅助

平台目录只保留与 SDK、API、事件模型直接相关的实现。

## 五、细化 `internal/agent`

当前 `internal/agent` 下有：

- `agent.go`
- `model.go`
- `system/`

如果后续 agent 工厂逻辑继续增长，可以考虑演化为：

```text
internal/agent/
  factory/
  model/
  system/
```

目前还不算必须，但已经有这个趋势。

## 推荐的目标结构

下面是一个更适合后续扩展的目录组织方式：

```text
cmd/
  butter/

internal/
  app/
  agent/
    factory/
    system/
  channel/
    common/
    telegram/
    discord/
  config/
  runtime/
    runner/
    cron/
    session/
    memory/
  service/
  store/
    config/
  transport/
    http/
    twirp/

proto/
  agents/v1/

pkg/
  agent/
  proto/
```

## 建议的实施顺序

为了降低重构风险，建议按下面顺序做：

### 第一阶段：纯目录和命名整理

- `internal/service/configapi` -> `internal/transport/twirp`
- `internal/repo/configstore` -> `internal/store/config`

这一步不改逻辑，只调整 import 和命名，风险最低。

### 第二阶段：拆分 bootstrap

- 拆分 `internal/bootstrap/channels.go`
- 让入口只负责 orchestrate，不承担所有构建细节

### 第三阶段：整理 runtime 和 channel 共性

- `runner/cron/session/memory` 收敛到 `internal/runtime`
- 将 channel 平台之间重复的共性抽象上提

## 额外观察

- `README.md` 中对 channel 的描述仍偏向 Telegram，但代码里已经有 Discord，文档和实现略有漂移。
- 根目录存在 `.kilocode/`，而当前仓库环境约定建议使用 `.kilo/`，建议统一，避免工具配置来源混乱。
- 根目录同时存在 `config.yaml` 和 `config/`，如果后续配置入口继续增加，建议统一约定默认配置位置。

## 总结

这个仓库的目录结构当前不算乱，问题主要不是“已经失控”，而是“已经能看出未来膨胀点”。

最值得优先处理的不是大规模重构，而是两个低风险动作：

1. 把 `configapi` 从 `service` 里挪出去
2. 把 `configstore` 从 `repo` 语义里挪出去

做完这两步后，再拆 bootstrap，整体结构就会清楚很多，也更适合继续扩展。
