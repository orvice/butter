// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_file.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'agent_file.pb.dart' as $1;
import 'agent_file.pbjson.dart';

export 'agent_file.pb.dart';

abstract class AgentFileServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListAgentFileSpacesResponse> listAgentFileSpaces(
      $pb.ServerContext ctx, $1.ListAgentFileSpacesRequest request);
  $async.Future<$1.GetAgentFileSpaceResponse> getAgentFileSpace(
      $pb.ServerContext ctx, $1.GetAgentFileSpaceRequest request);
  $async.Future<$1.CreateAgentFileSpaceResponse> createAgentFileSpace(
      $pb.ServerContext ctx, $1.CreateAgentFileSpaceRequest request);
  $async.Future<$1.UpdateAgentFileSpaceResponse> updateAgentFileSpace(
      $pb.ServerContext ctx, $1.UpdateAgentFileSpaceRequest request);
  $async.Future<$1.DeleteAgentFileSpaceResponse> deleteAgentFileSpace(
      $pb.ServerContext ctx, $1.DeleteAgentFileSpaceRequest request);
  $async.Future<$1.ListAgentFilesResponse> listAgentFiles(
      $pb.ServerContext ctx, $1.ListAgentFilesRequest request);
  $async.Future<$1.GetAgentFileResponse> getAgentFile(
      $pb.ServerContext ctx, $1.GetAgentFileRequest request);
  $async.Future<$1.WriteAgentFileResponse> writeAgentFile(
      $pb.ServerContext ctx, $1.WriteAgentFileRequest request);
  $async.Future<$1.DeleteAgentFileResponse> deleteAgentFile(
      $pb.ServerContext ctx, $1.DeleteAgentFileRequest request);
  $async.Future<$1.SearchAgentFilesResponse> searchAgentFiles(
      $pb.ServerContext ctx, $1.SearchAgentFilesRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListAgentFileSpaces':
        return $1.ListAgentFileSpacesRequest();
      case 'GetAgentFileSpace':
        return $1.GetAgentFileSpaceRequest();
      case 'CreateAgentFileSpace':
        return $1.CreateAgentFileSpaceRequest();
      case 'UpdateAgentFileSpace':
        return $1.UpdateAgentFileSpaceRequest();
      case 'DeleteAgentFileSpace':
        return $1.DeleteAgentFileSpaceRequest();
      case 'ListAgentFiles':
        return $1.ListAgentFilesRequest();
      case 'GetAgentFile':
        return $1.GetAgentFileRequest();
      case 'WriteAgentFile':
        return $1.WriteAgentFileRequest();
      case 'DeleteAgentFile':
        return $1.DeleteAgentFileRequest();
      case 'SearchAgentFiles':
        return $1.SearchAgentFilesRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListAgentFileSpaces':
        return listAgentFileSpaces(
            ctx, request as $1.ListAgentFileSpacesRequest);
      case 'GetAgentFileSpace':
        return getAgentFileSpace(ctx, request as $1.GetAgentFileSpaceRequest);
      case 'CreateAgentFileSpace':
        return createAgentFileSpace(
            ctx, request as $1.CreateAgentFileSpaceRequest);
      case 'UpdateAgentFileSpace':
        return updateAgentFileSpace(
            ctx, request as $1.UpdateAgentFileSpaceRequest);
      case 'DeleteAgentFileSpace':
        return deleteAgentFileSpace(
            ctx, request as $1.DeleteAgentFileSpaceRequest);
      case 'ListAgentFiles':
        return listAgentFiles(ctx, request as $1.ListAgentFilesRequest);
      case 'GetAgentFile':
        return getAgentFile(ctx, request as $1.GetAgentFileRequest);
      case 'WriteAgentFile':
        return writeAgentFile(ctx, request as $1.WriteAgentFileRequest);
      case 'DeleteAgentFile':
        return deleteAgentFile(ctx, request as $1.DeleteAgentFileRequest);
      case 'SearchAgentFiles':
        return searchAgentFiles(ctx, request as $1.SearchAgentFilesRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => AgentFileServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => AgentFileServiceBase$messageJson;
}
