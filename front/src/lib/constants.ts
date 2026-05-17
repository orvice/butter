export const AGENT_TYPE_LABELS: Record<string, string> = {
  AGENT_TYPE_UNSPECIFIED: "Unspecified",
  AGENT_TYPE_LLM: "LLM",
  AGENT_TYPE_LOOP: "Loop",
  AGENT_TYPE_SEQUENTIAL: "Sequential",
  AGENT_TYPE_PARALLEL: "Parallel",
};

export const MCP_TRANSPORT_LABELS: Record<string, string> = {
  MCP_SERVER_TRANSPORT_UNSPECIFIED: "Unspecified",
  MCP_SERVER_TRANSPORT_STDIO: "Stdio",
  MCP_SERVER_TRANSPORT_STREAMABLE_HTTP: "HTTP",
  MCP_SERVER_TRANSPORT_SSE: "SSE",
};

export const CRON_DELIVERY_LABELS: Record<string, string> = {
  CRON_DELIVERY_TYPE_UNSPECIFIED: "Unspecified",
  CRON_DELIVERY_TYPE_LOG: "Log",
  CRON_DELIVERY_TYPE_WEBHOOK: "Webhook",
  CRON_DELIVERY_TYPE_CHANNEL: "Channel",
};

export const CRON_STATUS_LABELS: Record<string, string> = {
  CRON_EXECUTION_STATUS_UNSPECIFIED: "Unknown",
  CRON_EXECUTION_STATUS_SUCCESS: "Success",
  CRON_EXECUTION_STATUS_ERROR: "Error",
};

export const TOKEN_KEY = "butter_token";
export const WORKSPACE_KEY = "butter_workspace_id";
