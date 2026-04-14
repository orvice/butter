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
  | "REMOTE_AGENT_PROTOCOL_A2A";

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
}

export interface SessionInfo {
  session_id: string;
  app_name: string;
  user_id: string;
  state?: Record<string, unknown>;
  last_update_time?: string;
}

export interface SessionEvent {
  event_id: string;
  invocation_id?: string;
  author?: string;
  branch?: string;
  content_json?: string;
  timestamp?: string;
}

export interface SessionDetail {
  session: SessionInfo;
  events: SessionEvent[];
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
