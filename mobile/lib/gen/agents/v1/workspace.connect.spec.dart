//
//  Generated code. Do not modify.
//  source: agents/v1/workspace.proto
//

import "package:connectrpc/connect.dart" as connect;
import "workspace.pb.dart" as agentsv1workspace;

/// WorkspaceService manages workspaces and memberships.
abstract final class WorkspaceService {
  /// Fully-qualified name of the WorkspaceService service.
  static const name = 'agents.v1.WorkspaceService';

  /// ListWorkspaces returns all workspaces the caller can access. Global
  /// admins receive every workspace; other users receive only the ones they
  /// are members of.
  static const listWorkspaces = connect.Spec(
    '/$name/ListWorkspaces',
    connect.StreamType.unary,
    agentsv1workspace.ListWorkspacesRequest.new,
    agentsv1workspace.ListWorkspacesResponse.new,
  );

  /// GetWorkspace returns a single workspace by id.
  static const getWorkspace = connect.Spec(
    '/$name/GetWorkspace',
    connect.StreamType.unary,
    agentsv1workspace.GetWorkspaceRequest.new,
    agentsv1workspace.GetWorkspaceResponse.new,
  );

  /// CreateWorkspace creates a new workspace. The caller is added as the
  /// initial owner.
  static const createWorkspace = connect.Spec(
    '/$name/CreateWorkspace',
    connect.StreamType.unary,
    agentsv1workspace.CreateWorkspaceRequest.new,
    agentsv1workspace.CreateWorkspaceResponse.new,
  );

  /// UpdateWorkspace updates workspace metadata.
  static const updateWorkspace = connect.Spec(
    '/$name/UpdateWorkspace',
    connect.StreamType.unary,
    agentsv1workspace.UpdateWorkspaceRequest.new,
    agentsv1workspace.UpdateWorkspaceResponse.new,
  );

  /// DeleteWorkspace permanently removes a workspace and all of its scoped
  /// configs. Caller must be an owner.
  static const deleteWorkspace = connect.Spec(
    '/$name/DeleteWorkspace',
    connect.StreamType.unary,
    agentsv1workspace.DeleteWorkspaceRequest.new,
    agentsv1workspace.DeleteWorkspaceResponse.new,
  );

  /// ListWorkspaceMembers returns the membership list of a workspace.
  static const listWorkspaceMembers = connect.Spec(
    '/$name/ListWorkspaceMembers',
    connect.StreamType.unary,
    agentsv1workspace.ListWorkspaceMembersRequest.new,
    agentsv1workspace.ListWorkspaceMembersResponse.new,
  );

  /// AddWorkspaceMember adds a user to a workspace with a given role.
  static const addWorkspaceMember = connect.Spec(
    '/$name/AddWorkspaceMember',
    connect.StreamType.unary,
    agentsv1workspace.AddWorkspaceMemberRequest.new,
    agentsv1workspace.AddWorkspaceMemberResponse.new,
  );

  /// UpdateWorkspaceMember changes a member's role.
  static const updateWorkspaceMember = connect.Spec(
    '/$name/UpdateWorkspaceMember',
    connect.StreamType.unary,
    agentsv1workspace.UpdateWorkspaceMemberRequest.new,
    agentsv1workspace.UpdateWorkspaceMemberResponse.new,
  );

  /// RemoveWorkspaceMember removes a user from a workspace.
  static const removeWorkspaceMember = connect.Spec(
    '/$name/RemoveWorkspaceMember',
    connect.StreamType.unary,
    agentsv1workspace.RemoveWorkspaceMemberRequest.new,
    agentsv1workspace.RemoveWorkspaceMemberResponse.new,
  );
}
