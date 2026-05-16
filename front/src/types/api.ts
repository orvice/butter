// --- Enums ---

export type AgentType =
  | "AGENT_TYPE_UNSPECIFIED"
  | "AGENT_TYPE_LLM"
  | "AGENT_TYPE_LOOP"
  | "AGENT_TYPE_SEQUENTIAL"
  | "AGENT_TYPE_PARALLEL";

export type MCPServerTransport =
  | "MCP_SERVER_TRANSPORT_UNSPECIFIED"
  | "MCP_SERVER_TRANSPORT_STDIO"
  | "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP"
  | "MCP_SERVER_TRANSPORT_SSE";

export type RemoteAgentProtocol =
  | "REMOTE_AGENT_PROTOCOL_UNSPECIFIED"
  | "REMOTE_AGENT_PROTOCOL_A2A"
  | "REMOTE_AGENT_PROTOCOL_DAEMON";

export type AgentChannelPlatform =
  | "AGENT_CHANNEL_PLATFORM_UNSPECIFIED"
  | "AGENT_CHANNEL_PLATFORM_TELEGRAM"
  | "AGENT_CHANNEL_PLATFORM_DISCORD";

export type AgentTriggerType =
  | "AGENT_TRIGGER_TYPE_UNSPECIFIED"
  | "AGENT_TRIGGER_TYPE_ALL"
  | "AGENT_TRIGGER_TYPE_COMMAND"
  | "AGENT_TRIGGER_TYPE_MESSAGE";

export type InvocationStatus =
  | "INVOCATION_STATUS_UNSPECIFIED"
  | "INVOCATION_STATUS_RUNNING"
  | "INVOCATION_STATUS_SUCCEEDED"
  | "INVOCATION_STATUS_FAILED";

export type AgentRuntimeState =
  | "AGENT_RUNTIME_STATE_UNSPECIFIED"
  | "AGENT_RUNTIME_STATE_IDLE"
  | "AGENT_RUNTIME_STATE_RUNNING"
  | "AGENT_RUNTIME_STATE_FAILED";

export type ComponentHealthStatus =
  | "STATUS_UNSPECIFIED"
  | "STATUS_HEALTHY"
  | "STATUS_DEGRADED"
  | "STATUS_DOWN";

export type DaemonState =
  | "STATE_UNSPECIFIED"
  | "STATE_ONLINE"
  | "STATE_IDLE"
  | "STATE_OFFLINE";

export type ChannelState =
  | "STATE_UNSPECIFIED"
  | "STATE_LIVE"
  | "STATE_PAUSED"
  | "STATE_DISABLED"
  | "STATE_ERROR";

export type MCPServerState =
  | "STATE_UNSPECIFIED"
  | "STATE_CONFIGURED"
  | "STATE_CONNECTED"
  | "STATE_DISCONNECTED"
  | "STATE_ERROR";

export type RemoteAgentState =
  | "STATE_UNSPECIFIED"
  | "STATE_CONFIGURED"
  | "STATE_ACTIVE"
  | "STATE_IDLE"
  | "STATE_UNREACHABLE"
  | "STATE_ERROR";

export type CronTimeseriesRange =
  | "RANGE_UNSPECIFIED"
  | "RANGE_1D"
  | "RANGE_7D"
  | "RANGE_30D";

export type CronDeliveryType =
  | "CRON_DELIVERY_TYPE_UNSPECIFIED"
  | "CRON_DELIVERY_TYPE_LOG"
  | "CRON_DELIVERY_TYPE_WEBHOOK"
  | "CRON_DELIVERY_TYPE_CHANNEL";

export type CronExecutionStatus =
  | "CRON_EXECUTION_STATUS_UNSPECIFIED"
  | "CRON_EXECUTION_STATUS_SUCCESS"
  | "CRON_EXECUTION_STATUS_ERROR";

export type StreamingMode =
  | "STREAMING_MODE_UNSPECIFIED"
  | "STREAMING_MODE_NONE"
  | "STREAMING_MODE_SSE";

export type LLMIncludeContents =
  | "LLM_INCLUDE_CONTENTS_UNSPECIFIED"
  | "LLM_INCLUDE_CONTENTS_DEFAULT"
  | "LLM_INCLUDE_CONTENTS_NONE";

export type ContextGuardStrategy =
  | "CONTEXT_GUARD_STRATEGY_UNSPECIFIED"
  | "CONTEXT_GUARD_STRATEGY_THRESHOLD"
  | "CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW";

// --- Models ---

export interface ContextGuardConfig {
  strategy?: ContextGuardStrategy;
  max_turns?: number;
  max_tokens?: number;
}

export interface AgentRuntime {
  streaming_mode?: StreamingMode;
  save_input_blobs_as_artifacts?: boolean;
}

export interface AgentConfig {
  runtime?: AgentRuntime;
  mcp_servers?: MCPServer[];
  context_guard?: ContextGuardConfig;
  mcp_server_ids?: string[];
  remote_agent_ids?: string[];
  model?: string;
  instruction?: string;
  global_instruction?: string;
  disallow_transfer_to_parent?: boolean;
  disallow_transfer_to_peers?: boolean;
  include_contents?: LLMIncludeContents;
  output_key?: string;
  input_schema_json?: string;
  output_schema_json?: string;
  max_iterations?: number;
}

export interface Agent {
  name: string;
  description?: string;
  sub_agents?: Agent[];
  labels?: Record<string, string>;
  metadata?: Record<string, string>;
  config?: AgentConfig;
  type?: AgentType;
  enable_a2a?: boolean;
}

export interface MCPServer {
  id?: string;
  name: string;
  transport?: MCPServerTransport;
  command?: string;
  args?: string[];
  env?: Record<string, string>;
  url?: string;
  headers?: Record<string, string>;
  tool_filter?: string[];
  metadata?: Record<string, string>;
}

export interface RemoteAgent {
  id: string;
  name: string;
  url: string;
  protocol?: RemoteAgentProtocol;
  daemon_capability?: string;
}

export interface SessionInfo {
  session_id: string;
  app_name: string;
  user_id: string;
  state?: Record<string, unknown>;
  last_update_time?: string;
  turn_count?: number;
}

export interface SessionEvent {
  event_id: string;
  invocation_id?: string;
  author?: string;
  branch?: string;
  content_json?: string;
  timestamp?: string;
  trace_id?: string;
  trace_url?: string;
}

export interface SessionDetail {
  session: SessionInfo;
  events: SessionEvent[];
  duration?: string; // protobuf Duration as "3.5s"
}

// --- Channels ---

export interface AgentTrigger {
  type?: AgentTriggerType;
  commands?: string[];
  prefixes?: string[];
  require_mention?: boolean;
}

export interface TelegramChannelConfig {
  bot_token?: string;
  allow_chat_ids?: string[];
}

export interface DiscordChannelConfig {
  bot_token?: string;
  allow_channel_ids?: string[];
}

export interface AgentChannel {
  name: string;
  agent_name: string;
  platform?: AgentChannelPlatform;
  enabled?: boolean;
  triggers?: AgentTrigger[];
  telegram?: TelegramChannelConfig;
  discord?: DiscordChannelConfig;
  model?: string;
  metadata?: Record<string, string>;
}

export interface ChannelStatus {
  name: string;
  platform?: AgentChannelPlatform;
  state?: ChannelState;
  last_poll_at?: string;
  detail?: string;
}

// --- Invocations ---

export interface Invocation {
  id: string;
  agent_name: string;
  app_name?: string;
  user_id?: string;
  session_id?: string;
  status?: InvocationStatus;
  input?: string;
  output?: string;
  error?: string;
  started_at?: string;
  finished_at?: string;
  latency_ms?: number;
  model_override?: string;
  source?: string;
}

export interface AgentRuntimeStatus {
  name: string;
  state?: AgentRuntimeState;
  last_run_at?: string;
  last_invocation_id?: string;
  in_flight?: number;
}

// --- Dashboard ---

export interface OverviewCounts {
  active_agents?: number;
  mcp_servers?: number;
  connected_daemons?: number;
  remote_agents?: number;
  channels?: number;
  cron_jobs?: number;
  active_sessions?: number;
}

export interface ComponentHealth {
  status?: ComponentHealthStatus;
  detail?: string;
  checked_at?: string;
  latency_ms?: number;
}

export interface HealthSummary {
  mongodb?: ComponentHealth;
  redis?: ComponentHealth;
  runner?: ComponentHealth;
}

export interface DaemonHandshake {
  daemon_id?: string;
  name?: string;
  capabilities?: string[];
  connected_at?: string;
  os?: string;
}

export interface GetOverviewResponse {
  counts?: OverviewCounts;
  health?: HealthSummary;
  latest_daemon_handshake?: DaemonHandshake;
}

export interface ActivityEvent {
  id: string;
  kind?: string;
  actor?: string;
  message?: string;
  timestamp?: string;
}

export interface CronExecutionBucket {
  start?: string;
  success?: number;
  error?: number;
}

// --- Daemons ---

export interface DaemonStatus {
  daemon_id: string;
  name?: string;
  capabilities?: string[];
  labels?: Record<string, string>;
  state?: DaemonState;
  connected_at?: string;
  uptime?: string;
  active_tasks?: number;
  version?: string;
  os?: string;
  executors?: string[];
  remote_addr?: string;
}

export interface DaemonTaskInFlight {
  task_id: string;
  daemon_id?: string;
  daemon_name?: string;
  capability?: string;
  started_at?: string;
  elapsed?: string;
  current_step?: string;
  progress?: number;
  agent_name?: string;
}

export interface LatencyPoint {
  timestamp?: string;
  latency_ms?: number;
}

export interface BridgeDiagnostics {
  cpu_percent?: number;
  memory_used_bytes?: number;
  memory_limit_bytes?: number;
  goroutines?: number;
  checked_at?: string;
  latency?: LatencyPoint[];
}

// --- MCP / Remote Status ---

export interface MCPServerStatus {
  id: string;
  name?: string;
  state?: MCPServerState;
  tool_count?: number;
  detail?: string;
  checked_at?: string;
}

export interface MCPTool {
  name: string;
  description?: string;
  server_id?: string;
  server_name?: string;
  allowed?: boolean;
}

export interface RemoteAgentStatus {
  id: string;
  protocol?: RemoteAgentProtocol;
  state?: RemoteAgentState;
  detail?: string;
  serving_daemon_id?: string;
  checked_at?: string;
  latency_ms?: number;
}

// --- API Tokens ---

export interface APIToken {
  id: string;
  name: string;
  prefix?: string;
  created_at?: string;
  last_used_at?: string;
  revoked?: boolean;
}

export interface CronDelivery {
  type?: CronDeliveryType;
  webhook_url?: string;
  channel_name?: string;
  chat_id?: string;
}

export interface CronJob {
  name: string;
  schedule: string;
  agent_name: string;
  input?: string;
  timezone?: string;
  enabled?: boolean;
  delivery?: CronDelivery;
  metadata?: Record<string, string>;
}

export interface CronExecution {
  id: string;
  job_name: string;
  agent_name: string;
  status: CronExecutionStatus;
  input?: string;
  output?: string;
  started_at?: string;
  finished_at?: string;
}

// --- Twirp Error ---

export interface TwirpError {
  code: string;
  msg: string;
}
