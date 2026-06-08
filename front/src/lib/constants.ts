export const AGENT_TYPE_LABELS: Record<string, string> = {
  AGENT_TYPE_UNSPECIFIED: "Unspecified",
  AGENT_TYPE_LLM: "LLM",
  AGENT_TYPE_LOOP: "Loop",
  AGENT_TYPE_SEQUENTIAL: "Sequential",
  AGENT_TYPE_PARALLEL: "Parallel",
};

export const MCP_TRANSPORT_LABELS: Record<string, string> = {
  MCP_SERVER_TRANSPORT_UNSPECIFIED: "Unspecified",
  MCP_SERVER_TRANSPORT_STREAMABLE_HTTP: "HTTP",
  MCP_SERVER_TRANSPORT_SSE: "SSE",
};

export const MCP_AUTH_TYPE_LABELS: Record<string, string> = {
  MCP_SERVER_AUTH_TYPE_UNSPECIFIED: "Unspecified",
  MCP_SERVER_AUTH_TYPE_NONE: "None",
  MCP_SERVER_AUTH_TYPE_STATIC_HEADERS: "Static headers",
  MCP_SERVER_AUTH_TYPE_OAUTH2: "OAuth 2.0",
};

export const MCP_OAUTH_CONNECTION_LABELS: Record<string, string> = {
  MCPO_AUTH_CONNECTION_STATE_UNSPECIFIED: "Unknown",
  MCPO_AUTH_CONNECTION_STATE_DISCONNECTED: "Disconnected",
  MCPO_AUTH_CONNECTION_STATE_CONNECTED: "Connected",
  MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED: "Reconnect Required",
  MCPO_AUTH_CONNECTION_STATE_ERROR: "Error",
};

export const REMOTE_AGENT_PROTOCOL_LABELS: Record<string, string> = {
  REMOTE_AGENT_PROTOCOL_UNSPECIFIED: "Unspecified",
  REMOTE_AGENT_PROTOCOL_A2A: "A2A",
  REMOTE_AGENT_PROTOCOL_DAEMON: "Daemon Runtime",
};

export const AGENT_CHANNEL_PLATFORM_LABELS: Record<string, string> = {
  AGENT_CHANNEL_PLATFORM_UNSPECIFIED: "Unspecified",
  AGENT_CHANNEL_PLATFORM_TELEGRAM: "Telegram",
  AGENT_CHANNEL_PLATFORM_DISCORD: "Discord",
};

export const AGENT_TRIGGER_TYPE_LABELS: Record<string, string> = {
  AGENT_TRIGGER_TYPE_UNSPECIFIED: "Unspecified",
  AGENT_TRIGGER_TYPE_MESSAGE: "All messages",
  AGENT_TRIGGER_TYPE_COMMAND: "Commands",
  AGENT_TRIGGER_TYPE_MENTION: "Mentions",
  AGENT_TRIGGER_TYPE_PRIVATE_CHAT: "Private chat only",
};

export const INVOCATION_STATUS_LABELS: Record<string, string> = {
  INVOCATION_STATUS_UNSPECIFIED: "Unknown",
  INVOCATION_STATUS_RUNNING: "Running",
  INVOCATION_STATUS_SUCCEEDED: "Succeeded",
  INVOCATION_STATUS_FAILED: "Failed",
};

export const AGENT_RUNTIME_STATE_LABELS: Record<string, string> = {
  AGENT_RUNTIME_STATE_UNSPECIFIED: "Unknown",
  AGENT_RUNTIME_STATE_IDLE: "Idle",
  AGENT_RUNTIME_STATE_RUNNING: "Running",
  AGENT_RUNTIME_STATE_FAILED: "Failed",
};

export const COMPONENT_HEALTH_STATUS_LABELS: Record<string, string> = {
  STATUS_UNSPECIFIED: "Unknown",
  STATUS_HEALTHY: "Healthy",
  STATUS_DEGRADED: "Degraded",
  STATUS_DOWN: "Down",
};

export const RUNTIME_STATE_LABELS: Record<string, string> = {
  STATE_UNSPECIFIED: "Unknown",
  STATE_ONLINE: "Online",
  STATE_IDLE: "Idle",
  STATE_OFFLINE: "Offline",
  STATE_LIVE: "Live",
  STATE_PAUSED: "Paused",
  STATE_DISABLED: "Disabled",
  STATE_CONFIGURED: "Configured",
  STATE_CONNECTED: "Connected",
  STATE_DISCONNECTED: "Disconnected",
  STATE_ACTIVE: "Active",
  STATE_UNREACHABLE: "Unreachable",
  STATE_ERROR: "Error",
};

export const CRON_RANGE_LABELS: Record<string, string> = {
  RANGE_UNSPECIFIED: "Unspecified",
  RANGE_1D: "1 day",
  RANGE_7D: "7 days",
  RANGE_30D: "30 days",
};

export const CRON_DELIVERY_LABELS: Record<string, string> = {
  CRON_DELIVERY_TYPE_UNSPECIFIED: "Unspecified",
  CRON_DELIVERY_TYPE_LOG: "Log",
  CRON_DELIVERY_TYPE_WEBHOOK: "Webhook",
  CRON_DELIVERY_TYPE_CHANNEL: "Channel",
  CRON_DELIVERY_TYPE_NOTIFY_GROUP: "Notify Group",
};

export const CRON_STATUS_LABELS: Record<string, string> = {
  CRON_EXECUTION_STATUS_UNSPECIFIED: "Unknown",
  CRON_EXECUTION_STATUS_SUCCESS: "Success",
  CRON_EXECUTION_STATUS_ERROR: "Error",
};

export const NOTIFY_TARGET_TYPE_LABELS: Record<string, string> = {
  NOTIFY_TARGET_TYPE_UNSPECIFIED: "Unspecified",
  NOTIFY_TARGET_TYPE_TELEGRAM: "Telegram",
  NOTIFY_TARGET_TYPE_LARK_WEBHOOK: "Lark Webhook",
  NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK: "Discord Webhook",
};

export const STREAMING_MODE_LABELS: Record<string, string> = {
  STREAMING_MODE_UNSPECIFIED: "Unspecified",
  STREAMING_MODE_NONE: "None",
  STREAMING_MODE_SSE: "SSE",
};

export const LLM_INCLUDE_CONTENTS_LABELS: Record<string, string> = {
  LLM_INCLUDE_CONTENTS_UNSPECIFIED: "Unspecified",
  LLM_INCLUDE_CONTENTS_DEFAULT: "Default",
  LLM_INCLUDE_CONTENTS_NONE: "None",
};

export const CONTEXT_GUARD_STRATEGY_LABELS: Record<string, string> = {
  CONTEXT_GUARD_STRATEGY_UNSPECIFIED: "Unspecified",
  CONTEXT_GUARD_STRATEGY_THRESHOLD: "Threshold",
  CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW: "Sliding Window",
};

export const AGENT_FILE_MOUNT_PERMISSION_LABELS: Record<string, string> = {
  AGENT_FILE_MOUNT_PERMISSION_UNSPECIFIED: "Unspecified",
  AGENT_FILE_MOUNT_PERMISSION_READ: "Read",
  AGENT_FILE_MOUNT_PERMISSION_READ_WRITE: "Read / write",
  AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE: "Read / write / delete",
};

export const DAEMON_TASK_STATUS_LABELS: Record<string, string> = {
  DAEMON_TASK_STATUS_UNSPECIFIED: "Unknown",
  DAEMON_TASK_STATUS_ACCEPTED: "Accepted",
  DAEMON_TASK_STATUS_RUNNING: "Running",
  DAEMON_TASK_STATUS_COMPLETED: "Completed",
  DAEMON_TASK_STATUS_FAILED: "Failed",
  DAEMON_TASK_STATUS_CANCELLED: "Cancelled",
};

const ENUM_LABELS: Record<string, string> = {
  ...AGENT_TYPE_LABELS,
  ...MCP_TRANSPORT_LABELS,
  ...MCP_AUTH_TYPE_LABELS,
  ...MCP_OAUTH_CONNECTION_LABELS,
  ...REMOTE_AGENT_PROTOCOL_LABELS,
  ...AGENT_CHANNEL_PLATFORM_LABELS,
  ...AGENT_TRIGGER_TYPE_LABELS,
  ...INVOCATION_STATUS_LABELS,
  ...AGENT_RUNTIME_STATE_LABELS,
  ...COMPONENT_HEALTH_STATUS_LABELS,
  ...RUNTIME_STATE_LABELS,
  ...CRON_RANGE_LABELS,
  ...CRON_DELIVERY_LABELS,
  ...CRON_STATUS_LABELS,
  ...NOTIFY_TARGET_TYPE_LABELS,
  ...STREAMING_MODE_LABELS,
  ...LLM_INCLUDE_CONTENTS_LABELS,
  ...CONTEXT_GUARD_STRATEGY_LABELS,
  ...AGENT_FILE_MOUNT_PERMISSION_LABELS,
  ...DAEMON_TASK_STATUS_LABELS,
};

const ACRONYMS = new Set(["A2A", "ADK", "API", "HTML", "HTTP", "ID", "LLM", "MCP", "OAuth", "SSE", "URL"]);

export function enumLabel(value: unknown, fallback = "Unspecified"): string {
  if (value === undefined || value === null || value === "") return fallback;
  if (Array.isArray(value)) return value.map((item) => enumLabel(item, fallback)).join(", ");
  if (typeof value !== "string") return String(value);
  if (ENUM_LABELS[value]) return ENUM_LABELS[value];
  if (!/^[A-Z][A-Z0-9_]*$/.test(value)) return value;
  return value
    .split("_")
    .filter(Boolean)
    .map((part) => {
      if (ACRONYMS.has(part)) return part;
      return part.charAt(0) + part.slice(1).toLowerCase();
    })
    .join(" ");
}

export const TOKEN_KEY = "butter_token";
export const WORKSPACE_KEY = "butter_workspace_id";
export const COLOR_THEME_KEY = "butter_color_theme";
export const LAYOUT_DENSITY_KEY = "butter_layout_density";
