// This is a generated file - do not edit.
//
// Generated from agents/v1/automation.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'automation.pb.dart' as $2;
import 'automation.pbjson.dart';

export 'automation.pb.dart';

abstract class AutomationServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListAutomationsResponse> listAutomations(
      $pb.ServerContext ctx, $2.ListAutomationsRequest request);
  $async.Future<$2.GetAutomationResponse> getAutomation(
      $pb.ServerContext ctx, $2.GetAutomationRequest request);
  $async.Future<$2.CreateAutomationResponse> createAutomation(
      $pb.ServerContext ctx, $2.CreateAutomationRequest request);
  $async.Future<$2.UpdateAutomationResponse> updateAutomation(
      $pb.ServerContext ctx, $2.UpdateAutomationRequest request);
  $async.Future<$2.DeleteAutomationResponse> deleteAutomation(
      $pb.ServerContext ctx, $2.DeleteAutomationRequest request);
  $async.Future<$2.RunAutomationNowResponse> runAutomationNow(
      $pb.ServerContext ctx, $2.RunAutomationNowRequest request);
  $async.Future<$2.ListAutomationRunsResponse> listAutomationRuns(
      $pb.ServerContext ctx, $2.ListAutomationRunsRequest request);
  $async.Future<$2.GetAutomationRunResponse> getAutomationRun(
      $pb.ServerContext ctx, $2.GetAutomationRunRequest request);
  $async.Future<$2.ListAutomationStepRunsResponse> listAutomationStepRuns(
      $pb.ServerContext ctx, $2.ListAutomationStepRunsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListAutomations':
        return $2.ListAutomationsRequest();
      case 'GetAutomation':
        return $2.GetAutomationRequest();
      case 'CreateAutomation':
        return $2.CreateAutomationRequest();
      case 'UpdateAutomation':
        return $2.UpdateAutomationRequest();
      case 'DeleteAutomation':
        return $2.DeleteAutomationRequest();
      case 'RunAutomationNow':
        return $2.RunAutomationNowRequest();
      case 'ListAutomationRuns':
        return $2.ListAutomationRunsRequest();
      case 'GetAutomationRun':
        return $2.GetAutomationRunRequest();
      case 'ListAutomationStepRuns':
        return $2.ListAutomationStepRunsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListAutomations':
        return listAutomations(ctx, request as $2.ListAutomationsRequest);
      case 'GetAutomation':
        return getAutomation(ctx, request as $2.GetAutomationRequest);
      case 'CreateAutomation':
        return createAutomation(ctx, request as $2.CreateAutomationRequest);
      case 'UpdateAutomation':
        return updateAutomation(ctx, request as $2.UpdateAutomationRequest);
      case 'DeleteAutomation':
        return deleteAutomation(ctx, request as $2.DeleteAutomationRequest);
      case 'RunAutomationNow':
        return runAutomationNow(ctx, request as $2.RunAutomationNowRequest);
      case 'ListAutomationRuns':
        return listAutomationRuns(ctx, request as $2.ListAutomationRunsRequest);
      case 'GetAutomationRun':
        return getAutomationRun(ctx, request as $2.GetAutomationRunRequest);
      case 'ListAutomationStepRuns':
        return listAutomationStepRuns(
            ctx, request as $2.ListAutomationStepRunsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      AutomationServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => AutomationServiceBase$messageJson;
}
