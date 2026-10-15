//
//  Generated code. Do not modify.
//  source: agents/v1/workspace.proto
//

import "package:connectrpc/connect.dart" as connect;
import "workspace.pb.dart" as agentsv1workspace;
import "workspace.connect.spec.dart" as specs;

/// WorkspaceService manages workspaces and memberships.
extension type WorkspaceServiceClient(connect.Transport _transport) {
  /// ListWorkspaces returns all workspaces the caller can access. Global
  /// admins receive every workspace; other users receive only the ones they
  /// are members of.
  Future<agentsv1workspace.ListWorkspacesResponse> listWorkspaces(
    agentsv1workspace.ListWorkspacesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.listWorkspaces,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetWorkspace returns a single workspace by id.
  Future<agentsv1workspace.GetWorkspaceResponse> getWorkspace(
    agentsv1workspace.GetWorkspaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.getWorkspace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateWorkspace creates a new workspace. The caller is added as the
  /// initial owner.
  Future<agentsv1workspace.CreateWorkspaceResponse> createWorkspace(
    agentsv1workspace.CreateWorkspaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.createWorkspace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateWorkspace updates workspace metadata.
  Future<agentsv1workspace.UpdateWorkspaceResponse> updateWorkspace(
    agentsv1workspace.UpdateWorkspaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.updateWorkspace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteWorkspace permanently removes a workspace and all of its scoped
  /// configs. Caller must be an owner.
  Future<agentsv1workspace.DeleteWorkspaceResponse> deleteWorkspace(
    agentsv1workspace.DeleteWorkspaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.deleteWorkspace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListWorkspaceMembers returns the membership list of a workspace.
  Future<agentsv1workspace.ListWorkspaceMembersResponse> listWorkspaceMembers(
    agentsv1workspace.ListWorkspaceMembersRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.listWorkspaceMembers,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// AddWorkspaceMember adds a user to a workspace with a given role.
  Future<agentsv1workspace.AddWorkspaceMemberResponse> addWorkspaceMember(
    agentsv1workspace.AddWorkspaceMemberRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.addWorkspaceMember,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateWorkspaceMember changes a member's role.
  Future<agentsv1workspace.UpdateWorkspaceMemberResponse> updateWorkspaceMember(
    agentsv1workspace.UpdateWorkspaceMemberRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.updateWorkspaceMember,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RemoveWorkspaceMember removes a user from a workspace.
  Future<agentsv1workspace.RemoveWorkspaceMemberResponse> removeWorkspaceMember(
    agentsv1workspace.RemoveWorkspaceMemberRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.removeWorkspaceMember,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
