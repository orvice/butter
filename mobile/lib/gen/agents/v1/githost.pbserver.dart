// This is a generated file - do not edit.
//
// Generated from agents/v1/githost.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'githost.pb.dart' as $1;
import 'githost.pbjson.dart';

export 'githost.pb.dart';

abstract class GitHostServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListGitHostsResponse> listGitHosts(
      $pb.ServerContext ctx, $1.ListGitHostsRequest request);
  $async.Future<$1.GetGitHostResponse> getGitHost(
      $pb.ServerContext ctx, $1.GetGitHostRequest request);
  $async.Future<$1.CreateGitHostResponse> createGitHost(
      $pb.ServerContext ctx, $1.CreateGitHostRequest request);
  $async.Future<$1.UpdateGitHostResponse> updateGitHost(
      $pb.ServerContext ctx, $1.UpdateGitHostRequest request);
  $async.Future<$1.DeleteGitHostResponse> deleteGitHost(
      $pb.ServerContext ctx, $1.DeleteGitHostRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListGitHosts':
        return $1.ListGitHostsRequest();
      case 'GetGitHost':
        return $1.GetGitHostRequest();
      case 'CreateGitHost':
        return $1.CreateGitHostRequest();
      case 'UpdateGitHost':
        return $1.UpdateGitHostRequest();
      case 'DeleteGitHost':
        return $1.DeleteGitHostRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListGitHosts':
        return listGitHosts(ctx, request as $1.ListGitHostsRequest);
      case 'GetGitHost':
        return getGitHost(ctx, request as $1.GetGitHostRequest);
      case 'CreateGitHost':
        return createGitHost(ctx, request as $1.CreateGitHostRequest);
      case 'UpdateGitHost':
        return updateGitHost(ctx, request as $1.UpdateGitHostRequest);
      case 'DeleteGitHost':
        return deleteGitHost(ctx, request as $1.DeleteGitHostRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => GitHostServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => GitHostServiceBase$messageJson;
}
