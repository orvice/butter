// This is a generated file - do not edit.
//
// Generated from agents/v1/linear.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'linear.pb.dart' as $1;
import 'linear.pbjson.dart';

export 'linear.pb.dart';

abstract class LinearAppServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListLinearAppsResponse> listLinearApps(
      $pb.ServerContext ctx, $1.ListLinearAppsRequest request);
  $async.Future<$1.GetLinearAppResponse> getLinearApp(
      $pb.ServerContext ctx, $1.GetLinearAppRequest request);
  $async.Future<$1.CreateLinearAppResponse> createLinearApp(
      $pb.ServerContext ctx, $1.CreateLinearAppRequest request);
  $async.Future<$1.UpdateLinearAppResponse> updateLinearApp(
      $pb.ServerContext ctx, $1.UpdateLinearAppRequest request);
  $async.Future<$1.PutLinearAppCredentialsResponse> putLinearAppCredentials(
      $pb.ServerContext ctx, $1.PutLinearAppCredentialsRequest request);
  $async.Future<$1.DeleteLinearAppResponse> deleteLinearApp(
      $pb.ServerContext ctx, $1.DeleteLinearAppRequest request);
  $async.Future<$1.BeginLinearInstallResponse> beginLinearInstall(
      $pb.ServerContext ctx, $1.BeginLinearInstallRequest request);
  $async.Future<$1.ListLinearInstallationsResponse> listLinearInstallations(
      $pb.ServerContext ctx, $1.ListLinearInstallationsRequest request);
  $async.Future<$1.DeleteLinearInstallationResponse> deleteLinearInstallation(
      $pb.ServerContext ctx, $1.DeleteLinearInstallationRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListLinearApps':
        return $1.ListLinearAppsRequest();
      case 'GetLinearApp':
        return $1.GetLinearAppRequest();
      case 'CreateLinearApp':
        return $1.CreateLinearAppRequest();
      case 'UpdateLinearApp':
        return $1.UpdateLinearAppRequest();
      case 'PutLinearAppCredentials':
        return $1.PutLinearAppCredentialsRequest();
      case 'DeleteLinearApp':
        return $1.DeleteLinearAppRequest();
      case 'BeginLinearInstall':
        return $1.BeginLinearInstallRequest();
      case 'ListLinearInstallations':
        return $1.ListLinearInstallationsRequest();
      case 'DeleteLinearInstallation':
        return $1.DeleteLinearInstallationRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListLinearApps':
        return listLinearApps(ctx, request as $1.ListLinearAppsRequest);
      case 'GetLinearApp':
        return getLinearApp(ctx, request as $1.GetLinearAppRequest);
      case 'CreateLinearApp':
        return createLinearApp(ctx, request as $1.CreateLinearAppRequest);
      case 'UpdateLinearApp':
        return updateLinearApp(ctx, request as $1.UpdateLinearAppRequest);
      case 'PutLinearAppCredentials':
        return putLinearAppCredentials(
            ctx, request as $1.PutLinearAppCredentialsRequest);
      case 'DeleteLinearApp':
        return deleteLinearApp(ctx, request as $1.DeleteLinearAppRequest);
      case 'BeginLinearInstall':
        return beginLinearInstall(ctx, request as $1.BeginLinearInstallRequest);
      case 'ListLinearInstallations':
        return listLinearInstallations(
            ctx, request as $1.ListLinearInstallationsRequest);
      case 'DeleteLinearInstallation':
        return deleteLinearInstallation(
            ctx, request as $1.DeleteLinearInstallationRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => LinearAppServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => LinearAppServiceBase$messageJson;
}

abstract class LinearAdminServiceBase extends $pb.GeneratedService {
  $async.Future<$1.GetLinearSettingsResponse> getLinearSettings(
      $pb.ServerContext ctx, $1.GetLinearSettingsRequest request);
  $async.Future<$1.UpdateLinearSettingsResponse> updateLinearSettings(
      $pb.ServerContext ctx, $1.UpdateLinearSettingsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'GetLinearSettings':
        return $1.GetLinearSettingsRequest();
      case 'UpdateLinearSettings':
        return $1.UpdateLinearSettingsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'GetLinearSettings':
        return getLinearSettings(ctx, request as $1.GetLinearSettingsRequest);
      case 'UpdateLinearSettings':
        return updateLinearSettings(
            ctx, request as $1.UpdateLinearSettingsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      LinearAdminServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => LinearAdminServiceBase$messageJson;
}

abstract class LinearProcessingServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListLinearProcessingRecordsResponse>
      listLinearProcessingRecords(
          $pb.ServerContext ctx, $1.ListLinearProcessingRecordsRequest request);
  $async.Future<$1.GetLinearProcessingRecordResponse> getLinearProcessingRecord(
      $pb.ServerContext ctx, $1.GetLinearProcessingRecordRequest request);
  $async.Future<$1.ResendLinearReplyResponse> resendLinearReply(
      $pb.ServerContext ctx, $1.ResendLinearReplyRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListLinearProcessingRecords':
        return $1.ListLinearProcessingRecordsRequest();
      case 'GetLinearProcessingRecord':
        return $1.GetLinearProcessingRecordRequest();
      case 'ResendLinearReply':
        return $1.ResendLinearReplyRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListLinearProcessingRecords':
        return listLinearProcessingRecords(
            ctx, request as $1.ListLinearProcessingRecordsRequest);
      case 'GetLinearProcessingRecord':
        return getLinearProcessingRecord(
            ctx, request as $1.GetLinearProcessingRecordRequest);
      case 'ResendLinearReply':
        return resendLinearReply(ctx, request as $1.ResendLinearReplyRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      LinearProcessingServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => LinearProcessingServiceBase$messageJson;
}
