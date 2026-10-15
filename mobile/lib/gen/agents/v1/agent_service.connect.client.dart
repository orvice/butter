//
//  Generated code. Do not modify.
//  source: agents/v1/agent_service.proto
//

import "package:connectrpc/connect.dart" as connect;
import "agent_service.pb.dart" as agentsv1agent_service;
import "agent_service.connect.spec.dart" as specs;
import "dashboard.pb.dart" as agentsv1dashboard;

/// AgentService manages Agent configurations.
extension type AgentServiceClient(connect.Transport _transport) {
  Future<agentsv1agent_service.ListAgentsResponse> listAgents(
    agentsv1agent_service.ListAgentsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.listAgents,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.GetAgentResponse> getAgent(
    agentsv1agent_service.GetAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.getAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.CreateAgentResponse> createAgent(
    agentsv1agent_service.CreateAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.createAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.UpdateAgentResponse> updateAgent(
    agentsv1agent_service.UpdateAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.updateAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.DeleteAgentResponse> deleteAgent(
    agentsv1agent_service.DeleteAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.deleteAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// InvokeAgent runs an agent once with the given input. If session_id is empty
  /// an ephemeral session id is generated; otherwise the existing session is used.
  Future<agentsv1agent_service.InvokeAgentResponse> invokeAgent(
    agentsv1agent_service.InvokeAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.invokeAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListAgentInvocations returns recorded invocations, optionally filtered by
  /// agent name, with simple opaque-token pagination.
  Future<agentsv1agent_service.ListAgentInvocationsResponse>
  listAgentInvocations(
    agentsv1agent_service.ListAgentInvocationsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.listAgentInvocations,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ReloadAgents drops cached runners and re-resolves agents/MCP/remote-agents
  /// from the current configuration. Exposes the runtime hot-reload trigger
  /// used internally by config mutations to the dashboard "Hot-reload" button.
  Future<agentsv1agent_service.ReloadAgentsResponse> reloadAgents(
    agentsv1agent_service.ReloadAgentsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.reloadAgents,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetAgentRuntimeStatus returns the latest invocation state for an agent,
  /// suitable for the Agents table Status / Last Run columns.
  Future<agentsv1agent_service.GetAgentRuntimeStatusResponse>
  getAgentRuntimeStatus(
    agentsv1agent_service.GetAgentRuntimeStatusRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.getAgentRuntimeStatus,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListAgentRuntimeStatuses returns runtime statuses for the given agent
  /// names in a single round trip. When `names` is empty, statuses for all
  /// configured agents are returned.
  Future<agentsv1agent_service.ListAgentRuntimeStatusesResponse>
  listAgentRuntimeStatuses(
    agentsv1agent_service.ListAgentRuntimeStatusesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.listAgentRuntimeStatuses,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CancelAgentInvocation cancels an in-flight invocation by its ID. The
  /// invocation transitions to FAILED with a cancellation error.
  Future<agentsv1agent_service.CancelAgentInvocationResponse>
  cancelAgentInvocation(
    agentsv1agent_service.CancelAgentInvocationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.cancelAgentInvocation,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// StreamAgent runs an agent and streams progress events back to the
  /// caller. The dashboard chat UI uses this Connect server-stream RPC
  /// (replaces the removed POST /api/chat/stream SSE endpoint). One of the
  /// StreamAgentResponse.event oneof variants is sent for each runner event
  /// (started, mid-stream text delta, full ADK event, final response).
  /// Errors bubble up as the RPC's terminal connect.Error so the client
  /// gets typed cancellation / failed-precondition codes instead of an
  /// ad-hoc string payload.
  Stream<agentsv1agent_service.StreamAgentResponse> streamAgent(
    agentsv1agent_service.StreamAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).server(
      specs.AgentService.streamAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// AssignAgentID is retired (issue #241): the Agent ID rollout is complete,
  /// every Agent carries an immutable agent_id assigned at creation, and the
  /// repository no longer resolves Agents by name. Always returns
  /// Unimplemented.
  @deprecated
  Future<agentsv1agent_service.AssignAgentIDResponse> assignAgentID(
    agentsv1agent_service.AssignAgentIDRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.assignAgentID,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetMigrationReadiness is retired (issue #241). Use VerifyAgentIDCutover
  /// for post-cutover diagnostics. Always returns Unimplemented.
  @deprecated
  Future<agentsv1agent_service.GetMigrationReadinessResponse>
  getMigrationReadiness(
    agentsv1agent_service.GetMigrationReadinessRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.getMigrationReadiness,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// MigrateAgentsV2 is retired (issue #241): the V2 migration observation
  /// window is closed. Use VerifyAgentIDCutover for post-cutover diagnostics.
  /// Always returns Unimplemented.
  @deprecated
  Future<agentsv1agent_service.MigrateAgentsV2Response> migrateAgentsV2(
    agentsv1agent_service.MigrateAgentsV2Request input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.migrateAgentsV2,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// VerifyAgentIDCutover runs the read-only final cutover verifier (issue
  /// #241) across every workspace: it confirms all Agents carry valid,
  /// workspace-unique Agent IDs, no legacy embedded sub_agents or name-based
  /// workflow references remain, all child/workflow references resolve, no
  /// Agent is stuck in MIGRATION_REQUIRED, consumer records (channels, cron
  /// jobs, automations, forum threads) carry Agent IDs, and no runtime-name
  /// conflict violates the runner's naming constraints. Global admins only.
  Future<agentsv1agent_service.VerifyAgentIDCutoverResponse>
  verifyAgentIDCutover(
    agentsv1agent_service.VerifyAgentIDCutoverRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.verifyAgentIDCutover,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateAgentConfiguration is the composite-save command: it coordinates an
  /// Agent operational patch (DB-owned) with Agent Content changes (Git-owned)
  /// under a single durable operation while keeping ownership separate. The
  /// patch is applied with optimistic concurrency (expected_agent_version); a
  /// mismatch aborts. Returns the operation record for status/retry.
  Future<agentsv1agent_service.UpdateAgentConfigurationResponse>
  updateAgentConfiguration(
    agentsv1agent_service.UpdateAgentConfigurationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.updateAgentConfiguration,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RestoreAgent reactivates a soft-deleted (tombstoned) Agent from its
  /// retained configuration and Agent Content, flipping it from DELETED back to
  /// ACTIVE and re-publishing the retained content.
  Future<agentsv1agent_service.RestoreAgentResponse> restoreAgent(
    agentsv1agent_service.RestoreAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.restoreAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetAgentInvocation returns the authoritative state of one invocation by
  /// its ID, scoped by workspace and user ownership.
  Future<agentsv1agent_service.GetAgentInvocationResponse> getAgentInvocation(
    agentsv1agent_service.GetAgentInvocationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.getAgentInvocation,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetAgentOperation returns a durable lifecycle operation record by ID.
  Future<agentsv1agent_service.GetAgentOperationResponse> getAgentOperation(
    agentsv1agent_service.GetAgentOperationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.getAgentOperation,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListAgentOperations lists lifecycle operations in the workspace, optionally
  /// filtered by status, with opaque-token pagination.
  Future<agentsv1agent_service.ListAgentOperationsResponse> listAgentOperations(
    agentsv1agent_service.ListAgentOperationsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.listAgentOperations,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RetryAgentOperation resumes a FAILED lifecycle operation from its first
  /// unfinished step. Steps that already succeeded are skipped (idempotent).
  Future<agentsv1agent_service.RetryAgentOperationResponse> retryAgentOperation(
    agentsv1agent_service.RetryAgentOperationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentService.retryAgentOperation,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}

/// MCPServerService manages MCP Server configurations.
extension type MCPServerServiceClient(connect.Transport _transport) {
  Future<agentsv1agent_service.ListMCPServersResponse> listMCPServers(
    agentsv1agent_service.ListMCPServersRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.listMCPServers,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.GetMCPServerResponse> getMCPServer(
    agentsv1agent_service.GetMCPServerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.getMCPServer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.CreateMCPServerResponse> createMCPServer(
    agentsv1agent_service.CreateMCPServerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.createMCPServer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.UpdateMCPServerResponse> updateMCPServer(
    agentsv1agent_service.UpdateMCPServerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.updateMCPServer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.DeleteMCPServerResponse> deleteMCPServer(
    agentsv1agent_service.DeleteMCPServerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.deleteMCPServer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetMCPServerStatus returns the current runtime status of a configured
  /// MCP server. Implemented in dashboard.proto.
  Future<agentsv1dashboard.GetMCPServerStatusResponse> getMCPServerStatus(
    agentsv1dashboard.GetMCPServerStatusRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.getMCPServerStatus,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListMCPTools enumerates tools exposed by the configured MCP servers.
  /// When server_id is empty, all servers are probed.
  Future<agentsv1dashboard.ListMCPToolsResponse> listMCPTools(
    agentsv1dashboard.ListMCPToolsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.listMCPTools,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// StartMCPServerOAuth prepares a workspace-scoped OAuth2 authorization
  /// flow and returns the user-facing authorization URL.
  Future<agentsv1agent_service.StartMCPServerOAuthResponse> startMCPServerOAuth(
    agentsv1agent_service.StartMCPServerOAuthRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.startMCPServerOAuth,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CompleteMCPServerOAuth completes an OAuth2 authorization code callback.
  Future<agentsv1agent_service.CompleteMCPServerOAuthResponse>
  completeMCPServerOAuth(
    agentsv1agent_service.CompleteMCPServerOAuthRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.completeMCPServerOAuth,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetMCPServerOAuthStatus returns non-secret OAuth connection state.
  Future<agentsv1agent_service.GetMCPServerOAuthStatusResponse>
  getMCPServerOAuthStatus(
    agentsv1agent_service.GetMCPServerOAuthStatusRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.getMCPServerOAuthStatus,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DisconnectMCPServerOAuth removes local OAuth2 credentials for a server.
  Future<agentsv1agent_service.DisconnectMCPServerOAuthResponse>
  disconnectMCPServerOAuth(
    agentsv1agent_service.DisconnectMCPServerOAuthRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.MCPServerService.disconnectMCPServerOAuth,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}

/// GlobalMCPServerService manages workspace-agnostic MCP server presets that
/// admins curate and end users install into their workspace. Browse / install
/// is open to authenticated callers; admin role is required for the
/// create / update / delete mutations.
extension type GlobalMCPServerServiceClient(connect.Transport _transport) {
  Future<agentsv1agent_service.ListGlobalMCPServersResponse>
  listGlobalMCPServers(
    agentsv1agent_service.ListGlobalMCPServersRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GlobalMCPServerService.listGlobalMCPServers,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.CreateGlobalMCPServerResponse>
  createGlobalMCPServer(
    agentsv1agent_service.CreateGlobalMCPServerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GlobalMCPServerService.createGlobalMCPServer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.UpdateGlobalMCPServerResponse>
  updateGlobalMCPServer(
    agentsv1agent_service.UpdateGlobalMCPServerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GlobalMCPServerService.updateGlobalMCPServer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.DeleteGlobalMCPServerResponse>
  deleteGlobalMCPServer(
    agentsv1agent_service.DeleteGlobalMCPServerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GlobalMCPServerService.deleteGlobalMCPServer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// InstallGlobalMCPServer clones a preset into the caller's workspace.
  /// Admins may pass a different target workspace via `workspace_id`; the
  /// server audit-logs cross-workspace installs.
  Future<agentsv1agent_service.InstallGlobalMCPServerResponse>
  installGlobalMCPServer(
    agentsv1agent_service.InstallGlobalMCPServerRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GlobalMCPServerService.installGlobalMCPServer,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}

/// ModelProviderService manages LLM provider configurations.
extension type ModelProviderServiceClient(connect.Transport _transport) {
  Future<agentsv1agent_service.ListModelProvidersResponse> listModelProviders(
    agentsv1agent_service.ListModelProvidersRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ModelProviderService.listModelProviders,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.GetModelProviderResponse> getModelProvider(
    agentsv1agent_service.GetModelProviderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ModelProviderService.getModelProvider,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.CreateModelProviderResponse> createModelProvider(
    agentsv1agent_service.CreateModelProviderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ModelProviderService.createModelProvider,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.UpdateModelProviderResponse> updateModelProvider(
    agentsv1agent_service.UpdateModelProviderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ModelProviderService.updateModelProvider,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.DeleteModelProviderResponse> deleteModelProvider(
    agentsv1agent_service.DeleteModelProviderRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ModelProviderService.deleteModelProvider,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}

/// NotifyGroupService manages outbound notification groups.
extension type NotifyGroupServiceClient(connect.Transport _transport) {
  Future<agentsv1agent_service.ListNotifyGroupsResponse> listNotifyGroups(
    agentsv1agent_service.ListNotifyGroupsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.NotifyGroupService.listNotifyGroups,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.GetNotifyGroupResponse> getNotifyGroup(
    agentsv1agent_service.GetNotifyGroupRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.NotifyGroupService.getNotifyGroup,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.CreateNotifyGroupResponse> createNotifyGroup(
    agentsv1agent_service.CreateNotifyGroupRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.NotifyGroupService.createNotifyGroup,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.UpdateNotifyGroupResponse> updateNotifyGroup(
    agentsv1agent_service.UpdateNotifyGroupRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.NotifyGroupService.updateNotifyGroup,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.DeleteNotifyGroupResponse> deleteNotifyGroup(
    agentsv1agent_service.DeleteNotifyGroupRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.NotifyGroupService.deleteNotifyGroup,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}

/// RemoteAgentService manages Remote Agent configurations.
extension type RemoteAgentServiceClient(connect.Transport _transport) {
  Future<agentsv1agent_service.ListRemoteAgentsResponse> listRemoteAgents(
    agentsv1agent_service.ListRemoteAgentsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RemoteAgentService.listRemoteAgents,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.GetRemoteAgentResponse> getRemoteAgent(
    agentsv1agent_service.GetRemoteAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RemoteAgentService.getRemoteAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.CreateRemoteAgentResponse> createRemoteAgent(
    agentsv1agent_service.CreateRemoteAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RemoteAgentService.createRemoteAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.UpdateRemoteAgentResponse> updateRemoteAgent(
    agentsv1agent_service.UpdateRemoteAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RemoteAgentService.updateRemoteAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.DeleteRemoteAgentResponse> deleteRemoteAgent(
    agentsv1agent_service.DeleteRemoteAgentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RemoteAgentService.deleteRemoteAgent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetRemoteAgentStatus probes the configured remote agent endpoint and
  /// returns its liveness state (A2A: HTTP agent.json; daemon: registry).
  Future<agentsv1dashboard.GetRemoteAgentStatusResponse> getRemoteAgentStatus(
    agentsv1dashboard.GetRemoteAgentStatusRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RemoteAgentService.getRemoteAgentStatus,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}

/// ChannelService manages AgentChannel configurations.
extension type ChannelServiceClient(connect.Transport _transport) {
  Future<agentsv1agent_service.ListChannelsResponse> listChannels(
    agentsv1agent_service.ListChannelsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ChannelService.listChannels,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.GetChannelResponse> getChannel(
    agentsv1agent_service.GetChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ChannelService.getChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.CreateChannelResponse> createChannel(
    agentsv1agent_service.CreateChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ChannelService.createChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.UpdateChannelResponse> updateChannel(
    agentsv1agent_service.UpdateChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ChannelService.updateChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_service.DeleteChannelResponse> deleteChannel(
    agentsv1agent_service.DeleteChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ChannelService.deleteChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetChannelStatus returns the current runtime status of a channel poller.
  /// Implemented in dashboard.proto.
  Future<agentsv1dashboard.GetChannelStatusResponse> getChannelStatus(
    agentsv1dashboard.GetChannelStatusRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ChannelService.getChannelStatus,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RestartChannel reloads the channel manager. Effectively bounces all pollers.
  Future<agentsv1agent_service.RestartChannelResponse> restartChannel(
    agentsv1agent_service.RestartChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ChannelService.restartChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// PauseChannel disables the named channel and reloads the channel manager so
  /// its poller stops. The configuration `enabled` flag is set to false.
  Future<agentsv1agent_service.PauseChannelResponse> pauseChannel(
    agentsv1agent_service.PauseChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ChannelService.pauseChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ResumeChannel re-enables the named channel and reloads the channel manager.
  Future<agentsv1agent_service.ResumeChannelResponse> resumeChannel(
    agentsv1agent_service.ResumeChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ChannelService.resumeChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
extension type SessionServiceClient(connect.Transport _transport) {
  /// CreateSession creates a new session for the given agent.
  Future<agentsv1agent_service.CreateSessionResponse> createSession(
    agentsv1agent_service.CreateSessionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SessionService.createSession,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetSession retrieves a session by ID, optionally with recent events.
  Future<agentsv1agent_service.GetSessionResponse> getSession(
    agentsv1agent_service.GetSessionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SessionService.getSession,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListSessions lists sessions for a given app and user.
  Future<agentsv1agent_service.ListSessionsResponse> listSessions(
    agentsv1agent_service.ListSessionsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SessionService.listSessions,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteSession deletes a session and its events.
  Future<agentsv1agent_service.DeleteSessionResponse> deleteSession(
    agentsv1agent_service.DeleteSessionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SessionService.deleteSession,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ReplySession sends a user message to an existing session and returns the agent response.
  Future<agentsv1agent_service.ReplySessionResponse> replySession(
    agentsv1agent_service.ReplySessionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SessionService.replySession,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateSessionTitle sets a first-class title on a session.
  Future<agentsv1agent_service.UpdateSessionTitleResponse> updateSessionTitle(
    agentsv1agent_service.UpdateSessionTitleRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SessionService.updateSessionTitle,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GenerateSessionTitle derives a best-effort title from the first
  /// conversation turn and persists it only while no effective title exists.
  /// Manual, legacy, and concurrent titles always win. Returns the effective
  /// SessionInfo and whether this call actually generated a new title.
  Future<agentsv1agent_service.GenerateSessionTitleResponse>
  generateSessionTitle(
    agentsv1agent_service.GenerateSessionTitleRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SessionService.generateSessionTitle,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// MarkSessionRead records that the authenticated user has viewed the
  /// session up to (and including) the given event. The server persists the
  /// read cursor; subsequent ListSessions / GetSession responses reflect it
  /// through SessionInfo.unread. Only the session owner may mark read;
  /// another user's attempt returns NotFound.
  Future<agentsv1agent_service.MarkSessionReadResponse> markSessionRead(
    agentsv1agent_service.MarkSessionReadRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.SessionService.markSessionRead,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
