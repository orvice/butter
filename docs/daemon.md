# Butter Daemon

Butter daemon 是一个反向连接的远程执行面。`cmd/butter` 在主 HTTP 服务的 `/api`
ConnectRPC 入口中挂载 daemon connector，`cmd/butter-daemon` 主动连接该入口、注册 runtime，然后等待服务端下发
daemon-backed RemoteAgent 的任务。

当前实现的核心约束：

- 服务端 daemon worker 入口与 dashboard API 共用主 HTTP 端口，路径为
  `/api/agents.v1.DaemonConnectorService/Connect`。
- daemon token 由服务端签发，类型为 `API_TOKEN_KIND_DAEMON`，scope 为 `daemon:connect`。
- token 绑定 workspace 和 `daemon_runtime_id`，服务端以 token 中的值为准，不信任 daemon 自报的 workspace/runtime。
- 同一个 workspace 下同一个 `DaemonRuntime` 同时只允许一个 daemon 连接。
- daemon-backed RemoteAgent 需要选择一个 `DaemonRuntime` 和一个 `acp_runtime`。当前支持 `opencode` 和 `codex`。

## 1. 创建 DaemonRuntime 和 Token

在前端 Daemon Monitor 页面创建 `DaemonRuntime`，再为该 runtime 生成 runtime token。token secret
只展示一次，需要保存下来用于启动 daemon。

也可以通过 `DaemonService` API 创建：

- `CreateDaemonRuntime`
- `CreateDaemonRuntimeToken`
- `ListDaemonRuntimes`
- `ListDaemons`
- `ListDaemonTasks`
- `CancelDaemonTask`

API 字段和请求格式见 [docs/api.md](./api.md) 的 `DaemonService` 章节。

## 2. 启动 Daemon

### 本地二进制

```bash
go run ./cmd/butter-daemon \
  --url https://butter.example.com/api \
  --token bt_daemon_runtime_xxx
```

`--url` 传 ConnectRPC base URL，例如 `https://butter.example.com/api`。daemon client
会自动追加 `/agents.v1.DaemonConnectorService/Connect`，不要把方法路径手动拼进去。
HTTPS/HTTP2 入口是推荐部署方式；没有 scheme 的值会被视为 cleartext HTTP 地址，需要入口支持 h2c。

如果前面有 Nginx 之类的反代，daemon connector 可以和普通 `/api` 一样转发到 Butter
主 HTTP 端口，但这个长连接路径必须关闭 buffering 并拉长超时：

```nginx
location /api/agents.v1.DaemonConnectorService/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;

    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto https;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

    proxy_buffering off;
    proxy_request_buffering off;
    proxy_read_timeout 3600s;
    proxy_send_timeout 3600s;
}
```

### 环境变量

`butter-daemon` 支持用环境变量提供连接参数：

```bash
export BUTTER_DAEMON_URL=https://butter.example.com/api
export BUTTER_DAEMON_TOKEN=bt_daemon_runtime_xxx
go run ./cmd/butter-daemon
```

优先级如下：

1. CLI flags：`--url`、`--token`
2. 配置文件字段：`server` / `url`、`credential` / `token`
3. 环境变量：`BUTTER_DAEMON_URL`、`BUTTER_DAEMON_TOKEN`

### systemd

Linux 服务器上可以用 systemd 托管本地二进制。先构建并安装 daemon：

```bash
make build
sudo install -m 0755 bin/butter-daemon /usr/local/bin/butter-daemon
```

建议用独立系统用户运行 daemon，并准备配置目录：

```bash
sudo useradd --system --home /var/lib/butter-daemon --shell /usr/sbin/nologin butter || true
sudo install -d -o butter -g butter /var/lib/butter-daemon
sudo install -d -m 0755 /etc/butter
sudo install -d -o butter -g butter /tmp/butter-daemon-workdirs
```

把连接信息和运行 ACP executor 需要的凭证写入 `/etc/butter/butter-daemon.env`：

```dotenv
BUTTER_DAEMON_URL=https://butter.example.com/api
BUTTER_DAEMON_TOKEN=bt_daemon_runtime_xxx
OPENAI_API_KEY=sk-xxx
GH_TOKEN=ghp_xxx
```

保护 token 文件权限：

```bash
sudo chown root:butter /etc/butter/butter-daemon.env
sudo chmod 0640 /etc/butter/butter-daemon.env
```

创建 `/etc/systemd/system/butter-daemon.service`：

```ini
[Unit]
Description=Butter daemon worker
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=butter
Group=butter
WorkingDirectory=/var/lib/butter-daemon
EnvironmentFile=/etc/butter/butter-daemon.env
ExecStart=/usr/local/bin/butter-daemon
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

启动并设置开机自启：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now butter-daemon
sudo systemctl status butter-daemon
```

查看日志：

```bash
journalctl -u butter-daemon -f
```

如果 Butter server 和 daemon 不在同一个文件系统中，systemd 部署同样需要保证
`/tmp/butter-daemon-workdirs` 是两边可访问的同一个绝对路径；否则 daemon 收到任务后无法
`chdir` 到服务端创建的 workdir。

### Docker Image

daemon 镜像由 `.github/workflows/daemon-publish.yml` 发布：

```bash
ghcr.io/orvice/butter-daemon:main
```

镜像 entrypoint 已经是 `/app/butter-daemon`，所以可以直接传 flags：

```bash
docker run -d \
  --name butter-daemon \
  --restart unless-stopped \
  --add-host=host.docker.internal:host-gateway \
  -v /tmp/butter-daemon-workdirs:/tmp/butter-daemon-workdirs \
  -e OPENAI_API_KEY="$OPENAI_API_KEY" \
  -e GH_TOKEN="$GH_TOKEN" \
  ghcr.io/orvice/butter-daemon:main \
  --url https://butter.example.com/api \
  --token bt_daemon_runtime_xxx
```

也可以只用 env 启动：

```bash
docker run -d \
  --name butter-daemon \
  --restart unless-stopped \
  --add-host=host.docker.internal:host-gateway \
  -v /tmp/butter-daemon-workdirs:/tmp/butter-daemon-workdirs \
  -e BUTTER_DAEMON_URL=https://butter.example.com/api \
  -e BUTTER_DAEMON_TOKEN=bt_daemon_runtime_xxx \
  -e OPENAI_API_KEY="$OPENAI_API_KEY" \
  -e GH_TOKEN="$GH_TOKEN" \
  ghcr.io/orvice/butter-daemon:main
```

Linux 上如果 daemon 容器要访问宿主机的 Butter 服务，通常需要
`--add-host=host.docker.internal:host-gateway`。macOS/Windows Docker Desktop 通常已经内置
`host.docker.internal`。

如果 Butter server 和 daemon 在同一个 Docker network 中，仍建议让 daemon 连接经过同一层
HTTPS 反代入口，保持 URL 形态与 dashboard 一致：

```yaml
services:
  butter:
    image: ghcr.io/orvice/butter:main
    volumes:
      - butter-daemon-workdirs:/tmp/butter-daemon-workdirs

  butter-daemon:
    image: ghcr.io/orvice/butter-daemon:main
    restart: unless-stopped
    environment:
      BUTTER_DAEMON_URL: https://butter.example.com/api
      BUTTER_DAEMON_TOKEN: ${BUTTER_DAEMON_TOKEN}
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
      GH_TOKEN: ${GH_TOKEN:-}
    volumes:
      - butter-daemon-workdirs:/tmp/butter-daemon-workdirs

volumes:
  butter-daemon-workdirs:
```

## 3. 配置 RemoteAgent

创建 RemoteAgent 时选择：

- `protocol`: `REMOTE_AGENT_PROTOCOL_DAEMON`
- `daemon_runtime_id`: 已创建并在线的 `DaemonRuntime`
- `acp_runtime`: `opencode` 或 `codex`

当用户调用该 RemoteAgent 时，Butter runner 会创建 daemon bridge，把 ADK invocation 转成
`DaemonTask`，并按 `workspace_id + daemon_runtime_id` 找到在线 daemon 连接。

## 4. Executor 和本地配置

daemon 默认内置两个 ACP executor：

| `acp_runtime` | 默认命令 | 说明 |
| --- | --- | --- |
| `opencode` | `opencode acp` | 使用 opencode 的 ACP stdio 入口 |
| `codex` | `codex-acp` | 使用 Codex ACP adapter |

镜像内已包含 `opencode`、`codex`、`codex-acp`、`gh`、`glab` 和 `git`。模型/API 凭证、
GitHub/GitLab token 等仍需要通过 env 或挂载配置提供。

可以通过 YAML 覆盖或扩展 executor：

```yaml
server: https://butter.example.com/api
credential: bt_daemon_runtime_xxx
name: local-worker

executors:
  acp:
    - runtime: opencode
      command: opencode
      args: ["acp"]
      permission_policy: deny
      fs:
        read: true
        write: true
      terminal: true
```

启动时指定配置文件：

```bash
butter-daemon --config ./daemon.yaml
```

## 5. Workdir 和 Volume 约束

当前实现中，server bridge 会按 workspace/session 在服务端创建工作目录：

```text
/tmp/butter-daemon-workdirs/<session-hash>
```

这个绝对路径会随 `DaemonTask.work_dir` 下发给 daemon。daemon 执行 ACP 进程时会 `chdir` 到该
路径，所以 daemon 必须能访问同一个绝对路径。

部署时需要满足以下条件之一：

- daemon 和 server 运行在同一台机器/同一个文件系统中。
- daemon 和 server 容器把同一个 volume 挂载到相同绝对路径
  `/tmp/butter-daemon-workdirs`。

如果路径不共享，daemon 会因为无法进入 workdir 而执行失败。

## 6. 状态和排障

前端 Daemon Monitor 页面可以查看：

- 已配置的 `DaemonRuntime`
- 在线 daemon 列表
- daemon version / OS / executors / remote address
- in-flight tasks
- task step / progress / elapsed
- task cancel 操作

常见问题：

- `url is required`：没有传 `--url`，也没有设置 `BUTTER_DAEMON_URL` 或配置文件 `server/url`。
- `token is required`：没有传 `--token`，也没有设置 `BUTTER_DAEMON_TOKEN` 或配置文件
  `credential/token`。
- `invalid token` / `token is not a daemon runtime token`：使用了普通 API token，或 token 已失效。
- `daemon runtime already connected`：同一个 workspace/runtime 已经有 daemon 在线。
- `unsupported acp_runtime`：RemoteAgent 选择的 `acp_runtime` 不在 daemon 注册的 executor 列表中。
- 任务无法进入目录：检查 server 和 daemon 是否共享 `/tmp/butter-daemon-workdirs`。
