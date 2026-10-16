// This is a generated file - do not edit.
//
// Generated from agents/v1/cron.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'cron.pb.dart' as $2;
import 'cron.pbjson.dart';

export 'cron.pb.dart';

abstract class CronJobServiceBase extends $pb.GeneratedService {
  $async.Future<$2.ListCronJobsResponse> listCronJobs(
      $pb.ServerContext ctx, $2.ListCronJobsRequest request);
  $async.Future<$2.GetCronJobResponse> getCronJob(
      $pb.ServerContext ctx, $2.GetCronJobRequest request);
  $async.Future<$2.CreateCronJobResponse> createCronJob(
      $pb.ServerContext ctx, $2.CreateCronJobRequest request);
  $async.Future<$2.UpdateCronJobResponse> updateCronJob(
      $pb.ServerContext ctx, $2.UpdateCronJobRequest request);
  $async.Future<$2.DeleteCronJobResponse> deleteCronJob(
      $pb.ServerContext ctx, $2.DeleteCronJobRequest request);
  $async.Future<$2.ListCronExecutionsResponse> listCronExecutions(
      $pb.ServerContext ctx, $2.ListCronExecutionsRequest request);
  $async.Future<$2.RunCronJobNowResponse> runCronJobNow(
      $pb.ServerContext ctx, $2.RunCronJobNowRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListCronJobs':
        return $2.ListCronJobsRequest();
      case 'GetCronJob':
        return $2.GetCronJobRequest();
      case 'CreateCronJob':
        return $2.CreateCronJobRequest();
      case 'UpdateCronJob':
        return $2.UpdateCronJobRequest();
      case 'DeleteCronJob':
        return $2.DeleteCronJobRequest();
      case 'ListCronExecutions':
        return $2.ListCronExecutionsRequest();
      case 'RunCronJobNow':
        return $2.RunCronJobNowRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListCronJobs':
        return listCronJobs(ctx, request as $2.ListCronJobsRequest);
      case 'GetCronJob':
        return getCronJob(ctx, request as $2.GetCronJobRequest);
      case 'CreateCronJob':
        return createCronJob(ctx, request as $2.CreateCronJobRequest);
      case 'UpdateCronJob':
        return updateCronJob(ctx, request as $2.UpdateCronJobRequest);
      case 'DeleteCronJob':
        return deleteCronJob(ctx, request as $2.DeleteCronJobRequest);
      case 'ListCronExecutions':
        return listCronExecutions(ctx, request as $2.ListCronExecutionsRequest);
      case 'RunCronJobNow':
        return runCronJobNow(ctx, request as $2.RunCronJobNowRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => CronJobServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => CronJobServiceBase$messageJson;
}
