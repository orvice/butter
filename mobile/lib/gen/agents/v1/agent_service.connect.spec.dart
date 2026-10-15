//
//  Generated code. Do not modify.
//  source: agents/v1/agent_service.proto
//

import "package:connectrpc/connect.dart" as connect;
import "agent_service.pb.dart" as agentsv1agent_service;
import "dashboard.pb.dart" as agentsv1dashboard;

/// AgentService manages Agent configurations.
abstract final class AgentService {
  /// Fully-qualified name of the AgentService service.
  static const name = 'agents.v1.AgentService';

  static const listAgents = connect.Spec(
    '/$name/ListAgents',
    connect.StreamType.unary,
    agentsv1agent_service.ListAgentsRequest.new,
    agentsv1agent_service.ListAgentsResponse.new,
  );

  static const getAgent = connect.Spec(
    '/$name/GetAgent',
    connect.StreamType.unary,
    agentsv1agent_service.GetAgentRequest.new,
    agentsv1agent_service.GetAgentResponse.new,
  );

  static const createAgent = connect.Spec(
    '/$name/CreateAgent',
    connect.StreamType.unary,
    agentsv1agent_service.CreateAgentRequest.new,
    agentsv1agent_service.CreateAgentResponse.new,
  );

  static const updateAgent = connect.Spec(
    '/$name/UpdateAgent',
    connect.StreamType.unary,
    agentsv1agent_service.UpdateAgentRequest.new,
    agentsv1agent_service.UpdateAgentResponse.new,
  );

  static const deleteAgent = connect.Spec(
    '/$name/DeleteAgent',
    connect.StreamType.unary,
    agentsv1agent_service.DeleteAgentRequest.new,
    agentsv1agent_service.DeleteAgentResponse.new,
  );

  /// InvokeAgent runs an agent once with the given input. If session_id is empty
  /// an ephemeral session id is generated; otherwise the existing session is used.
  static const invokeAgent = connect.Spec(
    '/$name/InvokeAgent',
    connect.StreamType.unary,
    agentsv1agent_service.InvokeAgentRequest.new,
    agentsv1agent_service.InvokeAgentResponse.new,
  );

  /// ListAgentInvocations returns recorded invocations, optionally filtered by
  /// agent name, with simple opaque-token pagination.
  static const listAgentInvocations = connect.Spec(
    '/$name/ListAgentInvocations',
    connect.StreamType.unary,
    agentsv1agent_service.ListAgentInvocationsRequest.new,
    agentsv1agent_service.ListAgentInvocationsResponse.new,
  );

  /// ReloadAgents drops cached runners and re-resolves agents/MCP/remote-agents
  /// from the current configuration. Exposes the runtime hot-reload trigger
  /// used internally by config mutations to the dashboard "Hot-reload" button.
  static const reloadAgents = connect.Spec(
    '/$name/ReloadAgents',
    connect.StreamType.unary,
    agentsv1agent_service.ReloadAgentsRequest.new,
    agentsv1agent_service.ReloadAgentsResponse.new,
  );

  /// GetAgentRuntimeStatus returns the latest invocation state for an agent,
  /// suitable for the Agents table Status / Last Run columns.
  static const getAgentRuntimeStatus = connect.Spec(
    '/$name/GetAgentRuntimeStatus',
    connect.StreamType.unary,
    agentsv1agent_service.GetAgentRuntimeStatusRequest.new,
    agentsv1agent_service.GetAgentRuntimeStatusResponse.new,
  );

  /// ListAgentRuntimeStatuses returns runtime statuses for the given agent
  /// names in a single round trip. When `names` is empty, statuses for all
  /// configured agents are returned.
  static const listAgentRuntimeStatuses = connect.Spec(
    '/$name/ListAgentRuntimeStatuses',
    connect.StreamType.unary,
    agentsv1agent_service.ListAgentRuntimeStatusesRequest.new,
    agentsv1agent_service.ListAgentRuntimeStatusesResponse.new,
  );

  /// CancelAgentInvocation cancels an in-flight invocation by its ID. The
  /// invocation transitions to FAILED with a cancellation error.
  static const cancelAgentInvocation = connect.Spec(
    '/$name/CancelAgentInvocation',
    connect.StreamType.unary,
    agentsv1agent_service.CancelAgentInvocationRequest.new,
    agentsv1agent_service.CancelAgentInvocationResponse.new,
  );

  /// StreamAgent runs an agent and streams progress events back to the
  /// caller. The dashboard chat UI uses this Connect server-stream RPC
  /// (replaces the removed POST /api/chat/stream SSE endpoint). One of the
  /// StreamAgentResponse.event oneof variants is sent for each runner event
  /// (started, mid-stream text delta, full ADK event, final response).
  /// Errors bubble up as the RPC's terminal connect.Error so the client
  /// gets typed cancellation / failed-precondition codes instead of an
  /// ad-hoc string payload.
  static const streamAgent = connect.Spec(
    '/$name/StreamAgent',
    connect.StreamType.server,
    agentsv1agent_service.StreamAgentRequest.new,
    agentsv1agent_service.StreamAgentResponse.new,
  );

  /// AssignAgentID is retired (issue #241): the Agent ID rollout is complete,
  /// every Agent carries an immutable agent_id assigned at creation, and the
  /// repository no longer resolves Agents by name. Always returns
  /// Unimplemented.
  static const assignAgentID = connect.Spec(
    '/$name/AssignAgentID',
    connect.StreamType.unary,
    agentsv1agent_service.AssignAgentIDRequest.new,
    agentsv1agent_service.AssignAgentIDResponse.new,
  );

  /// GetMigrationReadiness is retired (issue #241). Use VerifyAgentIDCutover
  /// for post-cutover diagnostics. Always returns Unimplemented.
  static const getMigrationReadiness = connect.Spec(
    '/$name/GetMigrationReadiness',
    connect.StreamType.unary,
    agentsv1agent_service.GetMigrationReadinessRequest.new,
    agentsv1agent_service.GetMigrationReadinessResponse.new,
  );

  /// MigrateAgentsV2 is retired (issue #241): the V2 migration observation
  /// window is closed. Use VerifyAgentIDCutover for post-cutover diagnostics.
  /// Always returns Unimplemented.
  static const migrateAgentsV2 = connect.Spec(
    '/$name/MigrateAgentsV2',
    connect.StreamType.unary,
    agentsv1agent_service.MigrateAgentsV2Request.new,
    agentsv1agent_service.MigrateAgentsV2Response.new,
  );

  /// VerifyAgentIDCutover runs the read-only final cutover verifier (issue
  /// #241) across every workspace: it confirms all Agents carry valid,
  /// workspace-unique Agent IDs, no legacy embedded sub_agents or name-based
  /// workflow references remain, all child/workflow references resolve, no
  /// Agent is stuck in MIGRATION_REQUIRED, consumer records (channels, cron
  /// jobs, automations, forum threads) carry Agent IDs, and no runtime-name
  /// conflict violates the runner's naming constraints. Global admins only.
  static const verifyAgentIDCutover = connect.Spec(
    '/$name/VerifyAgentIDCutover',
    connect.StreamType.unary,
    agentsv1agent_service.VerifyAgentIDCutoverRequest.new,
    agentsv1agent_service.VerifyAgentIDCutoverResponse.new,
  );

  /// UpdateAgentConfiguration is the composite-save command: it coordinates an
  /// Agent operational patch (DB-owned) with Agent Content changes (Git-owned)
  /// under a single durable operation while keeping ownership separate. The
  /// patch is applied with optimistic concurrency (expected_agent_version); a
  /// mismatch aborts. Returns the operation record for status/retry.
  static const updateAgentConfiguration = connect.Spec(
    '/$name/UpdateAgentConfiguration',
    connect.StreamType.unary,
    agentsv1agent_service.UpdateAgentConfigurationRequest.new,
    agentsv1agent_service.UpdateAgentConfigurationResponse.new,
  );

  /// RestoreAgent reactivates a soft-deleted (tombstoned) Agent from its
  /// retained configuration and Agent Content, flipping it from DELETED back to
  /// ACTIVE and re-publishing the retained content.
  static const restoreAgent = connect.Spec(
    '/$name/RestoreAgent',
    connect.StreamType.unary,
    agentsv1agent_service.RestoreAgentRequest.new,
    agentsv1agent_service.RestoreAgentResponse.new,
  );

  /// GetAgentInvocation returns the authoritative state of one invocation by
  /// its ID, scoped by workspace and user ownership.
  static const getAgentInvocation = connect.Spec(
    '/$name/GetAgentInvocation',
    connect.StreamType.unary,
    agentsv1agent_service.GetAgentInvocationRequest.new,
    agentsv1agent_service.GetAgentInvocationResponse.new,
  );

  /// GetAgentOperation returns a durable lifecycle operation record by ID.
  static const getAgentOperation = connect.Spec(
    '/$name/GetAgentOperation',
    connect.StreamType.unary,
    agentsv1agent_service.GetAgentOperationRequest.new,
    agentsv1agent_service.GetAgentOperationResponse.new,
  );

  /// ListAgentOperations lists lifecycle operations in the workspace, optionally
  /// filtered by status, with opaque-token pagination.
  static const listAgentOperations = connect.Spec(
    '/$name/ListAgentOperations',
    connect.StreamType.unary,
    agentsv1agent_service.ListAgentOperationsRequest.new,
    agentsv1agent_service.ListAgentOperationsResponse.new,
  );

  /// RetryAgentOperation resumes a FAILED lifecycle operation from its first
  /// unfinished step. Steps that already succeeded are skipped (idempotent).
  static const retryAgentOperation = connect.Spec(
    '/$name/RetryAgentOperation',
    connect.StreamType.unary,
    agentsv1agent_service.RetryAgentOperationRequest.new,
    agentsv1agent_service.RetryAgentOperationResponse.new,
  );
}

/// MCPServerService manages MCP Server configurations.
abstract final class MCPServerService {
  /// Fully-qualified name of the MCPServerService service.
  static const name = 'agents.v1.MCPServerService';

  static const listMCPServers = connect.Spec(
    '/$name/ListMCPServers',
    connect.StreamType.unary,
    agentsv1agent_service.ListMCPServersRequest.new,
    agentsv1agent_service.ListMCPServersResponse.new,
  );

  static const getMCPServer = connect.Spec(
    '/$name/GetMCPServer',
    connect.StreamType.unary,
    agentsv1agent_service.GetMCPServerRequest.new,
    agentsv1agent_service.GetMCPServerResponse.new,
  );

  static const createMCPServer = connect.Spec(
    '/$name/CreateMCPServer',
    connect.StreamType.unary,
    agentsv1agent_service.CreateMCPServerRequest.new,
    agentsv1agent_service.CreateMCPServerResponse.new,
  );

  static const updateMCPServer = connect.Spec(
    '/$name/UpdateMCPServer',
    connect.StreamType.unary,
    agentsv1agent_service.UpdateMCPServerRequest.new,
    agentsv1agent_service.UpdateMCPServerResponse.new,
  );

  static const deleteMCPServer = connect.Spec(
    '/$name/DeleteMCPServer',
    connect.StreamType.unary,
    agentsv1agent_service.DeleteMCPServerRequest.new,
    agentsv1agent_service.DeleteMCPServerResponse.new,
  );

  /// GetMCPServerStatus returns the current runtime status of a configured
  /// MCP server. Implemented in dashboard.proto.
  static const getMCPServerStatus = connect.Spec(
    '/$name/GetMCPServerStatus',
    connect.StreamType.unary,
    agentsv1dashboard.GetMCPServerStatusRequest.new,
    agentsv1dashboard.GetMCPServerStatusResponse.new,
  );

  /// ListMCPTools enumerates tools exposed by the configured MCP servers.
  /// When server_id is empty, all servers are probed.
  static const listMCPTools = connect.Spec(
    '/$name/ListMCPTools',
    connect.StreamType.unary,
    agentsv1dashboard.ListMCPToolsRequest.new,
    agentsv1dashboard.ListMCPToolsResponse.new,
  );

  /// StartMCPServerOAuth prepares a workspace-scoped OAuth2 authorization
  /// flow and returns the user-facing authorization URL.
  static const startMCPServerOAuth = connect.Spec(
    '/$name/StartMCPServerOAuth',
    connect.StreamType.unary,
    agentsv1agent_service.StartMCPServerOAuthRequest.new,
    agentsv1agent_service.StartMCPServerOAuthResponse.new,
  );

  /// CompleteMCPServerOAuth completes an OAuth2 authorization code callback.
  static const completeMCPServerOAuth = connect.Spec(
    '/$name/CompleteMCPServerOAuth',
    connect.StreamType.unary,
    agentsv1agent_service.CompleteMCPServerOAuthRequest.new,
    agentsv1agent_service.CompleteMCPServerOAuthResponse.new,
  );

  /// GetMCPServerOAuthStatus returns non-secret OAuth connection state.
  static const getMCPServerOAuthStatus = connect.Spec(
    '/$name/GetMCPServerOAuthStatus',
    connect.StreamType.unary,
    agentsv1agent_service.GetMCPServerOAuthStatusRequest.new,
    agentsv1agent_service.GetMCPServerOAuthStatusResponse.new,
  );

  /// DisconnectMCPServerOAuth removes local OAuth2 credentials for a server.
  static const disconnectMCPServerOAuth = connect.Spec(
    '/$name/DisconnectMCPServerOAuth',
    connect.StreamType.unary,
    agentsv1agent_service.DisconnectMCPServerOAuthRequest.new,
    agentsv1agent_service.DisconnectMCPServerOAuthResponse.new,
  );
}

/// GlobalMCPServerService manages workspace-agnostic MCP server presets that
/// admins curate and end users install into their workspace. Browse / install
/// is open to authenticated callers; admin role is required for the
/// create / update / delete mutations.
abstract final class GlobalMCPServerService {
  /// Fully-qualified name of the GlobalMCPServerService service.
  static const name = 'agents.v1.GlobalMCPServerService';

  static const listGlobalMCPServers = connect.Spec(
    '/$name/ListGlobalMCPServers',
    connect.StreamType.unary,
    agentsv1agent_service.ListGlobalMCPServersRequest.new,
    agentsv1agent_service.ListGlobalMCPServersResponse.new,
  );

  static const createGlobalMCPServer = connect.Spec(
    '/$name/CreateGlobalMCPServer',
    connect.StreamType.unary,
    agentsv1agent_service.CreateGlobalMCPServerRequest.new,
    agentsv1agent_service.CreateGlobalMCPServerResponse.new,
  );

  static const updateGlobalMCPServer = connect.Spec(
    '/$name/UpdateGlobalMCPServer',
    connect.StreamType.unary,
    agentsv1agent_service.UpdateGlobalMCPServerRequest.new,
    agentsv1agent_service.UpdateGlobalMCPServerResponse.new,
  );

  static const deleteGlobalMCPServer = connect.Spec(
    '/$name/DeleteGlobalMCPServer',
    connect.StreamType.unary,
    agentsv1agent_service.DeleteGlobalMCPServerRequest.new,
    agentsv1agent_service.DeleteGlobalMCPServerResponse.new,
  );

  /// InstallGlobalMCPServer clones a preset into the caller's workspace.
  /// Admins may pass a different target workspace via `workspace_id`; the
  /// server audit-logs cross-workspace installs.
  static const installGlobalMCPServer = connect.Spec(
    '/$name/InstallGlobalMCPServer',
    connect.StreamType.unary,
    agentsv1agent_service.InstallGlobalMCPServerRequest.new,
    agentsv1agent_service.InstallGlobalMCPServerResponse.new,
  );
}

/// ModelProviderService manages LLM provider configurations.
abstract final class ModelProviderService {
  /// Fully-qualified name of the ModelProviderService service.
  static const name = 'agents.v1.ModelProviderService';

  static const listModelProviders = connect.Spec(
    '/$name/ListModelProviders',
    connect.StreamType.unary,
    agentsv1agent_service.ListModelProvidersRequest.new,
    agentsv1agent_service.ListModelProvidersResponse.new,
  );

  static const getModelProvider = connect.Spec(
    '/$name/GetModelProvider',
    connect.StreamType.unary,
    agentsv1agent_service.GetModelProviderRequest.new,
    agentsv1agent_service.GetModelProviderResponse.new,
  );

  static const createModelProvider = connect.Spec(
    '/$name/CreateModelProvider',
    connect.StreamType.unary,
    agentsv1agent_service.CreateModelProviderRequest.new,
    agentsv1agent_service.CreateModelProviderResponse.new,
  );

  static const updateModelProvider = connect.Spec(
    '/$name/UpdateModelProvider',
    connect.StreamType.unary,
    agentsv1agent_service.UpdateModelProviderRequest.new,
    agentsv1agent_service.UpdateModelProviderResponse.new,
  );

  static const deleteModelProvider = connect.Spec(
    '/$name/DeleteModelProvider',
    connect.StreamType.unary,
    agentsv1agent_service.DeleteModelProviderRequest.new,
    agentsv1agent_service.DeleteModelProviderResponse.new,
  );
}

/// NotifyGroupService manages outbound notification groups.
abstract final class NotifyGroupService {
  /// Fully-qualified name of the NotifyGroupService service.
  static const name = 'agents.v1.NotifyGroupService';

  static const listNotifyGroups = connect.Spec(
    '/$name/ListNotifyGroups',
    connect.StreamType.unary,
    agentsv1agent_service.ListNotifyGroupsRequest.new,
    agentsv1agent_service.ListNotifyGroupsResponse.new,
  );

  static const getNotifyGroup = connect.Spec(
    '/$name/GetNotifyGroup',
    connect.StreamType.unary,
    agentsv1agent_service.GetNotifyGroupRequest.new,
    agentsv1agent_service.GetNotifyGroupResponse.new,
  );

  static const createNotifyGroup = connect.Spec(
    '/$name/CreateNotifyGroup',
    connect.StreamType.unary,
    agentsv1agent_service.CreateNotifyGroupRequest.new,
    agentsv1agent_service.CreateNotifyGroupResponse.new,
  );

  static const updateNotifyGroup = connect.Spec(
    '/$name/UpdateNotifyGroup',
    connect.StreamType.unary,
    agentsv1agent_service.UpdateNotifyGroupRequest.new,
    agentsv1agent_service.UpdateNotifyGroupResponse.new,
  );

  static const deleteNotifyGroup = connect.Spec(
    '/$name/DeleteNotifyGroup',
    connect.StreamType.unary,
    agentsv1agent_service.DeleteNotifyGroupRequest.new,
    agentsv1agent_service.DeleteNotifyGroupResponse.new,
  );
}

/// RemoteAgentService manages Remote Agent configurations.
abstract final class RemoteAgentService {
  /// Fully-qualified name of the RemoteAgentService service.
  static const name = 'agents.v1.RemoteAgentService';

  static const listRemoteAgents = connect.Spec(
    '/$name/ListRemoteAgents',
    connect.StreamType.unary,
    agentsv1agent_service.ListRemoteAgentsRequest.new,
    agentsv1agent_service.ListRemoteAgentsResponse.new,
  );

  static const getRemoteAgent = connect.Spec(
    '/$name/GetRemoteAgent',
    connect.StreamType.unary,
    agentsv1agent_service.GetRemoteAgentRequest.new,
    agentsv1agent_service.GetRemoteAgentResponse.new,
  );

  static const createRemoteAgent = connect.Spec(
    '/$name/CreateRemoteAgent',
    connect.StreamType.unary,
    agentsv1agent_service.CreateRemoteAgentRequest.new,
    agentsv1agent_service.CreateRemoteAgentResponse.new,
  );

  static const updateRemoteAgent = connect.Spec(
    '/$name/UpdateRemoteAgent',
    connect.StreamType.unary,
    agentsv1agent_service.UpdateRemoteAgentRequest.new,
    agentsv1agent_service.UpdateRemoteAgentResponse.new,
  );

  static const deleteRemoteAgent = connect.Spec(
    '/$name/DeleteRemoteAgent',
    connect.StreamType.unary,
    agentsv1agent_service.DeleteRemoteAgentRequest.new,
    agentsv1agent_service.DeleteRemoteAgentResponse.new,
  );

  /// GetRemoteAgentStatus probes the configured remote agent endpoint and
  /// returns its liveness state (A2A: HTTP agent.json; daemon: registry).
  static const getRemoteAgentStatus = connect.Spec(
    '/$name/GetRemoteAgentStatus',
    connect.StreamType.unary,
    agentsv1dashboard.GetRemoteAgentStatusRequest.new,
    agentsv1dashboard.GetRemoteAgentStatusResponse.new,
  );
}

/// ChannelService manages AgentChannel configurations.
abstract final class ChannelService {
  /// Fully-qualified name of the ChannelService service.
  static const name = 'agents.v1.ChannelService';

  static const listChannels = connect.Spec(
    '/$name/ListChannels',
    connect.StreamType.unary,
    agentsv1agent_service.ListChannelsRequest.new,
    agentsv1agent_service.ListChannelsResponse.new,
  );

  static const getChannel = connect.Spec(
    '/$name/GetChannel',
    connect.StreamType.unary,
    agentsv1agent_service.GetChannelRequest.new,
    agentsv1agent_service.GetChannelResponse.new,
  );

  static const createChannel = connect.Spec(
    '/$name/CreateChannel',
    connect.StreamType.unary,
    agentsv1agent_service.CreateChannelRequest.new,
    agentsv1agent_service.CreateChannelResponse.new,
  );

  static const updateChannel = connect.Spec(
    '/$name/UpdateChannel',
    connect.StreamType.unary,
    agentsv1agent_service.UpdateChannelRequest.new,
    agentsv1agent_service.UpdateChannelResponse.new,
  );

  static const deleteChannel = connect.Spec(
    '/$name/DeleteChannel',
    connect.StreamType.unary,
    agentsv1agent_service.DeleteChannelRequest.new,
    agentsv1agent_service.DeleteChannelResponse.new,
  );

  /// GetChannelStatus returns the current runtime status of a channel poller.
  /// Implemented in dashboard.proto.
  static const getChannelStatus = connect.Spec(
    '/$name/GetChannelStatus',
    connect.StreamType.unary,
    agentsv1dashboard.GetChannelStatusRequest.new,
    agentsv1dashboard.GetChannelStatusResponse.new,
  );

  /// RestartChannel reloads the channel manager. Effectively bounces all pollers.
  static const restartChannel = connect.Spec(
    '/$name/RestartChannel',
    connect.StreamType.unary,
    agentsv1agent_service.RestartChannelRequest.new,
    agentsv1agent_service.RestartChannelResponse.new,
  );

  /// PauseChannel disables the named channel and reloads the channel manager so
  /// its poller stops. The configuration `enabled` flag is set to false.
  static const pauseChannel = connect.Spec(
    '/$name/PauseChannel',
    connect.StreamType.unary,
    agentsv1agent_service.PauseChannelRequest.new,
    agentsv1agent_service.PauseChannelResponse.new,
  );

  /// ResumeChannel re-enables the named channel and reloads the channel manager.
  static const resumeChannel = connect.Spec(
    '/$name/ResumeChannel',
    connect.StreamType.unary,
    agentsv1agent_service.ResumeChannelRequest.new,
    agentsv1agent_service.ResumeChannelResponse.new,
  );
}

abstract final class SessionService {
  /// Fully-qualified name of the SessionService service.
  static const name = 'agents.v1.SessionService';

  /// CreateSession creates a new session for the given agent.
  static const createSession = connect.Spec(
    '/$name/CreateSession',
    connect.StreamType.unary,
    agentsv1agent_service.CreateSessionRequest.new,
    agentsv1agent_service.CreateSessionResponse.new,
  );

  /// GetSession retrieves a session by ID, optionally with recent events.
  static const getSession = connect.Spec(
    '/$name/GetSession',
    connect.StreamType.unary,
    agentsv1agent_service.GetSessionRequest.new,
    agentsv1agent_service.GetSessionResponse.new,
  );

  /// ListSessions lists sessions for a given app and user.
  static const listSessions = connect.Spec(
    '/$name/ListSessions',
    connect.StreamType.unary,
    agentsv1agent_service.ListSessionsRequest.new,
    agentsv1agent_service.ListSessionsResponse.new,
  );

  /// DeleteSession deletes a session and its events.
  static const deleteSession = connect.Spec(
    '/$name/DeleteSession',
    connect.StreamType.unary,
    agentsv1agent_service.DeleteSessionRequest.new,
    agentsv1agent_service.DeleteSessionResponse.new,
  );

  /// ReplySession sends a user message to an existing session and returns the agent response.
  static const replySession = connect.Spec(
    '/$name/ReplySession',
    connect.StreamType.unary,
    agentsv1agent_service.ReplySessionRequest.new,
    agentsv1agent_service.ReplySessionResponse.new,
  );

  /// UpdateSessionTitle sets a first-class title on a session.
  static const updateSessionTitle = connect.Spec(
    '/$name/UpdateSessionTitle',
    connect.StreamType.unary,
    agentsv1agent_service.UpdateSessionTitleRequest.new,
    agentsv1agent_service.UpdateSessionTitleResponse.new,
  );

  /// GenerateSessionTitle derives a best-effort title from the first
  /// conversation turn and persists it only while no effective title exists.
  /// Manual, legacy, and concurrent titles always win. Returns the effective
  /// SessionInfo and whether this call actually generated a new title.
  static const generateSessionTitle = connect.Spec(
    '/$name/GenerateSessionTitle',
    connect.StreamType.unary,
    agentsv1agent_service.GenerateSessionTitleRequest.new,
    agentsv1agent_service.GenerateSessionTitleResponse.new,
  );

  /// MarkSessionRead records that the authenticated user has viewed the
  /// session up to (and including) the given event. The server persists the
  /// read cursor; subsequent ListSessions / GetSession responses reflect it
  /// through SessionInfo.unread. Only the session owner may mark read;
  /// another user's attempt returns NotFound.
  static const markSessionRead = connect.Spec(
    '/$name/MarkSessionRead',
    connect.StreamType.unary,
    agentsv1agent_service.MarkSessionReadRequest.new,
    agentsv1agent_service.MarkSessionReadResponse.new,
  );
}
