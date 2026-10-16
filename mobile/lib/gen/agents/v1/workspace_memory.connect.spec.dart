//
//  Generated code. Do not modify.
//  source: agents/v1/workspace_memory.proto
//

import "package:connectrpc/connect.dart" as connect;
import "workspace_memory.pb.dart" as agentsv1workspace_memory;

/// WorkspaceMemoryConfigService manages the workspace's mem0 connection. Any
/// member may read it; owners, admins, and global admins may change it.
abstract final class WorkspaceMemoryConfigService {
  /// Fully-qualified name of the WorkspaceMemoryConfigService service.
  static const name = 'agents.v1.WorkspaceMemoryConfigService';

  /// GetWorkspaceMemoryConfig returns the config, or an unset `config` when
  /// the workspace has none.
  static const getWorkspaceMemoryConfig = connect.Spec(
    '/$name/GetWorkspaceMemoryConfig',
    connect.StreamType.unary,
    agentsv1workspace_memory.GetWorkspaceMemoryConfigRequest.new,
    agentsv1workspace_memory.GetWorkspaceMemoryConfigResponse.new,
  );

  /// PutWorkspaceMemoryConfig creates or replaces the config. When the result
  /// is enabled, the server is probed first: a rejected API key refuses the
  /// save, while an unreachable server saves and reports a warning.
  static const putWorkspaceMemoryConfig = connect.Spec(
    '/$name/PutWorkspaceMemoryConfig',
    connect.StreamType.unary,
    agentsv1workspace_memory.PutWorkspaceMemoryConfigRequest.new,
    agentsv1workspace_memory.PutWorkspaceMemoryConfigResponse.new,
  );

  /// DeleteWorkspaceMemoryConfig removes the config and its API key. Agents
  /// that enable memory keep running, without memory.
  static const deleteWorkspaceMemoryConfig = connect.Spec(
    '/$name/DeleteWorkspaceMemoryConfig',
    connect.StreamType.unary,
    agentsv1workspace_memory.DeleteWorkspaceMemoryConfigRequest.new,
    agentsv1workspace_memory.DeleteWorkspaceMemoryConfigResponse.new,
  );

  /// TestWorkspaceMemoryConnection probes the stored config's server with the
  /// stored API key and reports the outcome.
  static const testWorkspaceMemoryConnection = connect.Spec(
    '/$name/TestWorkspaceMemoryConnection',
    connect.StreamType.unary,
    agentsv1workspace_memory.TestWorkspaceMemoryConnectionRequest.new,
    agentsv1workspace_memory.TestWorkspaceMemoryConnectionResponse.new,
  );
}

abstract final class WorkspaceMemoryService {
  /// Fully-qualified name of the WorkspaceMemoryService service.
  static const name = 'agents.v1.WorkspaceMemoryService';

  /// ListWorkspaceMemories returns the newest memories of one scope, capped
  /// at `limit` (mem0 OSS listing has no pagination; search reaches the rest).
  static const listWorkspaceMemories = connect.Spec(
    '/$name/ListWorkspaceMemories',
    connect.StreamType.unary,
    agentsv1workspace_memory.ListWorkspaceMemoriesRequest.new,
    agentsv1workspace_memory.ListWorkspaceMemoriesResponse.new,
  );

  /// SearchWorkspaceMemories runs a semantic search in one scope.
  static const searchWorkspaceMemories = connect.Spec(
    '/$name/SearchWorkspaceMemories',
    connect.StreamType.unary,
    agentsv1workspace_memory.SearchWorkspaceMemoriesRequest.new,
    agentsv1workspace_memory.SearchWorkspaceMemoriesResponse.new,
  );

  /// DeleteWorkspaceMemory deletes one memory of this workspace. A memory
  /// that does not exist or belongs to another workspace is NotFound.
  static const deleteWorkspaceMemory = connect.Spec(
    '/$name/DeleteWorkspaceMemory',
    connect.StreamType.unary,
    agentsv1workspace_memory.DeleteWorkspaceMemoryRequest.new,
    agentsv1workspace_memory.DeleteWorkspaceMemoryResponse.new,
  );
}
