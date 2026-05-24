# Frontend Required API Follow-ups

更新时间：2026-05-24

这份清单只记录前端仍需要后端新增语义或扩展字段的能力。已经存在的后端接口已在前端接入，不再作为缺口跟踪。

## 已接入的既有后端能力

| 前端能力 | 已使用接口 | 说明 |
|---|---|---|
| Global MCP preset 安装到当前 workspace | `POST /api/global-mcp-servers/{id}/install` | Admin Global MCP 页面现在可以直接安装 preset。 |
| Forum thread/post 管理 | `ForumService.UpdateThread`、`DeleteThread`、`DeletePost` | Thread 页面新增编辑、删除 thread、删除 post。 |
| Operations 内嵌 Session 过滤 | `SessionService.ListSessions` | 复用 `app_name`、`user_id`、`page_size` 参数。 |
| Dashboard 环境选择 | `DashboardService.GetOverview.environment` | 前端会传 environment；当前后端仍只是接收字段，不做过滤。 |

## 仍需要后端新增或扩展

### 1. 全局搜索

当前顶栏有搜索入口，但没有统一搜索接口。建议新增只读聚合 RPC：

```proto
service SearchService {
  rpc Search(SearchRequest) returns (SearchResponse);
}

message SearchRequest {
  string query = 1;
  repeated string scopes = 2; // agents, sessions, forum, mcp_servers, cron_jobs, workspaces
  int32 limit = 3;
}

message SearchResult {
  string id = 1;
  string scope = 2;
  string title = 3;
  string subtitle = 4;
  string url = 5;
  map<string, string> metadata = 6;
}

message SearchResponse {
  repeated SearchResult results = 1;
}
```

### 2. 通知中心

顶栏通知按钮还没有数据源。建议先做最小可用的系统通知列表：

```proto
service NotificationService {
  rpc ListNotifications(ListNotificationsRequest) returns (ListNotificationsResponse);
  rpc MarkNotificationRead(MarkNotificationReadRequest) returns (MarkNotificationReadResponse);
  rpc MarkAllNotificationsRead(MarkAllNotificationsReadRequest) returns (MarkAllNotificationsReadResponse);
}
```

通知来源可以先聚合 daemon 断连、cron 执行失败、MCP OAuth 失效、channel runtime error。

### 3. Storage 状态

顶栏 Storage 按钮需要后端暴露静态存储/S3 配置与健康状态。建议新增到 `DashboardService.GetOverview` 或独立 RPC：

```proto
message StorageHealth {
  ComponentHealth.Status status = 1;
  string provider = 2; // local, s3
  string bucket = 3;
  string public_base_url = 4;
  string detail = 5;
  google.protobuf.Timestamp checked_at = 6;
}
```

### 4. Dashboard 趋势数据

Overview 四个 stat card 目前只能展示当前计数，缺少真实趋势。建议扩展 `OverviewCounts`：

```proto
message CountMetric {
  int32 value = 1;
  int32 previous_value = 2;
  double delta_percent = 3;
  string window = 4; // 24h, 7d
}
```

可逐步替换 `active_agents`、`mcp_servers`、`connected_daemons`、`active_sessions` 的裸 int 字段，或新增 `metrics` map 保持兼容。

### 5. Environment 维度过滤

`GetOverviewRequest.environment` 已存在，但后端注释说明当前不做过滤。若前端要让 Production/Staging/Development 真正影响数据，需要给配置实体、invocation、cron execution、daemon handshake 增加 environment 标签，并在 dashboard 聚合中按标签过滤。

## 暂不需要新增后端的项

- Copy 按钮：纯前端 Clipboard API。
- Forum 删除/编辑：已有 Twirp RPC。
- Global MCP preset 安装：已有 HTTP route。
- Session Explorer 过滤：已有 `ListSessions` 参数。
