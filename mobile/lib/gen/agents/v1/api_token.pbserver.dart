// This is a generated file - do not edit.
//
// Generated from agents/v1/api_token.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'api_token.pb.dart' as $1;
import 'api_token.pbjson.dart';

export 'api_token.pb.dart';

abstract class APITokenServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListAPITokensResponse> listAPITokens(
      $pb.ServerContext ctx, $1.ListAPITokensRequest request);
  $async.Future<$1.CreateAPITokenResponse> createAPIToken(
      $pb.ServerContext ctx, $1.CreateAPITokenRequest request);
  $async.Future<$1.RevokeAPITokenResponse> revokeAPIToken(
      $pb.ServerContext ctx, $1.RevokeAPITokenRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListAPITokens':
        return $1.ListAPITokensRequest();
      case 'CreateAPIToken':
        return $1.CreateAPITokenRequest();
      case 'RevokeAPIToken':
        return $1.RevokeAPITokenRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListAPITokens':
        return listAPITokens(ctx, request as $1.ListAPITokensRequest);
      case 'CreateAPIToken':
        return createAPIToken(ctx, request as $1.CreateAPITokenRequest);
      case 'RevokeAPIToken':
        return revokeAPIToken(ctx, request as $1.RevokeAPITokenRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => APITokenServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => APITokenServiceBase$messageJson;
}
