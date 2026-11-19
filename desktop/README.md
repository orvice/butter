# Butter Desktop

Butter 的原生 macOS 客户端，用 [MyGo](https://github.com/egoist/mygo) 的原生 UI 绘制，不使用 WebView。范围与验收见 [`docs/desktop-macos-phase1.md`](../docs/desktop-macos-phase1.md)（随 PR #532 提交），调研与风险见 [`docs/research/mygo-butter-desktop.md`](../docs/research/mygo-butter-desktop.md)。

当前状态：**M0（骨架、构建、密码登录）**。

## 现在能做什么

- 输入服务器地址、用户名、密码，调 `AuthService.Login` 登录；token 只写入系统钥匙串（Keychain service `go.orx.me/apps/butter/desktop`），密码只用于这一次请求。
- 登录失败按类别提示：密码错误、账号被禁用、请求缺字段、服务端认证存储不可用、网络不可达；**不自动重试**（服务端没有登录限流，重试只会放大问题）。
- 多工作区时先选工作区；只有一个时自动选中；全局管理员没有显式成员关系时在登录后补一次 `WorkspaceService.ListWorkspaces`。
- 惰性恢复：启动时若保存过服务器地址，从钥匙串取 token 调 `AuthService.Me` 校验；过期则清掉本地 token 并要求重新登录。
- 退出登录会尽力撤销服务端会话，然后删除本地 token；服务端不可达时也保证本地退出。
- 更换服务器前先删除旧服务器的 token；切换期间返回的登录结果会被丢弃（generation 校验），不会把上一个服务器的会话写进新上下文。
- 服务器地址、用户名、工作区保存在 `~/Library/Application Support/butter/desktop.json`（权限 0600，不含 token 与密码）。

## 还不做

- 聊天、AG-UI、A2UI（M1、M2）。
- 签名与公证（M3）；本地构建出的 `.app` 未经 Developer ID 签名，只能在构建机或已允许的环境打开。
- 自动更新、OAuth、Windows/Linux。
- 没有工作区的普通账号目前会停在"没有可访问的工作区"提示上；管理员用 `WorkspaceService.ListWorkspaces` 补全列表。

## 运行

需要 macOS 与 Go 1.27.3（本 module 要求；仓库根 module 仍是 1.27.1，两者互不影响）。

```bash
cd desktop
go run ./cmd/butter-desktop
```

打包成 `.app` / DMG 用 MyGo 的构建流程（`mygo build`），还没接进 CI。

## 验证

```bash
cd desktop
go test ./...                                  # 传输、登录分类、凭证作用域、上下文切换、偏好读写
go vet ./...
GOOS=darwin CGO_ENABLED=0 go vet ./...         # 类型检查 macOS 专属的 UI 代码（Linux 上也能跑）
```

`internal/app` 与 `cmd/butter-desktop` 带 `//go:build darwin`：第一阶段的窗口只在 macOS 上构建，Linux 的开发机因此也能编译与测试其余部分。macOS 上的 `go build ./...` 会同时编译窗口代码，这是 GitHub Actions `desktop.yml` 的 `build-macos` 任务存在的意义。

没有在本机通过网络请求真实 Butter server 做过端到端登录验证；协议侧的依据是 `pkg/proto/agents/v1` 与 `docs/api.md`。

## 代码结构

```text
cmd/butter-desktop   main：读偏好、装配 Keychain 与 account.Manager、开窗口（darwin）
internal/app         窗口与状态机：登录、工作区选择、已登录视图（darwin）
internal/account     服务器/账号/工作区上下文，token 存取与 generation 校验
internal/authapi     AuthService/WorkspaceService 的 Connect 客户端与错误分类
internal/endpoint    服务器地址归一化与按服务器区分的凭证 key
internal/credentials 凭证存储接口、内存实现、Keychain 实现（darwin）
internal/prefs       非敏感偏好（服务器、用户名、工作区）
internal/version     版本与 VCS 修订
```

约定：UI 不直接发 HTTP；`account.Manager` 是唯一持有 token 与当前上下文的类型；后台 goroutine 只通过 `Window.Update` 改 UI 状态。
