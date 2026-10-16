//
//  Generated code. Do not modify.
//  source: agents/v1/workspace_memory.proto
//

import "package:connectrpc/connect.dart" as connect;
import "workspace_memory.pb.dart" as agentsv1workspace_memory;
import "workspace_memory.connect.spec.dart" as specs;

/// WorkspaceMemoryConfigService manages the workspace's mem0 connection. Any
/// member may read it; owners, admins, and global admins may change it.
extension type WorkspaceMemoryConfigServiceClient(
  connect.Transport _transport
) {
  /// GetWorkspaceMemoryConfig returns the config, or an unset `config` when
  /// the workspace has none.
  Future<agentsv1workspace_memory.GetWorkspaceMemoryConfigResponse>
  getWorkspaceMemoryConfig(
    agentsv1workspace_memory.GetWorkspaceMemoryConfigRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceMemoryConfigService.getWorkspaceMemoryConfig,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// PutWorkspaceMemoryConfig creates or replaces the config. When the result
  /// is enabled, the server is probed first: a rejected API key refuses the
  /// save, while an unreachable server saves and reports a warning.
  Future<agentsv1workspace_memory.PutWorkspaceMemoryConfigResponse>
  putWorkspaceMemoryConfig(
    agentsv1workspace_memory.PutWorkspaceMemoryConfigRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceMemoryConfigService.putWorkspaceMemoryConfig,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteWorkspaceMemoryConfig removes the config and its API key. Agents
  /// that enable memory keep running, without memory.
  Future<agentsv1workspace_memory.DeleteWorkspaceMemoryConfigResponse>
  deleteWorkspaceMemoryConfig(
    agentsv1workspace_memory.DeleteWorkspaceMemoryConfigRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceMemoryConfigService.deleteWorkspaceMemoryConfig,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// TestWorkspaceMemoryConnection probes the stored config's server with the
  /// stored API key and reports the outcome.
  Future<agentsv1workspace_memory.TestWorkspaceMemoryConnectionResponse>
  testWorkspaceMemoryConnection(
    agentsv1workspace_memory.TestWorkspaceMemoryConnectionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceMemoryConfigService.testWorkspaceMemoryConnection,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
extension type WorkspaceMemoryServiceClient(connect.Transport _transport) {
  /// ListWorkspaceMemories returns the newest memories of one scope, capped
  /// at `limit` (mem0 OSS listing has no pagination; search reaches the rest).
  Future<agentsv1workspace_memory.ListWorkspaceMemoriesResponse>
  listWorkspaceMemories(
    agentsv1workspace_memory.ListWorkspaceMemoriesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceMemoryService.listWorkspaceMemories,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SearchWorkspaceMemories runs a semantic search in one scope.
  Future<agentsv1workspace_memory.SearchWorkspaceMemoriesResponse>
  searchWorkspaceMemories(
    agentsv1workspace_memory.SearchWorkspaceMemoriesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceMemoryService.searchWorkspaceMemories,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteWorkspaceMemory deletes one memory of this workspace. A memory
  /// that does not exist or belongs to another workspace is NotFound.
  Future<agentsv1workspace_memory.DeleteWorkspaceMemoryResponse>
  deleteWorkspaceMemory(
    agentsv1workspace_memory.DeleteWorkspaceMemoryRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceMemoryService.deleteWorkspaceMemory,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
