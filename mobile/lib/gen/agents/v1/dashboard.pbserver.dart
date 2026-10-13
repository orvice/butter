// This is a generated file - do not edit.
//
// Generated from agents/v1/dashboard.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'dashboard.pb.dart' as $5;
import 'dashboard.pbjson.dart';

export 'dashboard.pb.dart';

abstract class DashboardServiceBase extends $pb.GeneratedService {
  $async.Future<$5.GetOverviewResponse> getOverview(
      $pb.ServerContext ctx, $5.GetOverviewRequest request);
  $async.Future<$5.GetActivityFeedResponse> getActivityFeed(
      $pb.ServerContext ctx, $5.GetActivityFeedRequest request);
  $async.Future<$5.GetCronExecutionTimeseriesResponse>
      getCronExecutionTimeseries(
          $pb.ServerContext ctx, $5.GetCronExecutionTimeseriesRequest request);
  $async.Future<$5.GetActivityMetricsResponse> getActivityMetrics(
      $pb.ServerContext ctx, $5.GetActivityMetricsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'GetOverview':
        return $5.GetOverviewRequest();
      case 'GetActivityFeed':
        return $5.GetActivityFeedRequest();
      case 'GetCronExecutionTimeseries':
        return $5.GetCronExecutionTimeseriesRequest();
      case 'GetActivityMetrics':
        return $5.GetActivityMetricsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'GetOverview':
        return getOverview(ctx, request as $5.GetOverviewRequest);
      case 'GetActivityFeed':
        return getActivityFeed(ctx, request as $5.GetActivityFeedRequest);
      case 'GetCronExecutionTimeseries':
        return getCronExecutionTimeseries(
            ctx, request as $5.GetCronExecutionTimeseriesRequest);
      case 'GetActivityMetrics':
        return getActivityMetrics(ctx, request as $5.GetActivityMetricsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => DashboardServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => DashboardServiceBase$messageJson;
}

abstract class DaemonServiceBase extends $pb.GeneratedService {
  $async.Future<$5.ListDaemonRuntimesResponse> listDaemonRuntimes(
      $pb.ServerContext ctx, $5.ListDaemonRuntimesRequest request);
  $async.Future<$5.GetDaemonRuntimeResponse> getDaemonRuntime(
      $pb.ServerContext ctx, $5.GetDaemonRuntimeRequest request);
  $async.Future<$5.CreateDaemonRuntimeResponse> createDaemonRuntime(
      $pb.ServerContext ctx, $5.CreateDaemonRuntimeRequest request);
  $async.Future<$5.UpdateDaemonRuntimeResponse> updateDaemonRuntime(
      $pb.ServerContext ctx, $5.UpdateDaemonRuntimeRequest request);
  $async.Future<$5.DeleteDaemonRuntimeResponse> deleteDaemonRuntime(
      $pb.ServerContext ctx, $5.DeleteDaemonRuntimeRequest request);
  $async.Future<$5.CreateDaemonRuntimeTokenResponse> createDaemonRuntimeToken(
      $pb.ServerContext ctx, $5.CreateDaemonRuntimeTokenRequest request);
  $async.Future<$5.ListDaemonsResponse> listDaemons(
      $pb.ServerContext ctx, $5.ListDaemonsRequest request);
  $async.Future<$5.GetDaemonResponse> getDaemon(
      $pb.ServerContext ctx, $5.GetDaemonRequest request);
  $async.Future<$5.CancelDaemonTaskResponse> cancelDaemonTask(
      $pb.ServerContext ctx, $5.CancelDaemonTaskRequest request);
  $async.Future<$5.ListDaemonTasksResponse> listDaemonTasks(
      $pb.ServerContext ctx, $5.ListDaemonTasksRequest request);
  $async.Future<$5.GetBridgeDiagnosticsResponse> getBridgeDiagnostics(
      $pb.ServerContext ctx, $5.GetBridgeDiagnosticsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListDaemonRuntimes':
        return $5.ListDaemonRuntimesRequest();
      case 'GetDaemonRuntime':
        return $5.GetDaemonRuntimeRequest();
      case 'CreateDaemonRuntime':
        return $5.CreateDaemonRuntimeRequest();
      case 'UpdateDaemonRuntime':
        return $5.UpdateDaemonRuntimeRequest();
      case 'DeleteDaemonRuntime':
        return $5.DeleteDaemonRuntimeRequest();
      case 'CreateDaemonRuntimeToken':
        return $5.CreateDaemonRuntimeTokenRequest();
      case 'ListDaemons':
        return $5.ListDaemonsRequest();
      case 'GetDaemon':
        return $5.GetDaemonRequest();
      case 'CancelDaemonTask':
        return $5.CancelDaemonTaskRequest();
      case 'ListDaemonTasks':
        return $5.ListDaemonTasksRequest();
      case 'GetBridgeDiagnostics':
        return $5.GetBridgeDiagnosticsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListDaemonRuntimes':
        return listDaemonRuntimes(ctx, request as $5.ListDaemonRuntimesRequest);
      case 'GetDaemonRuntime':
        return getDaemonRuntime(ctx, request as $5.GetDaemonRuntimeRequest);
      case 'CreateDaemonRuntime':
        return createDaemonRuntime(
            ctx, request as $5.CreateDaemonRuntimeRequest);
      case 'UpdateDaemonRuntime':
        return updateDaemonRuntime(
            ctx, request as $5.UpdateDaemonRuntimeRequest);
      case 'DeleteDaemonRuntime':
        return deleteDaemonRuntime(
            ctx, request as $5.DeleteDaemonRuntimeRequest);
      case 'CreateDaemonRuntimeToken':
        return createDaemonRuntimeToken(
            ctx, request as $5.CreateDaemonRuntimeTokenRequest);
      case 'ListDaemons':
        return listDaemons(ctx, request as $5.ListDaemonsRequest);
      case 'GetDaemon':
        return getDaemon(ctx, request as $5.GetDaemonRequest);
      case 'CancelDaemonTask':
        return cancelDaemonTask(ctx, request as $5.CancelDaemonTaskRequest);
      case 'ListDaemonTasks':
        return listDaemonTasks(ctx, request as $5.ListDaemonTasksRequest);
      case 'GetBridgeDiagnostics':
        return getBridgeDiagnostics(
            ctx, request as $5.GetBridgeDiagnosticsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => DaemonServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => DaemonServiceBase$messageJson;
}
