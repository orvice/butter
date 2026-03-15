# butter

基于 `butterfly.orx.me/core` 初始化的服务骨架，已经按常见业务分层拆好入口、配置、handler、service、repo。

## 项目结构

```text
.
├── .env.example
├── cmd
│   └── butter
│       └── main.go
├── config
│   └── butter.yaml
├── internal
│   ├── config
│   │   └── config.go
│   ├── handler
│   │   └── http
│   │       └── health.go
│   ├── repo
│   │   └── health.go
│   └── service
│       └── health.go
├── go.mod
└── go.sum
```

## 当前分层

- `cmd/butter`: butterfly 应用启动和依赖装配
- `internal/config`: 应用配置结构
- `internal/handler/http`: HTTP 路由注册和请求处理
- `internal/service`: 业务逻辑
- `internal/repo`: 数据访问抽象占位，方便后续接 Redis/MySQL/Mongo

## 本地运行

1. 准备环境变量

```bash
cp .env.example .env
export $(grep -v '^#' .env | xargs)
```

2. 安装依赖并启动

```bash
go mod tidy
go run ./cmd/butter
```

服务启动后可访问：

```bash
curl http://127.0.0.1:8080/ping
```

返回示例：

```json
{"service":"butter","message":"pong"}
```
