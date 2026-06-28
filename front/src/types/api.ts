// --- Enums ---

export type AgentType =
  | "AGENT_TYPE_UNSPECIFIED"
  | "AGENT_TYPE_LLM"
  | "AGENT_TYPE_LOOP"
  | "AGENT_TYPE_SEQUENTIAL"
  | "AGENT_TYPE_PARALLEL";

export type MCPServerTransport =
  | "MCP_SERVER_TRANSPORT_UNSPECIFIED"
  | "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP"
  | "MCP_SERVER_TRANSPORT_SSE";

export type MCPServerAuthType =
  | "MCP_SERVER_AUTH_TYPE_UNSPECIFIED"
  | "MCP_SERVER_AUTH_TYPE_NONE"
  | "MCP_SERVER_AUTH_TYPE_STATIC_HEADERS"
  | "MCP_SERVER_AUTH_TYPE_OAUTH2";

export type RemoteAgentProtocol =
  | "REMOTE_AGENT_PROTOCOL_UNSPECIFIED"
  | "REMOTE_AGENT_PROTOCOL_A2A"
  | "REMOTE_AGENT_PROTOCOL_DAEMON"
  | "REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP";

export type AgentChannelPlatform =
  | "AGENT_CHANNEL_PLATFORM_UNSPECIFIED"
  | "AGENT_CHANNEL_PLATFORM_TELEGRAM"
  | "AGENT_CHANNEL_PLATFORM_DISCORD";

export type AgentTriggerType =
  | "AGENT_TRIGGER_TYPE_UNSPECIFIED"
  | "AGENT_TRIGGER_TYPE_MESSAGE"
  | "AGENT_TRIGGER_TYPE_COMMAND"
  | "AGENT_TRIGGER_TYPE_MENTION"
  | "AGENT_TRIGGER_TYPE_PRIVATE_CHAT";

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

export type MCPOAuthConnectionState =
  | "MCPO_AUTH_CONNECTION_STATE_UNSPECIFIED"
  | "MCPO_AUTH_CONNECTION_STATE_DISCONNECTED"
  | "MCPO_AUTH_CONNECTION_STATE_CONNECTED"
  | "MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED"
  | "MCPO_AUTH_CONNECTION_STATE_ERROR";

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
  | "CRON_DELIVERY_TYPE_CHANNEL"
  | "CRON_DELIVERY_TYPE_NOTIFY_GROUP";

export type NotifyTargetType =
  | "NOTIFY_TARGET_TYPE_UNSPECIFIED"
  | "NOTIFY_TARGET_TYPE_TELEGRAM"
  | "NOTIFY_TARGET_TYPE_LARK_WEBHOOK"
  | "NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK";

export type CronExecutionStatus =
  | "CRON_EXECUTION_STATUS_UNSPECIFIED"
  | "CRON_EXECUTION_STATUS_SUCCESS"
  | "CRON_EXECUTION_STATUS_ERROR"
  | "CRON_EXECUTION_STATUS_SKIPPED"
  | "CRON_EXECUTION_STATUS_CANCELLED";

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

export type AgentFileMountPermission =
  | "AGENT_FILE_MOUNT_PERMISSION_UNSPECIFIED"
  | "AGENT_FILE_MOUNT_PERMISSION_READ"
  | "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE"
  | "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE";

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
  file_mounts?: AgentFileMount[];
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

export interface AgentFileSpace {
  id?: string;
  name: string;
  description?: string;
  metadata?: Record<string, string>;
  created_at?: string;
  updated_at?: string;
  workspace_id?: string;
}

export interface AgentFile {
  id?: string;
  space_id?: string;
  path: string;
  content_type?: string;
  size_bytes?: number | string;
  version?: number | string;
  metadata?: Record<string, string>;
  created_at?: string;
  updated_at?: string;
  workspace_id?: string;
}

export interface AgentFileSearchResult {
  file?: AgentFile;
  snippets?: string[];
}

export interface AgentFileMount {
  space_id: string;
  mount_path?: string;
  permission?: AgentFileMountPermission;
}

export interface MCPServer {
  id?: string;
  name: string;
  transport?: MCPServerTransport;
  url?: string;
  headers?: Record<string, string>;
  tool_filter?: string[];
  metadata?: Record<string, string>;
  timeout_seconds?: number;
  auth?: MCPServerAuth;
}

export interface MCPServerAuth {
  type?: MCPServerAuthType;
  oauth2?: MCPServerOAuth2Config;
}

export interface MCPServerOAuth2Config {
  client_id?: string;
  client_secret?: string;
  scopes?: string[];
  authorization_url?: string;
  token_url?: string;
  resource_metadata_url?: string;
  authorization_server_url?: string;
  resource?: string;
}

export interface RemoteAgent {
  id: string;
  name: string;
  url: string;
  protocol?: RemoteAgentProtocol;
  daemon_runtime_id?: string;
  acp_runtime?: string;
  opencode_agent?: string;
  opencode_model?: string;
  username?: string;
  password?: string;
}

export interface ModelConfig {
  name: string;
  alias?: string;
}

export interface ModelProvider {
  name: string;
  type: string;
  api_key?: string;
  base_url?: string;
  models?: ModelConfig[];
}

export interface TelegramNotifyTarget {
  bot_token?: string;
  chat_id?: string;
  parse_mode?: string;
  message_thread_id?: number;
}

export interface LarkNotifyTarget {
  webhook_url?: string;
  secret?: string;
}

export interface DiscordNotifyTarget {
  webhook_url?: string;
  username?: string;
  avatar_url?: string;
  thread_id?: string;
}

export interface NotifyTarget {
  name?: string;
  enabled?: boolean;
  type?: NotifyTargetType;
  telegram?: TelegramNotifyTarget;
  lark?: LarkNotifyTarget;
  discord?: DiscordNotifyTarget;
  metadata?: Record<string, string>;
}

export interface NotifyGroup {
  name: string;
  enabled?: boolean;
  targets?: NotifyTarget[];
  metadata?: Record<string, string>;
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

// --- Forum ---

export interface ForumThread {
  id: string;
  title: string;
  body?: string;
  created_by?: string;
  status?: string;
  agent_names?: string[];
  labels?: string[];
  metadata?: Record<string, string>;
  created_at?: string;
  updated_at?: string;
  workspace_id?: string;
}

export interface ForumPost {
  id: string;
  thread_id: string;
  body: string;
  author_user_id?: string;
  author_agent_name?: string;
  author_kind?: "user" | "agent" | "system" | string;
  invocation_id?: string;
  parent_post_id?: string;
  created_at?: string;
  updated_at?: string;
  workspace_id?: string;
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
  allowed_chat_ids?: string[];
  allow_chat_ids?: string[];
}

export interface DiscordChannelConfig {
  bot_token?: string;
  allowed_channel_ids?: string[];
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
  daemon_runtime_id?: string;
  name?: string;
  acp_runtimes?: string[];
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
  daemon_runtime_id: string;
  name?: string;
  acp_runtimes?: string[];
  labels?: Record<string, string>;
  state?: DaemonState;
  connected_at?: string;
  uptime?: string;
  active_tasks?: number;
  version?: string;
  os?: string;
  executors?: string[];
  remote_addr?: string;
  workspace_id?: string;
}

export interface DaemonRuntime {
  id: string;
  name: string;
  description?: string;
  labels?: Record<string, string>;
  created_at?: string;
  created_by?: string;
  workspace_id?: string;
}

export interface CreateDaemonRuntimeTokenInput {
  daemon_runtime_id: string;
  name?: string;
  ttl_hours?: number;
}

export interface CreateDaemonRuntimeTokenResult {
  token?: APIToken;
  secret: string;
}

export interface DaemonTaskInFlight {
  task_id: string;
  daemon_runtime_id?: string;
  daemon_name?: string;
  acp_runtime?: string;
  started_at?: string;
  elapsed?: string;
  current_step?: string;
  progress?: number;
  agent_name?: string;
  workspace_id?: string;
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

export interface MCPOAuthConnectionStatus {
  server_id?: string;
  state?: MCPOAuthConnectionState;
  detail?: string;
  scopes?: string[];
  connected_at?: string;
  expires_at?: string;
  checked_at?: string;
}

export interface RemoteAgentStatus {
  id: string;
  protocol?: RemoteAgentProtocol;
  state?: RemoteAgentState;
  detail?: string;
  serving_daemon_runtime_id?: string;
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
  kind?: string;
  scopes?: string[];
  expires_at?: string;
  daemon_runtime_id?: string;
  workspace_id?: string;
}

export interface CronDelivery {
  type?: CronDeliveryType;
  webhook_url?: string;
  channel_name?: string;
  chat_id?: string;
  notify_group_name?: string;
}

export interface CronJob {
  name: string;
  schedule: string;
  agent_name: string;
  input?: string;
  timezone?: string;
  enabled?: boolean;
  delivery?: CronDelivery;
  timeout_seconds?: number;
  retry?: { max_attempts?: number; backoff_seconds?: number };
  concurrency_policy?: CronConcurrencyPolicy;
  notify_on?: CronNotifyOn;
  max_output_bytes?: number;
  metadata?: Record<string, string>;
}

export interface CronExecution {
  id: string;
  job_name: string;
  agent_name: string;
  status: CronExecutionStatus;
  input?: string;
  output?: string;
  error?: string;
  started_at?: string;
  finished_at?: string;
  duration_ms?: number;
  attempt_count?: number;
  trigger_type?: CronExecutionTriggerType;
  skipped_reason?: string;
  truncated?: boolean;
}

export type CronConcurrencyPolicy =
  | "CRON_CONCURRENCY_POLICY_UNSPECIFIED"
  | "CRON_CONCURRENCY_POLICY_SKIP"
  | "CRON_CONCURRENCY_POLICY_QUEUE"
  | "CRON_CONCURRENCY_POLICY_REPLACE"
  | "CRON_CONCURRENCY_POLICY_ALLOW";

export type CronNotifyOn =
  | "CRON_NOTIFY_ON_UNSPECIFIED"
  | "CRON_NOTIFY_ON_ALWAYS"
  | "CRON_NOTIFY_ON_FAILURE"
  | "CRON_NOTIFY_ON_SUCCESS";

export type CronExecutionTriggerType =
  | "CRON_EXECUTION_TRIGGER_TYPE_UNSPECIFIED"
  | "CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE"
  | "CRON_EXECUTION_TRIGGER_TYPE_MANUAL";

export type AutomationTriggerType =
  | "AUTOMATION_TRIGGER_TYPE_UNSPECIFIED"
  | "AUTOMATION_TRIGGER_TYPE_MANUAL"
  | "AUTOMATION_TRIGGER_TYPE_SCHEDULE"
  | "AUTOMATION_TRIGGER_TYPE_WEBHOOK"
  | "AUTOMATION_TRIGGER_TYPE_FORUM_EVENT"
  | "AUTOMATION_TRIGGER_TYPE_CHANNEL_EVENT"
  | "AUTOMATION_TRIGGER_TYPE_DAEMON_EVENT";

export type AutomationConditionOperator =
  | "AUTOMATION_CONDITION_OPERATOR_UNSPECIFIED"
  | "AUTOMATION_CONDITION_OPERATOR_EQUALS"
  | "AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS"
  | "AUTOMATION_CONDITION_OPERATOR_CONTAINS"
  | "AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH"
  | "AUTOMATION_CONDITION_OPERATOR_EXISTS"
  | "AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS";

export type AutomationStepType =
  | "AUTOMATION_STEP_TYPE_UNSPECIFIED"
  | "AUTOMATION_STEP_TYPE_INVOKE_AGENT"
  | "AUTOMATION_STEP_TYPE_CALL_WEBHOOK"
  | "AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP"
  | "AUTOMATION_STEP_TYPE_CREATE_FORUM_POST";

export type AutomationConcurrencyPolicy =
  | "AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED"
  | "AUTOMATION_CONCURRENCY_POLICY_SKIP"
  | "AUTOMATION_CONCURRENCY_POLICY_QUEUE"
  | "AUTOMATION_CONCURRENCY_POLICY_REPLACE"
  | "AUTOMATION_CONCURRENCY_POLICY_ALLOW";

export type AutomationRunStatus =
  | "AUTOMATION_RUN_STATUS_UNSPECIFIED"
  | "AUTOMATION_RUN_STATUS_RUNNING"
  | "AUTOMATION_RUN_STATUS_SUCCEEDED"
  | "AUTOMATION_RUN_STATUS_FAILED"
  | "AUTOMATION_RUN_STATUS_SKIPPED"
  | "AUTOMATION_RUN_STATUS_CANCELLED";

export type AutomationStepRunStatus =
  | "AUTOMATION_STEP_RUN_STATUS_UNSPECIFIED"
  | "AUTOMATION_STEP_RUN_STATUS_RUNNING"
  | "AUTOMATION_STEP_RUN_STATUS_SUCCEEDED"
  | "AUTOMATION_STEP_RUN_STATUS_FAILED"
  | "AUTOMATION_STEP_RUN_STATUS_SKIPPED"
  | "AUTOMATION_STEP_RUN_STATUS_CANCELLED";

export interface AutomationTrigger {
  type: AutomationTriggerType;
  schedule?: { schedule?: string; timezone?: string };
}

export interface AutomationCondition {
  selector: string;
  operator: AutomationConditionOperator;
  value?: string;
}

export interface AutomationPolicy {
  timeout_seconds?: number;
  retry?: { max_attempts?: number; backoff_seconds?: number };
  concurrency?: AutomationConcurrencyPolicy;
  max_output_bytes?: number;
}

export interface AutomationStep {
  name: string;
  type: AutomationStepType;
  invoke_agent?: { agent_name: string; input?: string; model_override?: string };
  call_webhook?: { url: string; method?: string; payload_json?: string; headers?: Record<string, string> };
  send_notify_group?: { notify_group_name: string; title?: string; message?: string };
  create_forum_post?: { thread_id: string; body: string };
  policy?: AutomationPolicy;
}

export interface Automation {
  name: string;
  enabled?: boolean;
  trigger?: AutomationTrigger;
  conditions?: AutomationCondition[];
  steps?: AutomationStep[];
  policy?: AutomationPolicy;
  metadata?: Record<string, string>;
  created_at?: string;
  updated_at?: string;
  workspace_id?: string;
}

export interface AutomationRun {
  id: string;
  automation_name: string;
  trigger_type: AutomationTriggerType;
  status: AutomationRunStatus;
  trigger_payload_json?: string;
  error?: string;
  started_at?: string;
  finished_at?: string;
  duration_ms?: number;
  workspace_id?: string;
}

export interface AutomationStepRun {
  id: string;
  run_id: string;
  automation_name: string;
  step_name: string;
  step_type: AutomationStepType;
  status: AutomationStepRunStatus;
  attempt_count?: number;
  input_json?: string;
  output_json?: string;
  error?: string;
  invocation_id?: string;
  started_at?: string;
  finished_at?: string;
  duration_ms?: number;
  order?: number;
  truncated?: boolean;
  workspace_id?: string;
}
