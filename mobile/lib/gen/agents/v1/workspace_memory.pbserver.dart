// This is a generated file - do not edit.
//
// Generated from agents/v1/workspace_memory.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'workspace_memory.pb.dart' as $1;
import 'workspace_memory.pbjson.dart';

export 'workspace_memory.pb.dart';

abstract class WorkspaceMemoryConfigServiceBase extends $pb.GeneratedService {
  $async.Future<$1.GetWorkspaceMemoryConfigResponse> getWorkspaceMemoryConfig(
      $pb.ServerContext ctx, $1.GetWorkspaceMemoryConfigRequest request);
  $async.Future<$1.PutWorkspaceMemoryConfigResponse> putWorkspaceMemoryConfig(
      $pb.ServerContext ctx, $1.PutWorkspaceMemoryConfigRequest request);
  $async.Future<$1.DeleteWorkspaceMemoryConfigResponse>
      deleteWorkspaceMemoryConfig(
          $pb.ServerContext ctx, $1.DeleteWorkspaceMemoryConfigRequest request);
  $async.Future<$1.TestWorkspaceMemoryConnectionResponse>
      testWorkspaceMemoryConnection($pb.ServerContext ctx,
          $1.TestWorkspaceMemoryConnectionRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'GetWorkspaceMemoryConfig':
        return $1.GetWorkspaceMemoryConfigRequest();
      case 'PutWorkspaceMemoryConfig':
        return $1.PutWorkspaceMemoryConfigRequest();
      case 'DeleteWorkspaceMemoryConfig':
        return $1.DeleteWorkspaceMemoryConfigRequest();
      case 'TestWorkspaceMemoryConnection':
        return $1.TestWorkspaceMemoryConnectionRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'GetWorkspaceMemoryConfig':
        return getWorkspaceMemoryConfig(
            ctx, request as $1.GetWorkspaceMemoryConfigRequest);
      case 'PutWorkspaceMemoryConfig':
        return putWorkspaceMemoryConfig(
            ctx, request as $1.PutWorkspaceMemoryConfigRequest);
      case 'DeleteWorkspaceMemoryConfig':
        return deleteWorkspaceMemoryConfig(
            ctx, request as $1.DeleteWorkspaceMemoryConfigRequest);
      case 'TestWorkspaceMemoryConnection':
        return testWorkspaceMemoryConnection(
            ctx, request as $1.TestWorkspaceMemoryConnectionRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      WorkspaceMemoryConfigServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => WorkspaceMemoryConfigServiceBase$messageJson;
}

abstract class WorkspaceMemoryServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListWorkspaceMemoriesResponse> listWorkspaceMemories(
      $pb.ServerContext ctx, $1.ListWorkspaceMemoriesRequest request);
  $async.Future<$1.SearchWorkspaceMemoriesResponse> searchWorkspaceMemories(
      $pb.ServerContext ctx, $1.SearchWorkspaceMemoriesRequest request);
  $async.Future<$1.DeleteWorkspaceMemoryResponse> deleteWorkspaceMemory(
      $pb.ServerContext ctx, $1.DeleteWorkspaceMemoryRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListWorkspaceMemories':
        return $1.ListWorkspaceMemoriesRequest();
      case 'SearchWorkspaceMemories':
        return $1.SearchWorkspaceMemoriesRequest();
      case 'DeleteWorkspaceMemory':
        return $1.DeleteWorkspaceMemoryRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListWorkspaceMemories':
        return listWorkspaceMemories(
            ctx, request as $1.ListWorkspaceMemoriesRequest);
      case 'SearchWorkspaceMemories':
        return searchWorkspaceMemories(
            ctx, request as $1.SearchWorkspaceMemoriesRequest);
      case 'DeleteWorkspaceMemory':
        return deleteWorkspaceMemory(
            ctx, request as $1.DeleteWorkspaceMemoryRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      WorkspaceMemoryServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => WorkspaceMemoryServiceBase$messageJson;
}
