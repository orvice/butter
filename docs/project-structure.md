# 项目目录结构文档

更新时间：2026-04-12

```text
butter/
├── cmd/
│   └── butter/
│       └── main.go
├── config/
├── docs/
│   ├── api.md
│   ├── structure-review.md
│   └── project-structure.md
├── internal/
│   ├── agent/
│   │   ├── agent.go
│   │   ├── model.go
│   │   ├── model_test.go
│   │   └── system/
│   ├── app/
│   │   ├── channels.go
│   │   ├── cron.go
│   │   ├── routes.go
│   │   ├── runtime.go
│   │   └── system_agent.go
│   ├── application/
│   │   ├── agent_service.go
│   │   ├── cron_service.go
│   │   ├── mcpserver_service.go
│   │   ├── remoteagent_service.go
│   │   └── session_service.go
│   ├── channel/
│   │   ├── manager.go
│   │   ├── discord/
│   │   └── telegram/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   └── http/
│   ├── repo/
│   │   └── health.go
│   ├── runtime/
│   │   ├── cron/
│   │   ├── memory/
│   │   │   └── mongo/
│   │   ├── runner/
│   │   └── session/
│   │       └── mongo/
│   ├── service/
│   │   └── health.go
│   └── store/
│       └── config/
├── openspec/
│   ├── changes/
│   └── specs/
├── pkg/
│   ├── agent/
│   └── proto/
│       └── agents/
├── proto/
│   └── agents/
│       └── v1/
├── .github/
│   └── workflows/
├── .claude/
├── .codex/
├── .kilocode/
├── .env.example
├── buf.gen.yaml
├── buf.lock
├── buf.yaml
├── config.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
└── README.md
```

## 目录说明

- `cmd/`：服务启动入口。
- `internal/app/`：应用装配与初始化（路由、运行时、渠道、系统 Agent）。
- `internal/application/`：RPC 服务实现层。
- `internal/channel/`：渠道适配与渠道管理（Telegram、Discord）。
- `internal/runtime/`：运行时能力（调度、会话、记忆、Runner）。
- `internal/store/`：配置类存储实现。
- `proto/`：Proto 定义源文件。
- `pkg/proto/`：Proto 生成代码。
- `openspec/`：变更提案、任务和规范文档。
- `docs/`：项目文档。

## 维护建议

- 新增模块优先放在现有分层下，避免在 `internal/` 根目录继续平铺。
- `pkg/proto/` 为生成代码目录，手动变更应在 `proto/` 中进行后重新生成。
- 结构变更后同步更新本文件，保证文档与仓库一致。
