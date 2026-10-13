// This is a generated file - do not edit.
//
// Generated from agents/v1/workspace.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'workspace.pb.dart' as $1;
import 'workspace.pbjson.dart';

export 'workspace.pb.dart';

abstract class WorkspaceServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListWorkspacesResponse> listWorkspaces(
      $pb.ServerContext ctx, $1.ListWorkspacesRequest request);
  $async.Future<$1.GetWorkspaceResponse> getWorkspace(
      $pb.ServerContext ctx, $1.GetWorkspaceRequest request);
  $async.Future<$1.CreateWorkspaceResponse> createWorkspace(
      $pb.ServerContext ctx, $1.CreateWorkspaceRequest request);
  $async.Future<$1.UpdateWorkspaceResponse> updateWorkspace(
      $pb.ServerContext ctx, $1.UpdateWorkspaceRequest request);
  $async.Future<$1.DeleteWorkspaceResponse> deleteWorkspace(
      $pb.ServerContext ctx, $1.DeleteWorkspaceRequest request);
  $async.Future<$1.ListWorkspaceMembersResponse> listWorkspaceMembers(
      $pb.ServerContext ctx, $1.ListWorkspaceMembersRequest request);
  $async.Future<$1.AddWorkspaceMemberResponse> addWorkspaceMember(
      $pb.ServerContext ctx, $1.AddWorkspaceMemberRequest request);
  $async.Future<$1.UpdateWorkspaceMemberResponse> updateWorkspaceMember(
      $pb.ServerContext ctx, $1.UpdateWorkspaceMemberRequest request);
  $async.Future<$1.RemoveWorkspaceMemberResponse> removeWorkspaceMember(
      $pb.ServerContext ctx, $1.RemoveWorkspaceMemberRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListWorkspaces':
        return $1.ListWorkspacesRequest();
      case 'GetWorkspace':
        return $1.GetWorkspaceRequest();
      case 'CreateWorkspace':
        return $1.CreateWorkspaceRequest();
      case 'UpdateWorkspace':
        return $1.UpdateWorkspaceRequest();
      case 'DeleteWorkspace':
        return $1.DeleteWorkspaceRequest();
      case 'ListWorkspaceMembers':
        return $1.ListWorkspaceMembersRequest();
      case 'AddWorkspaceMember':
        return $1.AddWorkspaceMemberRequest();
      case 'UpdateWorkspaceMember':
        return $1.UpdateWorkspaceMemberRequest();
      case 'RemoveWorkspaceMember':
        return $1.RemoveWorkspaceMemberRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListWorkspaces':
        return listWorkspaces(ctx, request as $1.ListWorkspacesRequest);
      case 'GetWorkspace':
        return getWorkspace(ctx, request as $1.GetWorkspaceRequest);
      case 'CreateWorkspace':
        return createWorkspace(ctx, request as $1.CreateWorkspaceRequest);
      case 'UpdateWorkspace':
        return updateWorkspace(ctx, request as $1.UpdateWorkspaceRequest);
      case 'DeleteWorkspace':
        return deleteWorkspace(ctx, request as $1.DeleteWorkspaceRequest);
      case 'ListWorkspaceMembers':
        return listWorkspaceMembers(
            ctx, request as $1.ListWorkspaceMembersRequest);
      case 'AddWorkspaceMember':
        return addWorkspaceMember(ctx, request as $1.AddWorkspaceMemberRequest);
      case 'UpdateWorkspaceMember':
        return updateWorkspaceMember(
            ctx, request as $1.UpdateWorkspaceMemberRequest);
      case 'RemoveWorkspaceMember':
        return removeWorkspaceMember(
            ctx, request as $1.RemoveWorkspaceMemberRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => WorkspaceServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => WorkspaceServiceBase$messageJson;
}
