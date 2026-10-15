// This is a generated file - do not edit.
//
// Generated from agents/v1/butterbox.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'butterbox.pb.dart' as $1;
import 'butterbox.pbjson.dart';

export 'butterbox.pb.dart';

abstract class ButterBoxServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListButterBoxesResponse> listButterBoxes(
      $pb.ServerContext ctx, $1.ListButterBoxesRequest request);
  $async.Future<$1.GetButterBoxResponse> getButterBox(
      $pb.ServerContext ctx, $1.GetButterBoxRequest request);
  $async.Future<$1.CreateButterBoxResponse> createButterBox(
      $pb.ServerContext ctx, $1.CreateButterBoxRequest request);
  $async.Future<$1.UpdateButterBoxResponse> updateButterBox(
      $pb.ServerContext ctx, $1.UpdateButterBoxRequest request);
  $async.Future<$1.DeleteButterBoxResponse> deleteButterBox(
      $pb.ServerContext ctx, $1.DeleteButterBoxRequest request);
  $async.Future<$1.SetButterBoxTokenResponse> setButterBoxToken(
      $pb.ServerContext ctx, $1.SetButterBoxTokenRequest request);
  $async.Future<$1.GetButterBoxStatusResponse> getButterBoxStatus(
      $pb.ServerContext ctx, $1.GetButterBoxStatusRequest request);
  $async.Future<$1.ListButterBoxModelsResponse> listButterBoxModels(
      $pb.ServerContext ctx, $1.ListButterBoxModelsRequest request);
  $async.Future<$1.ListButterBoxCursorModelsResponse> listButterBoxCursorModels(
      $pb.ServerContext ctx, $1.ListButterBoxCursorModelsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListButterBoxes':
        return $1.ListButterBoxesRequest();
      case 'GetButterBox':
        return $1.GetButterBoxRequest();
      case 'CreateButterBox':
        return $1.CreateButterBoxRequest();
      case 'UpdateButterBox':
        return $1.UpdateButterBoxRequest();
      case 'DeleteButterBox':
        return $1.DeleteButterBoxRequest();
      case 'SetButterBoxToken':
        return $1.SetButterBoxTokenRequest();
      case 'GetButterBoxStatus':
        return $1.GetButterBoxStatusRequest();
      case 'ListButterBoxModels':
        return $1.ListButterBoxModelsRequest();
      case 'ListButterBoxCursorModels':
        return $1.ListButterBoxCursorModelsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListButterBoxes':
        return listButterBoxes(ctx, request as $1.ListButterBoxesRequest);
      case 'GetButterBox':
        return getButterBox(ctx, request as $1.GetButterBoxRequest);
      case 'CreateButterBox':
        return createButterBox(ctx, request as $1.CreateButterBoxRequest);
      case 'UpdateButterBox':
        return updateButterBox(ctx, request as $1.UpdateButterBoxRequest);
      case 'DeleteButterBox':
        return deleteButterBox(ctx, request as $1.DeleteButterBoxRequest);
      case 'SetButterBoxToken':
        return setButterBoxToken(ctx, request as $1.SetButterBoxTokenRequest);
      case 'GetButterBoxStatus':
        return getButterBoxStatus(ctx, request as $1.GetButterBoxStatusRequest);
      case 'ListButterBoxModels':
        return listButterBoxModels(
            ctx, request as $1.ListButterBoxModelsRequest);
      case 'ListButterBoxCursorModels':
        return listButterBoxCursorModels(
            ctx, request as $1.ListButterBoxCursorModelsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => ButterBoxServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ButterBoxServiceBase$messageJson;
}
